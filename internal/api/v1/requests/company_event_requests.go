package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"

	"github.com/google/uuid"
)

type AssociateCompanyEventRequest struct {
	CompanyID uuid.UUID `json:"company_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventID   uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *AssociateCompanyEventRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("CreateCompanyEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.CompanyID == uuid.Nil {
		message := "CompanyID is invalid"
		slog.Info("CreateCompanyEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventID == uuid.Nil {
		message := "EventID is invalid"
		slog.Info("CreateCompanyEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *AssociateCompanyEventRequest) ToModel() (*models.AssociateCompanyEvent, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.AssociateCompanyEvent{
		CompanyID: request.CompanyID,
		EventID:   request.EventID,
	}

	return &model, nil
}

type DeleteCompanyEventRequest struct {
	CompanyID uuid.UUID `json:"company_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventID   uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *DeleteCompanyEventRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("DeleteCompanyEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.CompanyID == uuid.Nil {
		message := "CompanyID is invalid"
		slog.Info("DeleteCompanyEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventID == uuid.Nil {
		message := "EventID is invalid"
		slog.Info("DeleteCompanyEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *DeleteCompanyEventRequest) ToModel() (*models.DeleteCompanyEvent, error) {
	if request == nil {
		return nil, nil
	}

	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.DeleteCompanyEvent{
		CompanyID: request.CompanyID,
		EventID:   request.EventID,
	}

	return &model, nil
}
