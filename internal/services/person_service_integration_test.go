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

	personToInsert := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Dude Janesson",
		PersonType:  models.PersonTypeCEO,
		Email:       testutil.ToPtr("em@ai.l"),
		Phone:       testutil.ToPtr("321"),
		Notes:       testutil.ToPtr("Text"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(1, 0, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, -2, 0)),
	}

	insertedPerson, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.Equal(t, *personToInsert.ID, insertedPerson.ID)
	assert.Equal(t, personToInsert.Name, *insertedPerson.Name)
	assert.Equal(t, personToInsert.PersonType.String(), insertedPerson.PersonType.String())
	assert.Equal(t, personToInsert.Email, insertedPerson.Email)
	assert.Equal(t, personToInsert.Phone, insertedPerson.Phone)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert.CreatedDate, insertedPerson.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert.UpdatedDate, insertedPerson.UpdatedDate)
}

func TestCreatePerson_ShouldHandleEmptyFields(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	personToInsert := models.CreatePerson{
		Name:       "Sven Joe",
		PersonType: models.PersonTypeCEO,
	}

	insertedDateApproximation := time.Now()
	insertedPerson, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.NotNil(t, insertedPerson.ID)
	assert.Equal(t, personToInsert.Name, *insertedPerson.Name)
	assert.Equal(t, personToInsert.PersonType.String(), insertedPerson.PersonType.String())
	assert.Nil(t, insertedPerson.Email)
	assert.Nil(t, insertedPerson.Phone)
	testutil.AssertDateTimesWithinDelta(t, &insertedDateApproximation, insertedPerson.CreatedDate, time.Second)
	assert.Nil(t, insertedPerson.UpdatedDate)
}

// -------- GetPersonById tests: --------

func TestGetPersonById_ShouldWork(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	personToInsert := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Some Name",
		PersonType:  models.PersonTypeOther,
		Email:       testutil.ToPtr("an@email.address"),
		Phone:       testutil.ToPtr("128932019"),
		Notes:       testutil.ToPtr("No notes here..."),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, -1, 0)),
	}

	_, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)

	retrievedPerson, err := personService.GetPersonById(personToInsert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPerson)

	assert.NotNil(t, retrievedPerson.ID)
	assert.Equal(t, personToInsert.Name, *retrievedPerson.Name)
	assert.Equal(t, personToInsert.PersonType.String(), retrievedPerson.PersonType.String())
	assert.Equal(t, personToInsert.Email, retrievedPerson.Email)
	assert.Equal(t, personToInsert.Phone, retrievedPerson.Phone)
	testutil.AssertDateTimesWithinDelta(t, personToInsert.CreatedDate, retrievedPerson.CreatedDate, time.Second)
	testutil.AssertDateTimesWithinDelta(t, personToInsert.UpdatedDate, retrievedPerson.UpdatedDate, time.Second)
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

func TestGetPersonsByName_ShouldReturnMultiplePersons(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	// insert persons

	personToInsert1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Sonny Brak",
		PersonType:  models.PersonTypeDeveloper,
		Phone:       testutil.ToPtr("31425683"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -3, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, -4, 0)),
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	personToInsert2 := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Mary Sparks",
		PersonType: models.PersonTypeOther,
	}
	_, err = personService.CreatePerson(&personToInsert2)
	assert.NoError(t, err)

	personToInsert3 := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "David Jonesson",
		PersonType: models.PersonTypeExternalRecruiter,
	}
	_, err = personService.CreatePerson(&personToInsert3)
	assert.NoError(t, err)

	// GetByName

	persons, err := personService.GetPersonsByName(testutil.ToPtr("son"))
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)

	assert.Equal(t, *personToInsert3.ID, persons[0].ID)

	assert.Equal(t, *personToInsert1.ID, persons[1].ID)
	assert.Equal(t, personToInsert1.Name, *persons[1].Name)
	assert.Equal(t, personToInsert1.PersonType.String(), persons[1].PersonType.String())
	assert.Equal(t, personToInsert1.Email, persons[1].Email)
	assert.Equal(t, personToInsert1.Phone, persons[1].Phone)
	assert.Equal(t, personToInsert1.Notes, persons[1].Notes)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert1.CreatedDate, persons[1].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert1.UpdatedDate, persons[1].UpdatedDate)
}

func TestGetPersonsByName_ShouldReturnNotFoundErrorIfNoNamesMatch(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	// insert persons
	personToInsert1 := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Debbie Star",
		PersonType: models.PersonTypeUnknown,
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	personToInsert2 := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Manny Dee",
		PersonType: models.PersonTypeJobAdvertiser,
	}
	_, err = personService.CreatePerson(&personToInsert2)
	assert.NoError(t, err)

	// GetByName
	nameToGet := "Bee"
	persons, err := personService.GetPersonsByName(testutil.ToPtr(nameToGet))
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

	personToInsert1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "abc def",
		PersonType:  models.PersonTypeHR,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&personToInsert1)
	assert.NoError(t, err)

	personToInsert2 := models.CreatePerson{
		Name:        "ghi jkl",
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

	assert.Equal(t, personToInsert2.Name, *persons[0].Name)

	assert.Equal(t, *personToInsert1.ID, persons[1].ID)
	assert.Equal(t, personToInsert1.Name, *persons[1].Name)
	assert.Equal(t, personToInsert1.PersonType.String(), persons[1].PersonType.String())
	assert.Equal(t, personToInsert1.Email, persons[1].Email)
	assert.Equal(t, personToInsert1.Phone, persons[1].Phone)
	assert.Equal(t, personToInsert1.Notes, persons[1].Notes)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert1.CreatedDate, persons[1].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, personToInsert1.UpdatedDate, persons[1].UpdatedDate)
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
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: *company1.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, *person1.ID, persons[0].ID)
	assert.Len(t, *(persons[0]).Companies, 2)

	person1Company1 := (*(*persons[0]).Companies)[0]
	assert.Equal(t, *company1.ID, person1Company1.ID)
	assert.Equal(t, company1.Name, *person1Company1.Name)
	assert.Equal(t, company1.CompanyType.String(), person1Company1.CompanyType.String())
	assert.Equal(t, company1.Notes, person1Company1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, company1.LastContact, person1Company1.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company1.CreatedDate, person1Company1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company1.UpdatedDate, person1Company1.UpdatedDate)

	assert.Equal(t, *company2.ID, (*(*persons[0]).Companies)[1].ID)

	assert.Equal(t, *person2.ID, persons[1].ID)
	assert.Len(t, *(persons[1]).Companies, 1)
	assert.Equal(t, *company2.ID, (*(*persons[1]).Companies)[0].ID)

	assert.Equal(t, *person3.ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personService, companyRepository, _ := setupPersonService(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
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

	assert.Equal(t, *person1.ID, persons[0].ID)
	assert.Nil(t, persons[0].Companies)

	assert.Equal(t, *person2.ID, persons[1].ID)
	assert.Nil(t, persons[1].Companies)

	assert.Equal(t, *person3.ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	personService, companyRepository, companyPersonRepository := setupPersonService(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: *company1.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, *person1.ID, persons[0].ID)
	assert.Len(t, *(persons[0]).Companies, 2)

	person1Company1 := (*(*persons[0]).Companies)[0]
	assert.Equal(t, *company1.ID, person1Company1.ID)
	assert.Nil(t, person1Company1.Name)
	assert.Nil(t, person1Company1.CompanyType)
	assert.Nil(t, person1Company1.Notes)
	assert.Nil(t, person1Company1.LastContact)
	assert.Nil(t, person1Company1.CreatedDate)
	assert.Nil(t, person1Company1.UpdatedDate)

	assert.Equal(t, *person2.ID, persons[1].ID)
	assert.Len(t, *(persons[1]).Companies, 1)
	assert.Equal(t, *company2.ID, (*(*persons[1]).Companies)[0].ID)

	assert.Equal(t, *person3.ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personService, companyRepository, _ := setupPersonService(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
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

	assert.Equal(t, *person1.ID, persons[0].ID)
	assert.Nil(t, persons[0].Companies)

	assert.Equal(t, *person2.ID, persons[1].ID)
	assert.Nil(t, persons[1].Companies)

	assert.Equal(t, *person3.ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	personService, companyRepository, companyPersonRepository := setupPersonService(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personService.CreatePerson(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personService.CreatePerson(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personService.CreatePerson(&person3)
	assert.NoError(t, err)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
	}
	_, err = companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: *company1.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	persons, err := personService.GetAllPersons(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 3)

	assert.Equal(t, *person1.ID, persons[0].ID)
	assert.Nil(t, persons[0].Companies)

	assert.Equal(t, *person2.ID, persons[1].ID)
	assert.Nil(t, persons[1].Companies)

	assert.Equal(t, *person3.ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

// -------- UpdatePerson tests: --------
func TestUpdatePerson_ShouldWork(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	// insert person

	id := uuid.New()
	personToInsert := models.CreatePerson{
		ID:          &id,
		Name:        "Bolt",
		PersonType:  models.PersonTypeCEO,
		Email:       testutil.ToPtr("some email"),
		Phone:       testutil.ToPtr("48908"),
		Notes:       testutil.ToPtr("Some Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(1, 0, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, -2, 0)),
	}
	_, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)

	// update person

	var personTypeToUpdate models.PersonType = models.PersonTypeCTO
	personToUpdate := models.UpdatePerson{
		ID:         id,
		Name:       testutil.ToPtr("Another Name"),
		PersonType: &personTypeToUpdate,
		Email:      testutil.ToPtr("Another Email"),
		Phone:      testutil.ToPtr("5940358"),
		Notes:      testutil.ToPtr("Another notes"),
	}

	updatedDateApproximation := time.Now()
	err = personService.UpdatePerson(&personToUpdate)
	assert.NoError(t, err)

	// get ById
	retrievedPerson, err := personService.GetPersonById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPerson)

	assert.Equal(t, personToUpdate.ID, retrievedPerson.ID)
	assert.Equal(t, personToUpdate.Name, retrievedPerson.Name)
	assert.Equal(t, personToUpdate.PersonType.String(), retrievedPerson.PersonType.String())
	assert.Equal(t, personToUpdate.Email, retrievedPerson.Email)
	assert.Equal(t, personToUpdate.Phone, retrievedPerson.Phone)
	assert.Equal(t, personToUpdate.Notes, retrievedPerson.Notes)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, retrievedPerson.UpdatedDate, time.Second)
}

func TestUpdatePerson_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	personToUpdate := models.UpdatePerson{
		ID:    uuid.New(),
		Notes: testutil.ToPtr("Random Notes"),
	}

	err := personService.UpdatePerson(&personToUpdate)
	assert.NoError(t, err)
}

// -------- DeletePerson tests: --------

func TestDeletePerson_ShouldWork(t *testing.T) {
	personService, _, _ := setupPersonService(t)

	// insert person

	personToInsert := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Dave Davesson",
		PersonType: models.PersonTypeDeveloper,
	}
	_, err := personService.CreatePerson(&personToInsert)
	assert.NoError(t, err)

	// delete person

	err = personService.DeletePerson(personToInsert.ID)
	assert.NoError(t, err)

	//ensure that person is deleted

	retrievedPerson, err := personService.GetPersonById(personToInsert.ID)
	assert.Nil(t, retrievedPerson)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+personToInsert.ID.String()+"'", notFoundError.Error())
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
