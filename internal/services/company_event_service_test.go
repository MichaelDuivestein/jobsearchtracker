package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateCompanyEvent tests: --------

func TestAssociateCompanyEvent_ShouldReturnValidationErrorIfModelIsNil(t *testing.T) {
	service := NewCompanyEventService(nil)

	nilCompany, err := service.AssociateCompanyEvent(nil)
	assert.Nil(t, nilCompany)

	var validationError *internalErrors.ValidationError
	assert.ErrorAs(t, err, &validationError)

	assert.Equal(t, validationError.Error(), "validation error: AssociateCompanyEvent model is nil")
}

// -------- GetByID tests: --------

func TestGetByID_ShouldReturnValidationErrorIfCompanyIDAndEventIDAreEmpty(t *testing.T) {
	tests := []struct {
		testName  string
		companyID *uuid.UUID
		eventID   *uuid.UUID
	}{
		{
			testName:  "nil companyID and nil eventID ",
			eventID:   nil,
			companyID: nil,
		},
		{
			testName:  "empty companyID and nil eventID",
			companyID: testutil.ToPtr(uuid.UUID{}),
			eventID:   nil,
		},
		{
			testName:  "nil companyID and empty eventID",
			companyID: nil,
			eventID:   testutil.ToPtr(uuid.UUID{}),
		},
		{
			testName:  "empty companyID and empty eventID",
			companyID: testutil.ToPtr(uuid.UUID{}),
			eventID:   testutil.ToPtr(uuid.UUID{}),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewCompanyEventService(nil)

			eventCompanies, err := service.GetByID(test.companyID, test.eventID)
			assert.Nil(t, eventCompanies)

			assert.Error(t, err)
			assert.Equal(
				t,
				internalErrors.NewValidationError(nil, "companyID and eventID cannot both be empty"),
				err)
		})
	}
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfCompanyIDOrEventIDisEmpty(t *testing.T) {
	tests := []struct {
		testName  string
		companyID uuid.UUID
		eventID   uuid.UUID
		errorText string
	}{
		{
			testName:  "empty CompanyID",
			companyID: uuid.UUID{},
			eventID:   uuid.New(),
			errorText: "validation error: CompanyID cannot be empty",
		},
		{
			testName:  "empty EventID",
			companyID: uuid.New(),
			eventID:   uuid.UUID{},
			errorText: "validation error: EventID cannot be empty",
		},
		{
			testName:  "empty CompanyID and EventID",
			companyID: uuid.UUID{},
			eventID:   uuid.UUID{},
			errorText: "validation error: CompanyID cannot be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewCompanyEventService(nil)

			deleteModel := models.DeleteCompanyEvent{
				CompanyID: test.companyID,
				EventID:   test.eventID,
			}

			err := service.Delete(&deleteModel)
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.ErrorAs(t, err, &validationError)

			assert.Equal(t, test.errorText, validationError.Error())
		})
	}
}
