package responses

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewEventDTO tests: --------

func TestNewEventDTO_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied
	model := models.Event{
		ID:          uuid.New(),
		EventType:   &eventType,
		Description: testutil.ToPtr("Description"),
		Notes:       testutil.ToPtr("Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
	}

	eventDTO, err := NewEventDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventDTO)

	assert.Equal(t, model.ID, *eventDTO.ID)
	assert.Equal(t, model.EventType.String(), eventDTO.EventType.String())
	assert.Equal(t, model.Description, eventDTO.Description)
	assert.Equal(t, model.Notes, eventDTO.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.EventDate, eventDTO.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, eventDTO.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.UpdatedDate, eventDTO.UpdatedDate)
}

func TestNewEventDTO_ShouldWorkWithOnlyID(t *testing.T) {
	var model = models.Event{
		ID: uuid.New(),
	}
	eventDTO, err := NewEventDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventDTO)
	assert.Equal(t, model.ID, *eventDTO.ID)
}

func TestNewEventDTO_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilDTO, err := NewEventDTO(nil)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, internalServiceError.Error(), "internal service error: Error building DTO: Event is nil")
}

func TestNewEventDTO_ShouldReturnInternalServiceErrorIfEventTypeIsEmpty(t *testing.T) {
	var empty models.EventType = ""
	emptyEventType := models.Event{
		ID:        uuid.New(),
		EventType: &empty,
	}
	nilDTO, err := NewEventDTO(&emptyEventType)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: ''",
		internalServiceError.Error())
}

func TestNewEventDTO_ShouldReturnInternalServiceErrorIfEventTypeIsInvalid(t *testing.T) {
	var invalid models.EventType = "hiringPaused"
	emptyEventType := models.Event{
		ID:        uuid.New(),
		EventType: &invalid,
	}
	nilDTO, err := NewEventDTO(&emptyEventType)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: 'hiringPaused'",
		internalServiceError.Error())
}

// -------- NewEventDTOs tests: --------
func TestNewEventDTOs_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied

	eventModels := []*models.Event{
		{
			ID:          uuid.New(),
			EventType:   &eventType,
			Description: testutil.ToPtr("Description"),
			Notes:       testutil.ToPtr("Notes"),
			EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
			UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
		},
		{
			ID: uuid.New(),
		},
	}

	eventDTOs, err := NewEventDTOs(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, eventDTOs)
	assert.Len(t, eventDTOs, 2)
}

func TestNewEventDTOs_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewEventDTOs(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewEventDTOs_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var eventModels []*models.Event
	response, err := NewEventDTOs(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewEventDTOs_ShouldReturnNilIfOneEventTypeIsInvalid(t *testing.T) {
	var eventTypeApplied models.EventType = models.EventTypeApplied
	var eventTypeEmpty models.EventType = ""

	eventModels := []*models.Event{
		{
			ID:        uuid.New(),
			EventType: &eventTypeApplied,
		},
		{
			ID:        uuid.New(),
			EventType: &eventTypeEmpty,
		},
	}
	nilDTOs, err := NewEventDTOs(eventModels)
	assert.Nil(t, nilDTOs)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: ''",
		internalServiceError.Error())
}

// -------- NewEventResponse tests: --------

func TestNewEventResponse_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied
	model := models.Event{
		ID:          uuid.New(),
		EventType:   &eventType,
		Description: testutil.ToPtr("Description"),
		Notes:       testutil.ToPtr("Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
	}

	eventResponse, err := NewEventResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventResponse)

	assert.Equal(t, model.ID, *eventResponse.ID)
	assert.Equal(t, model.EventType.String(), eventResponse.EventType.String())
	assert.Equal(t, model.Description, eventResponse.Description)
	assert.Equal(t, model.Notes, eventResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.EventDate, eventResponse.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, eventResponse.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.UpdatedDate, eventResponse.UpdatedDate)
}

func TestNewEventResponse_ShouldWorkWithOnlyID(t *testing.T) {
	var model = models.Event{
		ID: uuid.New(),
	}
	eventResponse, err := NewEventResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventResponse)
	assert.Equal(t, model.ID, *eventResponse.ID)
}

func TestNewEventResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilResponse, err := NewEventResponse(nil)
	assert.Nil(t, nilResponse)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, internalServiceError.Error(), "internal service error: Error building response: Event is nil")
}

// -------- NewEventsResponse tests: --------

func TestNewEventsResponse_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied

	eventModels := []*models.Event{
		{
			ID:          uuid.New(),
			EventType:   &eventType,
			Description: testutil.ToPtr("Description"),
			Notes:       testutil.ToPtr("Notes"),
			EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
			UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
		},
		{
			ID: uuid.New(),
		},
	}

	EventsResponse, err := NewEventsResponse(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, EventsResponse)
	assert.Len(t, EventsResponse, 2)
}

func TestNewEventsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewEventsResponse(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewEventsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var eventModels []*models.Event
	response, err := NewEventsResponse(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
