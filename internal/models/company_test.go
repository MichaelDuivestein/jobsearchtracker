package models

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreateCompany.Validate tests: --------

func TestCreateCompanyValidate_ShouldReturnNilIfCompanyIsValid(t *testing.T) {
	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := CreateCompany{
		ID:          &id,
		Name:        "Something should be here",
		CompanyType: CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	err := company.Validate()
	assert.Nil(t, err, "error should be nil")
}

func TestCreateCompanyValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	company := CreateCompany{
		Name:        "Something should be here",
		CompanyType: CompanyTypeRecruiter,
	}

	err := company.Validate()
	assert.Nil(t, err, "error should be nil")
}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnEmptyName(t *testing.T) {
	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := &CreateCompany{
		ID:          &id,
		Name:        "",
		CompanyType: CompanyTypeRecruiter,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	err := company.Validate()
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'Name': company name is empty", err.Error())
}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnEmptyCompanyType(t *testing.T) {
	id := uuid.New()
	notes := "More stuff"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := &CreateCompany{
		ID:          &id,
		Name:        "A random person",
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	err := company.Validate()
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", err.Error())

}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnInvalidCompanyType(t *testing.T) {
	id := uuid.New()
	notes := "Noted"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)
	updatedDate := time.Now().AddDate(0, 0, -3)

	company := &CreateCompany{
		ID:          &id,
		Name:        "Jan Janssen",
		CompanyType: "Nothing",
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: &createdDate,
		UpdatedDate: &updatedDate,
	}

	err := company.Validate()
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", err.Error())
}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	id := uuid.New()
	notes := "some notes"
	lastContact := time.Now().AddDate(-1, 0, 0)
	createdDate := time.Now().AddDate(0, -5, 0)

	company := &CreateCompany{
		ID:          &id,
		Name:        "Pick one",
		CompanyType: CompanyTypeEmployer,
		Notes:       &notes,
		UpdatedDate: &time.Time{},
		LastContact: &lastContact,
		CreatedDate: &createdDate,
	}

	err := company.Validate()
	assert.NotNil(t, err, "error should not be nil")

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(
		t,
		"validation error on field 'UpdatedDate': updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		err.Error())
}

// -------- CompanyType.IsValid tests: --------

func TestCompanyTypeisValid_ShouldReturnTrue(t *testing.T) {
	employer := CompanyType(CompanyTypeEmployer)
	assert.True(t, employer.IsValid())

	recruiter := CompanyType(CompanyTypeRecruiter)
	assert.True(t, recruiter.IsValid())

	consultancy := CompanyType(CompanyTypeConsultancy)
	assert.True(t, consultancy.IsValid())
}

func TestCompanyTypeIsValid_ShouldReturnFalseOnInvalidCompanyType(t *testing.T) {
	empty := CompanyType("")
	assert.False(t, empty.IsValid())

	spammer := CompanyType("Spammer")
	assert.False(t, spammer.IsValid())
}
