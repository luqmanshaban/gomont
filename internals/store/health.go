package store

import (
	"database/sql"
	"log/slog"
)

type HealthStore struct {
	DB *sql.DB
}

func NewHealthStore(db *sql.DB) *HealthStore {
	return &HealthStore{DB: db}
}

func (hs *HealthStore) CheckDBHealth() error {
	err := hs.DB.Ping()
	if err != nil {
		slog.Error("failed to ping database", "error", err)
		return err 
	}
	return  nil
}