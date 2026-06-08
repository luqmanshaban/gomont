package core

import "time"

type Notification struct {
	ID int 
	UserID int 
	NotificationChannelID int 
	URLID int 
	ErrorMsg string 
	CreatedAt time.Time
}