package responses

import (
	"jobsearchtracker/internal/api/v1/requests"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type ApplicationResponse struct {
	ID                   uuid.UUID                  `json:"id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=00"`
	CompanyID            *uuid.UUID                 `json:"company_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=01"`
	RecruiterID          *uuid.UUID                 `json:"recruiter_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=02"`
	JobTitle             *string                    `json:"job_title,omitempty" example:"Job Title" extensions:"x-order=03"`
	JobAdURL             *string                    `json:"job_ad_url,omitempty" example:"https://job.ad.url" extensions:"x-order=04"`
	Country              *string                    `json:"country,omitempty" example:"Sweden" extensions:"x-order=05"`
	Area                 *string                    `json:"area,omitempty" example:"Stockholm" extensions:"x-order=06"`
	RemoteStatusType     *requests.RemoteStatusType `json:"remote_status_type" example:"hybrid" extensions:"x-order=07"`
	WeekdaysInOffice     *int                       `json:"weekdays_in_office,omitempty" example:"2" extensions:"x-order=08"`
	EstimatedCycleTime   *int                       `json:"estimated_cycle_time,omitempty" example:"25" extensions:"x-order=09"`
	EstimatedCommuteTime *int                       `json:"estimated_commute_time,omitempty" example:"35" extensions:"x-order=10"`
	ApplicationDate      *time.Time                 `json:"application_date,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=11"`
	CreatedDate          *time.Time                 `json:"created_date" example:"2025-12-31T23:59Z" extensions:"x-order=12"`
	UpdatedDate          *time.Time                 `json:"updated_date" example:"2025-12-31T23:59Z" extensions:"x-order=13"`
}

// NewApplicationResponse can return InternalServerError
func NewApplicationResponse(applicationModel *models.Application) (*ApplicationResponse, error) {
	if applicationModel == nil {
		slog.Error("responses.NewApplicationResponse: Application is nil")
		return nil, internalErrors.NewInternalServiceError("Error building response: Application is nil")
	}

	var remoteStatusType *requests.RemoteStatusType = nil
	if applicationModel.RemoteStatusType != nil {
		nonNilRemoteStatusType, err := requests.NewRemoteStatusType(applicationModel.RemoteStatusType)
		if err != nil {
			return nil, err
		}
		remoteStatusType = &nonNilRemoteStatusType
	}

	applicationResponse := ApplicationResponse{
		ID:                   applicationModel.ID,
		CompanyID:            applicationModel.CompanyID,
		RecruiterID:          applicationModel.RecruiterID,
		JobTitle:             applicationModel.JobTitle,
		JobAdURL:             applicationModel.JobAdURL,
		Country:              applicationModel.Country,
		Area:                 applicationModel.Area,
		RemoteStatusType:     remoteStatusType,
		WeekdaysInOffice:     applicationModel.WeekdaysInOffice,
		EstimatedCycleTime:   applicationModel.EstimatedCycleTime,
		EstimatedCommuteTime: applicationModel.EstimatedCommuteTime,
		ApplicationDate:      applicationModel.ApplicationDate,
		CreatedDate:          applicationModel.CreatedDate,
		UpdatedDate:          applicationModel.UpdatedDate,
	}

	return &applicationResponse, nil
}

// NewApplicationsResponse can return InternalServerError
func NewApplicationsResponse(applications []*models.Application) ([]*ApplicationResponse, error) {
	if applications == nil || len(applications) == 0 {
		return []*ApplicationResponse{}, nil
	}

	var applicationResponses = make([]*ApplicationResponse, len(applications))
	for index := range applications {
		applicationResponse, err := NewApplicationResponse(applications[index])
		if err != nil {
			return nil, err
		}
		applicationResponses[index] = applicationResponse
	}
	return applicationResponses, nil
}
