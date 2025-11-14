package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateApplicationPerson tests: --------

func TestAssociateApplicationPerson_ShouldReturnValidationErrorIfModelIsNil(t *testing.T) {
	service := NewApplicationPersonService(nil)

	nilApplication, err := service.AssociateApplicationPerson(nil)
	assert.Nil(t, nilApplication)

	var validationError *internalErrors.ValidationError
	assert.ErrorAs(t, err, &validationError)

	assert.Equal(t, validationError.Error(), "validation error: AssociateApplicationPerson model is nil")
}

// -------- GetByID tests: --------

func TestGetByID_ShouldReturnValidationErrorIfApplicationIDAndPersonIDAreEmpty(t *testing.T) {
	tests := []struct {
		testName      string
		applicationID *uuid.UUID
		personID      *uuid.UUID
	}{
		{
			testName:      "nil applicationID and nil personID ",
			personID:      nil,
			applicationID: nil,
		},
		{
			testName:      "empty applicationID and nil personID",
			applicationID: testutil.ToPtr(uuid.UUID{}),
			personID:      nil,
		},
		{
			testName:      "nil applicationID and empty personID",
			applicationID: nil,
			personID:      testutil.ToPtr(uuid.UUID{}),
		},
		{
			testName:      "empty applicationID and empty personID",
			applicationID: testutil.ToPtr(uuid.UUID{}),
			personID:      testutil.ToPtr(uuid.UUID{}),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewApplicationPersonService(nil)

			personCompanies, err := service.GetByID(test.applicationID, test.personID)
			assert.Nil(t, personCompanies)

			assert.Error(t, err)
			assert.Equal(t, internalErrors.NewValidationError(nil, "applicationID and personID cannot both be empty"), err)
		})
	}
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfApplicationIDOrPersonIDisEmpty(t *testing.T) {
	tests := []struct {
		testName      string
		applicationID uuid.UUID
		personID      uuid.UUID
		errorText     string
	}{
		{
			testName:      "empty ApplicationID",
			applicationID: uuid.UUID{},
			personID:      uuid.New(),
			errorText:     "validation error: ApplicationID cannot be empty",
		},
		{
			testName:      "empty PersonID",
			applicationID: uuid.New(),
			personID:      uuid.UUID{},
			errorText:     "validation error: PersonID cannot be empty",
		},
		{
			testName:      "empty ApplicationID and PersonID",
			applicationID: uuid.UUID{},
			personID:      uuid.UUID{},
			errorText:     "validation error: ApplicationID cannot be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewApplicationPersonService(nil)

			deleteModel := models.DeleteApplicationPerson{
				ApplicationID: test.applicationID,
				PersonID:      test.personID,
			}

			err := service.Delete(&deleteModel)
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.ErrorAs(t, err, &validationError)

			assert.Equal(t, test.errorText, validationError.Error())
		})
	}
}
