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

// -------- AssociateCompanyEvent.Validate tests: --------

func TestAssociateCompanyEventValidate_ShouldReturnNilIfAssociateCompanyEventIsValid(t *testing.T) {
	model := AssociateCompanyEvent{
		CompanyID:   uuid.New(),
		EventID:     uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateCompanyEventValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	model := AssociateCompanyEvent{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateCompanyEventValidate_ShouldReturnValidationErrorIfCompanyIDIsEmpty(t *testing.T) {
	var CompanyID uuid.UUID
	model := AssociateCompanyEvent{
		CompanyID:   CompanyID,
		EventID:     uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CompanyID is empty", validationError.Error())
}

func TestAssociateCompanyEventValidate_ShouldReturnValidationErrorIfEventIDIsEmpty(t *testing.T) {
	var eventID uuid.UUID
	model := AssociateCompanyEvent{
		CompanyID:   uuid.New(),
		EventID:     eventID,
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: EventID is empty", validationError.Error())
}

// -------- DeleteCompanyEvent.Validate tests: --------

func TestDeleteCompanyEventValidate_ShouldReturnNilIfAssociateCompanyEventIsValid(t *testing.T) {
	model := DeleteCompanyEvent{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}
	err := model.Validate()
	assert.NoError(t, err)
}

func TestDeleteCompanyEventValidate_ShouldReturnValidationErrorIfCompanyIDIsEmpty(t *testing.T) {
	var CompanyID uuid.UUID
	model := DeleteCompanyEvent{
		CompanyID: CompanyID,
		EventID:   uuid.New(),
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CompanyID cannot be empty", validationError.Error())
}

func TestDeleteCompanyEventValidate_ShouldReturnValidationErrorIfEventIDIsEmpty(t *testing.T) {
	var eventID uuid.UUID
	model := DeleteCompanyEvent{
		CompanyID: uuid.New(),
		EventID:   eventID,
	}
	err := model.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: EventID cannot be empty", validationError.Error())
}
