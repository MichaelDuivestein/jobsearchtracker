package responses

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewPersonResponse tests: --------

func TestNewPersonResponse_ShouldWork(t *testing.T) {
	email := "e@m.ai"
	phone := "2345"
	notes := "sdfgkljherwkl"
	updatedDate := time.Now().AddDate(0, 0, 27)

	model := models.Person{
		ID:          uuid.New(),
		Name:        "Blah blah",
		PersonType:  models.PersonTypeJobContact,
		Email:       &email,
		Phone:       &phone,
		Notes:       &notes,
		CreatedDate: time.Now().AddDate(0, -1, 0),
		UpdatedDate: &updatedDate,
	}

	response, err := NewPersonResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.PersonType.String(), response.PersonType.String())
	assert.Equal(t, model.Email, response.Email)
	assert.Equal(t, model.Phone, response.Phone)
	assert.Equal(t, model.Notes, response.Notes)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Equal(t, model.UpdatedDate, response.UpdatedDate)
}

func TestNewPersonResponse_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	model := models.Person{
		ID:          uuid.New(),
		Name:        "Anker",
		PersonType:  models.PersonTypeUnknown,
		CreatedDate: time.Now().AddDate(0, 3, 0),
	}

	response, err := NewPersonResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.PersonType.String(), response.PersonType.String())
	assert.Nil(t, response.Email)
	assert.Nil(t, response.Phone)
	assert.Nil(t, response.Notes)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Nil(t, response.UpdatedDate)
}

func TestNewPersonResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewPersonResponse(nil)
	assert.Nil(t, nilModel)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, err.Error(), "internal service error: Error building response: Person is nil")
}

func TestNewPersonResponse_ShouldReturnInternalServiceErrorIfCompanyTypeIsInvalid(t *testing.T) {
	emptyPersonType := models.Person{
		ID:          uuid.New(),
		Name:        "Dave",
		PersonType:  "",
		CreatedDate: time.Now().AddDate(0, 0, 16),
	}
	emptyResponse, err := NewPersonResponse(&emptyPersonType)
	assert.Nil(t, emptyResponse)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(
		t,
		"internal service error: Error converting internal PersonType to external PersonType: ''",
		err.Error())

	invalidPersonType := models.Person{
		ID:          uuid.New(),
		Name:        "Dave",
		PersonType:  "Blah",
		CreatedDate: time.Now().AddDate(0, 0, 16),
	}
	invalidResponse, err := NewPersonResponse(&invalidPersonType)
	assert.Nil(t, invalidResponse)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(
		t,
		"internal service error: Error converting internal PersonType to external PersonType: 'Blah'",
		err.Error())
}
