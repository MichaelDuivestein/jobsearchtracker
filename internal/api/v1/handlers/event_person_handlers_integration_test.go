package handlers_test

import (
	"bytes"
	"encoding/json"
	"jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	configPackage "jobsearchtracker/internal/config"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupEventPersonHandler(t *testing.T) (
	*handlers.EventPersonHandler,
	*repositories.EventRepository,
	*repositories.PersonRepository,
	*repositories.EventPersonRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupEventPersonHandlerTestContainer(t, config)

	var eventPersonHandler *handlers.EventPersonHandler
	err := container.Invoke(func(handler *handlers.EventPersonHandler) {
		eventPersonHandler = handler
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

	var eventPersonRepository *repositories.EventPersonRepository
	err = container.Invoke(func(repository *repositories.EventPersonRepository) {
		eventPersonRepository = repository
	})
	assert.NoError(t, err)

	return eventPersonHandler,
		eventRepository,
		personRepository,
		eventPersonRepository
}

// -------- AssociateEventPerson tests: --------

func TestAssociateEventPerson_ShouldWork(t *testing.T) {
	eventPersonHandler,
		eventRepository,
		personRepository,
		_ := setupEventPersonHandler(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := requests.AssociateEventPersonRequest{
		EventID:  event.ID,
		PersonID: person.ID,
	}

	requestBytes, err := json.Marshal(eventPerson)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/event-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	eventPersonHandler.AssociateEventPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var eventPersonResponse responses.EventPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&eventPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, event.ID, eventPersonResponse.EventID)
	assert.Equal(t, person.ID, eventPersonResponse.PersonID)
	assert.NotNil(t, eventPersonResponse.CreatedDate)
}

// -------- GetEventPersonsByID tests: --------

func TestGetEventPersonsByID_ShouldWork(t *testing.T) {
	eventPersonHandler,
		eventRepository,
		personRepository,
		eventPersonRepository := setupEventPersonHandler(t)

	_, event2ID, person1ID, _ := setupEventPersonTestData(
		t,
		eventRepository,
		personRepository,
		eventPersonRepository,
		false)

	queryParams := "event-id=" + event2ID.String() + "&person-id=" + person1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/event-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventPersonHandler.GetEventPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, event2ID, response[0].EventID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)
}

func TestGetEventPersonsByID_ShouldReturnAllMatchingCompanies(t *testing.T) {
	eventPersonHandler,
		eventRepository,
		personRepository,
		eventPersonRepository := setupEventPersonHandler(t)

	_, event2ID, person1ID, person2ID := setupEventPersonTestData(
		t,
		eventRepository,
		personRepository,
		eventPersonRepository,
		true)

	queryParams := "event-id=" + event2ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/event-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventPersonHandler.GetEventPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, event2ID, response[0].EventID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, event2ID, response[1].EventID)
	assert.Equal(t, person2ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetEventPersonsByID_ShouldReturnAllMatchingPersons(t *testing.T) {
	eventPersonHandler,
		eventRepository,
		personRepository,
		eventPersonRepository := setupEventPersonHandler(t)

	event1ID, event2ID, person1ID, _ := setupEventPersonTestData(
		t,
		eventRepository,
		personRepository,
		eventPersonRepository,
		true)

	queryParams := "person-id=" + person1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/event-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventPersonHandler.GetEventPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, event2ID, response[0].EventID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, event1ID, response[1].EventID)
	assert.Equal(t, person1ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetEventPersonsByID_ShouldReturnEmptyResponseIfNoMatchingEventPersons(t *testing.T) {
	eventPersonHandler,
		eventRepository,
		personRepository,
		eventPersonRepository := setupEventPersonHandler(t)

	setupEventPersonTestData(
		t,
		eventRepository,
		personRepository,
		eventPersonRepository,
		false)

	queryParams := "event-id=" + uuid.New().String() + "&person-id=" + uuid.New().String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/event-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventPersonHandler.GetEventPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- GetAllEventPersons tests: --------

func TestGetAllEventPersons_ShouldReturnAllEventPersons(t *testing.T) {
	eventPersonHandler,
		eventRepository,
		personRepository,
		eventPersonRepository := setupEventPersonHandler(t)

	event1ID, event2ID, person1ID, person2ID := setupEventPersonTestData(
		t,
		eventRepository,
		personRepository,
		eventPersonRepository,
		true)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/event-person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventPersonHandler.GetAllEventPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, event2ID, response[0].EventID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, event2ID, response[1].EventID)
	assert.Equal(t, person2ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)

	assert.Equal(t, event1ID, response[2].EventID)
	assert.Equal(t, person1ID, response[2].PersonID)
	assert.NotNil(t, response[2].CreatedDate)
}

func TestGetAllEventPersons_ShouldReturnNothingIfNothingInDatabase(t *testing.T) {
	eventPersonHandler, _, _, _ := setupEventPersonHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/event-person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventPersonHandler.GetAllEventPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- DeleteEventPerson tests: --------

func TestDeleteEventPerson_ShouldDeleteEventPerson(t *testing.T) {
	eventPersonHandler,
		eventRepository,
		personRepository,
		_ := setupEventPersonHandler(t)

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	eventPerson := requests.AssociateEventPersonRequest{
		EventID:  event.ID,
		PersonID: person.ID,
	}

	requestBytes, err := json.Marshal(eventPerson)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/event-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	eventPersonHandler.AssociateEventPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	deleteRequest := requests.DeleteEventPersonRequest{
		EventID:  event.ID,
		PersonID: person.ID,
	}

	requestBytes, err = json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err = http.NewRequest("POST", "/api/v1/event-person/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()
	eventPersonHandler.DeleteEventPerson(responseRecorder, request)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "", responseBodyString)
}

func TestDeleteEventPerson_ShouldReturnErrorIfNoMatchingEventPersonToDelete(t *testing.T) {
	eventPersonHandler, _, _, _ := setupEventPersonHandler(t)

	eventID, personID := uuid.New(), uuid.New()
	deleteRequest := requests.DeleteEventPersonRequest{
		EventID:  eventID,
		PersonID: personID,
	}

	requestBytes, err := json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/event-person/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	eventPersonHandler.DeleteEventPerson(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"error: object not found: EventPerson does not exist. eventID: "+
			eventID.String()+", personID: "+personID.String()+"\n",
		responseBodyString)
}

// -------- test helpers: --------

func setupEventPersonTestData(
	t *testing.T,
	eventRepository *repositories.EventRepository,
	personRepository *repositories.PersonRepository,
	eventPersonRepository *repositories.EventPersonRepository,
	sleep bool) (
	uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {

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

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	eventPerson2 := models.AssociateEventPerson{
		EventID:  event2.ID,
		PersonID: person2.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson2)
	assert.NoError(t, err)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	eventPerson3 := models.AssociateEventPerson{
		EventID:  event2.ID,
		PersonID: person1.ID,
	}
	_, err = eventPersonRepository.AssociateEventPerson(&eventPerson3)
	assert.NoError(t, err)

	return event1.ID, event2.ID, person1.ID, person2.ID
}
