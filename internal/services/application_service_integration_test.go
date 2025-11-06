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

func setupApplicationService(t *testing.T) (
	*services.ApplicationService,
	*repositories.CompanyRepository,
	*repositories.PersonRepository,
	*repositories.ApplicationPersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupApplicationServiceTestContainer(t, *config)

	var applicationService *services.ApplicationService
	err := container.Invoke(func(applicationSvc *services.ApplicationService) {
		applicationService = applicationSvc
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

	var applicationPersonRepository *repositories.ApplicationPersonRepository
	err = container.Invoke(func(repository *repositories.ApplicationPersonRepository) {
		applicationPersonRepository = repository
	})
	assert.NoError(t, err)

	return applicationService, companyRepository, personRepository, applicationPersonRepository
}

// -------- CreateApplication tests: --------

func TestCreateApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	applicationToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &company.ID,
		RecruiterID:          &recruiter.ID,
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(9),
		EstimatedCycleTime:   testutil.ToPtr(8),
		EstimatedCommuteTime: testutil.ToPtr(7),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}

	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	assert.Equal(t, *applicationToInsert.ID, insertedApplication.ID)
	assert.Equal(t, applicationToInsert.CompanyID, insertedApplication.CompanyID)
	assert.Equal(t, applicationToInsert.RecruiterID, insertedApplication.RecruiterID)
	assert.Equal(t, applicationToInsert.JobTitle, insertedApplication.JobTitle)
	assert.Equal(t, applicationToInsert.JobAdURL, insertedApplication.JobAdURL)
	assert.Equal(t, applicationToInsert.Country, insertedApplication.Country)
	assert.Equal(t, applicationToInsert.Area, insertedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType.String(), insertedApplication.RemoteStatusType.String())
	assert.Equal(t, applicationToInsert.WeekdaysInOffice, insertedApplication.WeekdaysInOffice)
	assert.Equal(t, applicationToInsert.EstimatedCycleTime, insertedApplication.EstimatedCycleTime)
	assert.Equal(t, applicationToInsert.EstimatedCommuteTime, insertedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, insertedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, insertedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, insertedApplication.UpdatedDate)
}

func TestCreateApplication_ShouldHandleEmptyFields(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	application := models.CreateApplication{
		CompanyID:        &company.ID,
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	createdDateApproximation := time.Now()
	insertedApplication, err := applicationService.CreateApplication(&application)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	assert.NotNil(t, insertedApplication.ID)
	assert.Equal(t, application.CompanyID, insertedApplication.CompanyID)
	assert.Nil(t, insertedApplication.RecruiterID)
	assert.Nil(t, insertedApplication.JobTitle)
	assert.Equal(t, insertedApplication.JobAdURL, insertedApplication.JobAdURL)
	assert.Nil(t, insertedApplication.Country)
	assert.Nil(t, insertedApplication.Area)
	assert.Equal(t, models.RemoteStatusTypeHybrid, insertedApplication.RemoteStatusType.String())
	assert.Nil(t, insertedApplication.WeekdaysInOffice)
	assert.Nil(t, insertedApplication.EstimatedCycleTime)
	assert.Nil(t, insertedApplication.EstimatedCommuteTime)
	assert.Nil(t, insertedApplication.ApplicationDate)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, insertedApplication.CreatedDate, time.Second)
	assert.Nil(t, insertedApplication.UpdatedDate)
}

func TestCreateApplication_ShouldReturnErrorIfCompanyIdIsNotInCompany(t *testing.T) {
	applicationService, _, _, _ := setupApplicationService(t)

	application := models.CreateApplication{
		CompanyID:        testutil.ToPtr(uuid.New()),
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	response, err := applicationService.CreateApplication(&application)
	assert.Nil(t, response)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

func TestCreateApplication_ShouldReturnErrorIfRecruiterIdIsNotInCompany(t *testing.T) {
	applicationService, _, _, _ := setupApplicationService(t)

	application := models.CreateApplication{
		RecruiterID:      testutil.ToPtr(uuid.New()),
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	response, err := applicationService.CreateApplication(&application)
	assert.Nil(t, response)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
}

// -------- GetApplicationById tests: --------

func TestGetApplicationById_ShouldWork(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	applicationToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &company.ID,
		RecruiterID:          &recruiter.ID,
		JobTitle:             testutil.ToPtr("JobTitle"),
		JobAdURL:             testutil.ToPtr("JobAdURL"),
		Country:              testutil.ToPtr("SomeCountry"),
		Area:                 testutil.ToPtr("SomeArea"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(9),
		EstimatedCycleTime:   testutil.ToPtr(8),
		EstimatedCommuteTime: testutil.ToPtr(7),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}

	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	retrievedApplication, err := applicationService.GetApplicationById(applicationToInsert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, applicationToInsert.CompanyID, retrievedApplication.CompanyID)
	assert.Equal(t, applicationToInsert.RecruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, applicationToInsert.JobAdURL, retrievedApplication.JobAdURL)
	assert.Equal(t, applicationToInsert.Country, retrievedApplication.Country)
	assert.Equal(t, applicationToInsert.Area, retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, *applicationToInsert.WeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, *applicationToInsert.EstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, *applicationToInsert.EstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, insertedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, insertedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, insertedApplication.UpdatedDate)
}

func TestGetApplicationById_ShouldReturnNotFoundErrorForAnIdThatDoesNotExist(t *testing.T) {
	applicationService, _, _, _ := setupApplicationService(t)

	id := uuid.New()
	nilApplication, err := applicationService.GetApplicationById(&id)
	assert.Nil(t, nilApplication)

	assert.NotNil(t, err)
	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notFoundError.Error())
}

// -------- GetApplicationsByJobTitle tests: --------

func TestGetApplicationsByJobTitle_ShouldReturnApplications(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// insert applications

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	application1ToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &company.ID,
		RecruiterID:          &recruiter.ID,
		JobTitle:             testutil.ToPtr("developer"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Country"),
		Area:                 testutil.ToPtr("Area"),
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedApplication, err := applicationService.CreateApplication(&application1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	application2ToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        &company.ID,
		JobTitle:         testutil.ToPtr("Backend Developer"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	_, err = applicationService.CreateApplication(&application2ToInsert)
	assert.NoError(t, err)

	application3ToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      &recruiter.ID,
		JobTitle:         testutil.ToPtr("utvecklare till en företag"),
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}
	_, err = applicationService.CreateApplication(&application3ToInsert)
	assert.NoError(t, err)

	// GetByJobTitle

	applications, err := applicationService.GetApplicationsByJobTitle(testutil.ToPtr("developer"))
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Len(t, applications, 2)

	assert.Equal(t, *application1ToInsert.ID, applications[0].ID)
	assert.Equal(t, application1ToInsert.CompanyID, applications[0].CompanyID)
	assert.Equal(t, application1ToInsert.RecruiterID, applications[0].RecruiterID)
	assert.Equal(t, application1ToInsert.JobAdURL, applications[0].JobAdURL)
	assert.Equal(t, application1ToInsert.Country, applications[0].Country)
	assert.Equal(t, application1ToInsert.Area, applications[0].Area)
	assert.Equal(t, application1ToInsert.RemoteStatusType.String(), applications[0].RemoteStatusType.String())
	assert.Equal(t, application1ToInsert.WeekdaysInOffice, applications[0].WeekdaysInOffice)
	assert.Equal(t, application1ToInsert.EstimatedCycleTime, applications[0].EstimatedCycleTime)
	assert.Equal(t, application1ToInsert.EstimatedCommuteTime, applications[0].EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, application1ToInsert.ApplicationDate, applications[0].ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, application1ToInsert.CreatedDate, applications[0].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, application1ToInsert.UpdatedDate, applications[0].UpdatedDate)

	assert.Equal(t, application2ToInsert.JobTitle, applications[1].JobTitle)
}

func TestGetApplicationsByJobTitle_ShouldReturnNotFoundErrorIfNoNamesMatch(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// insert applications

	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      &recruiter.ID,
		JobTitle:         testutil.ToPtr("Backend Engineer"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	_, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)

	// GetByJobTitle

	jobTitleToGet := "Developer"
	applications, err := applicationService.GetApplicationsByJobTitle(&jobTitleToGet)
	assert.Nil(t, applications)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: JobTitle: '"+jobTitleToGet+"'", notFoundError.Error())
}

// -------- GetAllApplications - Base tests: --------

func TestGetAlLApplications_ShouldWork(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// insert applications

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	applicationToInsert1 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &company.ID,
		RecruiterID:          &recruiter.ID,
		JobTitle:             testutil.ToPtr("Job Title 1"),
		JobAdURL:             testutil.ToPtr("Job Ad URL 1"),
		Country:              testutil.ToPtr("Some Country 1"),
		Area:                 testutil.ToPtr("Some Area 1"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(12),
		EstimatedCycleTime:   testutil.ToPtr(43),
		EstimatedCommuteTime: testutil.ToPtr(77),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 12)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -8)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err := applicationService.CreateApplication(&applicationToInsert1)
	assert.NoError(t, err)

	applicationToInsert2 := models.CreateApplication{
		CompanyID:        &company.ID,
		JobTitle:         testutil.ToPtr("JobTitle2"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	insertedApplication2, err := applicationService.CreateApplication(&applicationToInsert2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication2)

	// getAll

	applications, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Len(t, applications, 2)

	assert.Equal(t, insertedApplication2.JobTitle, applications[0].JobTitle)

	application1 := applications[1]
	assert.Equal(t, *applicationToInsert1.ID, application1.ID)
	assert.Equal(t, company.ID, *application1.CompanyID)
	assert.Equal(t, recruiter.ID, *application1.RecruiterID)
	assert.Equal(t, applicationToInsert1.JobTitle, application1.JobTitle)
	assert.Equal(t, applicationToInsert1.JobAdURL, application1.JobAdURL)
	assert.Equal(t, applicationToInsert1.Country, application1.Country)
	assert.Equal(t, applicationToInsert1.Area, application1.Area)
	assert.Equal(t, applicationToInsert1.RemoteStatusType.String(), application1.RemoteStatusType.String())
	assert.Equal(t, applicationToInsert1.WeekdaysInOffice, application1.WeekdaysInOffice)
	assert.Equal(t, applicationToInsert1.EstimatedCycleTime, application1.EstimatedCycleTime)
	assert.Equal(t, applicationToInsert1.EstimatedCommuteTime, application1.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert1.ApplicationDate, application1.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert1.CreatedDate, application1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert1.UpdatedDate, application1.UpdatedDate)
}

func TestGetAlLApplications_ShouldReturnNilIfNoApplicationsInDatabase(t *testing.T) {
	applicationService, _, _, _ := setupApplicationService(t)

	applications, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)
	assert.Nil(t, applications)
}

// -------- GetAllApplications - Company tests: --------

func TestGetAllApplications_ShouldReturnCompanyIfIncludeCompanyIsSetToAll(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.NotNil(t, retrievedApplication.Company)

	assert.Equal(t, *companyToInsert.ID, retrievedApplication.Company.ID)
	assert.Equal(t, companyToInsert.Name, *retrievedApplication.Company.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), retrievedApplication.Company.CompanyType.String())
	assert.Equal(t, companyToInsert.Notes, retrievedApplication.Company.Notes)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.LastContact, retrievedApplication.Company.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.CreatedDate, retrievedApplication.Company.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.UpdatedDate, retrievedApplication.Company.UpdatedDate)
}

func TestGetAllApplications_ShouldReturnNoCompanyIfIncludeCompanyIsSetToAllAndThereIsNoCompany(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        nil,
		RecruiterID:      recruiterID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company)
}

func TestGetAllApplications_ShouldReturnCompanyWithOnlyIDIfIncludeCompanyIsSetToIDs(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.NotNil(t, retrievedApplication.Company)

	assert.Equal(t, *companyToInsert.ID, retrievedApplication.Company.ID)
	assert.Nil(t, retrievedApplication.Company.Name)
	assert.Nil(t, retrievedApplication.Company.CompanyType)
	assert.Nil(t, retrievedApplication.Company.Notes)
	assert.Nil(t, retrievedApplication.Company.LastContact)
	assert.Nil(t, retrievedApplication.Company.CreatedDate)
	assert.Nil(t, retrievedApplication.Company.UpdatedDate)
}

func TestGetAllApplications_ShouldReturnNoCompanyIfIncludeCompanyIsSetToIDsAndThereIsNoCompany(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        nil,
		RecruiterID:      recruiterID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company)
}

func TestGetAllApplications_ShouldReturnNoCompanyIfIncludeCompanyIsSetToNone(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title1"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication1, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company)
}

// -------- GetAllApplications - Recruiter tests: --------

func TestGetAllApplications_ShouldReturnRecruiterIfIncludeRecruiterIsSetToAll(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      companyToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.RecruiterID)
	assert.NotNil(t, retrievedApplication.RecruiterID)

	assert.Equal(t, *companyToInsert.ID, retrievedApplication.Recruiter.ID)
	assert.Equal(t, companyToInsert.Name, *retrievedApplication.Recruiter.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), retrievedApplication.Recruiter.CompanyType.String())
	assert.Equal(t, companyToInsert.Notes, retrievedApplication.Recruiter.Notes)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.LastContact, retrievedApplication.Recruiter.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.CreatedDate, retrievedApplication.Recruiter.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.UpdatedDate, retrievedApplication.Recruiter.UpdatedDate)
}

func TestGetAllApplications_ShouldReturnNoRecruiterIfIncludeRecruiterIsSetToAllAndThereIsNoRecruiter(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyID,
		RecruiterID:      nil,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.RecruiterID)
	assert.Nil(t, retrievedApplication.Recruiter)
}

func TestGetAllApplications_ShouldReturnRecruiterWithOnlyIDIfIncludeRecruiterIsSetToIDs(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      companyToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.RecruiterID)
	assert.NotNil(t, retrievedApplication.Recruiter)

	assert.Equal(t, *companyToInsert.ID, retrievedApplication.Recruiter.ID)
	assert.Nil(t, retrievedApplication.Recruiter.Name)
	assert.Nil(t, retrievedApplication.Recruiter.CompanyType)
	assert.Nil(t, retrievedApplication.Recruiter.Notes)
	assert.Nil(t, retrievedApplication.Recruiter.LastContact)
	assert.Nil(t, retrievedApplication.Recruiter.CreatedDate)
	assert.Nil(t, retrievedApplication.Recruiter.UpdatedDate)
}

func TestGetAllApplications_ShouldReturnNoRecruiterIfIncludeRecruiterIsSetToIDsAndThereIsNoRecruiter(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyID,
		RecruiterID:      nil,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.RecruiterID)
	assert.Nil(t, retrievedApplication.Recruiter)
}

func TestGetAllApplications_ShouldReturnNoRecruiterIfIncludeRecruiterIsSetToNone(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// create an application

	companyToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      companyToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title1"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication1, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)

	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.RecruiterID)
	assert.Nil(t, retrievedApplication.Recruiter)
}

// -------- GetAllApplications - Persons tests: --------

func TestGetAllApplications_ShouldReturnPersonsIfIncludePersonsIsSetToAll(t *testing.T) {
	applicationService, companyRepository, personRepository, applicationPersonRepository :=
		setupApplicationService(t)

	// create application

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      &companyID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	person1ToInsert := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1Name",
		PersonType:  models.PersonTypeOther,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err = personRepository.Create(&person1ToInsert)
	assert.NoError(t, err)

	person2ID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	repositoryhelpers.AssociateApplicationPerson(
		t,
		applicationPersonRepository,
		insertedApplication.ID,
		*person1ToInsert.ID,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5)))

	repositoryhelpers.AssociateApplicationPerson(
		t,
		applicationPersonRepository,
		insertedApplication.ID,
		person2ID,
		testutil.ToPtr(time.Now().AddDate(0, 0, 6)))

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.NotNil(t, retrievedApplication.Persons)
	assert.Len(t, *retrievedApplication.Persons, 2)

	assert.Equal(t, person2ID, (*retrievedApplication.Persons)[1].ID)

	person1 := (*retrievedApplication.Persons)[0]
	assert.Equal(t, *person1ToInsert.ID, person1.ID)
	assert.Equal(t, person1ToInsert.Name, *person1.Name)
	assert.Equal(t, person1ToInsert.PersonType.String(), person1.PersonType.String())
	assert.Equal(t, person1ToInsert.Email, person1.Email)
	assert.Equal(t, person1ToInsert.Phone, person1.Phone)
	assert.Equal(t, person1ToInsert.Notes, person1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, person1ToInsert.CreatedDate, person1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, person1ToInsert.UpdatedDate, person1.UpdatedDate)
	assert.Nil(t, person1.Companies)
}

func TestGetAllApplications_ShouldReturnNoPersonsIfIncludePersonsIsSetToAllAndThereAreNoApplicationPersons(t *testing.T) {
	applicationService, companyRepository, personRepository, _ := setupApplicationService(t)

	// create application

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      &companyID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.Persons)
}

func TestGetAllApplications_ShouldReturnPersonIDsIfIncludePersonsIsSetToIDs(t *testing.T) {
	applicationService, companyRepository, personRepository, applicationPersonRepository :=
		setupApplicationService(t)

	// create application

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      &companyID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	person1ToInsert := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1Name",
		PersonType:  models.PersonTypeOther,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err = personRepository.Create(&person1ToInsert)
	assert.NoError(t, err)

	person2ID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	repositoryhelpers.AssociateApplicationPerson(
		t,
		applicationPersonRepository,
		insertedApplication.ID,
		*person1ToInsert.ID,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5)))

	repositoryhelpers.AssociateApplicationPerson(
		t,
		applicationPersonRepository,
		insertedApplication.ID,
		person2ID,
		testutil.ToPtr(time.Now().AddDate(0, 0, 6)))

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.NotNil(t, retrievedApplication.Persons)
	assert.Len(t, *retrievedApplication.Persons, 2)

	assert.Equal(t, person2ID, (*retrievedApplication.Persons)[1].ID)

	person1 := (*retrievedApplication.Persons)[0]
	assert.Equal(t, *person1ToInsert.ID, person1.ID)
	assert.Nil(t, person1.Name)
	assert.Nil(t, person1.PersonType)
	assert.Nil(t, person1.Email)
	assert.Nil(t, person1.Phone)
	assert.Nil(t, person1.Notes)
	assert.Nil(t, person1.CreatedDate)
	assert.Nil(t, person1.CreatedDate)
	assert.Nil(t, person1.Companies)
}

func TestGetAllApplications_ShouldReturnNoPersonsIfIncludePersonsIsSetToIDsAndThereAreNoApplicationPersons(t *testing.T) {
	applicationService, companyRepository, personRepository, _ := setupApplicationService(t)

	// create application

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      &companyID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.Persons)
}

func TestGetAllApplications_ShouldReturnNoPersonsIfIncludePersonsIsSetToNone(t *testing.T) {
	applicationService, companyRepository, personRepository, applicationPersonRepository :=
		setupApplicationService(t)

	// create application

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      &companyID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	repositoryhelpers.AssociateApplicationPerson(
		t,
		applicationPersonRepository,
		insertedApplication.ID,
		personID,
		testutil.ToPtr(time.Now().AddDate(0, 0, 6)))

	// get all applications

	results, err := applicationService.GetAllApplications(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.Persons)
}

// -------- UpdateApplication tests: --------

func TestUpdateApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// insert application

	id := uuid.New()
	originalCompany := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	originalRecruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	applicationToInsert := models.CreateApplication{
		ID:                   &id,
		CompanyID:            &originalCompany.ID,
		RecruiterID:          &originalRecruiter.ID,
		JobTitle:             testutil.ToPtr("OriginalJobTitle"),
		JobAdURL:             testutil.ToPtr("OriginalJobAdURL"),
		Country:              testutil.ToPtr("OriginalCountry"),
		Area:                 testutil.ToPtr("OriginalArea"),
		RemoteStatusType:     models.RemoteStatusTypeOffice,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(1, 0, 0)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(2, 0, 0)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(3, 0, 0)),
	}
	_, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)

	// update application

	newCompany := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	newRecruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	var newRemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeOffice

	applicationToUpdate := models.UpdateApplication{
		ID:                   id,
		CompanyID:            &newCompany.ID,
		RecruiterID:          &newRecruiter.ID,
		JobTitle:             testutil.ToPtr("NewJobTitle"),
		JobAdURL:             testutil.ToPtr("NewJobAdURL"),
		Country:              testutil.ToPtr("NewCountry"),
		Area:                 testutil.ToPtr("NewArea"),
		RemoteStatusType:     &newRemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(4),
		EstimatedCycleTime:   testutil.ToPtr(5),
		EstimatedCommuteTime: testutil.ToPtr(6),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}

	updatedDateApproximation := time.Now()
	err = applicationService.UpdateApplication(&applicationToUpdate)
	assert.NoError(t, err)

	// get ById

	application, err := applicationService.GetApplicationById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, application)

	assert.Equal(t, id, application.ID)
	assert.Equal(t, newCompany.ID, *application.CompanyID)
	assert.Equal(t, newRecruiter.ID, *application.RecruiterID)
	assert.Equal(t, applicationToUpdate.JobTitle, application.JobTitle)
	assert.Equal(t, applicationToUpdate.JobAdURL, application.JobAdURL)
	assert.Equal(t, applicationToUpdate.Country, application.Country)
	assert.Equal(t, applicationToUpdate.Area, application.Area)
	assert.Equal(t, applicationToUpdate.RemoteStatusType.String(), application.RemoteStatusType.String())
	assert.Equal(t, applicationToUpdate.WeekdaysInOffice, application.WeekdaysInOffice)
	assert.Equal(t, applicationToUpdate.EstimatedCycleTime, application.EstimatedCycleTime)
	assert.Equal(t, applicationToUpdate.EstimatedCommuteTime, application.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToUpdate.ApplicationDate, application.ApplicationDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, application.UpdatedDate, time.Second)
}

func TestUpdateApplication_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	applicationService, _, _, _ := setupApplicationService(t)

	applicationToUpdate := models.UpdateApplication{
		ID:       uuid.New(),
		JobTitle: testutil.ToPtr("JobTitle"),
	}

	err := applicationService.UpdateApplication(&applicationToUpdate)
	assert.NoError(t, err)
}

// -------- DeleteApplication tests: --------

func TestDeleteApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository, _, _ := setupApplicationService(t)

	// insert application

	id := uuid.New()
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	applicationToInsert := models.CreateApplication{
		ID:               &id,
		RecruiterID:      &recruiter.ID,
		JobAdURL:         testutil.ToPtr("JobAdURL"),
		RemoteStatusType: models.PersonTypeUnknown,
	}
	_, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)

	// delete application

	err = applicationService.DeleteApplication(&id)
	assert.NoError(t, err)

	//ensure that application is deleted

	retrievedApplication, err := applicationService.GetApplicationById(&id)
	assert.Nil(t, retrievedApplication)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notFoundError.Error())
}

func TestDeleteApplication_ShouldReturnNotFoundErrorIfIdToDeleteDoesNotExist(t *testing.T) {
	applicationService, _, _, _ := setupApplicationService(t)

	id := uuid.New()
	err := applicationService.DeleteApplication(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Application does not exist. ID: "+id.String(), notFoundError.Error())
}
