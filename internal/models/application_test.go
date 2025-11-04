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

// -------- CreateApplication.Validate tests: --------
func TestCreateApplicationValidate_ShouldReturnNilIfApplicationIsValid(t *testing.T) {

	id := uuid.New()
	companyID := uuid.New()
	recruiterID := uuid.New()
	jobTitle := "Job1 Title"
	jobAdUrl := "Job1 Ad URL"
	country := "Job 1 Country"
	area := "Job 1 Area"
	weekdaysInOffice := 2
	estimatedCycleTime := 30
	estimatedCommuteTime := 40
	applicationDate := time.Now().AddDate(0, 0, 1)
	createdDate := time.Now().AddDate(0, 0, 2)
	updatedDate := time.Now().AddDate(0, 0, 3)

	application := CreateApplication{
		ID:                   &id,
		CompanyID:            &companyID,
		RecruiterID:          &recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdUrl,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     RemoteStatusTypeHybrid,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
		CreatedDate:          &createdDate,
		UpdatedDate:          &updatedDate,
	}
	err := application.Validate()
	assert.NoError(t, err)
}

func TestCreateApplicationValidate_ShouldReturnNilWithOnlyRequiredFields(t *testing.T) {

	recruiterID := uuid.New()
	jobTitle := "Job1 Title"

	application := CreateApplication{
		RecruiterID:      &recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: RemoteStatusTypeHybrid,
	}
	err := application.Validate()
	assert.NoError(t, err)
}

func TestCreateApplicationValidate_ShouldReturnValidationErrorIfCompanyIDAndRecruiterIDAreNull(t *testing.T) {
	tests := []struct {
		testName    string
		companyId   *uuid.UUID
		recruiterId *uuid.UUID
	}{
		{
			testName:    "companyId is nil and recruiterId is nil",
			companyId:   nil,
			recruiterId: nil,
		},
		{
			testName:    "companyId is uuid.Nil and recruiterId is nil",
			companyId:   &uuid.Nil,
			recruiterId: nil,
		},
		{
			testName:    "companyId is nil and recruiterId is uuid.Nil",
			companyId:   nil,
			recruiterId: &uuid.Nil,
		},
		{
			testName:    "companyId is uuid.Nil and recruiterId is uuid.Nil",
			companyId:   &uuid.Nil,
			recruiterId: &uuid.Nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			application := CreateApplication{
				CompanyID:   test.companyId,
				RecruiterID: test.recruiterId,
			}
			err := application.Validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, "validation error: CompanyID and RecruiterID cannot both be empty", validationError.Error())
		})
	}
}

func TestCreateApplicationValidate_ShouldReturnValidationErrorIfJobTitleIsEmpty(t *testing.T) {
	application := CreateApplication{
		CompanyID: testutil.ToPtr(uuid.New()),
		JobTitle:  testutil.ToPtr(""),
	}
	err := application.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: JobTitle is empty", validationError.Error())
}

func TestCreateApplicationValidate_ShouldReturnValidationErrorIfJobAdUrlIsEmpty(t *testing.T) {
	application := CreateApplication{
		CompanyID: testutil.ToPtr(uuid.New()),
		JobAdURL:  testutil.ToPtr(""),
	}
	err := application.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: JobAdURL is empty", validationError.Error())
}

func TestCreateApplicationValidate_ShouldReturnValidationErrorIfJobTitleIsEmptyAndJobAdUrlIsNil(t *testing.T) {
	application := CreateApplication{
		CompanyID: testutil.ToPtr(uuid.New()),
		JobTitle:  nil,
		JobAdURL:  nil,
	}
	err := application.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: JobTitle and JobAdURL cannot be both be nil", validationError.Error())
}

func TestCreateApplicationValidate_ShouldReturnValidationErrorIfRemoteStatusIsInvalid(t *testing.T) {
	var fakeRemoteStatusType RemoteStatusType = "something that should never happen"
	application := CreateApplication{
		CompanyID:        testutil.ToPtr(uuid.New()),
		JobTitle:         testutil.ToPtr("not important"),
		RemoteStatusType: fakeRemoteStatusType,
	}
	err := application.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: remoteStatusType is invalid", validationError.Error())
}

func TestCreateApplicationValidate_ShouldReturnValidationErrorOnUnsetApplicationDate(t *testing.T) {
	application := CreateApplication{
		CompanyID:        testutil.ToPtr(uuid.New()),
		JobTitle:         testutil.ToPtr("not important"),
		RemoteStatusType: RemoteStatusTypeOffice,
		ApplicationDate:  &time.Time{},
	}
	err := application.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'ApplicationDate': ApplicationDate is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

func TestCreateApplicationValidate_ShouldReturnValidationErrorOnUnsetUpdatedDate(t *testing.T) {
	application := CreateApplication{
		CompanyID:        testutil.ToPtr(uuid.New()),
		JobTitle:         testutil.ToPtr("not important"),
		RemoteStatusType: RemoteStatusTypeUnknown,
		UpdatedDate:      &time.Time{},
	}
	err := application.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'UpdatedDate': UpdatedDate is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

// -------- RemoteStatusType.IsValid tests: --------

func TestRemoteStatusTypesValid_ShouldReturnTrue(t *testing.T) {
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

	spammer := RemoteStatusType("offshore")
	assert.False(t, spammer.IsValid())
}
