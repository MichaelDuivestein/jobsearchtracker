package models

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreateCompany.Validate tests: --------

func TestCreateCompanyValidate_ShouldReturnNilIfCompanyIsValid(t *testing.T) {
	company := CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Something should be here",
		CompanyType: CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	err := company.Validate()
	assert.NoError(t, err)
}

func TestCreateCompanyValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	company := CreateCompany{
		Name:        "Something should be here",
		CompanyType: CompanyTypeRecruiter,
	}

	err := company.Validate()
	assert.NoError(t, err)
}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnEmptyName(t *testing.T) {
	company := CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "",
		CompanyType: CompanyTypeRecruiter,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	err := company.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'Name': company name is empty", validationError.Error())
}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnEmptyCompanyType(t *testing.T) {
	company := CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "A random person",
		Notes:       testutil.ToPtr("More stuff"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	err := company.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", validationError.Error())
}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnInvalidCompanyType(t *testing.T) {
	company := CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Jan Janssen",
		CompanyType: "Nothing",
		Notes:       testutil.ToPtr("Noted"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
	}

	err := company.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'CompanyType': company type is invalid", validationError.Error())
}

func TestCreateCompanyValidate_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	company := CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Pick one",
		CompanyType: CompanyTypeEmployer,
		Notes:       testutil.ToPtr("some notes"),
		UpdatedDate: &time.Time{},
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -5, 0)),
	}

	err := company.Validate()
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'UpdatedDate': updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

// -------- CompanyType.IsValid tests: --------

func TestCompanyTypeIsValid_ShouldReturnTrue(t *testing.T) {
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
