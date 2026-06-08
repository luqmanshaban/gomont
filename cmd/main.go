package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luqmanshaban/gomont/internals/api"
	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/store"
)

func main() {
	cfg := config.LoadEnv()

	db := store.ConnectToDb(cfg)
	defer db.Close()

	healthStore := store.NewHealthStore(db)
	userStore := store.NewUserStore(db)
    notificationStore := store.NewNotificationStore(db)

	srv := api.NewServer(cfg, healthStore, userStore, notificationStore)
	srv.Start()
	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<- quit

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	slog.Info("Shutting down server")
	srv.ShutDown(ctx)
}