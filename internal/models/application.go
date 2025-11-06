package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID                   uuid.UUID
	CompanyID            *uuid.UUID
	RecruiterID          *uuid.UUID
	JobTitle             *string
	JobAdURL             *string
	Country              *string
	Area                 *string
	RemoteStatusType     *RemoteStatusType
	WeekdaysInOffice     *int
	EstimatedCycleTime   *int
	EstimatedCommuteTime *int
	ApplicationDate      *time.Time
	CreatedDate          *time.Time
	UpdatedDate          *time.Time
	Company              *Company
	Recruiter            *Company
	Persons              *[]*Person
}

type CreateApplication struct {
	ID                   *uuid.UUID
	CompanyID            *uuid.UUID
	RecruiterID          *uuid.UUID
	JobTitle             *string
	JobAdURL             *string
	Country              *string
	Area                 *string
	RemoteStatusType     RemoteStatusType
	WeekdaysInOffice     *int
	EstimatedCycleTime   *int
	EstimatedCommuteTime *int
	ApplicationDate      *time.Time
	CreatedDate          *time.Time
	UpdatedDate          *time.Time
}

func (application *CreateApplication) Validate() error {

	if application.CompanyID != nil {
		err := uuid.Validate(application.CompanyID.String())
		if err != nil {
			return errors.NewValidationError(nil, "CompanyID is invalid")
		}
	}

	if application.RecruiterID != nil {
		err := uuid.Validate(application.RecruiterID.String())
		if err != nil {
			return errors.NewValidationError(nil, "RecruiterID is invalid")
		}
	}

	if (application.CompanyID == nil || *application.CompanyID == uuid.Nil) &&
		(application.RecruiterID == nil || *application.RecruiterID == uuid.Nil) {
		return errors.NewValidationError(nil, "CompanyID and RecruiterID cannot both be empty")
	}

	if application.JobTitle != nil && *application.JobTitle == "" {
		return errors.NewValidationError(nil, "JobTitle is empty")
	}

	if application.JobAdURL != nil && *application.JobAdURL == "" {
		return errors.NewValidationError(nil, "JobAdURL is empty")
	}

	if application.JobTitle == nil && application.JobAdURL == nil {
		return errors.NewValidationError(nil, "JobTitle and JobAdURL cannot be both be nil")
	}

	if !application.RemoteStatusType.IsValid() {
		return errors.NewValidationError(nil, "remoteStatusType is invalid")
	}

	if application.ApplicationDate != nil && application.ApplicationDate.IsZero() {
		updatedDate := "ApplicationDate"
		return errors.NewValidationError(
			&updatedDate,
			"ApplicationDate is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	if application.UpdatedDate != nil && application.UpdatedDate.IsZero() {
		updatedDate := "UpdatedDate"
		return errors.NewValidationError(
			&updatedDate,
			"UpdatedDate is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	return nil
}

type UpdateApplication struct {
	ID                   uuid.UUID
	CompanyID            *uuid.UUID
	RecruiterID          *uuid.UUID
	JobTitle             *string
	JobAdURL             *string
	Country              *string
	Area                 *string
	RemoteStatusType     *RemoteStatusType
	WeekdaysInOffice     *int
	EstimatedCycleTime   *int
	EstimatedCommuteTime *int
	ApplicationDate      *time.Time
}

func (application *UpdateApplication) Validate() error {
	if (application.CompanyID == nil || *application.CompanyID == uuid.Nil) &&
		(application.RecruiterID == nil || *application.RecruiterID == uuid.Nil) &&
		application.JobTitle == nil && application.JobAdURL == nil && application.Country == nil &&
		application.Area == nil && application.RemoteStatusType == nil && application.WeekdaysInOffice == nil &&
		application.EstimatedCycleTime == nil && application.EstimatedCommuteTime == nil &&
		application.ApplicationDate == nil {
		return errors.NewValidationError(nil, "nothing to update")
	}

	return nil
}

type RemoteStatusType string

const (
	RemoteStatusTypeHybrid  = "hybrid"
	RemoteStatusTypeOffice  = "office"
	RemoteStatusTypeRemote  = "remote"
	RemoteStatusTypeUnknown = "unknown"
)

func (remoteStatusType RemoteStatusType) IsValid() bool {
	switch remoteStatusType {
	case RemoteStatusTypeHybrid, RemoteStatusTypeOffice, RemoteStatusTypeRemote, RemoteStatusTypeUnknown:
		return true
	}
	return false
}

func (remoteStatusType RemoteStatusType) String() string {
	return string(remoteStatusType)
}

func (remoteStatusType RemoteStatusType) ToPtr() *RemoteStatusType {
	return &remoteStatusType
}
