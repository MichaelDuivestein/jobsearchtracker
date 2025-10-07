package repositories_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupApplicationRepository(t *testing.T) (*repositories.ApplicationRepository, *repositories.CompanyRepository) {
	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupApplicationRepositoryTestContainer(t, *config)

	var applicationRepository *repositories.ApplicationRepository
	err := container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	return applicationRepository, companyRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertAndReturnApplication(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	id := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Some Country"
	area := "Some Area"
	weekdaysInOffice := 1
	estimatedCycleTime := 2
	estimatedCommuteTime := 3
	applicationDate := time.Now().AddDate(0, 0, -1)
	createdDate := time.Now().AddDate(0, 0, -2)
	updatedDate := time.Now().AddDate(0, 0, -3)

	application := models.CreateApplication{
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

	insertedApplication, err := applicationRepository.Create(&application)
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

func TestCreate_ShouldInsertAndReturnWithMinimumRequiredFields(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	jobAdURL := "Job Ad URL"
	application := models.CreateApplication{
		CompanyID:        companyID,
		JobAdURL:         &jobAdURL,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	createdDateApproximation := time.Now().Format(time.RFC3339)
	insertedApplication, err := applicationRepository.Create(&application)
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

func TestCreate_ShouldReturnErrorIfCompanyIDAndRecruiterIDIsNil(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	jobTitle := "JobTitle"
	createApplication := models.CreateApplication{
		CompanyID:        nil,
		RecruiterID:      nil,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	application, err := applicationRepository.Create(&createApplication)
	assert.Nil(t, application)
	assert.Error(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: CompanyID and RecruiterID cannot both be empty", err.Error())
}

func TestCreate_ShouldReturnErrorIfCompanyIDIsNotInCompany(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	companyID := uuid.New()
	jobTitle := "JobTitle"
	createApplication := models.CreateApplication{
		CompanyID:        &companyID,
		RecruiterID:      nil,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	application, err := applicationRepository.Create(&createApplication)
	assert.Nil(t, application)
	assert.Error(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: Foreign key does not exist", err.Error())
}

func TestCreate_ShouldReturnErrorIfRecruiterIDIsNotInCompany(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	recruiterID := uuid.New()
	jobTitle := "JobTitle"
	createApplication := models.CreateApplication{
		CompanyID:        nil,
		RecruiterID:      &recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	application, err := applicationRepository.Create(&createApplication)
	assert.Nil(t, application)
	assert.Error(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: Foreign key does not exist", err.Error())
}

func TestCreate_ShouldReturnErrorIfJobTitleAndJobAdURLIsNil(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	createApplication := models.CreateApplication{
		CompanyID:        companyID,
		JobTitle:         nil,
		JobAdURL:         nil,
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}

	application, err := applicationRepository.Create(&createApplication)
	assert.Nil(t, application)
	assert.Error(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: JobTitle and JobAdURL cannot both be empty", err.Error())

}

// -------- GetById tests: --------

func TestGetById_ShouldGetApplication(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	id := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Country"
	area := "Area"
	weekdaysInOffice := 1
	estimatedCycleTime := 2
	estimatedCommuteTime := 3

	applicationDate := time.Now().AddDate(0, 0, -1)
	createdDate := time.Now().AddDate(0, 0, -2)
	updatedDate := time.Now().AddDate(0, 0, -3)

	applicationToInsert := models.CreateApplication{
		ID:                   &id,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     models.RemoteStatusTypeOffice,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
	}

	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	retrievedApplication, err := applicationRepository.GetById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

	assert.Equal(t, id, retrievedApplication.ID)
	assert.Equal(t, companyID, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, jobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, country, *retrievedApplication.Country)
	assert.Equal(t, area, *retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, weekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, estimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, estimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)

	retrievedApplicationLastContact := retrievedApplication.ApplicationDate.Format(time.RFC3339)
	applicationToInsertLastContact := applicationToInsert.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertLastContact, retrievedApplicationLastContact)

	retrievedApplicationCreatedDate := retrievedApplication.CreatedDate.Format(time.RFC3339)
	applicationToInsertCreatedDate := applicationToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertCreatedDate, retrievedApplicationCreatedDate)

	retrievedApplicationUpdatedDate := retrievedApplication.UpdatedDate.Format(time.RFC3339)
	applicationToInsertUpdatedDate := applicationToInsert.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertUpdatedDate, retrievedApplicationUpdatedDate)
}

func TestGetById_ShouldReturnErrorIfApplicationIDIsNil(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	response, err := applicationRepository.GetById(nil)
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, "validation error: ID is nil", err.Error())
}

func TestGetById_ShouldReturnErrorIfApplicationIDDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	id := uuid.New()

	response, err := applicationRepository.GetById(&id)
	assert.Nil(t, response, "response should be nil")
	assert.NotNil(t, err, err.Error(), "Wrong error returned")
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
}

// -------- GetByJobTitle tests: --------

func TestGetAllByJobTitle_ShouldReturnApplication(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	jobTitle := "Some Job Title"

	applicationToInsert := models.CreateApplication{
		RecruiterID:      recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	retrievedApplications, err := applicationRepository.GetAllByJobTitle(insertedApplication.JobTitle)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplications)
	assert.Equal(t, 1, len(retrievedApplications))

	assert.Equal(t, "Some Job Title", *retrievedApplications[0].JobTitle)

}

func TestGetAllByJobTitle_ShouldReturnValidationErrorIfApplicationNameIsNil(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	retrievedApplications, err := applicationRepository.GetAllByJobTitle(nil)
	assert.Nil(t, retrievedApplications)
	assert.NotNil(t, err)
	assert.Equal(t, "validation error: JobTitle is nil", err.Error())
}

func TestGetAllByJobTitle_ShouldReturnNotFoundErrorIfApplicationNameDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	jobTitle := "Doesnt Exist"

	application, err := applicationRepository.GetAllByJobTitle(&jobTitle)
	assert.Nil(t, application)
	assert.NotNil(t, err)
	assert.Equal(t, "error: object not found: JobTitle: '"+jobTitle+"'", err.Error())
}

func TestGetAllByJobTitle_ShouldReturnMultipleApplicationsWithSameJobTitle(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// insert some applications

	application1ID := uuid.New()
	application1CompanyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1JobTitle := "Developer"
	application1 := models.CreateApplication{
		ID:               &application1ID,
		CompanyID:        application1CompanyID,
		JobTitle:         &application1JobTitle,
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}
	insertedApplication1, err := applicationRepository.Create(&application1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	application2ID := uuid.New()
	application2RecruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application2JobTitle := "Software Engineer"
	application2 := models.CreateApplication{
		ID:               &application2ID,
		RecruiterID:      application2RecruiterID,
		JobTitle:         &application2JobTitle,
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication2, err := applicationRepository.Create(&application2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication2)

	application3ID := uuid.New()
	application3CompanyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application3JobTitle := "Backend Developer"
	application3 := models.CreateApplication{
		ID:               &application3ID,
		CompanyID:        application3CompanyID,
		JobTitle:         &application3JobTitle,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	insertedApplication3, err := applicationRepository.Create(&application3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication3)

	developer := "developer"

	retrievedApplications, err := applicationRepository.GetAllByJobTitle(&developer)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplications)
	assert.Equal(t, 2, len(retrievedApplications))

	foundApplication1 := retrievedApplications[0]
	assert.Equal(t, insertedApplication1.ID.String(), foundApplication1.ID.String())

	foundApplication2 := retrievedApplications[1]
	assert.Equal(t, insertedApplication3.ID.String(), foundApplication2.ID.String())
}

// -------- GetAll tests: --------

func TestGetAll_ShouldReturnAllApplications(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)
	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	id1 := uuid.New()
	jobTitle1 := "Job Title1"
	jobAdURL1 := "Job Ad URL1"
	country1 := "Some Country1"
	area1 := "Some Area1"
	weekdaysInOffice1 := 1
	estimatedCycleTime1 := 2
	estimatedCommuteTime1 := 3
	applicationDate1 := time.Now().AddDate(0, 0, -1)
	createdDate1 := time.Now().AddDate(0, 0, -2)
	updatedDate1 := time.Now().AddDate(0, 0, -3)

	application1ToInsert := models.CreateApplication{
		ID:                   &id1,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle1,
		JobAdURL:             &jobAdURL1,
		Country:              &country1,
		Area:                 &area1,
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     &weekdaysInOffice1,
		EstimatedCycleTime:   &estimatedCycleTime1,
		EstimatedCommuteTime: &estimatedCommuteTime1,
		ApplicationDate:      &applicationDate1,
		CreatedDate:          &createdDate1,
		UpdatedDate:          &updatedDate1,
	}

	insertedApplication1, err := applicationRepository.Create(&application1ToInsert)
	assert.Nil(t, err, "Error on applicationRepository.Create(): '%s'.", err)
	assert.NotNil(t, insertedApplication1, "inserted application 1 is nil")

	id2 := uuid.New()
	jobTitle2 := "Job Title2"
	jobAdURL2 := "Job Ad URL2"
	country2 := "Some Country2"
	area2 := "Some Area2"
	weekdaysInOffice2 := 1
	estimatedCycleTime2 := 2
	estimatedCommuteTime2 := 3
	applicationDate2 := time.Now().AddDate(0, 0, -4)
	createdDate2 := time.Now().AddDate(0, 0, -5)
	updatedDate2 := time.Now().AddDate(0, 0, -6)

	application2ToInsert := models.CreateApplication{
		ID:                   &id2,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle2,
		JobAdURL:             &jobAdURL2,
		Country:              &country2,
		Area:                 &area2,
		RemoteStatusType:     models.RemoteStatusTypeOffice,
		WeekdaysInOffice:     &weekdaysInOffice2,
		EstimatedCycleTime:   &estimatedCycleTime2,
		EstimatedCommuteTime: &estimatedCommuteTime2,
		ApplicationDate:      &applicationDate2,
		CreatedDate:          &createdDate2,
		UpdatedDate:          &updatedDate2,
	}

	insertedApplication2, err := applicationRepository.Create(&application2ToInsert)
	assert.Nil(t, err, "Error on applicationRepository.Create(): '%s'.", err)
	assert.NotNil(t, insertedApplication2, "inserted application 2 is nil")

	results, err := applicationRepository.GetAll()
	assert.Nil(t, err, "Error on applicationRepository.GetAll(): '%s'.", err)

	assert.NotNil(t, results, "results should not be nil")
	assert.Equal(t, 2, len(results), "number of results should be 2")

	assert.Equal(t, id2.String(), results[1].ID.String())
	assert.Equal(t, id1.String(), results[0].ID.String())
}

func TestGetAll_ShouldReturnNilIfNoApplicationsInDatabase(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	applications, err := applicationRepository.GetAll()
	assert.Nil(t, err, "Error on applicationRepository.GetAll(): '%s'.", err)
	assert.Nil(t, applications, "applications should be nil")
}

// -------- Update tests: --------

func TestUpdate_ShouldUpdateApplication(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create an application
	id := uuid.New()
	jobTitle := "Old Job Title"
	jobAdURL := "Old Job Ad URL"
	country := "Old Country"
	area := "Old Area"
	weekdaysInOffice := 1
	estimatedCycleTime := 2
	estimatedCommuteTime := 3
	applicationDate := time.Now().AddDate(0, 0, 10)
	createdDate := time.Now().AddDate(0, 0, 20)
	updatedDate := time.Now().AddDate(0, 0, 30)

	applicationToInsert := models.CreateApplication{
		ID:                   &id,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	newCompanyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	newRecruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	newJobTitle := "New Job Title"
	newJobAdURL := "New Job Ad URL"
	newCountry := "New Country"
	newArea := "New Area"
	var newRemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeOffice
	newWeekdaysInOffice := 1
	newEstimatedCycleTime := 2
	newEstimatedCommuteTime := 3
	newApplicationDate := time.Now().AddDate(0, 0, 40)

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

	// update the application

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	err = applicationRepository.Update(&applicationToUpdate)
	assert.NoError(t, err)

	// get the company and verify that it's updated

	retrievedApplication, err := applicationRepository.GetById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

	assert.Equal(t, id, retrievedApplication.ID)
	assert.Equal(t, newCompanyID.String(), retrievedApplication.CompanyID.String())
	assert.Equal(t, newRecruiterID.String(), retrievedApplication.RecruiterID.String())
	assert.Equal(t, newJobTitle, *retrievedApplication.JobTitle)
	assert.Equal(t, newJobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, newCountry, *retrievedApplication.Country)
	assert.Equal(t, newArea, *retrievedApplication.Area)
	assert.Equal(t, newRemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, newWeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, newEstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, newEstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)

	retrievedApplicationDate := retrievedApplication.ApplicationDate.Format(time.RFC3339)
	expectedApplicationDate := newApplicationDate.Format(time.RFC3339)
	assert.Equal(t, expectedApplicationDate, retrievedApplicationDate)

	retrievedUpdatedDate := retrievedApplication.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, retrievedUpdatedDate)
}

func TestUpdate_ShouldReturnValidationErrorIfNoApplicationFieldsToUpdate(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	id := uuid.New()
	applicationToUpdate := models.UpdateApplication{
		ID: id,
	}

	err := applicationRepository.Update(&applicationToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

func TestUpdate_ShouldNotReturnErrorIfApplicationDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	id := uuid.New()
	jobTitle := "Another Job Title"

	applicationToUpdate := models.UpdateApplication{
		ID:       id,
		JobTitle: &jobTitle,
	}

	err := applicationRepository.Update(&applicationToUpdate)
	assert.NoError(t, err)
}

// -------- Delete tests: --------

func TestDelete_ShouldDeleteApplication(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	id := uuid.New()
	jobTitle := "JobTitle"

	applicationToAdd := models.CreateApplication{
		ID:               &id,
		RecruiterID:      recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	_, err := applicationRepository.Create(&applicationToAdd)
	assert.NoError(t, err)

	err = applicationRepository.Delete(&id)
	assert.NoError(t, err)

	retrievedApplication, err := applicationRepository.GetById(&id)
	assert.Nil(t, retrievedApplication)
	assert.Error(t, err)
}

func TestDelete_ShouldReturnValidationErrorIfApplicationIDIsNil(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	err := applicationRepository.Delete(nil)
	assert.Error(t, err)
	assert.Equal(t, "validation error: ID is nil", err.Error())
}

func TestDelete_ShouldReturnNotFoundErrorIfApplicationIdDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	id := uuid.New()
	err := applicationRepository.Delete(&id)
	assert.Error(t, err)
	assert.Equal(t, "error: object not found: Application does not exist. ID: "+id.String(), err.Error())
}

// -------- Test helpers: --------
