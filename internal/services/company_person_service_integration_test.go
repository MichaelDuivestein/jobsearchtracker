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

func TestAssociateCompanyToPerson_ShouldAssociateCompaniesToPersons(t *testing.T) {
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
		CompanyID:   company2.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 0)),
	}
	_, err = companyPersonService.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	personCompanies, err := companyPersonService.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, personCompanies)
	assert.Len(t, personCompanies, 3)

	associatedCompanyPerson1 := personCompanies[0]
	assert.Equal(t, company2.ID, associatedCompanyPerson1.CompanyID)
	assert.Equal(t, person2.ID, associatedCompanyPerson1.PersonID)
	assert.NotNil(t, associatedCompanyPerson1.CreatedDate)

	associatedCompanyPerson2 := personCompanies[1]
	assert.Equal(t, company1.ID, associatedCompanyPerson2.CompanyID)
	assert.Equal(t, person1.ID, associatedCompanyPerson2.PersonID)
	assert.NotNil(t, associatedCompanyPerson2.CreatedDate)

	associatedCompanyPerson3 := personCompanies[2]
	assert.Equal(t, company2.ID, associatedCompanyPerson3.CompanyID)
	assert.Equal(t, person1.ID, associatedCompanyPerson3.PersonID)
	assert.NotNil(t, associatedCompanyPerson3.CreatedDate)
}

// -------- GetByID tests: --------

func TestGetByID_ShouldGetRecordsMatchingCompanyID(t *testing.T) {
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

func TestGetByID_ShouldGetRecordsMatchingPersonID(t *testing.T) {
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

	createdDateToInsert1 := companyPerson2.CreatedDate.Format(time.RFC3339)
	insertedCreatedDate1 := insertedCompanyPerson1.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateToInsert1, insertedCreatedDate1)

	insertedCompanyPerson2 := personCompanies[1]
	assert.Equal(t, company2.ID, insertedCompanyPerson2.CompanyID)
	assert.Equal(t, person2.ID, insertedCompanyPerson2.PersonID)

	createdDateToInsert2 := companyPerson3.CreatedDate.Format(time.RFC3339)
	insertedCreatedDate2 := insertedCompanyPerson2.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateToInsert2, insertedCreatedDate2)

	insertedCompanyPerson3 := personCompanies[2]
	assert.Equal(t, company1.ID, insertedCompanyPerson3.CompanyID)
	assert.Equal(t, person1.ID, insertedCompanyPerson3.PersonID)

	createdDateToInsert3 := companyPerson1.CreatedDate.Format(time.RFC3339)
	insertedCreatedDate3 := insertedCompanyPerson3.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateToInsert3, insertedCreatedDate3)
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
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: CompanyPerson does not exist. companyID: "+deleteModel.CompanyID.String()+
			", personID: "+deleteModel.PersonID.String(), notFoundError.Error())
}
