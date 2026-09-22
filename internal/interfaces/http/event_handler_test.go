package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eclesiaste/event-processor/internal/application"
	"github.com/eclesiaste/event-processor/internal/domain"
)

type mockEventService struct {
	createFunc  func(ctx context.Context, eventType, source string, payload []byte) (*domain.Event, error)
	getByIDFunc func(ctx context.Context, id string) (*domain.Event, error)
}

func (m *mockEventService) Create(
	ctx context.Context,
	eventType string,
	source string,
	payload []byte,
) (*domain.Event, error) {
	return m.createFunc(ctx, eventType, source, payload)
}

func (m *mockEventService) GetByID(
	ctx context.Context,
	id string,
) (*domain.Event, error) {
	return m.getByIDFunc(ctx, id)
}

func TestEventHandler_Create(t *testing.T) {
	service := &mockEventService{
		createFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			payload []byte,
		) (*domain.Event, error) {
			return &domain.Event{
				ID:        "event-123",
				Type:      eventType,
				Source:    source,
				Payload:   payload,
				CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	handler := NewEventHandler(service)

	body := `{
		"type": "payment.created",
		"source": "revofin",
		"payload": {
			"payment_id": "123",
			"amount": 150.00
		}
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/events",
		strings.NewReader(body),
	)

	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	responseBody := recorder.Body.String()

	if !strings.Contains(responseBody, `"event-123"`) {
		t.Fatalf("expected event ID in response, got %s", responseBody)
	}

	if !strings.Contains(responseBody, `"payment.created"`) {
		t.Fatalf("expected event type in response, got %s", responseBody)
	}

	if !strings.Contains(responseBody, `"payment_id":"123"`) {
		t.Fatalf("expected payload in response, got %s", responseBody)
	}
}

func TestEventHandler_Create_InvalidJSON(t *testing.T) {
	service := &mockEventService{
		createFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			payload []byte,
		) (*domain.Event, error) {
			t.Fatal("service should not be called")
			return nil, nil
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/events",
		strings.NewReader(`invalid json`),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestEventHandler_Create_ServiceError(t *testing.T) {
	serviceError := errors.New("service error")

	service := &mockEventService{
		createFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			payload []byte,
		) (*domain.Event, error) {
			return nil, serviceError
		},
	}

	handler := NewEventHandler(service)

	body := `{
		"type": "payment.created",
		"source": "revofin",
		"payload": {
			"payment_id": "123"
		}
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/events",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestEventHandler_GetByID(t *testing.T) {
	service := &mockEventService{
		getByIDFunc: func(
			ctx context.Context,
			id string,
		) (*domain.Event, error) {
			if id != "event-123" {
				t.Fatalf("unexpected ID: %s", id)
			}

			return &domain.Event{
				ID:        "event-123",
				Type:      "payment.created",
				Source:    "revofin",
				Payload:   []byte(`{"payment_id":"123"}`),
				CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/events/event-123",
		nil,
	)

	request.SetPathValue("id", "event-123")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	responseBody := recorder.Body.String()

	if !strings.Contains(responseBody, `"event-123"`) {
		t.Fatalf("expected event ID in response, got %s", responseBody)
	}
}

func TestEventHandler_GetByID_NotFound(t *testing.T) {
	service := &mockEventService{
		getByIDFunc: func(
			ctx context.Context,
			id string,
		) (*domain.Event, error) {
			return nil, application.ErrEventNotFound
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/events/event-123",
		nil,
	)

	request.SetPathValue("id", "event-123")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}

func TestEventHandler_GetByID_InternalError(t *testing.T) {
	service := &mockEventService{
		getByIDFunc: func(
			ctx context.Context,
			id string,
		) (*domain.Event, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/events/event-123",
		nil,
	)

	request.SetPathValue("id", "event-123")

	recorder := httptest.NewRecorder()

	handler.GetByID(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
}
