package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateEventPersonRequest.validate tests: --------

func TestAssociateEventPersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := AssociateEventPersonRequest{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestAssociateEventPersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		eventID              uuid.UUID
		personID             uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid EventID",
			uuid.UUID{},
			uuid.New(),
			"validation error: EventID is invalid"},
		{
			"invalid PersonID",
			uuid.New(),
			uuid.UUID{},
			"validation error: PersonID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := AssociateEventPersonRequest{
				EventID:  test.eventID,
				PersonID: test.personID,
			}

			err := request.validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- AssociateEventPersonRequest.ToModel tests: --------

func TestAssociateEventPersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := AssociateEventPersonRequest{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.EventID, model.EventID)
	assert.Equal(t, request.PersonID, model.PersonID)
	assert.Nil(t, model.CreatedDate)
}

// -------- DeleteEventPersonRequest.validate tests: --------

func TestDeleteEventPersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := DeleteEventPersonRequest{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestDeleteEventPersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		eventID              uuid.UUID
		personID             uuid.UUID
		expectedErrorMessage string
	}{
		{
			"invalid EventID",
			uuid.UUID{},
			uuid.New(),
			"validation error: EventID is invalid"},
		{
			"invalid PersonID",
			uuid.New(),
			uuid.UUID{},
			"validation error: PersonID is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			request := DeleteEventPersonRequest{
				EventID:  test.eventID,
				PersonID: test.personID,
			}

			err := request.validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

// -------- DeleteEventPersonRequest.ToModel tests: --------

func TestADeleteEventPersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := DeleteEventPersonRequest{
		EventID:  uuid.New(),
		PersonID: uuid.New(),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.EventID, model.EventID)
	assert.Equal(t, request.PersonID, model.PersonID)
}
