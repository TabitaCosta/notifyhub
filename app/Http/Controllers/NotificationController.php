<?php

namespace App\Http\Controllers;

use App\Models\Notification;
use App\Jobs\ProcessNotification;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class NotificationController extends Controller
{
    public function index(): JsonResponse
    {
        $notifications = Notification::latest()->get();
        return response()->json($notifications, 200);
    }

    public function store(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'recipient' => ['required', 'string', 'max:255'],
            'channel'   => ['required', 'string', 'in:email,sms,push'],
            'subject'   => ['nullable', 'string', 'max:255'],
            'content'   => ['required', 'string'],
        ]);

        $notification = Notification::create($validated);

        ProcessNotification::dispatch($notification);

        return response()->json($notification, 201);
    }

    public function show(string $id): JsonResponse
    {
        $notification = Notification::find($id);

        if (!$notification) {
            return response()->json(['message' => 'Notification not found'], 404);
        }

        return response()->json($notification, 200);
    }
}