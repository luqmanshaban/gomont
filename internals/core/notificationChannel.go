package core

import "time"

type NotificationChannel struct {
	ID int 
	UserID int 
	Emails []string 
	CreatedAt time.Time
}