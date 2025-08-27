package main

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/dig"
	"jobsearchtracker/internal/api"
	configPackage "jobsearchtracker/internal/config"
	databasePackage "jobsearchtracker/internal/database"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	container, err := setupContainer()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = container.Invoke(func(database *sql.DB, config *configPackage.Config) error {
		return databasePackage.RunMigrations(database, config)
	})
	if err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 1)
	go func() {
		errChan <- container.Invoke(startServer)
	}()

	// Wait for interruption
	select {
	case err = <-errChan:
		// Error when starting HTTP server.
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		// Stop receiving signal notifications as soon as possible.
		slog.Info("Shutting down gracefully...")
		stop()
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

func startServer(server *api.Server, config *configPackage.Config) {

	log.Printf("Server starting on port %d", config.ServerPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.ServerPort), server))
}
