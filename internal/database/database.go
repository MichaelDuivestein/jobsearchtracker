package database

import (
	"database/sql"
	"fmt"
	"jobsearchtracker/internal/config"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type Database interface {
	Connect(config *config.Config) (*sql.DB, error)
}

type FileDatabase struct{}

func NewFileDatabase() Database {
	return &FileDatabase{}
}

func (database *FileDatabase) Connect(config *config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", database.buildAndEnsureFilePath(config))
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to SQLite file database.")
	return db, nil
}

func (database *FileDatabase) buildAndEnsureFilePath(config *config.Config) string {
	var dbFilePath string
	if config.IsDatabaseFileLocationAbsolutePath {
		dbFilePath = config.DatabaseFilePath
	} else {
		absoluteFileLocation, err := filepath.Abs(config.DatabaseFilePath)
		if err != nil {
			slog.Error("Error getting database path", "error", err)
			os.Exit(1)
		}
		dbFilePath = absoluteFileLocation
	}
	if !strings.HasSuffix("/", dbFilePath) && !strings.HasSuffix(dbFilePath, "\\") {
		dbFilePath += string(os.PathSeparator)
	}

	if _, err := os.Stat(dbFilePath); os.IsNotExist(err) {
		err = os.MkdirAll(dbFilePath, os.ModePerm)
		if err != nil {
			slog.Error("Error ensuring database path exists", "error", err)
			os.Exit(1)
		}
	}

	dbFilePath += config.DatabaseFileName
	return dbFilePath
}

type InMemoryDatabase struct{}

func NewInMemoryDatabase() Database {
	return &InMemoryDatabase{}
}

func (database *InMemoryDatabase) Connect(_ *config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	slog.Info("Connected to SQLite in-memory database.")

	return db, nil
}
