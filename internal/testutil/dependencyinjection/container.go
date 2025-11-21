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

// -------- Application containers: --------

func SetupApplicationRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.ApplicationRepository {
		return repositories.NewApplicationRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationRepository", err)
	}

	// the CompanyRepository is also needed due to a FK dependency.
	err = container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyRepository in SetupApplicationRepositoryTestContainer", err)
	}

	// Add EventRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.EventRepository {
		return repositories.NewEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide eventRepository in SetupApplicationRepositoryTestContainer", err)
	}

	// Add PersonRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.PersonRepository {
		return repositories.NewPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide personRepository", err)
	}

	// Add ApplicationEventRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.ApplicationEventRepository {
		return repositories.NewApplicationEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationEventRepository", err)
	}

	// Add ApplicationPersonRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.ApplicationPersonRepository {
		return repositories.NewApplicationPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationPersonRepository", err)
	}

	return container
}

func SetupApplicationServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupApplicationRepositoryTestContainer(t, config)

	err := container.Provide(func(applicationRepository *repositories.ApplicationRepository) *services.ApplicationService {
		return services.NewApplicationService(applicationRepository)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationService", err)
	}

	return container
}

func SetupApplicationHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupApplicationServiceTestContainer(t, config)

	err := container.Provide(func(applicationService *services.ApplicationService) *apiV1.ApplicationHandler {
		return apiV1.NewApplicationHandler(applicationService)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationHandler", err)
	}

	return container
}

// -------- ApplicationEvent containers: --------

func SetupApplicationEventRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.ApplicationEventRepository {
		return repositories.NewApplicationEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationEventRepository", err)
	}

	// Add EventRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.EventRepository {
		return repositories.NewEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide eventRepository", err)
	}

	// Add CompanyRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyRepository", err)
	}

	// Add ApplicationRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.ApplicationRepository {
		return repositories.NewApplicationRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationRepository", err)
	}

	return container
}

func SetupApplicationEventServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupApplicationEventRepositoryTestContainer(t, config)

	err := container.Provide(
		func(repository *repositories.ApplicationEventRepository) *services.ApplicationEventService {

			return services.NewApplicationEventService(repository)
		})
	if err != nil {
		log.Fatal("Failed to provide applicationEventService", err)
	}

	return container
}

func SetupApplicationEventHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupApplicationEventServiceTestContainer(t, config)

	err := container.Provide(func(service *services.ApplicationEventService) *apiV1.ApplicationEventHandler {
		return apiV1.NewApplicationEventHandler(service)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationEventHandler", err)
	}

	return container
}

// -------- ApplicationPerson containers: --------

func SetupApplicationPersonRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.ApplicationPersonRepository {
		return repositories.NewApplicationPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationPersonRepository", err)
	}

	// Add PersonRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.PersonRepository {
		return repositories.NewPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide personRepository", err)
	}

	// Add CompanyRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyRepository", err)
	}

	// Add ApplicationRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.ApplicationRepository {
		return repositories.NewApplicationRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationRepository", err)
	}

	return container
}

func SetupApplicationPersonServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupApplicationPersonRepositoryTestContainer(t, config)

	err := container.Provide(func(repository *repositories.ApplicationPersonRepository) *services.ApplicationPersonService {
		return services.NewApplicationPersonService(repository)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationPersonService", err)
	}

	return container
}

func SetupApplicationPersonHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupApplicationPersonServiceTestContainer(t, config)

	err := container.Provide(func(service *services.ApplicationPersonService) *apiV1.ApplicationPersonHandler {
		return apiV1.NewApplicationPersonHandler(service)
	})
	if err != nil {
		log.Fatal("Failed to provide applicationPersonHandler", err)
	}

	return container
}

// -------- Company containers: --------

func SetupCompanyRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.ApplicationRepository {
		return repositories.NewApplicationRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide ApplicationRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.EventRepository {
		return repositories.NewEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide EventRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.PersonRepository {
		return repositories.NewPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide PersonRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.CompanyEventRepository {
		return repositories.NewCompanyEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide CompanyEventRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.CompanyPersonRepository {
		return repositories.NewCompanyPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide CompanyPersonRepository", err)
	}

	return container
}

func SetupCompanyServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupCompanyRepositoryTestContainer(t, config)

	err := container.Provide(
		func(
			companyRepository *repositories.CompanyRepository) *services.CompanyService {
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

// -------- CompanyEvent containers: --------

func SetupCompanyEventRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.CompanyEventRepository {
		return repositories.NewCompanyEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyEventRepository", err)
	}

	// Add CompanyRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyRepository", err)
	}

	// Add EventRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.EventRepository {
		return repositories.NewEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide eventRepository", err)
	}

	return container
}

func SetupCompanyEventServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupCompanyEventRepositoryTestContainer(t, config)

	err := container.Provide(func(repository *repositories.CompanyEventRepository) *services.CompanyEventService {
		return services.NewCompanyEventService(repository)
	})
	if err != nil {
		log.Fatal("Failed to provide companyEventService", err)
	}

	return container
}

func SetupCompanyEventHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupCompanyEventServiceTestContainer(t, config)

	err := container.Provide(func(service *services.CompanyEventService) *apiV1.CompanyEventHandler {
		return apiV1.NewCompanyEventHandler(service)
	})
	if err != nil {
		log.Fatal("Failed to provide companyEventHandler", err)
	}

	return container
}

// -------- CompanyPerson containers: --------

func SetupCompanyPersonRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.CompanyPersonRepository {
		return repositories.NewCompanyPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyPersonRepository", err)
	}

	// Add PersonRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.PersonRepository {
		return repositories.NewPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide personRepository", err)
	}

	// Add CompanyRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide companyRepository", err)
	}

	return container
}

func SetupCompanyPersonServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupCompanyPersonRepositoryTestContainer(t, config)

	err := container.Provide(func(repository *repositories.CompanyPersonRepository) *services.CompanyPersonService {
		return services.NewCompanyPersonService(repository)
	})
	if err != nil {
		log.Fatal("Failed to provide companyPersonService", err)
	}

	return container
}

func SetupCompanyPersonHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupCompanyPersonServiceTestContainer(t, config)

	err := container.Provide(func(service *services.CompanyPersonService) *apiV1.CompanyPersonHandler {
		return apiV1.NewCompanyPersonHandler(service)
	})
	if err != nil {
		log.Fatal("Failed to provide companyPersonHandler", err)
	}

	return container
}

// -------- Event containers: --------

func SetupEventRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.EventRepository {
		return repositories.NewEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide eventRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.ApplicationRepository {
		return repositories.NewApplicationRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide ApplicationRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide CompanyRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.PersonRepository {
		return repositories.NewPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide PersonRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.ApplicationEventRepository {
		return repositories.NewApplicationEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide ApplicationEventRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.CompanyEventRepository {
		return repositories.NewCompanyEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide CompanyEventRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.EventPersonRepository {
		return repositories.NewEventPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide EventPersonRepository", err)
	}

	return container
}

func SetupEventServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupEventRepositoryTestContainer(t, config)

	err := container.Provide(func(repository *repositories.EventRepository) *services.EventService {
		return services.NewEventService(repository)
	})
	if err != nil {
		log.Fatal("Failed to provide eventService", err)
	}
	return container
}

func SetupEventHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupEventServiceTestContainer(t, config)

	err := container.Provide(func(service *services.EventService) *apiV1.EventHandler {
		return apiV1.NewEventHandler(service)
	})
	if err != nil {
		log.Fatal("Failed to provide eventHandler", err)
	}
	return container
}

// -------- EventPerson containers: --------

func SetupEventPersonRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.EventPersonRepository {
		return repositories.NewEventPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide eventPersonRepository", err)
	}

	// Add PersonRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.PersonRepository {
		return repositories.NewPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide personRepository", err)
	}

	// Add EventRepository in order to insert data for testing
	err = container.Provide(func(db *sql.DB) *repositories.EventRepository {
		return repositories.NewEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide eventRepository", err)
	}

	return container
}

func SetupEventPersonServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupEventPersonRepositoryTestContainer(t, config)

	err := container.Provide(func(repository *repositories.EventPersonRepository) *services.EventPersonService {
		return services.NewEventPersonService(repository)
	})
	if err != nil {
		log.Fatal("Failed to provide eventPersonService", err)
	}

	return container
}

func SetupEventPersonHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupEventPersonServiceTestContainer(t, config)

	err := container.Provide(func(service *services.EventPersonService) *apiV1.EventPersonHandler {
		return apiV1.NewEventPersonHandler(service)
	})
	if err != nil {
		log.Fatal("Failed to provide EventPersonHandler", err)
	}

	return container
}

// -------- Person containers: --------

func SetupPersonRepositoryTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupDatabaseTestContainer(t, config)

	err := container.Provide(func(db *sql.DB) *repositories.PersonRepository {
		return repositories.NewPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide personRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.ApplicationRepository {
		return repositories.NewApplicationRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide ApplicationRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.CompanyRepository {
		return repositories.NewCompanyRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide CompanyRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.ApplicationPersonRepository {
		return repositories.NewApplicationPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide ApplicationPersonRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.CompanyPersonRepository {
		return repositories.NewCompanyPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide CompanyPersonRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.EventRepository {
		return repositories.NewEventRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide EventRepository", err)
	}

	err = container.Provide(func(db *sql.DB) *repositories.EventPersonRepository {
		return repositories.NewEventPersonRepository(db)
	})
	if err != nil {
		log.Fatal("Failed to provide EventPersonRepository", err)
	}

	return container
}

func SetupPersonServiceTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupPersonRepositoryTestContainer(t, config)

	err := container.Provide(func(personRepository *repositories.PersonRepository) *services.PersonService {
		return services.NewPersonService(personRepository)
	})
	if err != nil {
		log.Fatal("Failed to provide personService", err)
	}

	return container
}

func SetupPersonHandlerTestContainer(t *testing.T, config configPackage.Config) *dig.Container {
	container := SetupPersonServiceTestContainer(t, config)

	err := container.Provide(func(personService *services.PersonService) *apiV1.PersonHandler {
		return apiV1.NewPersonHandler(personService)
	})
	if err != nil {
		log.Fatal("Failed to provide personHandler", err)
	}

	return container
}
