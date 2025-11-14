package repositories_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupApplicationEventRepository(t *testing.T) (
	*repositories.ApplicationEventRepository,
	*repositories.ApplicationRepository,
	*repositories.EventRepository,
	*repositories.CompanyRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupApplicationEventRepositoryTestContainer(t, *config)

	var applicationEventRepository *repositories.ApplicationEventRepository
	err := container.Invoke(func(repository *repositories.ApplicationEventRepository) {
		applicationEventRepository = repository
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	return applicationEventRepository, applicationRepository, eventRepository, companyRepository
}

// -------- AssociateApplicationEvent tests: --------

func TestAssociateApplicationToEvent_ShouldWork(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	associatedApplicationEvent, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)
	assert.NotNil(t, associatedApplicationEvent)

	assert.Equal(t, application.ID, associatedApplicationEvent.ApplicationID)
	assert.Equal(t, event.ID, associatedApplicationEvent.EventID)
	testutil.AssertEqualFormattedDateTimes(t, applicationEvent.CreatedDate, &associatedApplicationEvent.CreatedDate)
}

func TestAssociateApplicationToEvent_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	associatedApplicationEvent, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)
	assert.NotNil(t, associatedApplicationEvent)

	assert.Equal(t, application.ID, associatedApplicationEvent.ApplicationID)
	assert.Equal(t, event.ID, associatedApplicationEvent.EventID)
	assert.NotNil(t, associatedApplicationEvent.CreatedDate)
}

func TestAssociateApplicationToEvent_ShouldAssociateAnApplicationToMultipleEvents(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	eventCompanies, err := applicationEventRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, eventCompanies)
	assert.Len(t, eventCompanies, 2)

	associatedApplicationEvent1 := eventCompanies[0]
	assert.Equal(t, application.ID, associatedApplicationEvent1.ApplicationID)
	assert.Equal(t, event2.ID, associatedApplicationEvent1.EventID)
	assert.NotNil(t, associatedApplicationEvent1.CreatedDate)

	associatedApplicationEvent2 := eventCompanies[1]
	assert.Equal(t, application.ID, associatedApplicationEvent2.ApplicationID)
	assert.Equal(t, event1.ID, associatedApplicationEvent2.EventID)
	assert.NotNil(t, associatedApplicationEvent2.CreatedDate)
}

func TestAssociateApplicationToEvent_ShouldAssociateMultipleApplicationsToAEvent(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	eventCompanies, err := applicationEventRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, eventCompanies)
	assert.Len(t, eventCompanies, 2)

	associatedApplicationEvent1 := eventCompanies[0]
	assert.Equal(t, application2.ID, associatedApplicationEvent1.ApplicationID)
	assert.Equal(t, event.ID, associatedApplicationEvent1.EventID)
	assert.NotNil(t, associatedApplicationEvent1.CreatedDate)

	associatedApplicationEvent2 := eventCompanies[1]
	assert.Equal(t, application1.ID, associatedApplicationEvent2.ApplicationID)
	assert.Equal(t, event.ID, associatedApplicationEvent2.EventID)
	assert.NotNil(t, associatedApplicationEvent2.CreatedDate)
}

func TestAssociateApplicationToEvent_ShouldReturnConflictErrorIfApplicationIDAndEventIDCombinationAlreadyExist(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: ApplicationID and EventID combination already exists in database.",
		conflictError.Error())
}

func TestAssociateApplicationToEvent_ShouldReturnValidationErrorIfEventIDDoesNotExist(t *testing.T) {
	applicationEventRepository, applicationRepository, _, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       uuid.New(),
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateApplicationToEvent_ShouldReturnValidationErrorIfApplicationIDDoesNotExist(t *testing.T) {
	applicationEventRepository, _, eventRepository, _ := setupApplicationEventRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       event.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateApplicationToEvent_ShouldReturnValidationErrorIfApplicationIDAndEventIDDoNotExist(t *testing.T) {
	applicationEventRepository, _, _, _ := setupApplicationEventRepository(t)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

// -------- GetByID tests: --------

func TestApplicationEventGetByID_ShouldGetRecordsMatchingApplicationID(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	applicationEvents, err := applicationEventRepository.GetByID(&application1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, applicationEvents, 2)

	assert.Equal(t, applicationEvents[0].ApplicationID, application1.ID)
	assert.Equal(t, applicationEvents[0].EventID, event2.ID)

	assert.Equal(t, applicationEvents[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationEvents[1].EventID, event1.ID)
}

func TestApplicationEventGetByID_ShouldGetRecordsMatchingEventID(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	applicationEvents, err := applicationEventRepository.GetByID(nil, &event1.ID)
	assert.NoError(t, err)
	assert.Len(t, applicationEvents, 2)

	assert.Equal(t, applicationEvents[0].ApplicationID, application2.ID)
	assert.Equal(t, applicationEvents[0].EventID, event1.ID)

	assert.Equal(t, applicationEvents[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationEvents[1].EventID, event1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingApplicationIDAndEventID(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository :=
		setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	events, err := applicationEventRepository.GetByID(&application1.ID, &event1.ID)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, application1.ID, events[0].ApplicationID)
	assert.Equal(t, event1.ID, events[0].EventID)
}

func TestApplicationEventGetByID_ShouldGetNoRecordsIfApplicationIDDoesNotMatch(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	events, err := applicationEventRepository.GetByID(testutil.ToPtr(uuid.New()), &event1.ID)
	assert.NoError(t, err)
	assert.Nil(t, events)
}

func TestApplicationEventGetByID_ShouldGetNoRecordsIfEventIDDoesNotMatch(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	events, err := applicationEventRepository.GetByID(&application1.ID, testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, events)
}

func TestGetByID_ShouldGetNoRecordsIfApplicationIDAndEventIDDoesNotMatch(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	events, err := applicationEventRepository.GetByID(testutil.ToPtr(uuid.New()), testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, events)
}

func TestApplicationEventGetByID_ShouldGetNoRecordsIfNoRecordsInDB(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	events, err := applicationEventRepository.GetByID(&application1.ID, &event1.ID)
	assert.NoError(t, err)
	assert.Nil(t, events)
}

// -------- GetAll tests: --------

func TestGetAllApplicationEvents_ShouldReturnAllApplicationEvents(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	eventCompanies, err := applicationEventRepository.GetAll()
	assert.NoError(t, err)

	assert.Len(t, eventCompanies, 3)

	insertedApplicationEvent1 := eventCompanies[0]
	assert.Equal(t, application1.ID, insertedApplicationEvent1.ApplicationID)
	assert.Equal(t, event2.ID, insertedApplicationEvent1.EventID)
	testutil.AssertEqualFormattedDateTimes(t, applicationEvent2.CreatedDate, &insertedApplicationEvent1.CreatedDate)

	insertedApplicationEvent2 := eventCompanies[1]
	assert.Equal(t, application2.ID, insertedApplicationEvent2.ApplicationID)
	assert.Equal(t, event2.ID, insertedApplicationEvent2.EventID)
	testutil.AssertEqualFormattedDateTimes(t, applicationEvent3.CreatedDate, &insertedApplicationEvent2.CreatedDate)

	insertedApplicationEvent3 := eventCompanies[2]
	assert.Equal(t, application1.ID, insertedApplicationEvent3.ApplicationID)
	assert.Equal(t, event1.ID, insertedApplicationEvent3.EventID)
	testutil.AssertEqualFormattedDateTimes(t, applicationEvent1.CreatedDate, &insertedApplicationEvent3.CreatedDate)
}

func TestGetAllApplicationEvents_ShouldReturnNilIfNoEventsInDatabase(t *testing.T) {
	applicationEventRepository, _, _, _ := setupApplicationEventRepository(t)

	results, err := applicationEventRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteApplicationEvent_ShouldDeleteApplicationEvent(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	model := models.DeleteApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}

	err = applicationEventRepository.Delete(&model)
	assert.NoError(t, err)
}

func TestDeleteApplicationEvent_ShouldReturnNotFoundErrorIfNoMatchingApplicationEventInDatabase(t *testing.T) {
	applicationEventRepository, applicationRepository, eventRepository, companyRepository := setupApplicationEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	model := models.DeleteApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}

	err = applicationEventRepository.Delete(&model)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ApplicationEvent does not exist. applicationID: "+model.ApplicationID.String()+
			", eventID: "+model.EventID.String(), notFoundError.Error())
}
