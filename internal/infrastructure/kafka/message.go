package kafka

import (
	"encoding/json"
	"time"

	"github.com/eclesiaste/event-processor/internal/domain"
)

type eventMessage struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Source    string          `json:"source"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

func encodeEvent(event *domain.Event) ([]byte, error) {
	return json.Marshal(eventMessage{
		ID:        event.ID,
		Type:      event.Type,
		Source:    event.Source,
		Payload:   json.RawMessage(event.Payload),
		CreatedAt: event.CreatedAt,
	})
}

func decodeEvent(value []byte) (*domain.Event, error) {
	var message eventMessage

	if error := json.Unmarshal(value, &message); error != nil {
		return nil, error
	}

	return &domain.Event{
		ID:        message.ID,
		Type:      message.Type,
		Source:    message.Source,
		Payload:   []byte(message.Payload),
		CreatedAt: message.CreatedAt,
	}, nil
}
