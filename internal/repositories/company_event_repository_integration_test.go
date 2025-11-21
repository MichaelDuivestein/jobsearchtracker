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

func setupCompanyEventRepository(t *testing.T) (
	*repositories.CompanyEventRepository,
	*repositories.CompanyRepository,
	*repositories.EventRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyEventRepositoryTestContainer(t, *config)

	var companyEventRepository *repositories.CompanyEventRepository
	err := container.Invoke(func(repository *repositories.CompanyEventRepository) {
		companyEventRepository = repository
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

	return companyEventRepository, companyRepository, eventRepository
}

// -------- AssociateCompanyEvent tests: --------

func TestAssociateCompanyToEvent_ShouldWork(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent := models.AssociateCompanyEvent{
		CompanyID:   companyID,
		EventID:     eventID,
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	associatedCompanyEvent, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)
	assert.NotNil(t, associatedCompanyEvent)

	assert.Equal(t, companyID, associatedCompanyEvent.CompanyID)
	assert.Equal(t, eventID, associatedCompanyEvent.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent.CreatedDate, &associatedCompanyEvent.CreatedDate)
}

func TestAssociateCompanyToEvent_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: companyID,
		EventID:   eventID,
	}
	associatedCompanyEvent, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)
	assert.NotNil(t, associatedCompanyEvent)

	assert.Equal(t, companyID, associatedCompanyEvent.CompanyID)
	assert.Equal(t, eventID, associatedCompanyEvent.EventID)
	assert.NotNil(t, associatedCompanyEvent.CreatedDate)
}

func TestAssociateCompanyToEvent_ShouldAssociateAnCompanyToMultipleEvents(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   companyID,
		EventID:     event1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   companyID,
		EventID:     event2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	eventCompanies, err := companyEventRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, eventCompanies)
	assert.Len(t, eventCompanies, 2)

	associatedCompanyEvent1 := eventCompanies[0]
	assert.Equal(t, companyID, associatedCompanyEvent1.CompanyID)
	assert.Equal(t, event2ID, associatedCompanyEvent1.EventID)
	assert.NotNil(t, associatedCompanyEvent1.CreatedDate)

	associatedCompanyEvent2 := eventCompanies[1]
	assert.Equal(t, companyID, associatedCompanyEvent2.CompanyID)
	assert.Equal(t, event1ID, associatedCompanyEvent2.EventID)
	assert.NotNil(t, associatedCompanyEvent2.CreatedDate)
}

func TestAssociateCompanyToEvent_ShouldAssociateMultipleCompaniesToAEvent(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   company1ID,
		EventID:     event.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   company2ID,
		EventID:     event.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	eventCompanies, err := companyEventRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, eventCompanies)
	assert.Len(t, eventCompanies, 2)

	associatedCompanyEvent1 := eventCompanies[0]
	assert.Equal(t, company2ID, associatedCompanyEvent1.CompanyID)
	assert.Equal(t, event.ID, associatedCompanyEvent1.EventID)
	assert.NotNil(t, associatedCompanyEvent1.CreatedDate)

	associatedCompanyEvent2 := eventCompanies[1]
	assert.Equal(t, company1ID, associatedCompanyEvent2.CompanyID)
	assert.Equal(t, event.ID, associatedCompanyEvent2.EventID)
	assert.NotNil(t, associatedCompanyEvent2.CreatedDate)
}

func TestAssociateCompanyToEvent_ShouldReturnConflictErrorIfCompanyIDAndEventIDCombinationAlreadyExist(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: companyID,
		EventID:   eventID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: CompanyID and EventID combination already exists in database.",
		conflictError.Error())
}

func TestAssociateCompanyToEvent_ShouldReturnValidationErrorIfEventIDDoesNotExist(t *testing.T) {
	companyEventRepository, companyRepository, _ := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: companyID,
		EventID:   uuid.New(),
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateCompanyToEvent_ShouldReturnValidationErrorIfCompanyIDDoesNotExist(t *testing.T) {
	companyEventRepository, _, eventRepository := setupCompanyEventRepository(t)

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: uuid.New(),
		EventID:   eventID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateCompanyToEvent_ShouldReturnValidationErrorIfCompanyIDAndEventIDDoNotExist(t *testing.T) {
	companyEventRepository, _, _ := setupCompanyEventRepository(t)

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

// -------- GetByID tests: --------

func TestCompanyEventGetByID_ShouldGetRecordsMatchingCompanyID(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   company1ID,
		EventID:     event1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   company1ID,
		EventID:     event2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID:   company2ID,
		EventID:     event1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	companyEvents, err := companyEventRepository.GetByID(&company1ID, nil)
	assert.NoError(t, err)
	assert.Len(t, companyEvents, 2)

	assert.Equal(t, companyEvents[0].CompanyID, company1ID)
	assert.Equal(t, companyEvents[0].EventID, event2ID)

	assert.Equal(t, companyEvents[1].CompanyID, company1ID)
	assert.Equal(t, companyEvents[1].EventID, event1ID)
}

func TestCompanyEventGetByID_ShouldGetRecordsMatchingEventID(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   company1ID,
		EventID:     event1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   company1ID,
		EventID:     event2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID:   company2ID,
		EventID:     event1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	companyEvents, err := companyEventRepository.GetByID(nil, &event1ID)
	assert.NoError(t, err)
	assert.Len(t, companyEvents, 2)

	assert.Equal(t, companyEvents[0].CompanyID, company2ID)
	assert.Equal(t, companyEvents[0].EventID, event1ID)

	assert.Equal(t, companyEvents[1].CompanyID, company1ID)
	assert.Equal(t, companyEvents[1].EventID, event1ID)
}

func TestGetByID_ShouldGetRecordsMatchingCompanyIDAndEventID(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event1ID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	events, err := companyEventRepository.GetByID(&company1ID, &event1ID)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, company1ID, events[0].CompanyID)
	assert.Equal(t, event1ID, events[0].EventID)
}

func TestCompanyEventGetByID_ShouldGetNoRecordsIfCompanyIDDoesNotMatch(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event1ID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	events, err := companyEventRepository.GetByID(testutil.ToPtr(uuid.New()), &event1ID)
	assert.NoError(t, err)
	assert.Nil(t, events)
}

func TestCompanyEventGetByID_ShouldGetNoRecordsIfEventIDDoesNotMatch(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event1ID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	events, err := companyEventRepository.GetByID(&company1ID, testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, events)
}

func TestGetByID_ShouldGetNoRecordsIfCompanyIDAndEventIDDoesNotMatch(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event1ID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID: company1ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	events, err := companyEventRepository.GetByID(testutil.ToPtr(uuid.New()), testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, events)
}

func TestCompanyEventGetByID_ShouldGetNoRecordsIfNoRecordsInDB(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	events, err := companyEventRepository.GetByID(&companyID, &eventID)
	assert.NoError(t, err)
	assert.Nil(t, events)
}

// -------- GetAll tests: --------

func TestGetAllCompanyEvents_ShouldReturnAllCompanyEvents(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	event1ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	event2ID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent1 := models.AssociateCompanyEvent{
		CompanyID:   company1ID,
		EventID:     event1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent1)
	assert.NoError(t, err)

	companyEvent2 := models.AssociateCompanyEvent{
		CompanyID:   company1ID,
		EventID:     event2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent2)
	assert.NoError(t, err)

	companyEvent3 := models.AssociateCompanyEvent{
		CompanyID:   company2ID,
		EventID:     event2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&companyEvent3)
	assert.NoError(t, err)

	eventCompanies, err := companyEventRepository.GetAll()
	assert.NoError(t, err)

	assert.Len(t, eventCompanies, 3)

	insertedCompanyEvent1 := eventCompanies[0]
	assert.Equal(t, company1ID, insertedCompanyEvent1.CompanyID)
	assert.Equal(t, event2ID, insertedCompanyEvent1.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent2.CreatedDate, &insertedCompanyEvent1.CreatedDate)

	insertedCompanyEvent2 := eventCompanies[1]
	assert.Equal(t, company2ID, insertedCompanyEvent2.CompanyID)
	assert.Equal(t, event2ID, insertedCompanyEvent2.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent3.CreatedDate, &insertedCompanyEvent2.CreatedDate)

	insertedCompanyEvent3 := eventCompanies[2]
	assert.Equal(t, company1ID, insertedCompanyEvent3.CompanyID)
	assert.Equal(t, event1ID, insertedCompanyEvent3.EventID)
	testutil.AssertEqualFormattedDateTimes(t, companyEvent1.CreatedDate, &insertedCompanyEvent3.CreatedDate)
}

func TestGetAllCompanyEvents_ShouldReturnNilIfNoEventsInDatabase(t *testing.T) {
	companyEventRepository, _, _ := setupCompanyEventRepository(t)

	results, err := companyEventRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteCompanyEvent_ShouldDeleteCompanyEvent(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: companyID,
		EventID:   eventID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	model := models.DeleteCompanyEvent{
		CompanyID: companyID,
		EventID:   eventID,
	}

	err = companyEventRepository.Delete(&model)
	assert.NoError(t, err)
}

func TestDeleteCompanyEvent_ShouldReturnNotFoundErrorIfNoMatchingCompanyEventInDatabase(t *testing.T) {
	companyEventRepository, companyRepository, eventRepository := setupCompanyEventRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	companyEvent := models.AssociateCompanyEvent{
		CompanyID: companyID,
		EventID:   eventID,
	}
	_, err := companyEventRepository.AssociateCompanyEvent(&companyEvent)
	assert.NoError(t, err)

	model := models.DeleteCompanyEvent{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}

	err = companyEventRepository.Delete(&model)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: CompanyEvent does not exist. companyID: "+model.CompanyID.String()+
			", eventID: "+model.EventID.String(), notFoundError.Error())
}
