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

// -------- NewEventDTO tests: --------

func TestNewEventDTO_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied
	model := models.Event{
		ID:          uuid.New(),
		EventType:   &eventType,
		Description: testutil.ToPtr("Description"),
		Notes:       testutil.ToPtr("Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
	}

	eventDTO, err := NewEventDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventDTO)

	assert.Equal(t, model.ID, eventDTO.ID)
	assert.Equal(t, model.EventType.String(), eventDTO.EventType.String())
	assert.Equal(t, model.Description, eventDTO.Description)
	assert.Equal(t, model.Notes, eventDTO.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.EventDate, eventDTO.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, eventDTO.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.UpdatedDate, eventDTO.UpdatedDate)
}

func TestNewEventDTO_ShouldWorkWithOnlyID(t *testing.T) {
	var model = models.Event{
		ID: uuid.New(),
	}
	eventDTO, err := NewEventDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventDTO)
	assert.Equal(t, model.ID, eventDTO.ID)
}

func TestNewEventDTO_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilDTO, err := NewEventDTO(nil)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, internalServiceError.Error(), "internal service error: Error building DTO: Event is nil")
}

func TestNewEventDTO_ShouldReturnInternalServiceErrorIfEventTypeIsEmpty(t *testing.T) {
	var empty models.EventType = ""
	emptyEventType := models.Event{
		ID:        uuid.New(),
		EventType: &empty,
	}
	nilDTO, err := NewEventDTO(&emptyEventType)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: ''",
		internalServiceError.Error())
}

func TestNewEventDTO_ShouldReturnInternalServiceErrorIfEventTypeIsInvalid(t *testing.T) {
	var invalid models.EventType = "hiringPaused"
	emptyEventType := models.Event{
		ID:        uuid.New(),
		EventType: &invalid,
	}
	nilDTO, err := NewEventDTO(&emptyEventType)
	assert.Nil(t, nilDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: 'hiringPaused'",
		internalServiceError.Error())
}

// -------- NewEventDTOs tests: --------
func TestNewEventDTOs_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied

	eventModels := []*models.Event{
		{
			ID:          uuid.New(),
			EventType:   &eventType,
			Description: testutil.ToPtr("Description"),
			Notes:       testutil.ToPtr("Notes"),
			EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
			UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
		},
		{
			ID: uuid.New(),
		},
	}

	eventDTOs, err := NewEventDTOs(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, eventDTOs)
	assert.Len(t, eventDTOs, 2)
}

func TestNewEventDTOs_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewEventDTOs(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewEventDTOs_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var eventModels []*models.Event
	response, err := NewEventDTOs(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewEventDTOs_ShouldReturnNilIfOneEventTypeIsInvalid(t *testing.T) {
	var eventTypeApplied models.EventType = models.EventTypeApplied
	var eventTypeEmpty models.EventType = ""

	eventModels := []*models.Event{
		{
			ID:        uuid.New(),
			EventType: &eventTypeApplied,
		},
		{
			ID:        uuid.New(),
			EventType: &eventTypeEmpty,
		},
	}
	nilDTOs, err := NewEventDTOs(eventModels)
	assert.Nil(t, nilDTOs)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: ''",
		internalServiceError.Error())
}

// -------- NewEventResponse tests: --------

func TestNewEventResponse_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied
	model := models.Event{
		ID:          uuid.New(),
		EventType:   &eventType,
		Description: testutil.ToPtr("Description"),
		Notes:       testutil.ToPtr("Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
	}

	eventResponse, err := NewEventResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventResponse)

	assert.Equal(t, model.ID, eventResponse.ID)
	assert.Equal(t, model.EventType.String(), eventResponse.EventType.String())
	assert.Equal(t, model.Description, eventResponse.Description)
	assert.Equal(t, model.Notes, eventResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.EventDate, eventResponse.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, eventResponse.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.UpdatedDate, eventResponse.UpdatedDate)
}

func TestNewEventResponse_ShouldWorkWithOnlyID(t *testing.T) {
	var model = models.Event{
		ID: uuid.New(),
	}
	eventResponse, err := NewEventResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, eventResponse)
	assert.Equal(t, model.ID, eventResponse.ID)
}

func TestNewEventResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilResponse, err := NewEventResponse(nil)
	assert.Nil(t, nilResponse)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, internalServiceError.Error(), "internal service error: Error building response: Event is nil")
}

func TestNewEventResponse_ShouldHandleApplications(t *testing.T) {
	var application1RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	application1Model := models.Application{
		ID:                   uuid.New(),
		CompanyID:            testutil.ToPtr(uuid.New()),
		RecruiterID:          testutil.ToPtr(uuid.New()),
		JobTitle:             testutil.ToPtr("Application 1 Job Title"),
		JobAdURL:             testutil.ToPtr("Application 1 Job Ad URL"),
		Country:              testutil.ToPtr("Application 1 Job Country"),
		Area:                 testutil.ToPtr("Application 1 Job Area"),
		RemoteStatusType:     &application1RemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(2),
		EstimatedCycleTime:   testutil.ToPtr(30),
		EstimatedCommuteTime: testutil.ToPtr(40),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}

	application2Model := models.Application{
		ID: uuid.New(),
	}

	applicationModels := []*models.Application{&application1Model, &application2Model}
	model := models.Event{
		ID:           uuid.New(),
		Applications: &applicationModels,
	}

	response, err := NewEventResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.NotNil(t, response.Applications)

	assert.Len(t, *response.Applications, 2)

	application1 := (*response.Applications)[0]
	assert.Equal(t, application1Model.ID, application1.ID)
	assert.Equal(t, application1Model.CompanyID, application1.CompanyID)
	assert.Equal(t, application1Model.RecruiterID, application1.RecruiterID)
	assert.Equal(t, application1Model.JobTitle, application1.JobTitle)
	assert.Equal(t, application1Model.JobAdURL, application1.JobAdURL)
	assert.Equal(t, application1Model.Country, application1.Country)
	assert.Equal(t, application1Model.Area, application1.Area)
	assert.Equal(t, application1Model.RemoteStatusType.String(), application1.RemoteStatusType.String())
	assert.Equal(t, application1Model.WeekdaysInOffice, application1.WeekdaysInOffice)
	assert.Equal(t, application1Model.EstimatedCycleTime, application1.EstimatedCycleTime)
	assert.Equal(t, application1Model.EstimatedCommuteTime, application1.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, application1Model.ApplicationDate, application1.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, application1Model.CreatedDate, application1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, application1Model.UpdatedDate, application1.UpdatedDate)

	application2 := (*response.Applications)[1]
	assert.Equal(t, application2Model.ID, application2.ID)
}

func TestNewEventResponse_ShouldHandleCompanies(t *testing.T) {
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
	model := models.Event{
		ID:        uuid.New(),
		Companies: &companyModels,
	}

	response, err := NewEventResponse(&model)
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

func TestNewEventResponse_ShouldHandlePersons(t *testing.T) {

	var personType models.PersonType = models.PersonTypeDeveloper
	person1 := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Person Name"),
		PersonType:  &personType,
		Email:       testutil.ToPtr("Person Email"),
		Phone:       testutil.ToPtr("Person Phone"),
		Notes:       testutil.ToPtr("Person Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 12)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 13)),
	}

	person2 := models.Person{
		ID:          uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 14)),
	}

	persons := []*models.Person{
		&person1,
		&person2,
	}

	model := models.Event{
		ID:      uuid.New(),
		Persons: &persons,
	}

	response, err := NewEventResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.NotNil(t, response.Persons)
	assert.Len(t, *response.Persons, 2)

	personDTO1 := (*response.Persons)[0]
	assert.Equal(t, person1.ID, personDTO1.ID)
	assert.Equal(t, person1.Name, personDTO1.Name)
	assert.Equal(t, person1.PersonType.String(), personDTO1.PersonType.String())
	assert.Equal(t, person1.Email, personDTO1.Email)
	assert.Equal(t, person1.Phone, personDTO1.Phone)
	assert.Equal(t, person1.Notes, personDTO1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, person1.CreatedDate, personDTO1.CreatedDate)

	assert.Equal(t, person2.ID, (*response.Persons)[1].ID)
}

// -------- NewEventsResponse tests: --------

func TestNewEventsResponse_ShouldWork(t *testing.T) {
	var eventType models.EventType = models.EventTypeApplied

	eventModels := []*models.Event{
		{
			ID:          uuid.New(),
			EventType:   &eventType,
			Description: testutil.ToPtr("Description"),
			Notes:       testutil.ToPtr("Notes"),
			EventDate:   testutil.ToPtr(time.Now().AddDate(0, 4, 0)),
			CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
			UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
		},
		{
			ID: uuid.New(),
		},
	}

	EventsResponse, err := NewEventsResponse(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, EventsResponse)
	assert.Len(t, EventsResponse, 2)
}

func TestNewEventsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewEventsResponse(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewEventsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var eventModels []*models.Event
	response, err := NewEventsResponse(eventModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
