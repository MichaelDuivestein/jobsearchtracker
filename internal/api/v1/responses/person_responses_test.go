package responses

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

// -------- NewPersonDTO tests: --------

func TestNewPersonDTO_ShouldWork(t *testing.T) {
	var personType models.PersonType = models.PersonTypeJobContact
	model := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Blah blah"),
		PersonType:  &personType,
		Email:       testutil.ToPtr("e@m.ai"),
		Phone:       testutil.ToPtr("2345"),
		Notes:       testutil.ToPtr("sdfgkljherwkl"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -1, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 27)),
	}

	personDTO, err := NewPersonDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, personDTO)

	assert.Equal(t, model.ID.String(), personDTO.ID.String())
	assert.Equal(t, model.Name, personDTO.Name)
	assert.Equal(t, model.PersonType.String(), personDTO.PersonType.String())
	assert.Equal(t, model.Email, personDTO.Email)
	assert.Equal(t, model.Phone, personDTO.Phone)
	assert.Equal(t, model.Notes, personDTO.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, personDTO.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.UpdatedDate, personDTO.UpdatedDate)
}

func TestNewPersonDTO_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	var personType models.PersonType = models.PersonTypeUnknown
	model := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Anker"),
		PersonType:  &personType,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
	}

	personDTO, err := NewPersonDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, personDTO)

	assert.Equal(t, model.ID.String(), personDTO.ID.String())
	assert.Equal(t, model.Name, personDTO.Name)
	assert.Equal(t, model.PersonType.String(), personDTO.PersonType.String())
	assert.Nil(t, personDTO.Email)
	assert.Nil(t, personDTO.Phone)
	assert.Nil(t, personDTO.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, personDTO.CreatedDate)
	assert.Nil(t, personDTO.UpdatedDate)
}

func TestNewPersonDTO_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilDTO, err := NewPersonDTO(nil)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, internalServiceError.Error(), "internal service error: Error building DTO: Person is nil")
}

func TestNewPersonDTO_ShouldReturnInternalServiceErrorIfPersonTypeIsInvalid(t *testing.T) {
	var personTypeEmpty models.PersonType = ""
	emptyPersonType := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Dave"),
		PersonType:  &personTypeEmpty,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 16)),
	}
	nilDTO, err := NewPersonDTO(&emptyPersonType)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal PersonType to external PersonType: ''",
		internalServiceError.Error())

	var personTypeBlah models.PersonType = "Blah"
	invalidPersonType := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Dave"),
		PersonType:  &personTypeBlah,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 16)),
	}
	invalidDTO, err := NewPersonDTO(&invalidPersonType)
	assert.Nil(t, invalidDTO)
	assert.Error(t, err)

	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal PersonType to external PersonType: 'Blah'",
		internalServiceError.Error())
}

// -------- NewPersonDTOs tests: --------

func TestNewPersonDTOs_ShouldWork(t *testing.T) {
	var personTypeUnknown models.PersonType = models.PersonTypeUnknown
	var personTypeCTO models.PersonType = models.PersonTypeCTO
	personModels := []*models.Person{
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Aaron"),
			PersonType:  &personTypeUnknown,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		},
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Bru"),
			PersonType:  &personTypeCTO,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		},
	}

	personDTOs, err := NewPersonDTOs(personModels)
	assert.NoError(t, err)
	assert.NotNil(t, personDTOs)
	assert.Len(t, personDTOs, 2)
}

func TestNewPersonDTOs_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	emptyDTOs, err := NewPersonDTOs(nil)
	assert.NoError(t, err)
	assert.NotNil(t, emptyDTOs)
	assert.Len(t, emptyDTOs, 0)
}

func TestNewPersonDTOs_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var personModels []*models.Person
	emptyDTOs, err := NewPersonDTOs(personModels)
	assert.NoError(t, err)
	assert.NotNil(t, emptyDTOs)
	assert.Len(t, emptyDTOs, 0)
}

func TestNewPersonDTOs_ShouldReturnInternalServiceErrorIfOnePersonTypeIsInvalid(t *testing.T) {
	var personTypeJobAdvertiser models.PersonType = models.PersonTypeJobAdvertiser
	var personTypeEmpty models.PersonType = ""

	personModels := []*models.Person{
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Sammy"),
			PersonType:  &personTypeJobAdvertiser,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 7)),
		},
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Britt"),
			PersonType:  &personTypeEmpty,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 0)),
		},
	}
	persons, err := NewPersonDTOs(personModels)
	assert.Nil(t, persons)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal PersonType to external PersonType: ''",
		internalServiceError.Error())
}

// -------- NewPersonResponse tests: --------

func TestNewPersonResponse_ShouldWork(t *testing.T) {
	var personType models.PersonType = models.PersonTypeJobContact
	model := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Blah blah"),
		PersonType:  &personType,
		Email:       testutil.ToPtr("e@m.ai"),
		Phone:       testutil.ToPtr("2345"),
		Notes:       testutil.ToPtr("sdfgkljherwkl"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, -1, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 27)),
	}

	response, err := NewPersonResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.PersonType.String(), response.PersonType.String())
	assert.Equal(t, model.Email, response.Email)
	assert.Equal(t, model.Phone, response.Phone)
	assert.Equal(t, model.Notes, response.Notes)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Equal(t, model.UpdatedDate, response.UpdatedDate)
}

func TestNewPersonResponse_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	var personType models.PersonType = models.PersonTypeUnknown
	model := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Anker"),
		PersonType:  &personType,
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
	}

	response, err := NewPersonResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.PersonType.String(), response.PersonType.String())
	assert.Nil(t, response.Email)
	assert.Nil(t, response.Phone)
	assert.Nil(t, response.Notes)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Nil(t, response.UpdatedDate)
}

func TestNewPersonResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewPersonResponse(nil)
	assert.Nil(t, nilModel)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, internalServiceError.Error(), "internal service error: Error building response: Person is nil")
}

func TestNewPersonResponse_ShouldHandleCompanies(t *testing.T) {
	var company1CompanyType models.CompanyType = models.CompanyTypeRecruiter
	company1Model := models.Company{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Company1Name"),
		CompanyType: &company1CompanyType,
		Notes:       testutil.ToPtr("Company1Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}

	company2Model := models.Company{
		ID: uuid.New(),
	}

	companyModels := []*models.Company{&company1Model, &company2Model}
	var personType models.PersonType = models.PersonTypeJobContact
	model := models.Person{
		ID:         uuid.New(),
		Name:       testutil.ToPtr("PersonName"),
		PersonType: &personType,
		Companies:  &companyModels,
	}

	response, err := NewPersonResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.NotNil(t, response.Companies)

	assert.Len(t, *response.Companies, 2)

	company1 := (*response.Companies)[0]
	assert.Equal(t, company1Model.ID, company1.ID)
	assert.Equal(t, company1Model.Name, company1.Name)
	assert.Equal(t, company1Model.CompanyType.String(), company1.CompanyType.String())
	assert.Equal(t, company1Model.Notes, company1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, company1.LastContact, company1Model.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company1.CreatedDate, company1Model.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company1.UpdatedDate, company1Model.UpdatedDate)

	company2 := (*response.Companies)[1]
	assert.Equal(t, company2Model.ID, company2.ID)
}

func TestNewPersonResponse_ShouldHandleEvents(t *testing.T) {
	var event1EventType models.EventType = models.EventTypeApplied
	event1Model := models.Event{
		ID:          uuid.New(),
		EventType:   &event1EventType,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}

	event2Model := models.Event{
		ID: uuid.New(),
	}

	eventModels := []*models.Event{&event1Model, &event2Model}
	var personType models.PersonType = models.PersonTypeJobContact
	model := models.Person{
		ID:         uuid.New(),
		Name:       testutil.ToPtr("PersonName"),
		PersonType: &personType,
		Events:     &eventModels,
	}

	response, err := NewPersonResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.NotNil(t, response.Events)

	assert.Len(t, *response.Events, 2)

	event1 := (*response.Events)[0]
	assert.Equal(t, event1Model.ID, event1.ID)
	assert.Equal(t, event1Model.EventType.String(), event1.EventType.String())
	assert.Equal(t, event1Model.Description, event1.Description)
	assert.Equal(t, event1Model.Notes, event1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, event1.EventDate, event1Model.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.CreatedDate, event1Model.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.UpdatedDate, event1Model.UpdatedDate)

	event2 := (*response.Events)[1]
	assert.Equal(t, event2Model.ID, event2.ID)
}

// -------- NewPersonsResponse tests: --------

func TestNewPersonsResponse_ShouldWork(t *testing.T) {
	var personTypeUnknown models.PersonType = models.PersonTypeUnknown
	var personTypeCTO models.PersonType = models.PersonTypeCTO
	personModels := []*models.Person{
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Aaron"),
			PersonType:  &personTypeUnknown,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		},
		{
			ID:          uuid.New(),
			Name:        testutil.ToPtr("Bru"),
			PersonType:  &personTypeCTO,
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		},
	}

	persons, err := NewPersonsResponse(personModels)
	assert.NoError(t, err)
	assert.NotNil(t, persons)
	assert.Len(t, persons, 2)
}

func TestNewPersonsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewPersonsResponse(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewPersonsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var personModels []*models.Person
	response, err := NewPersonsResponse(personModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
