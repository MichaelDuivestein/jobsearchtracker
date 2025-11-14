package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"

	"github.com/google/uuid"
)

type AssociateApplicationEventRequest struct {
	ApplicationID uuid.UUID `json:"application_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventID       uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *AssociateApplicationEventRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("CreateApplicationEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.ApplicationID == uuid.Nil {
		message := "ApplicationID is invalid"
		slog.Info("CreateApplicationEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventID == uuid.Nil {
		message := "EventID is invalid"
		slog.Info("CreateApplicationEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *AssociateApplicationEventRequest) ToModel() (*models.AssociateApplicationEvent, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.AssociateApplicationEvent{
		ApplicationID: request.ApplicationID,
		EventID:       request.EventID,
	}

	return &model, nil
}

type DeleteApplicationEventRequest struct {
	ApplicationID uuid.UUID `json:"application_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventID       uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *DeleteApplicationEventRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("DeleteApplicationEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.ApplicationID == uuid.Nil {
		message := "ApplicationID is invalid"
		slog.Info("DeleteApplicationEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventID == uuid.Nil {
		message := "EventID is invalid"
		slog.Info("DeleteApplicationEventRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *DeleteApplicationEventRequest) ToModel() (*models.DeleteApplicationEvent, error) {
	if request == nil {
		return nil, nil
	}

	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.DeleteApplicationEvent{
		ApplicationID: request.ApplicationID,
		EventID:       request.EventID,
	}

	return &model, nil
}
