package repositories_test

import (
	"errors"
	"jobsearchtracker/internal/api/v1/requests"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupPersonRepository(t *testing.T) (
	*repositories.PersonRepository,
	*repositories.CompanyRepository,
	*repositories.CompanyPersonRepository) {

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

	return personRepository, companyRepository, companyPersonRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertPerson(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	person := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person Name",
		PersonType:  models.PersonTypeDeveloper,
		Email:       testutil.ToPtr("some@email.tld"),
		Phone:       testutil.ToPtr("123456"),
		Notes:       testutil.ToPtr("Some Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
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
	testutil.AssertEqualFormattedDateTimes(t, person.CreatedDate, insertedPerson.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, person.UpdatedDate, insertedPerson.UpdatedDate)
}

func TestCreate_ShouldInsertPersonWithMinimumRequiredFields(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	person := models.CreatePerson{
		Name:       "Abc Def",
		PersonType: models.PersonTypeCEO,
	}

	createdDateApproximation := time.Now()
	insertedPerson, err := personRepository.Create(&person)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	assert.NotNil(t, insertedPerson.ID)
	assert.NotNil(t, insertedPerson.Name)
	assert.NotNil(t, insertedPerson.PersonType)
	assert.Nil(t, insertedPerson.Email)
	assert.Nil(t, insertedPerson.Phone)
	assert.Nil(t, insertedPerson.Notes)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, insertedPerson.CreatedDate, time.Second)
	assert.Nil(t, insertedPerson.UpdatedDate)
}

func TestCreate_ShouldReturnConflictErrorOnDuplicatePersonId(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

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
		conflictError.Error())
}

// -------- GetById tests: --------

func TestGetById_ShouldGetPerson(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	personToInsert := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Joe Sparks",
		PersonType: models.PersonTypeDeveloper,
	}
	insertedPerson, err := personRepository.Create(&personToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson)

	retrievedPerson, err := personRepository.GetById(personToInsert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPerson)

	assert.Equal(t, *personToInsert.ID, retrievedPerson.ID)
}

func TestGetById_ShouldReturnNotFoundErrorIfPersonIDDoesNotExist(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	id := uuid.New()

	person, err := personRepository.GetById(&id)
	assert.Nil(t, person)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ID: '"+id.String()+"'",
		notFoundError.Error())
}

// -------- GetAllByName tests: --------

func TestGetAllByName_ShouldReturnPerson(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	personToInsert := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
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
	assert.Equal(t, *personToInsert.ID, person.ID)
	assert.Equal(t, "John Smith", *person.Name)

}

func TestGetAllByName_ShouldReturnNotFoundErrorIfPersonNameDoesNotExist(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	name := "Doesnt Exist"

	person, err := personRepository.GetAllByName(&name)
	assert.Nil(t, person)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: Name: '"+name+"'",
		notFoundError.Error(),
		"Wrong error returned")
}

func TestGetAllByName_ShouldReturnMultiplePersonsWithSameNameSubstring(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	// insert some humans

	person1 := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "frank john",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson1, err := personRepository.Create(&person1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson1)

	person2 := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Frank Jones",
		PersonType: models.PersonTypeCEO,
	}
	insertedPerson2, err := personRepository.Create(&person2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson2)

	person3 := models.CreatePerson{
		ID:         testutil.ToPtr(uuid.New()),
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

// -------- GetAll tests: --------

func TestGetAll_ShouldReturnAllPersons(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	// add some humans

	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Frank Jones",
		PersonType:  models.PersonTypeDeveloper,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	insertedPerson1, err := personRepository.Create(&person1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson1)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Anne Gale",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	insertedPerson2, err := personRepository.Create(&person2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedPerson2)

	// get all humans
	persons, err := personRepository.GetAll(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)

	assert.Equal(t, *person2.ID, persons[0].ID)

	assert.Equal(t, *person1.ID, persons[1].ID)
	assert.Equal(t, person1.Name, *persons[1].Name)
	assert.Equal(t, person1.PersonType.String(), persons[1].PersonType.String())
	assert.Equal(t, person1.Email, persons[1].Email)
	assert.Equal(t, person1.Phone, persons[1].Phone)
	assert.Equal(t, person1.Notes, persons[1].Notes)
	testutil.AssertEqualFormattedDateTimes(t, person1.CreatedDate, persons[1].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, person1.UpdatedDate, persons[1].UpdatedDate)
}

func TestGetAll_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	persons, err := personRepository.GetAll(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestGetAll_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	personRepository, companyRepository, companyPersonRepository := setupPersonRepository(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personRepository.Create(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personRepository.Create(&person3)
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

	persons, err := personRepository.GetAll(models.IncludeExtraDataTypeAll)
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

	person1Company2 := (*(*persons[0]).Companies)[1]
	assert.Equal(t, *company2.ID, person1Company2.ID)

	assert.Equal(t, *person2.ID, persons[1].ID)
	assert.Len(t, *(persons[1]).Companies, 1)
	assert.Equal(t, *company2.ID, (*(*persons[1]).Companies)[0].ID)

	assert.Equal(t, *person3.ID, persons[2].ID)
	assert.Nil(t, persons[2].Companies)
}

func TestGetAll_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personRepository, companyRepository, _ := setupPersonRepository(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personRepository.Create(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personRepository.Create(&person3)
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

	persons, err := personRepository.GetAll(models.IncludeExtraDataTypeAll)
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

func TestGetAll_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	personRepository, companyRepository, companyPersonRepository := setupPersonRepository(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personRepository.Create(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personRepository.Create(&person3)
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

	persons, err := personRepository.GetAll(models.IncludeExtraDataTypeIDs)
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

func TestGetAll_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personRepository, companyRepository, _ := setupPersonRepository(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personRepository.Create(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personRepository.Create(&person3)
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

	persons, err := personRepository.GetAll(models.IncludeExtraDataTypeIDs)
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

func TestGetAll_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	personRepository, companyRepository, companyPersonRepository := setupPersonRepository(t)

	// setup persons
	person1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1",
		PersonType:  models.PersonTypeDeveloper,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := personRepository.Create(&person1)
	assert.NoError(t, err)

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person2",
		PersonType:  models.PersonTypeCTO,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person2)
	assert.NoError(t, err)

	person3 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "person3",
		PersonType:  models.PersonTypeHR,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = personRepository.Create(&person3)
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

	persons, err := personRepository.GetAll(models.IncludeExtraDataTypeNone)
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

// -------- Update tests: --------

func TestUpdate_ShouldUpdatePerson(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

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

	personType := models.PersonType(models.PersonTypeHR)
	personToUpdate := models.UpdatePerson{
		ID:         id,
		Name:       testutil.ToPtr("Another Name"),
		PersonType: &personType,
		Email:      testutil.ToPtr("a@b.c"),
		Phone:      testutil.ToPtr("312765"),
		Notes:      testutil.ToPtr("Something noteworthy"),
	}

	// update the person

	updatedDateApproximation := time.Now()
	err = personRepository.Update(&personToUpdate)
	assert.NoError(t, err)

	// get the company and verify that it's updated
	retrievedPerson, err := personRepository.GetById(&id)
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

func TestUpdate_ShouldNotReturnErrorIfPersonDoesNotExist(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

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
	personRepository, _, _ := setupPersonRepository(t)

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

func TestDelete_ShouldReturnNotFoundErrorIfPersonIdDoesNotExist(t *testing.T) {
	personRepository, _, _ := setupPersonRepository(t)

	id := uuid.New()
	err := personRepository.Delete(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Person does not exist. ID: "+id.String(), notFoundError.Error())
}
