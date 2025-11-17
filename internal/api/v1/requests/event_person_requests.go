package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"

	"github.com/google/uuid"
)

type AssociateEventPersonRequest struct {
	EventID  uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *AssociateEventPersonRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("CreateEventPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventID == uuid.Nil {
		message := "EventID is invalid"
		slog.Info("CreateEventPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.PersonID == uuid.Nil {
		message := "PersonID is invalid"
		slog.Info("CreateEventPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *AssociateEventPersonRequest) ToModel() (*models.AssociateEventPerson, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.AssociateEventPerson{
		EventID:  request.EventID,
		PersonID: request.PersonID,
	}

	return &model, nil
}

type DeleteEventPersonRequest struct {
	EventID  uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *DeleteEventPersonRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("DeleteEventPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventID == uuid.Nil {
		message := "EventID is invalid"
		slog.Info("DeleteEventPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.PersonID == uuid.Nil {
		message := "PersonID is invalid"
		slog.Info("DeleteEventPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *DeleteEventPersonRequest) ToModel() (*models.DeleteEventPerson, error) {
	if request == nil {
		return nil, nil
	}

	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.DeleteEventPerson{
		EventID:  request.EventID,
		PersonID: request.PersonID,
	}

	return &model, nil
}
