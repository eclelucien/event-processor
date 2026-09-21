package http

import (
	"encoding/json"
	"net/http"

	"github.com/eclesiaste/event-processor/internal/application"
	"github.com/eclesiaste/event-processor/internal/domain"
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

type eventResponse struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Source    string          `json:"source"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt string          `json:"created_at"`
}

func newEventResponse(event *domain.Event) eventResponse {
	return eventResponse{
		ID:        event.ID,
		Type:      event.Type,
		Source:    event.Source,
		Payload:   json.RawMessage(event.Payload),
		CreatedAt: event.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
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

	response := newEventResponse(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(response)
}

func (h *EventHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	event, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}

	response := newEventResponse(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}
