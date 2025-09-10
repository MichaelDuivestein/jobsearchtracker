package services_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/services"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupPersonService(t *testing.T) *services.PersonService {
	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupPersonServiceTestContainer(t, *config)

	var personService *services.PersonService
	err := container.Invoke(func(personSvc *services.PersonService) {
		personService = personSvc
	})
	assert.NoError(t, err)

	return personService
}

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldWork(t *testing.T) {
	personService := setupPersonService(t)

	id := uuid.New()
	name := "Dude Janesson"
	email := "em@ai.l"
	phone := "321"
	Notes := "Text"
	createdDate := time.Now().AddDate(1, 0, 0)
	updatedDate := time.Now().AddDate(0, -2, 0)

	personToInsert := models.CreatePerson{
		ID:          &id,
		Name:        name,
		PersonType:  models.PersonTypeCEO,
		Email:       &email,
		Phone:       &phone,
		Notes:       &Notes,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	insertedPerson, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.Equal(t, id, insertedPerson.ID)
	assert.Equal(t, name, insertedPerson.Name)
	assert.Equal(t, personToInsert.PersonType, insertedPerson.PersonType)
	assert.Equal(t, &email, insertedPerson.Email)
	assert.Equal(t, &phone, insertedPerson.Phone)

	createdDateToInsert := createdDate.Format(time.RFC3339)
	insertedCreatedDate := insertedPerson.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateToInsert, insertedCreatedDate)

	updatedDateToInsert := updatedDate.Format(time.RFC3339)
	insertedUpdatedDate := insertedPerson.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateToInsert, insertedUpdatedDate)
}

func TestCreatePerson_ShouldHandleEmptyFields(t *testing.T) {
	personService := setupPersonService(t)

	name := "Sven Joe"

	personToInsert := models.CreatePerson{
		Name:       name,
		PersonType: models.PersonTypeCEO,
	}

	insertedDateApproximation := time.Now().Format(time.RFC3339)
	insertedPerson, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.NotNil(t, insertedPerson.ID)
	assert.Equal(t, name, insertedPerson.Name)
	assert.Equal(t, personToInsert.PersonType, insertedPerson.PersonType)
	assert.Nil(t, insertedPerson.Email)
	assert.Nil(t, insertedPerson.Phone)

	insertedCreatedDate := insertedPerson.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, insertedDateApproximation, insertedCreatedDate)

	assert.Nil(t, insertedPerson.UpdatedDate)
}

// -------- GetPersonById tests: --------

func TestGetPersonById_ShouldWork(t *testing.T) {
	personService := setupPersonService(t)

	id := uuid.New()
	name := "Some Name"
	email := "an@email.address"
	phone := "128932019"
	Notes := "No notes here..."
	createdDate := time.Now().AddDate(0, 2, 0)
	updatedDate := time.Now().AddDate(0, -1, 0)

	personToInsert := models.CreatePerson{
		ID:          &id,
		Name:        name,
		PersonType:  models.PersonTypeOther,
		Email:       &email,
		Phone:       &phone,
		Notes:       &Notes,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	insertedPerson, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	retrievedPerson, err := personService.GetPersonById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPerson)

}

func TestGetPersonById_ShouldReturnNotFoundErrorForAnIdThatDoesNotExist(t *testing.T) {
	personService := setupPersonService(t)

	id := uuid.New()
	nilPerson, err := personService.GetPersonById(&id)
	assert.Nil(t, nilPerson)

	assert.NotNil(t, err)
	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
}
