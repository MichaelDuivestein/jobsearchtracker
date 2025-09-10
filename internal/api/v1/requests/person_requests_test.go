package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreatePersonRequest tests: --------

func TestCreatePersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	id := uuid.New()
	email := "no email here"
	phone := "6839023748"
	notes := "Something not noteworthy"

	request := CreatePersonRequest{
		ID:         &id,
		Name:       "Nameless",
		PersonType: PersonTypeDeveloper,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestCreatePersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		name                 string
		personType           PersonType
		expectedErrorMessage string
	}{
		{
			"Empty Name",
			"",
			PersonTypeCTO,
			"validation error on field 'Name': Name is empty"},
		{
			"Empty PersonType",
			"Name present",
			"",
			"validation error on field 'PersonType': PersonType is invalid"},
		{
			"Invalid PersonType",
			"Name present",
			"Invalid PersonType",
			"validation error on field 'PersonType': PersonType is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			email := "name@domain.tld"
			notes := "Lots of text"
			phone := "345023485"

			request := CreatePersonRequest{
				Name:       test.name,
				PersonType: test.personType,
				Email:      &email,
				Phone:      &phone,
				Notes:      &notes,
			}

			err := request.validate()
			assert.NotNil(t, err)

			var validationErr *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationErr))

			assert.Equal(t, test.expectedErrorMessage, err.Error())
		})
	}
}

func TestCreatePersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	id := uuid.New()
	email := "email@email.email"
	phone := "34543534"
	notes := "Blah Blah"

	request := CreatePersonRequest{
		ID:         &id,
		Name:       "Jane Doe",
		PersonType: PersonTypeCEO,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, id.String(), model.ID.String())
	assert.Equal(t, request.Name, model.Name)
	assert.Equal(t, request.PersonType.String(), model.PersonType.String())
	assert.Equal(t, &email, model.Email)
	assert.Equal(t, &phone, model.Phone)
	assert.Equal(t, &notes, model.Notes)
}

func TestCreatePersonRequestToModel_ShouldConvertToModelWithNilValues(t *testing.T) {
	request := CreatePersonRequest{
		Name:       "Jane Doe",
		PersonType: PersonTypeCEO,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Nil(t, model.ID)
	assert.Equal(t, request.Name, model.Name)
	assert.Equal(t, request.PersonType.String(), model.PersonType.String())
	assert.Nil(t, model.Email)
	assert.Nil(t, model.Phone)
	assert.Nil(t, model.Notes)
}
