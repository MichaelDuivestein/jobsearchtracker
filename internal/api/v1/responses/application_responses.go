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
	ID                   uuid.UUID                  `json:"id,omitempty"`
	CompanyID            *uuid.UUID                 `json:"company_id,omitempty"`
	RecruiterID          *uuid.UUID                 `json:"recruiter_id,omitempty"`
	JobTitle             *string                    `json:"job_title,omitempty"`
	JobAdURL             *string                    `json:"job_ad_url,omitempty"`
	Country              *string                    `json:"country,omitempty"`
	Area                 *string                    `json:"area,omitempty"`
	RemoteStatusType     *requests.RemoteStatusType `json:"remote_status_type"`
	WeekdaysInOffice     *int                       `json:"weekdays_in_office,omitempty"`
	EstimatedCycleTime   *int                       `json:"estimated_cycle_time,omitempty"`
	EstimatedCommuteTime *int                       `json:"estimated_commute_time,omitempty"`
	ApplicationDate      *time.Time                 `json:"application_date,omitempty"`
	CreatedDate          *time.Time                 `json:"created_date"`
	UpdatedDate          *time.Time                 `json:"updated_date"`
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
