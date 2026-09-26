package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (r *EventRepository) List(
	ctx context.Context,
	eventType string,
	source string,
	limit int,
	offset int,
) ([]domain.Event, error) {
	args := make([]any, 0, 4)
	conditions := make([]string, 0, 2)

	if eventType != "" {
		args = append(args, eventType)
		conditions = append(conditions, fmt.Sprintf("type = $%d", len(args)))
	}

	if source != "" {
		args = append(args, source)
		conditions = append(conditions, fmt.Sprintf("source = $%d", len(args)))
	}

	query := `
		SELECT id, type, source, payload, created_at
		FROM events
	`

	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + "\n"
	}

	args = append(args, limit, offset)
	query += fmt.Sprintf(
		"ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		len(args)-1,
		len(args),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)

	for rows.Next() {
		var event domain.Event

		if err := rows.Scan(
			&event.ID,
			&event.Type,
			&event.Source,
			&event.Payload,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
