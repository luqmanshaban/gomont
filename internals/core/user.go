package core

import "time"

type User struct {
	ID                 int       `json:"id"`
	Email              string    `json:"email"`
	Names              string    `json:"names"`
	VerificationCode   int       `json:"verification_code"`
	VerificationExpiry time.Time `json:"verification_expiry"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}