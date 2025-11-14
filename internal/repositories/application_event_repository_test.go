package repositories

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestApplicationEventGetByID_ShouldReturnValidationError(t *testing.T) {
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
			repository := NewApplicationEventRepository(nil)

			eventCompanies, err := repository.GetByID(test.applicationID, test.eventID)
			assert.Nil(t, eventCompanies)

			assert.Error(t, err)
			assert.Equal(
				t,
				internalErrors.NewValidationError(nil, "applicationID and eventID cannot both be empty"),
				err)
		})
	}
}
