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

// -------- AssociateCompanyPerson.Validate tests: --------

func TestAssociateCompanyPersonValidate_ShouldReturnNilIfAssociateCompanyPersonIsValid(t *testing.T) {
	model := AssociateCompanyPerson{
		CompanyID:   uuid.New(),
		PersonID:    uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now()),
	}

	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateCompanyPersonValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	model := AssociateCompanyPerson{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	err := model.Validate()
	assert.NoError(t, err)
}

func TestAssociateCompanyPersonValidate_ShouldReturnValidationErrorIfCompanyIDIsEmpty(t *testing.T) {
	var companyID uuid.UUID
	model := AssociateCompanyPerson{
		CompanyID:   companyID,
		PersonID:    uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now()),
	}

	err := model.Validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: CompanyID is empty", validationErr.Error())

}

func TestAssociateCompanyPersonValidate_ShouldReturnValidationErrorIfPersonIDIsEmpty(t *testing.T) {
	var personID uuid.UUID
	model := AssociateCompanyPerson{
		CompanyID:   uuid.New(),
		PersonID:    personID,
		CreatedDate: testutil.ToPtr(time.Now()),
	}

	err := model.Validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: PersonID is empty", validationErr.Error())

}

// -------- DeleteCompanyPerson.Validate tests: --------

func TestDeleteCompanyPersonValidate_ShouldReturnNilIfAssociateCompanyPersonIsValid(t *testing.T) {
	model := DeleteCompanyPerson{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	err := model.Validate()
	assert.NoError(t, err)
}

func TestDeleteCompanyPersonValidate_ShouldReturnValidationErrorIfCompanyIDIsEmpty(t *testing.T) {
	var companyID uuid.UUID
	model := DeleteCompanyPerson{
		CompanyID: companyID,
		PersonID:  uuid.New(),
	}

	err := model.Validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: CompanyID cannot be empty", validationErr.Error())

}

func TestDeleteCompanyPersonValidate_ShouldReturnValidationErrorIfPersonIDIsEmpty(t *testing.T) {
	var personID uuid.UUID
	model := DeleteCompanyPerson{
		CompanyID: uuid.New(),
		PersonID:  personID,
	}

	err := model.Validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: PersonID cannot be empty", validationErr.Error())

}
