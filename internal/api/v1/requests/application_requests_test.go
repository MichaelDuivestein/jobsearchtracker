package requests

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

// -------- CreateApplicationRequest tests: --------

func TestCreateApplicationRequestValidate_ShouldValidateRequest(t *testing.T) {
	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Some Country"
	area := "Some Area"
	weekdaysInOffice := 1
	estimatedCycleTime := 2
	estimatedCommuteTime := 3
	applicationDate := time.Now().AddDate(0, 0, -1)

	request := models.CreateApplication{
		ID:                   &id,
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
	}

	err := request.Validate()
	assert.NoError(t, err)
}

func TestCreateApplicationRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		CompanyID            *uuid.UUID
		RecruiterID          *uuid.UUID
		JobTitle             *string
		JobAdURL             *string
		remoteStatusType     RemoteStatusType
		expectedErrorMessage string
	}{
		{
			testName:             "nil CompanyID and nil RecruiterID",
			CompanyID:            nil,
			RecruiterID:          nil,
			JobTitle:             testutil.ToPtr("Job Title"),
			JobAdURL:             testutil.ToPtr("Job Ad URL"),
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty"},
		{
			testName:             "uuid.Nil CompanyID and nil RecruiterID",
			CompanyID:            &uuid.Nil,
			RecruiterID:          nil,
			JobTitle:             testutil.ToPtr("Job Title"),
			JobAdURL:             testutil.ToPtr("Job Ad URL"),
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty"},
		{
			testName:             "nil CompanyID and uuid.Nil RecruiterID",
			CompanyID:            nil,
			RecruiterID:          &uuid.Nil,
			JobTitle:             testutil.ToPtr("Job Title"),
			JobAdURL:             testutil.ToPtr("Job Ad URL"),
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty"},
		{
			testName:             "uuid.Nil CompanyID and uuid.Nil RecruiterID",
			CompanyID:            &uuid.Nil,
			RecruiterID:          &uuid.Nil,
			JobTitle:             testutil.ToPtr("Job Title"),
			JobAdURL:             testutil.ToPtr("Job Ad URL"),
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty"},
		{
			testName:             "nil JobTitle and nil JobAdURL",
			CompanyID:            testutil.ToPtr(uuid.New()),
			RecruiterID:          testutil.ToPtr(uuid.New()),
			JobTitle:             nil,
			JobAdURL:             nil,
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: JobTitle and JobAdURL cannot be both be empty"},
		{
			testName:             "empty JobTitle and nil JobAdURL",
			CompanyID:            testutil.ToPtr(uuid.New()),
			RecruiterID:          testutil.ToPtr(uuid.New()),
			JobTitle:             testutil.ToPtr(""),
			JobAdURL:             nil,
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: JobTitle is empty"},
		{
			testName:             "nil JobTitle and empty JobAdURL",
			CompanyID:            testutil.ToPtr(uuid.New()),
			RecruiterID:          testutil.ToPtr(uuid.New()),
			JobTitle:             nil,
			JobAdURL:             testutil.ToPtr(""),
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: JobAdURL is empty"},
		{
			testName:             "empty JobTitle and empty JobAdURL",
			CompanyID:            testutil.ToPtr(uuid.New()),
			RecruiterID:          testutil.ToPtr(uuid.New()),
			JobTitle:             testutil.ToPtr(""),
			JobAdURL:             testutil.ToPtr(""),
			remoteStatusType:     RemoteStatusTypeOffice,
			expectedErrorMessage: "validation error: JobTitle is empty"},
		{
			testName:             "empty RemoteStatusType",
			CompanyID:            testutil.ToPtr(uuid.New()),
			RecruiterID:          testutil.ToPtr(uuid.New()),
			JobTitle:             testutil.ToPtr("Job Title"),
			JobAdURL:             testutil.ToPtr("Job Ad URL"),
			remoteStatusType:     "",
			expectedErrorMessage: "validation error on field 'RemoteStatusType': RemoteStatusType is invalid"},
		{
			testName:             "invalid RemoteStatusType",
			CompanyID:            testutil.ToPtr(uuid.New()),
			RecruiterID:          testutil.ToPtr(uuid.New()),
			JobTitle:             testutil.ToPtr("Job Title"),
			JobAdURL:             testutil.ToPtr("Job Ad URL"),
			remoteStatusType:     "invalid RemoteStatusType",
			expectedErrorMessage: "validation error on field 'RemoteStatusType': RemoteStatusType is invalid"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			request := CreateApplicationRequest{
				CompanyID:        test.CompanyID,
				RecruiterID:      test.RecruiterID,
				JobTitle:         test.JobTitle,
				JobAdURL:         test.JobAdURL,
				RemoteStatusType: test.remoteStatusType,
			}

			err := request.validate()
			assert.NotNil(t, err)

			var validationErr *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationErr))

			assert.Equal(t, test.expectedErrorMessage, err.Error())
		})
	}
}

func TestCreateApplicationRequestToModel_ShouldConvertToModel(t *testing.T) {
	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Some Country"
	area := "Some Area"
	weekdaysInOffice := 1
	estimatedCycleTime := 2
	estimatedCommuteTime := 3
	applicationDate := time.Now().AddDate(0, 0, -1)

	request := CreateApplicationRequest{
		ID:                   &id,
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
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, id.String(), model.ID.String())
	assert.Equal(t, companyID.String(), model.CompanyID.String())
	assert.Equal(t, recruiterID.String(), model.RecruiterID.String())
	assert.Equal(t, jobTitle, *model.JobTitle)
	assert.Equal(t, jobAdURL, *model.JobAdURL)
	assert.Equal(t, country, *model.Country)
	assert.Equal(t, area, *model.Area)
	assert.Equal(t, request.RemoteStatusType.String(), model.RemoteStatusType.String())
	assert.Equal(t, weekdaysInOffice, *model.WeekdaysInOffice)
	assert.Equal(t, estimatedCycleTime, *model.EstimatedCycleTime)
	assert.Equal(t, estimatedCommuteTime, *model.EstimatedCommuteTime)

	reqestApplicationDate := applicationDate.Format(time.RFC3339)
	modelApplicationDate := model.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, reqestApplicationDate, modelApplicationDate)
}

func TestCreateApplicationRequestToModel_ShouldConvertToModelWithNilValues(t *testing.T) {
	recruiterID := uuid.New()
	jobAdURL := "Job Ad URL"

	request := CreateApplicationRequest{
		RecruiterID:      &recruiterID,
		JobAdURL:         &jobAdURL,
		RemoteStatusType: RemoteStatusTypeRemote,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Nil(t, model.ID)
	assert.Nil(t, model.CompanyID)
	assert.Equal(t, request.RecruiterID, model.RecruiterID)
	assert.Nil(t, model.JobTitle)
	assert.Equal(t, jobAdURL, *model.JobAdURL)
	assert.Nil(t, model.Country)
	assert.Nil(t, model.Area)
	assert.Equal(t, request.RemoteStatusType.String(), model.RemoteStatusType.String())
	assert.Nil(t, model.WeekdaysInOffice)
	assert.Nil(t, model.EstimatedCycleTime)
	assert.Nil(t, model.EstimatedCommuteTime)
	assert.Nil(t, model.ApplicationDate)
}

// --------UpdateApplicationRequest tests: --------

func TestUpdateApplicationRequestValidate_ShouldValidateRequest(t *testing.T) {
	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Some Country"
	area := "Some Area"
	var remoteStatusType RemoteStatusType = RemoteStatusTypeOffice
	weekdaysInOffice := 1
	estimatedCycleTime := 2
	estimatedCommuteTime := 3
	applicationDate := time.Now().AddDate(0, 0, -1)

	request := UpdateApplicationRequest{
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
	}

	err := request.validate()
	assert.NoError(t, err)
}

func TestUpdateApplicationRequestValidate_ShouldReturnValidationErrorIfNothingToUpdate(t *testing.T) {
	id := uuid.New()

	request := UpdateApplicationRequest{
		ID: id,
	}

	err := request.validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error: nothing to update", validationErr.Error())
}

func TestUpdateApplicationRequestToModel_ShouldReturnValidationErrorIfRemoteStatusTypeIsInvalid(t *testing.T) {
	id := uuid.New()
	var fakeRemoteStatusType RemoteStatusType = "something that should never happen"

	request := UpdateApplicationRequest{
		ID:               id,
		RemoteStatusType: &fakeRemoteStatusType,
	}

	err := request.validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error on field 'RemoteStatusType': RemoteStatusType is invalid", validationErr.Error())
}

func TestUpdateApplicationRequestToModel_ShouldConvertToModel(t *testing.T) {
	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Some Country"
	area := "Some Area"
	var remoteStatusType RemoteStatusType = RemoteStatusTypeOffice
	weekdaysInOffice := 1
	estimatedCycleTime := 2
	estimatedCommuteTime := 3
	applicationDate := time.Now().AddDate(0, 0, -1)

	request := UpdateApplicationRequest{
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
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, id.String(), model.ID.String())
	assert.Equal(t, companyID.String(), model.CompanyID.String())
	assert.Equal(t, recruiterID.String(), model.RecruiterID.String())
	assert.Equal(t, jobTitle, *model.JobTitle)
	assert.Equal(t, jobAdURL, *model.JobAdURL)
	assert.Equal(t, country, *model.Country)
	assert.Equal(t, area, *model.Area)
	assert.Equal(t, request.RemoteStatusType.String(), model.RemoteStatusType.String())
	assert.Equal(t, weekdaysInOffice, *model.WeekdaysInOffice)
	assert.Equal(t, estimatedCycleTime, *model.EstimatedCycleTime)
	assert.Equal(t, estimatedCommuteTime, *model.EstimatedCommuteTime)

	requestApplicationDate := applicationDate.Format(time.RFC3339)
	modelApplicationDate := model.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, requestApplicationDate, modelApplicationDate)
}

func TestUpdateApplicationRequestToModel_ShouldConvertToModelWithNilValues(t *testing.T) {
	id := uuid.New()
	companyID := uuid.New()

	request := UpdateApplicationRequest{
		ID:        id,
		CompanyID: &companyID,
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, id, model.ID)
	assert.Equal(t, companyID, *model.CompanyID)
	assert.Nil(t, model.RecruiterID)
	assert.Nil(t, model.JobTitle)
	assert.Nil(t, model.JobAdURL)
	assert.Nil(t, model.Country)
	assert.Nil(t, model.Area)
	assert.Nil(t, model.RemoteStatusType)
	assert.Nil(t, model.WeekdaysInOffice)
	assert.Nil(t, model.EstimatedCycleTime)
	assert.Nil(t, model.EstimatedCommuteTime)
	assert.Nil(t, model.ApplicationDate)
}

func TestUpdateApplicationRequestToModel_ShouldReturnValidationErrorIfNothingToUpdate(t *testing.T) {
	id := uuid.New()

	request := UpdateApplicationRequest{
		ID: id,
	}

	model, err := request.ToModel()
	assert.Nil(t, model)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error: nothing to update", err.Error())
}

// -------- RemoteStatusType tests: --------

func TestRemoteStatusTypeIsValid_ShouldReturnTrue(t *testing.T) {
	hybrid := RemoteStatusType(RemoteStatusTypeHybrid)
	assert.True(t, hybrid.IsValid())

	office := RemoteStatusType(RemoteStatusTypeOffice)
	assert.True(t, office.IsValid())

	remote := RemoteStatusType(RemoteStatusTypeRemote)
	assert.True(t, remote.IsValid())

	unknown := RemoteStatusType(RemoteStatusTypeUnknown)
	assert.True(t, unknown.IsValid())
}

func TestRemoteStatusTypeIsValid_ShouldReturnFalseOnInvalidRemoteStatusType(t *testing.T) {

	empty := RemoteStatusType("")
	assert.False(t, empty.IsValid())

	nothing := RemoteStatusType("Nothing")
	assert.False(t, nothing.IsValid())
}

func TestRemoteStatusTypeToModel_ShouldConvertToModel(t *testing.T) {
	tests := []struct {
		testName              string
		applicationType       RemoteStatusType
		modelRemoteStatusType models.RemoteStatusType
	}{
		{"hybrid", RemoteStatusTypeHybrid, models.RemoteStatusTypeHybrid},
		{"office", RemoteStatusTypeOffice, models.RemoteStatusTypeOffice},
		{"remote", RemoteStatusTypeRemote, models.RemoteStatusTypeRemote},
		{"Unknown", RemoteStatusTypeUnknown, models.RemoteStatusTypeUnknown},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			modelRemoteStatusType, err := test.applicationType.ToModel()
			assert.NoError(t, err)
			assert.NotNil(t, modelRemoteStatusType)
			assert.Equal(t, test.applicationType.String(), modelRemoteStatusType.String())
		})
	}
}

func TestRemoteStatusTypeToModel_ShouldReturnValidationErrorOnInvalidRemoteStatusType(t *testing.T) {
	empty := RemoteStatusType("")
	emptyModel, err := empty.ToModel()
	assert.NotNil(t, emptyModel)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "", emptyModel.String())
	assert.Equal(t, "validation error on field 'RemoteStatusType': invalid RemoteStatusType: ''", err.Error())

	blah := RemoteStatusType("Blah")
	blahModel, err := blah.ToModel()
	assert.NotNil(t, blahModel)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "", blahModel.String())
	assert.Equal(t, "validation error on field 'RemoteStatusType': invalid RemoteStatusType: 'Blah'", err.Error())
}

func TestNewRemoteStatusType_ShouldConvertFromModel(t *testing.T) {
	tests := []struct {
		testName              string
		modelRemoteStatusType models.RemoteStatusType
		applicationType       RemoteStatusType
	}{
		{"hybrid", models.RemoteStatusTypeHybrid, RemoteStatusTypeHybrid},
		{"office", models.RemoteStatusTypeOffice, RemoteStatusTypeOffice},
		{"remote", models.RemoteStatusTypeRemote, RemoteStatusTypeRemote},
		{"Unknown", models.RemoteStatusTypeUnknown, RemoteStatusTypeUnknown},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			applicationType, err := NewRemoteStatusType(&test.modelRemoteStatusType)
			assert.NoError(t, err)
			assert.NotNil(t, applicationType)
			assert.Equal(t, test.applicationType.String(), applicationType.String())
		})
	}
}

func TestRemoteStatusTypeToModel_ShouldReturnInternalServiceErrorOnNilRemoteStatusType(t *testing.T) {
	applicationType, err := NewRemoteStatusType(nil)
	assert.NotNil(t, applicationType)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, "", applicationType.String())
	assert.Equal(
		t,
		"internal service error: Error trying to convert internal RemoteStatusType to external RemoteStatusType.",
		err.Error())
}

func TestRemoteStatusTypeToModel_ShouldReturnInternalServiceErrorOnInvalidRemoteStatusType(t *testing.T) {
	emptyModel := models.RemoteStatusType("")
	emptyApplication, err := NewRemoteStatusType(&emptyModel)
	assert.NotNil(t, err)
	assert.NotNil(t, emptyApplication)
	assert.Equal(t, "", emptyApplication.String())

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, "", emptyApplication.String())

	specialistModel := models.RemoteStatusType("specialist")
	specialist, err := NewRemoteStatusType(&specialistModel)
	assert.NotNil(t, err)
	assert.NotNil(t, specialist)
	assert.Equal(t, "", specialist.String())

	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal RemoteStatusType to external RemoteStatusType: 'specialist'",
		err.Error())
}
