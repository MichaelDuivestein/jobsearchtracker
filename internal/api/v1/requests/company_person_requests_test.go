package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateCompanyPersonRequest.validate tests: --------

func TestAssociateCompanyPersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := AssociateCompanyPersonRequest{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestAssociateCompanyPersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		companyID            uuid.UUID
		personID             uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid CompanyID",
			uuid.UUID{},
			uuid.New(),
			"validation error: CompanyID is invalid"},
		{
			"invalid PersonID",
			uuid.New(),
			uuid.UUID{},
			"validation error: PersonID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := AssociateCompanyPersonRequest{
				CompanyID: test.companyID,
				PersonID:  test.personID,
			}

			err := request.validate()
			assert.NotNil(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- AssociateCompanyPersonRequest.ToModel tests: --------

func TestAssociateCompanyPersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := AssociateCompanyPersonRequest{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.CompanyID, model.CompanyID)
	assert.Equal(t, request.PersonID, model.PersonID)
	assert.Nil(t, model.CreatedDate)
}

// -------- DeleteCompanyPersonRequest.validate tests: --------

func TestDeleteCompanyPersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := DeleteCompanyPersonRequest{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestDeleteCompanyPersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		companyID            uuid.UUID
		personID             uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid CompanyID",
			uuid.UUID{},
			uuid.New(),
			"validation error: CompanyID is invalid"},
		{
			"invalid PersonID",
			uuid.New(),
			uuid.UUID{},
			"validation error: PersonID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := DeleteCompanyPersonRequest{
				CompanyID: test.companyID,
				PersonID:  test.personID,
			}

			err := request.validate()
			assert.NotNil(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- DeleteCompanyPersonRequest.ToModel tests: --------

func TestADeleteCompanyPersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := DeleteCompanyPersonRequest{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.CompanyID, model.CompanyID)
	assert.Equal(t, request.PersonID, model.PersonID)
}
