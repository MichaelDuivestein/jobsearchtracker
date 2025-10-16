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
	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Job Country"
	area := "Job Area"
	var remoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	weekdaysInOffice := 2
	estimatedCycleTime := 30
	estimatedCommuteTime := 40
	applicationDate := time.Now().AddDate(0, 0, 1)
	createdDate := time.Now().AddDate(0, 0, 2)
	updatedDate := time.Now().AddDate(0, 0, 3)

	model := models.Application{
		ID:                   id,
		CompanyID:            &companyID,
		RecruiterID:          &recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     &remoteStatusType,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
	}

	dto, err := NewApplicationDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, dto)

	assert.Equal(t, id, dto.ID)
	assert.Equal(t, companyID, *dto.CompanyID)
	assert.Equal(t, recruiterID, *dto.RecruiterID)
	assert.Equal(t, jobTitle, *dto.JobTitle)
	assert.Equal(t, jobAdURL, *dto.JobAdURL)
	assert.Equal(t, country, *dto.Country)
	assert.Equal(t, area, *dto.Area)
	assert.Equal(t, models.RemoteStatusTypeHybrid, dto.RemoteStatusType.String())
	assert.Equal(t, weekdaysInOffice, *dto.WeekdaysInOffice)
	assert.Equal(t, estimatedCycleTime, *dto.EstimatedCycleTime)
	assert.Equal(t, estimatedCommuteTime, *dto.EstimatedCommuteTime)

	applicationToInsertApplicationDate := applicationDate.Format(time.RFC3339)
	insertedApplicationApplicationDate := dto.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertApplicationDate, insertedApplicationApplicationDate)

	applicationToInsertCreatedDate := createdDate.Format(time.RFC3339)
	insertedApplicationCreatedDate := dto.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertCreatedDate, insertedApplicationCreatedDate)

	applicationToInsertUpdatedDate := updatedDate.Format(time.RFC3339)
	insertedApplicationUpdatedDate := dto.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertUpdatedDate, insertedApplicationUpdatedDate)
}

func TestNewApplicationDTO_ShouldWorkWithOnlyRequiredFields(t *testing.T) {

	companyID := uuid.New()
	jobAdURL := "Job Ad URL"
	var remoteStatusType models.RemoteStatusType = models.RemoteStatusTypeRemote

	model := models.Application{
		ID:               uuid.New(),
		CompanyID:        &companyID,
		JobAdURL:         &jobAdURL,
		RemoteStatusType: &remoteStatusType,
		CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
	}

	dto, err := NewApplicationDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, dto)

	assert.Equal(t, model.ID.String(), dto.ID.String())
	assert.Equal(t, model.CompanyID, dto.CompanyID)
	assert.Nil(t, model.RecruiterID)
	assert.Nil(t, dto.JobTitle)
	assert.Equal(t, model.JobAdURL, dto.JobAdURL)
	assert.Nil(t, dto.Country)
	assert.Nil(t, dto.Area)
	assert.Equal(t, model.RemoteStatusType.String(), dto.RemoteStatusType.String())
	assert.Nil(t, dto.WeekdaysInOffice)
	assert.Nil(t, dto.EstimatedCycleTime)
	assert.Nil(t, dto.EstimatedCommuteTime)
	assert.Nil(t, dto.ApplicationDate)
	assert.Equal(t, model.CreatedDate, dto.CreatedDate)
	assert.Nil(t, dto.UpdatedDate)
}

func TestNewApplicationDTO_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewApplicationDTO(nil)
	assert.Nil(t, nilModel)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, err.Error(), "internal service error: Error building DTO: Application is nil")
}

func TestNewApplicationDTO_ShouldReturnInternalServiceErrorIfRemoteStatusTypeIsInvalid(t *testing.T) {
	recruiterID := uuid.New()
	JobAdURL := "Job Ad URL"

	var remoteStatusTypeEmpty models.RemoteStatusType = ""
	emptyRemoteStatusType := models.Application{
		ID:               uuid.New(),
		RecruiterID:      &recruiterID,
		JobAdURL:         &JobAdURL,
		RemoteStatusType: &remoteStatusTypeEmpty,
		CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 16)),
	}
	emptyDTO, err := NewApplicationDTO(&emptyRemoteStatusType)
	assert.Nil(t, emptyDTO)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: ''",
		err.Error())

	var remoteStatusTypeBlah models.RemoteStatusType = "Blah"
	invalidRemoteStatusType := models.Application{
		ID:               uuid.New(),
		RecruiterID:      &recruiterID,
		JobAdURL:         &JobAdURL,
		RemoteStatusType: &remoteStatusTypeBlah,
		CreatedDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 16)),
	}
	invalidDTO, err := NewApplicationDTO(&invalidRemoteStatusType)
	assert.Nil(t, invalidDTO)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: 'Blah'",
		err.Error())
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

func TestNewApplicationDTOs_ShouldReturnEmptySliceIfOneRemoteStatusTypeIsInvalid(t *testing.T) {
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
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(
		t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: ''",
		err.Error())
}

// -------- NewApplicationResponse tests: --------

func TestNewApplicationResponse_ShouldWork(t *testing.T) {
	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Job Country"
	area := "Job Area"
	var remoteStatusType models.RemoteStatusType = models.RemoteStatusTypeHybrid
	weekdaysInOffice := 2
	estimatedCycleTime := 30
	estimatedCommuteTime := 40
	applicationDate := time.Now().AddDate(0, 0, 1)
	createdDate := time.Now().AddDate(0, 0, 2)
	updatedDate := time.Now().AddDate(0, 0, 3)

	model := models.Application{
		ID:                   id,
		CompanyID:            &companyID,
		RecruiterID:          &recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     &remoteStatusType,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
	}

	response, err := NewApplicationResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, id, response.ID)
	assert.Equal(t, companyID, *response.CompanyID)
	assert.Equal(t, recruiterID, *response.RecruiterID)
	assert.Equal(t, jobTitle, *response.JobTitle)
	assert.Equal(t, jobAdURL, *response.JobAdURL)
	assert.Equal(t, country, *response.Country)
	assert.Equal(t, area, *response.Area)
	assert.Equal(t, models.RemoteStatusTypeHybrid, response.RemoteStatusType.String())
	assert.Equal(t, weekdaysInOffice, *response.WeekdaysInOffice)
	assert.Equal(t, estimatedCycleTime, *response.EstimatedCycleTime)
	assert.Equal(t, estimatedCommuteTime, *response.EstimatedCommuteTime)

	applicationToInsertApplicationDate := applicationDate.Format(time.RFC3339)
	insertedApplicationApplicationDate := response.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertApplicationDate, insertedApplicationApplicationDate)

	applicationToInsertCreatedDate := createdDate.Format(time.RFC3339)
	insertedApplicationCreatedDate := response.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertCreatedDate, insertedApplicationCreatedDate)

	applicationToInsertUpdatedDate := updatedDate.Format(time.RFC3339)
	insertedApplicationUpdatedDate := response.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertUpdatedDate, insertedApplicationUpdatedDate)
}

func TestNewApplicationResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewApplicationResponse(nil)
	assert.Nil(t, nilModel)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, err.Error(), "internal service error: Error building response: Application is nil")
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
			JobTitle:         testutil.ToPtr("Job Title "),
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
