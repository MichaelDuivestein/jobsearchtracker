package repositories_test

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"testing"
	"time"
)

func setupCompanyRepository(t *testing.T) *repositories.CompanyRepository {
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

	return companyRepository
}

// -------- Create tests: --------

func TestCreate_ShouldInsertCompany(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	insertedCompany, err := companyRepository.Create(&company)

	assert.Nil(t, err, "Error on companyRepository.Create(): '%s'.", err)
	assert.NotNil(t, insertedCompany, "inserted company is nil")

	assert.Equal(t, *company.ID, insertedCompany.ID, "insertedCompany.ID should be the same as company.ID")
	assert.Equal(t, company.Name, insertedCompany.Name, "insertedCompany.Name should be the same as company.Name")
	assert.Equal(t, company.CompanyType, insertedCompany.CompanyType, "insertedCompany.CompanyType should be the same as company.CompanyType")
	assert.Equal(t, company.Notes, insertedCompany.Notes, "insertedCompany.Notes should be the same as company.Notes")

	insertedCompanyLastContact := insertedCompany.LastContact.Format(time.RFC3339)
	companyToInsertLastContact := company.LastContact.Format(time.RFC3339)
	assert.Equal(t, companyToInsertLastContact, insertedCompanyLastContact, "insertedCompany.LastContact should be the same as company.LastContact")

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := company.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, insertedCompanyCreatedDate, "insertedCompany.CreatedDate should be the same as company.CreatedDate")

	insertedCompanyUpdatedDate := insertedCompany.UpdatedDate.Format(time.RFC3339)
	companyToInsertUpdatedDate := company.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertUpdatedDate, insertedCompanyUpdatedDate, "insertedCompany.UpdatedDate should be the same as company.UpdatedDate")
}

func TestCreate_ShouldInsertCompanyWithMinimumRequiredFields(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	company := models.CreateCompany{
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
	}

	createdDateApproximation := time.Now().Format(time.RFC3339)
	insertedCompany, err := companyRepository.Create(&company)

	assert.Nil(t, err, "Error on companyRepository.Create(): '%s'.", err)
	assert.NotNil(t, insertedCompany, "inserted company is nil")

	assert.Equal(t, company.Name, insertedCompany.Name, "insertedCompany.Name should be the same as company.Name")
	assert.Equal(t, company.CompanyType, insertedCompany.CompanyType, "insertedCompany.CompanyType should be the same as company.CompanyType")
	assert.Nil(t, insertedCompany.Notes, "inserted company.Notes should be nil, but got '%s'", insertedCompany.Notes)
	assert.Nil(t, insertedCompany.LastContact, "inserted company.LastContact should be nil, but got '%s'", insertedCompany.LastContact)

	insertedCompanyCreatedDate := insertedCompany.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, insertedCompanyCreatedDate, "insertedCompany.CreatedDate should be the same as '%s'", createdDateApproximation)

	assert.Nil(t, insertedCompany.UpdatedDate, "inserted company.UpdatedDate should be nil, but got '%s'", insertedCompany.UpdatedDate)
}

func TestCreate_ShouldReturnConflictErrorOnDuplicateId(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now()
	createdDate := time.Now().AddDate(0, -5, 0)

	firstCompany := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
	}

	firstInsertedCompany, err := companyRepository.Create(&firstCompany)

	assert.Nil(t, err, "error on companyRepository.Create(): '%s'.", err)
	assert.NotNil(t, firstInsertedCompany, "inserted company is nil")

	assert.Equal(t, firstInsertedCompany.ID, id, "insertedCompany.ID should be '%s'", id)

	secondCompany := models.CreateCompany{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
	}

	shouldBeNil, err := companyRepository.Create(&secondCompany)
	assert.Nil(t, shouldBeNil, "expected returned company to be nil")
	assert.NotNil(t, err, "expected error but got nil")

	var conflictError *internalErrors.ConflictError
	assert.True(t, errors.As(err, &conflictError))
	assert.Equal(t, "conflict error on insert: companyID already exists in database: '"+id.String()+"'", err.Error(), "returned error is not expected error")
}

// -------- GetById tests: --------

func TestGetById_ShouldGetCompany(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

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

	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.Nil(t, err, "Error on companyRepository.Create(): '%s'.", err)
	assert.NotNil(t, insertedCompany, "inserted company is nil")

	retrievedCompany, err := companyRepository.GetById(&id)
	assert.Nil(t, err, "Error on companyRepository.GetById(): '%s'.", err)
	assert.NotNil(t, retrievedCompany, "retrieved company is nil")

	assert.Equal(t, *companyToInsert.ID, retrievedCompany.ID, "retrievedCompany.ID should be the same as companyToInsert.ID")
	assert.Equal(t, companyToInsert.Name, retrievedCompany.Name, "retrievedCompany.Name should be the same as companyToInsert.Name")
	assert.Equal(t, companyToInsert.CompanyType, retrievedCompany.CompanyType, "retrievedCompany.CompanyType should be the same as companyToInsert.CompanyType")
	assert.Equal(t, companyToInsert.Notes, retrievedCompany.Notes, "retrievedCompany.Notes should be the same as companyToInsert.Notes")

	retrievedCompanyLastContact := retrievedCompany.LastContact.Format(time.RFC3339)
	companyToInsertLastContact := companyToInsert.LastContact.Format(time.RFC3339)
	assert.Equal(t, companyToInsertLastContact, retrievedCompanyLastContact, "retrievedCompany.LastContact should be the same as companyToInsert.LastContact")

	retrievedCompanyCreatedDate := retrievedCompany.CreatedDate.Format(time.RFC3339)
	companyToInsertCreatedDate := companyToInsert.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertCreatedDate, retrievedCompanyCreatedDate, "retrievedCompany.CreatedDate should be the same as companyToInsert.CreatedDate")

	retrievedCompanyUpdatedDate := retrievedCompany.UpdatedDate.Format(time.RFC3339)
	companyToInsertUpdatedDate := companyToInsert.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, companyToInsertUpdatedDate, retrievedCompanyUpdatedDate, "retrievedCompany.UpdatedDate should be the same as companyToInsert.UpdatedDate")
}

func TestGetById_ShouldReturnErrorIfIdIsNil(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	response, err := companyRepository.GetById(nil)
	assert.Nil(t, response, "Response should be nil")
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "validation error on field 'ID': ID is nil", err.Error(), "Wrong error returned")
}

func TestGetById_ShouldReturnErrorIfIdDoesNotExist(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	id := uuid.New()

	response, err := companyRepository.GetById(&id)
	assert.Nil(t, response, "response should be nil")
	assert.NotNil(t, err, "error should ID: '"+id.String()+"'", err.Error(), "Wrong error returned")
}
