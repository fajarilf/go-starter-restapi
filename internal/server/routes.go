package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) regiterMiddleware() {
	r := s.router

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(s.requestLogger)
}

func (s *Server) registerRoutes() {
	r := s.router

	r.Get("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api/rooms", func(r chi.Router) {
		r.Post("/", s.roomHandler.Create)
		r.Get("/{id}", s.roomHandler.GetById)
	})
}
