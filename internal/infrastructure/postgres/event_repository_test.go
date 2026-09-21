package postgres

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/eclesiaste/event-processor/internal/domain"
	"github.com/google/uuid"
)

func TestEventRepository_CreateAndGetByID(t *testing.T) {
	ctx := context.Background()

	databaseURL := "postgres://event_processor:event_processor@localhost:5432/event_processor?sslmode=disable"

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
