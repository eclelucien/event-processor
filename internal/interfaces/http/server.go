package http

import (
	"net/http"

	"github.com/eclesiaste/event-processor/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg config.Config, eventHandler *EventHandler) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("POST /api/v1/events", eventHandler.Create)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.HTTPPort,
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
