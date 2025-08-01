package main

import (
	"fmt"
	"go.uber.org/dig"
	"jobsearchtracker/internal/api"
	configPackage "jobsearchtracker/internal/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {

	container, err := setupContainer()
	if err != nil {
		log.Fatal(err)
	}

	err = container.Invoke(startServer)
	if err != nil {
		log.Fatal("Failed to start server", err)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	go func() {
		err = container.Invoke(startServer)
		if err != nil {
			log.Fatal("Failed to start server", err)
		}
	}()

	<-signalChannel
	log.Println("Shutting down gracefully...")
}

func setupContainer() (*dig.Container, error) {
	container := dig.New()

	config, err := configPackage.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err = container.Provide(func() *configPackage.Config { return config }); err != nil {
		return nil, fmt.Errorf("failed to provide config: %w", err)
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
