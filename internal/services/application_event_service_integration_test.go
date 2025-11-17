package services_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/services"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupApplicationEventService(t *testing.T) (
	*services.ApplicationEventService,
	*repositories.ApplicationRepository,
	*repositories.EventRepository,
	*repositories.CompanyRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupApplicationEventServiceTestContainer(t, *config)

	var applicationEventService *services.ApplicationEventService
	err := container.Invoke(func(service *services.ApplicationEventService) {
		applicationEventService = service
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

	return applicationEventService, applicationRepository, eventRepository, companyRepository
}

// -------- AssociateApplicationEvent tests: --------

func TestAssociateApplicationToEvent_ShouldAssociateAApplicationToAEvent(t *testing.T) {
	applicationEventService,
		applicationRepository,
		eventRepository,
		companyRepository := setupApplicationEventService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	associatedApplicationEvent, err := applicationEventService.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	assert.Equal(t, applicationEvent.ApplicationID, associatedApplicationEvent.ApplicationID)
	assert.Equal(t, applicationEvent.EventID, associatedApplicationEvent.EventID)
	testutil.AssertEqualFormattedDateTimes(t, applicationEvent.CreatedDate, &associatedApplicationEvent.CreatedDate)
}

func TestAssociateApplicationToEvent_ShouldAssociateAApplicationToAEventWithOnlyRequiredFields(t *testing.T) {
	applicationEventService,
		applicationRepository,
		eventRepository,
		companyRepository := setupApplicationEventService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	associatedApplicationEvent, err := applicationEventService.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	assert.Equal(t, applicationEvent.ApplicationID, associatedApplicationEvent.ApplicationID)
	assert.Equal(t, applicationEvent.EventID, associatedApplicationEvent.EventID)
	assert.NotNil(t, associatedApplicationEvent.CreatedDate)
}

func TestAssociateApplicationToEvent_ShouldReturnConflictErrorIfApplicationIDAndEventIDCombinationAlreadyExist(t *testing.T) {
	applicationEventService,
		applicationRepository,
		eventRepository,
		companyRepository := setupApplicationEventService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	_, err := applicationEventService.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: ApplicationID and EventID combination already exists in database.",
		conflictError.Error())
}

// -------- GetByID tests: --------

func TestApplicationPersonGetByID_ShouldGetRecordsMatchingApplicationID(t *testing.T) {
	applicationEventService, applicationRepository, eventRepository, companyRepository := setupApplicationEventService(t)

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
	_, err := applicationEventService.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	applicationEvents, err := applicationEventService.GetByID(&application1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, applicationEvents, 2)

	assert.Equal(t, applicationEvents[0].ApplicationID, application1.ID)
	assert.Equal(t, applicationEvents[0].EventID, event2.ID)

	assert.Equal(t, applicationEvents[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationEvents[1].EventID, event1.ID)
}

func TestApplicationEventGetByID_ShouldGetRecordsMatchingEventID(t *testing.T) {
	applicationEventService, applicationRepository, eventRepository, companyRepository := setupApplicationEventService(t)

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
	_, err := applicationEventService.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	applicationEvents, err := applicationEventService.GetByID(nil, &event1.ID)
	assert.NoError(t, err)
	assert.Len(t, applicationEvents, 2)

	assert.Equal(t, applicationEvents[0].ApplicationID, application2.ID)
	assert.Equal(t, applicationEvents[0].EventID, event1.ID)

	assert.Equal(t, applicationEvents[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationEvents[1].EventID, event1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingApplicationIDAndEventID(t *testing.T) {
	applicationEventService, applicationRepository, eventRepository, companyRepository := setupApplicationEventService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
	}
	_, err := applicationEventService.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	events, err := applicationEventService.GetByID(&application1.ID, &event1.ID)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, application1.ID, events[0].ApplicationID)
	assert.Equal(t, event1.ID, events[0].EventID)
}

// -------- GetAll tests: --------

func TestGetAllApplicationEvents_ShouldReturnAllApplicationEvents(t *testing.T) {
	applicationEventService, applicationRepository, eventRepository, companyRepository := setupApplicationEventService(t)

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
	_, err := applicationEventService.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationEventService.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	eventCompanies, err := applicationEventService.GetAll()
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
	applicationEventService, _, _, _ := setupApplicationEventService(t)

	results, err := applicationEventService.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteApplicationEvent_ShouldDeleteApplicationEvent(t *testing.T) {
	applicationEventService, applicationRepository, eventRepository, companyRepository := setupApplicationEventService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	_, err := applicationEventService.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	deleteModel := models.DeleteApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}

	err = applicationEventService.Delete(&deleteModel)
	assert.NoError(t, err)
}

func TestDeleteApplicationEvent_ShouldReturnNotFoundErrorIfNoMatchingApplicationEventInDatabase(t *testing.T) {
	applicationEventService, applicationRepository, eventRepository, companyRepository := setupApplicationEventService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := models.AssociateApplicationEvent{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}
	_, err := applicationEventService.AssociateApplicationEvent(&applicationEvent)
	assert.NoError(t, err)

	deleteModel := models.DeleteApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}

	err = applicationEventService.Delete(&deleteModel)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ApplicationEvent does not exist. applicationID: "+deleteModel.ApplicationID.String()+
			", eventID: "+deleteModel.EventID.String(), notFoundError.Error())
}
