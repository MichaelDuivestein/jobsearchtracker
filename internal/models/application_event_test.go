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

// -------- AssociateApplicationEvent.Validate tests: --------

func TestAssociateApplicationEventValidate_ShouldReturnNilIfAssociateApplicationEventIsValid(t *testing.T) {
	model := AssociateApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateApplicationEventValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	model := AssociateApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateApplicationEventValidate_ShouldReturnValidationErrorIfApplicationIDIsEmpty(t *testing.T) {
	var ApplicationID uuid.UUID
	model := AssociateApplicationEvent{
		ApplicationID: ApplicationID,
		EventID:       uuid.New(),
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: ApplicationID is empty", validationError.Error())
}

func TestAssociateApplicationEventValidate_ShouldReturnValidationErrorIfEventIDIsEmpty(t *testing.T) {
	var eventID uuid.UUID
	model := AssociateApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       eventID,
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: EventID is empty", validationError.Error())
}

// -------- DeleteApplicationEvent.Validate tests: --------

func TestDeleteApplicationEventValidate_ShouldReturnNilIfAssociateApplicationEventIsValid(t *testing.T) {
	model := DeleteApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestDeleteApplicationEventValidate_ShouldReturnValidationErrorIfApplicationIDIsEmpty(t *testing.T) {
	var ApplicationID uuid.UUID
	model := DeleteApplicationEvent{
		ApplicationID: ApplicationID,
		EventID:       uuid.New(),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: ApplicationID cannot be empty", validationError.Error())
}

func TestDeleteApplicationEventValidate_ShouldReturnValidationErrorIfEventIDIsEmpty(t *testing.T) {
	var eventID uuid.UUID
	model := DeleteApplicationEvent{
		ApplicationID: uuid.New(),
		EventID:       eventID,
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: EventID cannot be empty", validationError.Error())
}
