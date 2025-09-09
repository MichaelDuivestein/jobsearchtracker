package services_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/services"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupCompanyService(t *testing.T) *services.CompanyService {
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

	return companyService
}

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldWork(t *testing.T) {
	companyService := setupCompanyService(t)

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
	assert.Nil(t, err, "Failed to create company: '%s'", err)
	assert.NotNil(t, insertedCompany, "CreateCompany should return a company")

	assert.Equal(t, *companyToInsert.ID, id)
	assert.Equal(t, companyToInsert.Name, insertedCompany.Name, "insertedCompany.Name should be the same as companyToInsert.Name")
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType, "insertedCompany.CompanyType should be the same as companyToInsert.CompanyType")
	assert.Equal(t, companyToInsert.Notes, insertedCompany.Notes, "insertedCompany.Notes should be the same as companyToInsert.Notes")

	insertedCompanyLastContact := insertedCompany.LastContact.Format(time.RFC3339)
	companyToInsertLastContact := companyToInsert.LastContact.Format(time.RFC3339)
	assert.Equal(t, companyToInsertLastContact, insertedCompanyLastContact, "insertedCompany.LastContact should be the same as companyToInsert.LastContact")

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := companyToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, insertedCompanyCreatedDate, "insertedCompany.CreatedDate should be the same as companyToInsert.CreatedDate")

	insertedCompanyUpdatedDate := insertedCompany.UpdatedDate.Format(time.RFC3339)
	companyToInsertUpdatedDate := companyToInsert.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertUpdatedDate, insertedCompanyUpdatedDate, "insertedCompany.UpdatedDate should be the same as companyToInsert.UpdatedDate")
}

func TestCreateCompany_ShouldHandleEmptyFields(t *testing.T) {
	companyService := setupCompanyService(t)

	companyToInsert := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
	}

	insertedDateApproximation := time.Now().Format(time.RFC3339)
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)

	assert.Nil(t, err, "Failed to create company: '%s'", err)
	assert.NotNil(t, insertedCompany, "CreateCompany should return a company")

	assert.Equal(t, companyToInsert.Name, insertedCompany.Name, "insertedCompany.Name should be the same as company.Name")
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType, "insertedCompany.CompanyType should be the same as company.CompanyType")
	assert.Nil(t, insertedCompany.Notes, "inserted company.Notes should be nil, but got '%s'", insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact, "inserted company.LastContact should be nil, but got '%s'", insertedCompany.LastContact)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, insertedDateApproximation, insertedCompanyCreatedDate, "insertedCompany.CreatedDate should be the same as '%s'", insertedDateApproximation)

	assert.Nil(t, insertedCompany.UpdatedDate, "inserted company.UpdatedDate should be nil, but got '%s'", insertedCompany.UpdatedDate)
}

func TestCreateCompany_ShouldHandleUnsetCreatedDate(t *testing.T) {
	companyService := setupCompanyService(t)

	companyToInsert := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: &time.Time{},
	}

	insertedDateApproximation := time.Now().Format(time.RFC3339)
	insertedCompany, err := companyService.CreateCompany(&companyToInsert)

	assert.Nil(t, err, "Failed to create company: '%s'", err)
	assert.NotNil(t, insertedCompany, "CreateCompany should return a company")

	assert.Equal(t, companyToInsert.Name, insertedCompany.Name, "insertedCompany.Name should be the same as company.Name")
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType, "insertedCompany.CompanyType should be the same as company.CompanyType")
	assert.Nil(t, insertedCompany.Notes, "inserted company.Notes should be nil, but got '%s'", insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact, "inserted company.LastContact should be nil, but got '%s'", insertedCompany.LastContact)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, insertedDateApproximation, insertedCompanyCreatedDate, "insertedCompany.CreatedDate should be the same as '%s'", insertedDateApproximation)

	assert.Nil(t, insertedCompany.UpdatedDate, "inserted company.UpdatedDate should be nil, but got '%s'", insertedCompany.UpdatedDate)
}

func TestCreateCompany_ShouldSetUnsetLastContactToCreatedDate(t *testing.T) {
	companyService := setupCompanyService(t)

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

	assert.Nil(t, err, "Failed to create company: '%s'", err)
	assert.NotNil(t, insertedCompany, "CreateCompany should return a company")

	assert.Equal(t, companyToInsert.Name, insertedCompany.Name, "insertedCompany.Name should be the same as company.Name")
	assert.Equal(t, companyToInsert.CompanyType, insertedCompany.CompanyType, "insertedCompany.CompanyType should be the same as company.CompanyType")
	assert.Nil(t, insertedCompany.Notes, "inserted company.Notes should be nil, but got '%s'", insertedCompany.Notes)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := companyToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, insertedCompanyCreatedDate, "insertedCompany.CreatedDate should be the same as companyToInsert.CreatedDate")

	insertedCompanyLastContact := insertedCompany.LastContact.Format(time.RFC3339)
	assert.Equal(t, insertedCompanyCreatedDate, insertedCompanyLastContact, "insertedCompany.LastContact should be the same as companyToInsert.CreatedDate")

	assert.Nil(t, insertedCompany.UpdatedDate, "inserted company.UpdatedDate should be nil, but got '%s'", insertedCompany.UpdatedDate)
}

// -------- GetCompanyById tests: --------

func TestGetCompanyById_ShouldWork(t *testing.T) {
	companyService := setupCompanyService(t)

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
	assert.Nil(t, err, "Failed to create company: '%s'", err)

	retrievedCompany, err := companyService.GetCompanyById(&id)
	assert.Nil(t, err, "Failed to create retrieve: '%s'", err)
	assert.NotNil(t, retrievedCompany, "retrieved company is nil")

	assert.Equal(t, *companyToInsert.ID, retrievedCompany.ID, "retrievedCompany.ID should match companyToInsert.ID")
	assert.Equal(t, companyToInsert.Name, retrievedCompany.Name, "retrievedCompany.Name should match companyToInsert.Name")
	assert.Equal(t, companyToInsert.CompanyType, retrievedCompany.CompanyType, "retrievedCompany.CompanyType should match companyToInsert.CompanyType")
	assert.Equal(t, companyToInsert.Notes, retrievedCompany.Notes, "retrievedCompany.Notes should match companyToInsert.Notes")

	retrievedCompanyLastContact := retrievedCompany.LastContact.Format(time.RFC3339)
	companyToInsertLastContact := companyToInsert.LastContact.Format(time.RFC3339)
	assert.Equal(t, companyToInsertLastContact, retrievedCompanyLastContact, "retrievedCompany.CreatedDate should be the same as companyToInsert.CreatedDate")

	retrievedCompanyCreatedDate := retrievedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := companyToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, retrievedCompanyCreatedDate, "retrievedCompany.CreatedDate should be the same as companyToInsert.CreatedDate")

	retrievedCompanyUpdatedDate := retrievedCompany.UpdatedDate.Format(time.RFC3339)
	companyToInsertUpdatedDate := companyToInsert.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertUpdatedDate, retrievedCompanyUpdatedDate, "retrievedCompany.CreatedDate should be the same as companyToInsert.CreatedDate")
}

func TestGetCompanyById_ShouldReturnNotFoundErrorForAnIdThatDoesNotExist(t *testing.T) {
	companyService := setupCompanyService(t)

	nonExistingId := uuid.New()
	retrievedCompany, err := companyService.GetCompanyById(&nonExistingId)
	assert.NotNil(t, err, "Error should not be nil", err)
	assert.Nil(t, retrievedCompany, "retrieved company should be nil")

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+nonExistingId.String()+"'", err.Error())

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
	assert.Nil(t, err, "Failed to create company: '%s'", err)

	retrievedCompany, err = companyService.GetCompanyById(&nonExistingId)
	assert.NotNil(t, err, "Error should not be nil", err)
	assert.Nil(t, retrievedCompany, "retrieved company should be nil")

	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+nonExistingId.String()+"'", err.Error())
}

// -------- GetCompaniesByName tests: --------

func TestGetCompaniesByName_ShouldReturnASingleCompany(t *testing.T) {
	companyService := setupCompanyService(t)

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
	assert.Equal(t, 1, len(companies))

	assert.Equal(t, id2, companies[0].ID)
}

func TestGetCompaniesByName_ShouldReturnMultipleCompanies(t *testing.T) {
	companyService := setupCompanyService(t)

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
	assert.Equal(t, 2, len(companies))

	assert.Equal(t, id1, companies[1].ID)
	assert.Equal(t, id3, companies[0].ID)
}

func TestGetCompaniesByName_ShouldReturnNotFoundErrorIfNoNamesMatch(t *testing.T) {
	companyService := setupCompanyService(t)

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
	companyService := setupCompanyService(t)

	company1Id := uuid.New()
	company1Notes := "some notes"
	company1LastContact := time.Now().AddDate(-1, 0, 0)
	company1CreatedDate := time.Now().AddDate(0, -5, 0)
	company1UpdatedDate := time.Now().AddDate(0, 0, -3)

	company1ToInsert := models.CreateCompany{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       &company1Notes,
		LastContact: &company1LastContact,
		CreatedDate: &company1CreatedDate,
		UpdatedDate: &company1UpdatedDate,
	}

	insertedCompany1, err := companyService.CreateCompany(&company1ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2Id := uuid.New()
	company2Notes := "some notes"
	company2LastContact := time.Now().AddDate(-1, 0, 0)
	company2CreatedDate := time.Now().AddDate(0, -4, 22)
	company2UpdatedDate := time.Now().AddDate(0, 0, -3)

	company2ToInsert := models.CreateCompany{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       &company2Notes,
		LastContact: &company2LastContact,
		CreatedDate: &company2CreatedDate,
		UpdatedDate: &company2UpdatedDate,
	}
	insertedCompany2, err := companyService.CreateCompany(&company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	results, err := companyService.GetAllCompanies()
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Equal(t, 2, len(results))

	assert.Equal(t, company2Id, results[0].ID)
	assert.Equal(t, company1Id, results[1].ID)
}

func TestGetAllCompanies_ShouldReturnNilIfNoCompaniesInDatabase(t *testing.T) {
	companyService := setupCompanyService(t)

	results, err := companyService.GetAllCompanies()
	assert.NoError(t, err)
	assert.Nil(t, results)
}

// -------- UpdateCompany tests: --------
func TestUpdateCompany_ShouldWork(t *testing.T) {
	companyService := setupCompanyService(t)

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
	assert.Nil(t, err)
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
	companyService := setupCompanyService(t)

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
