package models

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateEventPerson.Validate tests: --------

func TestAssociateEventPersonValidate_ShouldReturnNilIfAssociateEventPersonIsValid(t *testing.T) {
	model := AssociateEventPerson{
		EventID:     uuid.New(),
		PersonID:    uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateEventPersonValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	model := AssociateEventPerson{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateEventPersonValidate_ShouldReturnValidationErrorIfEventIDIsEmpty(t *testing.T) {
	var EventID uuid.UUID
	model := AssociateEventPerson{
		EventID:     EventID,
		PersonID:    uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: EventID is empty", validationError.Error())
}

func TestAssociateEventPersonValidate_ShouldReturnValidationErrorIfPersonIDIsEmpty(t *testing.T) {
	var personID uuid.UUID
	model := AssociateEventPerson{
		EventID:     uuid.New(),
		PersonID:    personID,
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: PersonID is empty", validationError.Error())
}

// -------- DeleteEventPerson.Validate tests: --------

func TestDeleteEventPersonValidate_ShouldReturnNilIfAssociateEventPersonIsValid(t *testing.T) {
	model := DeleteEventPerson{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestDeleteEventPersonValidate_ShouldReturnValidationErrorIfEventIDIsEmpty(t *testing.T) {
	var EventID uuid.UUID
	model := DeleteEventPerson{
		EventID:  EventID,
		PersonID: uuid.New(),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: EventID cannot be empty", validationError.Error())
}

func TestDeleteEventPersonValidate_ShouldReturnValidationErrorIfPersonIDIsEmpty(t *testing.T) {
	var personID uuid.UUID
	model := DeleteEventPerson{
		EventID:  uuid.New(),
		PersonID: personID,
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: PersonID cannot be empty", validationError.Error())
}
