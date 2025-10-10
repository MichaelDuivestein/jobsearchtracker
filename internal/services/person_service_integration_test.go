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
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notFoundError.Error())
}

// -------- GetPersonsByName tests: --------

func TestGetPersonsByName_ShouldReturnASinglePerson(t *testing.T) {
	personService := setupPersonService(t)

	// insert persons
	id1 := uuid.New()
	name1 := "Dane Joe"
	personToInsert1 := models.CreatePerson{
		ID:         &id1,
		Name:       name1,
		PersonType: models.PersonTypeCTO,
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	name2 := "Bruce Pritt"
	personToInsert2 := models.CreatePerson{
		ID:         &id2,
		Name:       name2,
		PersonType: models.PersonTypeHR,
	}
	_, err = personService.CreatePerson(&personToInsert2)
	assert.NoError(t, err)

	// GetByName
	nameToGet := "Joe"
	persons, err := personService.GetPersonsByName(&nameToGet)
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 1)

	assert.Equal(t, id1, persons[0].ID)
}

func TestGetPersonsByName_ShouldReturnMultiplePersons(t *testing.T) {
	personService := setupPersonService(t)

	// insert persons

	id1 := uuid.New()
	name1 := "Sonny Brak"
	personToInsert1 := models.CreatePerson{
		ID:         &id1,
		Name:       name1,
		PersonType: models.PersonTypeDeveloper,
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	name2 := "Mary Sparks"
	personToInsert2 := models.CreatePerson{
		ID:         &id2,
		Name:       name2,
		PersonType: models.PersonTypeOther,
	}
	_, err = personService.CreatePerson(&personToInsert2)
	assert.NoError(t, err)

	id3 := uuid.New()
	name3 := "David Jonesson"
	personToInsert3 := models.CreatePerson{
		ID:         &id3,
		Name:       name3,
		PersonType: models.PersonTypeExternalRecruiter,
	}
	_, err = personService.CreatePerson(&personToInsert3)
	assert.NoError(t, err)

	// GetByName

	nameToGet := "son"
	persons, err := personService.GetPersonsByName(&nameToGet)
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)

	assert.Equal(t, id3, persons[0].ID)
	assert.Equal(t, id1, persons[1].ID)
}

func TestGetPersonsByName_ShouldReturnNotFoundErrorIfNoNamesMatch(t *testing.T) {
	personService := setupPersonService(t)

	// insert persons
	id1 := uuid.New()
	name1 := "Debbie Star"
	personToInsert1 := models.CreatePerson{
		ID:         &id1,
		Name:       name1,
		PersonType: models.PersonTypeUnknown,
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	name2 := "Manny Dee"
	personToInsert2 := models.CreatePerson{
		ID:         &id2,
		Name:       name2,
		PersonType: models.PersonTypeJobAdvertiser,
	}
	_, err = personService.CreatePerson(&personToInsert2)
	assert.NoError(t, err)

	// GetByName
	nameToGet := "Bee"
	persons, err := personService.GetPersonsByName(&nameToGet)
	assert.Nil(t, persons)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Name: '"+nameToGet+"'", notFoundError.Error())
}

// -------- GetAllPersons tests: --------
func TestGetAlLPersons_ShouldWork(t *testing.T) {
	personService := setupPersonService(t)

	// insert persons

	name1 := "abc def"
	personToInsert1 := models.CreatePerson{
		Name:       name1,
		PersonType: models.PersonTypeHR,
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	name2 := "ghi jkl"
	personToInsert2 := models.CreatePerson{
		Name:       name2,
		PersonType: models.PersonTypeHR,
	}
	_, err = personService.CreatePerson(&personToInsert2)
	assert.NoError(t, err)

	// getAll
	persons, err := personService.GetAllPersons()
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)

	assert.Equal(t, name1, persons[0].Name)
	assert.Equal(t, name2, persons[1].Name)
}

func TestGetAlLPersons_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	personService := setupPersonService(t)

	persons, err := personService.GetAllPersons()
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

// -------- UpdatePerson tests: --------
func TestUpdatePerson_ShouldWork(t *testing.T) {
	personService := setupPersonService(t)

	// insert person

	id := uuid.New()
	originalName := "Bolt"
	originalEmail := "some email"
	originalPhone := "48908"
	originalNotes := "Some Notes"
	originalCreatedDate := time.Now().AddDate(1, 0, 0)
	originalUpdatedDate := time.Now().AddDate(0, -2, 0)

	personToInsert := models.CreatePerson{
		ID:          &id,
		Name:        originalName,
		PersonType:  models.PersonTypeCEO,
		Email:       &originalEmail,
		Phone:       &originalPhone,
		Notes:       &originalNotes,
		CreatedDate: &originalCreatedDate,
		UpdatedDate: &originalUpdatedDate,
	}
	_, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)

	// update person

	newName := "Another Name"
	newEmail := "Another Email"
	newPhone := "5940358"
	newNotes := "Another notes"
	personToUpdate := models.UpdatePerson{
		ID:    id,
		Name:  &newName,
		Email: &newEmail,
		Phone: &newPhone,
		Notes: &newNotes,
	}

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	err = personService.UpdatePerson(&personToUpdate)
	assert.NoError(t, err)

	// get ById
	retrievedPerson, err := personService.GetPersonById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPerson)

	assert.Equal(t, id, retrievedPerson.ID)
	assert.Equal(t, newName, retrievedPerson.Name)
	assert.Equal(t, newEmail, *retrievedPerson.Email)
	assert.Equal(t, newPhone, *retrievedPerson.Phone)
	assert.Equal(t, newNotes, *retrievedPerson.Notes)

	updatedDate := retrievedPerson.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, updatedDate)
}

func TestUpdatePerson_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	personService := setupPersonService(t)

	id := uuid.New()
	notes := "Random Notes"
	personToUpdate := models.UpdatePerson{
		ID:    id,
		Notes: &notes,
	}

	err := personService.UpdatePerson(&personToUpdate)
	assert.NoError(t, err)
}

// -------- DeletePerson tests: --------

func TestDeletePerson_ShouldWork(t *testing.T) {
	personService := setupPersonService(t)

	// insert person

	id := uuid.New()
	name := "Dave Davesson"
	personToInsert := models.CreatePerson{
		ID:         &id,
		Name:       name,
		PersonType: models.PersonTypeDeveloper,
	}
	_, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)

	// delete person

	err = personService.DeletePerson(&id)
	assert.NoError(t, err)

	//ensure that person is deleted

	retrievedPerson, err := personService.GetPersonById(&id)
	assert.Nil(t, retrievedPerson)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notFoundError.Error())
}

func TestDeletePerson_ShouldReturnNotFoundErrorIfIdToDeleteDoesNotExist(t *testing.T) {
	personService := setupPersonService(t)

	id := uuid.New()
	err := personService.DeletePerson(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Person does not exist. ID: "+id.String(), notFoundError.Error())
}
