package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- AssociateCompanyPerson tests: --------

func TestAssociateCompanyPerson_ShouldReturnValidationErrorIfModelIsNil(t *testing.T) {
	service := NewCompanyPersonService(nil)

	nilCompany, err := service.AssociateCompanyPerson(nil)
	assert.Nil(t, nilCompany)

	var validationError *internalErrors.ValidationError
	assert.ErrorAs(t, err, &validationError)

	assert.Equal(t, validationError.Error(), "validation error: AssociateCompanyPerson model is nil")
}

func TestAssociateCompanyPerson_ShouldReturnValidationErrorIfModelIsInvalid(t *testing.T) {
	tests := []struct {
		testName  string
		companyID *uuid.UUID
		personID  *uuid.UUID
	}{
		{
			testName:  "nil companyID and nil personID ",
			personID:  nil,
			companyID: nil,
		},
		{
			testName:  "empty companyID and nil personID",
			companyID: testutil.ToPtr(uuid.UUID{}),
			personID:  nil,
		},
		{
			testName:  "nil companyID and empty personID",
			companyID: nil,
			personID:  testutil.ToPtr(uuid.UUID{}),
		},
		{
			testName:  "empty companyID and empty personID",
			companyID: testutil.ToPtr(uuid.UUID{}),
			personID:  testutil.ToPtr(uuid.UUID{}),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewCompanyPersonService(nil)

			personCompanies, err := service.GetByID(test.companyID, test.personID)
			assert.Nil(t, personCompanies)

			assert.NotNil(t, err)
			assert.Equal(t, internalErrors.NewValidationError(nil, "companyID and personID cannot both be empty"), err)
		})
	}

}

// -------- GetByID tests: --------

func TestGetByID_ShouldReturnValidationErrorIfCompanyIDAndPersonIDAreEmpty(t *testing.T) {
	tests := []struct {
		testName  string
		companyID *uuid.UUID
		personID  *uuid.UUID
	}{
		{
			testName:  "nil companyID and nil personID ",
			personID:  nil,
			companyID: nil,
		},
		{
			testName:  "empty companyID and nil personID",
			companyID: testutil.ToPtr(uuid.UUID{}),
			personID:  nil,
		},
		{
			testName:  "nil companyID and empty personID",
			companyID: nil,
			personID:  testutil.ToPtr(uuid.UUID{}),
		},
		{
			testName:  "empty companyID and empty personID",
			companyID: testutil.ToPtr(uuid.UUID{}),
			personID:  testutil.ToPtr(uuid.UUID{}),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewCompanyPersonService(nil)

			personCompanies, err := service.GetByID(test.companyID, test.personID)
			assert.Nil(t, personCompanies)

			assert.NotNil(t, err)
			assert.Equal(t, internalErrors.NewValidationError(nil, "companyID and personID cannot both be empty"), err)
		})
	}
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfCompanyIDORPersonIDAreEmpty(t *testing.T) {
	tests := []struct {
		testName  string
		companyID uuid.UUID
		personID  uuid.UUID
		errorText string
	}{
		{
			testName:  "empty CompanyID",
			companyID: uuid.UUID{},
			personID:  uuid.New(),
			errorText: "validation error: CompanyID cannot be empty",
		},
		{
			testName:  "empty PersonID",
			companyID: uuid.New(),
			personID:  uuid.UUID{},
			errorText: "validation error: PersonID cannot be empty",
		},
		{
			testName:  "empty CompanyID and PersonID",
			companyID: uuid.UUID{},
			personID:  uuid.UUID{},
			errorText: "validation error: CompanyID cannot be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service := NewCompanyPersonService(nil)

			deleteModel := models.DeleteCompanyPerson{
				CompanyID: test.companyID,
				PersonID:  test.personID,
			}

			err := service.Delete(&deleteModel)
			assert.NotNil(t, err)

			var validationError *internalErrors.ValidationError
			assert.ErrorAs(t, err, &validationError)

			assert.Equal(t, test.errorText, validationError.Error())

		})
	}
}
