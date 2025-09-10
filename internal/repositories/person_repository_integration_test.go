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

// -------- GetAllByName tests: --------

func TestGetAllByName_ShouldReturnPerson(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()
	personToInsert := models.CreatePerson{
		ID:         &id,
		Name:       "John Smith",
		PersonType: models.PersonTypeDeveloper,
	}
	insertedPerson, err := personRepository.Create(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	retrievedPersons, err := personRepository.GetAllByName(&insertedPerson.Name)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPersons)
	assert.Equal(t, 1, len(retrievedPersons))

	person := retrievedPersons[0]
	assert.Equal(t, id, person.ID)
	assert.Equal(t, "John Smith", person.Name)

}

func TestGetAllByName_ShouldReturnValidationErrorIfPersonNameIsNil(t *testing.T) {
	personRepository := setupPersonRepository(t)

	retrievedPersons, err := personRepository.GetAllByName(nil)
	assert.Nil(t, retrievedPersons)
	assert.NotNil(t, err)
	assert.Equal(t, "validation error on field 'Name': Name is nil", err.Error())
}

func TestGetAllByName_ShouldReturnNotFoundErrorIfPersonNameDoesNotExist(t *testing.T) {
	personRepository := setupPersonRepository(t)

	name := "Doesnt Exist"

	person, err := personRepository.GetAllByName(&name)
	assert.Nil(t, person)
	assert.NotNil(t, err)
	assert.Equal(t,
		"error: object not found: Name: '"+name+"'",
		err.Error(),
		"Wrong error returned")
}

func TestGetAllByName_ShouldReturnMultiplePersonsWithSameName(t *testing.T) {
	personRepository := setupPersonRepository(t)

	// insert some humans

	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:         &person1ID,
		Name:       "frank john",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson1, err := personRepository.Create(&person1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson1)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:         &person2ID,
		Name:       "Frank Jones",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson2, err := personRepository.Create(&person2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson2)

	person3ID := uuid.New()
	person3 := models.CreatePerson{
		ID:         &person3ID,
		Name:       "Frank John",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson3, err := personRepository.Create(&person3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson3)

	// get humans with name Frank John
	frankJohn := "Frank John"

	retrievedPersons, err := personRepository.GetAllByName(&frankJohn)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPersons)
	assert.Equal(t, 2, len(retrievedPersons))

	foundPerson1 := retrievedPersons[0]
	assert.Equal(t, insertedPerson3.ID, foundPerson1.ID)

	foundPerson2 := retrievedPersons[1]
	assert.Equal(t, insertedPerson1.ID, foundPerson2.ID)
}

func TestGetAllByName_ShouldReturnMultiplePersonsWithSameNamePart(t *testing.T) {
	personRepository := setupPersonRepository(t)

	// insert some humans

	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:         &person1ID,
		Name:       "Anne Gale",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson1, err := personRepository.Create(&person1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson1)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:         &person2ID,
		Name:       "Anna Davies",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson2, err := personRepository.Create(&person2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson2)

	person3ID := uuid.New()
	person3 := models.CreatePerson{
		ID:         &person3ID,
		Name:       "Steven Annerson",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson3, err := personRepository.Create(&person3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson3)

	// get humans containing "ann"
	ann := "ann"

	retrievedPersons, err := personRepository.GetAllByName(&ann)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPersons)
	assert.Equal(t, 3, len(retrievedPersons))

	foundPerson1 := retrievedPersons[0]
	assert.Equal(t, insertedPerson2.ID, foundPerson1.ID)

	foundPerson2 := retrievedPersons[1]
	assert.Equal(t, insertedPerson1.ID, foundPerson2.ID)

	foundPerson3 := retrievedPersons[2]
	assert.Equal(t, insertedPerson3.ID, foundPerson3.ID)
}

// -------- GetAll tests: --------

func TestGetAll_ShouldReturnAllPersons(t *testing.T) {
	personRepository := setupPersonRepository(t)

	// add some humans

	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:         &person1ID,
		Name:       "Frank Jones",
		PersonType: models.PersonTypeDeveloper,
	}
	insertedPerson1, err := personRepository.Create(&person1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson1)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:         &person2ID,
		Name:       "Anne Gale",
		PersonType: models.PersonTypeCTO,
	}
	insertedPerson2, err := personRepository.Create(&person2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson2)

	// get all humans
	persons, err := personRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Equal(t, 2, len(persons))

	assert.Equal(t, person2ID, persons[0].ID)
	assert.Equal(t, person1ID, persons[1].ID)
}

func TestGetAll_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	personRepository := setupPersonRepository(t)

	persons, err := personRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, persons)
}
