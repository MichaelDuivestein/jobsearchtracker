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

func setupCompanyService(t *testing.T) (*services.CompanyService,
	*repositories.ApplicationRepository,
	*repositories.PersonRepository,
	*repositories.CompanyPersonRepository) {
	config := &configPackage.Config{
		DatabaseMigrationsPath:               "../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyServiceTestContainer(t, *config)

	var companyService *services.CompanyService
	err := container.Invoke(func(companySvc *services.CompanyService) {
		companyService = companySvc
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(applicationRepo *repositories.ApplicationRepository) {
		applicationRepository = applicationRepo
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

	return companyService, applicationRepository, personRepository, companyPersonRepository
}

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldWork(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

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
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, *companyToInsert.ID, id)
	assert.Equal(t, companyToInsert.Name, *insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), insertedCompany.CompanyType.String())
	assert.Equal(t, companyToInsert.Notes, insertedCompany.Notes)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.LastContact, companyToInsert.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.CreatedDate, companyToInsert.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.UpdatedDate, companyToInsert.UpdatedDate)
}

func TestCreateCompany_ShouldHandleEmptyFields(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	companyToInsert := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
	}

	insertedDateApproximation := time.Now()
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, companyToInsert.Name, *insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), insertedCompany.CompanyType.String())
	assert.Nil(t, insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact)
	testutil.AssertDateTimesWithinDelta(t, &insertedDateApproximation, insertedCompany.CreatedDate, time.Second)
	assert.Nil(t, insertedCompany.UpdatedDate)
}

func TestCreateCompany_ShouldHandleUnsetCreatedDate(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	companyToInsert := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: &time.Time{},
	}

	insertedDateApproximation := time.Now()
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, companyToInsert.Name, *insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), insertedCompany.CompanyType.String())
	assert.Nil(t, insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact)
	testutil.AssertDateTimesWithinDelta(t, &insertedDateApproximation, insertedCompany.CreatedDate, time.Second)
	assert.Nil(t, insertedCompany.UpdatedDate)
}

func TestCreateCompany_ShouldSetUnsetLastContactToCreatedDate(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	companyToInsert := models.CreateCompany{
		ID:          nil,
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       nil,
		LastContact: &time.Time{},
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate: nil,
	}
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, companyToInsert.Name, *insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType.String(), insertedCompany.CompanyType.String())
	assert.Nil(t, insertedCompany.Notes)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.LastContact, insertedCompany.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, companyToInsert.CreatedDate, insertedCompany.CreatedDate)
	assert.Nil(t, insertedCompany.UpdatedDate)
}

// -------- GetCompanyById tests: --------

func TestGetCompanyById_ShouldWork(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	id := uuid.New()
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now()),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)

	retrievedCompany, err := companyService.GetCompanyById(&id)
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

func TestGetCompanyById_ShouldReturnNotFoundErrorForAnIdThatDoesNotExist(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	nonExistingId := uuid.New()
	retrievedCompany, err := companyService.GetCompanyById(&nonExistingId)
	assert.NotNil(t, err)
	assert.Nil(t, retrievedCompany)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+nonExistingId.String()+"'", notFoundError.Error())

	id := uuid.New()
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now()),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)

	retrievedCompany, err = companyService.GetCompanyById(&nonExistingId)
	assert.NotNil(t, err)
	assert.Nil(t, retrievedCompany)

	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+nonExistingId.String()+"'", err.Error())
}

// -------- GetCompaniesByName tests: --------

func TestGetCompaniesByName_ShouldReturnASingleCompany(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// insert companies
	id1 := uuid.New()
	companyToInsert1 := models.CreateCompany{
		ID:          &id1,
		Name:        "Software House",
		CompanyType: models.CompanyTypeConsultancy,
	}
	_, err := companyService.CreateCompany(&companyToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	companyToInsert2 := models.CreateCompany{
		ID:          &id2,
		Name:        "Development Corp",
		CompanyType: models.CompanyTypeRecruiter,
	}
	_, err = companyService.CreateCompany(&companyToInsert2)
	assert.NoError(t, err)

	// GetByName
	nameToGet := "Corp"
	companies, err := companyService.GetCompaniesByName(&nameToGet)
	assert.NoError(t, err)
	assert.NotNil(t, companies)
	assert.Len(t, companies, 1)

	assert.Equal(t, id2, companies[0].ID)
}

func TestGetCompaniesByName_ShouldReturnMultipleCompanies(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// insert companies

	id1 := uuid.New()
	companyToInsert1 := models.CreateCompany{
		ID:          &id1,
		Name:        "Sunday Developers",
		CompanyType: models.CompanyTypeEmployer,
	}
	_, err := companyService.CreateCompany(&companyToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	companyToInsert2 := models.CreateCompany{
		ID:          &id2,
		Name:        "Brand AB",
		CompanyType: models.CompanyTypeEmployer,
	}
	_, err = companyService.CreateCompany(&companyToInsert2)
	assert.NoError(t, err)

	id3 := uuid.New()
	companyToInsert3 := models.CreateCompany{
		ID:          &id3,
		Name:        "Day Workers",
		CompanyType: models.CompanyTypeRecruiter,
	}
	_, err = companyService.CreateCompany(&companyToInsert3)
	assert.NoError(t, err)

	// GetByName

	nameToGet := "day"
	companies, err := companyService.GetCompaniesByName(&nameToGet)
	assert.NoError(t, err)
	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, id1, companies[1].ID)
	assert.Equal(t, id3, companies[0].ID)
}

func TestGetCompaniesByName_ShouldReturnNotFoundErrorIfNoNamesMatch(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// insert companies
	id1 := uuid.New()
	companyToInsert1 := models.CreateCompany{
		ID:          &id1,
		Name:        "Trickery AB",
		CompanyType: models.CompanyTypeConsultancy,
	}
	_, err := companyService.CreateCompany(&companyToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	companyToInsert2 := models.CreateCompany{
		ID:          &id2,
		Name:        "Offshoring Inc.",
		CompanyType: models.CompanyTypeEmployer,
	}
	_, err = companyService.CreateCompany(&companyToInsert2)
	assert.NoError(t, err)

	// GetByName
	nameToGet := "Bee"
	companies, err := companyService.GetCompaniesByName(&nameToGet)
	assert.Nil(t, companies)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Name: '"+nameToGet+"'", err.Error())
}

// -------- GetAllCompanies tests: --------

func TestGetAllCompanies_ShouldWork(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// insert companies

	company1Id := uuid.New()
	company1ToInsert := models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("company 1 notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedCompany1, err := companyService.CreateCompany(&company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2Id := uuid.New()
	company2ToInsert := models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("company 2 notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -4, 22)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	insertedCompany2, err := companyService.CreateCompany(&company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3Id := uuid.New()
	company3ToInsert := models.CreateCompany{
		ID:          &company3Id,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	insertedCompany3, err := companyService.CreateCompany(&company3ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 3)

	assert.Equal(t, company3Id, results[0].ID)
	assert.Equal(t, company3ToInsert.Name, *results[0].Name)
	assert.Equal(t, company3ToInsert.CompanyType.String(), results[0].CompanyType.String())
	assert.Nil(t, results[0].Notes)
	assert.Nil(t, results[0].LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company3ToInsert.CreatedDate, insertedCompany3.CreatedDate)
	assert.Nil(t, results[0].UpdatedDate)

	assert.Equal(t, company2Id, results[1].ID)
	assert.Equal(t, company2ToInsert.Name, *results[1].Name)
	assert.Equal(t, company2ToInsert.CompanyType.String(), results[1].CompanyType.String())
	assert.Equal(t, company2ToInsert.Notes, results[1].Notes)
	testutil.AssertEqualFormattedDateTimes(t, company2ToInsert.LastContact, insertedCompany2.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company2ToInsert.CreatedDate, insertedCompany2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company2ToInsert.UpdatedDate, insertedCompany2.UpdatedDate)

	assert.Equal(t, company1Id, results[2].ID)
	assert.Equal(t, company1ToInsert.Name, *results[2].Name)
	assert.Equal(t, company1ToInsert.CompanyType.String(), results[2].CompanyType.String())
	assert.Equal(t, company1ToInsert.Notes, results[2].Notes)
	testutil.AssertEqualFormattedDateTimes(t, company1ToInsert.LastContact, insertedCompany1.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company1ToInsert.CreatedDate, insertedCompany1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company1ToInsert.UpdatedDate, insertedCompany1.UpdatedDate)
}

func TestGetAllCompanies_ShouldReturnNilIfNoCompaniesInDatabase(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestGetAllCompanies_ShouldReturnASingleEntryWhenCompanyIDAndRecruiterIDAreTheSame(t *testing.T) {
	companyService, applicationRepository, _, _ := setupCompanyService(t)

	// insert company

	companyId := uuid.New()
	companyToInsert := &models.CreateCompany{
		ID:          &companyId,
		Name:        "companyName",
		CompanyType: models.CompanyTypeConsultancy,
	}
	insertedCompany, err := companyService.CreateCompany(companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	// insert application

	applicationId := uuid.New()
	application := models.CreateApplication{
		ID:               &applicationId,
		CompanyID:        &companyId,
		RecruiterID:      &companyId,
		JobTitle:         testutil.ToPtr("ApplicationJobTitle"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}

	_, err = applicationRepository.Create(&application)
	assert.NoError(t, err)

	// get all companies

	idResults, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeIDs, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, idResults)
	assert.Len(t, idResults, 1)

	assert.Equal(t, companyId, idResults[0].ID)
	assert.Len(t, *idResults[0].Applications, 1)

	allResults, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeAll, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, allResults)
	assert.Len(t, allResults, 1)

	assert.Equal(t, companyId, allResults[0].ID)
	assert.Len(t, *allResults[0].Applications, 1)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithApplicationIDsIfIncludeApplicationsIsIDs(t *testing.T) {
	companyService, applicationRepository, _, _ := setupCompanyService(t)

	// insert companies

	company1Id := uuid.New()
	company1ToInsert := &models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	insertedCompany1, err := companyService.CreateCompany(company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2Id := uuid.New()
	company2ToInsert := &models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	insertedCompany2, err := companyService.CreateCompany(company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3Id := uuid.New()
	company3ToInsert := &models.CreateCompany{
		ID:          &company3Id,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	insertedCompany3, err := companyService.CreateCompany(company3ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// insert applications

	application1Id := uuid.New()
	application1 := models.CreateApplication{
		ID:                   &application1Id,
		CompanyID:            &company2Id,
		RecruiterID:          &company3Id,
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
	_, err = applicationRepository.Create(&application1)
	assert.NoError(t, err)

	application2Id := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application2Id,
		&company2Id,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)))

	application3Id := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application3Id,
		nil,
		&company2Id,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)))

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeIDs, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 3)

	assert.Equal(t, company3Id, results[0].ID)
	assert.Len(t, *results[0].Applications, 1)

	results0Application := (*results[0].Applications)[0]
	assert.Equal(t, application1Id, results0Application.ID)
	assert.Equal(t, company2Id, *results0Application.CompanyID)
	assert.Equal(t, company3Id, *results0Application.RecruiterID)

	assert.Equal(t, company2Id, results[1].ID)
	assert.Len(t, *results[1].Applications, 3)

	results1Applications := *results[1].Applications
	assert.Equal(t, application1Id, results1Applications[0].ID)
	assert.Equal(t, company2Id, *results1Applications[0].CompanyID)
	assert.Equal(t, company3Id, *results1Applications[0].RecruiterID)
	assert.Nil(t, results1Applications[0].JobTitle)
	assert.Nil(t, results1Applications[0].JobAdURL)
	assert.Nil(t, results1Applications[0].Country)
	assert.Nil(t, results1Applications[0].Area)
	assert.Nil(t, results1Applications[0].RemoteStatusType)
	assert.Nil(t, results1Applications[0].WeekdaysInOffice)
	assert.Nil(t, results1Applications[0].EstimatedCycleTime)
	assert.Nil(t, results1Applications[0].EstimatedCommuteTime)
	assert.Nil(t, results1Applications[0].ApplicationDate)
	assert.Nil(t, results1Applications[0].CreatedDate)
	assert.Nil(t, results1Applications[0].UpdatedDate)

	assert.Equal(t, application2Id, results1Applications[1].ID)
	assert.Equal(t, company2Id, *results1Applications[1].CompanyID)
	assert.Nil(t, results1Applications[1].RecruiterID)

	assert.Equal(t, application3Id, results1Applications[2].ID)
	assert.Nil(t, results1Applications[2].CompanyID)
	assert.Equal(t, company2Id, *results1Applications[2].RecruiterID)

	assert.Equal(t, company1Id, results[2].ID)
	assert.Nil(t, results[2].Applications)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsIDsAndThereAreNoApplications(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// insert companies

	company1Id := uuid.New()
	company1ToInsert := &models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	insertedCompany1, err := companyService.CreateCompany(company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2Id := uuid.New()
	company2ToInsert := &models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	insertedCompany2, err := companyService.CreateCompany(company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3Id := uuid.New()
	company3ToInsert := &models.CreateCompany{
		ID:          &company3Id,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	insertedCompany3, err := companyService.CreateCompany(company3ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeIDs, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 3)

	assert.Equal(t, company3Id, results[0].ID)
	assert.Nil(t, results[0].Applications)

	assert.Equal(t, company2Id, results[1].ID)
	assert.Nil(t, results[1].Applications)

	assert.Equal(t, company1Id, results[2].ID)
	assert.Nil(t, results[2].Applications)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithApplicationsIfIncludeApplicationsIsAll(t *testing.T) {
	companyService, applicationRepository, _, _ := setupCompanyService(t)

	// insert companies

	company1Id := uuid.New()
	company1ToInsert := &models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	insertedCompany1, err := companyService.CreateCompany(company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2Id := uuid.New()
	company2ToInsert := &models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	insertedCompany2, err := companyService.CreateCompany(company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3Id := uuid.New()
	company3ToInsert := &models.CreateCompany{
		ID:          &company3Id,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	insertedCompany3, err := companyService.CreateCompany(company3ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// insert applications

	application1Id := uuid.New()
	application1 := models.CreateApplication{
		ID:                   &application1Id,
		CompanyID:            &company2Id,
		RecruiterID:          &company3Id,
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
	_, err = applicationRepository.Create(&application1)
	assert.NoError(t, err)

	application2Id := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application2Id,
		&company2Id,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)))

	application3Id := uuid.New()
	repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		&application3Id,
		nil,
		&company2Id,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)))

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeAll, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 3)

	assert.Equal(t, company3Id, results[0].ID)
	assert.Len(t, *results[0].Applications, 1)

	results0Application := (*results[0].Applications)[0]
	assert.Equal(t, application1Id, results0Application.ID)
	assert.Equal(t, company2Id, *results0Application.CompanyID)
	assert.Equal(t, company3Id, *results0Application.RecruiterID)

	assert.Equal(t, company2Id, results[1].ID)
	assert.Len(t, *results[1].Applications, 3)

	results1Applications := *results[1].Applications
	assert.Equal(t, application1Id, results1Applications[0].ID)
	assert.Equal(t, company2Id, *results1Applications[0].CompanyID)
	assert.Equal(t, company3Id, *results1Applications[0].RecruiterID)
	assert.Equal(t, application1.JobTitle, results1Applications[0].JobTitle)
	assert.Equal(t, application1.JobAdURL, results1Applications[0].JobAdURL)
	assert.Equal(t, application1.Country, results1Applications[0].Country)
	assert.Equal(t, application1.Area, results1Applications[0].Area)
	assert.Equal(t, application1.RemoteStatusType.String(), results1Applications[0].RemoteStatusType.String())
	assert.Equal(t, application1.WeekdaysInOffice, results1Applications[0].WeekdaysInOffice)
	assert.Equal(t, application1.EstimatedCycleTime, results1Applications[0].EstimatedCycleTime)
	assert.Equal(t, application1.EstimatedCommuteTime, results1Applications[0].EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, application1.ApplicationDate, results1Applications[0].ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, application1.CreatedDate, results1Applications[0].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, application1.UpdatedDate, results1Applications[0].UpdatedDate)

	assert.Equal(t, application2Id, results1Applications[1].ID)
	assert.Equal(t, company2Id, *results1Applications[1].CompanyID)
	assert.Nil(t, results1Applications[1].RecruiterID)

	assert.Equal(t, application3Id, results1Applications[2].ID)
	assert.Nil(t, results1Applications[2].CompanyID)
	assert.Equal(t, company2Id, *results1Applications[2].RecruiterID)

	assert.Equal(t, company1Id, results[2].ID)
	assert.Nil(t, results[2].Applications)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsAllAndThereAreNoApplications(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// insert companies

	company1Id := uuid.New()
	company1ToInsert := &models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	insertedCompany1, err := companyService.CreateCompany(company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2Id := uuid.New()
	company2ToInsert := &models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	insertedCompany2, err := companyService.CreateCompany(company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3Id := uuid.New()
	company3ToInsert := &models.CreateCompany{
		ID:          &company3Id,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	insertedCompany3, err := companyService.CreateCompany(company3ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeAll, models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 3)

	assert.Equal(t, company3Id, results[0].ID)
	assert.Nil(t, results[0].Applications)

	assert.Equal(t, company2Id, results[1].ID)
	assert.Nil(t, results[1].Applications)

	assert.Equal(t, company1Id, results[2].ID)
	assert.Nil(t, results[2].Applications)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithPersonIDsIfIncludePersonsIsIDs(t *testing.T) {
	companyService, _, personRepository, companyPersonRepository := setupCompanyService(t)

	// insert companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyService.CreateCompany(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyService.CreateCompany(&createCompany2)
	assert.NoError(t, err)

	company3ID := uuid.New()
	createCompany3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	_, err = companyService.CreateCompany(&createCompany3)
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

	// get all companies

	companies, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeIDs)
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

func TestGetAllCompanies_ShouldReturnCompaniesWithNoPersonsIfIncludePersonsIsSetToIDsAndThereAreNoPersons(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyService.CreateCompany(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyService.CreateCompany(&createCompany2)
	assert.NoError(t, err)

	companies, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Nil(t, companies[0].Persons)

	assert.Equal(t, company1ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)
}

func TestGetAll_ShouldReturnCompaniesWithNilPersonsIfIncludePersonsIsSetToIDsAndThereAreNoCompanyPersons(t *testing.T) {
	companyService, _, personRepository, _ := setupCompanyService(t)

	// create companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyService.CreateCompany(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyService.CreateCompany(&createCompany2)
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

	companies, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company2ID, companies[0].ID)
	assert.Nil(t, companies[0].Persons)

	assert.Equal(t, company1ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)
}

func TestGetAll_ShouldReturnCompaniesWithPersonsIfIncludePersonsIsSetToAll(t *testing.T) {
	companyService, _, personRepository, companyPersonRepository := setupCompanyService(t)

	// create companies

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}
	_, err := companyService.CreateCompany(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}
	_, err = companyService.CreateCompany(&createCompany2)
	assert.NoError(t, err)

	company3ID := uuid.New()
	createCompany3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}
	_, err = companyService.CreateCompany(&createCompany3)
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

	companies, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeAll)
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
	testutil.AssertEqualFormattedDateTimes(t, person1.CreatedDate, person1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, person1.UpdatedDate, person1.UpdatedDate)
}

func TestGetAll_ShouldReturnCompaniesWithNilPersonsIfIncludePersonsIsSetToAllAndThereAreNoPersonsInDB(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	company1ID := uuid.New()
	createCompany1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err := companyService.CreateCompany(&createCompany1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	createCompany2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyService.CreateCompany(&createCompany2)
	assert.NoError(t, err)

	companies, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone, models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)

	assert.Equal(t, company1ID, companies[0].ID)
	assert.Nil(t, companies[0].Persons)

	assert.Equal(t, company2ID, companies[1].ID)
	assert.Nil(t, companies[1].Persons)
}

// -------- UpdateCompany tests: --------

func TestUpdateCompany_ShouldWork(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// insert a company:
	id := uuid.New()
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "Some Stockholm-based AB",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("Notes about an AB"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
	}
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	// update a company:

	var companyTypeToUpdate models.CompanyType = models.CompanyTypeConsultancy
	updateModel := models.UpdateCompany{
		ID:          id,
		Name:        testutil.ToPtr("Updated Name"),
		CompanyType: &companyTypeToUpdate,
		Notes:       testutil.ToPtr("Updated Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	updatedDateApproximation := time.Now()
	err = companyService.UpdateCompany(&updateModel)
	assert.NoError(t, err)

	// get the company to ensure that the changes have been applied.
	retrievedCompany, err := companyService.GetCompanyById(&id)
	assert.NoError(t, err)

	assert.NotNil(t, retrievedCompany)
	assert.Equal(t, id, retrievedCompany.ID)
	assert.Equal(t, updateModel.Name, retrievedCompany.Name)
	assert.Equal(t, companyTypeToUpdate.String(), retrievedCompany.CompanyType.String())
	testutil.AssertEqualFormattedDateTimes(t, updateModel.LastContact, retrievedCompany.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, insertedCompany.CreatedDate, retrievedCompany.CreatedDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, retrievedCompany.UpdatedDate, time.Second)
}

func TestUpdateCompany_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	var companyTypeToUpdate models.CompanyType = models.CompanyTypeConsultancy
	updateModel := models.UpdateCompany{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Updated Name"),
		CompanyType: &companyTypeToUpdate,
		Notes:       testutil.ToPtr("Updated Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	err := companyService.UpdateCompany(&updateModel)
	assert.NoError(t, err)
}

// -------- DeleteCompany tests: --------

func TestDeleteCompany_ShouldWork(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	// create a company:

	id := uuid.New()
	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now()),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}

	_, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)

	// delete the company:

	err = companyService.DeleteCompany(&id)
	assert.NoError(t, err)

	// try to get the company:
	// this should return an error as the company no longer exists.

	deletedCompany, err := companyService.GetCompanyById(&id)
	assert.NotNil(t, err)
	assert.Nil(t, deletedCompany)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
}

func TestDeleteCompany_ShouldReturnNotFoundErrorIfIdToDeleteDoesNotExist(t *testing.T) {
	companyService, _, _, _ := setupCompanyService(t)

	id := uuid.New()

	err := companyService.DeleteCompany(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Company does not exist. ID: "+id.String(), err.Error())
}
