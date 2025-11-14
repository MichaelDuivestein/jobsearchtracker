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

func setupApplicationPersonService(t *testing.T) (
	*services.ApplicationPersonService,
	*repositories.ApplicationRepository,
	*repositories.PersonRepository,
	*repositories.CompanyRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupApplicationPersonServiceTestContainer(t, *config)

	var applicationPersonService *services.ApplicationPersonService
	err := container.Invoke(func(service *services.ApplicationPersonService) {
		applicationPersonService = service
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

	return applicationPersonService, applicationRepository, personRepository, companyRepository
}

// -------- AssociateApplicationPerson tests: --------

func TestAssociateApplicationToPerson_ShouldAssociateAApplicationToAPerson(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	associatedApplicationPerson, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	assert.Equal(t, applicationPerson.ApplicationID, associatedApplicationPerson.ApplicationID)
	assert.Equal(t, applicationPerson.PersonID, associatedApplicationPerson.PersonID)
	testutil.AssertEqualFormattedDateTimes(t, applicationPerson.CreatedDate, &associatedApplicationPerson.CreatedDate)
}

func TestAssociateApplicationToPerson_ShouldAssociateAApplicationToAPersonWithOnlyRequiredFields(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	associatedApplicationPerson, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	assert.Equal(t, applicationPerson.ApplicationID, associatedApplicationPerson.ApplicationID)
	assert.Equal(t, applicationPerson.PersonID, associatedApplicationPerson.PersonID)
	assert.NotNil(t, associatedApplicationPerson.CreatedDate)
}

func TestAssociateApplicationToPerson_ShouldReturnConflictErrorIfApplicationIDAndPersonIDCombinationAlreadyExist(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	_, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(
		t,
		"conflict error on insert: ApplicationID and PersonID combination already exists in database.",
		conflictError.Error())
}

// -------- GetByID tests: --------

func TestGetByID_ShouldGetRecordsMatchingApplicationID(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

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
	_, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	applicationPersons, err := applicationPersonService.GetByID(&application1.ID, nil)
	assert.NoError(t, err)
	assert.Len(t, applicationPersons, 2)

	assert.Equal(t, applicationPersons[0].ApplicationID, application1.ID)
	assert.Equal(t, applicationPersons[0].PersonID, person2.ID)

	assert.Equal(t, applicationPersons[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationPersons[1].PersonID, person1.ID)
}

func TestApplicationPersonGetByID_ShouldGetRecordsMatchingPersonID(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

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
	_, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	applicationPersons, err := applicationPersonService.GetByID(nil, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, applicationPersons, 2)

	assert.Equal(t, applicationPersons[0].ApplicationID, application2.ID)
	assert.Equal(t, applicationPersons[0].PersonID, person1.ID)

	assert.Equal(t, applicationPersons[1].ApplicationID, application1.ID)
	assert.Equal(t, applicationPersons[1].PersonID, person1.ID)
}

func TestApplicationPersonGetByID_ShouldGetRecordsMatchingApplicationIDAndPersonID(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
	}
	_, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	persons, err := applicationPersonService.GetByID(&application1.ID, &person1.ID)
	assert.NoError(t, err)
	assert.Len(t, persons, 1)
	assert.Equal(t, application1.ID, persons[0].ApplicationID)
	assert.Equal(t, person1.ID, persons[0].PersonID)
}

// -------- GetAll tests: --------

func TestGetAllApplicationPersons_ShouldReturnAllApplicationPersons(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

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
	_, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson1)
	assert.NoError(t, err)

	applicationPerson2 := models.AssociateApplicationPerson{
		ApplicationID: application1.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson2)
	assert.NoError(t, err)

	applicationPerson3 := models.AssociateApplicationPerson{
		ApplicationID: application2.ID,
		PersonID:      person2.ID,
		CreatedDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationPersonService.AssociateApplicationPerson(&applicationPerson3)
	assert.NoError(t, err)

	personCompanies, err := applicationPersonService.GetAll()
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
	applicationPersonService, _, _, _ := setupApplicationPersonService(t)

	results, err := applicationPersonService.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- Delete tests: --------

func TestDeleteApplicationPerson_ShouldDeleteApplicationPerson(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	_, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	deleteModel := models.DeleteApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}

	err = applicationPersonService.Delete(&deleteModel)
	assert.NoError(t, err)
}

func TestDeleteApplicationPerson_ShouldReturnNotFoundErrorIfNoMatchingApplicationPersonInDatabase(t *testing.T) {
	applicationPersonService, applicationRepository, personRepository, companyRepository := setupApplicationPersonService(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := models.AssociateApplicationPerson{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}
	_, err := applicationPersonService.AssociateApplicationPerson(&applicationPerson)
	assert.NoError(t, err)

	deleteModel := models.DeleteApplicationPerson{
		ApplicationID: uuid.New(),
		PersonID:      uuid.New(),
	}

	err = applicationPersonService.Delete(&deleteModel)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t,
		"error: object not found: ApplicationPerson does not exist. applicationID: "+deleteModel.ApplicationID.String()+
			", personID: "+deleteModel.PersonID.String(), notFoundError.Error())
}
