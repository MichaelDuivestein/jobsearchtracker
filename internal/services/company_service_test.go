package services

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldReturnValidationErrorOnNilCompany(t *testing.T) {
	companyService := NewCompanyService(nil)

	nilCompany, err := companyService.CreateCompany(nil)
	assert.Nil(t, nilCompany, "company should be nil")
	assert.NotNil(t, err, "error should not be nil")

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CreateCompany is nil", err.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnEmptyName(t *testing.T) {
	companyService := NewCompanyService(nil)

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := &models.CreateCompany{
		ID:          &id,
		Name:        "",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany, "company should be nil")
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'Name': company name is empty", err.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnEmptyCompanyType(t *testing.T) {
	companyService := NewCompanyService(nil)

	id := uuid.New()
	notes := "More stuff"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := &models.CreateCompany{
		ID:          &id,
		Name:        "A random person",
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany, "company should be nil")
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", err.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnInvalidCompanyType(t *testing.T) {
	companyService := NewCompanyService(nil)

	id := uuid.New()
	notes := "Noted"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := &models.CreateCompany{
		ID:          &id,
		Name:        "Jan Janssen",
		CompanyType: "Nothing",
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany, "company should be nil")
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", err.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	companyService := NewCompanyService(nil)

	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)

	company := &models.CreateCompany{
		ID:          &id,
		Name:        "Pick one",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       &notes,
		UpdatedDate: &time.Time{},
		LastContact: &lastContact,
		CreatedDate: &createdDate,
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany, "company should be nil")
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'UpdatedDate': updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil", err.Error())
}

// -------- GetCompanyById tests: --------

func TestGetCompanyById_ShouldReturnValidationErrorIfCompanyIdIsNil(t *testing.T) {
	companyService := NewCompanyService(nil)

	company, err := companyService.GetCompanyById(nil)
	assert.NotNil(t, err, "error should not be nil")
	assert.Equal(t, "validation error on field 'company ID': companyId is required", err.Error())
	assert.Nil(t, company, "company should be nil")
}

// -------- GetCompaniesByName tests: --------
func TestGetCompaniesByName_ShouldReturnValidationErrorIfCompanyNameIsNil(t *testing.T) {
	companyService := NewCompanyService(nil)

	nilCompany, err := companyService.GetCompaniesByName(nil)
	assert.Nil(t, nilCompany)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: companyName is required", err.Error())
}
func TestGetCompaniesByName_ShouldReturnValidationErrorIfCompanyNameIsEmpty(t *testing.T) {
	companyService := NewCompanyService(nil)

	emptyName := ""
	nilCompany, err := companyService.GetCompaniesByName(&emptyName)
	assert.Nil(t, nilCompany)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: companyName is required", err.Error())
}
