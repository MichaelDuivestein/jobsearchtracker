package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"

	"github.com/google/uuid"
)

type AssociateApplicationPersonRequest struct {
	ApplicationID uuid.UUID `json:"application_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID      uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *AssociateApplicationPersonRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("CreateApplicationPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.ApplicationID == uuid.Nil {
		message := "ApplicationID is invalid"
		slog.Info("CreateApplicationPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.PersonID == uuid.Nil {
		message := "PersonID is invalid"
		slog.Info("CreateApplicationPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *AssociateApplicationPersonRequest) ToModel() (*models.AssociateApplicationPerson, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.AssociateApplicationPerson{
		ApplicationID: request.ApplicationID,
		PersonID:      request.PersonID,
	}

	return &model, nil
}

type DeleteApplicationPersonRequest struct {
	ApplicationID uuid.UUID `json:"application_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID      uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *DeleteApplicationPersonRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("DeleteApplicationPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.ApplicationID == uuid.Nil {
		message := "ApplicationID is invalid"
		slog.Info("DeleteApplicationPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.PersonID == uuid.Nil {
		message := "PersonID is invalid"
		slog.Info("DeleteApplicationPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *DeleteApplicationPersonRequest) ToModel() (*models.DeleteApplicationPerson, error) {
	if request == nil {
		return nil, nil
	}

	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.DeleteApplicationPerson{
		ApplicationID: request.ApplicationID,
		PersonID:      request.PersonID,
	}

	return &model, nil
}
