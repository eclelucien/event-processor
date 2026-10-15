package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/eclesiaste/event-processor/internal/domain"
)

type Producer struct {
	writer *kafkago.Writer
}

func NewProducer(brokers string, topic string) (*Producer, error) {
	addresses := splitBrokers(brokers)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	if topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	return &Producer{
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(addresses...),
			Topic:                  topic,
			Balancer:               &kafkago.LeastBytes{},
			RequiredAcks:           kafkago.RequireOne,
			AllowAutoTopicCreation: true,
			BatchSize:              1,
			BatchTimeout:           10 * time.Millisecond,
		},
	}, nil
}

func (p *Producer) Publish(requestContext context.Context, event *domain.Event) error {
	value, error := encodeEvent(event)
	if error != nil {
		return error
	}

	return p.writer.WriteMessages(requestContext, kafkago.Message{
		Key:   []byte(event.ID),
		Value: value,
		Time:  event.CreatedAt,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func splitBrokers(brokers string) []string {
	parts := strings.Split(brokers, ",")
	addresses := make([]string, 0, len(parts))

	for _, part := range parts {
		address := strings.TrimSpace(part)
		if address != "" {
			addresses = append(addresses, address)
		}
	}

	return addresses
}
