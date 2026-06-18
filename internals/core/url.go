package core

import "time"

type URL struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	UserEmail         string    `json:"user_email"`
	Endpoint          string    `json:"endpoint"`
	IsHealthy         bool      `json:"is_healthy"`
	NotifcationSent   bool      `json:"notification_sent"`
	Retries           int       `json:"retries"`
	MaxRetries        int       `json:"max_retries"`
	Interval          int       `json:"interval"`
	RunsAt            time.Time `json:"runs_at"`
	RetryAt           time.Time `json:"retry_at"`
	LastManualRetryAt time.Time `json:"last_manual_retry_at"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
