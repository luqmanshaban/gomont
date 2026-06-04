package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/gomont/internals/api/handlers"
	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/store"
)

type Server struct {
	cfg *config.Config
	httpServer *http.Server
}

func NewServer(cfg *config.Config, hs *store.HealthStore, us *store.UserStore) *Server {
	mux := http.NewServeMux()

	// initiating handlers
	healthH := &handlers.HealthHandler{Store: hs}
	userH := &handlers.UserHandler{Store: us, Cfg: cfg}

	mux.HandleFunc("GET /health", healthH.CheckDBStatus)

	// users
	mux.HandleFunc("POST /auth",userH.CreateUser)

	s := &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr: fmt.Sprintf(":%d",cfg.PORT),
			Handler: mux,
		},
	}
	return s
}

func (s *Server) Start()  {
	slog.Info("SERVER RUNNING... ", "port", s.cfg.PORT)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("failed to run server", "error", err)
	}
}

func (s *Server) ShutDown(ctx context.Context) error {
	return  s.httpServer.Shutdown(ctx)
}