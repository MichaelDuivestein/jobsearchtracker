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
