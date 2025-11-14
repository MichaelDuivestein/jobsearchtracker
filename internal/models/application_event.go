package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type ApplicationEvent struct {
	ApplicationID uuid.UUID
	EventID       uuid.UUID
	CreatedDate   time.Time
}

type AssociateApplicationEvent struct {
	ApplicationID uuid.UUID
	EventID       uuid.UUID
	CreatedDate   *time.Time
}

// Validate can return ValidationError
func (applicationEvent *AssociateApplicationEvent) Validate() error {
	if applicationEvent.ApplicationID == uuid.Nil {
		return errors.NewValidationError(nil, "ApplicationID is empty")
	}

	if applicationEvent.EventID == uuid.Nil {
		return errors.NewValidationError(nil, "EventID is empty")
	}

	return nil
}

type DeleteApplicationEvent struct {
	ApplicationID uuid.UUID
	EventID       uuid.UUID
}

// Validate can return ValidationError
func (applicationEvent *DeleteApplicationEvent) Validate() error {
	if applicationEvent.ApplicationID == uuid.Nil {
		return errors.NewValidationError(nil, "ApplicationID cannot be empty")
	}

	if applicationEvent.EventID == uuid.Nil {
		return errors.NewValidationError(nil, "EventID cannot be empty")
	}

	return nil
}
