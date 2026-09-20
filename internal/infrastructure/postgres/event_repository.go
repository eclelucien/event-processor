package postgres

import (
	"context"
	"database/sql"

	"github.com/eclesiaste/event-processor/internal/domain"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

func (r *EventRepository) Create(
	ctx context.Context,
	event *domain.Event,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO events (
			id,
			type,
			source,
			payload,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		event.ID,
		event.Type,
		event.Source,
		event.Payload,
		event.CreatedAt,
	)

	return err
}

func (r *EventRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Event, error) {
	event := &domain.Event{}

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT id, type, source, payload, created_at
		FROM events
		WHERE id = $1
		`,
		id,
	).Scan(
		&event.ID,
		&event.Type,
		&event.Source,
		&event.Payload,
		&event.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return event, nil
}
