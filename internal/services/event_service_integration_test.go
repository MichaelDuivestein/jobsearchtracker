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

func setupEventService(t *testing.T) (*services.EventService, *repositories.EventRepository) {
	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupEventServiceTestContainer(t, *config)

	var eventService *services.EventService
	err := container.Invoke(func(service *services.EventService) {
		eventService = service
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	return eventService, eventRepository
}

// -------- CreateEvent tests: --------

func TestCreateEvent_ShouldWork(t *testing.T) {
	eventService, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	insertedEvent, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)
	assert.NotNil(t, insertedEvent)

	assert.Equal(t, *createEvent.ID, insertedEvent.ID)
	assert.Equal(t, createEvent.EventType.String(), insertedEvent.EventType.String())
	assert.Equal(t, createEvent.Description, insertedEvent.Description)
	assert.Equal(t, createEvent.Notes, insertedEvent.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent.EventDate, insertedEvent.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, insertedEvent.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.UpdatedDate, insertedEvent.UpdatedDate)
}

func TestCreate_ShouldInsertEventWithOnlyRequiredFields(t *testing.T) {
	eventService, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	createdDateApproximation := time.Now()

	insertedEvent, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)
	assert.NotNil(t, insertedEvent)

	assert.NotNil(t, insertedEvent.ID)
	assert.Equal(t, createEvent.EventType.String(), insertedEvent.EventType.String())
	assert.Nil(t, insertedEvent.Description)
	assert.Nil(t, insertedEvent.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent.EventDate, insertedEvent.EventDate)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, insertedEvent.CreatedDate, time.Second)
	assert.Nil(t, insertedEvent.UpdatedDate)
}

// -------- GetEventByID tests: --------

func TestGetEventByID_ShouldWork(t *testing.T) {
	eventService, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 7, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 6, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 5, 0)),
	}
	_, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)

	event, err := eventService.GetEventByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, *createEvent.ID, event.ID)
	assert.Equal(t, createEvent.EventType.String(), event.EventType.String())
	assert.Equal(t, createEvent.Description, event.Description)
	assert.Equal(t, createEvent.Notes, event.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent.EventDate, event.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, event.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.UpdatedDate, event.UpdatedDate)
}

func TestGetEventByID_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventService, _ := setupEventService(t)

	id := uuid.New()
	nilEvent, err := eventService.GetEventByID(&id)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ID: '"+id.String()+"'",
		notFoundError.Error())
}

// -------- GetAllEvents tests: --------

func TestGetAllEvents_ShouldReturnAllEvents(t *testing.T) {
	eventService, eventRepository := setupEventService(t)

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventService.CreateEvent(&createEvent1)
	assert.NoError(t, err)

	createEvent2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	events, err := eventService.GetAllEvents()
	assert.NoError(t, err)
	assert.NotNil(t, events)
	assert.Equal(t, 2, len(events))

	assert.Equal(t, *createEvent1.ID, events[0].ID)
	assert.Equal(t, createEvent1.EventType.String(), events[0].EventType.String())
	assert.Equal(t, createEvent1.Description, events[0].Description)
	assert.Equal(t, createEvent1.Notes, events[0].Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent1.EventDate, events[0].EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent1.CreatedDate, events[0].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent1.UpdatedDate, events[0].UpdatedDate)

	assert.Equal(t, createEvent2.ID, events[1].ID)
	assert.Equal(t, createEvent2.EventType.String(), events[1].EventType.String())
	assert.Nil(t, createEvent2.Description)
	assert.Nil(t, createEvent2.Notes)
	testutil.AssertEqualFormattedDateTimes(t, createEvent2.EventDate, events[1].EventDate)
	assert.NotNil(t, createEvent2.CreatedDate)
	assert.Nil(t, createEvent2.UpdatedDate)
}

func TestGetAllEvents_ShouldReturnNilIfNoEventsInDatabase(t *testing.T) {
	eventService, _ := setupEventService(t)

	events, err := eventService.GetAllEvents()
	assert.NoError(t, err)
	assert.Nil(t, events)
}

// -------- Update tests: --------

func TestUpdateEvent_ShouldWork(t *testing.T) {
	eventService, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)

	var eventType models.EventType = models.EventTypeCallBooked
	updateEvent := models.UpdateEvent{
		ID:          *createEvent.ID,
		EventType:   &eventType,
		Description: testutil.ToPtr("New Description"),
		Notes:       testutil.ToPtr("New Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, -3, 0)),
	}
	updatedDateApproximation := time.Now()
	err = eventService.UpdateEvent(&updateEvent)
	assert.NoError(t, err)

	event, err := eventService.GetEventByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, updateEvent.ID, event.ID)
	assert.Equal(t, updateEvent.EventType.String(), event.EventType.String())
	assert.Equal(t, updateEvent.Description, event.Description)
	assert.Equal(t, updateEvent.Notes, event.Notes)
	testutil.AssertEqualFormattedDateTimes(t, updateEvent.EventDate, event.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, event.CreatedDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, event.UpdatedDate, time.Second)
}

func TestUpdateEvent_ShouldUpdateASingleField(t *testing.T) {
	eventService, _ := setupEventService(t)

	createEvent := models.CreateEvent{
		ID:        testutil.ToPtr(uuid.New()),
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	_, err := eventService.CreateEvent(&createEvent)
	assert.NoError(t, err)

	updateEvent := models.UpdateEvent{
		ID:    *createEvent.ID,
		Notes: testutil.ToPtr("New Notes"),
	}
	err = eventService.UpdateEvent(&updateEvent)
	assert.NoError(t, err)

	event, err := eventService.GetEventByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, updateEvent.ID, event.ID)
	assert.Equal(t, updateEvent.Notes, event.Notes)
}

func TestUpdateEvent_ShouldNotReturnErrorIfEventDoesNotExist(t *testing.T) {
	eventService, _ := setupEventService(t)

	updateEvent := models.UpdateEvent{
		ID:    uuid.New(),
		Notes: testutil.ToPtr("New Notes"),
	}
	err := eventService.UpdateEvent(&updateEvent)
	assert.NoError(t, err)
}

// -------- DeleteEvent tests: --------

func TestDeleteEvent_ShouldDeleteEvent(t *testing.T) {
	eventService, eventRepository := setupEventService(t)

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	err := eventService.DeleteEvent(&eventID)
	assert.NoError(t, err)

	retrievedPerson, err := eventService.GetEventByID(&eventID)
	assert.Nil(t, retrievedPerson)
	assert.Error(t, err)
}

func TestDeleteEvent_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventService, _ := setupEventService(t)

	id := uuid.New()
	err := eventService.DeleteEvent(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: event does not exist. ID: "+id.String(), notFoundError.Error())
}
