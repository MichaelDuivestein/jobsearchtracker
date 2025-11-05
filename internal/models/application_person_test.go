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

// -------- AssociateApplicationPerson.Validate tests: --------

func TestAssociateApplicationPersonValidate_ShouldReturnNilIfAssociateApplicationPersonIsValid(t *testing.T) {
	model := AssociateApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateApplicationPersonValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	model := AssociateApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateApplicationPersonValidate_ShouldReturnValidationErrorIfApplicationIDIsEmpty(t *testing.T) {
	var ApplicationID uuid.UUID
	model := AssociateApplicationPerson{
		ApplicationID: ApplicationID,
		PersonID:      uuid.New(),
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: ApplicationID is empty", validationError.Error())
}

func TestAssociateApplicationPersonValidate_ShouldReturnValidationErrorIfPersonIDIsEmpty(t *testing.T) {
	var personID uuid.UUID
	model := AssociateApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      personID,
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: PersonID is empty", validationError.Error())
}

// -------- DeleteApplicationPerson.Validate tests: --------

func TestDeleteApplicationPersonValidate_ShouldReturnNilIfAssociateApplicationPersonIsValid(t *testing.T) {
	model := DeleteApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestDeleteApplicationPersonValidate_ShouldReturnValidationErrorIfApplicationIDIsEmpty(t *testing.T) {
	var ApplicationID uuid.UUID
	model := DeleteApplicationPerson{
		ApplicationID: ApplicationID,
		PersonID:      uuid.New(),
	}
	err := model.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: ApplicationID cannot be empty", validationError.Error())
}

func TestDeleteApplicationPersonValidate_ShouldReturnValidationErrorIfPersonIDIsEmpty(t *testing.T) {
	var personID uuid.UUID
	model := DeleteApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      personID,
	}
	err := model.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: PersonID cannot be empty", validationError.Error())
}
