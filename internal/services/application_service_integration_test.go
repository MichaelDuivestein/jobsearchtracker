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

func setupApplicationService(t *testing.T) (*services.ApplicationService, *repositories.CompanyRepository) {
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

	return applicationService, companyRepository
}

// -------- CreateApplication tests: --------

func TestCreateApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

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
	assert.Equal(t, company.ID, *insertedApplication.CompanyID)
	assert.Equal(t, recruiter.ID, *insertedApplication.RecruiterID)
	assert.Equal(t, applicationToInsert.JobTitle, insertedApplication.JobTitle)
	assert.Equal(t, applicationToInsert.JobAdURL, insertedApplication.JobAdURL)
	assert.Equal(t, applicationToInsert.Country, insertedApplication.Country)
	assert.Equal(t, applicationToInsert.Area, insertedApplication.Area)
	assert.Equal(t, models.RemoteStatusTypeHybrid, insertedApplication.RemoteStatusType.String())
	assert.Equal(t, applicationToInsert.WeekdaysInOffice, insertedApplication.WeekdaysInOffice)
	assert.Equal(t, applicationToInsert.EstimatedCycleTime, insertedApplication.EstimatedCycleTime)
	assert.Equal(t, applicationToInsert.EstimatedCommuteTime, insertedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, insertedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, insertedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, insertedApplication.UpdatedDate)
}

func TestCreateApplication_ShouldHandleEmptyFields(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	application := models.CreateApplication{
		CompanyID:        &company.ID,
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	createdDateApproximation := time.Now().Format(time.RFC3339)
	insertedApplication, err := applicationService.CreateApplication(&application)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	assert.NotNil(t, insertedApplication.ID)
	assert.Equal(t, company.ID, *insertedApplication.CompanyID)
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

	insertedCompanyCreatedDate := insertedApplication.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, insertedCompanyCreatedDate)
	assert.Nil(t, insertedApplication.UpdatedDate)
}

func TestCreateApplication_ShouldReturnErrorIfCompanyIdIsNotInCompany(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	application := models.CreateApplication{
		CompanyID:        testutil.ToPtr(uuid.New()),
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	response, err := applicationService.CreateApplication(&application)
	assert.Nil(t, response)
	assert.Error(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: Foreign key does not exist", err.Error())
}

func TestCreateApplication_ShouldReturnErrorIfRecruiterIdIsNotInCompany(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	application := models.CreateApplication{
		RecruiterID:      testutil.ToPtr(uuid.New()),
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	response, err := applicationService.CreateApplication(&application)
	assert.Nil(t, response)
	assert.Error(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: Foreign key does not exist", err.Error())
}

// -------- GetApplicationById tests: --------

func TestGetApplicationById_ShouldWork(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	id := uuid.New()

	applicationToInsert := models.CreateApplication{
		ID:                   &id,
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

	retrievedApplication, err := applicationService.GetApplicationById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

}

func TestGetApplicationById_ShouldReturnNotFoundErrorForAnIdThatDoesNotExist(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	id := uuid.New()
	nilApplication, err := applicationService.GetApplicationById(&id)
	assert.Nil(t, nilApplication)

	assert.NotNil(t, err)
	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
}

// -------- GetApplicationsByJobTitle tests: --------

func TestGetAllByJobTitle_ShouldReturnApplication(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// insert applications

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	jobTitle := "Some Job Title"

	applicationToInsert := models.CreateApplication{
		CompanyID:        &company.ID,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// GetByName
	applications, err := applicationService.GetApplicationsByJobTitle(insertedApplication.JobTitle)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Len(t, applications, 1)

	assert.Equal(t, jobTitle, *applications[0].JobTitle)
}

func TestGetApplicationsByJobTitle_ShouldReturnMultipleApplications(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// insert applications

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	applicationToInsert1 := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        &company.ID,
		JobTitle:         testutil.ToPtr("developer"),
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}
	_, err := applicationService.CreateApplication(&applicationToInsert1)
	assert.NoError(t, err)

	applicationToInsert2 := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        &company.ID,
		JobTitle:         testutil.ToPtr("Backend Developer"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	_, err = applicationService.CreateApplication(&applicationToInsert2)
	assert.NoError(t, err)

	applicationToInsert3 := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        &company.ID,
		JobTitle:         testutil.ToPtr("utvecklare till en företag"),
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}
	_, err = applicationService.CreateApplication(&applicationToInsert3)
	assert.NoError(t, err)

	// GetByJobTitle

	jobTitleToGet := "developer"
	applications, err := applicationService.GetApplicationsByJobTitle(&jobTitleToGet)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Len(t, applications, 2)

	assert.Equal(t, *applicationToInsert1.ID, applications[0].ID)
	assert.Equal(t, *applicationToInsert2.ID, applications[1].ID)
}

func TestGetApplicationsByJobTitle_ShouldReturnNotFoundErrorIfNoNamesMatch(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

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
	assert.Equal(t, "error: object not found: JobTitle: '"+jobTitleToGet+"'", err.Error())
}

// -------- GetAllApplications tests: --------

func TestGetAlLApplications_ShouldWork(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

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
	_, err = applicationService.CreateApplication(&applicationToInsert2)
	assert.NoError(t, err)

	// getAll

	applications, err := applicationService.GetAllApplications(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Len(t, applications, 2)

	assert.Equal(t, applicationToInsert2.JobTitle, applications[0].JobTitle)

	insertedApplication1 := applications[1]
	assert.Equal(t, *applicationToInsert1.ID, insertedApplication1.ID)
	assert.Equal(t, company.ID, *insertedApplication1.CompanyID)
	assert.Equal(t, recruiter.ID, *insertedApplication1.RecruiterID)
	assert.Equal(t, applicationToInsert1.JobTitle, insertedApplication1.JobTitle)
	assert.Equal(t, applicationToInsert1.JobAdURL, insertedApplication1.JobAdURL)
	assert.Equal(t, applicationToInsert1.Country, insertedApplication1.Country)
	assert.Equal(t, applicationToInsert1.Area, insertedApplication1.Area)
	assert.Equal(t, applicationToInsert1.RemoteStatusType.String(), insertedApplication1.RemoteStatusType.String())
	assert.Equal(t, applicationToInsert1.WeekdaysInOffice, insertedApplication1.WeekdaysInOffice)
	assert.Equal(t, applicationToInsert1.EstimatedCycleTime, insertedApplication1.EstimatedCycleTime)
	assert.Equal(t, applicationToInsert1.EstimatedCommuteTime, insertedApplication1.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert1.ApplicationDate, insertedApplication1.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert1.CreatedDate, insertedApplication1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert1.UpdatedDate, insertedApplication1.UpdatedDate)
}

func TestGetAlLApplications_ShouldReturnNilIfNoApplicationsInDatabase(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	applications, err := applicationService.GetAllApplications(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, applications)
}

func TestGetAllApplications_ShouldReturnCompanyIfIncludeCompanyIsSetToAll(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// create an application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

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
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyToInsert.ID,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, *applicationToInsert.JobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, *applicationToInsert.Country, *retrievedApplication.Country)
	assert.Equal(t, *applicationToInsert.Area, *retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, *applicationToInsert.WeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, *applicationToInsert.EstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, *applicationToInsert.EstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, retrievedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, retrievedApplication.UpdatedDate)
	assert.NotNil(t, retrievedApplication.Company)

	assert.Equal(t, retrievedApplication.Company.ID, *retrievedApplication.CompanyID)
	assert.Equal(t, companyToInsert.Name, *retrievedApplication.Company.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), retrievedApplication.Company.CompanyType.String())
	assert.Equal(t, companyToInsert.Notes, retrievedApplication.Company.Notes)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.LastContact, retrievedApplication.Company.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.CreatedDate, retrievedApplication.Company.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.UpdatedDate, retrievedApplication.Company.UpdatedDate)
}

func TestGetAllApplications_ShouldReturnNoCompanyIfIncludeCompanyIsSetToAllAndThereIsNoCompany(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// create an application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            nil,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, *applicationToInsert.JobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, *applicationToInsert.Country, *retrievedApplication.Country)
	assert.Equal(t, *applicationToInsert.Area, *retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, *applicationToInsert.WeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, *applicationToInsert.EstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, *applicationToInsert.EstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, retrievedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, retrievedApplication.UpdatedDate)
	assert.Nil(t, retrievedApplication.Company)
}

func TestGetAllApplications_ShouldReturnCompanyWithOnlyIDIfIncludeCompanyIsSetToIDs(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// create an application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

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
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyToInsert.ID,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, *applicationToInsert.JobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, *applicationToInsert.Country, *retrievedApplication.Country)
	assert.Equal(t, *applicationToInsert.Area, *retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, *applicationToInsert.WeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, *applicationToInsert.EstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, *applicationToInsert.EstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, retrievedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, retrievedApplication.UpdatedDate)
	assert.NotNil(t, retrievedApplication.Company)

	assert.Equal(t, retrievedApplication.Company.ID, *retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company.Name)
	assert.Nil(t, retrievedApplication.Company.CompanyType)
	assert.Nil(t, retrievedApplication.Company.Notes)
	assert.Nil(t, retrievedApplication.Company.LastContact)
	assert.Nil(t, retrievedApplication.Company.CreatedDate)
	assert.Nil(t, retrievedApplication.Company.UpdatedDate)
}

func TestGetAllApplications_ShouldReturnNoCompanyIncludeCompanyIsSetToIDsAndThereIsNoCompany(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// create an application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            nil,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationService.GetAllApplications(models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, *applicationToInsert.JobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, *applicationToInsert.Country, *retrievedApplication.Country)
	assert.Equal(t, *applicationToInsert.Area, *retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, *applicationToInsert.WeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, *applicationToInsert.EstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, *applicationToInsert.EstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, retrievedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, retrievedApplication.UpdatedDate)
	assert.Nil(t, retrievedApplication.Company)
}

func TestGetAllApplications_ShouldReturnNoCompanyIfIncludeCompanyIsSetToNone(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// create an application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

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
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyToInsert.ID,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Job Title1"),
		JobAdURL:             testutil.ToPtr("Job Ad URL1"),
		Country:              testutil.ToPtr("Some Country1"),
		Area:                 testutil.ToPtr("Some Area1"),
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedApplication1, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	// get all applications

	results, err := applicationService.GetAllApplications(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, *applicationToInsert.JobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, *applicationToInsert.Country, *retrievedApplication.Country)
	assert.Equal(t, *applicationToInsert.Area, *retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, *applicationToInsert.WeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, *applicationToInsert.EstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, *applicationToInsert.EstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, retrievedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, retrievedApplication.UpdatedDate)
	assert.Nil(t, retrievedApplication.Company)
}

// -------- UpdateApplication tests: --------

func TestUpdateApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

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

	retrievedApplication, err := applicationService.GetApplicationById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

	assert.Equal(t, id, retrievedApplication.ID)
	assert.Equal(t, newCompany.ID, *retrievedApplication.CompanyID)
	assert.Equal(t, newRecruiter.ID, *retrievedApplication.RecruiterID)
	assert.Equal(t, applicationToUpdate.JobTitle, retrievedApplication.JobTitle)
	assert.Equal(t, applicationToUpdate.JobAdURL, retrievedApplication.JobAdURL)
	assert.Equal(t, applicationToUpdate.Country, retrievedApplication.Country)
	assert.Equal(t, applicationToUpdate.Area, retrievedApplication.Area)
	assert.Equal(t, applicationToUpdate.RemoteStatusType.String(), retrievedApplication.RemoteStatusType.String())
	assert.Equal(t, applicationToUpdate.WeekdaysInOffice, retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, applicationToUpdate.EstimatedCycleTime, retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, applicationToUpdate.EstimatedCommuteTime, retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToUpdate.ApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, &updatedDateApproximation, retrievedApplication.UpdatedDate)
}

func TestUpdateApplication_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	applicationToUpdate := models.UpdateApplication{
		ID:       uuid.New(),
		JobTitle: testutil.ToPtr("JobTitle"),
	}

	err := applicationService.UpdateApplication(&applicationToUpdate)
	assert.NoError(t, err)
}

// -------- DeleteApplication tests: --------

func TestDeleteApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

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
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
}

func TestDeleteApplication_ShouldReturnNotFoundErrorIfIdToDeleteDoesNotExist(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	id := uuid.New()
	err := applicationService.DeleteApplication(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Application does not exist. ID: "+id.String(), err.Error())
}
