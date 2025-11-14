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

func setupApplicationPersonRepository(t *testing.T) (
	*repositories.ApplicationPersonRepository,
	*repositories.ApplicationRepository,
	*repositories.PersonRepository,
	*repositories.CompanyRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupApplicationPersonRepositoryTestContainer(t, *config)

	var applicationPersonRepository *repositories.ApplicationPersonRepository
	err := container.Invoke(func(repository *repositories.ApplicationPersonRepository) {
		applicationPersonRepository = repository
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	return applicationPersonRepository, applicationRepository, personRepository, companyRepository
}

// -------- AssociateApplicationPerson tests: --------

func TestAssociateApplicationToPerson_ShouldWork(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
		CreatedDate:   testutil.ToPtr(time.Now()),
	}
	associatedApplicationPerson, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)
	assert.NotNil(t, associatedApplicationPerson)

	assert.Equal(t, application.ID, associatedApplicationPerson.ApplicationID)
	assert.Equal(t, person.ID, associatedApplicationPerson.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, applicationPerson.CreatedDate, &associatedApplicationPerson.CreatedDate)
}

func TestAssociateApplicationToPerson_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	associatedApplicationPerson, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)
	assert.NotNil(t, associatedApplicationPerson)

	assert.Equal(t, application.ID, associatedApplicationPerson.ApplicationID)
	assert.Equal(t, person.ID, associatedApplicationPerson.PersonID)
	assert.NotNil(t, associatedApplicationPerson.CreatedDate)
}

func TestAssociateApplicationToPerson_ShouldAssociateAnApplicationToMultiplePersons(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	personCompanies, err := applicationPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, personCompanies)
	assert.Len(t, personCompanies, 2)

	associatedApplicationPerson1 := personCompanies[0]
	assert.Equal(t, application.ID, associatedApplicationPerson1.ApplicationID)
	assert.Equal(t, person2.ID, associatedApplicationPerson1.PersonID)
	assert.NotNil(t, associatedApplicationPerson1.CreatedDate)

	associatedApplicationPerson2 := personCompanies[1]
	assert.Equal(t, application.ID, associatedApplicationPerson2.ApplicationID)
	assert.Equal(t, person1.ID, associatedApplicationPerson2.PersonID)
	assert.NotNil(t, associatedApplicationPerson2.CreatedDate)
}

func TestAssociateApplicationToPerson_ShouldAssociateMultipleApplicationsToAPerson(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	personCompanies, err := applicationPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, personCompanies)
	assert.Len(t, personCompanies, 2)

	associatedApplicationPerson1 := personCompanies[0]
	assert.Equal(t, application2.ID, associatedApplicationPerson1.ApplicationID)
	assert.Equal(t, person.ID, associatedApplicationPerson1.PersonID)
	assert.NotNil(t, associatedApplicationPerson1.CreatedDate)

	associatedApplicationPerson2 := personCompanies[1]
	assert.Equal(t, application1.ID, associatedApplicationPerson2.ApplicationID)
	assert.Equal(t, person.ID, associatedApplicationPerson2.PersonID)
	assert.NotNil(t, associatedApplicationPerson2.CreatedDate)
}

func TestAssociateApplicationToPerson_ShouldReturnConflictErrorIfApplicationIDAndPersonIDCombinationAlreadyExist(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: ApplicationID and PersonID combination already exists in database.",
		conflictError.Error())
}

func TestAssociateApplicationToPerson_ShouldReturnValidationErrorIfPersonIDDoesNotExist(t *testing.T) {
	applicationPersonRepository, applicationRepository, _, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      uuid.New(),
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateApplicationToPerson_ShouldReturnValidationErrorIfApplicationIDDoesNotExist(t *testing.T) {
	applicationPersonRepository, _, personRepository, _ := setupApplicationPersonRepository(t)

	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      person.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestAssociateApplicationToPerson_ShouldReturnValidationErrorIfApplicationIDAndPersonIDDoNotExist(t *testing.T) {
	applicationPersonRepository, _, _, _ := setupApplicationPersonRepository(t)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

// -------- GetByID tests: --------

func TestApplicationPersonGetByID_ShouldGetRecordsMatchingApplicationID(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	applicationPersons, err := applicationPersonRepository.GetByID(&application1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, applicationPersons, 2)

	assert.Equal(t, applicationPersons[0].ApplicationID, application1.ID)
	assert.Equal(t, applicationPersons[0].PersonID, person2.ID)

	assert.Equal(t, applicationPersons[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationPersons[1].PersonID, person1.ID)
}

func TestApplicationPersonGetByID_ShouldGetRecordsMatchingPersonID(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	applicationPersons, err := applicationPersonRepository.GetByID(nil, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, applicationPersons, 2)

	assert.Equal(t, applicationPersons[0].ApplicationID, application2.ID)
	assert.Equal(t, applicationPersons[0].PersonID, person1.ID)

	assert.Equal(t, applicationPersons[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationPersons[1].PersonID, person1.ID)
}

func TestGetByID_ShouldGetRecordsMatchingApplicationIDAndPersonID(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository :=
		setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	persons, err := applicationPersonRepository.GetByID(&application1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, persons, 1)
	assert.Equal(t, application1.ID, persons[0].ApplicationID)
	assert.Equal(t, person1.ID, persons[0].PersonID)
}

func TestApplicationPersonGetByID_ShouldGetNoRecordsIfApplicationIDDoesNotMatch(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	persons, err := applicationPersonRepository.GetByID(testutil.ToPtr(uuid.New()), &person1.ID)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestApplicationPersonGetByID_ShouldGetNoRecordsIfPersonIDDoesNotMatch(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	persons, err := applicationPersonRepository.GetByID(&application1.ID, testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestGetByID_ShouldGetNoRecordsIfApplicationIDAndPersonIDDoesNotMatch(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	persons, err := applicationPersonRepository.GetByID(testutil.ToPtr(uuid.New()), testutil.ToPtr(uuid.New()))
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

func TestApplicationPersonGetByID_ShouldGetNoRecordsIfNoRecordsInDB(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	persons, err := applicationPersonRepository.GetByID(&application1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

// -------- GetAll tests: --------

func TestGetAllApplicationPersons_ShouldReturnAllApplicationPersons(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonRepository.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	personCompanies, err := applicationPersonRepository.GetAll()
	assert.NoError(t, err)

	assert.Len(t, personCompanies, 3)

	insertedApplicationPerson1 := personCompanies[0]
	assert.Equal(t, application1.ID, insertedApplicationPerson1.ApplicationID)
	assert.Equal(t, person2.ID, insertedApplicationPerson1.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, applicationPerson2.CreatedDate, &insertedApplicationPerson1.CreatedDate)

	insertedApplicationPerson2 := personCompanies[1]
	assert.Equal(t, application2.ID, insertedApplicationPerson2.ApplicationID)
	assert.Equal(t, person2.ID, insertedApplicationPerson2.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, applicationPerson3.CreatedDate, &insertedApplicationPerson2.CreatedDate)

	insertedApplicationPerson3 := personCompanies[2]
	assert.Equal(t, application1.ID, insertedApplicationPerson3.ApplicationID)
	assert.Equal(t, person1.ID, insertedApplicationPerson3.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, applicationPerson1.CreatedDate, &insertedApplicationPerson3.CreatedDate)
}

func TestGetAllApplicationPersons_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	applicationPersonRepository, _, _, _ := setupApplicationPersonRepository(t)

	results, err := applicationPersonRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteApplicationPerson_ShouldDeleteApplicationPerson(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	model := models.DeleteApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}

	err = applicationPersonRepository.Delete(&model)
	assert.NoError(t, err)
}

func TestDeleteApplicationPerson_ShouldReturnNotFoundErrorIfNoMatchingApplicationPersonInDatabase(t *testing.T) {
	applicationPersonRepository, applicationRepository, personRepository, companyRepository := setupApplicationPersonRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	_, err := applicationPersonRepository.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	model := models.DeleteApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}

	err = applicationPersonRepository.Delete(&model)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ApplicationPerson does not exist. applicationID: "+model.ApplicationID.String()+
			", personID: "+model.PersonID.String(), notFoundError.Error())
}
