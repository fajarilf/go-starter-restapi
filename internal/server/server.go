package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/handler"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	cfg         config.Config
	router      *chi.Mux
	httpServer  *http.Server
	roomHandler *handler.RoomHandler
}

func New(cfg config.Config, roomHandler *handler.RoomHandler) *Server {
	r := chi.NewRouter()

	s := &Server{
		cfg:         cfg,
		router:      r,
		roomHandler: roomHandler,
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
