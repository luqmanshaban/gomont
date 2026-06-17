package core

import "time"

type NotificationChannel struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Emails    []string  `json:"emails"`
	CreatedAt time.Time `json:"created_at"`
}