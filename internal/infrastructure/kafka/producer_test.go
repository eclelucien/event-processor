package kafka

import "testing"

func TestNewProducer_Validation(t *testing.T) {
	if _, error := NewProducer("", "events"); error == nil {
		t.Fatal("expected error for empty brokers")
	}

	if _, error := NewProducer("localhost:9092", ""); error == nil {
		t.Fatal("expected error for empty topic")
	}

	producer, error := NewProducer("localhost:9092, localhost:9093", "events")
	if error != nil {
		t.Fatalf("expected producer, got %v", error)
	}

	if error := producer.Close(); error != nil {
		t.Fatalf("close: %v", error)
	}
}

func TestNewConsumer_Validation(t *testing.T) {
	if _, error := NewConsumer("localhost:9092", "events", ""); error == nil {
		t.Fatal("expected error for empty group id")
	}

	consumer, error := NewConsumer("localhost:9092", "events", "event-processor")
	if error != nil {
		t.Fatalf("expected consumer, got %v", error)
	}

	if error := consumer.Close(); error != nil {
		t.Fatalf("close: %v", error)
	}
}
