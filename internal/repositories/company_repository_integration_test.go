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

func setupCompanyRepository(t *testing.T) (*repositories.CompanyRepository, *repositories.ApplicationRepository) {
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

	return companyRepository, applicationRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertAndReturnCompany(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

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
	assert.Equal(t, company.Name, insertedCompany.Name)
	assert.Equal(t, company.CompanyType, insertedCompany.CompanyType)
	assert.Equal(t, company.Notes, insertedCompany.Notes)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.LastContact, company.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, &insertedCompany.CreatedDate, company.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.UpdatedDate, company.UpdatedDate)
}

func TestCreate_ShouldInsertCompanyWithMinimumRequiredFields(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	company := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
	}
	createdDateApproximation := time.Now()
	insertedCompany, err := companyRepository.Create(&company)

	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, company.Name, insertedCompany.Name)
	assert.Equal(t, company.CompanyType, insertedCompany.CompanyType)
	assert.Nil(t, insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, &createdDateApproximation, &insertedCompany.CreatedDate)
	assert.Nil(t, insertedCompany.UpdatedDate)
}

func TestCreate_ShouldReturnConflictErrorOnDuplicateCompanyId(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	id := uuid.New()
	firstCompany := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now()),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
	}
	firstInsertedCompany, err := companyRepository.Create(&firstCompany)
	assert.NoError(t, err)
	assert.NotNil(t, firstInsertedCompany)
	assert.Equal(t, firstInsertedCompany.ID, id)

	secondCompany := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("Other notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	shouldBeNil, err := companyRepository.Create(&secondCompany)
	assert.Nil(t, shouldBeNil)
	assert.NotNil(t, err)

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(t,
		"conflict error on insert: ID already exists in database: '"+id.String()+"'",
		err.Error())
}

// -------- GetById tests: --------

func TestGetById_ShouldGetCompany(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

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
	assert.Equal(t, companyToInsert.Name, retrievedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType, retrievedCompany.CompanyType)
	assert.Equal(t, companyToInsert.Notes, retrievedCompany.Notes)

	retrievedCompanyLastContact := retrievedCompany.LastContact.Format(time.RFC3339)
	companyToInsertLastContact := companyToInsert.LastContact.Format(time.RFC3339)
	assert.Equal(t, companyToInsertLastContact, retrievedCompanyLastContact)

	retrievedCompanyCreatedDate := retrievedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := companyToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, retrievedCompanyCreatedDate)

	retrievedCompanyUpdatedDate := retrievedCompany.UpdatedDate.Format(time.RFC3339)
	companyToInsertUpdatedDate := companyToInsert.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertUpdatedDate, retrievedCompanyUpdatedDate)
}

func TestGetById_ShouldReturnErrorIfCompanyIDIsNil(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	response, err := companyRepository.GetById(nil)
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, "validation error on field 'ID': ID is nil", err.Error())
}

func TestGetById_ShouldReturnErrorIfCompanyIDDoesNotExist(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	id := uuid.New()
	response, err := companyRepository.GetById(&id)
	assert.Nil(t, response)
	assert.NotNil(t, err, err.Error())
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
}

// -------- GetAllByName tests: --------

func TestGetAllByName_ShouldReturnCompany(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	companyToInsert := models.CreateCompany{
		Name:        "Company Bee",
		CompanyType: models.CompanyTypeRecruiter,
	}
	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	retrievedCompanies, err := companyRepository.GetAllByName(&insertedCompany.Name)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompanies)
	assert.Len(t, retrievedCompanies, 1)

	assert.Equal(t, "Company Bee", retrievedCompanies[0].Name)
}

func TestGetAllByName_ShouldReturnValidationErrorIfCompanyNameIsNil(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	retrievedCompanies, err := companyRepository.GetAllByName(nil)
	assert.Nil(t, retrievedCompanies)
	assert.NotNil(t, err)
	assert.Equal(t, "validation error: name is nil", err.Error())
}

func TestGetAllByName_ShouldReturnNotFoundErrorIfCompanyNameDoesNotExist(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	name := "Doesnt Exist"

	company, err := companyRepository.GetAllByName(&name)
	assert.Nil(t, company)
	assert.NotNil(t, err)
	assert.Equal(t, "error: object not found: Name: '"+name+"'", err.Error())
}

func TestGetAllByName_ShouldReturnMultipleCompaniesWithSameName(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

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

func TestGetAllByName_ShouldReturnMultipleCompaniesWithSameNamePart(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	// insert some companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Some AB",
		CompanyType: models.CompanyTypeRecruiter,
	}
	insertedCompany1, err := companyRepository.Create(&company1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Absolutely not a company name",
		CompanyType: models.CompanyTypeConsultancy,
	}
	insertedCompany2, err := companyRepository.Create(&company2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Different AB",
		CompanyType: models.CompanyTypeEmployer,
	}
	insertedCompany3, err := companyRepository.Create(&company3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// get companies containing "ab"

	retrievedCompanies, err := companyRepository.GetAllByName(testutil.ToPtr("ab"))
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompanies)
	assert.Len(t, retrievedCompanies, 3)
	assert.Equal(t, insertedCompany2.ID, retrievedCompanies[0].ID)
	assert.Equal(t, insertedCompany3.ID, retrievedCompanies[1].ID)
	assert.Equal(t, insertedCompany1.ID, retrievedCompanies[2].ID)
}

// -------- GetAll tests: --------

func TestGetAll_ShouldReturnAllCompanies(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

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
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -4, 22)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedCompany2, err := companyRepository.Create(&company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	results, err := companyRepository.GetAll(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)
	assert.Equal(t, *company2ToInsert.ID, results[0].ID)
	assert.Equal(t, *company1ToInsert.ID, results[1].ID)
}

func TestGetAll_ShouldReturnNilIfNoCompaniesInDatabase(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	companies, err := companyRepository.GetAll(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, companies)
}

func TestGetAll_ShouldReturnApplicationIDsIfIncludeApplicationsIsSetToIDs(t *testing.T) {
	companyRepository, applicationRepository := setupCompanyRepository(t)

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

	companies, err := companyRepository.GetAll(models.IncludeExtraDataTypeIDs)
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
	assert.Equal(t, application1ID, company2Application1.ID)
	assert.Equal(t, company1ID, *company2Application1.CompanyID)
	assert.Equal(t, company2ID, *company2Application1.RecruiterID)
	assert.Nil(t, company2Application1.JobTitle)
	assert.Nil(t, company2Application1.JobAdURL)
	assert.Nil(t, company2Application1.Country)
	assert.Nil(t, company2Application1.Area)
	assert.Nil(t, company2Application1.RemoteStatusType)
	assert.Nil(t, company2Application1.WeekdaysInOffice)
	assert.Nil(t, company2Application1.EstimatedCycleTime)
	assert.Nil(t, company2Application1.EstimatedCommuteTime)
	assert.Nil(t, company2Application1.ApplicationDate)
	assert.Nil(t, company2Application1.CreatedDate)
	assert.Nil(t, company2Application1.UpdatedDate)

	company2Application2 := (*companies[0].Applications)[1]
	assert.Equal(t, application2ID, company2Application2.ID)
	assert.Equal(t, company2ID, *company2Application2.CompanyID)
	assert.Nil(t, company2Application2.RecruiterID)

	company2Application3 := (*companies[0].Applications)[2]
	assert.Equal(t, application3ID, company2Application3.ID)
	assert.Nil(t, company2Application3.CompanyID)
	assert.Equal(t, company2ID, *company2Application3.RecruiterID)

	company1Application1 := (*companies[2].Applications)[0]
	assert.Equal(t, application1ID, company1Application1.ID)
	assert.Equal(t, company1ID, *company1Application1.CompanyID)
	assert.Equal(t, company2ID, *company1Application1.RecruiterID)
}

func TestGetAll_ShouldReturnNilApplicationsIfIncludeApplicationsIsSetToIDsAndThereAreNoApplicationsInDB(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

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

	companies, err := companyRepository.GetAll(models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Nil(t, companies[0].Applications)

	assert.Equal(t, company1ID, companies[1].ID)
	assert.Nil(t, companies[1].Applications)
}

func TestGetAll_ShouldReturnApplicationsIfIncludeApplicationsIsSetToAll(t *testing.T) {
	companyRepository, applicationRepository := setupCompanyRepository(t)

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
	applicationDate := time.Now().AddDate(0, 0, 1)
	createdDate := time.Now().AddDate(0, 0, 2)
	updatedDate := time.Now().AddDate(0, 0, 3)
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
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
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

	companies, err := companyRepository.GetAll(models.IncludeExtraDataTypeAll)
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
	assert.Equal(t, application1ID, company2Application1.ID)
	assert.Equal(t, company1ID, *company2Application1.CompanyID)
	assert.Equal(t, company2ID, *company2Application1.RecruiterID)
	assert.Equal(t, "Application1JobTitle", *company2Application1.JobTitle)
	assert.Equal(t, "Application1JobAdURL", *company2Application1.JobAdURL)
	assert.Equal(t, "Application1Country", *company2Application1.Country)
	assert.Equal(t, "Application1Area", *company2Application1.Area)
	assert.Equal(t, application1RemoteStatusType.String(), company2Application1.RemoteStatusType.String())
	assert.Equal(t, 10, *company2Application1.WeekdaysInOffice)
	assert.Equal(t, 11, *company2Application1.EstimatedCycleTime)
	assert.Equal(t, 12, *company2Application1.EstimatedCommuteTime)

	retrievedApplicationDate := company2Application1.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationDate.Format(time.RFC3339), retrievedApplicationDate)

	retrievedCreatedDate := company2Application1.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDate.Format(time.RFC3339), retrievedCreatedDate)

	retrievedUpdatedDate := company2Application1.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDate.Format(time.RFC3339), retrievedUpdatedDate)

	company2Application2 := (*companies[0].Applications)[1]
	assert.Equal(t, application2ID, company2Application2.ID)
	assert.Equal(t, company2ID, *company2Application2.CompanyID)
	assert.Nil(t, company2Application2.RecruiterID)

	company2Application3 := (*companies[0].Applications)[2]
	assert.Equal(t, application3ID, company2Application3.ID)
	assert.Nil(t, company2Application3.CompanyID)
	assert.Equal(t, company2ID, *company2Application3.RecruiterID)

	company1Application1 := (*companies[2].Applications)[0]
	assert.Equal(t, application1ID, company1Application1.ID)
	assert.Equal(t, company1ID, *company1Application1.CompanyID)
	assert.Equal(t, company2ID, *company1Application1.RecruiterID)
}

func TestGetAll_ShouldReturnNilApplicationsIfIncludeApplicationsIsSetToAllAndThereAreNoApplicationsInDB(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

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

	companies, err := companyRepository.GetAll(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company1ID, companies[0].ID)
	assert.Nil(t, companies[0].Applications)

	assert.Equal(t, company2ID, companies[1].ID)
	assert.Nil(t, companies[1].Applications)
}

// -------- Update tests: --------

func TestUpdate_ShouldUpdateCompany(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	// create a company

	id := uuid.New()
	notes := "More notes"
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "Some AB",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}

	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	// update a company

	nameToUpdate := "a different name"
	companyTypeToUpdate := models.CompanyType(models.CompanyTypeConsultancy)
	notesToUpdate := "Different notes"
	lastContactToUpdate := time.Now().AddDate(0, 2, 0)

	updateModel := models.UpdateCompany{
		ID:          id,
		Name:        &nameToUpdate,
		CompanyType: &companyTypeToUpdate,
		Notes:       &notesToUpdate,
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
	assert.Equal(t, *updateModel.Name, retrievedCompany.Name)
	assert.Equal(t, *updateModel.CompanyType, retrievedCompany.CompanyType)
	assert.Equal(t, *updateModel.Notes, *retrievedCompany.Notes)
	testutil.AssertEqualFormattedDateTimes(t, retrievedCompany.LastContact, retrievedCompany.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, &retrievedCompany.CreatedDate, &insertedCompany.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, &updatedDateApproximation, retrievedCompany.UpdatedDate)
}

func TestUpdate_ShouldUpdateASingleField(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

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

	nameToUpdate := "a different name"
	nameUpdateModel := models.UpdateCompany{
		ID:   id,
		Name: &nameToUpdate,
	}
	retrievedCompany := updateAndGetCompany(t, companyRepository, nameUpdateModel)
	assert.Equal(t, nameToUpdate, retrievedCompany.Name)

	// update CompanyType

	var companyTypeToUpdate models.CompanyType = models.CompanyTypeRecruiter
	companyTypeUpdateModel := models.UpdateCompany{
		ID:          id,
		CompanyType: &companyTypeToUpdate,
	}
	retrievedCompany = updateAndGetCompany(t, companyRepository, companyTypeUpdateModel)
	assert.Equal(t, *companyTypeUpdateModel.CompanyType, retrievedCompany.CompanyType)

	// update CompanyType

	notesToUpdate := "additional notes"
	notesUpdateModel := models.UpdateCompany{
		ID:    id,
		Notes: &notesToUpdate,
	}
	retrievedCompany = updateAndGetCompany(t, companyRepository, notesUpdateModel)
	assert.Equal(t, notesToUpdate, *retrievedCompany.Notes)

	// update CompanyType

	lastContactUpdateModel := models.UpdateCompany{
		ID:          id,
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	retrievedCompany = updateAndGetCompany(t, companyRepository, lastContactUpdateModel)
	testutil.AssertEqualFormattedDateTimes(t, retrievedCompany.LastContact, lastContactUpdateModel.LastContact)
}

func TestUpdate_ShouldNotReturnErrorIfCompanyDoesNotExist(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	id := uuid.New()
	updateModel := models.UpdateCompany{
		ID:          id,
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
	companyRepository, _ := setupCompanyRepository(t)

	id := uuid.New()

	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now()),
		CreatedDate: testutil.ToPtr(time.Now()),
		UpdatedDate: testutil.ToPtr(time.Now()),
	}

	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	err = companyRepository.Delete(&id)
	assert.NoError(t, err)

	deletedCompany, err := companyRepository.GetById(&id)
	assert.NotNil(t, err)
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
	assert.Nil(t, deletedCompany)
}

func TestDelete_ShouldReturnErrorIfCompanyIdIsNil(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	err := companyRepository.Delete(nil)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationErr.Error())
}

func TestDelete_ShouldReturnNotFoundErrorIfCompanyIdDoesNotExist(t *testing.T) {
	companyRepository, _ := setupCompanyRepository(t)

	id := uuid.New()

	err := companyRepository.Delete(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Company does not exist. ID: "+id.String(), err.Error())
}
