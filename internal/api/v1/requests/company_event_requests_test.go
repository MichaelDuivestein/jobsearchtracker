package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateCompanyEventRequest.validate tests: --------

func TestAssociateCompanyEventRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := AssociateCompanyEventRequest{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestAssociateCompanyEventRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		companyID            uuid.UUID
		eventID              uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid CompanyID",
			uuid.UUID{},
			uuid.New(),
			"validation error: CompanyID is invalid"},
		{
			"invalid EventID",
			uuid.New(),
			uuid.UUID{},
			"validation error: EventID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := AssociateCompanyEventRequest{
				CompanyID: test.companyID,
				EventID:   test.eventID,
			}

			err := request.validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- AssociateCompanyEventRequest.ToModel tests: --------

func TestAssociateCompanyEventRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := AssociateCompanyEventRequest{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.CompanyID, model.CompanyID)
	assert.Equal(t, request.EventID, model.EventID)
	assert.Nil(t, model.CreatedDate)
}

// -------- DeleteCompanyEventRequest.validate tests: --------

func TestDeleteCompanyEventRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := DeleteCompanyEventRequest{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestDeleteCompanyEventRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		companyID            uuid.UUID
		eventID              uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid CompanyID",
			uuid.UUID{},
			uuid.New(),
			"validation error: CompanyID is invalid"},
		{
			"invalid EventID",
			uuid.New(),
			uuid.UUID{},
			"validation error: EventID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := DeleteCompanyEventRequest{
				CompanyID: test.companyID,
				EventID:   test.eventID,
			}

			err := request.validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- DeleteCompanyEventRequest.ToModel tests: --------

func TestADeleteCompanyEventRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := DeleteCompanyEventRequest{
		CompanyID: uuid.New(),
		EventID:   uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.CompanyID, model.CompanyID)
	assert.Equal(t, request.EventID, model.EventID)
}
