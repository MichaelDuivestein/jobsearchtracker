package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateEventPerson tests: --------

func TestAssociateEventPerson_ShouldReturnValidationErrorIfModelIsNil(t *testing.T) {
	service := NewEventPersonService(nil)

	nilEvent, err := service.AssociateEventPerson(nil)
	assert.Nil(t, nilEvent)

	var validationError *internalErrors.ValidationError
	assert.ErrorAs(t, err, &validationError)

	assert.Equal(t, validationError.Error(), "validation error: AssociateEventPerson model is nil")
}

// -------- GetByID tests: --------

func TestGetByID_ShouldReturnValidationErrorIfEventIDAndPersonIDAreEmpty(t *testing.T) {
	tests := []struct {
		testName string
		eventID  *uuid.UUID
		personID *uuid.UUID
	}{
		{
			testName: "nil eventID and nil personID ",
			personID: nil,
			eventID:  nil,
		},
		{
			testName: "empty eventID and nil personID",
			eventID:  testutil.ToPtr(uuid.UUID{}),
			personID: nil,
		},
		{
			testName: "nil eventID and empty personID",
			eventID:  nil,
			personID: testutil.ToPtr(uuid.UUID{}),
		},
		{
			testName: "empty eventID and empty personID",
			eventID:  testutil.ToPtr(uuid.UUID{}),
			personID: testutil.ToPtr(uuid.UUID{}),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewEventPersonService(nil)

			personCompanies, err := service.GetByID(test.eventID, test.personID)
			assert.Nil(t, personCompanies)

			assert.Error(t, err)
			assert.Equal(t, internalErrors.NewValidationError(nil, "eventID and personID cannot both be empty"), err)
		})
	}
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfEventIDOrPersonIDisEmpty(t *testing.T) {
	tests := []struct {
		testName  string
		eventID   uuid.UUID
		personID  uuid.UUID
		errorText string
	}{
		{
			testName:  "empty EventID",
			eventID:   uuid.UUID{},
			personID:  uuid.New(),
			errorText: "validation error: EventID cannot be empty",
		},
		{
			testName:  "empty PersonID",
			eventID:   uuid.New(),
			personID:  uuid.UUID{},
			errorText: "validation error: PersonID cannot be empty",
		},
		{
			testName:  "empty EventID and PersonID",
			eventID:   uuid.UUID{},
			personID:  uuid.UUID{},
			errorText: "validation error: EventID cannot be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewEventPersonService(nil)

			deleteModel := models.DeleteEventPerson{
				EventID:  test.eventID,
				PersonID: test.personID,
			}

			err := service.Delete(&deleteModel)
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.ErrorAs(t, err, &validationError)

			assert.Equal(t, test.errorText, validationError.Error())
		})
	}
}
