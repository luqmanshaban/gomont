package worker

import (
	"log/slog"
	"sync"

	"github.com/luqmanshaban/gomont/internals/core"
	"github.com/luqmanshaban/gomont/internals/store"
)

type Pool struct {
	Store *store.URLStore
}

func NewPool(s *store.URLStore) *Pool {
	return &Pool{Store: s}
}

func (p *Pool) Start(jobCh <- chan core.URL) *sync.WaitGroup {
	slog.Info("pool started...")
	var wg sync.WaitGroup 
	
	w := &Worker{Store: p.Store};
	
	for i := range 4 {
		go func(id int){
			wg.Add(1)
			defer wg.Done()
			w.Worker(id, jobCh)
		}(i)
	}

	return &wg
}
