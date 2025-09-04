package dependencyinjection

import (
	"database/sql"
	apiV1 "jobsearchtracker/internal/api/v1/handlers"
	configPackage "jobsearchtracker/internal/config"
	databasePackage "jobsearchtracker/internal/database"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/services"
	"log"
	"log/slog"
	"os"
	"testing"

	"go.uber.org/dig"
)

func SetupDatabaseTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := dig.New()

	err := container.Provide(func() *configPackage.Config {
		return &config
	})
	if err != nil {
		slog.Error("Failed to provide test config", "error", err)
		os.Exit(1)
	}

	if err = container.Provide(databasePackage.NewInMemoryDatabase); err != nil {
		slog.Error("Failed to provide in-memory database", "error", err)
		os.Exit(1)
	}

	err = container.Provide(func(db databasePackage.Database, config *configPackage.Config) (*sql.DB, error) {
		return db.Connect(config)
	})
	if err != nil {
		slog.Error("Failed to provide database", "error", err)
		os.Exit(1)
	}

	err = container.Invoke(func(database *sql.DB, config *configPackage.Config) error {
		return databasePackage.RunMigrations(database, config)
	})
	if err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	t.Cleanup(func() {
		err = container.Invoke(func(database *sql.DB) error {
			t.Logf("Closing database connection for test : %s", t.Name())
			return database.Close()
		})
		if err != nil {
			t.Errorf("Failed to close database connection for test : %s", t.Name())
		}
	})

	return container
}

func SetupCompanyRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyRepository", err)
	}

	return container
}

func SetupCompanyServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupCompanyRepositoryTestContainer(t, config)

	err := container.Provide(func(companyRepository *repositories.CompanyRepository) *services.CompanyService {
		return services.NewCompanyService(companyRepository)
	})
	if err != nil {
		log.Fatal("Failed to provide companyService", err)
	}

	return container
}

func SetupCompanyHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupCompanyServiceTestContainer(t, config)

	err := container.Provide(func(companyService *services.CompanyService) *apiV1.CompanyHandler {
		return apiV1.NewCompanyHandler(companyService)
	})
	if err != nil {
		log.Fatal("Failed to provide companyHandler", err)
	}

	return container
}
