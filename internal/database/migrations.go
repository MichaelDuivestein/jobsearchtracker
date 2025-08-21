package database

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"jobsearchtracker/internal/config"
	"log/slog"
	"os"
	"path/filepath"
)

func RunMigrations(database *sql.DB, config *config.Config) error {
	slog.Info("Starting to run DB migrations")
	driver, err := sqlite.WithInstance(database, &sqlite.Config{})
	if err != nil {
		return err
	}

	var migrationsPath string
	if config.IsDatabaseMigrationsPathAbsolutePath {
		migrationsPath = config.DatabaseMigrationsPath
	} else {
		absoluteMigrationLocation, err := filepath.Abs(config.DatabaseMigrationsPath)
		if err != nil {
			slog.Error("Error getting migrations path", err)
			os.Exit(1)
		}
		migrationsPath = absoluteMigrationLocation
	}

	fullMigrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}

	migrations, err := migrate.NewWithDatabaseInstance("file://"+fullMigrationsPath, "sqlite", driver)

	if err != nil {
		return err
	}

	if err := migrations.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	slog.Info("DB migrations complete.")
	return nil
}
