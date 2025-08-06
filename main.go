package main

import (
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

	err = container.Invoke(startServer)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		err = container.Invoke(startServer)
		if err != nil {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	<-signalChannel
	log.Println("Shutting down gracefully...")
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
