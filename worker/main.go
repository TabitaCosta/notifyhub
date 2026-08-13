package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type LaravelJobPayload struct {
	Data struct {
		Command string `json:"command"`
	} `json:"data"`
}

const (
	numWorkers = 3
	queueKey   = "laravel-database-queues:default"
)

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Erro ao conectar no Redis: %v", err)
	}

	db, err := sql.Open("sqlite3", "../database/database.sqlite")
	if err != nil {
		log.Fatalf("Erro ao conectar no SQLite: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	jobsChannel := make(chan string, 100)
	var wg sync.WaitGroup

	fmt.Printf("Worker Pool iniciado com %d Workers concorrentes em Go!\n", numWorkers)

	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobsChannel, db, &wg)
	}

	go func() {
		for {
			result, err := rdb.BLPop(ctx, 2*time.Second, queueKey).Result()
			if err == redis.Nil {
				continue
			} else if err != nil {
				log.Printf("Erro na leitura do Redis: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			jobsChannel <- result[1]
		}
	}()

	select {}
}

func worker(id int, jobs <-chan string, db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	for jobData := range jobs {
		var payload LaravelJobPayload
		if err := json.Unmarshal([]byte(jobData), &payload); err != nil {
			log.Printf("[Worker %d] Erro no JSON: %v", id, err)
			continue
		}

		notificationID := extractNotificationID(payload.Data.Command)
		if notificationID == 0 {
			log.Printf("[Worker %d] ID não encontrado no payload.", id)
			continue
		}

		fmt.Printf("[Worker %d] Processando Notificação ID: %d...\n", id, notificationID)

		time.Sleep(2 * time.Second)

		query := `UPDATE notifications SET status = ?, attempts = attempts + 1, updated_at = ? WHERE id = ?`
		now := time.Now().Format("2006-01-02 15:04:05")

		_, err := db.Exec(query, "processed", now, notificationID)
		if err != nil {
			log.Printf("[Worker %d] Erro ao atualizar ID %d: %v", id, notificationID, err)
		} else {
			fmt.Printf("[Worker %d] Notificação ID %d concluída!\n", id, notificationID)
		}
	}
}

func extractNotificationID(command string) int {
	re := regexp.MustCompile(`"id";i:(\d+)`)
	matches := re.FindStringSubmatch(command)
	if len(matches) > 1 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}
	return 0
}
