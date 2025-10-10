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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupCompanyService(t *testing.T) (*services.CompanyService, *repositories.ApplicationRepository) {
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

	return companyService, applicationRepository
}

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldWork(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	insertedCompany, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, *companyToInsert.ID, id)
	assert.Equal(t, companyToInsert.Name, insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType)
	assert.Equal(t, companyToInsert.Notes, insertedCompany.Notes)

	insertedCompanyLastContact := insertedCompany.LastContact.Format(time.RFC3339)
	companyToInsertLastContact := companyToInsert.LastContact.Format(time.RFC3339)
	assert.Equal(t, companyToInsertLastContact, insertedCompanyLastContact)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := companyToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, insertedCompanyCreatedDate)

	insertedCompanyUpdatedDate := insertedCompany.UpdatedDate.Format(time.RFC3339)
	companyToInsertUpdatedDate := companyToInsert.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertUpdatedDate, insertedCompanyUpdatedDate)
}

func TestCreateCompany_ShouldHandleEmptyFields(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	companyToInsert := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
	}

	insertedDateApproximation := time.Now().Format(time.RFC3339)
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)

	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, companyToInsert.Name, insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType)
	assert.Nil(t, insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, insertedDateApproximation, insertedCompanyCreatedDate)

	assert.Nil(t, insertedCompany.UpdatedDate)
}

func TestCreateCompany_ShouldHandleUnsetCreatedDate(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	companyToInsert := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: &time.Time{},
	}

	insertedDateApproximation := time.Now().Format(time.RFC3339)
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)

	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, companyToInsert.Name, insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType)
	assert.Nil(t, insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, insertedDateApproximation, insertedCompanyCreatedDate)

	assert.Nil(t, insertedCompany.UpdatedDate)
}

func TestCreateCompany_ShouldSetUnsetLastContactToCreatedDate(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	createdDate := time.Now().AddDate(0, 0, -2)

	companyToInsert := models.CreateCompany{
		ID:          nil,
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       nil,
		LastContact: &time.Time{},
		CreatedDate: &createdDate,
		UpdatedDate: nil,
	}

	insertedCompany, err := companyService.CreateCompany(&companyToInsert)

	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	assert.Equal(t, companyToInsert.Name, insertedCompany.Name)
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType)
	assert.Nil(t, insertedCompany.Notes)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := companyToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, insertedCompanyCreatedDate)

	insertedCompanyLastContact := insertedCompany.LastContact.Format(time.RFC3339)
	assert.Equal(t, insertedCompanyCreatedDate, insertedCompanyLastContact)

	assert.Nil(t, insertedCompany.UpdatedDate)
}

// -------- GetCompanyById tests: --------

func TestGetCompanyById_ShouldWork(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now()
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, 2)

	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	_, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)

	retrievedCompany, err := companyService.GetCompanyById(&id)
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

func TestGetCompanyById_ShouldReturnNotFoundErrorForAnIdThatDoesNotExist(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	nonExistingId := uuid.New()
	retrievedCompany, err := companyService.GetCompanyById(&nonExistingId)
	assert.NotNil(t, err)
	assert.Nil(t, retrievedCompany)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+nonExistingId.String()+"'", notFoundError.Error())

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now()
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, 2)

	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
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
	companyService, _ := setupCompanyService(t)

	// insert companies
	id1 := uuid.New()
	name1 := "Software House"
	companyToInsert1 := models.CreateCompany{
		ID:          &id1,
		Name:        name1,
		CompanyType: models.CompanyTypeConsultancy,
	}
	_, err := companyService.CreateCompany(&companyToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	name2 := "Development Corp"
	companyToInsert2 := models.CreateCompany{
		ID:          &id2,
		Name:        name2,
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
	companyService, _ := setupCompanyService(t)

	// insert companies

	id1 := uuid.New()
	name1 := "Sunday Developers"
	companyToInsert1 := models.CreateCompany{
		ID:          &id1,
		Name:        name1,
		CompanyType: models.CompanyTypeEmployer,
	}
	_, err := companyService.CreateCompany(&companyToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	name2 := "Brand AB"
	companyToInsert2 := models.CreateCompany{
		ID:          &id2,
		Name:        name2,
		CompanyType: models.CompanyTypeEmployer,
	}
	_, err = companyService.CreateCompany(&companyToInsert2)
	assert.NoError(t, err)

	id3 := uuid.New()
	name3 := "Day Workers"
	companyToInsert3 := models.CreateCompany{
		ID:          &id3,
		Name:        name3,
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
	companyService, _ := setupCompanyService(t)

	// insert companies
	id1 := uuid.New()
	name1 := "Trickery AB"
	companyToInsert1 := models.CreateCompany{
		ID:          &id1,
		Name:        name1,
		CompanyType: models.CompanyTypeConsultancy,
	}
	_, err := companyService.CreateCompany(&companyToInsert1)
	assert.NoError(t, err)

	id2 := uuid.New()
	name2 := "Offshoring Inc."
	companyToInsert2 := models.CreateCompany{
		ID:          &id2,
		Name:        name2,
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
	companyService, applicationRepository := setupCompanyService(t)

	// insert companies

	company1Id := uuid.New()
	company1LastContact := time.Now().AddDate(-1, 0, 0)
	company1CreatedDate := time.Now().AddDate(0, -5, 0)
	company1UpdatedDate := time.Now().AddDate(0, 0, -3)

	company1ToInsert := models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("company 1 notes"),
		LastContact: &company1LastContact,
		CreatedDate: &company1CreatedDate,
		UpdatedDate: &company1UpdatedDate,
	}

	insertedCompany1, err := companyService.CreateCompany(&company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2Id := uuid.New()
	company2LastContact := time.Now().AddDate(-1, 0, 0)
	company2CreatedDate := time.Now().AddDate(0, -4, 22)
	company2UpdatedDate := time.Now().AddDate(0, 0, -3)

	company2ToInsert := models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("company 2 notes"),
		LastContact: &company2LastContact,
		CreatedDate: &company2CreatedDate,
		UpdatedDate: &company2UpdatedDate,
	}
	insertedCompany2, err := companyService.CreateCompany(&company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3Id := uuid.New()
	company3CreatedDate := time.Now().AddDate(0, 0, 4)
	company3ToInsert := models.CreateCompany{
		ID:          &company3Id,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: &company3CreatedDate,
	}
	insertedCompany3, err := companyService.CreateCompany(&company3ToInsert)
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
		RemoteStatusType:     models.RemoteStatusTypeRemote,
		WeekdaysInOffice:     testutil.ToPtr(0),
		EstimatedCycleTime:   testutil.ToPtr(1),
		EstimatedCommuteTime: testutil.ToPtr(2),
		ApplicationDate:      testutil.ToPtr(time.Now()),
		CreatedDate:          testutil.ToPtr(time.Now()),
		UpdatedDate:          testutil.ToPtr(time.Now()),
	}
	_, err = applicationRepository.Create(&application1)
	assert.NoError(t, err)

	application2Id := uuid.New()
	application2 := models.CreateApplication{
		ID:               &application2Id,
		CompanyID:        &company2Id,
		JobAdURL:         testutil.ToPtr("Application2JobAdURL"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
		CreatedDate:      testutil.ToPtr(time.Now()),
	}
	_, err = applicationRepository.Create(&application2)
	assert.NoError(t, err)

	application3Id := uuid.New()
	application3 := models.CreateApplication{
		ID:               &application3Id,
		RecruiterID:      &company2Id,
		JobTitle:         testutil.ToPtr("Application3JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeOffice,
		CreatedDate:      testutil.ToPtr(time.Now()),
	}
	_, err = applicationRepository.Create(&application3)
	assert.NoError(t, err)

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Len(t, results, 3)

	assert.Equal(t, company3Id, results[0].ID)
	assert.Equal(t, "company3Name", results[0].Name)
	assert.Equal(t, company3ToInsert.CompanyType, results[0].CompanyType)
	assert.Nil(t, results[0].Notes)
	assert.Nil(t, results[0].LastContact)

	company3ToInsertCreatedDate := company3ToInsert.CreatedDate.Format(time.RFC3339)
	insertedCompany3CreatedDate := insertedCompany3.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, company3ToInsertCreatedDate, insertedCompany3CreatedDate)

	assert.Nil(t, results[0].UpdatedDate)
	assert.Len(t, *results[0].Applications, 1)

	results0Application := (*results[0].Applications)[0]
	assert.Equal(t, application1Id, results0Application.ID)
	assert.Equal(t, company2Id, *results0Application.CompanyID)
	assert.Equal(t, company3Id, *results0Application.RecruiterID)
	assert.Equal(t, "Application1JobTitle", *results0Application.JobTitle)
	assert.Equal(t, "Application1JobAdURL", *results0Application.JobAdURL)
	assert.Equal(t, "Application1Country", *results0Application.Country)
	assert.Equal(t, "Application1Area", *results0Application.Area)
	assert.Equal(t, models.RemoteStatusTypeRemote, results0Application.RemoteStatusType.String())
	assert.Equal(t, 0, *results0Application.WeekdaysInOffice)
	assert.Equal(t, 1, *results0Application.EstimatedCycleTime)
	assert.Equal(t, 2, *results0Application.EstimatedCommuteTime)

	assert.Equal(t, company2Id, results[1].ID)
	assert.Equal(t, "company2Name", results[1].Name)
	assert.Equal(t, company2ToInsert.CompanyType, results[1].CompanyType)
	assert.Equal(t, "company 2 notes", *results[1].Notes)

	company2ToInsertLastContact := company2ToInsert.LastContact.Format(time.RFC3339)
	insertedCompany2LastContact := insertedCompany2.LastContact.Format(time.RFC3339)
	assert.Equal(t, company2ToInsertLastContact, insertedCompany2LastContact)

	company2ToInsertCreatedDate := company2ToInsert.CreatedDate.Format(time.RFC3339)
	insertedCompany2CreatedDate := insertedCompany2.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, company2ToInsertCreatedDate, insertedCompany2CreatedDate)

	company2ToInsertUpdatedDate := company2ToInsert.UpdatedDate.Format(time.RFC3339)
	insertedCompany2UpdatedDate := insertedCompany2.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, company2ToInsertUpdatedDate, insertedCompany2UpdatedDate)

	assert.Len(t, *results[1].Applications, 3)

	results1Applications := *results[1].Applications

	assert.Equal(t, application1Id, results1Applications[0].ID)
	assert.Equal(t, company2Id, *results1Applications[0].CompanyID)
	assert.Equal(t, company3Id, *results1Applications[0].RecruiterID)
	assert.Equal(t, "Application1JobTitle", *results1Applications[0].JobTitle)
	assert.Equal(t, "Application1JobAdURL", *results1Applications[0].JobAdURL)
	assert.Equal(t, "Application1Country", *results1Applications[0].Country)
	assert.Equal(t, "Application1Area", *results1Applications[0].Area)
	assert.Equal(t, models.RemoteStatusTypeRemote, results1Applications[0].RemoteStatusType.String())
	assert.Equal(t, 0, *results1Applications[0].WeekdaysInOffice)
	assert.Equal(t, 1, *results1Applications[0].EstimatedCycleTime)
	assert.Equal(t, 2, *results1Applications[0].EstimatedCommuteTime)

	application1ToInsertApplicationDate := application1.ApplicationDate.Format(time.RFC3339)
	insertedApplication2ApplicationDate := results1Applications[0].ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, application1ToInsertApplicationDate, insertedApplication2ApplicationDate)

	application1ToInsertCreatedDate := application1.CreatedDate.Format(time.RFC3339)
	insertedApplication1CreatedDate := results1Applications[0].CreatedDate.Format(time.RFC3339)
	assert.Equal(t, application1ToInsertCreatedDate, insertedApplication1CreatedDate)

	application1ToInsertUpdatedDate := application1.UpdatedDate.Format(time.RFC3339)
	insertedApplication2UpdatedDate := results1Applications[0].UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, application1ToInsertUpdatedDate, insertedApplication2UpdatedDate)

	assert.Equal(t, application2Id, results1Applications[1].ID)
	assert.Equal(t, company2Id, *results1Applications[1].CompanyID)
	assert.Nil(t, results1Applications[1].RecruiterID)
	assert.Nil(t, results1Applications[1].JobTitle)
	assert.Equal(t, "Application2JobAdURL", *results1Applications[1].JobAdURL)
	assert.Nil(t, results1Applications[1].Country)
	assert.Nil(t, results1Applications[1].Area)
	assert.Equal(t, models.RemoteStatusTypeHybrid, results1Applications[1].RemoteStatusType.String())
	assert.Nil(t, results1Applications[1].WeekdaysInOffice)
	assert.Nil(t, results1Applications[1].EstimatedCycleTime)
	assert.Nil(t, results1Applications[1].EstimatedCommuteTime)
	assert.Nil(t, results1Applications[1].ApplicationDate)

	application2ToInsertCreatedDate := application2.CreatedDate.Format(time.RFC3339)
	insertedApplication2CreatedDate := results1Applications[1].CreatedDate.Format(time.RFC3339)
	assert.Equal(t, application2ToInsertCreatedDate, insertedApplication2CreatedDate)

	assert.Nil(t, results1Applications[1].UpdatedDate)

	assert.Equal(t, application3Id, results1Applications[2].ID)
	assert.Nil(t, results1Applications[2].CompanyID)
	assert.Equal(t, company2Id, *results1Applications[2].RecruiterID)
	assert.Equal(t, "Application3JobTitle", *results1Applications[2].JobTitle)
	assert.Equal(t, models.RemoteStatusTypeOffice, results1Applications[2].RemoteStatusType.String())

	application3ToInsertCreatedDate := application3.CreatedDate.Format(time.RFC3339)
	insertedApplication3CreatedDate := results1Applications[2].CreatedDate.Format(time.RFC3339)
	assert.Equal(t, application3ToInsertCreatedDate, insertedApplication3CreatedDate)

	assert.Equal(t, company1Id, results[2].ID)
	assert.Equal(t, "company1Name", results[2].Name)
	assert.Equal(t, company1ToInsert.CompanyType, results[2].CompanyType)
	assert.Equal(t, "company 1 notes", *results[2].Notes)

	company1ToInsertLastContact := company1ToInsert.LastContact.Format(time.RFC3339)
	insertedCompany1LastContact := insertedCompany1.LastContact.Format(time.RFC3339)
	assert.Equal(t, company1ToInsertLastContact, insertedCompany1LastContact)

	company1ToInsertCreatedDate := company1ToInsert.CreatedDate.Format(time.RFC3339)
	insertedCompany1CreatedDate := insertedCompany1.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, company1ToInsertCreatedDate, insertedCompany1CreatedDate)

	company1ToInsertUpdatedDate := company1ToInsert.UpdatedDate.Format(time.RFC3339)
	insertedCompany1UpdatedDate := insertedCompany1.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, company1ToInsertUpdatedDate, insertedCompany1UpdatedDate)
	assert.Nil(t, results[2].Applications)

}

func TestGetAllCompanies_ShouldReturnNilIfNoCompaniesInDatabase(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsNone(t *testing.T) {
	companyService, applicationRepository := setupCompanyService(t)

	// insert companies
	company1Id := uuid.New()
	company1ToInsert := &models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := companyService.CreateCompany(company1ToInsert)
	assert.NoError(t, err)

	company2Id := uuid.New()
	company2ToInsert := &models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = companyService.CreateCompany(company2ToInsert)
	assert.NoError(t, err)

	company3Id := uuid.New()
	company3ToInsert := &models.CreateCompany{
		ID:          &company3Id,
		Name:        "company3Name",
		CompanyType: models.CompanyTypeRecruiter,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = companyService.CreateCompany(company3ToInsert)
	assert.NoError(t, err)

	// insert applications

	application1Id := uuid.New()
	application1 := models.CreateApplication{
		ID:               &application1Id,
		CompanyID:        &company2Id,
		RecruiterID:      &company3Id,
		JobTitle:         testutil.ToPtr("Application1JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
	}
	_, err = applicationRepository.Create(&application1)
	assert.NoError(t, err)

	application2Id := uuid.New()
	application2 := models.CreateApplication{
		ID:               &application2Id,
		CompanyID:        &company2Id,
		JobTitle:         testutil.ToPtr("Application2JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}
	_, err = applicationRepository.Create(&application2)
	assert.NoError(t, err)

	application3Id := uuid.New()
	application3 := models.CreateApplication{
		ID:               &application3Id,
		RecruiterID:      &company2Id,
		JobAdURL:         testutil.ToPtr("Application3JobAdUrl"),
		RemoteStatusType: models.RemoteStatusTypeOffice,
	}
	_, err = applicationRepository.Create(&application3)
	assert.NoError(t, err)

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeNone)
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

func TestGetAllCompanies_ShouldReturnCompaniesWithApplicationIDsIfIncludeApplicationsIsIDs(t *testing.T) {
	companyService, applicationRepository := setupCompanyService(t)

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
		ID:               &application1Id,
		CompanyID:        &company2Id,
		RecruiterID:      &company3Id,
		JobTitle:         testutil.ToPtr("Application1JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
		CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err = applicationRepository.Create(&application1)
	assert.NoError(t, err)

	application2Id := uuid.New()
	application2 := models.CreateApplication{
		ID:               &application2Id,
		CompanyID:        &company2Id,
		JobTitle:         testutil.ToPtr("Application2JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeOffice,
		CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err = applicationRepository.Create(&application2)
	assert.NoError(t, err)

	application3Id := uuid.New()
	application3 := models.CreateApplication{
		ID:               &application3Id,
		RecruiterID:      &company2Id,
		JobAdURL:         testutil.ToPtr("Application3JobAdUrl"),
		RemoteStatusType: models.RemoteStatusTypeOffice,
		CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err = applicationRepository.Create(&application3)
	assert.NoError(t, err)

	// get all companies

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeIDs)
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
	assert.Nil(t, results1Applications[1].JobTitle)
	assert.Nil(t, results1Applications[1].JobAdURL)
	assert.Nil(t, results1Applications[1].Country)
	assert.Nil(t, results1Applications[1].Area)
	assert.Nil(t, results1Applications[1].RemoteStatusType)
	assert.Nil(t, results1Applications[1].WeekdaysInOffice)
	assert.Nil(t, results1Applications[1].EstimatedCycleTime)
	assert.Nil(t, results1Applications[1].EstimatedCommuteTime)
	assert.Nil(t, results1Applications[1].ApplicationDate)
	assert.Nil(t, results1Applications[1].CreatedDate)
	assert.Nil(t, results1Applications[1].UpdatedDate)

	assert.Equal(t, application3Id, results1Applications[2].ID)
	assert.Nil(t, results1Applications[2].CompanyID)
	assert.Equal(t, company2Id, *results1Applications[2].RecruiterID)
	assert.Nil(t, results1Applications[2].JobTitle)
	assert.Nil(t, results1Applications[2].JobAdURL)
	assert.Nil(t, results1Applications[2].Country)
	assert.Nil(t, results1Applications[2].Area)
	assert.Nil(t, results1Applications[2].RemoteStatusType)
	assert.Nil(t, results1Applications[2].WeekdaysInOffice)
	assert.Nil(t, results1Applications[2].EstimatedCycleTime)
	assert.Nil(t, results1Applications[2].EstimatedCommuteTime)
	assert.Nil(t, results1Applications[2].ApplicationDate)
	assert.Nil(t, results1Applications[2].CreatedDate)
	assert.Nil(t, results1Applications[2].UpdatedDate)

	assert.Equal(t, company1Id, results[2].ID)
	assert.Nil(t, results[2].Applications)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsIDsAndThereAreNoApplications(t *testing.T) {
	companyService, _ := setupCompanyService(t)

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

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeIDs)
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

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsAllAndThereAreNoApplications(t *testing.T) {
	companyService, _ := setupCompanyService(t)

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

	results, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeAll)
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

func TestGetAllCompanies_ShouldReturnASingleEntryWhenCompanyIDAndRecruiterIDAreTheSame(t *testing.T) {
	companyService, applicationRepository := setupCompanyService(t)

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

	idResults, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeIDs)
	assert.NoError(t, err)

	assert.NotNil(t, idResults)
	assert.Len(t, idResults, 1)

	assert.Equal(t, companyId, idResults[0].ID)
	assert.Len(t, *idResults[0].Applications, 1)

	allResults, err := companyService.GetAllCompanies(models.IncludeExtraDataTypeAll)
	assert.NoError(t, err)

	assert.NotNil(t, allResults)
	assert.Len(t, allResults, 1)

	assert.Equal(t, companyId, allResults[0].ID)
	assert.Len(t, *allResults[0].Applications, 1)
}

// -------- UpdateCompany tests: --------

func TestUpdateCompany_ShouldWork(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	// insert a company:
	id := uuid.New()
	notes := "Notes about an AB"
	lastContact := time.Now().AddDate(0, 0, -3)
	createdDate := time.Now().AddDate(0, 0, -2)
	updatedDate := time.Now().AddDate(0, 0, -1)

	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "Some Stockholm-based AB",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	insertedCompany, err := companyService.CreateCompany(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	// update a company:

	nameToUpdate := "Updated Name"
	var companyTypeToUpdate models.CompanyType = models.CompanyTypeConsultancy
	notesToUpdate := "Updated Notes"
	lastContactToUpdate := time.Now().AddDate(0, 1, 0)

	updateModel := models.UpdateCompany{
		ID:          id,
		Name:        &nameToUpdate,
		CompanyType: &companyTypeToUpdate,
		Notes:       &notesToUpdate,
		LastContact: &lastContactToUpdate,
	}

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	err = companyService.UpdateCompany(&updateModel)
	assert.NoError(t, err)

	// get the company to ensure that the changes have been applied.
	retrievedCompany, err := companyService.GetCompanyById(&id)
	assert.NoError(t, err)

	assert.NotNil(t, retrievedCompany)
	assert.Equal(t, id, retrievedCompany.ID)
	assert.Equal(t, nameToUpdate, retrievedCompany.Name)
	assert.Equal(t, companyTypeToUpdate, retrievedCompany.CompanyType)

	updatedLastContact := lastContactToUpdate.Format(time.RFC3339)
	retrievedLastContact := retrievedCompany.LastContact.Format(time.RFC3339)
	assert.Equal(t, updatedLastContact, retrievedLastContact)

	insertedCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	retrievedCreatedDate := retrievedCompany.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, insertedCreatedDate, retrievedCreatedDate)

	retrievedUpdatedDate := retrievedCompany.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, retrievedUpdatedDate)
}

func TestUpdateCompany_ShouldNotReturnErrorIfIdToUpdateDoesNotExist(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	nameToUpdate := "Updated Name"
	var companyTypeToUpdate models.CompanyType = models.CompanyTypeConsultancy
	notesToUpdate := "Updated Notes"
	lastContactToUpdate := time.Now().AddDate(0, 1, 0)

	updateModel := models.UpdateCompany{
		ID:          uuid.New(),
		Name:        &nameToUpdate,
		CompanyType: &companyTypeToUpdate,
		Notes:       &notesToUpdate,
		LastContact: &lastContactToUpdate,
	}

	err := companyService.UpdateCompany(&updateModel)
	assert.NoError(t, err)
}

// -------- DeleteCompany tests: --------

func TestDeleteCompany_ShouldWork(t *testing.T) {
	companyService, _ := setupCompanyService(t)

	// create a company:

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now()
	createdDate := time.Now().AddDate(0, 0, 0)
	updatedDate := time.Now().AddDate(0, 0, 0)

	companyToInsert := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
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
	companyService, _ := setupCompanyService(t)

	id := uuid.New()

	err := companyService.DeleteCompany(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Company does not exist. ID: "+id.String(), err.Error())
}
