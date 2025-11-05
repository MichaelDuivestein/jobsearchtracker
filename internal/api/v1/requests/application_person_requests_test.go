package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateApplicationPersonRequest.validate tests: --------

func TestAssociateApplicationPersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := AssociateApplicationPersonRequest{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestAssociateApplicationPersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		applicationID        uuid.UUID
		personID             uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid ApplicationID",
			uuid.UUID{},
			uuid.New(),
			"validation error: ApplicationID is invalid"},
		{
			"invalid PersonID",
			uuid.New(),
			uuid.UUID{},
			"validation error: PersonID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := AssociateApplicationPersonRequest{
				ApplicationID: test.applicationID,
				PersonID:      test.personID,
			}

			err := request.validate()
			assert.NotNil(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- AssociateApplicationPersonRequest.ToModel tests: --------

func TestAssociateApplicationPersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := AssociateApplicationPersonRequest{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.ApplicationID, model.ApplicationID)
	assert.Equal(t, request.PersonID, model.PersonID)
	assert.Nil(t, model.CreatedDate)
}

// -------- DeleteApplicationPersonRequest.validate tests: --------

func TestDeleteApplicationPersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := DeleteApplicationPersonRequest{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestDeleteApplicationPersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		applicationID        uuid.UUID
		personID             uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid ApplicationID",
			uuid.UUID{},
			uuid.New(),
			"validation error: ApplicationID is invalid"},
		{
			"invalid PersonID",
			uuid.New(),
			uuid.UUID{},
			"validation error: PersonID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := DeleteApplicationPersonRequest{
				ApplicationID: test.applicationID,
				PersonID:      test.personID,
			}

			err := request.validate()
			assert.NotNil(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- DeleteApplicationPersonRequest.ToModel tests: --------

func TestADeleteApplicationPersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := DeleteApplicationPersonRequest{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.ApplicationID, model.ApplicationID)
	assert.Equal(t, request.PersonID, model.PersonID)
}
