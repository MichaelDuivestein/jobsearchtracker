package services

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldReturnValidationErrorOnNilCompany(t *testing.T) {
	companyService := NewCompanyService(nil)

	nilCompany, err := companyService.CreateCompany(nil)
	assert.Nil(t, nilCompany)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CreateCompany is nil", validationError.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnEmptyName(t *testing.T) {
	companyService := NewCompanyService(nil)

	company := &models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "",
		CompanyType: models.CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'Name': company name is empty", validationError.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnEmptyCompanyType(t *testing.T) {
	companyService := NewCompanyService(nil)

	company := &models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "A random person",
		Notes:       testutil.ToPtr("More stuff"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", validationError.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnInvalidCompanyType(t *testing.T) {
	companyService := NewCompanyService(nil)

	company := &models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Jan Janssen",
		CompanyType: "Nothing",
		Notes:       testutil.ToPtr("Noted"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", validationError.Error())
}

func TestCreateCompany_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	companyService := NewCompanyService(nil)

	company := &models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Pick one",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		UpdatedDate: &time.Time{},
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
	}

	nilCompany, err := companyService.CreateCompany(company)
	assert.Nil(t, nilCompany)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'UpdatedDate': updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

// -------- GetCompanyById tests: --------

func TestGetCompanyById_ShouldReturnValidationErrorIfCompanyIdIsNil(t *testing.T) {
	companyService := NewCompanyService(nil)

	company, err := companyService.GetCompanyById(nil)
	assert.Nil(t, company)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'company ID': companyId is required", validationError.Error())
}

// -------- GetCompaniesByName tests: --------
func TestGetCompaniesByName_ShouldReturnValidationErrorIfCompanyNameIsNil(t *testing.T) {
	companyService := NewCompanyService(nil)

	nilCompany, err := companyService.GetCompaniesByName(nil)
	assert.Nil(t, nilCompany)
	assert.Error(t, err)
	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: companyName is required", validationError.Error())
}

func TestGetCompaniesByName_ShouldReturnValidationErrorIfCompanyNameIsEmpty(t *testing.T) {
	companyService := NewCompanyService(nil)

	nilCompany, err := companyService.GetCompaniesByName(testutil.ToPtr(""))
	assert.Nil(t, nilCompany)
	assert.Error(t, err)
	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: companyName is required", validationError.Error())
}

// -------- UpdateCompany tests: --------

func TestUpdateCompany_ShouldReturnValidationErrorIfCompanyIsNil(t *testing.T) {
	companyService := NewCompanyService(nil)

	err := companyService.UpdateCompany(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: UpdateCompany model is nil", validationError.Error())
}

func TestUpdateCompany_ShouldReturnValidationErrorIfCompanyContainsNothingToUpdate(t *testing.T) {
	companyService := NewCompanyService(nil)

	companyToUpdate := &models.UpdateCompany{
		ID: uuid.New(),
	}

	err := companyService.UpdateCompany(companyToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

// -------- DeleteCompany tests: --------

func TestDeleteCompany_ShouldReturnValidationErrorIfCompanyIdIsNil(t *testing.T) {
	companyService := NewCompanyService(nil)

	err := companyService.DeleteCompany(nil)
	assert.Error(t, err)
	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'company ID': companyId is required", validationError.Error())
}
