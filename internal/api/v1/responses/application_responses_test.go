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

// -------- NewApplicationDTO tests: --------

func TestNewApplicationDTO_ShouldWork(t *testing.T) {
	var remoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	model := models.Application{
		ID:                   uuid.New(),
		CompanyID:            testutil.ToPtr(uuid.New()),
		RecruiterID:          testutil.ToPtr(uuid.New()),
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Job Country"),
		Area:                 testutil.ToPtr("Job Area"),
		RemoteStatusType:     &remoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(2),
		EstimatedCycleTime:   testutil.ToPtr(30),
		EstimatedCommuteTime: testutil.ToPtr(40),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}

	dto, err := NewApplicationDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, dto)

	assert.Equal(t, model.ID, dto.ID)
	assert.Equal(t, model.CompanyID, dto.CompanyID)
	assert.Equal(t, model.RecruiterID, dto.RecruiterID)
	assert.Equal(t, model.JobTitle, dto.JobTitle)
	assert.Equal(t, model.JobAdURL, dto.JobAdURL)
	assert.Equal(t, model.Country, dto.Country)
	assert.Equal(t, model.Area, dto.Area)
	assert.Equal(t, model.RemoteStatusType.String(), dto.RemoteStatusType.String())
	assert.Equal(t, model.WeekdaysInOffice, dto.WeekdaysInOffice)
	assert.Equal(t, model.EstimatedCycleTime, dto.EstimatedCycleTime)
	assert.Equal(t, model.EstimatedCommuteTime, dto.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, model.ApplicationDate, dto.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, dto.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.UpdatedDate, dto.UpdatedDate)
}

func TestNewApplicationDTO_ShouldWorkWithOnlyID(t *testing.T) {
	model := models.Application{
		ID: uuid.New(),
	}
	dto, err := NewApplicationDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, dto)

	assert.Equal(t, model.ID.String(), dto.ID.String())
	assert.Nil(t, dto.CompanyID)
	assert.Nil(t, dto.RecruiterID)
	assert.Nil(t, dto.JobTitle)
	assert.Nil(t, dto.JobAdURL)
	assert.Nil(t, dto.Country)
	assert.Nil(t, dto.Area)
	assert.Nil(t, dto.RemoteStatusType)
	assert.Nil(t, dto.WeekdaysInOffice)
	assert.Nil(t, dto.EstimatedCycleTime)
	assert.Nil(t, dto.EstimatedCommuteTime)
	assert.Nil(t, dto.ApplicationDate)
	assert.Nil(t, dto.CreatedDate)
	assert.Nil(t, dto.UpdatedDate)
}

func TestNewApplicationDTO_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewApplicationDTO(nil)
	assert.Nil(t, nilModel)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, internalServiceError.Error(), "internal service error: Error building DTO: Application is nil")
}

func TestNewApplicationDTO_ShouldReturnInternalServiceErrorIfRemoteStatusTypeIsInvalid(t *testing.T) {
	var remoteStatusTypeEmpty models.RemoteStatusType = ""
	emptyRemoteStatusType := models.Application{
		ID:               uuid.New(),
		RemoteStatusType: &remoteStatusTypeEmpty,
	}
	emptyDTO, err := NewApplicationDTO(&emptyRemoteStatusType)
	assert.Nil(t, emptyDTO)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: ''",
		internalServiceError.Error())

	var remoteStatusTypeBlah models.RemoteStatusType = "Blah"
	invalidRemoteStatusType := models.Application{
		ID:               uuid.New(),
		RemoteStatusType: &remoteStatusTypeBlah,
	}
	invalidDTO, err := NewApplicationDTO(&invalidRemoteStatusType)
	assert.Nil(t, invalidDTO)
	assert.Error(t, err)

	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: 'Blah'",
		internalServiceError.Error())
}

// -------- NewApplicationDTOs tests: --------

func TestNewApplicationDTOs_ShouldWork(t *testing.T) {
	var application1RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeUnknown
	var application2RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeRemote

	applicationModels := []*models.Application{
		{
			ID:               uuid.New(),
			CompanyID:        testutil.ToPtr(uuid.New()),
			JobAdURL:         testutil.ToPtr("Job Ad URL"),
			RemoteStatusType: &application1RemoteStatusType,
			CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		},
		{
			ID:               uuid.New(),
			RecruiterID:      testutil.ToPtr(uuid.New()),
			JobTitle:         testutil.ToPtr("Job Title "),
			RemoteStatusType: &application2RemoteStatusType,
			CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		},
	}
	applicationDTOs, err := NewApplicationDTOs(applicationModels)
	assert.NoError(t, err)
	assert.NotNil(t, applicationDTOs)
	assert.Len(t, applicationDTOs, 2)
}

func TestNewApplicationDTOs_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	emptyDTOs, err := NewApplicationDTOs(nil)
	assert.NoError(t, err)
	assert.NotNil(t, emptyDTOs)
	assert.Len(t, emptyDTOs, 0)
}

func TestNewApplicationDTOs_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var applicationModels []*models.Application
	emptyDTOs, err := NewApplicationDTOs(applicationModels)
	assert.NoError(t, err)
	assert.NotNil(t, emptyDTOs)
	assert.Len(t, emptyDTOs, 0)
}

func TestNewApplicationDTOs_ShouldReturnNilIfOneRemoteStatusTypeIsInvalid(t *testing.T) {
	var application1RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeUnknown
	var application2RemoteStatusType models.RemoteStatusType = ""
	applicationModels := []*models.Application{
		{
			ID:               uuid.New(),
			RecruiterID:      testutil.ToPtr(uuid.New()),
			JobTitle:         testutil.ToPtr("Job Title "),
			RemoteStatusType: &application1RemoteStatusType,
			CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 7)),
		},
		{
			ID:               uuid.New(),
			RecruiterID:      testutil.ToPtr(uuid.New()),
			JobTitle:         testutil.ToPtr("Job Title "),
			RemoteStatusType: &application2RemoteStatusType,
			CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 0)),
		},
	}

	nilDTOs, err := NewApplicationDTOs(applicationModels)
	assert.Nil(t, nilDTOs)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: ''",
		internalServiceError.Error())
}

// -------- NewApplicationResponse tests: --------

func TestNewApplicationResponse_ShouldWork(t *testing.T) {
	var remoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	model := models.Application{
		ID:                   uuid.New(),
		CompanyID:            testutil.ToPtr(uuid.New()),
		RecruiterID:          testutil.ToPtr(uuid.New()),
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Job Country"),
		Area:                 testutil.ToPtr("Job Area"),
		RemoteStatusType:     &remoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(2),
		EstimatedCycleTime:   testutil.ToPtr(30),
		EstimatedCommuteTime: testutil.ToPtr(40),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}

	response, err := NewApplicationResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.CompanyID, response.CompanyID)
	assert.Equal(t, model.RecruiterID, response.RecruiterID)
	assert.Equal(t, model.JobTitle, response.JobTitle)
	assert.Equal(t, model.JobAdURL, response.JobAdURL)
	assert.Equal(t, model.Country, response.Country)
	assert.Equal(t, model.Area, response.Area)
	assert.Equal(t, model.RemoteStatusType.String(), response.RemoteStatusType.String())
	assert.Equal(t, model.WeekdaysInOffice, response.WeekdaysInOffice)
	assert.Equal(t, model.EstimatedCycleTime, response.EstimatedCycleTime)
	assert.Equal(t, model.EstimatedCommuteTime, response.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, model.ApplicationDate, response.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, model.CreatedDate, response.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.UpdatedDate, response.UpdatedDate)
}

func TestNewApplicationResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewApplicationResponse(nil)
	assert.Nil(t, nilModel)
	assert.Error(t, err)

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		internalServiceError.Error(),
		"internal service error: Error building response: Application is nil")
}

func TestNewApplicationResponse_ShouldHandleCompany(t *testing.T) {
	var companyType models.CompanyType = models.CompanyTypeEmployer
	company := models.Company{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Company Name"),
		CompanyType: &companyType,
		Notes:       testutil.ToPtr("Company Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 6)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 7)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 8)),
	}

	model := models.Application{
		ID:        uuid.New(),
		CompanyID: testutil.ToPtr(uuid.New()),
		Company:   &company,
	}

	response, err := NewApplicationResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.CompanyID, response.CompanyID)
	assert.NotNil(t, response.Company)

	assert.Equal(t, company.ID, response.Company.ID)
	assert.Equal(t, company.Name, response.Company.Name)
	assert.Equal(t, company.CompanyType.String(), response.Company.CompanyType.String())
	assert.Equal(t, company.Notes, response.Company.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.Company.LastContact, response.Company.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, model.Company.CreatedDate, response.Company.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.Company.UpdatedDate, response.Company.UpdatedDate)
}

func TestNewApplicationResponse_ShouldHandleRecruiter(t *testing.T) {
	var companyType models.CompanyType = models.CompanyTypeEmployer
	recruiter := models.Company{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Recruiter Name"),
		CompanyType: &companyType,
		Notes:       testutil.ToPtr("Recruiter Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 6)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 7)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 8)),
	}

	model := models.Application{
		ID:          uuid.New(),
		RecruiterID: testutil.ToPtr(uuid.New()),
		Recruiter:   &recruiter,
	}

	response, err := NewApplicationResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.RecruiterID, response.RecruiterID)
	assert.NotNil(t, response.Recruiter)

	assert.Equal(t, recruiter.ID, response.Recruiter.ID)
	assert.Equal(t, recruiter.Name, response.Recruiter.Name)
	assert.Equal(t, recruiter.CompanyType.String(), response.Recruiter.CompanyType.String())
	assert.Equal(t, recruiter.Notes, response.Recruiter.Notes)
	testutil.AssertEqualFormattedDateTimes(t, model.Recruiter.LastContact, response.Recruiter.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, model.Recruiter.CreatedDate, response.Recruiter.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, model.Recruiter.UpdatedDate, response.Recruiter.UpdatedDate)
}

func TestNewApplicationResponse_ShouldHandlePersons(t *testing.T) {

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

	model := models.Application{
		ID:      uuid.New(),
		Persons: &persons,
	}

	response, err := NewApplicationResponse(&model)
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

func TestNewApplicationResponse_ShouldHandleEvents(t *testing.T) {

	var eventType models.EventType = models.EventTypeApplied
	event1 := models.Event{
		ID:          uuid.New(),
		EventType:   &eventType,
		Description: testutil.ToPtr("Event Description"),
		Notes:       testutil.ToPtr("Event Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 0, 11)),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 12)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 13)),
	}

	event2 := models.Event{
		ID:          uuid.New(),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 14)),
	}

	events := []*models.Event{
		&event1,
		&event2,
	}

	model := models.Application{
		ID:     uuid.New(),
		Events: &events,
	}

	response, err := NewApplicationResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.NotNil(t, response.Events)
	assert.Len(t, *response.Events, 2)

	eventDTO1 := (*response.Events)[0]
	assert.Equal(t, event1.ID, *eventDTO1.ID)
	assert.Equal(t, event1.EventType.String(), eventDTO1.EventType.String())
	assert.Equal(t, event1.Description, eventDTO1.Description)
	assert.Equal(t, event1.Notes, eventDTO1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, event1.EventDate, eventDTO1.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.CreatedDate, eventDTO1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.UpdatedDate, eventDTO1.UpdatedDate)

	assert.Equal(t, event2.ID, *(*response.Events)[1].ID)
}

// -------- NewApplicationsResponse tests: --------

func TestNewApplicationsResponseShouldWork(t *testing.T) {
	var application1RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeUnknown
	var application2RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeRemote
	applicationModels := []*models.Application{
		{
			ID:               uuid.New(),
			CompanyID:        testutil.ToPtr(uuid.New()),
			JobAdURL:         testutil.ToPtr("Job Ad URL"),
			RemoteStatusType: &application1RemoteStatusType,
			CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		},
		{
			ID:               uuid.New(),
			RecruiterID:      testutil.ToPtr(uuid.New()),
			JobTitle:         testutil.ToPtr("Job Title"),
			RemoteStatusType: &application2RemoteStatusType,
			CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		},
	}

	applications, err := NewApplicationsResponse(applicationModels)
	assert.NoError(t, err)
	assert.NotNil(t, applications)
	assert.Len(t, applications, 2)
}

func TestNewApplicationsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	emptyResponse, err := NewApplicationsResponse(nil)
	assert.NoError(t, err)
	assert.NotNil(t, emptyResponse)
	assert.Len(t, emptyResponse, 0)
}

func TestNewApplicationsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var applicationModels []*models.Application
	emptyResponse, err := NewApplicationsResponse(applicationModels)
	assert.NoError(t, err)
	assert.NotNil(t, emptyResponse)
	assert.Len(t, emptyResponse, 0)
}
