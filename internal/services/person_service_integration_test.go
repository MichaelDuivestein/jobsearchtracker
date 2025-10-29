package services_test

import (
	"errors"
	"jobsearchtracker/internal/api/v1/requests"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/services"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupPersonService(t *testing.T) (
	*services.PersonService,
	*repositories.CompanyRepository,
	*repositories.CompanyPersonRepository) {
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

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var companyPersonRepository *repositories.CompanyPersonRepository
	err = container.Invoke(func(repository *repositories.CompanyPersonRepository) {
		companyPersonRepository = repository
	})
	assert.NoError(t, err)

	return personService, companyRepository, companyPersonRepository
}

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldWork(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	id := uuid.New()
	name := "Dude Janesson"
	email := "em@ai.l"
	phone := "321"
	Notes := "Text"

	personToInsert := models.CreatePerson{
		ID:          &id,
		Name:        name,
		PersonType:  models.PersonTypeCEO,
		Email:       &email,
		Phone:       &phone,
		Notes:       &Notes,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(1, 0, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, -2, 0)),
	}

	insertedPerson, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.Equal(t, id, insertedPerson.ID)
	assert.Equal(t, name, *insertedPerson.Name)
	assert.Equal(t, personToInsert.PersonType.String(), insertedPerson.PersonType.String())
	assert.Equal(t, &email, insertedPerson.Email)
	assert.Equal(t, &phone, insertedPerson.Phone)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert.CreatedDate, insertedPerson.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert.UpdatedDate, insertedPerson.UpdatedDate)
}

func TestCreatePerson_ShouldHandleEmptyFields(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	name := "Sven Joe"

	personToInsert := models.CreatePerson{
		Name:       name,
		PersonType: models.PersonTypeCEO,
	}

	insertedDateApproximation := time.Now()
	insertedPerson, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.NotNil(t, insertedPerson.ID)
	assert.Equal(t, name, *insertedPerson.Name)
	assert.Equal(t, personToInsert.PersonType.String(), insertedPerson.PersonType.String())
	assert.Nil(t, insertedPerson.Email)
	assert.Nil(t, insertedPerson.Phone)
	testutil.AssertDateTimesWithinDelta(t, &insertedDateApproximation, insertedPerson.CreatedDate, time.Second)
	assert.Nil(t, insertedPerson.UpdatedDate)
}

// -------- GetPersonById tests: --------

func TestGetPersonById_ShouldWork(t *testing.T) {
	personService, _, _ := setupPersonService(t)

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
	personService, _, _ := setupPersonService(t)

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
	personService, _, _ := setupPersonService(t)

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
	personService, _, _ := setupPersonService(t)

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
	personService, _, _ := setupPersonService(t)

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
	personService, _, _ := setupPersonService(t)

	// insert persons

	name1 := "abc def"
	personToInsert1 := models.CreatePerson{
		Name:        name1,
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	name2 := "ghi jkl"
	personToInsert2 := models.CreatePerson{
		Name:        name2,
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&personToInsert2)
	assert.NoError(t, err)

	// getAll
	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)

	assert.Equal(t, name2, *persons[0].Name)
	assert.Equal(t, name1, *persons[1].Name)
}

func TestGetAlLPersons_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestGetAll_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	personService, companyRepository, companyPersonRepository := setupPersonService(t)

	// setup persons
	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:          &person2ID,
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3ID := uuid.New()
	person3 := models.CreatePerson{
		ID:          &person3ID,
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: company1ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person2ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, person1ID, persons[0].ID)
	assert.Len(t, *(persons[0]).Companies, 2)

	person1Company1 := (*(*persons[0]).Companies)[0]
	assert.Equal(t, company1ID, person1Company1.ID)
	assert.Equal(t, company1.Name, *person1Company1.Name)
	assert.Equal(t, company1.CompanyType.String(), person1Company1.CompanyType.String())
	assert.Equal(t, company1.Notes, person1Company1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, company1.LastContact, person1Company1.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company1.CreatedDate, person1Company1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company1.UpdatedDate, person1Company1.UpdatedDate)

	person1Company2 := (*(*persons[0]).Companies)[1]
	assert.Equal(t, company2ID, person1Company2.ID)
	assert.Equal(t, company2.Name, *person1Company2.Name)
	assert.Equal(t, company2.CompanyType.String(), person1Company2.CompanyType.String())
	assert.Nil(t, person1Company2.Notes)
	assert.Nil(t, person1Company2.LastContact)
	assert.NotNil(t, person1Company2.CreatedDate)
	assert.Nil(t, person1Company2.UpdatedDate)

	assert.Equal(t, person2ID, persons[1].ID)
	assert.Len(t, *(persons[1]).Companies, 1)
	assert.Equal(t, company2ID, (*(*persons[1]).Companies)[0].ID)

	assert.Equal(t, person3ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personService, companyRepository, _ := setupPersonService(t)

	// setup persons
	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:          &person2ID,
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3ID := uuid.New()
	person3 := models.CreatePerson{
		ID:          &person3ID,
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, person1ID, persons[0].ID)
	assert.Nil(t, persons[0].Companies)

	assert.Equal(t, person2ID, persons[1].ID)
	assert.Nil(t, persons[1].Companies)

	assert.Equal(t, person3ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	personService, companyRepository, companyPersonRepository := setupPersonService(t)

	// setup persons
	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:          &person2ID,
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3ID := uuid.New()
	person3 := models.CreatePerson{
		ID:          &person3ID,
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: company1ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person2ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, person1ID, persons[0].ID)
	assert.Len(t, *(persons[0]).Companies, 2)

	person1Company1 := (*(*persons[0]).Companies)[0]
	assert.Equal(t, company1ID, person1Company1.ID)
	assert.Nil(t, person1Company1.Name)
	assert.Nil(t, person1Company1.CompanyType)
	assert.Nil(t, person1Company1.Notes)
	assert.Nil(t, person1Company1.LastContact)
	assert.Nil(t, person1Company1.CreatedDate)
	assert.Nil(t, person1Company1.UpdatedDate)

	assert.Equal(t, person2ID, persons[1].ID)
	assert.Len(t, *(persons[1]).Companies, 1)
	assert.Equal(t, company2ID, (*(*persons[1]).Companies)[0].ID)

	assert.Equal(t, person3ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personService, companyRepository, _ := setupPersonService(t)

	// setup persons
	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:          &person2ID,
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3ID := uuid.New()
	person3 := models.CreatePerson{
		ID:          &person3ID,
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, person1ID, persons[0].ID)
	assert.Nil(t, persons[0].Companies)

	assert.Equal(t, person2ID, persons[1].ID)
	assert.Nil(t, persons[1].Companies)

	assert.Equal(t, person3ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	personService, companyRepository, companyPersonRepository := setupPersonService(t)

	// setup persons
	person1ID := uuid.New()
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	person2 := models.CreatePerson{
		ID:          &person2ID,
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3ID := uuid.New()
	person3 := models.CreatePerson{
		ID:          &person3ID,
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: company1ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person2ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, person1ID, persons[0].ID)
	assert.Nil(t, persons[0].Companies)

	assert.Equal(t, person2ID, persons[1].ID)
	assert.Nil(t, persons[1].Companies)

	assert.Equal(t, person3ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

// -------- UpdatePerson tests: --------
func TestUpdatePerson_ShouldWork(t *testing.T) {
	personService, _, _ := setupPersonService(t)

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

	updatedDateApproximation := time.Now()
	err = personService.UpdatePerson(&personToUpdate)
	assert.NoError(t, err)

	// get ById
	retrievedPerson, err := personService.GetPersonById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPerson)

	assert.Equal(t, id, retrievedPerson.ID)
	assert.Equal(t, newName, *retrievedPerson.Name)
	assert.Equal(t, newEmail, *retrievedPerson.Email)
	assert.Equal(t, newPhone, *retrievedPerson.Phone)
	assert.Equal(t, newNotes, *retrievedPerson.Notes)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, retrievedPerson.UpdatedDate, time.Second)
}

func TestUpdatePerson_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	personService, _, _ := setupPersonService(t)

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
	personService, _, _ := setupPersonService(t)

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
	personService, _, _ := setupPersonService(t)

	id := uuid.New()
	err := personService.DeletePerson(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Person does not exist. ID: "+id.String(), notFoundError.Error())
}
