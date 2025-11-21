package repositories

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCompanyEventGetByID_ShouldReturnValidationError(t *testing.T) {
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
			repository := NewCompanyEventRepository(nil)

			eventCompanies, err := repository.GetByID(test.companyID, test.eventID)
			assert.Nil(t, eventCompanies)

			assert.Error(t, err)
			assert.Equal(
				t,
				internalErrors.NewValidationError(nil, "companyID and eventID cannot both be empty"),
				err)
		})
	}
}
