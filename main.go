package main

import (
	"context"
	"database/sql"
	"fmt"
	"jobsearchtracker/internal/api"
	configPackage "jobsearchtracker/internal/config"
	databasePackage "jobsearchtracker/internal/database"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/dig"
)

func run() error {
	// Handle SIGINT (CTRL+C) and SIGTERM gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	container, err := setupContainer()
	if err != nil {
		return fmt.Errorf("failed to setup container: %w", err)
	}

	err = container.Invoke(func(database *sql.DB, config *configPackage.Config) error {
		return databasePackage.RunMigrations(database, config)
	})
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- container.Invoke(startServer)
	}()

	// Wait for interruption
	select {
	case err = <-errChan:
		// Error when starting HTTP server.
		return fmt.Errorf("failed to start server: %w", err)
	case <-ctx.Done():
		// Stop receiving signal notifications as soon as possible.
		slog.Info("Shutting down gracefully...")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("application failed to run", "error", err)
		os.Exit(1)
	}
}

func setupContainer() (*dig.Container, error) {
	container := dig.New()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := container.Provide(func() *slog.Logger { return logger }); err != nil {
		return nil, fmt.Errorf("failed to provide logger: %w", err)
	}

	config, err := configPackage.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err = container.Provide(func() *configPackage.Config { return config }); err != nil {
		return nil, fmt.Errorf("failed to provide config: %w", err)
	}

	if err = container.Provide(databasePackage.NewFileDatabase); err != nil {
		return nil, fmt.Errorf("failed to provide file database: %w", err)
	}

	err = container.Provide(func(db databasePackage.Database, config *configPackage.Config) (*sql.DB, error) {
		return db.Connect(config)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to provide database: %w", err)
	}

	if err := container.Provide(api.NewServer); err != nil {
		return nil, fmt.Errorf("failed to provide api server: %w", err)
	}

	return container, nil
}

func startServer(server *api.Server, config *configPackage.Config) error {
	slog.Info("Starting server...", "port", config.ServerPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), server)
}
