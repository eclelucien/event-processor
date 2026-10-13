package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed files/*.sql
var migrationFiles embed.FS

func Run(databaseURL string) error {
	database, error := sql.Open("pgx", databaseURL)
	if error != nil {
		return fmt.Errorf("failed to open database for migrations: %w", error)
	}

	defer database.Close()

	if error := database.Ping(); error != nil {
		return fmt.Errorf("failed to ping database for migrations: %w", error)
	}

	migrationFS, error := fs.Sub(migrationFiles, "files")
	if error != nil {
		return fmt.Errorf("failed to create migration filesystem: %w", error)
	}

	driver, error := postgres.WithInstance(
		database,
		&postgres.Config{},
	)
	if error != nil {
		return fmt.Errorf("failed to create postgres migration driver: %w", error)
	}

	sourceDriver, error := iofs.New(migrationFS, ".")
	if error != nil {
		return fmt.Errorf("failed to create migration source: %w", error)
	}

	migration, error := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"postgres",
		driver,
	)
	if error != nil {
		return fmt.Errorf("failed to initialize migrations: %w", error)
	}

	defer func() {
		sourceError, databaseError := migration.Close()

		if sourceError != nil {
			fmt.Printf("migration source close error: %v\n", sourceError)
		}

		if databaseError != nil {
			fmt.Printf("migration database close error: %v\n", databaseError)
		}
	}()

	if error := migration.Up(); error != nil && !errors.Is(error, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", error)
	}

	return nil
}
