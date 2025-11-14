package services

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreateEvent tests: --------

func TestCreateEvent_ShouldReturnValidationErrorOnNilEvent(t *testing.T) {
	eventService := NewEventService(nil)

	nilEvent, err := eventService.CreateEvent(nil)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CreateEvent is nil", validationError.Error())
}

func TestCreateEvent_ShouldReturnValidationErrorOnEmptyEventType(t *testing.T) {
	eventService := NewEventService(nil)

	event := models.CreateEvent{
		EventType: "",
		EventDate: time.Now(),
	}
	nilEvent, err := eventService.CreateEvent(&event)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'eventType': event type is invalid", validationError.Error())
}

func TestCreateEvent_ShouldReturnValidationErrorOnUnsetEventDate(t *testing.T) {
	eventService := NewEventService(nil)

	event := models.CreateEvent{
		EventType: models.EventTypeApplied,
		EventDate: time.Time{},
	}
	nilEvent, err := eventService.CreateEvent(&event)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'eventDate': event date is zero. It should be a recent date",
		validationError.Error(),
	)
}

// -------- GetEventByID tests: --------

func TestGetEventByID_ShouldReturnValidationErrorIfEventIDIsNil(t *testing.T) {
	eventService := NewEventService(nil)

	nilEvent, err := eventService.GetEventByID(nil)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'event ID': eventID is required", validationError.Error())
}

// -------- UpdateEvent tests: --------

func TestUpdateEvent_ShouldReturnValidationErrorIfUpdateEventIsNil(t *testing.T) {
	eventService := NewEventService(nil)

	err := eventService.UpdateEvent(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: UpdateEvent model is nil", validationError.Error())
}

func TestUpdateEvent_ShouldReturnValidationErrorIfNoEventFieldsToUpdate(t *testing.T) {
	eventService := NewEventService(nil)

	eventToUpdate := &models.UpdateEvent{
		ID: uuid.New(),
	}
	err := eventService.UpdateEvent(eventToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

// -------- DeleteEvent tests: --------

func TestDeleteEvent_ShouldReturnValidationErrorIfEventIDIsNil(t *testing.T) {
	eventService := NewEventService(nil)

	err := eventService.DeleteEvent(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'event ID': eventID is required", validationError.Error())
}
