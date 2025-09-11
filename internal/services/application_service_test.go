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

// -------- CreateApplication tests: --------

func TestCreateApplication_ShouldReturnValidationErrorOnNilApplication(t *testing.T) {
	applicationService := NewApplicationService(nil)

	nilApplication, err := applicationService.CreateApplication(nil)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CreateApplication is nil", err.Error())
}

func TestCreateApplication_ShouldReturnValidationErrorOnNilCompanyIDAndRecruiterID(t *testing.T) {
	applicationService := NewApplicationService(nil)

	jobTitle := "JobTitle"
	application := models.CreateApplication{
		CompanyID:        nil,
		RecruiterID:      nil,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}

	nilApplication, err := applicationService.CreateApplication(&application)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: CompanyID and RecruiterID cannot both be empty", err.Error())
}

func TestCreateApplication_ShouldReturnValidationErrorOnEmptyCompanyIDAndEmptyJobAdURL(t *testing.T) {
	applicationService := NewApplicationService(nil)

	tests := []struct {
		testName     string
		jobTitle     *string
		jobAdURL     *string
		errorMessage string
	}{
		{
			"nil companyID and nil JobAdURL",
			nil,
			nil,
			"validation error: JobTitle and JobAdURL cannot be both be nil",
		},
		{
			"nil companyID and empty JobAdURL",
			nil,
			testutil.StringPtr(""),
			"validation error: JobAdURL is empty",
		},
		{
			"empty companyID and nil JobAdURL",
			testutil.StringPtr(""),
			nil,
			"validation error: JobTitle is empty",
		},
		{
			"empty companyID and empty JobAdURL",
			testutil.StringPtr(""),
			testutil.StringPtr(""),
			"validation error: JobTitle is empty",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			companyID := uuid.New()

			application := models.CreateApplication{
				CompanyID:        &companyID,
				JobTitle:         test.jobTitle,
				JobAdURL:         test.jobAdURL,
				RemoteStatusType: models.RemoteStatusTypeRemote,
			}

			nilApplication, err := applicationService.CreateApplication(&application)
			assert.Nil(t, nilApplication)
			assert.NotNil(t, err)

			var validationErr *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationErr))
			assert.Equal(t, test.errorMessage, err.Error())
		})
	}
}

func TestCreateApplication_ShouldReturnValidationErrorOnInvalidApplicationType(t *testing.T) {
	applicationService := NewApplicationService(nil)

	recruiterID := uuid.New()
	jobAdURL := "jobAdURL"
	var remoteStatusType models.RemoteStatusType = "Not Valid"
	application := models.CreateApplication{
		RecruiterID:      &recruiterID,
		JobAdURL:         &jobAdURL,
		RemoteStatusType: remoteStatusType,
	}

	nilApplication, err := applicationService.CreateApplication(&application)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: remoteStatusType is invalid", err.Error())
}

func TestCreateApplication_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	applicationService := NewApplicationService(nil)

	companyID := uuid.New()
	jobTitle := "JobTitle"
	application := models.CreateApplication{
		CompanyID:        &companyID,
		JobTitle:         &jobTitle,
		RemoteStatusType: models.RemoteStatusTypeRemote,
		UpdatedDate:      &time.Time{},
	}

	nilApplication, err := applicationService.CreateApplication(&application)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(
		t,
		"validation error on field 'UpdatedDate': UpdatedDate is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		err.Error())
}

// -------- GetApplicationById tests: --------

func TestGetApplicationById_ShouldReturnValidationErrorIfApplicationIdIsNil(t *testing.T) {
	applicationService := NewApplicationService(nil)

	nilApplication, err := applicationService.GetApplicationById(nil)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'application ID': applicationId is required", err.Error())
}

// -------- GetApplicationsByJobTitle tests: --------

func TestGetApplicationsByJobTitle_ShouldReturnValidationErrorIfApplicationNameIsNil(t *testing.T) {
	applicationService := NewApplicationService(nil)

	nilApplication, err := applicationService.GetApplicationsByJobTitle(nil)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'applicationJobTitle': applicationJobTitle is required", err.Error())
}

func TestGetApplicationsByJobTitle_ShouldReturnValidationErrorIfApplicationNameIsEmpty(t *testing.T) {
	applicationService := NewApplicationService(nil)

	jobTitle := ""
	nilApplication, err := applicationService.GetApplicationsByJobTitle(&jobTitle)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)
	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error on field 'applicationJobTitle': applicationJobTitle is required", err.Error())
}
