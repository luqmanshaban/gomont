package core

import "time"

type Notification struct {
	ID                    int       `json:"id"`
	UserID                int       `json:"user_id"`
	NotificationChannelID int       `json:"notification_channel_id"`
	URLID                 int       `json:"url_id"`
	ErrorMsg              string    `json:"error_msg"`
	CreatedAt             time.Time `json:"created_at"`
}