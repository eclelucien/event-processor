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
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	List(ctx context.Context, eventType string, source string, limit int, offset int) ([]domain.Event, error)
}

type EventService struct {
	repository EventRepository
}

func NewEventService(repository EventRepository) *EventService {
	return &EventService{
		repository: repository,
	}
}

func (s *EventService) Create(
	ctx context.Context,
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

	if err := s.repository.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetByID(
	ctx context.Context,
	id string,
) (*domain.Event, error) {
	event, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}

		return nil, err
	}

	return event, nil
}

func (s *EventService) List(
	ctx context.Context,
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

	return s.repository.List(ctx, eventType, source, limit, offset)
}
