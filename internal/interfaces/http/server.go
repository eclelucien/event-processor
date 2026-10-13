package http

import (
	"context"
	"net/http"

	"github.com/eclesiaste/event-processor/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(configurations config.Config, eventHandler *EventHandler) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("POST /api/v1/events", eventHandler.Create)
	mux.HandleFunc("GET /api/v1/events", eventHandler.List)
	mux.HandleFunc("GET /api/v1/events/{id}", eventHandler.GetByID)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + configurations.HTTPPort,
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) Shutdown(requestContext context.Context) error {
	return s.httpServer.Shutdown(requestContext)
}
