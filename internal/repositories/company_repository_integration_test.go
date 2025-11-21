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

func setupCompanyRepository(t *testing.T) (
	*repositories.CompanyRepository,
	*repositories.ApplicationRepository,
	*repositories.EventRepository,
	*repositories.PersonRepository,
	*repositories.CompanyEventRepository,
	*repositories.CompanyPersonRepository) {

	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyRepositoryTestContainer(t, *config)

	var companyRepository *repositories.CompanyRepository
	err := container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	var companyEventRepository *repositories.CompanyEventRepository
	err = container.Invoke(func(repository *repositories.CompanyEventRepository) {
		companyEventRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	var companyPersonRepository *repositories.CompanyPersonRepository
	err = container.Invoke(func(repository *repositories.CompanyPersonRepository) {
		companyPersonRepository = repository
	})
	assert.NoError(t, err)

	return companyRepository,
		applicationRepository,
		eventRepository,
		personRepository,
		companyEventRepository,
		companyPersonRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertAndReturnCompany(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	company := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedCompany, err := companyRepository.Create(&company)

	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, *company.ID, insertedCompany.ID)
	assert.Equal(t, company.Name, *insertedCompany.Name)
	assert.Equal(t, company.CompanyType.String(), insertedCompany.CompanyType.String())
	assert.Equal(t, company.Notes, insertedCompany.Notes)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.LastContact, company.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.CreatedDate, company.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.UpdatedDate, company.UpdatedDate)
}

func TestCreate_ShouldInsertCompanyWithMinimumRequiredFields(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	company := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
	}
	createdDateApproximation := time.Now()
	insertedCompany, err := companyRepository.Create(&company)

	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, company.Name, *insertedCompany.Name)
	assert.Equal(t, company.CompanyType.String(), insertedCompany.CompanyType.String())
	assert.Nil(t, insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, insertedCompany.CreatedDate, time.Second)
	assert.Nil(t, insertedCompany.UpdatedDate)
}

func TestCreate_ShouldReturnConflictErrorOnDuplicateCompanyId(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	id := uuid.New()
	firstCompany := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
	}
	firstInsertedCompany, err := companyRepository.Create(&firstCompany)
	assert.NoError(t, err)
	assert.NotNil(t, firstInsertedCompany)
	assert.Equal(t, firstInsertedCompany.ID, id)

	secondCompany := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
	}
	shouldBeNil, err := companyRepository.Create(&secondCompany)
	assert.Nil(t, shouldBeNil)
	assert.Error(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(t,
		"conflict error on insert: ID already exists in database: '"+id.String()+"'",
		conflictError.Error())
}

// -------- GetById tests: --------

func TestGetById_ShouldGetCompany(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	id := uuid.New()
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	retrievedCompany, err := companyRepository.GetById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompany)

	assert.Equal(t, *companyToInsert.ID, retrievedCompany.ID)
	assert.Equal(t, companyToInsert.Name, *retrievedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), retrievedCompany.CompanyType.String())
	assert.Equal(t, companyToInsert.Notes, retrievedCompany.Notes)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.LastContact, retrievedCompany.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.CreatedDate, retrievedCompany.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.UpdatedDate, retrievedCompany.UpdatedDate)
}

func TestGetById_ShouldReturnErrorIfCompanyIDIsNil(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	response, err := companyRepository.GetById(nil)
	assert.Nil(t, response)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

func TestGetById_ShouldReturnErrorIfCompanyIDDoesNotExist(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	id := uuid.New()
	response, err := companyRepository.GetById(&id)
	assert.Nil(t, response)
	assert.NotNil(t, err, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notFoundError.Error())
}

// -------- GetAllByName tests: --------

func TestGetAllByName_ShouldReturnCompany(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	companyToInsert := models.CreateCompany{
		Name:        "Company Bee",
		CompanyType: models.CompanyTypeRecruiter,
	}
	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	retrievedCompanies, err := companyRepository.GetAllByName(insertedCompany.Name)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompanies)
	assert.Len(t, retrievedCompanies, 1)

	assert.Equal(t, "Company Bee", *retrievedCompanies[0].Name)
}

func TestGetAllByName_ShouldReturnValidationErrorIfCompanyNameIsNil(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	retrievedCompanies, err := companyRepository.GetAllByName(nil)
	assert.Nil(t, retrievedCompanies)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: name is nil", validationError.Error())
}

func TestGetAllByName_ShouldReturnNotFoundErrorIfCompanyNameDoesNotExist(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	name := "Doesnt Exist"

	company, err := companyRepository.GetAllByName(&name)
	assert.Nil(t, company)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Name: '"+name+"'", notFoundError.Error())
}

func TestGetAllByName_ShouldReturnMultipleCompaniesWithSameNameSubstring(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	// insert some companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Some Name AB",
		CompanyType: models.CompanyTypeRecruiter,
	}
	insertedCompany1, err := companyRepository.Create(&company1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Brand AB",
		CompanyType: models.CompanyTypeEmployer,
	}
	insertedCompany2, err := companyRepository.Create(&company2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Another Company",
		CompanyType: models.CompanyTypeEmployer,
	}
	insertedCompany3, err := companyRepository.Create(&company3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// get all companies with a name that contains "ab"

	retrievedCompanies, err := companyRepository.GetAllByName(testutil.ToPtr("ab"))
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompanies)
	assert.Len(t, retrievedCompanies, 2)

	foundCompany1 := retrievedCompanies[0]
	assert.Equal(t, insertedCompany2.ID, foundCompany1.ID)

	foundCompany2 := retrievedCompanies[1]
	assert.Equal(t, insertedCompany1.ID, foundCompany2.ID)
}

// -------- GetAll - Base tests: --------

func TestGetAll_ShouldReturnAllCompanies(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	company1ToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedCompany1, err := companyRepository.Create(&company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2ToInsert := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "company2Name",
		CompanyType: models.CompanyTypeConsultancy,
	}
	insertedCompany2, err := companyRepository.Create(&company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	results, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)
	assert.Equal(t, *company2ToInsert.ID, results[0].ID)

	assert.Equal(t, *company1ToInsert.ID, results[1].ID)
	assert.Equal(t, company1ToInsert.Name, *results[1].Name)
	assert.Equal(t, company1ToInsert.CompanyType.String(), results[1].CompanyType.String())
	assert.Equal(t, company1ToInsert.Notes, results[1].Notes, results)
	testutil.AssertEqualFormattedDateTimes(t, company1ToInsert.LastContact, results[1].LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company1ToInsert.CreatedDate, results[1].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company1ToInsert.UpdatedDate, results[1].UpdatedDate)
}

func TestGetAll_ShouldReturnNilIfNoCompaniesInDatabase(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, companies)
}

// -------- GetAll - Applications tests: --------

func TestGetAll_ShouldReturnApplicationsIfIncludeApplicationsIsSetToAll(t *testing.T) {
	companyRepository, applicationRepository, _, _, _, _ := setupCompanyRepository(t)

	// create companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	company3ID := uuid.New()
	createCompany3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	_, err = companyRepository.Create(&createCompany3)
	assert.NoError(t, err)

	// create applications

	application1ID := uuid.New()
	var application1RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	createApplication1 := models.CreateApplication{
		ID:                   &application1ID,
		CompanyID:            &company1ID,
		RecruiterID:          &company2ID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     application1RemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(10),
		EstimatedCycleTime:   testutil.ToPtr(11),
		EstimatedCommuteTime: testutil.ToPtr(12),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application2ID,
		&company2ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	)

	application3ID := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application3ID,
		nil,
		&company2ID,
		nil,
	)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 3)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Len(t, *companies[0].Applications, 3)

	assert.Equal(t, company3ID, companies[1].ID)
	assert.Nil(t, companies[1].Applications)

	assert.Equal(t, company1ID, companies[2].ID)
	assert.Len(t, *companies[2].Applications, 1)

	company2Application1 := (*companies[0].Applications)[0]
	assert.Equal(t, application2ID, company2Application1.ID)
	assert.Equal(t, company2ID, *company2Application1.CompanyID)
	assert.Nil(t, company2Application1.RecruiterID)

	company2Application2 := (*companies[0].Applications)[1]
	assert.Equal(t, application1ID, company2Application2.ID)
	assert.Equal(t, company1ID, *company2Application2.CompanyID)
	assert.Equal(t, company2ID, *company2Application2.RecruiterID)
	assert.Equal(t, "Application1JobTitle", *company2Application2.JobTitle)
	assert.Equal(t, "Application1JobAdURL", *company2Application2.JobAdURL)
	assert.Equal(t, "Application1Country", *company2Application2.Country)
	assert.Equal(t, "Application1Area", *company2Application2.Area)
	assert.Equal(t, application1RemoteStatusType.String(), company2Application2.RemoteStatusType.String())
	assert.Equal(t, 10, *company2Application2.WeekdaysInOffice)
	assert.Equal(t, 11, *company2Application2.EstimatedCycleTime)
	assert.Equal(t, 12, *company2Application2.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.ApplicationDate, company2Application2.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.CreatedDate, company2Application2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.UpdatedDate, company2Application2.UpdatedDate)

	company2Application3 := (*companies[0].Applications)[2]
	assert.Equal(t, application3ID, company2Application3.ID)
	assert.Nil(t, company2Application3.CompanyID)
	assert.Equal(t, company2ID, *company2Application3.RecruiterID)

	company1Application1 := (*companies[2].Applications)[0]
	assert.Equal(t, application1ID, company1Application1.ID)
	assert.Equal(t, company1ID, *company1Application1.CompanyID)
	assert.Equal(t, company2ID, *company1Application1.RecruiterID)
}

func TestGetAll_ShouldReturnNilApplicationsIfIncludeApplicationsIsSetToAllAndThereAreNoApplications(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company1ID, companies[0].ID)
	assert.Nil(t, companies[0].Applications)

	assert.Equal(t, company2ID, companies[1].ID)
	assert.Nil(t, companies[1].Applications)
}

func TestGetAll_ShouldReturnApplicationIDsIfIncludeApplicationsIsSetToIDs(t *testing.T) {
	companyRepository, applicationRepository, _, _, _, _ := setupCompanyRepository(t)

	// create companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	company3ID := uuid.New()
	createCompany3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	_, err = companyRepository.Create(&createCompany3)
	assert.NoError(t, err)

	// create applications

	application1ID := uuid.New()
	var application1RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	createApplication1 := models.CreateApplication{
		ID:                   &application1ID,
		CompanyID:            &company1ID,
		RecruiterID:          &company2ID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     application1RemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(10),
		EstimatedCycleTime:   testutil.ToPtr(11),
		EstimatedCommuteTime: testutil.ToPtr(12),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application2ID,
		&company2ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	)

	application3ID := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application3ID,
		nil,
		&company2ID,
		nil,
	)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 3)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Len(t, *companies[0].Applications, 3)

	assert.Equal(t, company3ID, companies[1].ID)
	assert.Nil(t, companies[1].Applications)

	assert.Equal(t, company1ID, companies[2].ID)
	assert.Len(t, *companies[2].Applications, 1)

	company2Application1 := (*companies[0].Applications)[0]
	assert.Equal(t, application2ID, company2Application1.ID)
	assert.Equal(t, company2ID, *company2Application1.CompanyID)
	assert.Nil(t, company2Application1.RecruiterID)

	company2Application2 := (*companies[0].Applications)[1]
	assert.Equal(t, application1ID, company2Application2.ID)
	assert.Equal(t, company1ID, *company2Application2.CompanyID)
	assert.Equal(t, company2ID, *company2Application2.RecruiterID)
	assert.Nil(t, company2Application2.JobTitle)
	assert.Nil(t, company2Application2.JobAdURL)
	assert.Nil(t, company2Application2.Country)
	assert.Nil(t, company2Application2.Area)
	assert.Nil(t, company2Application2.RemoteStatusType)
	assert.Nil(t, company2Application2.WeekdaysInOffice)
	assert.Nil(t, company2Application2.EstimatedCycleTime)
	assert.Nil(t, company2Application2.EstimatedCommuteTime)
	assert.Nil(t, company2Application2.ApplicationDate)
	assert.Nil(t, company2Application2.CreatedDate)
	assert.Nil(t, company2Application2.UpdatedDate)

	company2Application3 := (*companies[0].Applications)[2]
	assert.Equal(t, application3ID, company2Application3.ID)
	assert.Nil(t, company2Application3.CompanyID)
	assert.Equal(t, company2ID, *company2Application3.RecruiterID)

	company1Application1 := (*companies[2].Applications)[0]
	assert.Equal(t, application1ID, company1Application1.ID)
	assert.Equal(t, company1ID, *company1Application1.CompanyID)
	assert.Equal(t, company2ID, *company1Application1.RecruiterID)
}

func TestGetAll_ShouldReturnNilApplicationsIfIncludeApplicationsIsSetToIDsAndThereAreNoApplications(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Nil(t, companies[0].Applications)

	assert.Equal(t, company1ID, companies[1].ID)
	assert.Nil(t, companies[1].Applications)
}

func TestGetAll_ShouldReturnNilApplicationsIfIncludeApplicationsIsSetToNone(t *testing.T) {
	companyRepository, applicationRepository, _, _, _, _ := setupCompanyRepository(t)

	// create company

	companyID := uuid.New()
	createCompany := models.CreateCompany{
		ID:          &companyID,
		Name:        "companyName",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany)
	assert.NoError(t, err)

	// create application

	applicationID := uuid.New()
	var applicationRemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	createApplication1 := models.CreateApplication{
		ID:               &applicationID,
		CompanyID:        &companyID,
		JobTitle:         testutil.ToPtr("ApplicationJobTitle"),
		RemoteStatusType: applicationRemoteStatusType,
	}
	_, err = applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companies)

	assert.Equal(t, companyID, companies[0].ID)
	assert.Nil(t, companies[0].Applications)
}

// -------- GetAll - Events tests: --------

func TestCompanyRepositoryGetAll_ShouldReturnEventsIfIncludeEventsIsSetToAll(t *testing.T) {
	companyRepository, _, eventRepository, _, companyEventRepository, _ := setupCompanyRepository(t)

	// create companies

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, -4))).ID

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, -2))).ID

	company3ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, -3))).ID

	// create events

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   time.Now().AddDate(0, 0, 7),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 8)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 9)),
	}
	_, err := eventRepository.Create(&createEvent1)
	assert.NoError(t, err)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 8))).ID

	// associate companies with events
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, *createEvent1.ID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, event2ID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, event2ID, nil)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 3)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Len(t, *companies[0].Events, 1)

	company2Event := (*companies[0].Events)[0]
	assert.Equal(t, event2ID, company2Event.ID)

	assert.Equal(t, company3ID, companies[1].ID)
	assert.Nil(t, companies[1].Events)

	assert.Equal(t, company1ID, companies[2].ID)
	assert.Len(t, *companies[2].Events, 2)

	company1Event1 := (*companies[2].Events)[0]
	assert.Equal(t, event2ID, company1Event1.ID)

	company1Event2 := (*companies[2].Events)[1]
	assert.Equal(t, *createEvent1.ID, company1Event2.ID)
	assert.Equal(t, createEvent1.EventType.String(), company1Event2.EventType.String())
	assert.Equal(t, createEvent1.Description, company1Event2.Description)
	assert.Equal(t, createEvent1.Notes, company1Event2.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent1.EventDate, company1Event2.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent1.CreatedDate, company1Event2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent1.UpdatedDate, company1Event2.UpdatedDate)
}

func TestGetAll_ShouldReturnNilEventsIfIncludeEventsIsSetToAllAndThereAreNoEventsInDB(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 1)

	assert.Equal(t, companyID, companies[0].ID)
	assert.Nil(t, companies[0].Events)
}

func TestGetAll_ShouldReturnEventIDsIfIncludeEventsIsSetToIDs(t *testing.T) {
	companyRepository, _, eventRepository, _, companyEventRepository, _ := setupCompanyRepository(t)

	// create companies

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, -4))).ID

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, -2))).ID

	company3ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, -3))).ID
	// create events

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   time.Now().AddDate(0, 0, 7),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 8)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 9)),
	}
	_, err := eventRepository.Create(&createEvent1)
	assert.NoError(t, err)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 8))).ID

	// create companyEvents

	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, *createEvent1.ID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, event2ID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, event2ID, nil)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 3)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Len(t, *companies[0].Events, 1)

	company2Event := (*companies[0].Events)[0]
	assert.Equal(t, event2ID, company2Event.ID)

	assert.Equal(t, company3ID, companies[1].ID)
	assert.Nil(t, companies[1].Events)

	assert.Equal(t, company1ID, companies[2].ID)
	assert.Len(t, *companies[2].Events, 2)

	company1Event1 := (*companies[2].Events)[0]
	assert.Equal(t, event2ID, company1Event1.ID)

	company1Event2 := (*companies[2].Events)[1]
	assert.Equal(t, *createEvent1.ID, company1Event2.ID)
	assert.Nil(t, company1Event2.EventType)
	assert.Nil(t, company1Event2.Description)
	assert.Nil(t, company1Event2.Notes)
	assert.Nil(t, company1Event2.EventDate)
	assert.Nil(t, company1Event2.CreatedDate)
	assert.Nil(t, company1Event2.UpdatedDate)
}

func TestGetAll_ShouldReturnNilEventsIfIncludeEventsIsSetToIDsAndThereAreNoEventsInDB(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 1)

	assert.Equal(t, companyID, companies[0].ID)
	assert.Nil(t, companies[0].Events)
}

func TestGetAll_ShouldReturnNilEventsIfIncludeEventsIsSetToIDsAndThereAreNoCompanyEventsInDB(t *testing.T) {
	companyRepository, _, eventRepository, _, _, _ := setupCompanyRepository(t)

	// create companies

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create events

	repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	// get all events

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 1)

	assert.Equal(t, companyID, companies[0].ID)
	assert.Nil(t, companies[0].Events)
}

func TestGetAll_ShouldReturnNilEventsIfIncludeEventsIsSetToNone(t *testing.T) {
	companyRepository, _, eventRepository, _, companyEventRepository, _ := setupCompanyRepository(t)

	// create company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// create companyEvent

	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 1)

	assert.Equal(t, companyID, companies[0].ID)
	assert.Nil(t, companies[0].Events)
}

// -------- GetAll - Persons tests: --------

func TestCompanyRepositoryGetAll_ShouldReturnPersonsIfIncludePersonsIsSetToAll(t *testing.T) {
	companyRepository, _, _, personRepository, _, companyPersonRepository := setupCompanyRepository(t)

	// create companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	company3ID := uuid.New()
	createCompany3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	_, err = companyRepository.Create(&createCompany3)
	assert.NoError(t, err)

	// create persons

	person1ID := uuid.New()
	var person1Type models.PersonType = models.PersonTypeJobContact
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1Name",
		PersonType:  person1Type,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	repositoryhelpers.CreatePerson(
		t,
		personRepository,
		&person2ID,
		nil,
	)

	// create companyPersons

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1ID,
		PersonID:    person1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1ID,
		PersonID:    person2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2ID,
		PersonID:    person2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 3)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Len(t, *companies[0].Persons, 1)

	company2Person := (*companies[0].Persons)[0]
	assert.Equal(t, person2ID, company2Person.ID)

	assert.Equal(t, company3ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)

	assert.Equal(t, company1ID, companies[2].ID)
	assert.Len(t, *companies[2].Persons, 2)

	company1Person1 := (*companies[2].Persons)[0]
	assert.Equal(t, person2ID, company1Person1.ID)

	company1Person2 := (*companies[2].Persons)[1]
	assert.Equal(t, person1ID, company1Person2.ID)
	assert.Equal(t, person1.Name, *company1Person2.Name)
	assert.Equal(t, person1.PersonType.String(), company1Person2.PersonType.String())
	assert.Equal(t, person1.Email, company1Person2.Email)
	assert.Equal(t, person1.Phone, company1Person2.Phone)
	assert.Equal(t, person1.Notes, company1Person2.Notes)
	testutil.AssertEqualFormattedDateTimes(t, person1.CreatedDate, company1Person2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, person1.UpdatedDate, company1Person2.UpdatedDate)
}

func TestGetAll_ShouldReturnNilPersonsIfIncludePersonsIsSetToAllAndThereAreNoPersonsInDB(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeAll,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company1ID, companies[0].ID)
	assert.Nil(t, companies[0].Persons)

	assert.Equal(t, company2ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)
}

func TestGetAll_ShouldReturnPersonIDsIfIncludePersonsIsSetToIDs(t *testing.T) {
	companyRepository, _, _, personRepository, _, companyPersonRepository := setupCompanyRepository(t)

	// create companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	company3ID := uuid.New()
	createCompany3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	_, err = companyRepository.Create(&createCompany3)
	assert.NoError(t, err)

	// create persons

	person1ID := uuid.New()
	var person1Type models.PersonType = models.PersonTypeJobContact
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1Name",
		PersonType:  person1Type,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	repositoryhelpers.CreatePerson(
		t,
		personRepository,
		&person2ID,
		nil,
	)

	// create companyPersons

	companyPerson1 := models.AssociateCompanyPerson{
		CompanyID:   company1ID,
		PersonID:    person1ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson1)
	assert.NoError(t, err)

	companyPerson2 := models.AssociateCompanyPerson{
		CompanyID:   company1ID,
		PersonID:    person2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson2)
	assert.NoError(t, err)

	companyPerson3 := models.AssociateCompanyPerson{
		CompanyID:   company2ID,
		PersonID:    person2ID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson3)
	assert.NoError(t, err)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 3)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Len(t, *companies[0].Persons, 1)

	company2Person := (*companies[0].Persons)[0]
	assert.Equal(t, person2ID, company2Person.ID)

	assert.Equal(t, company3ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)

	assert.Equal(t, company1ID, companies[2].ID)
	assert.Len(t, *companies[2].Persons, 2)

	company1Person1 := (*companies[2].Persons)[0]
	assert.Equal(t, person2ID, company1Person1.ID)

	company1Person2 := (*companies[2].Persons)[1]
	assert.Equal(t, person1ID, company1Person2.ID)
	assert.Nil(t, company1Person2.Name)
	assert.Nil(t, company1Person2.PersonType)
	assert.Nil(t, company1Person2.Email)
	assert.Nil(t, company1Person2.Phone)
	assert.Nil(t, company1Person2.Notes)
	assert.Nil(t, company1Person2.CreatedDate)
	assert.Nil(t, company1Person2.UpdatedDate)
}

func TestGetAll_ShouldReturnNilPersonsIfIncludePersonsIsSetToIDsAndThereAreNoPersonsInDB(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Nil(t, companies[0].Persons)

	assert.Equal(t, company1ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)
}

func TestGetAll_ShouldReturnNilPersonsIfIncludePersonsIsSetToIDsAndThereAreNoCompanyPersonsInDB(t *testing.T) {
	companyRepository, _, _, personRepository, _, _ := setupCompanyRepository(t)

	// create companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyRepository.Create(&createCompany2)
	assert.NoError(t, err)

	// create persons

	person1ID := uuid.New()
	var person1Type models.PersonType = models.PersonTypeJobContact
	person1 := models.CreatePerson{
		ID:          &person1ID,
		Name:        "Person1Name",
		PersonType:  person1Type,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = personRepository.Create(&person1)
	assert.NoError(t, err)

	person2ID := uuid.New()
	repositoryhelpers.CreatePerson(
		t,
		personRepository,
		&person2ID,
		nil)

	// get all persons

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Nil(t, companies[0].Persons)

	assert.Equal(t, company1ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)
}

func TestGetAll_ShouldReturnNilPersonsIfIncludePersonsIsSetToNone(t *testing.T) {
	companyRepository, _, _, personRepository, _, companyPersonRepository := setupCompanyRepository(t)

	// create company

	companyID := uuid.New()
	createCompany := models.CreateCompany{
		ID:          &companyID,
		Name:        "companyName",
		CompanyType: models.CompanyTypeConsultancy,
	}
	_, err := companyRepository.Create(&createCompany)
	assert.NoError(t, err)

	// create person

	personID := uuid.New()
	person := models.CreatePerson{
		ID:         &personID,
		Name:       "Person1Name",
		PersonType: models.PersonTypeJobContact,
	}
	_, err = personRepository.Create(&person)
	assert.NoError(t, err)

	// create companyPerson

	companyPerson := models.AssociateCompanyPerson{
		CompanyID:   companyID,
		PersonID:    personID,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&companyPerson)
	assert.NoError(t, err)

	// get companies

	companies, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 1)

	assert.Equal(t, companyID, companies[0].ID)
	assert.Nil(t, companies[0].Persons)
}

// -------- GetAll - combined objects tests: --------

func TestCompanyRepositoryGetAll_ShouldReturnTwoCompaniesEvenIfOneApplicationIsSharedBetweenTwoCompanies(t *testing.T) {
	companyRepository, applicationRepository, _, _, _, _ := setupCompanyRepository(t)

	// create two companies

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create an application using the companies as companyID and recruiterID

	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company1ID,
		&company2ID,
		nil)

	// ensure that two companies are returned

	companiesWithApplications, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companiesWithApplications)
	assert.Len(t, companiesWithApplications, 2)

	assert.Equal(t, company2ID, companiesWithApplications[0].ID)
	assert.Len(t, *companiesWithApplications[0].Applications, 1)

	assert.Equal(t, company1ID, companiesWithApplications[1].ID)
	assert.Len(t, *companiesWithApplications[1].Applications, 1)
}

func TestCompanyRepositoryGetAll_ShouldReturnTwoCompaniesEvenIfOneEventIsSharedBetweenTwoCompanies(t *testing.T) {
	companyRepository, _, eventRepository, _, companyEventRepository, _ := setupCompanyRepository(t)

	// create two companies

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create an event and associate it with both companies

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company1ID, eventID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, company2ID, eventID, nil)

	// ensure that two results are returned, each with an event

	companiesWithEvents, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, companiesWithEvents)
	assert.Len(t, companiesWithEvents, 2)

	assert.Equal(t, company2ID, companiesWithEvents[0].ID)
	assert.Len(t, *companiesWithEvents[0].Events, 1)

	assert.Equal(t, company1ID, companiesWithEvents[1].ID)
	assert.Len(t, *companiesWithEvents[1].Events, 1)
}

func TestCompanyRepositoryGetAll_ShouldReturnTwoCompaniesEvenIfOnePersonIsSharedBetweenTwoCompanies(t *testing.T) {
	companyRepository, _, _, personRepository, _, companyPersonRepository := setupCompanyRepository(t)

	// create two companies

	company1ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create a person and associate it with both companies

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, company1ID, personID, nil)
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, company2ID, personID, nil)

	// ensure that two results are returned, each with a person

	companiesWithEvents, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companiesWithEvents)
	assert.Len(t, companiesWithEvents, 2)

	assert.Equal(t, company2ID, companiesWithEvents[0].ID)
	assert.Len(t, *companiesWithEvents[0].Persons, 1)

	assert.Equal(t, company1ID, companiesWithEvents[1].ID)
	assert.Len(t, *companiesWithEvents[1].Persons, 1)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithOneApplicationAndTwoEvents(t *testing.T) {
	companyRepository,
		applicationRepository,
		eventRepository,
		_,
		companyEventRepository,
		_ := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create an application

	applicationID := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil).ID

	// Create two events and associate them with the company

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event1ID, nil)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event2ID, nil)

	// ensure that the company is returned with one application

	companyWithApplications, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplications)
	assert.Len(t, companyWithApplications, 1)
	assert.Len(t, *companyWithApplications[0].Applications, 1)

	// ensure that two events are returned with the company

	companyWithEvents, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithEvents)
	assert.Len(t, companyWithEvents, 1)
	assert.Len(t, *companyWithEvents[0].Events, 2)
	assert.Equal(t, event2ID, (*companyWithEvents[0].Events)[0].ID)
	assert.Equal(t, event1ID, (*companyWithEvents[0].Events)[1].ID)

	// Ensure that the company is returned with one application and two events

	companyWithApplicationsAndEvent, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplicationsAndEvent)
	assert.Len(t, companyWithApplicationsAndEvent, 1)
	assert.Len(t, *companyWithApplicationsAndEvent[0].Applications, 1)
	assert.Len(t, *companyWithApplicationsAndEvent[0].Events, 2)

	assert.Equal(t, applicationID, (*companyWithApplicationsAndEvent[0].Applications)[0].ID)

	assert.Equal(t, event2ID, (*companyWithApplicationsAndEvent[0].Events)[0].ID)
	assert.Equal(t, event1ID, (*companyWithApplicationsAndEvent[0].Events)[1].ID)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithTwoApplicationsAndOneEvent(t *testing.T) {
	companyRepository,
		applicationRepository,
		eventRepository,
		_,
		companyEventRepository,
		_ := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create two applications

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create an event and associate it with the company

	eventID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// Ensure that the company is returned with two applications and an event

	companyWithApplicationsAndEvent, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplicationsAndEvent)
	assert.Len(t, companyWithApplicationsAndEvent, 1)

	assert.Len(t, *companyWithApplicationsAndEvent[0].Applications, 2)
	assert.Equal(t, application2ID, (*companyWithApplicationsAndEvent[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*companyWithApplicationsAndEvent[0].Applications)[1].ID)

	assert.Len(t, *companyWithApplicationsAndEvent[0].Events, 1)
	assert.Equal(t, eventID, (*companyWithApplicationsAndEvent[0].Events)[0].ID)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithTwoApplicationsAndTwoEvents(t *testing.T) {
	companyRepository,
		applicationRepository,
		eventRepository,
		_,
		companyEventRepository,
		_ := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create two applications

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create two events and associate them with the company

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event1ID, nil)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event2ID, nil)

	// Ensure that the company is returned with two applications and two events

	companyWithApplicationsAndEvent, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplicationsAndEvent)
	assert.Len(t, companyWithApplicationsAndEvent, 1)

	assert.Len(t, *companyWithApplicationsAndEvent[0].Applications, 2)
	assert.Equal(t, application2ID, (*companyWithApplicationsAndEvent[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*companyWithApplicationsAndEvent[0].Applications)[1].ID)

	assert.Len(t, *companyWithApplicationsAndEvent[0].Events, 2)
	assert.Equal(t, event2ID, (*companyWithApplicationsAndEvent[0].Events)[0].ID)
	assert.Equal(t, event1ID, (*companyWithApplicationsAndEvent[0].Events)[1].ID)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithOneApplicationAndTwoPersons(t *testing.T) {
	companyRepository,
		applicationRepository,
		_,
		personRepository,
		_,
		companyPersonRepository := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// Create two persons and associate them with the company

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person2ID, nil)

	// ensure that the company is returned with one application

	companyWithApplications, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplications)
	assert.Len(t, companyWithApplications, 1)
	assert.Len(t, *companyWithApplications[0].Applications, 1)

	// ensure that two persons are returned with the company

	companyWithPersons, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithPersons)
	assert.Len(t, companyWithPersons, 1)
	assert.Len(t, *companyWithPersons[0].Persons, 2)
	assert.Equal(t, person2ID, (*companyWithPersons[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*companyWithPersons[0].Persons)[1].ID)

	// Ensure that the company is returned with one application and two persons

	companyWithApplicationsAndPerson, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplicationsAndPerson)
	assert.Len(t, companyWithApplicationsAndPerson, 1)
	assert.Len(t, *companyWithApplicationsAndPerson[0].Applications, 1)
	assert.Len(t, *companyWithApplicationsAndPerson[0].Persons, 2)

	assert.Equal(t, person2ID, (*companyWithApplicationsAndPerson[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*companyWithApplicationsAndPerson[0].Persons)[1].ID)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithTwoApplicationsAndOnePerson(t *testing.T) {
	companyRepository,
		applicationRepository,
		_,
		personRepository,
		_,
		companyPersonRepository := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create two applications

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// Create a person and associate it with the company

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, personID, nil)

	// ensure that the company is returned with two applications

	companyWithApplications, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplications)
	assert.Len(t, companyWithApplications, 1)
	assert.Len(t, *companyWithApplications[0].Applications, 2)
	assert.Equal(t, application2ID, (*companyWithApplications[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*companyWithApplications[0].Applications)[1].ID)

	// ensure that a single person is returned with the company

	companyWithPersons, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithPersons)
	assert.Len(t, companyWithPersons, 1)
	assert.Len(t, *companyWithPersons[0].Persons, 1)

	// Ensure that the company is returned with two applications and a single person

	companyWithApplicationsAndPerson, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplicationsAndPerson)
	assert.Len(t, companyWithApplicationsAndPerson, 1)
	assert.Len(t, *companyWithApplicationsAndPerson[0].Applications, 2)
	assert.Len(t, *companyWithApplicationsAndPerson[0].Persons, 1)

	assert.Equal(t, application2ID, (*companyWithApplicationsAndPerson[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*companyWithApplicationsAndPerson[0].Applications)[1].ID)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithTwoApplicationsAndTwoPersons(t *testing.T) {
	companyRepository,
		applicationRepository,
		_,
		personRepository,
		_,
		companyPersonRepository := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create two applications

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create two persons and associate them with the company

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person2ID, nil)

	// Ensure that the company is returned with two applications and two persons

	companyWithApplicationsAndPerson, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithApplicationsAndPerson)
	assert.Len(t, companyWithApplicationsAndPerson, 1)

	assert.Len(t, *companyWithApplicationsAndPerson[0].Applications, 2)
	assert.Equal(t, application2ID, (*companyWithApplicationsAndPerson[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*companyWithApplicationsAndPerson[0].Applications)[1].ID)

	assert.Len(t, *companyWithApplicationsAndPerson[0].Persons, 2)
	assert.Equal(t, person2ID, (*companyWithApplicationsAndPerson[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*companyWithApplicationsAndPerson[0].Persons)[1].ID)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithTwoEventsAndTwoPersons(t *testing.T) {
	companyRepository,
		_,
		eventRepository,
		personRepository,
		companyEventRepository,
		companyPersonRepository := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create two events and associate them with the company

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event1ID, nil)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event2ID, nil)

	// create two persons and associate them with the company

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person2ID, nil)

	// Ensure that the company is returned with two events and two persons

	companyWithEventsAndPerson, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeNone,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, companyWithEventsAndPerson)
	assert.Len(t, companyWithEventsAndPerson, 1)

	assert.Len(t, *companyWithEventsAndPerson[0].Events, 2)
	assert.Equal(t, event2ID, (*companyWithEventsAndPerson[0].Events)[0].ID)
	assert.Equal(t, event1ID, (*companyWithEventsAndPerson[0].Events)[1].ID)

	assert.Len(t, *companyWithEventsAndPerson[0].Persons, 2)
	assert.Equal(t, person2ID, (*companyWithEventsAndPerson[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*companyWithEventsAndPerson[0].Persons)[1].ID)
}

func TestCompanyRepositoryGetAll_ShouldReturnCompanyWithTwoApplicationsAndTwoEventsAndTwoPersons(t *testing.T) {
	companyRepository,
		applicationRepository,
		eventRepository,
		personRepository,
		companyEventRepository,
		companyPersonRepository := setupCompanyRepository(t)

	// create a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// create two applications

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID
	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	// create two events and associate them with the company

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event1ID, nil)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event2ID, nil)

	// create two persons and associate them with the company

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person1ID, nil)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person2ID, nil)

	// Ensure that the company is returned with two applications and two events and two persons

	company, err := companyRepository.GetAll(
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs,
		models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)
	assert.NotNil(t, company)
	assert.Len(t, company, 1)

	assert.Len(t, *company[0].Applications, 2)
	assert.Equal(t, application2ID, (*company[0].Applications)[0].ID)
	assert.Equal(t, application1ID, (*company[0].Applications)[1].ID)

	assert.Len(t, *company[0].Events, 2)
	assert.Equal(t, event2ID, (*company[0].Events)[0].ID)
	assert.Equal(t, event1ID, (*company[0].Events)[1].ID)

	assert.Len(t, *company[0].Persons, 2)
	assert.Equal(t, person2ID, (*company[0].Persons)[0].ID)
	assert.Equal(t, person1ID, (*company[0].Persons)[1].ID)
}

// -------- Update tests: --------

func TestUpdate_ShouldUpdateCompany(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	// create a company

	id := uuid.New()
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "Some AB",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("More notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}

	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	// update a company

	companyTypeToUpdate := models.CompanyType(models.CompanyTypeConsultancy)
	lastContactToUpdate := time.Now().AddDate(0, 2, 0)

	updateModel := models.UpdateCompany{
		ID:          id,
		Name:        testutil.ToPtr("a different name"),
		CompanyType: &companyTypeToUpdate,
		Notes:       testutil.ToPtr("Different notes"),
		LastContact: &lastContactToUpdate,
	}

	updatedDateApproximation := time.Now()
	err = companyRepository.Update(&updateModel)
	assert.NoError(t, err)

	// get the company and verify that it's updated

	retrievedCompany, err := companyRepository.GetById(&id)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompany)

	assert.Equal(t, updateModel.ID, retrievedCompany.ID)
	assert.Equal(t, *updateModel.Name, *retrievedCompany.Name)
	assert.Equal(t, updateModel.CompanyType.String(), retrievedCompany.CompanyType.String())
	assert.Equal(t, *updateModel.Notes, *retrievedCompany.Notes)
	testutil.AssertEqualFormattedDateTimes(t, retrievedCompany.LastContact, retrievedCompany.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, retrievedCompany.CreatedDate, insertedCompany.CreatedDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, retrievedCompany.UpdatedDate, time.Second)
}

func TestUpdateCompany_ShouldUpdateASingleField(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	// create a company

	id := uuid.New()
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	// update Name

	nameUpdateModel := models.UpdateCompany{
		ID:   id,
		Name: testutil.ToPtr("a different name"),
	}
	retrievedCompany := updateAndGetCompany(t, companyRepository, nameUpdateModel)
	assert.Equal(t, nameUpdateModel.Name, retrievedCompany.Name)

	// update CompanyType

	var companyTypeToUpdate models.CompanyType = models.CompanyTypeRecruiter
	companyTypeUpdateModel := models.UpdateCompany{
		ID:          id,
		CompanyType: &companyTypeToUpdate,
	}
	retrievedCompany = updateAndGetCompany(t, companyRepository, companyTypeUpdateModel)
	assert.Equal(t, companyTypeUpdateModel.CompanyType.String(), retrievedCompany.CompanyType.String())

	// update CompanyType

	notesUpdateModel := models.UpdateCompany{
		ID:    id,
		Notes: testutil.ToPtr("additional notes"),
	}
	retrievedCompany = updateAndGetCompany(t, companyRepository, notesUpdateModel)
	assert.Equal(t, notesUpdateModel.Notes, retrievedCompany.Notes)

	// update CompanyType

	lastContactUpdateModel := models.UpdateCompany{
		ID:          id,
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	retrievedCompany = updateAndGetCompany(t, companyRepository, lastContactUpdateModel)
	testutil.AssertEqualFormattedDateTimes(t, retrievedCompany.LastContact, lastContactUpdateModel.LastContact)
}

func TestUpdate_ShouldNotReturnErrorIfCompanyDoesNotExist(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	updateModel := models.UpdateCompany{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("a different name"),
		CompanyType: testutil.ToPtr(models.CompanyType(models.CompanyTypeConsultancy)),
		Notes:       testutil.ToPtr("Different notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
	}

	err := companyRepository.Update(&updateModel)
	assert.NoError(t, err)
}

func updateAndGetCompany(
	t *testing.T,
	companyRepository *repositories.CompanyRepository,
	updateCompany models.UpdateCompany,
) *models.Company {

	err := companyRepository.Update(&updateCompany)
	assert.NoError(t, err)

	retrievedCompany, err := companyRepository.GetById(&updateCompany.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompany)
	assert.Equal(t, updateCompany.ID, retrievedCompany.ID)

	return retrievedCompany
}

// -------- Delete tests: --------

func TestDelete_ShouldDeleteCompany(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	id := uuid.New()
	repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		&id,
		nil,
	)

	err := companyRepository.Delete(&id)
	assert.NoError(t, err)

	deletedCompany, err := companyRepository.GetById(&id)
	assert.Nil(t, deletedCompany)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", notFoundError.Error())
}

func TestDelete_ShouldReturnErrorIfCompanyIdIsNil(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	err := companyRepository.Delete(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

func TestDelete_ShouldReturnNotFoundErrorIfCompanyIdDoesNotExist(t *testing.T) {
	companyRepository, _, _, _, _, _ := setupCompanyRepository(t)

	id := uuid.New()

	err := companyRepository.Delete(&id)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Company does not exist. ID: "+id.String(), notFoundError.Error())
}
