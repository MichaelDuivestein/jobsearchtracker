package responses

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewApplicationResponse tests: --------

func TestNewApplicationResponse_ShouldWork(t *testing.T) {
	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Job Country"
	area := "Job Area"
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
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          createdDate,
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

func TestNewApplicationResponse_ShouldWorkWithOnlyRequiredFields(t *testing.T) {

	companyID := uuid.New()
	jobAdURL := "Job Ad URL"

	model := models.Application{
		ID:               uuid.New(),
		CompanyID:        &companyID,
		JobAdURL:         &jobAdURL,
		RemoteStatusType: models.RemoteStatusTypeRemote,
		CreatedDate:      time.Now().AddDate(0, 3, 0),
	}

	response, err := NewApplicationResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID.String(), response.ID.String())
	assert.Equal(t, model.CompanyID, response.CompanyID)
	assert.Nil(t, model.RecruiterID)
	assert.Nil(t, response.JobTitle)
	assert.Equal(t, model.JobAdURL, response.JobAdURL)
	assert.Nil(t, response.Country)
	assert.Nil(t, response.Area)
	assert.Equal(t, model.RemoteStatusType.String(), response.RemoteStatusType.String())
	assert.Nil(t, response.WeekdaysInOffice)
	assert.Nil(t, response.EstimatedCycleTime)
	assert.Nil(t, response.EstimatedCommuteTime)
	assert.Nil(t, response.ApplicationDate)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Nil(t, response.UpdatedDate)
}

func TestNewApplicationResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewApplicationResponse(nil)
	assert.Nil(t, nilModel)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, err.Error(), "internal service error: Error building response: Application is nil")
}

func TestNewApplicationResponse_ShouldReturnInternalServiceErrorIfRemoteStatusTypeIsInvalid(t *testing.T) {
	recruiterID := uuid.New()
	JobAdURL := "Job Ad URL"

	emptyRemoteStatusType := models.Application{
		ID:               uuid.New(),
		RecruiterID:      &recruiterID,
		JobAdURL:         &JobAdURL,
		RemoteStatusType: "",
		CreatedDate:      time.Now().AddDate(0, 0, 16),
	}
	emptyResponse, err := NewApplicationResponse(&emptyRemoteStatusType)
	assert.Nil(t, emptyResponse)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: ''",
		err.Error())

	invalidRemoteStatusType := models.Application{
		ID:               uuid.New(),
		RecruiterID:      &recruiterID,
		JobAdURL:         &JobAdURL,
		RemoteStatusType: "Blah",
		CreatedDate:      time.Now().AddDate(0, 0, 16),
	}
	invalidResponse, err := NewApplicationResponse(&invalidRemoteStatusType)
	assert.Nil(t, invalidResponse)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: 'Blah'",
		err.Error())
}
