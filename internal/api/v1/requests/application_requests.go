package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// CreateApplicationRequest represents a request to create an application.
//
// Either CompanyID or RecruiterID (or both) must be provided.
type CreateApplicationRequest struct {
	ID                   *uuid.UUID       `json:"id,omitempty" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=00"`
	CompanyID            *uuid.UUID       `json:"company_id,omitempty" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=01"`
	RecruiterID          *uuid.UUID       `json:"recruiter_id,omitempty" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=02"`
	JobTitle             *string          `json:"job_title,omitempty" example:"Job Title" extensions:"x-order=03"`
	JobAdURL             *string          `json:"job_ad_url,omitempty" example:"https://job.ad.url" extensions:"x-order=04"`
	Country              *string          `json:"country,omitempty" example:"Sweden" extensions:"x-order=05"`
	Area                 *string          `json:"area,omitempty" example:"Stockholm" extensions:"x-order=06"`
	RemoteStatusType     RemoteStatusType `json:"remote_status_type" example:"hybrid" extensions:"x-order=07"`
	WeekdaysInOffice     *int             `json:"weekdays_in_office,omitempty" example:"2" extensions:"x-order=08"`
	EstimatedCycleTime   *int             `json:"estimated_cycle_time,omitempty" example:"25" extensions:"x-order=09"`
	EstimatedCommuteTime *int             `json:"estimated_commute_time,omitempty" example:"35" extensions:"x-order=10"`
	ApplicationDate      *time.Time       `json:"application_date,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=11"`
}

func (request *CreateApplicationRequest) validate() error {
	if request.ID != nil {
		err := uuid.Validate(request.ID.String())
		if err != nil {
			message := "ID is invalid"
			slog.Info("CreateApplicationRequest.validate failed: " + message)
			return internalErrors.NewValidationError(nil, message)
		} else if *request.ID == uuid.Nil {
			message := "ID is empty"
			slog.Info("CreateApplicationRequest.Validate: "+message, "ID", request.ID)
			return internalErrors.NewValidationError(nil, message)
		}
	}

	if request.CompanyID != nil {
		err := uuid.Validate(request.CompanyID.String())
		if err != nil {
			return internalErrors.NewValidationError(nil, "CompanyID is invalid")
		}
	}

	if request.RecruiterID != nil {
		err := uuid.Validate(request.RecruiterID.String())
		if err != nil {
			return internalErrors.NewValidationError(nil, "RecruiterID is invalid")
		}
	}

	if (request.CompanyID == nil || *request.CompanyID == uuid.Nil) &&
		(request.RecruiterID == nil || *request.RecruiterID == uuid.Nil) {
		return internalErrors.NewValidationError(nil, "CompanyID and RecruiterID cannot both be empty")
	}

	if request.JobTitle != nil && *request.JobTitle == "" {
		return internalErrors.NewValidationError(nil, "JobTitle is empty")
	}

	if request.JobAdURL != nil && *request.JobAdURL == "" {
		return internalErrors.NewValidationError(nil, "JobAdURL is empty")
	}

	if request.JobTitle == nil && request.JobAdURL == nil {
		return internalErrors.NewValidationError(nil, "JobTitle and JobAdURL cannot be both be empty")
	}

	if !request.RemoteStatusType.IsValid() {
		message := "RemoteStatusType is invalid"
		slog.Info("CreateApplicationRequest.validate failed: " + message)
		companyType := "RemoteStatusType"
		return internalErrors.NewValidationError(&companyType, message)
	}

	if request.ApplicationDate != nil && request.ApplicationDate.IsZero() {
		updatedDate := "ApplicationDate"
		return internalErrors.NewValidationError(
			&updatedDate,
			"ApplicationDate is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	return nil
}

func (request *CreateApplicationRequest) ToModel() (*models.CreateApplication, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	remoteStatusType, _ := request.RemoteStatusType.ToModel()

	applicationModel := models.CreateApplication{
		ID:                   request.ID,
		CompanyID:            request.CompanyID,
		RecruiterID:          request.RecruiterID,
		JobTitle:             request.JobTitle,
		JobAdURL:             request.JobAdURL,
		Country:              request.Country,
		Area:                 request.Area,
		RemoteStatusType:     remoteStatusType,
		WeekdaysInOffice:     request.WeekdaysInOffice,
		EstimatedCycleTime:   request.EstimatedCycleTime,
		EstimatedCommuteTime: request.EstimatedCommuteTime,
		ApplicationDate:      request.ApplicationDate,
	}

	return &applicationModel, nil
}

type UpdateApplicationRequest struct {
	ID                   uuid.UUID         `json:"id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=00"`
	CompanyID            *uuid.UUID        `json:"company_id,omitempty" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=01"`
	RecruiterID          *uuid.UUID        `json:"recruiter_id,omitempty" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=02"`
	JobTitle             *string           `json:"job_title,omitempty" example:"Job Title" extensions:"x-order=03"`
	JobAdURL             *string           `json:"job_ad_url,omitempty" example:"https://job.ad.url" extensions:"x-order=04"`
	Country              *string           `json:"country,omitempty" example:"Sweden" extensions:"x-order=05"`
	Area                 *string           `json:"area,omitempty" example:"Stockholm" extensions:"x-order=06"`
	RemoteStatusType     *RemoteStatusType `json:"remote_status_type,omitempty" example:"hybrid" extensions:"x-order=07"`
	WeekdaysInOffice     *int              `json:"weekdays_in_office,omitempty" example:"2" extensions:"x-order=08"`
	EstimatedCycleTime   *int              `json:"estimated_cycle_time,omitempty" example:"25" extensions:"x-order=09"`
	EstimatedCommuteTime *int              `json:"estimated_commute_time,omitempty" example:"35" extensions:"x-order=10"`
	ApplicationDate      *time.Time        `json:"application_date,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=11"`
}

// Validate can return ValidationError
func (request *UpdateApplicationRequest) validate() error {
	err := uuid.Validate(request.ID.String())
	if err != nil {
		message := "ID is invalid"
		slog.Info("UpdateApplicationRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}
	if request.ID == uuid.Nil {
		message := "ID is empty"
		slog.Info("UpdateApplicationRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.CompanyID == nil && request.RecruiterID == nil && request.JobTitle == nil && request.JobAdURL == nil &&
		request.Country == nil && request.Area == nil && request.RemoteStatusType == nil &&
		request.WeekdaysInOffice == nil && request.EstimatedCycleTime == nil && request.EstimatedCommuteTime == nil &&
		request.ApplicationDate == nil {
		message := "nothing to update"
		slog.Info("UpdateApplicationRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.RemoteStatusType != nil && !request.RemoteStatusType.IsValid() {
		message := "RemoteStatusType is invalid"

		slog.Info("UpdateApplicationRequest.Validate: "+message, "ID", request.ID)

		remoteStatusType := "RemoteStatusType"
		return internalErrors.NewValidationError(&remoteStatusType, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *UpdateApplicationRequest) ToModel() (*models.UpdateApplication, error) {
	// can return ValidationError
	err := request.validate()
	if err != nil {
		slog.Info("validate updateApplicationRequest failed", "error", err)
		return nil, err
	}

	var remoteStatusType *models.RemoteStatusType
	if request.RemoteStatusType != nil {
		// can return ValidationError
		tempRemoteStatusType, _ := request.RemoteStatusType.ToModel()
		remoteStatusType = &tempRemoteStatusType
	} else {
		remoteStatusType = nil
	}

	updateModel := models.UpdateApplication{
		ID:                   request.ID,
		CompanyID:            request.CompanyID,
		RecruiterID:          request.RecruiterID,
		JobTitle:             request.JobTitle,
		JobAdURL:             request.JobAdURL,
		Country:              request.Country,
		Area:                 request.Area,
		RemoteStatusType:     remoteStatusType,
		WeekdaysInOffice:     request.WeekdaysInOffice,
		EstimatedCycleTime:   request.EstimatedCycleTime,
		EstimatedCommuteTime: request.EstimatedCommuteTime,
		ApplicationDate:      request.ApplicationDate,
	}

	return &updateModel, nil
}

// RemoteStatusType represents how an employer allows remote work
//
// @enum hybrid,office,remote,unknown
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

func (remoteStatusType RemoteStatusType) String() string { return string(remoteStatusType) }

// ToModel can return ValidationError
func (remoteStatusType RemoteStatusType) ToModel() (models.RemoteStatusType, error) {
	switch remoteStatusType {
	case RemoteStatusTypeHybrid:
		return models.RemoteStatusTypeHybrid, nil
	case RemoteStatusTypeOffice:
		return models.RemoteStatusTypeOffice, nil
	case RemoteStatusTypeRemote:
		return models.RemoteStatusTypeRemote, nil
	case RemoteStatusTypeUnknown:
		return models.RemoteStatusTypeUnknown, nil
	default:
		slog.Info("v1.types.toModel: Invalid RemoteStatusType: '" + remoteStatusType.String() + "'")
		remoteStatusTypeString := "RemoteStatusType"
		return "", internalErrors.NewValidationError(
			&remoteStatusTypeString,
			"invalid RemoteStatusType: '"+remoteStatusType.String()+"'")
	}
}

func NewRemoteStatusType(modelRemoteStatusType *models.RemoteStatusType) (RemoteStatusType, error) {
	if modelRemoteStatusType == nil {
		slog.Info("v1.types.NewRemoteStatusType: modelRemoteStatusType is nil")
		return "", internalErrors.NewInternalServiceError(
			"Error trying to convert internal RemoteStatusType to external RemoteStatusType.")
	}

	switch *modelRemoteStatusType {
	case models.RemoteStatusTypeHybrid:
		return RemoteStatusTypeHybrid, nil
	case models.RemoteStatusTypeOffice:
		return RemoteStatusTypeOffice, nil
	case models.RemoteStatusTypeRemote:
		return RemoteStatusTypeRemote, nil
	case models.RemoteStatusTypeUnknown:
		return RemoteStatusTypeUnknown, nil

	default:
		slog.Info("v1.types.NewRemoteStatusType: Invalid modelRemoteStatusType: '" + modelRemoteStatusType.String() + "'")
		return "", internalErrors.NewInternalServiceError(
			"Error converting internal RemoteStatusType to external RemoteStatusType: '" + modelRemoteStatusType.String() + "'")
	}
}

func (remoteStatusType RemoteStatusType) ToPointer() *RemoteStatusType {
	return &remoteStatusType
}
