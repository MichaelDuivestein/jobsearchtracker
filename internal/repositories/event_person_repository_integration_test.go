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

func setupEventPersonRepository(t *testing.T) (
	*repositories.EventPersonRepository, *repositories.EventRepository, *repositories.PersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupEventPersonRepositoryTestContainer(t, *config)

	var eventPersonRepository *repositories.EventPersonRepository
	err := container.Invoke(func(repository *repositories.EventPersonRepository) {
		eventPersonRepository = repository
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

	return eventPersonRepository, eventRepository, personRepository
}

// -------- AssociateEventPerson tests: --------

func TestAssociateEventToPerson_ShouldWork(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:     event.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	associatedEventPerson, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)
	assert.NotNil(t, associatedEventPerson)

	assert.Equal(t, event.ID, associatedEventPerson.EventID)
	assert.Equal(t, person.ID, associatedEventPerson.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, eventPerson.CreatedDate, &associatedEventPerson.CreatedDate)
}

func TestAssociateEventToPerson_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	associatedEventPerson, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)
	assert.NotNil(t, associatedEventPerson)

	assert.Equal(t, event.ID, associatedEventPerson.EventID)
	assert.Equal(t, person.ID, associatedEventPerson.PersonID)
	assert.NotNil(t, associatedEventPerson.CreatedDate)
}

func TestAssociateEventToPerson_ShouldAssociateAEventToMultiplePersons(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	personCompanies, err := eventPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, personCompanies)
	assert.Len(t, personCompanies, 2)

	associatedEventPerson1 := personCompanies[0]
	assert.Equal(t, event.ID, associatedEventPerson1.EventID)
	assert.Equal(t, person2.ID, associatedEventPerson1.PersonID)
	assert.NotNil(t, associatedEventPerson1.CreatedDate)

	associatedEventPerson2 := personCompanies[1]
	assert.Equal(t, event.ID, associatedEventPerson2.EventID)
	assert.Equal(t, person1.ID, associatedEventPerson2.PersonID)
	assert.NotNil(t, associatedEventPerson2.CreatedDate)
}

func TestAssociateEventToPerson_ShouldAssociateMultipleCompaniesToAPerson(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event2.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	personCompanies, err := eventPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, personCompanies)
	assert.Len(t, personCompanies, 2)

	associatedEventPerson1 := personCompanies[0]
	assert.Equal(t, event2.ID, associatedEventPerson1.EventID)
	assert.Equal(t, person.ID, associatedEventPerson1.PersonID)
	assert.NotNil(t, associatedEventPerson1.CreatedDate)

	associatedEventPerson2 := personCompanies[1]
	assert.Equal(t, event1.ID, associatedEventPerson2.EventID)
	assert.Equal(t, person.ID, associatedEventPerson2.PersonID)
	assert.NotNil(t, associatedEventPerson2.CreatedDate)
}

func TestAssociateEventToPerson_ShouldReturnConflictErrorIfEventIDAndPersonIDCombinationAlreadyExist(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: EventID and PersonID combination already exists in database.",
		conflictError.Error())
}

func TestAssociateEventToPerson_ShouldReturnValidationErrorIfPersonIDDoesNotExist(t *testing.T) {
	eventPersonRepository, eventRepository, _ := setupEventPersonRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: uuid.New(),
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateEventToPerson_ShouldReturnValidationErrorIfEventIDDoesNotExist(t *testing.T) {
	eventPersonRepository, _, personRepository := setupEventPersonRepository(t)

	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  uuid.New(),
		PersonID: person.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateEventToPerson_ShouldReturnValidationErrorIfEventIDAndPersonIDDoNotExist(t *testing.T) {
	eventPersonRepository, _, _ := setupEventPersonRepository(t)

	eventPerson := models.AssociateEventPerson{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

// -------- GetByID tests: --------

func TestGetByID_ShouldGetRecordsMatchingEventID(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:     event2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	eventPersons, err := eventPersonRepository.GetByID(&event1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, eventPersons, 2)

	assert.Equal(t, eventPersons[0].EventID, event1.ID)
	assert.Equal(t, eventPersons[0].PersonID, person2.ID)

	assert.Equal(t, eventPersons[1].EventID, event1.ID)
	assert.Equal(t, eventPersons[1].PersonID, person1.ID)
}

func TestEventPersonGetByID_ShouldGetRecordsMatchingPersonID(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:     event2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	eventPersons, err := eventPersonRepository.GetByID(nil, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, eventPersons, 2)

	assert.Equal(t, eventPersons[0].EventID, event2.ID)
	assert.Equal(t, eventPersons[0].PersonID, person1.ID)

	assert.Equal(t, eventPersons[1].EventID, event1.ID)
	assert.Equal(t, eventPersons[1].PersonID, person1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingEventIDAndPersonID(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person1.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person2.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:  event2.ID,
		PersonID: person1.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	persons, err := eventPersonRepository.GetByID(&event1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, persons, 1)
	assert.Equal(t, event1.ID, persons[0].EventID)
	assert.Equal(t, person1.ID, persons[0].PersonID)
}

func TestGetByID_ShouldGetNoRecordsIfEventIDDoesNotMatch(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person1.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person2.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:  event2.ID,
		PersonID: person1.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	persons, err := eventPersonRepository.GetByID(testutil.ToPtr(uuid.New()), &person1.ID)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestEventPersonGetByID_ShouldGetNoRecordsIfPersonIDDoesNotMatch(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person1.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person2.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:  event2.ID,
		PersonID: person1.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	persons, err := eventPersonRepository.GetByID(&event1.ID, testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestGetByID_ShouldGetNoRecordsIfEventIDAndPersonIDDoesNotMatch(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person1.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:  event1.ID,
		PersonID: person2.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:  event2.ID,
		PersonID: person1.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	persons, err := eventPersonRepository.GetByID(testutil.ToPtr(uuid.New()), testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestEventPersonGetByID_ShouldGetNoRecordsIfNoRecordsInDB(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	persons, err := eventPersonRepository.GetByID(&event1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

// -------- GetAll tests: --------

func TestGetAllEventPersons_ShouldReturnAllEventPersons(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson1 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson1)
	assert.NoError(t, err)

	eventPerson2 := models.AssociateEventPerson{
		EventID:     event1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	eventPerson3 := models.AssociateEventPerson{
		EventID:     event2.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	personCompanies, err := eventPersonRepository.GetAll()
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
	eventPersonRepository, _, _ := setupEventPersonRepository(t)

	results, err := eventPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteEventPerson_ShouldDeleteEventPerson(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	model := models.DeleteEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}

	err = eventPersonRepository.Delete(&model)
	assert.NoError(t, err)
}

func TestDeleteEventPerson_ShouldReturnNotFoundErrorIfNoMatchingEventPersonInDatabase(t *testing.T) {
	eventPersonRepository, eventRepository, personRepository := setupEventPersonRepository(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := models.AssociateEventPerson{
		EventID:  event.ID,
		PersonID: person.ID,
	}
	_, err := eventPersonRepository.AssociateEventPerson(&eventPerson)
	assert.NoError(t, err)

	model := models.DeleteEventPerson{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}

	err = eventPersonRepository.Delete(&model)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: EventPerson does not exist. eventID: "+model.EventID.String()+
			", personID: "+model.PersonID.String(), notFoundError.Error())
}
