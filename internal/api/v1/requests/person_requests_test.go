package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreatePersonRequest tests: --------

func TestCreatePersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	id := uuid.New()
	email := "no email here"
	phone := "6839023748"
	notes := "Something not noteworthy"

	request := CreatePersonRequest{
		ID:         &id,
		Name:       "Nameless",
		PersonType: PersonTypeDeveloper,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestCreatePersonRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		name                 string
		personType           PersonType
		expectedErrorMessage string
	}{
		{
			"Empty Name",
			"",
			PersonTypeCTO,
			"validation error on field 'Name': Name is empty"},
		{
			"Empty PersonType",
			"Name present",
			"",
			"validation error on field 'PersonType': PersonType is invalid"},
		{
			"Invalid PersonType",
			"Name present",
			"Invalid PersonType",
			"validation error on field 'PersonType': PersonType is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			email := "name@domain.tld"
			notes := "Lots of text"
			phone := "345023485"

			request := CreatePersonRequest{
				Name:       test.name,
				PersonType: test.personType,
				Email:      &email,
				Phone:      &phone,
				Notes:      &notes,
			}

			err := request.validate()
			assert.NotNil(t, err)

			var validationErr *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationErr))

			assert.Equal(t, test.expectedErrorMessage, err.Error())
		})
	}
}

func TestCreatePersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	id := uuid.New()
	email := "email@email.email"
	phone := "34543534"
	notes := "Blah Blah"

	request := CreatePersonRequest{
		ID:         &id,
		Name:       "Jane Doe",
		PersonType: PersonTypeCEO,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, id.String(), model.ID.String())
	assert.Equal(t, request.Name, model.Name)
	assert.Equal(t, request.PersonType.String(), model.PersonType.String())
	assert.Equal(t, &email, model.Email)
	assert.Equal(t, &phone, model.Phone)
	assert.Equal(t, &notes, model.Notes)
}

func TestCreatePersonRequestToModel_ShouldConvertToModelWithNilValues(t *testing.T) {
	request := CreatePersonRequest{
		Name:       "Jane Doe",
		PersonType: PersonTypeCEO,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Nil(t, model.ID)
	assert.Equal(t, request.Name, model.Name)
	assert.Equal(t, request.PersonType.String(), model.PersonType.String())
	assert.Nil(t, model.Email)
	assert.Nil(t, model.Phone)
	assert.Nil(t, model.Notes)
}

// --------UpdatePersonRequest tests: --------

func TestUpdatePersonRequestValidate_ShouldValidateRequest(t *testing.T) {
	id := uuid.New()
	name := "Blue Gray"
	var personType PersonType = PersonTypeJobAdvertiser
	email := "blue@grey.se"
	phone := "3459083459"
	notes := "Notes about Blue Gray"

	request := UpdatePersonRequest{
		ID:         id,
		Name:       &name,
		PersonType: &personType,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestUpdatePersonRequestValidate_ShouldReturnValidationErrorIfNothingToUpdate(t *testing.T) {
	id := uuid.New()

	request := UpdatePersonRequest{
		ID: id,
	}

	err := request.validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error: nothing to update", validationErr.Error())
}

func TestUpdatePersonRequestToModel_ShouldReturnValidationErrorIfPersonTypeIsInvalid(t *testing.T) {
	id := uuid.New()
	var fakePersonType PersonType = "something that should never happen"

	request := UpdatePersonRequest{
		ID:         id,
		PersonType: &fakePersonType,
	}

	err := request.validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error on field 'PersonType': PersonType is invalid", validationErr.Error())
}

func TestUpdatePersonRequestToModel_ShouldConvertToModel(t *testing.T) {
	id := uuid.New()
	name := "Blah Rah"
	var personType PersonType = PersonTypeCEO
	email := "blah@email.sd"
	phone := "23972314945"
	notes := "Nothing to do here"

	request := UpdatePersonRequest{
		ID:         id,
		Name:       &name,
		PersonType: &personType,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, id, model.ID)
	assert.Equal(t, name, *model.Name)
	assert.Equal(t, personType.String(), model.PersonType.String())
	assert.Equal(t, email, *model.Email)
	assert.Equal(t, phone, *model.Phone)
	assert.Equal(t, notes, *model.Notes)
}

func TestUpdatePersonRequestToModel_ShouldConvertToModelWithNilValues(t *testing.T) {
	id := uuid.New()
	name := "No Name Today"

	request := UpdatePersonRequest{
		ID:   id,
		Name: &name,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, id, model.ID)
	assert.Equal(t, name, *model.Name)
	assert.Nil(t, model.PersonType)
	assert.Nil(t, model.Email)
	assert.Nil(t, model.Phone)
	assert.Nil(t, model.Notes)
}

func TestUpdatePersonRequestToModel_ShouldReturnValidationErrorIfNothingToUpdate(t *testing.T) {
	id := uuid.New()

	request := UpdatePersonRequest{
		ID: id,
	}

	model, err := request.ToModel()
	assert.Nil(t, model)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error: nothing to update", err.Error())
}

// -------- PersonType tests: --------

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

func TestPersonTypeIsValid_ShouldReturnFalseOnInvalidPersonType(t *testing.T) {

	empty := PersonType("")
	assert.False(t, empty.IsValid())

	nobody := PersonType("nobody")
	assert.False(t, nobody.IsValid())
}

func TestPersonTypeToModel_ShouldConvertToModel(t *testing.T) {
	tests := []struct {
		testName        string
		personType      PersonType
		modelPersonType models.PersonType
	}{
		{"CEO", PersonTypeCEO, models.PersonTypeCEO},
		{"CTO", PersonTypeCTO, models.PersonTypeCTO},
		{"Developer", PersonTypeDeveloper, models.PersonTypeDeveloper},
		{"Recruiter", PersonTypeExternalRecruiter, models.PersonTypeExternalRecruiter},
		{"HR", PersonTypeHR, models.PersonTypeHR},
		{"JobAdvertiser", PersonTypeJobAdvertiser, models.PersonTypeJobAdvertiser},
		{"JobContact", PersonTypeJobContact, models.PersonTypeJobContact},
		{"Other", PersonTypeOther, models.PersonTypeOther},
		{"Unknown", PersonTypeUnknown, models.PersonTypeUnknown},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			modelPersonType, err := test.personType.ToModel()
			assert.NoError(t, err)
			assert.NotNil(t, modelPersonType)
			assert.Equal(t, test.personType.String(), modelPersonType.String())
		})
	}
}

func TestPersonTypeToModel_ShouldReturnValidationErrorOnInvalidPersonType(t *testing.T) {
	empty := PersonType("")
	emptyModel, err := empty.ToModel()
	assert.NotNil(t, emptyModel)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "", emptyModel.String())
	assert.Equal(t, "validation error on field 'PersonType': invalid PersonType: ''", err.Error())

	blah := PersonType("Blah")
	blahModel, err := blah.ToModel()
	assert.NotNil(t, blahModel)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "", blahModel.String())
	assert.Equal(t, "validation error on field 'PersonType': invalid PersonType: 'Blah'", err.Error())
}

func TestNewPersonType_ShouldConvertFromModel(t *testing.T) {
	tests := []struct {
		testName        string
		modelPersonType models.PersonType
		personType      PersonType
	}{
		{"CEO", models.PersonTypeCEO, PersonTypeCEO},
		{"CTO", models.PersonTypeCTO, PersonTypeCTO},
		{"Developer", models.PersonTypeDeveloper, PersonTypeDeveloper},
		{"Recruiter", models.PersonTypeExternalRecruiter, PersonTypeExternalRecruiter},
		{"HR", models.PersonTypeHR, PersonTypeHR},
		{"JobAdvertiser", models.PersonTypeJobAdvertiser, PersonTypeJobAdvertiser},
		{"JobContact", models.PersonTypeJobContact, PersonTypeJobContact},
		{"Other", models.PersonTypeOther, PersonTypeOther},
		{"Unknown", models.PersonTypeUnknown, PersonTypeUnknown},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			personType, err := NewPersonType(&test.modelPersonType)
			assert.NoError(t, err)
			assert.NotNil(t, personType)
			assert.Equal(t, test.personType.String(), personType.String())
		})
	}
}

func TestPersonTypeToModel_ShouldReturnInternalServiceErrorOnNilPersonType(t *testing.T) {
	personType, err := NewPersonType(nil)
	assert.NotNil(t, personType)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, "", personType.String())
	assert.Equal(t, "internal service error: Error trying to convert internal personType to external PersonType.", err.Error())
}

func TestPersonTypeToModel_ShouldReturnInternalServiceErrorOnInvalidPersonType(t *testing.T) {
	emptyModel := models.PersonType("")
	emptyPerson, err := NewPersonType(&emptyModel)
	assert.NotNil(t, err)
	assert.NotNil(t, emptyPerson)
	assert.Equal(t, "", emptyPerson.String())

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, "", emptyPerson.String())

	specialistModel := models.PersonType("specialist")
	specialist, err := NewPersonType(&specialistModel)
	assert.NotNil(t, err)
	assert.NotNil(t, specialist)
	assert.Equal(t, "", specialist.String())

	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal PersonType to external PersonType: 'specialist'",
		err.Error())
}
