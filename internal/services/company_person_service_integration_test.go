package services_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/services"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupCompanyPersonService(t *testing.T) (
	*services.CompanyPersonService, *repositories.CompanyRepository, *repositories.PersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyPersonServiceTestContainer(t, *config)

	var companyPersonService *services.CompanyPersonService
	err := container.Invoke(func(service *services.CompanyPersonService) {
		companyPersonService = service
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	return companyPersonService, companyRepository, personRepository
}

// -------- AssociateCompanyPerson tests: --------

func TestAssociateCompanyToPerson_ShouldAssociateACompanyToAPerson(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID:   company.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	associatedCompanyPerson, err := companyPersonService.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	assert.Equal(t, companyPerson.CompanyID, associatedCompanyPerson.CompanyID)
	assert.Equal(t, companyPerson.PersonID, associatedCompanyPerson.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, companyPerson.CreatedDate, &associatedCompanyPerson.CreatedDate)
}

func TestAssociateCompanyToPerson_ShouldAssociateACompanyToAPersonWithOnlyRequiredFields(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	associatedCompanyPerson, err := companyPersonService.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	assert.Equal(t, companyPerson.CompanyID, associatedCompanyPerson.CompanyID)
	assert.Equal(t, companyPerson.PersonID, associatedCompanyPerson.PersonID)
	assert.NotNil(t, associatedCompanyPerson.CreatedDate)
}

func TestAssociateCompanyToPerson_ShouldReturnConflictErrorIfCompanyIDAndPersonIDCombinationAlreadyExist(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	_, err := companyPersonService.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: CompanyID and PersonID combination already exists in database.",
		conflictError.Error())
}

// -------- GetByID tests: --------

func TestCompanyPersonServiceGetByID_ShouldGetRecordsMatchingCompanyID(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonService.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	companyPersons, err := companyPersonService.GetByID(&company1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, companyPersons, 2)

	assert.Equal(t, companyPersons[0].CompanyID, company1.ID)
	assert.Equal(t, companyPersons[0].PersonID, person2.ID)

	assert.Equal(t, companyPersons[1].CompanyID, company1.ID)
	assert.Equal(t, companyPersons[1].PersonID, person1.ID)
}

func TestCompanyPersonGetByID_ShouldGetRecordsMatchingPersonID(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonService.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	companyPersons, err := companyPersonService.GetByID(nil, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, companyPersons, 2)

	assert.Equal(t, companyPersons[0].CompanyID, company2.ID)
	assert.Equal(t, companyPersons[0].PersonID, person1.ID)

	assert.Equal(t, companyPersons[1].CompanyID, company1.ID)
	assert.Equal(t, companyPersons[1].PersonID, person1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingCompanyIDAndPersonID(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person1.ID,
	}
	_, err := companyPersonService.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person2.ID,
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID: company2.ID,
		PersonID:  person1.ID,
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	persons, err := companyPersonService.GetByID(&company1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, persons, 1)
	assert.Equal(t, company1.ID, persons[0].CompanyID)
	assert.Equal(t, person1.ID, persons[0].PersonID)
}

// -------- GetAll tests: --------

func TestGetAllCompanyPersons_ShouldReturnAllCompanyPersons(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonService.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	personCompanies, err := companyPersonService.GetAll()
	assert.NoError(t, err)

	assert.Len(t, personCompanies, 3)

	insertedCompanyPerson1 := personCompanies[0]
	assert.Equal(t, company1.ID, insertedCompanyPerson1.CompanyID)
	assert.Equal(t, person2.ID, insertedCompanyPerson1.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, companyPerson2.CreatedDate, &insertedCompanyPerson1.CreatedDate)

	insertedCompanyPerson2 := personCompanies[1]
	assert.Equal(t, company2.ID, insertedCompanyPerson2.CompanyID)
	assert.Equal(t, person2.ID, insertedCompanyPerson2.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, companyPerson3.CreatedDate, &insertedCompanyPerson2.CreatedDate)

	insertedCompanyPerson3 := personCompanies[2]
	assert.Equal(t, company1.ID, insertedCompanyPerson3.CompanyID)
	assert.Equal(t, person1.ID, insertedCompanyPerson3.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, companyPerson1.CreatedDate, &insertedCompanyPerson3.CreatedDate)
}

func TestGetAllCompanyPersons_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	companyPersonService, _, _ := setupCompanyPersonService(t)

	results, err := companyPersonService.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteCompanyPerson_ShouldDeleteCompanyPerson(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	_, err := companyPersonService.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	deleteModel := models.DeleteCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}

	err = companyPersonService.Delete(&deleteModel)
	assert.NoError(t, err)
}

func TestDeleteCompanyPerson_ShouldReturnNotFoundErrorIfNoMatchingCompanyPersonInDatabase(t *testing.T) {
	companyPersonService, companyRepository, personRepository := setupCompanyPersonService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	_, err := companyPersonService.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	deleteModel := models.DeleteCompanyPerson{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	err = companyPersonService.Delete(&deleteModel)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: CompanyPerson does not exist. companyID: "+deleteModel.CompanyID.String()+
			", personID: "+deleteModel.PersonID.String(), notFoundError.Error())
}
