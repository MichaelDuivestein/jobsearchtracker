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
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupPersonService(t *testing.T) (
	*services.PersonService,
	*repositories.ApplicationRepository,
	*repositories.CompanyRepository,
	*repositories.EventRepository,
	*repositories.PersonRepository,
	*repositories.ApplicationPersonRepository,
	*repositories.CompanyPersonRepository,
	*repositories.EventPersonRepository) {

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

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	var applicationPersonRepository *repositories.ApplicationPersonRepository
	err = container.Invoke(func(repository *repositories.ApplicationPersonRepository) {
		applicationPersonRepository = repository
	})
	assert.NoError(t, err)

	var companyPersonRepository *repositories.CompanyPersonRepository
	err = container.Invoke(func(repository *repositories.CompanyPersonRepository) {
		companyPersonRepository = repository
	})
	assert.NoError(t, err)

	var eventPersonRepository *repositories.EventPersonRepository
	err = container.Invoke(func(repository *repositories.EventPersonRepository) {
		eventPersonRepository = repository
	})
	assert.NoError(t, err)

	return personService,
		applicationRepository,
		companyRepository,
		eventRepository,
		personRepository,
		applicationPersonRepository,
		companyPersonRepository,
		eventPersonRepository
}

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldWork(t *testing.T) {
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

	id := uuid.New()
	nilPerson, err := personService.GetPersonById(&id)
	assert.Nil(t, nilPerson)

	assert.Error(t, err)
	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notFoundError.Error())
}

// -------- GetPersonsByName tests: --------

func TestGetPersonsByName_ShouldReturnMultiplePersons(t *testing.T) {
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Name: '"+nameToGet+"'", notFoundError.Error())
}

// -------- GetAllPersons - base tests: --------

func TestGetAllPersons_ShouldWork(t *testing.T) {
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
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

func TestGetAllPersons_ShouldReturnNilIfNoPersonsInDatabase(t *testing.T) {
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, persons)
}

// -------- GetAll - Application tests: --------

func TestPersonServiceGetAllPersons_ShouldReturnApplicationsIfIncludeApplicationsIsSetToAll(t *testing.T) {
	personService,
		applicationRepository,
		companyRepository,
		_,
		personRepository,
		applicationPersonRepository,
		_,
		_ := setupPersonService(t)

	// create persons

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	// add two companies

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	createApplication1 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &company1ID,
		RecruiterID:          &company2ID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(0),
		EstimatedCycleTime:   testutil.ToPtr(1),
		EstimatedCommuteTime: testutil.ToPtr(2),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}
	_, err := applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company1ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate persons and applications

	repositoryhelpers.AssociateApplicationPerson(t, applicationPersonRepository, *createApplication1.ID, person1ID, nil)
	repositoryhelpers.AssociateApplicationPerson(t, applicationPersonRepository, application2ID, person1ID, nil)
	repositoryhelpers.AssociateApplicationPerson(t, applicationPersonRepository, application2ID, person2ID, nil)

	// get all persons

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)

	assert.Equal(t, person1ID, persons[0].ID)
	assert.Len(t, *(persons[0]).Applications, 2)

	assert.Equal(t, application2ID, (*(*persons[0]).Applications)[0].ID)

	person1Application2 := (*(*persons[0]).Applications)[1]
	assert.Equal(t, *createApplication1.ID, person1Application2.ID)
	assert.Equal(t, createApplication1.CompanyID, person1Application2.CompanyID)
	assert.Equal(t, createApplication1.RecruiterID, person1Application2.RecruiterID)
	assert.Equal(t, createApplication1.JobTitle, person1Application2.JobTitle)
	assert.Equal(t, createApplication1.JobAdURL, person1Application2.JobAdURL)
	assert.Equal(t, createApplication1.Country, person1Application2.Country)
	assert.Equal(t, createApplication1.Area, person1Application2.Area)
	assert.Equal(t, createApplication1.RemoteStatusType.String(), person1Application2.RemoteStatusType.String())
	assert.Equal(t, createApplication1.WeekdaysInOffice, person1Application2.WeekdaysInOffice)
	assert.Equal(t, createApplication1.EstimatedCycleTime, person1Application2.EstimatedCycleTime)
	assert.Equal(t, createApplication1.EstimatedCommuteTime, person1Application2.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.ApplicationDate, person1Application2.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.CreatedDate, person1Application2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.UpdatedDate, person1Application2.UpdatedDate)

	assert.Len(t, *(persons[1]).Applications, 1)
	assert.Equal(t, application2ID, (*(*persons[1]).Applications)[0].ID)
}

func TestPersonServiceGetAllPersons_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToAllAndThereAreNoApplications(t *testing.T) {
	personService, applicationRepository, companyRepository, _, personRepository, _, _, _ := setupPersonService(t)

	// create person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all persons

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 1)
	assert.Equal(t, personID, persons[0].ID)
	assert.Nil(t, persons[0].Applications)
}

func TestPersonServiceGetAllPersons_ShouldReturnApplicationIDsIfIncludeApplicationsIsSetToIDs(t *testing.T) {
	personService,
		applicationRepository,
		companyRepository,
		_,
		personRepository,
		applicationPersonRepository,
		_,
		_ := setupPersonService(t)

	// create a person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	createApplication1 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &companyID,
		RecruiterID:          &companyID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(0),
		EstimatedCycleTime:   testutil.ToPtr(1),
		EstimatedCommuteTime: testutil.ToPtr(2),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}
	_, err := applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate person and applications

	repositoryhelpers.AssociateApplicationPerson(t, applicationPersonRepository, *createApplication1.ID, personID, nil)
	repositoryhelpers.AssociateApplicationPerson(t, applicationPersonRepository, application2ID, personID, nil)

	// get all persons

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 1)

	assert.Equal(t, personID, persons[0].ID)
	assert.Len(t, *(persons[0]).Applications, 2)

	assert.Equal(t, application2ID, (*(*persons[0]).Applications)[0].ID)

	person1Application2 := (*(*persons[0]).Applications)[1]
	assert.Equal(t, *createApplication1.ID, person1Application2.ID)
	assert.Nil(t, person1Application2.CompanyID)
	assert.Nil(t, person1Application2.RecruiterID)
	assert.Nil(t, person1Application2.JobTitle)
	assert.Nil(t, person1Application2.JobAdURL)
	assert.Nil(t, person1Application2.Country)
	assert.Nil(t, person1Application2.Area)
	assert.Nil(t, person1Application2.RemoteStatusType)
	assert.Nil(t, person1Application2.WeekdaysInOffice)
	assert.Nil(t, person1Application2.EstimatedCycleTime)
	assert.Nil(t, person1Application2.EstimatedCommuteTime)
	assert.Nil(t, person1Application2.ApplicationDate)
	assert.Nil(t, person1Application2.CreatedDate)
	assert.Nil(t, person1Application2.UpdatedDate)
}

func TestPersonServiceGetAllPersons_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToIDsAndThereAreNoApplications(t *testing.T) {
	personService,
		applicationRepository,
		companyRepository, _,
		personRepository,
		_,
		_,
		_ := setupPersonService(t)

	// create person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all persons

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 1)
	assert.Equal(t, personID, persons[0].ID)
	assert.Nil(t, persons[0].Applications)
}

func TestPersonServiceGetAllPersons_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToNone(t *testing.T) {
	personService,
		applicationRepository,
		companyRepository,
		_,
		personRepository,
		applicationPersonRepository,
		_,
		_ := setupPersonService(t)

	// create a person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	applicationID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate person and applications

	repositoryhelpers.AssociateApplicationPerson(t, applicationPersonRepository, applicationID, personID, nil)

	// get all persons

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, persons)
	assert.Len(t, persons, 1)

	assert.Equal(t, personID, persons[0].ID)
	assert.Nil(t, (persons[0]).Applications)
}

// -------- GetAllPersons - companies tests: --------

func TestGetAllPersons_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	personService, _, companyRepository, _, _, _, companyPersonRepository, _ := setupPersonService(t)

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

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
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

func TestGetAllPersons_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personService, _, companyRepository, _, _, _, _, _ := setupPersonService(t)

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

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
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

func TestGetAllPersons_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	personService, _, companyRepository, _, _, _, companyPersonRepository, _ := setupPersonService(t)

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

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
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

func TestGetAllPersons_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personService, _, companyRepository, _, _, _, _, _ := setupPersonService(t)

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

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
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

func TestGetAllPersons_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	personService, _, companyRepository, _, _, _, companyPersonRepository, _ := setupPersonService(t)

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

	persons, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
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

// -------- GetAllPersons events tests: --------

func TestGetAllPersons_ShouldReturnEventsIfIncludeEventsIsSetToAll(t *testing.T) {
	personService, _, _, eventRepository, personRepository, _, _, eventPersonRepository := setupPersonService(t)

	// create a person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add two events

	event1ToInsert := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   requests.EventTypeApplied,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   time.Now().AddDate(0, 0, 5),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := eventRepository.Create(&event1ToInsert)
	assert.NoError(t, err)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 6))).ID

	// associate persons and events

	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, *event1ToInsert.ID, personID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event2ID, personID, nil)

	// get all persons

	results, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedPerson := results[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.NotNil(t, retrievedPerson.Events)
	assert.Len(t, *retrievedPerson.Events, 2)

	assert.Equal(t, event2ID, (*retrievedPerson.Events)[0].ID)

	event2 := (*retrievedPerson.Events)[1]
	assert.Equal(t, *event1ToInsert.ID, event2.ID)
	assert.Equal(t, event1ToInsert.EventType.String(), event2.EventType.String())
	assert.Equal(t, event1ToInsert.Description, event2.Description)
	assert.Equal(t, event1ToInsert.Notes, event2.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &event1ToInsert.EventDate, event2.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, event1ToInsert.CreatedDate, event2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, event1ToInsert.UpdatedDate, event2.UpdatedDate)
}

func TestGetAllPersons_ShouldReturnNoEventsIfIncludeEventsIsSetToAllAndThereAreNoEventPersonsInRepository(t *testing.T) {
	personService, _, _, _, personRepository, _, _, _ := setupPersonService(t)

	// create a person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// get all persons

	results, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedPerson := results[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.Nil(t, retrievedPerson.Events)
}

func TestGetAllPersons_ShouldReturnEventIDsIfIncludeEventsIsSetToIds(t *testing.T) {
	personService, _, _, eventRepository, personRepository, _, _, eventPersonRepository := setupPersonService(t)

	// create a person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add two events

	event1ToInsert := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   requests.EventTypeApplied,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   time.Now().AddDate(0, 0, 5),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := eventRepository.Create(&event1ToInsert)
	assert.NoError(t, err)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 6))).ID

	// associate persons and events

	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, *event1ToInsert.ID, personID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event2ID, personID, nil)

	// get all persons

	results, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedPerson := results[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.NotNil(t, retrievedPerson.Events)
	assert.Len(t, *retrievedPerson.Events, 2)

	assert.Equal(t, event2ID, (*retrievedPerson.Events)[0].ID)

	event2 := (*retrievedPerson.Events)[1]
	assert.Equal(t, *event1ToInsert.ID, event2.ID)
	assert.Nil(t, event2.EventType)
	assert.Nil(t, event2.Description)
	assert.Nil(t, event2.Notes)
	assert.Nil(t, event2.EventDate)
	assert.Nil(t, event2.CreatedDate)
	assert.Nil(t, event2.UpdatedDate)
}

func TestGetAllPersons_ShouldReturnNoEventIDsIfIncludeEventsIsSetToIDsAndThereAreNoEventPersonsInRepository(t *testing.T) {
	personService, _, _, _, personRepository, _, _, _ := setupPersonService(t)

	// create a person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// get all persons

	results, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedPerson := results[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.Nil(t, retrievedPerson.Events)
}

func TestGetAllPersons_ShouldReturnNoEventsIfIncludeEventsIsSetToNone(t *testing.T) {
	personService, _, _, eventRepository, personRepository, _, _, eventPersonRepository := setupPersonService(t)

	// create a person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add an event and associate it to the person

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, personID, nil)

	// get all persons

	results, err := personService.GetAllPersons(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedPerson := results[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.Nil(t, retrievedPerson.Events)
}

// -------- UpdatePerson tests: --------

func TestUpdatePerson_ShouldWork(t *testing.T) {
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

	personToUpdate := models.UpdatePerson{
		ID:    uuid.New(),
		Notes: testutil.ToPtr("Random Notes"),
	}

	err := personService.UpdatePerson(&personToUpdate)
	assert.NoError(t, err)
}

// -------- DeletePerson tests: --------

func TestDeletePerson_ShouldWork(t *testing.T) {
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

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
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+personToInsert.ID.String()+"'", notFoundError.Error())
}

func TestDeletePerson_ShouldReturnNotFoundErrorIfIdToDeleteDoesNotExist(t *testing.T) {
	personService, _, _, _, _, _, _, _ := setupPersonService(t)

	id := uuid.New()
	err := personService.DeletePerson(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Person does not exist. ID: "+id.String(), notFoundError.Error())
}
