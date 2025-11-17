package repositories

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEventPersonGetByID_ShouldReturnValidationError(t *testing.T) {
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
			repository := NewEventPersonRepository(nil)

			personCompanies, err := repository.GetByID(test.eventID, test.personID)
			assert.Nil(t, personCompanies)

			assert.Error(t, err)
			assert.Equal(t, internalErrors.NewValidationError(nil, "eventID and personID cannot both be empty"), err)
		})
	}
}
