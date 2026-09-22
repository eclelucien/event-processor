package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eclesiaste/event-processor/internal/application"
	"github.com/eclesiaste/event-processor/internal/domain"
)

type EventService interface {
	Create(
		ctx context.Context,
		eventType string,
		source string,
		payload []byte,
	) (*domain.Event, error)

	GetByID(
		ctx context.Context,
		id string,
	) (*domain.Event, error)
}

type EventHandler struct {
	service EventService
}

func NewEventHandler(service EventService) *EventHandler {
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
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := h.service.Create(
		r.Context(),
		request.Type,
		request.Source,
		request.Payload,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
		if errors.Is(err, application.ErrEventNotFound) {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return

	}

	response := newEventResponse(event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}
