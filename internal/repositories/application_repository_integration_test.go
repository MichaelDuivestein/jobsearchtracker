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
	assert.Equal(t, application.JobTitle, insertedApplication.JobTitle)
	assert.Equal(t, application.JobAdURL, insertedApplication.JobAdURL)
	assert.Equal(t, application.Country, insertedApplication.Country)
	assert.Equal(t, application.Area, insertedApplication.Area)
	assert.Equal(t, application.RemoteStatusType.String(), insertedApplication.RemoteStatusType.String())
	assert.Equal(t, application.WeekdaysInOffice, insertedApplication.WeekdaysInOffice)
	assert.Equal(t, application.EstimatedCycleTime, insertedApplication.EstimatedCycleTime)
	assert.Equal(t, application.EstimatedCommuteTime, insertedApplication.EstimatedCommuteTime)
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

	createdDateApproximation := time.Now()
	insertedApplication, err := applicationRepository.Create(&application)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	assert.NotNil(t, insertedApplication.ID)
	assert.Equal(t, companyID, insertedApplication.CompanyID)
	assert.Nil(t, insertedApplication.RecruiterID)
	assert.Nil(t, insertedApplication.JobTitle)
	assert.Equal(t, application.JobAdURL, insertedApplication.JobAdURL)
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

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CompanyID and RecruiterID cannot both be empty", validationError.Error())
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

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
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

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: Foreign key does not exist", validationError.Error())
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

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: JobTitle and JobAdURL cannot both be empty", validationError.Error())
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
	assert.Equal(t, applicationToInsert.CompanyID, retrievedApplication.CompanyID)
	assert.Equal(t, applicationToInsert.RecruiterID, retrievedApplication.RecruiterID)
	assert.Equal(t, applicationToInsert.JobAdURL, retrievedApplication.JobAdURL)
	assert.Equal(t, applicationToInsert.Country, retrievedApplication.Country)
	assert.Equal(t, applicationToInsert.Area, retrievedApplication.Area)
	assert.Equal(t, applicationToInsert.RemoteStatusType.String(), retrievedApplication.RemoteStatusType.String())
	assert.Equal(t, applicationToInsert.WeekdaysInOffice, retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, applicationToInsert.EstimatedCycleTime, retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, applicationToInsert.EstimatedCommuteTime, retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.ApplicationDate, insertedApplication.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.CreatedDate, insertedApplication.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, applicationToInsert.UpdatedDate, insertedApplication.UpdatedDate)
}

func TestGetById_ShouldReturnErrorIfApplicationIDIsNil(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	response, err := applicationRepository.GetById(nil)
	assert.Nil(t, response)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: ID is nil", validationError.Error())
}

func TestGetById_ShouldReturnErrorIfApplicationIDDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	id := uuid.New()

	response, err := applicationRepository.GetById(&id)
	assert.Nil(t, response)
	assert.NotNil(t, err)

	var notfoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notfoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notfoundError.Error())
}

// -------- GetByJobTitle tests: --------

func TestGetAllByJobTitle_ShouldReturnApplications(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	application1ToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Some Job Title"),
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
	insertedApplication1, err := applicationRepository.Create(&application1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	application2ToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      recruiterID,
		JobTitle:         testutil.ToPtr("Another Job name"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	insertedApplication2, err := applicationRepository.Create(&application2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication2)

	applications, err := applicationRepository.GetAllByJobTitle(testutil.ToPtr("Job"))
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

	assert.Equal(t, *application2ToInsert.ID, applications[1].ID)
}

func TestGetAllByJobTitle_ShouldReturnValidationErrorIfApplicationNameIsNil(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	retrievedApplications, err := applicationRepository.GetAllByJobTitle(nil)
	assert.Nil(t, retrievedApplications)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: JobTitle is nil", validationError.Error())
}

func TestGetAllByJobTitle_ShouldReturnNotFoundErrorIfApplicationNameDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	jobTitle := "Doesnt Exist"

	application, err := applicationRepository.GetAllByJobTitle(&jobTitle)
	assert.Nil(t, application)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: JobTitle: '"+jobTitle+"'", notFoundError.Error())
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

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeNone)
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

	applications, err := applicationRepository.GetAll(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, applications)
}

func TestGetAll_ShouldReturnCompanyIfIncludeCompanyIsSetToAll(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// Create Application

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
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeAll, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.Equal(t, applicationToInsert.RemoteStatusType, *retrievedApplication.RemoteStatusType)
	assert.NotNil(t, retrievedApplication.Company)

	assert.Equal(t, retrievedApplication.Company.ID, *retrievedApplication.CompanyID)
	assert.Equal(t, companyToInsert.Name, *retrievedApplication.Company.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), retrievedApplication.Company.CompanyType.String())
	assert.Equal(t, companyToInsert.Notes, retrievedApplication.Company.Notes)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.LastContact, retrievedApplication.Company.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.CreatedDate, retrievedApplication.Company.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.UpdatedDate, retrievedApplication.Company.UpdatedDate)
}

func TestGetAll_ShouldReturnNoCompanyIfIncludeCompanyIsSetToAllAndThereIsNoCompany(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// Create Application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        nil,
		RecruiterID:      recruiterID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeAll, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company)
}

func TestGetAll_ShouldReturnCompanyWithOnlyIDIfIncludeCompanyIsSetToIDs(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// Create Application

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
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeIDs, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, applicationToInsert.CompanyID, retrievedApplication.CompanyID)
	assert.NotNil(t, retrievedApplication.Company)

	assert.Equal(t, retrievedApplication.Company.ID, *retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company.Name)
	assert.Nil(t, retrievedApplication.Company.CompanyType)
	assert.Nil(t, retrievedApplication.Company.Notes)
	assert.Nil(t, retrievedApplication.Company.LastContact)
	assert.Nil(t, retrievedApplication.Company.CreatedDate)
	assert.Nil(t, retrievedApplication.Company.UpdatedDate)
}

func TestGetAll_ShouldReturnNoCompanyIncludeCompanyIsSetToIDsAndThereIsNoCompany(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// Create Application

	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        nil,
		RecruiterID:      recruiterID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeIDs, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company)
}

func TestGetAll_ShouldReturnNoCompanyIfIncludeCompanyIsSetToNone(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// Create Application

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication1, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyID, retrievedApplication.CompanyID)
	assert.Nil(t, retrievedApplication.Company)
}

func TestGetAll_ShouldReturnRecruiterIfIncludeRecruiterIsSetToAll(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// create application

	recruiterToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&recruiterToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      recruiterToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, recruiterToInsert.ID, retrievedApplication.RecruiterID)
	assert.NotNil(t, retrievedApplication.Recruiter)

	assert.Equal(t, retrievedApplication.Recruiter.ID, *retrievedApplication.RecruiterID)
	assert.Equal(t, recruiterToInsert.Name, *retrievedApplication.Recruiter.Name)
	assert.Equal(t, recruiterToInsert.CompanyType.String(), retrievedApplication.Recruiter.CompanyType.String())
	assert.Equal(t, recruiterToInsert.Notes, retrievedApplication.Recruiter.Notes)
	testutil.AssertEqualFormattedDateTimes(t, recruiterToInsert.LastContact, retrievedApplication.Recruiter.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, recruiterToInsert.CreatedDate, retrievedApplication.Recruiter.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, recruiterToInsert.UpdatedDate, retrievedApplication.Recruiter.UpdatedDate)
}

func TestGetAll_ShouldReturnNoRecruiterIfIncludeRecruiterIsSetToAllAndThereIsNoRecruiter(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// create application

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyID,
		RecruiterID:      nil,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.RecruiterID)
	assert.Nil(t, retrievedApplication.Recruiter)
}

func TestGetAll_ShouldReturnRecruiterWithOnlyIDIfIncludeRecruiterIsSetToIDs(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// create application

	recruiterToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&recruiterToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      recruiterToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, recruiterToInsert.ID, retrievedApplication.RecruiterID)
	assert.NotNil(t, retrievedApplication.Recruiter)

	assert.Equal(t, retrievedApplication.Recruiter.ID, *retrievedApplication.RecruiterID)
	assert.Nil(t, retrievedApplication.Recruiter.Name)
	assert.Nil(t, retrievedApplication.Recruiter.CompanyType)
	assert.Nil(t, retrievedApplication.Recruiter.Notes)
	assert.Nil(t, retrievedApplication.Recruiter.LastContact)
	assert.Nil(t, retrievedApplication.Recruiter.CreatedDate)
	assert.Nil(t, retrievedApplication.Recruiter.UpdatedDate)
}

func TestGetAll_ShouldReturnNoCompanyIncludeRecruiterIsSetToIDsAndThereIsNoRecruiter(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// create application

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyID,
		RecruiterID:      nil,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeIDs, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Nil(t, retrievedApplication.RecruiterID)
	assert.Nil(t, retrievedApplication.Recruiter)
}

func TestGetAll_ShouldReturnNoRecruiterIfIncludeRecruiterIsSetToNone(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// create application

	recruitertoInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&recruitertoInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      recruitertoInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication1, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication1)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, recruitertoInsert.ID, retrievedApplication.RecruiterID)
	assert.Nil(t, retrievedApplication.Recruiter)
}

func TestGetAll_ShouldReturnCompanyAndRecruiter(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	// create application

	companyToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -7)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -6)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -5)),
	}
	_, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)

	recruiterToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "RecruiterName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("CompanyNotes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&recruiterToInsert)
	assert.NoError(t, err)

	applicationToInsert := models.CreateApplication{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        companyToInsert.ID,
		RecruiterID:      recruiterToInsert.ID,
		JobTitle:         testutil.ToPtr("Job Title"),
		RemoteStatusType: models.RemoteStatusTypeUnknown,
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	// get all applications

	results, err := applicationRepository.GetAll(models.IncludeExtraDataTypeAll, models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 1)

	retrievedApplication := results[0]
	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, companyToInsert.ID, retrievedApplication.CompanyID)
	assert.Equal(t, recruiterToInsert.ID, retrievedApplication.RecruiterID)

	assert.NotNil(t, retrievedApplication.Company)
	assert.Equal(t, retrievedApplication.Company.ID, *retrievedApplication.CompanyID)
	assert.Equal(t, companyToInsert.Name, *retrievedApplication.Company.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), retrievedApplication.Company.CompanyType.String())
	assert.Equal(t, companyToInsert.Notes, retrievedApplication.Company.Notes)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.LastContact, retrievedApplication.Company.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.CreatedDate, retrievedApplication.Company.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.UpdatedDate, retrievedApplication.Company.UpdatedDate)

	assert.NotNil(t, retrievedApplication.Recruiter)
	assert.Equal(t, retrievedApplication.Recruiter.ID, *retrievedApplication.RecruiterID)
	assert.Equal(t, recruiterToInsert.Name, *retrievedApplication.Recruiter.Name)
	assert.Equal(t, recruiterToInsert.CompanyType.String(), retrievedApplication.Recruiter.CompanyType.String())
	assert.Equal(t, recruiterToInsert.Notes, retrievedApplication.Recruiter.Notes)
	testutil.AssertEqualFormattedDateTimes(t, recruiterToInsert.LastContact, retrievedApplication.Recruiter.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, recruiterToInsert.CreatedDate, retrievedApplication.Recruiter.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, recruiterToInsert.UpdatedDate, retrievedApplication.Recruiter.UpdatedDate)
}

// -------- Update tests: --------

func TestUpdate_ShouldUpdateApplication(t *testing.T) {
	applicationRepository, companyRepository := setupApplicationRepository(t)

	companyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	recruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create an application

	applicationToInsert := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             testutil.ToPtr("Old Job Title"),
		JobAdURL:             testutil.ToPtr("Old Job Ad URL"),
		Country:              testutil.ToPtr("Old Country"),
		Area:                 testutil.ToPtr("Old Area"),
		RemoteStatusType:     models.RemoteStatusTypeUnknown,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 10)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 20)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 30)),
	}
	insertedApplication, err := applicationRepository.Create(&applicationToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedApplication)

	newCompanyID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	newRecruiterID := &repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// update the application

	var newRemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeOffice
	applicationToUpdate := models.UpdateApplication{
		ID:                   *applicationToInsert.ID,
		CompanyID:            newCompanyID,
		RecruiterID:          newRecruiterID,
		JobTitle:             testutil.ToPtr("New Job Title"),
		JobAdURL:             testutil.ToPtr("New Job Ad URL"),
		Country:              testutil.ToPtr("New Country"),
		Area:                 testutil.ToPtr("New Area"),
		RemoteStatusType:     &newRemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(4),
		EstimatedCycleTime:   testutil.ToPtr(5),
		EstimatedCommuteTime: testutil.ToPtr(6),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 40)),
	}

	updatedDateApproximation := time.Now()
	err = applicationRepository.Update(&applicationToUpdate)
	assert.NoError(t, err)

	// get the company and verify that it's updated

	retrievedApplication, err := applicationRepository.GetById(applicationToInsert.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedApplication)

	assert.Equal(t, *applicationToInsert.ID, retrievedApplication.ID)
	assert.Equal(t, newCompanyID.String(), retrievedApplication.CompanyID.String())
	assert.Equal(t, newRecruiterID.String(), retrievedApplication.RecruiterID.String())
	assert.Equal(t, applicationToUpdate.JobTitle, retrievedApplication.JobTitle)
	assert.Equal(t, applicationToUpdate.JobAdURL, retrievedApplication.JobAdURL)
	assert.Equal(t, applicationToUpdate.Country, retrievedApplication.Country)
	assert.Equal(t, applicationToUpdate.Area, retrievedApplication.Area)
	assert.Equal(t, newRemoteStatusType.String(), retrievedApplication.RemoteStatusType.String())
	assert.Equal(t, applicationToUpdate.WeekdaysInOffice, retrievedApplication.WeekdaysInOffice)
	assert.Equal(t, applicationToUpdate.EstimatedCycleTime, retrievedApplication.EstimatedCycleTime)
	assert.Equal(t, applicationToUpdate.EstimatedCommuteTime, retrievedApplication.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, applicationToUpdate.ApplicationDate, retrievedApplication.ApplicationDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, retrievedApplication.UpdatedDate, time.Second)
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

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: ID is nil", validationError.Error())
}

func TestDelete_ShouldReturnNotFoundErrorIfApplicationIdDoesNotExist(t *testing.T) {
	applicationRepository, _ := setupApplicationRepository(t)

	id := uuid.New()
	err := applicationRepository.Delete(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Application does not exist. ID: "+id.String(), notFoundError.Error())
}
