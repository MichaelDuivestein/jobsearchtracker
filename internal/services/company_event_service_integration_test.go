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

func setupCompanyEventService(t *testing.T) (
	*services.CompanyEventService, *repositories.CompanyRepository, *repositories.EventRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyEventServiceTestContainer(t, *config)

	var companyEventService *services.CompanyEventService
	err := container.Invoke(func(service *services.CompanyEventService) {
		companyEventService = service
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	return companyEventService, companyRepository, eventRepository
}

// -------- AssociateCompanyEvent tests: --------

func TestAssociateCompanyToEvent_ShouldAssociateACompanyToAEvent(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent := models.AssociateCompanyEvent{
		CompanyID:   company.ID,
		EventID:     event.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	associatedCompanyEvent, err := companyEventService.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	assert.Equal(t, companyEvent.CompanyID, associatedCompanyEvent.CompanyID)
	assert.Equal(t, companyEvent.EventID, associatedCompanyEvent.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent.CreatedDate, &associatedCompanyEvent.CreatedDate)
}

func TestAssociateCompanyToEvent_ShouldAssociateACompanyToAEventWithOnlyRequiredFields(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: company.ID,
		EventID:   event.ID,
	}
	associatedCompanyEvent, err := companyEventService.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	assert.Equal(t, companyEvent.CompanyID, associatedCompanyEvent.CompanyID)
	assert.Equal(t, companyEvent.EventID, associatedCompanyEvent.EventID)
	assert.NotNil(t, associatedCompanyEvent.CreatedDate)
}

func TestAssociateCompanyToEvent_ShouldReturnConflictErrorIfCompanyIDAndEventIDCombinationAlreadyExist(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: company.ID,
		EventID:   event.ID,
	}
	_, err := companyEventService.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	_, err = companyEventService.AssociateCompanyEvent(&companyEvent)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: CompanyID and EventID combination already exists in database.",
		conflictError.Error())
}

// -------- GetByID tests: --------

func TestCompanyEventServiceGetByID_ShouldGetRecordsMatchingCompanyID(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   company1.ID,
		EventID:     event1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventService.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   company1.ID,
		EventID:     event2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID:   company2.ID,
		EventID:     event1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	companyEvents, err := companyEventService.GetByID(&company1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, companyEvents, 2)

	assert.Equal(t, companyEvents[0].CompanyID, company1.ID)
	assert.Equal(t, companyEvents[0].EventID, event2.ID)

	assert.Equal(t, companyEvents[1].CompanyID, company1.ID)
	assert.Equal(t, companyEvents[1].EventID, event1.ID)
}

func TestCompanyEventGetByID_ShouldGetRecordsMatchingEventID(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   company1.ID,
		EventID:     event1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventService.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   company1.ID,
		EventID:     event2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID:   company2.ID,
		EventID:     event1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	companyEvents, err := companyEventService.GetByID(nil, &event1.ID)
	assert.NoError(t, err)
	assert.Len(t, companyEvents, 2)

	assert.Equal(t, companyEvents[0].CompanyID, company2.ID)
	assert.Equal(t, companyEvents[0].EventID, event1.ID)

	assert.Equal(t, companyEvents[1].CompanyID, company1.ID)
	assert.Equal(t, companyEvents[1].EventID, event1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingCompanyIDAndEventID(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID: company1.ID,
		EventID:   event1.ID,
	}
	_, err := companyEventService.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID: company1.ID,
		EventID:   event2.ID,
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID: company2.ID,
		EventID:   event1.ID,
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	events, err := companyEventService.GetByID(&company1.ID, &event1.ID)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, company1.ID, events[0].CompanyID)
	assert.Equal(t, event1.ID, events[0].EventID)
}

// -------- GetAll tests: --------

func TestGetAllCompanyEvents_ShouldReturnAllCompanyEvents(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   company1.ID,
		EventID:     event1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventService.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   company1.ID,
		EventID:     event2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID:   company2.ID,
		EventID:     event2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventService.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	eventCompanies, err := companyEventService.GetAll()
	assert.NoError(t, err)

	assert.Len(t, eventCompanies, 3)

	insertedCompanyEvent1 := eventCompanies[0]
	assert.Equal(t, company1.ID, insertedCompanyEvent1.CompanyID)
	assert.Equal(t, event2.ID, insertedCompanyEvent1.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent2.CreatedDate, &insertedCompanyEvent1.CreatedDate)

	insertedCompanyEvent2 := eventCompanies[1]
	assert.Equal(t, company2.ID, insertedCompanyEvent2.CompanyID)
	assert.Equal(t, event2.ID, insertedCompanyEvent2.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent3.CreatedDate, &insertedCompanyEvent2.CreatedDate)

	insertedCompanyEvent3 := eventCompanies[2]
	assert.Equal(t, company1.ID, insertedCompanyEvent3.CompanyID)
	assert.Equal(t, event1.ID, insertedCompanyEvent3.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent1.CreatedDate, &insertedCompanyEvent3.CreatedDate)
}

func TestGetAllCompanyEvents_ShouldReturnNilIfNoEventsInDatabase(t *testing.T) {
	companyEventService, _, _ := setupCompanyEventService(t)

	results, err := companyEventService.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteCompanyEvent_ShouldDeleteCompanyEvent(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: company.ID,
		EventID:   event.ID,
	}
	_, err := companyEventService.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	deleteModel := models.DeleteCompanyEvent{
		CompanyID: company.ID,
		EventID:   event.ID,
	}

	err = companyEventService.Delete(&deleteModel)
	assert.NoError(t, err)
}

func TestDeleteCompanyEvent_ShouldReturnNotFoundErrorIfNoMatchingCompanyEventInDatabase(t *testing.T) {
	companyEventService, companyRepository, eventRepository := setupCompanyEventService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: company.ID,
		EventID:   event.ID,
	}
	_, err := companyEventService.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	deleteModel := models.DeleteCompanyEvent{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}

	err = companyEventService.Delete(&deleteModel)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: CompanyEvent does not exist. companyID: "+deleteModel.CompanyID.String()+
			", eventID: "+deleteModel.EventID.String(), notFoundError.Error())
}
