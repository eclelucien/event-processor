package application

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/eclesiaste/event-processor/internal/domain"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

var (
	ErrInvalidEventType   = errors.New("event type is required")
	ErrInvalidEventSource = errors.New("event source is required")
	ErrEmptyPayload       = errors.New("event payload is required")
	ErrEventNotFound      = errors.New("event not found")
	ErrInvalidListLimit   = errors.New("limit must be between 1 and 100")
	ErrInvalidListOffset  = errors.New("offset must be zero or greater")
)

type EventRepository interface {
	Create(requestContext context.Context, event *domain.Event) error
	GetByID(requestContext context.Context, id string) (*domain.Event, error)
	List(requestContext context.Context, eventType string, source string, limit int, offset int) ([]domain.Event, error)
}

type EventPublisher interface {
	Publish(requestContext context.Context, event *domain.Event) error
}

type EventService struct {
	repository EventRepository
	publisher  EventPublisher
}

func NewEventService(repository EventRepository, publisher EventPublisher) *EventService {
	return &EventService{
		repository: repository,
		publisher:  publisher,
	}
}

func (s *EventService) Create(
	requestContext context.Context,
	eventType string,
	source string,
	payload []byte,
) (*domain.Event, error) {
	if eventType == "" {
		return nil, ErrInvalidEventType
	}

	if source == "" {
		return nil, ErrInvalidEventSource
	}

	if len(payload) == 0 {
		return nil, ErrEmptyPayload
	}

	event := &domain.Event{
		ID:        uuid.NewString(),
		Type:      eventType,
		Source:    source,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	if error := s.repository.Create(requestContext, event); error != nil {
		return nil, error
	}

	if error := s.publisher.Publish(requestContext, event); error != nil {
		return nil, error
	}

	return event, nil
}

func (s *EventService) GetByID(
	requestContext context.Context,
	id string,
) (*domain.Event, error) {
	event, error := s.repository.GetByID(requestContext, id)
	if error != nil {
		if errors.Is(error, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}

		return nil, error
	}

	return event, nil
}

func (s *EventService) List(
	requestContext context.Context,
	eventType string,
	source string,
	limit int,
	offset int,
) ([]domain.Event, error) {
	if limit == 0 {
		limit = defaultListLimit
	}

	if limit < 1 || limit > maxListLimit {
		return nil, ErrInvalidListLimit
	}

	if offset < 0 {
		return nil, ErrInvalidListOffset
	}

	return s.repository.List(requestContext, eventType, source, limit, offset)
}
