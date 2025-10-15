package responses

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewPersonResponse tests: --------

func TestNewPersonResponse_ShouldWork(t *testing.T) {
	var personType models.PersonType = models.PersonTypeJobContact

	model := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Blah blah"),
		PersonType:  &personType,
		Email:       testutil.ToPtr("e@m.ai"),
		Phone:       testutil.ToPtr("2345"),
		Notes:       testutil.ToPtr("sdfgkljherwkl"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -1, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 27)),
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
	var personType models.PersonType = models.PersonTypeUnknown
	model := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Anker"),
		PersonType:  &personType,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
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

func TestNewPersonResponse_ShouldReturnInternalServiceErrorIfPersonTypeIsInvalid(t *testing.T) {
	var personTypeEmpty models.PersonType = ""
	emptyPersonType := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Dave"),
		PersonType:  &personTypeEmpty,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 16)),
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

	var personTypeBlah models.PersonType = "Blah"
	invalidPersonType := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Dave"),
		PersonType:  &personTypeBlah,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 16)),
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

// -------- NewPersonsResponse tests: --------

func TestNewPersonsResponse_ShouldWork(t *testing.T) {
	var personTypeUnknown models.PersonType = models.PersonTypeUnknown
	var personTypeCTO models.PersonType = models.PersonTypeCTO
	personModels := []*models.Person{
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Aaron"),
			PersonType:  &personTypeUnknown,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		},
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Bru"),
			PersonType:  &personTypeCTO,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		},
	}

	persons, err := NewPersonsResponse(personModels)
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)
}

func TestNewPersonsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewPersonsResponse(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewPersonsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var personModels []*models.Person
	response, err := NewPersonsResponse(personModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewPersonsResponse_ShouldReturnEmptySliceIfOnePersonTypeIsInvalid(t *testing.T) {
	var personTypeJobAdvertiser models.PersonType = models.PersonTypeJobAdvertiser
	var personTypeEmpty models.PersonType = ""

	personModels := []*models.Person{
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Sammy"),
			PersonType:  &personTypeJobAdvertiser,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 7)),
		},
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Britt"),
			PersonType:  &personTypeEmpty,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 0)),
		},
	}

	persons, err := NewPersonsResponse(personModels)
	assert.Nil(t, persons)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(
		t,
		"internal service error: Error converting internal PersonType to external PersonType: ''",
		err.Error())
}
