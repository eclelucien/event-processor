package application

import (
	"context"
	"errors"
	"testing"

	"github.com/eclesiaste/event-processor/internal/domain"
)

type mockEventRepository struct {
	createFunc  func(ctx context.Context, event *domain.Event) error
	getByIDFunc func(ctx context.Context, id string) (*domain.Event, error)
	listFunc    func(ctx context.Context, eventType string, source string, limit int, offset int) ([]domain.Event, error)
}

func (m *mockEventRepository) Create(
	ctx context.Context,
	event *domain.Event,
) error {
	return m.createFunc(ctx, event)
}

func (m *mockEventRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Event, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockEventRepository) List(
	ctx context.Context,
	eventType string,
	source string,
	limit int,
	offset int,
) ([]domain.Event, error) {
	return m.listFunc(ctx, eventType, source, limit, offset)
}

func TestEventService_Create(t *testing.T) {
	repository := &mockEventRepository{
		createFunc: func(ctx context.Context, event *domain.Event) error {
			return nil
		},
	}

	service := NewEventService(repository)

	event, err := service.Create(
		context.Background(),
		"payment.created",
		"revofin",
		[]byte(`{"payment_id":"123"}`),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.ID == "" {
		t.Fatal("expected event ID to be generated")
	}

	if event.Type != "payment.created" {
		t.Fatalf("expected type payment.created, got %s", event.Type)
	}

	if event.Source != "revofin" {
		t.Fatalf("expected source revofin, got %s", event.Source)
	}

	if string(event.Payload) != `{"payment_id":"123"}` {
		t.Fatalf("unexpected payload: %s", event.Payload)
	}

	if event.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestEventService_Create_InvalidType(t *testing.T) {
	repository := &mockEventRepository{
		createFunc: func(ctx context.Context, event *domain.Event) error {
			return nil
		},
	}

	service := NewEventService(repository)

	_, err := service.Create(
		context.Background(),
		"",
		"revofin",
		[]byte(`{"payment_id":"123"}`),
	)

	if !errors.Is(err, ErrInvalidEventType) {
		t.Fatalf("expected ErrInvalidEventType, got %v", err)
	}
}

func TestEventService_Create_InvalidSource(t *testing.T) {
	repository := &mockEventRepository{
		createFunc: func(ctx context.Context, event *domain.Event) error {
			return nil
		},
	}

	service := NewEventService(repository)

	_, err := service.Create(
		context.Background(),
		"payment.created",
		"",
		[]byte(`{"payment_id":"123"}`),
	)

	if !errors.Is(err, ErrInvalidEventSource) {
		t.Fatalf("expected ErrInvalidEventSource, got %v", err)
	}
}

func TestEventService_Create_EmptyPayload(t *testing.T) {
	repository := &mockEventRepository{
		createFunc: func(ctx context.Context, event *domain.Event) error {
			return nil
		},
	}

	service := NewEventService(repository)

	_, err := service.Create(
		context.Background(),
		"payment.created",
		"revofin",
		nil,
	)

	if !errors.Is(err, ErrEmptyPayload) {
		t.Fatalf("expected ErrEmptyPayload, got %v", err)
	}
}

func TestEventService_Create_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repository := &mockEventRepository{
		createFunc: func(ctx context.Context, event *domain.Event) error {
			return expectedErr
		},
	}

	service := NewEventService(repository)

	_, err := service.Create(
		context.Background(),
		"payment.created",
		"revofin",
		[]byte(`{"payment_id":"123"}`),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestEventService_GetByID(t *testing.T) {
	expectedEvent := &domain.Event{
		ID:      "event-123",
		Type:    "payment.created",
		Source:  "revofin",
		Payload: []byte(`{"payment_id":"123"}`),
	}

	repository := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Event, error) {
			if id != "event-123" {
				t.Fatalf("unexpected ID: %s", id)
			}

			return expectedEvent, nil
		},
	}

	service := NewEventService(repository)

	event, err := service.GetByID(
		context.Background(),
		"event-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.ID != expectedEvent.ID {
		t.Fatalf("expected ID %s, got %s", expectedEvent.ID, event.ID)
	}

	if event.Type != expectedEvent.Type {
		t.Fatalf("expected type %s, got %s", expectedEvent.Type, event.Type)
	}
}

func TestEventService_GetByID_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repository := &mockEventRepository{
		getByIDFunc: func(ctx context.Context, id string) (*domain.Event, error) {
			return nil, expectedErr
		},
	}

	service := NewEventService(repository)

	_, err := service.GetByID(
		context.Background(),
		"event-123",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestEventService_List_DefaultLimit(t *testing.T) {
	repository := &mockEventRepository{
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

			if limit != 20 {
				t.Fatalf("expected default limit 20, got %d", limit)
			}

			if offset != 0 {
				t.Fatalf("expected offset 0, got %d", offset)
			}

			return []domain.Event{
				{
					ID:     "event-123",
					Type:   eventType,
					Source: source,
				},
			}, nil
		},
	}

	service := NewEventService(repository)

	events, err := service.List(
		context.Background(),
		"payment.created",
		"revofin",
		0,
		0,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].ID != "event-123" {
		t.Fatalf("expected ID event-123, got %s", events[0].ID)
	}
}

func TestEventService_List_InvalidLimit(t *testing.T) {
	repository := &mockEventRepository{
		listFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			limit int,
			offset int,
		) ([]domain.Event, error) {
			t.Fatal("repository should not be called")
			return nil, nil
		},
	}

	service := NewEventService(repository)

	_, err := service.List(
		context.Background(),
		"",
		"",
		101,
		0,
	)

	if !errors.Is(err, ErrInvalidListLimit) {
		t.Fatalf("expected ErrInvalidListLimit, got %v", err)
	}
}

func TestEventService_List_InvalidOffset(t *testing.T) {
	repository := &mockEventRepository{
		listFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			limit int,
			offset int,
		) ([]domain.Event, error) {
			t.Fatal("repository should not be called")
			return nil, nil
		},
	}

	service := NewEventService(repository)

	_, err := service.List(
		context.Background(),
		"",
		"",
		10,
		-1,
	)

	if !errors.Is(err, ErrInvalidListOffset) {
		t.Fatalf("expected ErrInvalidListOffset, got %v", err)
	}
}

func TestEventService_List_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repository := &mockEventRepository{
		listFunc: func(
			ctx context.Context,
			eventType string,
			source string,
			limit int,
			offset int,
		) ([]domain.Event, error) {
			return nil, expectedErr
		},
	}

	service := NewEventService(repository)

	_, err := service.List(
		context.Background(),
		"",
		"",
		10,
		0,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
