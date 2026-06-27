package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/handler"
	"github.com/fajarilf/go-starter-api/internal/service"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	cfg         config.Config
	router      *chi.Mux
	httpServer  *http.Server
	roomHandler *handler.RoomHandler
	authHandler *handler.AuthHandler
	authService *service.AuthService
}

func New(cfg config.Config, roomHandler *handler.RoomHandler, authHandler *handler.AuthHandler, authService *service.AuthService) *Server {
	r := chi.NewRouter()

	s := &Server{
		cfg:         cfg,
		router:      r,
		roomHandler: roomHandler,
		authHandler: authHandler,
		authService: authService,
	}

	s.regiterMiddleware()
	s.registerRoutes()

	s.httpServer = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Handler returns the configured router, for integration tests.
func (s *Server) Handler() http.Handler { return s.router }

func (s *Server) Start(ctx context.Context) error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Server starting", "addr", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	slog.Info("Server stopped cleanly")
	return nil
}
