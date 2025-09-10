package repositories_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupPersonRepository(t *testing.T) *repositories.PersonRepository {
	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupPersonRepositoryTestContainer(t, *config)

	var personRepository *repositories.PersonRepository
	err := container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	return personRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertPerson(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()
	email := "some@email.tld"
	phone := "123456"
	notes := "Some Notes"
	createdDate := time.Now().AddDate(0, 0, -2)
	updatedDate := time.Now().AddDate(0, 0, -1)

	person := models.CreatePerson{
		ID:          &id,
		Name:        "Person Name",
		PersonType:  models.PersonTypeDeveloper,
		Email:       &email,
		Phone:       &phone,
		Notes:       &notes,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	insertedPerson, err := personRepository.Create(&person)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.Equal(t, *person.ID, insertedPerson.ID)
	assert.Equal(t, person.Name, insertedPerson.Name)
	assert.Equal(t, person.PersonType, insertedPerson.PersonType)
	assert.Equal(t, person.Email, insertedPerson.Email)
	assert.Equal(t, person.Phone, insertedPerson.Phone)
	assert.Equal(t, person.Notes, insertedPerson.Notes)

	personToInsertCreatedDate := person.CreatedDate.Format(time.RFC3339)
	insertedPersonCreatedDate := insertedPerson.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, personToInsertCreatedDate, insertedPersonCreatedDate)

	personToInsertUpdatedDate := person.UpdatedDate.Format(time.RFC3339)
	insertedPersonUpdatedDate := insertedPerson.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, personToInsertUpdatedDate, insertedPersonUpdatedDate)
}

func TestCreate_ShouldInsertPersonWithMinimumRequiredFields(t *testing.T) {
	personRepository := setupPersonRepository(t)

	person := models.CreatePerson{
		Name:       "Abc Def",
		PersonType: models.PersonTypeCEO,
	}

	createdDateApproximation := time.Now().Format(time.RFC3339)
	insertedPerson, err := personRepository.Create(&person)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.NotNil(t, insertedPerson.ID)
	assert.NotNil(t, insertedPerson.Name)
	assert.NotNil(t, insertedPerson.PersonType)
	assert.Nil(t, insertedPerson.Email)
	assert.Nil(t, insertedPerson.Phone)
	assert.Nil(t, insertedPerson.Notes)

	insertedPersonCreatedDate := insertedPerson.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, insertedPersonCreatedDate)

	assert.Nil(t, insertedPerson.UpdatedDate)
}

func TestCreate_ShouldReturnConflictErrorOnDuplicatePersonId(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()

	person1 := models.CreatePerson{
		ID:         &id,
		Name:       "Not Real",
		PersonType: models.PersonTypeJobContact,
	}
	insertedPerson1, err := personRepository.Create(&person1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson1)
	assert.NotNil(t, insertedPerson1.ID)

	person2 := models.CreatePerson{
		ID:         &id,
		Name:       "Never Duplicated",
		PersonType: models.PersonTypeJobAdvertiser,
	}
	insertedPerson2, err := personRepository.Create(&person2)
	assert.Nil(t, insertedPerson2)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(t,
		"conflict error on insert: ID already exists in database: '"+id.String()+"'",
		err.Error())
}

// -------- GetById tests: --------

func TestGetById_ShouldGetPerson(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()
	personToInsert := models.CreatePerson{
		ID:         &id,
		Name:       "Joe Sparks",
		PersonType: models.PersonTypeDeveloper,
	}
	insertedPerson, err := personRepository.Create(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	retrievedPerson, err := personRepository.GetById(&id)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedPerson)

	assert.Equal(t, id, retrievedPerson.ID)
}

func TestGetById_ShouldReturnValidationErrorIfPersonIDIsNil(t *testing.T) {
	personRepository := setupPersonRepository(t)

	person, err := personRepository.GetById(nil)
	assert.Nil(t, person)
	assert.NotNil(t, err)
	assert.Equal(t, "validation error on field 'ID': ID is nil", err.Error())
}

func TestGetById_ShouldReturnNotFoundErrorIfPersonIDDoesNotExist(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()

	person, err := personRepository.GetById(&id)
	assert.Nil(t, person)
	assert.NotNil(t, err)
	assert.Equal(t,
		"error: object not found: ID: '"+id.String()+"'",
		err.Error(),
		"Wrong error returned")
}
