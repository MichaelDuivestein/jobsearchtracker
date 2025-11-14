package repositories

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- GetById tests: --------

func TestGetByID_ShouldReturnValidationErrorIfEventIDIsNil(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	nilEvent, err := eventRepository.GetByID(nil)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

// -------- Update tests: --------

func TestUpdate_ShouldReturnValidationErrorIfNoEventFieldsToUpdate(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	eventToUpdate := &models.UpdateEvent{
		ID: uuid.New(),
	}
	err := eventRepository.Update(eventToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfEventIDIsNil(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	err := eventRepository.Delete(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}
