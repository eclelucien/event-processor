package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(requestContext context.Context, databaseURL string) (*sql.DB, error) {
	database, error := sql.Open("pgx", databaseURL)
	if error != nil {
		return nil, error
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(30 * time.Minute)
	database.SetConnMaxIdleTime(5 * time.Minute)

	if error := database.PingContext(requestContext); error != nil {
		return nil, error
	}

	return database, nil
}
