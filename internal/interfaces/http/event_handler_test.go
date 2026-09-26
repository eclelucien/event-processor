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
	listFunc    func(ctx context.Context, eventType string, source string, limit int, offset int) ([]domain.Event, error)
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

func (m *mockEventService) List(
	ctx context.Context,
	eventType string,
	source string,
	limit int,
	offset int,
) ([]domain.Event, error) {
	return m.listFunc(ctx, eventType, source, limit, offset)
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

func TestEventHandler_Create_EmptyType(t *testing.T) {
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

	body := `{
		"type": "",
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

	if !strings.Contains(
		recorder.Body.String(),
		`"error":"type is required"`,
	) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestEventHandler_Create_EmptySource(t *testing.T) {
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

	body := `{
		"type": "payment.created",
		"source": "",
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

	if !strings.Contains(
		recorder.Body.String(),
		`"error":"source is required"`,
	) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestEventHandler_Create_EmptyPayload(t *testing.T) {
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

	body := `{
		"type": "payment.created",
		"source": "revofin"
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

	if !strings.Contains(
		recorder.Body.String(),
		`"error":"payload is required"`,
	) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestEventHandler_List(t *testing.T) {
	service := &mockEventService{
		listFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			limit int,
			offset int,
		) ([]domain.Event, error) {
			if eventType != "payment.created" {
				t.Fatalf("unexpected type: %s", eventType)
			}

			if source != "revofin" {
				t.Fatalf("unexpected source: %s", source)
			}

			if limit != 10 {
				t.Fatalf("unexpected limit: %d", limit)
			}

			if offset != 5 {
				t.Fatalf("unexpected offset: %d", offset)
			}

			return []domain.Event{
				{
					ID:        "event-123",
					Type:      "payment.created",
					Source:    "revofin",
					Payload:   []byte(`{"payment_id":"123"}`),
					CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/events?type=payment.created&source=revofin&limit=10&offset=5",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

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

func TestEventHandler_List_InvalidLimit(t *testing.T) {
	service := &mockEventService{
		listFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			limit int,
			offset int,
		) ([]domain.Event, error) {
			t.Fatal("service should not be called")
			return nil, nil
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/events?limit=abc",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if !strings.Contains(recorder.Body.String(), `"error":"invalid limit"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestEventHandler_List_LimitOutOfRange(t *testing.T) {
	service := &mockEventService{
		listFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			limit int,
			offset int,
		) ([]domain.Event, error) {
			return nil, application.ErrInvalidListLimit
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/events?limit=101",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestEventHandler_List_InternalError(t *testing.T) {
	service := &mockEventService{
		listFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			limit int,
			offset int,
		) ([]domain.Event, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewEventHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/events",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.List(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
}
