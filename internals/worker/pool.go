package worker

import (
	"log/slog"
	"sync"

	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/core"
	"github.com/luqmanshaban/gomont/internals/sse"
	"github.com/luqmanshaban/gomont/internals/store"
)

type Pool struct {
	URLStore  *store.URLStore
	NotificationChannelStore *store.NotificationChannelStore
	Cfg    *config.Config
	Broker *sse.Broker
}

func NewPool(s *store.URLStore, ns *store.NotificationChannelStore, cfg *config.Config, broker *sse.Broker) *Pool {
	return &Pool{URLStore: s, NotificationChannelStore: ns, Cfg: cfg, Broker: broker}
}

func (p *Pool) Start(jobCh <-chan core.URL) *sync.WaitGroup {
	slog.Info("pool started...")
	var wg sync.WaitGroup

	w := &Worker{URLStore: p.URLStore,NotificationChannelStore: p.NotificationChannelStore, Cfg: p.Cfg, Broker: p.Broker}

	for i := range 4 {
		go func(id int) {
			wg.Add(1)
			defer wg.Done()
			w.Worker(id, jobCh)
		}(i)
	}

	return &wg
}