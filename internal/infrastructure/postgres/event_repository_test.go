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
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not set")
	}
	db, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repository := NewEventRepository(db)

	event := &domain.Event{
		ID:        uuid.NewString(),
		Type:      "payment.created",
		Source:    "revofin",
		Payload:   []byte(`{"payment_id":"integration-test","amount":100}`),
		CreatedAt: time.Now().UTC(),
	}

	err = repository.Create(ctx, event)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	result, err := repository.GetByID(ctx, event.ID)
	if err != nil {
		t.Fatalf("failed to get event: %v", err)
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

	if err := json.Unmarshal(event.Payload, &expectedPayload); err != nil {
		t.Fatalf("failed to decode expected payload: %v", err)
	}

	if err := json.Unmarshal(result.Payload, &actualPayload); err != nil {
		t.Fatalf("failed to decode actual payload: %v", err)
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
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is not set")
	}

	db, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repository := NewEventRepository(db)

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
		if err := repository.Create(ctx, event); err != nil {
			t.Fatalf("failed to create event: %v", err)
		}
	}

	events, err := repository.List(ctx, "payment.created", source, 10, 0)
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
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

	page, err := repository.List(ctx, "", source, 1, 1)
	if err != nil {
		t.Fatalf("failed to list events page: %v", err)
	}

	if len(page) != 1 {
		t.Fatalf("expected 1 event, got %d", len(page))
	}

	if page[0].ID != second.ID {
		t.Fatalf("expected second page to start at %s, got %s", second.ID, page[0].ID)
	}
}
