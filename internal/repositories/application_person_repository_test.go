package repositories

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestApplicationGetByID_ShouldReturnValidationError(t *testing.T) {
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
			repository := NewApplicationPersonRepository(nil)

			personCompanies, err := repository.GetByID(test.applicationID, test.personID)
			assert.Nil(t, personCompanies)

			assert.Error(t, err)
			assert.Equal(t, internalErrors.NewValidationError(nil, "applicationID and personID cannot both be empty"), err)
		})
	}
}
