package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateApplicationEvent tests: --------

func TestAssociateApplicationEvent_ShouldReturnValidationErrorIfModelIsNil(t *testing.T) {
	service := NewApplicationEventService(nil)

	nilApplication, err := service.AssociateApplicationEvent(nil)
	assert.Nil(t, nilApplication)

	var validationError *internalErrors.ValidationError
	assert.ErrorAs(t, err, &validationError)

	assert.Equal(t, validationError.Error(), "validation error: AssociateApplicationEvent model is nil")
}

// -------- GetByID tests: --------

func TestGetByID_ShouldReturnValidationErrorIfApplicationIDAndEventIDAreEmpty(t *testing.T) {
	tests := []struct {
		testName      string
		applicationID *uuid.UUID
		eventID       *uuid.UUID
	}{
		{
			testName:      "nil applicationID and nil eventID ",
			eventID:       nil,
			applicationID: nil,
		},
		{
			testName:      "empty applicationID and nil eventID",
			applicationID: testutil.ToPtr(uuid.UUID{}),
			eventID:       nil,
		},
		{
			testName:      "nil applicationID and empty eventID",
			applicationID: nil,
			eventID:       testutil.ToPtr(uuid.UUID{}),
		},
		{
			testName:      "empty applicationID and empty eventID",
			applicationID: testutil.ToPtr(uuid.UUID{}),
			eventID:       testutil.ToPtr(uuid.UUID{}),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewApplicationEventService(nil)

			eventCompanies, err := service.GetByID(test.applicationID, test.eventID)
			assert.Nil(t, eventCompanies)

			assert.Error(t, err)
			assert.Equal(
				t,
				internalErrors.NewValidationError(nil, "applicationID and eventID cannot both be empty"),
				err)
		})
	}
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfApplicationIDOrEventIDisEmpty(t *testing.T) {
	tests := []struct {
		testName      string
		applicationID uuid.UUID
		eventID       uuid.UUID
		errorText     string
	}{
		{
			testName:      "empty ApplicationID",
			applicationID: uuid.UUID{},
			eventID:       uuid.New(),
			errorText:     "validation error: ApplicationID cannot be empty",
		},
		{
			testName:      "empty EventID",
			applicationID: uuid.New(),
			eventID:       uuid.UUID{},
			errorText:     "validation error: EventID cannot be empty",
		},
		{
			testName:      "empty ApplicationID and EventID",
			applicationID: uuid.UUID{},
			eventID:       uuid.UUID{},
			errorText:     "validation error: ApplicationID cannot be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewApplicationEventService(nil)

			deleteModel := models.DeleteApplicationEvent{
				ApplicationID: test.applicationID,
				EventID:       test.eventID,
			}

			err := service.Delete(&deleteModel)
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.ErrorAs(t, err, &validationError)

			assert.Equal(t, test.errorText, validationError.Error())
		})
	}
}
