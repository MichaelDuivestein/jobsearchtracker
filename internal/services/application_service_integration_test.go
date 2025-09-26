package services_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/services"
	"jobsearchtracker/internal/testutil/dependencyinjection"
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

	companyID := createCompany(t, companyRepository)
	recruiterID := createCompany(t, companyRepository)

	id := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Some Country"
	area := "Some Area"
	weekdaysInOffice := 9
	estimatedCycleTime := 8
	estimatedCommuteTime := 7
	applicationDate := time.Now().AddDate(0, 0, 4)
	createdDate := time.Now().AddDate(0, 0, 3)
	updatedDate := time.Now().AddDate(0, 0, 2)

	applicationToInsert := models.CreateApplication{
		ID:                   &id,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
	}

	insertedApplication, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	assert.Equal(t, id, insertedApplication.ID)
	assert.Equal(t, companyID, insertedApplication.CompanyID)
	assert.Equal(t, recruiterID, insertedApplication.RecruiterID)
	assert.Equal(t, jobTitle, *insertedApplication.JobTitle)
	assert.Equal(t, jobAdURL, *insertedApplication.JobAdURL)
	assert.Equal(t, country, *insertedApplication.Country)
	assert.Equal(t, area, *insertedApplication.Area)
	assert.Equal(t, models.RemoteStatusTypeHybrid, insertedApplication.RemoteStatusType.String())
	assert.Equal(t, weekdaysInOffice, *insertedApplication.WeekdaysInOffice)
	assert.Equal(t, estimatedCycleTime, *insertedApplication.EstimatedCycleTime)
	assert.Equal(t, estimatedCommuteTime, *insertedApplication.EstimatedCommuteTime)

	applicationToInsertApplicationDate := applicationDate.Format(time.RFC3339)
	insertedApplicationApplicationDate := insertedApplication.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertApplicationDate, insertedApplicationApplicationDate)

	applicationToInsertCreatedDate := createdDate.Format(time.RFC3339)
	insertedApplicationCreatedDate := insertedApplication.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertCreatedDate, insertedApplicationCreatedDate)

	applicationToInsertUpdatedDate := updatedDate.Format(time.RFC3339)
	insertedApplicationUpdatedDate := insertedApplication.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertUpdatedDate, insertedApplicationUpdatedDate)
}

func TestCreateApplication_ShouldHandleEmptyFields(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	companyID := createCompany(t, companyRepository)

	jobAdURL := "Job Ad URL"
	application := models.CreateApplication{
		CompanyID:        companyID,
		JobAdURL:         &jobAdURL,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	createdDateApproximation := time.Now().Format(time.RFC3339)
	insertedApplication, err := applicationService.CreateApplication(&application)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	assert.NotNil(t, insertedApplication.ID)
	assert.Equal(t, companyID, insertedApplication.CompanyID)
	assert.Nil(t, insertedApplication.RecruiterID)
	assert.Nil(t, insertedApplication.JobTitle)
	assert.Equal(t, jobAdURL, *insertedApplication.JobAdURL)
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

	companyID := uuid.New()
	jobAdURL := "Job Ad URL"
	application := models.CreateApplication{
		CompanyID:        &companyID,
		JobAdURL:         &jobAdURL,
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

	recruiterID := uuid.New()
	jobAdURL := "Job Ad URL"
	application := models.CreateApplication{
		RecruiterID:      &recruiterID,
		JobAdURL:         &jobAdURL,
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

	companyID := createCompany(t, companyRepository)
	recruiterID := createCompany(t, companyRepository)

	id := uuid.New()
	jobTitle := "JobTitle"
	jobAdURL := "JobAdURL"
	country := "SomeCountry"
	area := "SomeArea"
	weekdaysInOffice := 9
	estimatedCycleTime := 8
	estimatedCommuteTime := 7
	applicationDate := time.Now().AddDate(0, 0, 4)
	createdDate := time.Now().AddDate(0, 0, 3)
	updatedDate := time.Now().AddDate(0, 0, 2)

	applicationToInsert := models.CreateApplication{
		ID:                   &id,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
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
	companyID := createCompany(t, companyRepository)
	jobTitle := "Some Job Title"

	applicationToInsert := models.CreateApplication{
		CompanyID:        companyID,
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
	assert.Equal(t, 1, len(applications))

	assert.Equal(t, jobTitle, *applications[0].JobTitle)
}

func TestGetApplicationsByJobTitle_ShouldReturnMultipleApplications(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// insert applications

	companyID := createCompany(t, companyRepository)

	id1 := uuid.New()
	jobTitle1 := "developer"
	applicationToInsert1 := models.CreateApplication{
		ID:               &id1,
		CompanyID:        companyID,
		JobTitle:         &jobTitle1,
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}
	_, err := applicationService.CreateApplication(&applicationToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	jobTitle2 := "Backend Developer"
	applicationToInsert2 := models.CreateApplication{
		ID:               &id2,
		CompanyID:        companyID,
		JobTitle:         &jobTitle2,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	_, err = applicationService.CreateApplication(&applicationToInsert2)
	assert.NoError(t, err)

	id3 := uuid.New()
	jobTitle3 := "utvecklare till en företag"
	applicationToInsert3 := models.CreateApplication{
		ID:               &id3,
		CompanyID:        companyID,
		JobTitle:         &jobTitle3,
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}
	_, err = applicationService.CreateApplication(&applicationToInsert3)
	assert.NoError(t, err)

	// GetByJobTitle

	jobTitleToGet := "developer"
	applications, err := applicationService.GetApplicationsByJobTitle(&jobTitleToGet)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, 2, len(applications))

	assert.Equal(t, id2, applications[1].ID)
	assert.Equal(t, id1, applications[0].ID)
}

func TestGetApplicationsByJobTitle_ShouldReturnNotFoundErrorIfNoNamesMatch(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// insert applications

	recruiterID := createCompany(t, companyRepository)

	id := uuid.New()
	jobTitle := "Backend Engineer"
	applicationToInsert := models.CreateApplication{
		ID:               &id,
		RecruiterID:      recruiterID,
		JobTitle:         &jobTitle,
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

	companyID := createCompany(t, companyRepository)

	jobTitle1 := "JobTitle1"
	applicationToInsert1 := models.CreateApplication{
		CompanyID:        companyID,
		JobTitle:         &jobTitle1,
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}
	_, err := applicationService.CreateApplication(&applicationToInsert1)
	assert.NoError(t, err)

	jobTitle2 := "JobTitle2"
	applicationToInsert2 := models.CreateApplication{
		CompanyID:        companyID,
		JobTitle:         &jobTitle2,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	_, err = applicationService.CreateApplication(&applicationToInsert2)
	assert.NoError(t, err)

	// getAll

	applications, err := applicationService.GetAllApplications()
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Equal(t, 2, len(applications))

	assert.Equal(t, jobTitle1, *applications[0].JobTitle)
	assert.Equal(t, jobTitle2, *applications[1].JobTitle)
}

func TestGetAlLApplications_ShouldReturnNilIfNoApplicationsInDatabase(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	applications, err := applicationService.GetAllApplications()
	assert.NoError(t, err)
	assert.Nil(t, applications)
}

// -------- UpdateApplication tests: --------

func TestUpdateApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// insert application

	id := uuid.New()
	originalCompanyID := createCompany(t, companyRepository)
	originalRecruiterID := createCompany(t, companyRepository)
	originalJobTitle := "OriginalJobTitle"
	originalJobAdURL := "OriginalJobAdURL"
	originalCountry := "OriginalCountry"
	originalArea := "OriginalArea"
	originalWeekdaysInOffice := 1
	originalEstimatedCycleTime := 2
	originalEstimatedCommuteTime := 3
	originalApplicationDate := time.Now().AddDate(1, 0, 0)
	originalCreatedDate := time.Now().AddDate(2, 0, 0)
	originalUpdatedDate := time.Now().AddDate(3, 0, 0)

	applicationToInsert := models.CreateApplication{
		ID:                   &id,
		CompanyID:            originalCompanyID,
		RecruiterID:          originalRecruiterID,
		JobTitle:             &originalJobTitle,
		JobAdURL:             &originalJobAdURL,
		Country:              &originalCountry,
		Area:                 &originalArea,
		RemoteStatusType:     models.RemoteStatusTypeOffice,
		WeekdaysInOffice:     &originalWeekdaysInOffice,
		EstimatedCycleTime:   &originalEstimatedCycleTime,
		EstimatedCommuteTime: &originalEstimatedCommuteTime,
		ApplicationDate:      &originalApplicationDate,
		CreatedDate:          &originalCreatedDate,
		UpdatedDate:          &originalUpdatedDate,
	}
	_, err := applicationService.CreateApplication(&applicationToInsert)
	assert.NoError(t, err)

	// update application

	newCompanyID := createCompany(t, companyRepository)
	newRecruiterID := createCompany(t, companyRepository)
	newJobTitle := "NewJobTitle"
	newJobAdURL := "NewJobAdURL"
	newCountry := "NewCountry"
	newArea := "NewArea"
	var newRemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeOffice
	newWeekdaysInOffice := 4
	newEstimatedCycleTime := 5
	newEstimatedCommuteTime := 6
	newApplicationDate := time.Now().AddDate(0, 1, 0)
	applicationToUpdate := models.UpdateApplication{
		ID:                   id,
		CompanyID:            newCompanyID,
		RecruiterID:          newRecruiterID,
		JobTitle:             &newJobTitle,
		JobAdURL:             &newJobAdURL,
		Country:              &newCountry,
		Area:                 &newArea,
		RemoteStatusType:     &newRemoteStatusType,
		WeekdaysInOffice:     &newWeekdaysInOffice,
		EstimatedCycleTime:   &newEstimatedCycleTime,
		EstimatedCommuteTime: &newEstimatedCommuteTime,
		ApplicationDate:      &newApplicationDate,
	}

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	err = applicationService.UpdateApplication(&applicationToUpdate)
	assert.NoError(t, err)

	// get ById

	retrievedApplication, err := applicationService.GetApplicationById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

	assert.Equal(t, id, retrievedApplication.ID)
	assert.Equal(t, newCompanyID, retrievedApplication.CompanyID)
	assert.Equal(t, newRecruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, newJobTitle, *retrievedApplication.JobTitle)
	assert.Equal(t, newJobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, newCountry, *retrievedApplication.Country)
	assert.Equal(t, newArea, *retrievedApplication.Area)
	assert.Equal(t, newRemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, newWeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, newEstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, newEstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)

	applicationDate := retrievedApplication.ApplicationDate.Format(time.RFC3339)
	retrievedApplicationDate := retrievedApplication.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationDate, retrievedApplicationDate)

	updatedDate := retrievedApplication.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, updatedDate)
}

func TestUpdateApplication_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	applicationService, _ := setupApplicationService(t)

	id := uuid.New()
	jobTitle := "JobTitle"
	applicationToUpdate := models.UpdateApplication{
		ID:       id,
		JobTitle: &jobTitle,
	}

	err := applicationService.UpdateApplication(&applicationToUpdate)
	assert.NoError(t, err)
}

// -------- DeleteApplication tests: --------

func TestDeleteApplication_ShouldWork(t *testing.T) {
	applicationService, companyRepository := setupApplicationService(t)

	// insert application

	id := uuid.New()
	recruiterID := createCompany(t, companyRepository)
	jobAdURL := "JobAdURL"
	applicationToInsert := models.CreateApplication{
		ID:               &id,
		RecruiterID:      recruiterID,
		JobAdURL:         &jobAdURL,
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

// -------- Test helpers: --------

func createCompany(t *testing.T, companyRepository *repositories.CompanyRepository) *uuid.UUID {

	id := uuid.New()
	company := models.CreateCompany{
		ID:          &id,
		Name:        "Example Company Name",
		CompanyType: models.CompanyTypeEmployer,
	}

	insertedCompany, err := companyRepository.Create(&company)
	assert.NoError(t, err)

	return &insertedCompany.ID
}
