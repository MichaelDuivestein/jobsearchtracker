package repositories

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCompanyPersonGetByID_ShouldReturnValidationError(t *testing.T) {
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
			repository := NewCompanyPersonRepository(nil)

			personCompanies, err := repository.GetByID(test.companyID, test.personID)
			assert.Nil(t, personCompanies)

			assert.Error(t, err)
			assert.Equal(t, internalErrors.NewValidationError(nil, "companyID and personID cannot both be empty"), err)
		})
	}
}
