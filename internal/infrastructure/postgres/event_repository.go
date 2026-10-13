package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/eclesiaste/event-processor/internal/domain"
)

type EventRepository struct {
	database *sql.DB
}

func NewEventRepository(database *sql.DB) *EventRepository {
	return &EventRepository{
		database: database,
	}
}

func (r *EventRepository) Create(
	requestContext context.Context,
	event *domain.Event,
) error {
	_, error := r.database.ExecContext(
		requestContext,
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

	return error
}

func (r *EventRepository) GetByID(
	requestContext context.Context,
	id string,
) (*domain.Event, error) {
	event := &domain.Event{}

	error := r.database.QueryRowContext(
		requestContext,
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

	if error != nil {
		return nil, error
	}

	return event, nil
}

func (r *EventRepository) List(
	requestContext context.Context,
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

	rows, error := r.database.QueryContext(requestContext, query, args...)
	if error != nil {
		return nil, error
	}
	defer rows.Close()

	events := make([]domain.Event, 0)

	for rows.Next() {
		var event domain.Event

		if error := rows.Scan(
			&event.ID,
			&event.Type,
			&event.Source,
			&event.Payload,
			&event.CreatedAt,
		); error != nil {
			return nil, error
		}

		events = append(events, event)
	}

	if error := rows.Err(); error != nil {
		return nil, error
	}

	return events, nil
}
