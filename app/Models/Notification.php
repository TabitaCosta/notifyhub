<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Notification extends Model
{
    protected $fillable = [
        'recipient',
        'channel',
        'subject',
        'content',
        'status',
        'attempts',
    ];
}