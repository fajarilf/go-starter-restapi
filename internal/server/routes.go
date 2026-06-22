package server

import (
	"net/http"
	"time"

	"github.com/fajarilf/go-starter-api/docs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	httpSwagger "github.com/swaggo/http-swagger/v2"
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

	// Raw OpenAPI spec files (embedded at build time). Serving the whole
	// embedded FS means openapi.yaml and any file it $refs (room_docs.yaml)
	// are reachable for the resolver.
	specFiles := http.StripPrefix("/api/docs/", http.FileServer(http.FS(docs.FS)))
	r.Get("/api/docs/openapi.yaml", specFiles.ServeHTTP)
	r.Get("/api/docs/room_docs.yaml", specFiles.ServeHTTP)

	// Swagger UI, served from /api/docs/. The wildcard is required so the
	// handler can serve its own static assets.
	r.Get("/api/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/api/docs/openapi.yaml"),
	))

	r.Route("/api/rooms", func(r chi.Router) {
		r.Post("/", s.roomHandler.Create)
		r.Get("/", s.roomHandler.Get)
		r.Get("/{id}", s.roomHandler.GetById)
		r.Put("/{id}", s.roomHandler.Update)
		r.Delete("/{id}", s.roomHandler.Delete)
	})
}
