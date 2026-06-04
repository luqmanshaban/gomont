package store

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/luqmanshaban/gomont/internals/config"
	_ "github.com/lib/pq"
)

func ConnectToDb(cfg *config.Config) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.DB_HOST, cfg.DB_PORT, cfg.DB_USER, cfg.DB_PASS, cfg.DB_NAME)
	fmt.Println(connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("failed to connect to database")
		panic(err)
	}

	return db
}