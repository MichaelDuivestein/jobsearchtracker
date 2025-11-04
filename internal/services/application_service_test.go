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
	assert.Equal(t, "validation error: CreateApplication is nil", validationError.Error())
}

func TestCreateApplication_ShouldReturnValidationErrorOnNilCompanyIDAndNilRecruiterID(t *testing.T) {
	applicationService := NewApplicationService(nil)

	application := models.CreateApplication{
		CompanyID:        nil,
		RecruiterID:      nil,
		JobTitle:         testutil.ToPtr("JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}

	nilApplication, err := applicationService.CreateApplication(&application)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: CompanyID and RecruiterID cannot both be empty", validationError.Error())
}

func TestCreateApplication_ShouldReturnValidationErrorOnNilOrEmptyCompanyIDAndNilOrEmptyJobAdURL(t *testing.T) {
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
			testutil.ToPtr(""),
			"validation error: JobAdURL is empty",
		},
		{
			"empty companyID and nil JobAdURL",
			testutil.ToPtr(""),
			nil,
			"validation error: JobTitle is empty",
		},
		{
			"empty companyID and empty JobAdURL",
			testutil.ToPtr(""),
			testutil.ToPtr(""),
			"validation error: JobTitle is empty",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			application := models.CreateApplication{
				CompanyID:        testutil.ToPtr(uuid.New()),
				JobTitle:         test.jobTitle,
				JobAdURL:         test.jobAdURL,
				RemoteStatusType: models.RemoteStatusTypeRemote,
			}
			nilApplication, err := applicationService.CreateApplication(&application)
			assert.Nil(t, nilApplication)
			assert.NotNil(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.errorMessage, validationError.Error())
		})
	}
}

func TestCreateApplication_ShouldReturnValidationErrorOnInvalidRemoteStatusType(t *testing.T) {
	applicationService := NewApplicationService(nil)

	var remoteStatusType models.RemoteStatusType = "Not Valid"
	application := models.CreateApplication{
		RecruiterID:      testutil.ToPtr(uuid.New()),
		JobAdURL:         testutil.ToPtr("jobAdURL"),
		RemoteStatusType: remoteStatusType,
	}
	nilApplication, err := applicationService.CreateApplication(&application)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: remoteStatusType is invalid", validationError.Error())
}

func TestCreateApplication_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	applicationService := NewApplicationService(nil)

	application := models.CreateApplication{
		CompanyID:        testutil.ToPtr(uuid.New()),
		JobTitle:         testutil.ToPtr("JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeRemote,
		UpdatedDate:      &time.Time{},
	}
	nilApplication, err := applicationService.CreateApplication(&application)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'UpdatedDate': UpdatedDate is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

// -------- GetApplicationById tests: --------

func TestGetApplicationById_ShouldReturnValidationErrorIfApplicationIdIsNil(t *testing.T) {
	applicationService := NewApplicationService(nil)

	nilApplication, err := applicationService.GetApplicationById(nil)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'application ID': applicationId is required", validationError.Error())
}

// -------- GetApplicationsByJobTitle tests: --------

func TestGetApplicationsByJobTitle_ShouldReturnValidationErrorIfJobTitleIsNil(t *testing.T) {
	applicationService := NewApplicationService(nil)

	nilApplication, err := applicationService.GetApplicationsByJobTitle(nil)
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'applicationJobTitle': applicationJobTitle is required",
		validationError.Error())
}

func TestGetApplicationsByJobTitle_ShouldReturnValidationErrorIfJobTitleIsEmpty(t *testing.T) {
	applicationService := NewApplicationService(nil)

	nilApplication, err := applicationService.GetApplicationsByJobTitle(testutil.ToPtr(""))
	assert.Nil(t, nilApplication)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'applicationJobTitle': applicationJobTitle is required",
		validationError.Error())
}

// -------- UpdateApplication tests: --------

func TestUpdateApplication_ShouldReturnValidationErrorIfApplicationIsNil(t *testing.T) {
	applicationService := NewApplicationService(nil)

	err := applicationService.UpdateApplication(nil)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: UpdateApplication model is nil", validationError.Error())
}

func TestUpdateApplication_ShouldReturnValidationErrorIfApplicationContainsNothingToUpdate(t *testing.T) {
	applicationService := NewApplicationService(nil)

	application := models.UpdateApplication{
		ID: uuid.New(),
	}

	err := applicationService.UpdateApplication(&application)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())

}

// -------- DeleteApplication tests: --------

func TestDeleteApplication_ShouldReturnValidationErrorIfApplicationIdIsNil(t *testing.T) {
	applicationService := NewApplicationService(nil)

	err := applicationService.DeleteApplication(nil)
	assert.NotNil(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'application ID': applicationId is required", validationError.Error())
}
