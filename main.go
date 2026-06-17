package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luqmanshaban/gomont/internals/api"
	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/core"
	"github.com/luqmanshaban/gomont/internals/sse"
	"github.com/luqmanshaban/gomont/internals/store"
	"github.com/luqmanshaban/gomont/internals/worker"
)

func main() {
	cfg := config.LoadEnv()
	fmt.Println("port: ", cfg.PORT)
	fmt.Println("psql port: ", cfg.DB_PORT)

	// Initialize structured logging handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	db := store.ConnectToDb(cfg)
	defer db.Close()

	// Queue for holding urls
	jobCh := make(chan core.URL, 100)

	healthStore := store.NewHealthStore(db)
	userStore := store.NewUserStore(db)
	notificationChannelStore := store.NewNotificationChannelStore(db)
	urlStore := store.NewURLStore(db)

	// handle stale urls
	if err := urlStore.ResetStaleURLs(); err != nil {
		slog.Error("failed to reset stale urls", "error", err)
		panic(err)
	}

	// SSE broker — shared between the worker pool (which publishes status
	// changes) and the API server (which lets the dashboard subscribe to
	// them). One instance for the whole process.
	broker := sse.NewBroker()

	// Worker pool
	pool := worker.NewPool(urlStore, cfg, broker)
	wg := pool.Start(jobCh)

	// Producer
	prodCtx, prodCancel := context.WithCancel(context.Background())
	producer := worker.NewProducer(urlStore, jobCh)
	go producer.Start(prodCtx)

	// Start API Server
	srv := api.NewServer(cfg, healthStore, userStore, notificationChannelStore, urlStore, broker)
	go srv.Start()
	slog.Info("Server and background workers started successfully")

	// Graceful shutdown listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("Shutdown signal received. Starting graceful teardown...")

	// 1. Tell the producer to stop fetching from the DB
	prodCancel()

	// 2. Stop the HTTP server from accepting new traffic
	slog.Info("Shutting down API server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.ShutDown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	// 3. Close the channel.
	// This tells workers: "Finish what you're doing, no more work is coming."
	slog.Info("Closing job channel and draining remaining workers...")
	close(jobCh)

	// 4. Wait for all active worker goroutines to finish their current job
	wg.Wait()

	slog.Info("All components stopped cleanly. Exiting.")
}
