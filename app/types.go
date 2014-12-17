package main

import (
    "time"
)

type QueuedUser struct {
	UserID     string
	Discovered time.Time
}