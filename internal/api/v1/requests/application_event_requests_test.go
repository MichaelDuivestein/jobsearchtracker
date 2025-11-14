package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateApplicationEventRequest.validate tests: --------

func TestAssociateApplicationEventRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := AssociateApplicationEventRequest{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestAssociateApplicationEventRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		applicationID        uuid.UUID
		eventID              uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid ApplicationID",
			uuid.UUID{},
			uuid.New(),
			"validation error: ApplicationID is invalid"},
		{
			"invalid EventID",
			uuid.New(),
			uuid.UUID{},
			"validation error: EventID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := AssociateApplicationEventRequest{
				ApplicationID: test.applicationID,
				EventID:       test.eventID,
			}

			err := request.validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- AssociateApplicationEventRequest.ToModel tests: --------

func TestAssociateApplicationEventRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := AssociateApplicationEventRequest{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.ApplicationID, model.ApplicationID)
	assert.Equal(t, request.EventID, model.EventID)
	assert.Nil(t, model.CreatedDate)
}

// -------- DeleteApplicationEventRequest.validate tests: --------

func TestDeleteApplicationEventRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := DeleteApplicationEventRequest{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestDeleteApplicationEventRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		applicationID        uuid.UUID
		eventID              uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid ApplicationID",
			uuid.UUID{},
			uuid.New(),
			"validation error: ApplicationID is invalid"},
		{
			"invalid EventID",
			uuid.New(),
			uuid.UUID{},
			"validation error: EventID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := DeleteApplicationEventRequest{
				ApplicationID: test.applicationID,
				EventID:       test.eventID,
			}

			err := request.validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- DeleteApplicationEventRequest.ToModel tests: --------

func TestADeleteApplicationEventRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := DeleteApplicationEventRequest{
		ApplicationID: uuid.New(),
		EventID:       uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.ApplicationID, model.ApplicationID)
	assert.Equal(t, request.EventID, model.EventID)
}
