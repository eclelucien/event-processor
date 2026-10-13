package postgres

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/eclesiaste/event-processor/internal/domain"
	"github.com/google/uuid"
)

func TestEventRepository_CreateAndGetByID(t *testing.T) {
	requestContext := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not set")
	}
	database, error := New(requestContext, databaseURL)
	if error != nil {
		t.Fatalf("failed to connect to database: %v", error)
	}
	defer database.Close()

	repository := NewEventRepository(database)

	event := &domain.Event{
		ID:        uuid.NewString(),
		Type:      "payment.created",
		Source:    "revofin",
		Payload:   []byte(`{"payment_id":"integration-test","amount":100}`),
		CreatedAt: time.Now().UTC(),
	}

	error = repository.Create(requestContext, event)
	if error != nil {
		t.Fatalf("failed to create event: %v", error)
	}

	result, error := repository.GetByID(requestContext, event.ID)
	if error != nil {
		t.Fatalf("failed to get event: %v", error)
	}

	if result.ID != event.ID {
		t.Fatalf("expected ID %s, got %s", event.ID, result.ID)
	}

	if result.Type != event.Type {
		t.Fatalf("expected type %s, got %s", event.Type, result.Type)
	}

	if result.Source != event.Source {
		t.Fatalf("expected source %s, got %s", event.Source, result.Source)
	}

	var expectedPayload map[string]any
	var actualPayload map[string]any

	if error := json.Unmarshal(event.Payload, &expectedPayload); error != nil {
		t.Fatalf("failed to decode expected payload: %v", error)
	}

	if error := json.Unmarshal(result.Payload, &actualPayload); error != nil {
		t.Fatalf("failed to decode actual payload: %v", error)
	}

	if !reflect.DeepEqual(expectedPayload, actualPayload) {
		t.Fatalf(
			"expected payload %s, got %s",
			event.Payload,
			result.Payload,
		)
	}
}

func TestEventRepository_List(t *testing.T) {
	requestContext := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not set")
	}

	database, error := New(requestContext, databaseURL)
	if error != nil {
		t.Fatalf("failed to connect to database: %v", error)
	}
	defer database.Close()

	repository := NewEventRepository(database)

	source := "list-test-" + uuid.NewString()
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)

	first := &domain.Event{
		ID:        uuid.NewString(),
		Type:      "payment.created",
		Source:    source,
		Payload:   []byte(`{"payment_id":"first"}`),
		CreatedAt: older,
	}
	second := &domain.Event{
		ID:        uuid.NewString(),
		Type:      "payment.created",
		Source:    source,
		Payload:   []byte(`{"payment_id":"second"}`),
		CreatedAt: newer,
	}
	otherType := &domain.Event{
		ID:        uuid.NewString(),
		Type:      "payment.failed",
		Source:    source,
		Payload:   []byte(`{"payment_id":"other"}`),
		CreatedAt: newer.Add(time.Hour),
	}

	for _, event := range []*domain.Event{first, second, otherType} {
		if error := repository.Create(requestContext, event); error != nil {
			t.Fatalf("failed to create event: %v", error)
		}
	}

	events, error := repository.List(requestContext, "payment.created", source, 10, 0)
	if error != nil {
		t.Fatalf("failed to list events: %v", error)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].ID != second.ID {
		t.Fatalf("expected newest event first, got %s", events[0].ID)
	}

	if events[1].ID != first.ID {
		t.Fatalf("expected oldest event second, got %s", events[1].ID)
	}

	page, error := repository.List(requestContext, "", source, 1, 1)
	if error != nil {
		t.Fatalf("failed to list events page: %v", error)
	}

	if len(page) != 1 {
		t.Fatalf("expected 1 event, got %d", len(page))
	}

	if page[0].ID != second.ID {
		t.Fatalf("expected second page to start at %s, got %s", second.ID, page[0].ID)
	}
}
