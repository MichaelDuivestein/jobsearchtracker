package repositories_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupCompanyPersonRepository(t *testing.T) (
	*repositories.CompanyPersonRepository, *repositories.CompanyRepository, *repositories.PersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyPersonRepositoryTestContainer(t, *config)

	var companyPersonRepository *repositories.CompanyPersonRepository
	err := container.Invoke(func(repository *repositories.CompanyPersonRepository) {
		companyPersonRepository = repository
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

	return companyPersonRepository, companyRepository, personRepository
}

// -------- AssociateCompanyPerson tests: --------

func TestAssociateCompanyToPerson_ShouldWork(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID:   company.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now()),
	}
	associatedCompanyPerson, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)
	assert.NotNil(t, associatedCompanyPerson)

	assert.Equal(t, company.ID, associatedCompanyPerson.CompanyID)
	assert.Equal(t, person.ID, associatedCompanyPerson.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, companyPerson.CreatedDate, &associatedCompanyPerson.CreatedDate)
}

func TestAssociateCompanyToPerson_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	associatedCompanyPerson, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)
	assert.NotNil(t, associatedCompanyPerson)

	assert.Equal(t, company.ID, associatedCompanyPerson.CompanyID)
	assert.Equal(t, person.ID, associatedCompanyPerson.PersonID)
	assert.NotNil(t, associatedCompanyPerson.CreatedDate)
}

func TestAssociateCompanyToPerson_ShouldAssociateACompanyToMultiplePersons(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	personCompanies, err := companyPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, personCompanies)
	assert.Len(t, personCompanies, 2)

	associatedCompanyPerson1 := personCompanies[0]
	assert.Equal(t, company.ID, associatedCompanyPerson1.CompanyID)
	assert.Equal(t, person2.ID, associatedCompanyPerson1.PersonID)
	assert.NotNil(t, associatedCompanyPerson1.CreatedDate)

	associatedCompanyPerson2 := personCompanies[1]
	assert.Equal(t, company.ID, associatedCompanyPerson2.CompanyID)
	assert.Equal(t, person1.ID, associatedCompanyPerson2.PersonID)
	assert.NotNil(t, associatedCompanyPerson2.CreatedDate)
}

func TestAssociateCompanyToPerson_ShouldAssociateMultipleCompaniesToAPerson(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	personCompanies, err := companyPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, personCompanies)
	assert.Len(t, personCompanies, 2)

	associatedCompanyPerson1 := personCompanies[0]
	assert.Equal(t, company2.ID, associatedCompanyPerson1.CompanyID)
	assert.Equal(t, person.ID, associatedCompanyPerson1.PersonID)
	assert.NotNil(t, associatedCompanyPerson1.CreatedDate)

	associatedCompanyPerson2 := personCompanies[1]
	assert.Equal(t, company1.ID, associatedCompanyPerson2.CompanyID)
	assert.Equal(t, person.ID, associatedCompanyPerson2.PersonID)
	assert.NotNil(t, associatedCompanyPerson2.CreatedDate)
}

func TestAssociateCompanyToPerson_ShouldReturnConflictErrorIfCompanyIDAndPersonIDCombinationAlreadyExist(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: CompanyID and PersonID combination already exists in database.",
		conflictError.Error())
}

func TestAssociateCompanyToPerson_ShouldReturnValidationErrorIfPersonIDDoesNotExist(t *testing.T) {
	companyPersonRepository, companyRepository, _ := setupCompanyPersonRepository(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  uuid.New(),
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateCompanyToPerson_ShouldReturnValidationErrorIfCompanyIDDoesNotExist(t *testing.T) {
	companyPersonRepository, _, personRepository := setupCompanyPersonRepository(t)

	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: uuid.New(),
		PersonID:  person.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateCompanyToPerson_ShouldReturnValidationErrorIfCompanyIDAndPersonIDDoNotExist(t *testing.T) {
	companyPersonRepository, _, _ := setupCompanyPersonRepository(t)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

// -------- GetByID tests: --------

func TestGetByID_ShouldGetRecordsMatchingCompanyID(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	companyPersons, err := companyPersonRepository.GetByID(&company1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, companyPersons, 2)

	assert.Equal(t, companyPersons[0].CompanyID, company1.ID)
	assert.Equal(t, companyPersons[0].PersonID, person2.ID)

	assert.Equal(t, companyPersons[1].CompanyID, company1.ID)
	assert.Equal(t, companyPersons[1].PersonID, person1.ID)
}

func TestCompanyPersonGetByID_ShouldGetRecordsMatchingPersonID(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	companyPersons, err := companyPersonRepository.GetByID(nil, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, companyPersons, 2)

	assert.Equal(t, companyPersons[0].CompanyID, company2.ID)
	assert.Equal(t, companyPersons[0].PersonID, person1.ID)

	assert.Equal(t, companyPersons[1].CompanyID, company1.ID)
	assert.Equal(t, companyPersons[1].PersonID, person1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingCompanyIDAndPersonID(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person1.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID: company2.ID,
		PersonID:  person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	persons, err := companyPersonRepository.GetByID(&company1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, persons, 1)
	assert.Equal(t, company1.ID, persons[0].CompanyID)
	assert.Equal(t, person1.ID, persons[0].PersonID)
}

func TestGetByID_ShouldGetNoRecordsIfCompanyIDDoesNotMatch(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person1.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID: company2.ID,
		PersonID:  person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	persons, err := companyPersonRepository.GetByID(testutil.ToPtr(uuid.New()), &person1.ID)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestCompanyPersonGetByID_ShouldGetNoRecordsIfPersonIDDoesNotMatch(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person1.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID: company2.ID,
		PersonID:  person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	persons, err := companyPersonRepository.GetByID(&company1.ID, testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestGetByID_ShouldGetNoRecordsIfCompanyIDAndPersonIDDoesNotMatch(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person1.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID: company1.ID,
		PersonID:  person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID: company2.ID,
		PersonID:  person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	persons, err := companyPersonRepository.GetByID(testutil.ToPtr(uuid.New()), testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestCompanyPersonGetByID_ShouldGetNoRecordsIfNoRecordsInDB(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	persons, err := companyPersonRepository.GetByID(&company1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

// -------- GetAll tests: --------

func TestGetAllCompanyPersons_ShouldReturnAllCompanyPersons(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person1.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2.ID,
		PersonID:    person2.ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	personCompanies, err := companyPersonRepository.GetAll()
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
	companyPersonRepository, _, _ := setupCompanyPersonRepository(t)

	results, err := companyPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteCompanyPerson_ShouldDeleteCompanyPerson(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	model := models.DeleteCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}

	err = companyPersonRepository.Delete(&model)
	assert.NoError(t, err)
}

func TestDeleteCompanyPerson_ShouldReturnNotFoundErrorIfNoMatchingCompanyPersonInDatabase(t *testing.T) {
	companyPersonRepository, companyRepository, personRepository := setupCompanyPersonRepository(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := models.AssociateCompanyPerson{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}
	_, err := companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	model := models.DeleteCompanyPerson{
		CompanyID: uuid.New(),
		PersonID:  uuid.New(),
	}

	err = companyPersonRepository.Delete(&model)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: CompanyPerson does not exist. companyID: "+model.CompanyID.String()+
			", personID: "+model.PersonID.String(), notFoundError.Error())
}
