package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/gomont/internals/api/handlers"
	"github.com/luqmanshaban/gomont/internals/config"
	"github.com/luqmanshaban/gomont/internals/sse"
	"github.com/luqmanshaban/gomont/internals/store"
)

type Server struct {
	cfg        *config.Config
	httpServer *http.Server
}

func NewServer(
	cfg *config.Config,
	hs *store.HealthStore,
	us *store.UserStore,
	nchs *store.NotificationChannelStore,
	urlS *store.URLStore,
	broker *sse.Broker) *Server {

	mux := http.NewServeMux()

	// initiating handlers
	healthH := &handlers.HealthHandler{Store: hs}
	userH := &handlers.UserHandler{Store: us, Cfg: cfg}
	notificationChH := &handlers.NotificationChannelHandler{Store: nchs}
	urlH := &handlers.URLHandler{Store: urlS}
	eventsH := &handlers.EventsHandler{Broker: broker, URLStore: urlS, JWTSecret: cfg.JWT_SECRET}

	// static handlers
	pageH := handlers.NewPageHandler()

	// auth middleware
	auth := handlers.AuthMiddleware(cfg.JWT_SECRET)

	mux.HandleFunc("GET /health", healthH.CheckDBStatus)

	// static files
	mux.HandleFunc("GET /login", pageH.Login)
	mux.HandleFunc("GET /signup", pageH.Signup)
	mux.HandleFunc("GET /dashboard", pageH.Dashboard)
	mux.HandleFunc("GET /settings", pageH.Settings)
	mux.Handle("GET /static/", http.StripPrefix("/static/", pageH.Static()))

	// live updates (auth handled inside the handler via ?token=, since EventSource can't send an Authorization header — not wrapped in the auth middleware, which only checks that header)
	mux.HandleFunc("GET /events", eventsH.Stream)

	// auth
	mux.HandleFunc("POST /auth", userH.CreateUser)
	mux.HandleFunc("POST /auth/login", userH.VerifyLogin)

	//users
	mux.Handle("GET /users", auth(http.HandlerFunc(userH.GetUser)))
	mux.Handle("PATCH /users", auth(http.HandlerFunc(userH.UpdateUserNames)))
	mux.Handle("DELETE /users", auth(http.HandlerFunc(userH.DeleteUserById)))

	// notification channels
	mux.Handle("GET /notifications/channels", auth(http.HandlerFunc(notificationChH.GetNotificationChannelsByUserId)))
	mux.Handle("GET /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationChH.GetNotificationChannelsById)))
	mux.Handle("POST /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationChH.AddNotificationChannels)))
	mux.Handle("PATCH /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationChH.UpdateNotificationChannels)))
	mux.Handle("DELETE /notifications/channels/{row_id}", auth(http.HandlerFunc(notificationChH.DeleteNotificationChannels)))

	// urls
	mux.Handle("POST /urls", auth(http.HandlerFunc(urlH.AddNewEndpoint)))
	mux.Handle("GET /urls", auth(http.HandlerFunc(urlH.GetURLsByUserId)))
	mux.Handle("GET /urls/{url_id}", auth(http.HandlerFunc(urlH.GetURLById)))
	mux.Handle("PATCH /urls/{url_id}", auth(http.HandlerFunc(urlH.UpdateURL)))
	mux.Handle("DELETE /urls/{url_id}", auth(http.HandlerFunc(urlH.DeleteURL)))
	mux.Handle("POST /urls/{url_id}/retry", auth(http.HandlerFunc(urlH.ManualRetry)))

	s := &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.PORT),
			Handler: mux,
		},
	}
	return s
}

func (s *Server) Start() {
	slog.Info("SERVER RUNNING... ", "port", s.cfg.PORT)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("failed to run server", "error", err)
	}
}

func (s *Server) ShutDown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
