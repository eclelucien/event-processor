package kafka

import (
	"strings"
	"testing"
	"time"

	"github.com/eclesiaste/event-processor/internal/domain"
)

func TestEncodeEvent_KeepsPayloadAsJSON(t *testing.T) {
	event := &domain.Event{
		ID:        "event-123",
		Type:      "payment.created",
		Source:    "revofin",
		Payload:   []byte(`{"payment_id":"123"}`),
		CreatedAt: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	}

	encoded, error := encodeEvent(event)
	if error != nil {
		t.Fatalf("expected no error, got %v", error)
	}

	if !strings.Contains(string(encoded), `"payload":{"payment_id":"123"}`) {
		t.Fatalf("expected raw JSON payload, got %s", encoded)
	}
}

func TestDecodeEvent(t *testing.T) {
	event := &domain.Event{
		ID:        "event-123",
		Type:      "payment.created",
		Source:    "revofin",
		Payload:   []byte(`{"payment_id":"123"}`),
		CreatedAt: time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC),
	}

	encoded, error := encodeEvent(event)
	if error != nil {
		t.Fatalf("expected no error, got %v", error)
	}

	decoded, error := decodeEvent(encoded)
	if error != nil {
		t.Fatalf("expected no error, got %v", error)
	}

	if decoded.ID != event.ID {
		t.Fatalf("expected ID %s, got %s", event.ID, decoded.ID)
	}

	if decoded.Type != event.Type {
		t.Fatalf("expected type %s, got %s", event.Type, decoded.Type)
	}

	if decoded.Source != event.Source {
		t.Fatalf("expected source %s, got %s", event.Source, decoded.Source)
	}

	if string(decoded.Payload) != string(event.Payload) {
		t.Fatalf("expected payload %s, got %s", event.Payload, decoded.Payload)
	}

	if !decoded.CreatedAt.Equal(event.CreatedAt) {
		t.Fatalf("expected created_at %s, got %s", event.CreatedAt, decoded.CreatedAt)
	}
}

func TestDecodeEvent_InvalidJSON(t *testing.T) {
	_, error := decodeEvent([]byte(`{`))
	if error == nil {
		t.Fatal("expected error")
	}
}

func TestConsumer_Process(t *testing.T) {
	consumer := &Consumer{}

	error := consumer.process([]byte(`{"id":"event-123","type":"payment.created","source":"revofin","payload":{"payment_id":"123"},"created_at":"2026-10-14T12:00:00Z"}`))
	if error != nil {
		t.Fatalf("expected no error, got %v", error)
	}
}

func TestConsumer_Process_InvalidJSON(t *testing.T) {
	consumer := &Consumer{}

	if error := consumer.process([]byte(`{`)); error == nil {
		t.Fatal("expected error")
	}
}
