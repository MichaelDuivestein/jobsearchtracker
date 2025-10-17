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

	application := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	insertedApplication, err := applicationRepository.Create(&application)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	assert.Equal(t, *application.ID, insertedApplication.ID)
	assert.Equal(t, application.CompanyID, insertedApplication.CompanyID)
	assert.Equal(t, application.RecruiterID, insertedApplication.RecruiterID)
	assert.Equal(t, *application.JobTitle, *insertedApplication.JobTitle)
	assert.Equal(t, *application.JobAdURL, *insertedApplication.JobAdURL)
	assert.Equal(t, *application.Country, *insertedApplication.Country)
	assert.Equal(t, *application.Area, *insertedApplication.Area)
	assert.Equal(t, application.RemoteStatusType.String(), insertedApplication.RemoteStatusType.String())
	assert.Equal(t, *application.WeekdaysInOffice, *insertedApplication.WeekdaysInOffice)
	assert.Equal(t, *application.EstimatedCycleTime, *insertedApplication.EstimatedCycleTime)
	assert.Equal(t, *application.EstimatedCommuteTime, *insertedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, application.ApplicationDate, insertedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, application.CreatedDate, insertedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, application.UpdatedDate, insertedApplication.UpdatedDate)
}

func TestCreate_ShouldInsertAndReturnWithMinimumRequiredFields(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := models.CreateApplication{
		CompanyID:        companyID,
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
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
	assert.Equal(t, *application.JobAdURL, *insertedApplication.JobAdURL)
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

	createApplication := models.CreateApplication{
		CompanyID:        nil,
		RecruiterID:      nil,
		JobTitle:         testutil.ToPtr("JobTitle"),
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

	createApplication := models.CreateApplication{
		CompanyID:        testutil.ToPtr(uuid.New()),
		RecruiterID:      nil,
		JobTitle:         testutil.ToPtr("JobTitle"),
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

	createApplication := models.CreateApplication{
		CompanyID:        nil,
		RecruiterID:      testutil.ToPtr(uuid.New()),
		JobTitle:         testutil.ToPtr("JobTitle"),
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

	applicationToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Country"),
		Area:                 testutil.ToPtr("Area"),
		RemoteStatusType:     models.RemoteStatusTypeOffice,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	retrievedApplication, err := applicationRepository.GetById(applicationToInsert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyID, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, *applicationToInsert.JobAdURL, *retrievedApplication.JobAdURL)
	assert.Equal(t, *applicationToInsert.Country, *retrievedApplication.Country)
	assert.Equal(t, *applicationToInsert.Area, *retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.Equal(t, *applicationToInsert.WeekdaysInOffice, *retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, *applicationToInsert.EstimatedCycleTime, *retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, *applicationToInsert.EstimatedCommuteTime, *retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, insertedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, insertedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, insertedApplication.UpdatedDate)
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
	assert.Nil(t, response)
	assert.NotNil(t, err, err.Error())
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
	assert.Len(t, retrievedApplications, 1)

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

	application1CompanyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        application1CompanyID,
		JobTitle:         testutil.ToPtr("Developer"),
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}
	insertedApplication1, err := applicationRepository.Create(&application1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	application2RecruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application2 := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      application2RecruiterID,
		JobTitle:         testutil.ToPtr("Software Engineer"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication2, err := applicationRepository.Create(&application2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication2)

	application3CompanyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application3 := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        application3CompanyID,
		JobTitle:         testutil.ToPtr("Backend Developer"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	insertedApplication3, err := applicationRepository.Create(&application3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication3)

	developer := "developer"

	retrievedApplications, err := applicationRepository.GetAllByJobTitle(&developer)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplications)
	assert.Len(t, retrievedApplications, 2)

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

	application1ToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyID,
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
	insertedApplication1, err := applicationRepository.Create(&application1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	application2ToInsert := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		testutil.ToPtr(uuid.New()),
		nil,
		recruiterID,
		testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	)

	results, err := applicationRepository.GetAll()
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 2)

	retrievedApplication1 := results[0]
	assert.Equal(t, *application1ToInsert.ID, retrievedApplication1.ID)
	assert.Equal(t, companyID, retrievedApplication1.CompanyID)
	assert.Equal(t, recruiterID, retrievedApplication1.RecruiterID)
	assert.Equal(t, *application1ToInsert.JobAdURL, *retrievedApplication1.JobAdURL)
	assert.Equal(t, *application1ToInsert.Country, *retrievedApplication1.Country)
	assert.Equal(t, *application1ToInsert.Area, *retrievedApplication1.Area)
	assert.Equal(t, application1ToInsert.RemoteStatusType, *retrievedApplication1.RemoteStatusType)
	assert.Equal(t, *application1ToInsert.WeekdaysInOffice, *retrievedApplication1.WeekdaysInOffice)
	assert.Equal(t, *application1ToInsert.EstimatedCycleTime, *retrievedApplication1.EstimatedCycleTime)
	assert.Equal(t, *application1ToInsert.EstimatedCommuteTime, *retrievedApplication1.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, application1ToInsert.ApplicationDate, retrievedApplication1.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, application1ToInsert.CreatedDate, retrievedApplication1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, application1ToInsert.UpdatedDate, retrievedApplication1.UpdatedDate)

	assert.Equal(t, application2ToInsert.ID, results[1].ID)
}

func TestGetAll_ShouldReturnNilIfNoApplicationsInDatabase(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	applications, err := applicationRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, applications)
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

	updatedDateApproximation := time.Now()
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
	testutil.AssertEqualFormattedDateTimes(t, &newApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, &updatedDateApproximation, retrievedApplication.UpdatedDate)
}

func TestUpdate_ShouldReturnValidationErrorIfNoApplicationFieldsToUpdate(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	applicationToUpdate := models.UpdateApplication{
		ID: uuid.New(),
	}

	err := applicationRepository.Update(&applicationToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

func TestUpdate_ShouldNotReturnErrorIfApplicationDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	applicationToUpdate := models.UpdateApplication{
		ID:       uuid.New(),
		JobTitle: testutil.ToPtr("Another Job Title"),
	}
	err := applicationRepository.Update(&applicationToUpdate)
	assert.NoError(t, err)
}

// -------- Delete tests: --------

func TestDelete_ShouldDeleteApplication(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	id := uuid.New()

	applicationToAdd := models.CreateApplication{
		ID:               &id,
		RecruiterID:      recruiterID,
		JobTitle:         testutil.ToPtr("JobTitle"),
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
