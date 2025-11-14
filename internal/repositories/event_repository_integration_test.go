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

func setupEventRepository(t *testing.T) *repositories.EventRepository {
	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupEventRepositoryTestContainer(t, *config)

	var eventRepository *repositories.EventRepository
	err := container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	return eventRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertEvent(t *testing.T) {
	eventRepository := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	insertedEvent, err := eventRepository.Create(&createEvent)
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
	eventRepository := setupEventRepository(t)

	createEvent := models.CreateEvent{
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	createdDateApproximation := time.Now()

	insertedEvent, err := eventRepository.Create(&createEvent)
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

func TestCreate_ShouldReturnConflictErrorOnDuplicateEventID(t *testing.T) {
	eventRepository := setupEventRepository(t)

	id := uuid.New()

	event1 := models.CreateEvent{
		ID:        &id,
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	_, err := eventRepository.Create(&event1)
	assert.NoError(t, err)

	event2 := models.CreateEvent{
		ID:        &id,
		EventType: models.EventTypeOffer,
		EventDate: time.Now().AddDate(0, 3, 0),
	}
	nilEvent, err := eventRepository.Create(&event2)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(t,
		"conflict error on insert: ID already exists in database: '"+id.String()+"'",
		conflictError.Error())
}

// -------- GetByID tests: --------

func TestGetByID_ShouldGetEvent(t *testing.T) {
	eventRepository := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 7, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 6, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 5, 0)),
	}
	_, err := eventRepository.Create(&createEvent)
	assert.NoError(t, err)

	event, err := eventRepository.GetByID(createEvent.ID)
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

func TestGetByID_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventRepository := setupEventRepository(t)

	id := uuid.New()
	nilEvent, err := eventRepository.GetByID(&id)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ID: '"+id.String()+"'",
		notFoundError.Error())
}

// -------- GetAll tests: --------

func TestGetAll_ShouldReturnAllEvents(t *testing.T) {
	eventRepository := setupEventRepository(t)

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventRepository.Create(&createEvent1)
	assert.NoError(t, err)

	createEvent2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	events, err := eventRepository.GetAll()
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

func TestGetAll_ShouldReturnNilIfNoEventsInDatabase(t *testing.T) {
	eventRepository := setupEventRepository(t)

	events, err := eventRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, events)
}

// -------- Update tests: --------

func TestUpdate_ShouldUpdateEvent(t *testing.T) {
	eventRepository := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	_, err := eventRepository.Create(&createEvent)
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
	err = eventRepository.Update(&updateEvent)
	assert.NoError(t, err)

	event, err := eventRepository.GetByID(createEvent.ID)
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
	eventRepository := setupEventRepository(t)

	createEvent := models.CreateEvent{
		ID:        testutil.ToPtr(uuid.New()),
		EventType: models.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 12, 0),
	}
	_, err := eventRepository.Create(&createEvent)
	assert.NoError(t, err)

	updateEvent := models.UpdateEvent{
		ID:    *createEvent.ID,
		Notes: testutil.ToPtr("New Notes"),
	}
	err = eventRepository.Update(&updateEvent)
	assert.NoError(t, err)

	event, err := eventRepository.GetByID(createEvent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	assert.Equal(t, updateEvent.ID, event.ID)
	assert.Equal(t, updateEvent.Notes, event.Notes)
}

func TestUpdate_ShouldNotReturnErrorIfEventDoesNotExist(t *testing.T) {
	eventRepository := setupEventRepository(t)

	updateEvent := models.UpdateEvent{
		ID:    uuid.New(),
		Notes: testutil.ToPtr("New Notes"),
	}
	err := eventRepository.Update(&updateEvent)
	assert.NoError(t, err)
}

// -------- Delete tests: --------

func TestDelete_ShouldDeleteEvent(t *testing.T) {
	eventRepository := setupEventRepository(t)

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	err := eventRepository.Delete(&eventID)
	assert.NoError(t, err)

	retrievedPerson, err := eventRepository.GetByID(&eventID)
	assert.Nil(t, retrievedPerson)
	assert.Error(t, err)
}

func TestDelete_ShouldReturnNotFoundErrorIfEventIDDoesNotExist(t *testing.T) {
	eventRepository := setupEventRepository(t)

	id := uuid.New()
	err := eventRepository.Delete(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: event does not exist. ID: "+id.String(), notFoundError.Error())
}
