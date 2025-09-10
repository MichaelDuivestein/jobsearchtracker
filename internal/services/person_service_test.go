package services

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldReturnValidationErrorOnNilPerson(t *testing.T) {
	personService := NewPersonService(nil)

	nilPerson, err := personService.CreatePerson(nil)
	assert.Nil(t, nilPerson)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CreatePerson is nil", err.Error())
}

func TestCreatePerson_ShouldReturnValidationErrorOnEmptyName(t *testing.T) {
	personService := NewPersonService(nil)

	person := models.CreatePerson{
		Name:       "",
		PersonType: models.PersonTypeDeveloper,
	}

	nilPerson, err := personService.CreatePerson(&person)
	assert.Nil(t, nilPerson)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'Name': person name is empty", err.Error())
}

func TestCreatePerson_ShouldReturnValidationErrorOnInvalidPersonType(t *testing.T) {
	personService := NewPersonService(nil)

	person := models.CreatePerson{
		Name: "Random",
	}

	nilPerson, err := personService.CreatePerson(&person)
	assert.Nil(t, nilPerson)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'PersonType': person type is invalid", err.Error())
}

func TestCreatePerson_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	personService := NewPersonService(nil)

	person := models.CreatePerson{
		Name:        "something here",
		PersonType:  models.PersonTypeUnknown,
		UpdatedDate: &time.Time{},
	}

	nilPerson, err := personService.CreatePerson(&person)
	assert.Nil(t, nilPerson)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t,
		"validation error on field 'UpdatedDate': updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		err.Error())
}

// -------- GetPersonById tests: --------

func TestGetPersonById_ShouldReturnValidationErrorIfPersonIdIsNil(t *testing.T) {
	personService := NewPersonService(nil)

	nilPerson, err := personService.GetPersonById(nil)
	assert.Nil(t, nilPerson)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'person ID': personId is required", err.Error())
}

// -------- GetPersonsByName tests: --------

func TestGetPersonsByName_ShouldReturnValidationErrorIfPersonNameIsNil(t *testing.T) {
	personService := NewPersonService(nil)

	nilPerson, err := personService.GetPersonsByName(nil)
	assert.Nil(t, nilPerson)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'personName': personName is required", err.Error())
}

// -------- UpdatePerson tests: --------

func TestUpdatePerson_ShouldReturnValidationErrorIfPersonIsNil(t *testing.T) {
	personService := NewPersonService(nil)

	err := personService.UpdatePerson(nil)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: UpdatePerson model is nil", err.Error())
}

func TestUpdatePerson_ShouldReturnValidationErrorIfPersonContainsNothingToUpdate(t *testing.T) {
	personService := NewPersonService(nil)

	id := uuid.New()
	person := models.UpdatePerson{
		ID: id,
	}

	err := personService.UpdatePerson(&person)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: nothing to update", err.Error())
}

// -------- DeletePerson tests: --------

func TestDeletePerson_ShouldReturnValidationErrorIfPersonIdIsNil(t *testing.T) {
	personService := NewPersonService(nil)

	err := personService.DeletePerson(nil)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'person ID': personId is required", err.Error())
}
