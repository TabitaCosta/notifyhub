Markdown
# NotifyHub

O **NotifyHub** é um microserviço distribuído e orientado a eventos para envio de notificações em massa de alta performance. 

O objetivo do projeto é resolver o problema de gargalo em requisições HTTP: a API recebe o disparo de notificação, joga a tarefa em uma fila de alta velocidade e responde instantaneamente ao cliente, delegando todo o processamento pesado e concorrente para um Worker em segundo plano.

---

## Arquitetura e Tecnologias

A aplicação utiliza uma abordagem híbrida aproveitando o melhor de cada tecnologia:

* **PHP 8.2+ / Laravel (API REST):** Responsável pelo recebimento das requisições, validações, persistência inicial e despacho de Jobs.
* **Redis (Message Broker):** Fila em memória de altíssima velocidade para desacoplamento e retenção das mensagens.
* **Go / Golang (Worker Pool):** Consumidor da fila em segundo plano que utiliza **Goroutines** e **Channels** para processar disparos de forma concorrente e paralela.
* **SQLite / Docker:** Banco de dados relacional leve e ambiente para execução de serviços.

[ Cliente HTTP ]
│
▼
[ API Laravel ] ──(1. Salva 'pending')──► [ SQLite ]
│
▼ (2. Push Job em < 50ms)
[ Redis ]
│
▼ (3. BLPop & Distribuição)
[ Worker Pool (Go) ] ──(4. Processa e Atualiza 'processed')──► [ SQLite ]


---

## Concorrência e Performance

O Worker em Go implementa o padrão **Worker Pool**:
* **Goroutines:** Múltiplas threads leves processando registros em paralelo.
* **Channels:** Comunicação thread-safe entre a leitura do Redis e a execução dos trabalhadores.
* **Controle de Pool SQLite:** Configuração de `SetMaxOpenConns(1)` para escrita encadeada e segura, evitando travamentos de arquivo (`database is locked`) sem perder a concorrência do processamento.

---

## Pré-requisitos

* **PHP 8.2+** & **Composer**
* **Go 1.20+**
* **Docker** & **Docker Compose**

---

## Como Rodar o Projeto

### 1. Clonar o repositório

git clone [https://github.com/seu-usuario/notifyhub.git](https://github.com/seu-usuario/notifyhub.git)
cd notifyhub

2. Configurar a API Laravel

composer install
cp .env.example .env
php artisan key:generate
touch database/database.sqlite
php artisan migrate


3. Subir a Fila (Redis) com Docker

docker-compose up -d

4. Iniciar a API Laravel

php artisan serve
# A API estará disponível em: [http://127.0.0.1:8000](http://127.0.0.1:8000)


5. Iniciar o Worker em Go
Em outro terminal, acesse a pasta worker e inicie o serviço:

cd worker
go run main.go

Testando o Envio Concorrente
Em um terceiro terminal, execute o comando abaixo para disparar 5 requisições simuladas em paralelo:

for i in {1..5}; do
  curl -s -X POST [http://127.0.0.1:8000/api/notifications](http://127.0.0.1:8000/api/notifications) \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "{\"recipient\":\"dev$i@notifyhub.com\",\"channel\":\"email\",\"subject\":\"Teste $i\",\"content\":\"Corpo $i\"}" &
done
Acompanhe os logs no terminal do Go vendo os workers ([Worker 1], [Worker 2], etc.) dividindo a carga em tempo real.