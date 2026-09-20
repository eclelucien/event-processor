package http

import (
	"encoding/json"
	"net/http"

	"github.com/eclesiaste/event-processor/internal/application"
)

type EventHandler struct {
	service *application.EventService
}

func NewEventHandler(service *application.EventService) *EventHandler {
	return &EventHandler{
		service: service,
	}
}

type createEventRequest struct {
	Type    string          `json:"type"`
	Source  string          `json:"source"`
	Payload json.RawMessage `json:"payload"`
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createEventRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.service.Create(
		r.Context(),
		request.Type,
		request.Source,
		request.Payload,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(event)
}
