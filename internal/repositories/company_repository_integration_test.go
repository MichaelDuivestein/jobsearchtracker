package repositories_test

import (
	"errors"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func TestCreate_ShouldReturnConflictErrorOnDuplicateCompanyId(t *testing.T) {
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
	assert.Equal(t,
		"conflict error on insert: ID already exists in database: '"+id.String()+"'",
		err.Error(),
		"returned error is not expected error")
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

func TestGetById_ShouldReturnErrorIfCompanyIDIsNil(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	response, err := companyRepository.GetById(nil)
	assert.Nil(t, response, "Response should be nil")
	assert.NotNil(t, err, "Error should not be nil")
	assert.Equal(t, "validation error on field 'ID': ID is nil", err.Error(), "Wrong error returned")
}

func TestGetById_ShouldReturnErrorIfCompanyIDDoesNotExist(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	id := uuid.New()

	response, err := companyRepository.GetById(&id)
	assert.Nil(t, response, "response should be nil")
	assert.NotNil(t, err, err.Error(), "Wrong error returned")
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error(), "Wrong error returned")
}

// -------- GetAllByName tests: --------

func TestGetAllByName_ShouldReturnCompany(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

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
	assert.Equal(t, 1, len(retrievedCompanies))

	assert.Equal(t, "Company Bee", retrievedCompanies[0].Name)
}

func TestGetAllByName_ShouldReturnValidationErrorIfCompanyNameIsNil(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	retrievedCompanies, err := companyRepository.GetAllByName(nil)
	assert.Nil(t, retrievedCompanies)
	assert.NotNil(t, err)
	assert.Equal(t, "validation error: name is nil", err.Error())
}

func TestGetAllByName_ShouldReturnNotFoundErrorIfCompanyNameDoesNotExist(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	name := "Doesnt Exist"

	company, err := companyRepository.GetAllByName(&name)
	assert.Nil(t, company)
	assert.NotNil(t, err)
	assert.Equal(t, "error: object not found: Name: '"+name+"'", err.Error())
}

func TestGetAllByName_ShouldReturnMultipleCompaniesWithSameName(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	// insert some companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Some Name AB",
		CompanyType: models.CompanyTypeRecruiter,
	}
	insertedCompany1, err := companyRepository.Create(&company1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Brand AB",
		CompanyType: models.CompanyTypeEmployer,
	}
	insertedCompany2, err := companyRepository.Create(&company2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3ID := uuid.New()
	company3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "Another Company",
		CompanyType: models.CompanyTypeEmployer,
	}
	insertedCompany3, err := companyRepository.Create(&company3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// get humans with name Frank John
	ab := "ab"

	retrievedCompanies, err := companyRepository.GetAllByName(&ab)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompanies)
	assert.Equal(t, 2, len(retrievedCompanies))

	foundCompany1 := retrievedCompanies[0]
	assert.Equal(t, insertedCompany2.ID, foundCompany1.ID)

	foundCompany2 := retrievedCompanies[1]
	assert.Equal(t, insertedCompany1.ID, foundCompany2.ID)
}

func TestGetAllByName_ShouldReturnMultipleCompaniesWithSameNamePart(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	// insert some companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Some AB",
		CompanyType: models.CompanyTypeRecruiter,
	}
	insertedCompany1, err := companyRepository.Create(&company1)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany1)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Absolutely not a company name",
		CompanyType: models.CompanyTypeConsultancy,
	}
	insertedCompany2, err := companyRepository.Create(&company2)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	company3ID := uuid.New()
	company3 := models.CreateCompany{
		ID:          &company3ID,
		Name:        "Different AB",
		CompanyType: models.CompanyTypeEmployer,
	}
	insertedCompany3, err := companyRepository.Create(&company3)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany3)

	// get companies containing "ab"
	ab := "ab"

	retrievedCompanies, err := companyRepository.GetAllByName(&ab)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCompanies)
	assert.Equal(t, 3, len(retrievedCompanies))

	foundCompany1 := retrievedCompanies[0]
	assert.Equal(t, insertedCompany2.ID, foundCompany1.ID)

	foundCompany2 := retrievedCompanies[1]
	assert.Equal(t, insertedCompany3.ID, foundCompany2.ID)

	foundCompany3 := retrievedCompanies[2]
	assert.Equal(t, insertedCompany1.ID, foundCompany3.ID)
}

// -------- GetAll tests: --------

func TestGetAll_ShouldReturnAllCompanies(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

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

	insertedCompany1, err := companyRepository.Create(&company1ToInsert)
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

	insertedCompany2, err := companyRepository.Create(&company2ToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany2)

	results, err := companyRepository.GetAll()
	assert.NoError(t, err)

	assert.NotNil(t, results)
	assert.Equal(t, 2, len(results))

	assert.Equal(t, company2Id, results[0].ID)
	assert.Equal(t, company1Id, results[1].ID)
}

func TestGetAll_ShouldReturnNilIfNoCompaniesInDatabase(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	companies, err := companyRepository.GetAll()
	assert.NoError(t, err)
	assert.Nil(t, companies)
}

// -------- Delete tests: --------

func TestDelete_ShouldDeleteCompany(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(0, 0, 0)
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

	insertedCompany, err := companyRepository.Create(&companyToInsert)
	assert.NoError(t, err)
	assert.NotNil(t, insertedCompany)

	err = companyRepository.Delete(&id)
	assert.Nil(t, err, "error on companyRepository.Delete(): '%s'.", err)

	deletedCompany, err := companyRepository.GetById(&id)
	assert.NotNil(t, err)
	assert.Equal(t, "error: object not found: ID: '"+id.String()+"'", err.Error())
	assert.Nil(t, deletedCompany)
}

func TestDelete_ShouldReturnErrorIfCompanyIdIsNil(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	err := companyRepository.Delete(nil)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'ID': ID is nil", err.Error())
}

func TestDelete_ShouldReturnNotFoundErrorIfCompanyIdDoesNotExist(t *testing.T) {
	companyRepository := setupCompanyRepository(t)

	id := uuid.New()

	err := companyRepository.Delete(&id)
	assert.NotNil(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: Company does not exist. ID: "+id.String(), err.Error())
}
