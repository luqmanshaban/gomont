package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/luqmanshaban/gomont/internals/core"
	"github.com/luqmanshaban/gomont/internals/store"
)

type Producer struct {
	Store *store.URLStore
	ch    chan<- core.URL
}

func NewProducer(s *store.URLStore, ch chan<- core.URL) *Producer {
	return &Producer{Store: s, ch: ch}
}

func (p *Producer) Start(ctx context.Context) {
	// Ticks every 500ms as requested
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Producer shutting down...")
			return
		case <-ticker.C:
			// Implement a FetchURLs or FetchPending method in your URLStore
			urls, err := p.Store.FetchURLsToPing() 
			if err != nil {
				slog.Error("failed to fetch urls from db", "error", err)
				continue
			}

			for _, url := range urls {
				select {
				case p.ch <- url:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}