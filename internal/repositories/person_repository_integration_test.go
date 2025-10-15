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
	assert.Equal(t, person.Name, *insertedPerson.Name)
	assert.Equal(t, person.PersonType.String(), insertedPerson.PersonType.String())
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
	assert.NoError(t, err)
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

	retrievedPersons, err := personRepository.GetAllByName(insertedPerson.Name)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPersons)
	assert.Len(t, retrievedPersons, 1)

	person := retrievedPersons[0]
	assert.Equal(t, id, person.ID)
	assert.Equal(t, "John Smith", *person.Name)

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
	assert.Len(t, retrievedPersons, 2)

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
	assert.Len(t, retrievedPersons, 3)

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
	assert.Len(t, persons, 2)

	assert.Equal(t, person2ID, persons[0].ID)
	assert.Equal(t, person1ID, persons[1].ID)
}

func TestGetAll_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	personRepository := setupPersonRepository(t)

	persons, err := personRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

// -------- Update tests: --------

func TestUpdate_ShouldUpdatePerson(t *testing.T) {
	personRepository := setupPersonRepository(t)

	// create a person
	id := uuid.New()
	personToInsert := models.CreatePerson{
		ID:         &id,
		Name:       "Arr Grr",
		PersonType: models.PersonTypeOther,
	}
	insertedPerson, err := personRepository.Create(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	name := "Another Name"
	personType := models.PersonType(models.PersonTypeHR)
	email := "a@b.c"
	phone := "312765"
	notes := "Something noteworthy"

	personToUpdate := models.UpdatePerson{
		ID:         id,
		Name:       &name,
		PersonType: &personType,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	// update the person

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	err = personRepository.Update(&personToUpdate)
	assert.NoError(t, err)

	// get the company and verify that it's updated
	retrievedPerson, err := personRepository.GetById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPerson)

	assert.Equal(t, id, retrievedPerson.ID)
	assert.Equal(t, name, *retrievedPerson.Name)
	assert.Equal(t, personType.String(), retrievedPerson.PersonType.String())
	assert.Equal(t, email, *retrievedPerson.Email)
	assert.Equal(t, phone, *retrievedPerson.Phone)
	assert.Equal(t, notes, *retrievedPerson.Notes)

	retrievedUpdatedDate := retrievedPerson.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, retrievedUpdatedDate)
}

func TestUpdate_ShouldReturnValidationErrorIfNoPersonFieldsToUpdate(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()
	personToUpdate := models.UpdatePerson{
		ID: id,
	}

	err := personRepository.Update(&personToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

func TestUpdate_ShouldNotReturnErrorIfPersonDoesNotExist(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()
	name := "Another Name"

	personToUpdate := models.UpdatePerson{
		ID:   id,
		Name: &name,
	}

	err := personRepository.Update(&personToUpdate)
	assert.NoError(t, err)
}

// -------- Delete tests: --------

func TestDelete_ShouldDeletePerson(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()
	personToAdd := models.CreatePerson{
		ID:         &id,
		Name:       "Some Name",
		PersonType: models.PersonTypeUnknown,
	}
	_, err := personRepository.Create(&personToAdd)
	assert.NoError(t, err)

	err = personRepository.Delete(&id)
	assert.NoError(t, err)

	retrievedPerson, err := personRepository.GetById(&id)
	assert.Nil(t, retrievedPerson)
	assert.Error(t, err)
}

func TestDelete_ShouldReturnValidationErrorIfPersonIDIsNil(t *testing.T) {
	personRepository := setupPersonRepository(t)

	err := personRepository.Delete(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

func TestDelete_ShouldReturnNotFoundErrorIfPersonIdDoesNotExist(t *testing.T) {
	personRepository := setupPersonRepository(t)

	id := uuid.New()
	err := personRepository.Delete(&id)
	assert.Error(t, err)
	assert.Equal(t, "error: object not found: Person does not exist. ID: "+id.String(), err.Error())
}
