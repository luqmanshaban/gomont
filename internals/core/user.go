package core

import "time"

type User struct {
	ID int 
	Email string
	Names string
	VerificationCode int 
	VerificationExpiry time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}