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

// -------- CreatePerson.Validate tests: --------

func TestCreatePersonValidate_ShouldReturnNilIfPersonIsValid(t *testing.T) {
	person := CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Random Name",
		PersonType:  PersonTypeOther,
		Email:       testutil.ToPtr("Email Address"),
		Phone:       testutil.ToPtr("84323445"),
		Notes:       testutil.ToPtr("Noted"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	err := person.Validate()
	assert.NoError(t, err)
}

func TestCreatePersonValidate_ShouldReturnNilIfOnlyRequiredFieldsExist(t *testing.T) {
	person := CreatePerson{
		Name:       "Name Names",
		PersonType: PersonTypeCEO,
	}
	err := person.Validate()
	assert.NoError(t, err)
}

func TestCreatePersonValidate_ShouldReturnValidationErrorOnEmptyName(t *testing.T) {
	person := CreatePerson{
		Name:       "",
		PersonType: PersonTypeUnknown,
	}
	err := person.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'Name': person name is empty", validationError.Error())
}

func TestCreatePersonValidate_ShouldReturnValidationErrorOnEmptyPersonType(t *testing.T) {
	person := CreatePerson{
		Name: "Name Names",
	}
	err := person.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'PersonType': person type is invalid", validationError.Error())
}

func TestCreatePersonValidate_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	person := CreatePerson{
		PersonType:  PersonTypeJobAdvertiser,
		Name:        "Something here",
		UpdatedDate: &time.Time{},
	}
	err := person.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t,
		"validation error on field 'UpdatedDate': updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

// -------- PersonType.IsValid tests: --------

func TestPersonTypeIsValid_ShouldReturnTrue(t *testing.T) {
	ceo := PersonType(PersonTypeCEO)
	assert.True(t, ceo.IsValid())

	cto := PersonType(PersonTypeCTO)
	assert.True(t, cto.IsValid())

	developer := PersonType(PersonTypeDeveloper)
	assert.True(t, developer.IsValid())

	externalRecruiter := PersonType(PersonTypeExternalRecruiter)
	assert.True(t, externalRecruiter.IsValid())

	internalRecruiter := PersonType(PersonTypeInternalRecruiter)
	assert.True(t, internalRecruiter.IsValid())

	hr := PersonType(PersonTypeHR)
	assert.True(t, hr.IsValid())

	jobAdvertiser := PersonType(PersonTypeJobAdvertiser)
	assert.True(t, jobAdvertiser.IsValid())

	jobContact := PersonType(PersonTypeJobContact)
	assert.True(t, jobContact.IsValid())

	other := PersonType(PersonTypeOther)
	assert.True(t, other.IsValid())

	unknown := PersonType(PersonTypeUnknown)
	assert.True(t, unknown.IsValid())

}

func TestPersonType_IsValid_ShouldReturnFalseOnInvalidPersonType(t *testing.T) {
	empty := PersonType("")
	assert.False(t, empty.IsValid())

	spammer := PersonType("MisTyped")
	assert.False(t, spammer.IsValid())
}
