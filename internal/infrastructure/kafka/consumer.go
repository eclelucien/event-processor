package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafkago.Reader
	topic  string
}

func NewConsumer(brokers string, topic string, groupID string) (*Consumer, error) {
	addresses := splitBrokers(brokers)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	if topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	if groupID == "" {
		return nil, fmt.Errorf("kafka group id is required")
	}

	return &Consumer{
		topic: topic,
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  addresses,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6,
			MaxWait:  500 * time.Millisecond,
		}),
	}, nil
}

func (c *Consumer) Start(requestContext context.Context) error {
	log.Printf("Kafka consumer reading topic %s", c.topic)

	for {
		message, error := c.reader.FetchMessage(requestContext)
		if error != nil {
			if requestContext.Err() != nil {
				return nil
			}

			return error
		}

		if error := c.process(message.Value); error != nil {
			log.Printf(
				"kafka consumer: skipping invalid message at offset %d: %v",
				message.Offset,
				error,
			)
		}

		if error := c.reader.CommitMessages(requestContext, message); error != nil {
			if requestContext.Err() != nil {
				return nil
			}

			return error
		}
	}
}

func (c *Consumer) process(value []byte) error {
	event, error := decodeEvent(value)
	if error != nil {
		return error
	}

	log.Printf(
		"event processed id=%s type=%s source=%s payload=%s",
		event.ID,
		event.Type,
		event.Source,
		event.Payload,
	)

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
