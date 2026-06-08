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

func NewServer(cfg *config.Config, hs *store.HealthStore, us *store.UserStore, ns *store.NotificationStore) *Server {
	mux := http.NewServeMux()

	// initiating handlers
	healthH := &handlers.HealthHandler{Store: hs}
	userH := &handlers.UserHandler{Store: us, Cfg: cfg}
	notificationH := &handlers.NotificationChannelHandler{Store: ns}
	
	// auth middleware
	auth := handlers.AuthMiddleware(cfg.JWT_SECRET)

	mux.HandleFunc("GET /health", healthH.CheckDBStatus)

	// auth
	mux.HandleFunc("POST /auth",userH.CreateUser)
	mux.HandleFunc("POST /auth/login",userH.VerifyLogin)

	//users
	mux.Handle("GET /users", auth(http.HandlerFunc(userH.GetUser)))
	mux.Handle("PATCH /users", auth(http.HandlerFunc(userH.UpdateUserNames)))
	mux.Handle("DELETE /users", auth(http.HandlerFunc(userH.DeleteUserById)))

	// notification channels
	mux.Handle("GET /notifications/channels", auth(http.HandlerFunc(notificationH.GetNotificationChannelsByUserId)))
	mux.Handle("GET /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationH.GetNotificationChannelsById)))
	mux.Handle("POST /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationH.AddNotificationChannels)))
	mux.Handle("PATCH /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationH.UpdateNotificationChannels)))
	mux.Handle("DELETE /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationH.DeleteNotificationChannels)))

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