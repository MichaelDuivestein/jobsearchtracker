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

func setupEventPersonService(t *testing.T) (
	*services.EventPersonService, *repositories.EventRepository, *repositories.PersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupEventPersonServiceTestContainer(t, *config)

	var eventPersonService *services.EventPersonService
	err := container.Invoke(func(service *services.EventPersonService) {
		eventPersonService = service
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	return eventPersonService, eventRepository, personRepository
}

// -------- AssociateEventPerson tests: --------

func TestAssociateEventToPerson_ShouldAssociateAEventToAPerson(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:     event.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	associatedEventPerson, err := eventPersonService.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	assert.Equal(t, eventPerson.EventID, associatedEventPerson.EventID)
	assert.Equal(t, eventPerson.PersonID, associatedEventPerson.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, eventPerson.CreatedDate, &associatedEventPerson.CreatedDate)
}

func TestAssociateEventToPerson_ShouldAssociateAEventToAPersonWithOnlyRequiredFields(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	associatedEventPerson, err := eventPersonService.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	assert.Equal(t, eventPerson.EventID, associatedEventPerson.EventID)
	assert.Equal(t, eventPerson.PersonID, associatedEventPerson.PersonID)
	assert.NotNil(t, associatedEventPerson.CreatedDate)
}

func TestAssociateEventToPerson_ShouldReturnConflictErrorIfEventIDAndPersonIDCombinationAlreadyExist(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	_, err := eventPersonService.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	_, err = eventPersonService.AssociateEventPerson(&eventPerson)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: EventID and PersonID combination already exists in database.",
		conflictError.Error())
}

// -------- GetByID tests: --------

func TestGetByID_ShouldGetRecordsMatchingEventID(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonService.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:     event2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	eventPersons, err := eventPersonService.GetByID(&event1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, eventPersons, 2)

	assert.Equal(t, eventPersons[0].EventID, event1.ID)
	assert.Equal(t, eventPersons[0].PersonID, person2.ID)

	assert.Equal(t, eventPersons[1].EventID, event1.ID)
	assert.Equal(t, eventPersons[1].PersonID, person1.ID)
}

func TestEventPersonGetByID_ShouldGetRecordsMatchingPersonID(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonService.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:     event2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	eventPersons, err := eventPersonService.GetByID(nil, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, eventPersons, 2)

	assert.Equal(t, eventPersons[0].EventID, event2.ID)
	assert.Equal(t, eventPersons[0].PersonID, person1.ID)

	assert.Equal(t, eventPersons[1].EventID, event1.ID)
	assert.Equal(t, eventPersons[1].PersonID, person1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingEventIDAndPersonID(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person1.ID,
	}
	_, err := eventPersonService.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person2.ID,
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:  event2.ID,
		PersonID: person1.ID,
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	persons, err := eventPersonService.GetByID(&event1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, persons, 1)
	assert.Equal(t, event1.ID, persons[0].EventID)
	assert.Equal(t, person1.ID, persons[0].PersonID)
}

// -------- GetAll tests: --------

func TestGetAllEventPersons_ShouldReturnAllEventPersons(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonService.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:     event2.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonService.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	personCompanies, err := eventPersonService.GetAll()
	assert.NoError(t, err)

	assert.Len(t, personCompanies, 3)

	insertedEventPerson1 := personCompanies[0]
	assert.Equal(t, event1.ID, insertedEventPerson1.EventID)
	assert.Equal(t, person2.ID, insertedEventPerson1.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, eventPerson2.CreatedDate, &insertedEventPerson1.CreatedDate)

	insertedEventPerson2 := personCompanies[1]
	assert.Equal(t, event2.ID, insertedEventPerson2.EventID)
	assert.Equal(t, person2.ID, insertedEventPerson2.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, eventPerson3.CreatedDate, &insertedEventPerson2.CreatedDate)

	insertedEventPerson3 := personCompanies[2]
	assert.Equal(t, event1.ID, insertedEventPerson3.EventID)
	assert.Equal(t, person1.ID, insertedEventPerson3.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, eventPerson1.CreatedDate, &insertedEventPerson3.CreatedDate)
}

func TestGetAllEventPersons_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	eventPersonService, _, _ := setupEventPersonService(t)

	results, err := eventPersonService.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteEventPerson_ShouldDeleteEventPerson(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	_, err := eventPersonService.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	deleteModel := models.DeleteEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}

	err = eventPersonService.Delete(&deleteModel)
	assert.NoError(t, err)
}

func TestDeleteEventPerson_ShouldReturnNotFoundErrorIfNoMatchingEventPersonInDatabase(t *testing.T) {
	eventPersonService, eventRepository, personRepository := setupEventPersonService(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	_, err := eventPersonService.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	deleteModel := models.DeleteEventPerson{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}

	err = eventPersonService.Delete(&deleteModel)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: EventPerson does not exist. eventID: "+deleteModel.EventID.String()+
			", personID: "+deleteModel.PersonID.String(), notFoundError.Error())
}
