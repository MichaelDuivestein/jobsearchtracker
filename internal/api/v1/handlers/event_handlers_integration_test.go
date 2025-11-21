package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupEventHandler(t *testing.T) (
	*handlers.EventHandler,
	*repositories.EventRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupEventHandlerTestContainer(t, config)

	var eventHandler *handlers.EventHandler
	err := container.Invoke(func(handler *handlers.EventHandler) {
		eventHandler = handler
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	return eventHandler, eventRepository
}

// -------- CreateEvent tests: --------

func TestCreateEvent_ShouldInsertAndReturnEvent(t *testing.T) {
	eventHandler, _ := setupEventHandler(t)

	requestBody := requests.CreateEventRequest{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   requests.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 5, 0),
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	eventHandler.CreateEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var eventResponse responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&eventResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, eventResponse.ID)
	assert.Equal(t, requestBody.EventType.String(), eventResponse.EventType.String())
	assert.Equal(t, requestBody.Description, eventResponse.Description)
	assert.Equal(t, requestBody.Notes, eventResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &requestBody.EventDate, eventResponse.EventDate)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, eventResponse.CreatedDate, time.Second)
	assert.Nil(t, eventResponse.UpdatedDate)
}

func TestCreateEvent_ShouldInsertAndReturnEventWithOnlyRequiredFields(t *testing.T) {
	eventHandler, _ := setupEventHandler(t)

	requestBody := requests.CreateEventRequest{
		EventType: requests.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 5, 0),
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.CreateEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var eventResponse responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&eventResponse)
	assert.NoError(t, err)

	assert.NotNil(t, eventResponse.ID)
	assert.Equal(t, requestBody.EventType.String(), eventResponse.EventType.String())
	assert.Nil(t, eventResponse.Description)
	assert.Nil(t, eventResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &requestBody.EventDate, eventResponse.EventDate)
	assert.NotNil(t, eventResponse.CreatedDate)
	assert.Nil(t, eventResponse.UpdatedDate)
}

func TestCreateEvent_ShouldReturnStatusConflictIfEventIDIsAlreadyInDB(t *testing.T) {
	eventHandler, eventRepository := setupEventHandler(t)

	var id = uuid.New()

	repositoryhelpers.CreateEvent(t, eventRepository, &id, nil, nil)

	requestBody := requests.CreateEventRequest{
		ID:        &id,
		EventType: requests.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 5, 0),
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	eventHandler.CreateEvent(responseRecorder, request)
	assert.Equal(t, http.StatusConflict, responseRecorder.Code)

	expectedError := "Conflict error on insert: ID already exists\n"
	assert.Equal(t, expectedError, responseRecorder.Body.String())
}

// -------- GetAllEvents tests: --------

func TestGetAllEvents_ShouldReturnAllEvents(t *testing.T) {
	eventHandler, eventRepository := setupEventHandler(t)

	// insert events

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	event1, err := eventRepository.Create(&createEvent1)
	assert.NoError(t, err)
	assert.NotNil(t, event1)

	event2ID := uuid.New()
	event2EventDate := time.Now().AddDate(0, 15, 0)
	repositoryhelpers.CreateEvent(t, eventRepository, &event2ID, nil, &event2EventDate)

	// get all events:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, event2ID, response[0].ID)
	assert.NotNil(t, response[0].EventType)
	testutil.AssertEqualFormattedDateTimes(t, &event2EventDate, response[0].EventDate)
	assert.NotNil(t, response[0].CreatedDate)
	assert.Nil(t, response[0].UpdatedDate)

	assert.Equal(t, event1.ID, response[1].ID)
	assert.Equal(t, event1.EventType.String(), response[1].EventType.String())
	assert.Equal(t, event1.Description, response[1].Description)
	assert.Equal(t, event1.Notes, response[1].Notes)
	testutil.AssertEqualFormattedDateTimes(t, event1.EventDate, response[1].EventDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.CreatedDate, response[1].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.UpdatedDate, response[1].UpdatedDate)
}

func TestGetAllEvents_ShouldReturnEmptyResponseIfNoEventsInDatabase(t *testing.T) {
	eventHandler, _ := setupEventHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Len(t, response, 0)
}

// -------- UpdateEvent tests: --------

func TestUpdateEvent_ShouldUpdateEvent(t *testing.T) {
	eventHandler, eventRepository := setupEventHandler(t)

	// create an event

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	event, err := eventRepository.Create(&createEvent)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	// update the event

	var updatedEventType requests.EventType = requests.EventTypeOffer
	updateBody := requests.UpdateEventRequest{
		ID:          *createEvent.ID,
		EventType:   &updatedEventType,
		Description: testutil.ToPtr("Updated Description"),
		Notes:       testutil.ToPtr("Updated Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(4, 0, 0)),
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/event/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now()
	eventHandler.UpdateEvent(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the event by ID

	updatedEvent, err := eventRepository.GetByID(createEvent.ID)
	assert.NoError(t, err)

	assert.Equal(t, updateBody.ID, updatedEvent.ID)
	assert.Equal(t, updateBody.EventType.String(), updatedEvent.EventType.String())
	assert.Equal(t, updateBody.Description, updatedEvent.Description)
	assert.Equal(t, updateBody.Notes, updatedEvent.Notes)
	testutil.AssertEqualFormattedDateTimes(t, updateBody.EventDate, updatedEvent.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, updatedEvent.CreatedDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, updatedEvent.UpdatedDate, time.Second)

}

func TestUpdateEvent_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	eventHandler, eventRepository := setupEventHandler(t)

	// create an event

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	// update the event
	updateBody := requests.UpdateEventRequest{
		ID: event.ID,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/event/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.UpdateEvent(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(
		t,
		"Unable to convert request to internal model: validation error: nothing to update\n",
		responseBodyString)
}

// -------- DeleteEvent tests: --------

func TestDeleteEvent_ShouldDeleteEvent(t *testing.T) {
	eventHandler, eventRepository := setupEventHandler(t)

	// insert an event

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	// delete the event

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/event/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": event.ID.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	eventHandler.DeleteEvent(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusOK, deleteResponseRecorder.Code)

	// try to get the event

	nilEvent, err := eventRepository.GetByID(&event.ID)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+event.ID.String()+"'", notFoundError.Error())
}

func TestDeleteEvent_ShouldReturnStatusNotFoundIfEventDoesNotExist(t *testing.T) {
	eventHandler, _ := setupEventHandler(t)

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/event/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": uuid.New().String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	eventHandler.DeleteEvent(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusNotFound, deleteResponseRecorder.Code)
}
