package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type EventPerson struct {
	EventID     uuid.UUID
	PersonID    uuid.UUID
	CreatedDate time.Time
}

type AssociateEventPerson struct {
	EventID     uuid.UUID
	PersonID    uuid.UUID
	CreatedDate *time.Time
}

// Validate can return ValidationError
func (eventPerson *AssociateEventPerson) Validate() error {
	if eventPerson.EventID == uuid.Nil {
		return errors.NewValidationError(nil, "EventID is empty")
	}
	if eventPerson.PersonID == uuid.Nil {
		return errors.NewValidationError(nil, "PersonID is empty")
	}
	return nil
}

type DeleteEventPerson struct {
	EventID  uuid.UUID
	PersonID uuid.UUID
}

func (Delete *DeleteEventPerson) Validate() error {
	if Delete.EventID == uuid.Nil {
		return errors.NewValidationError(nil, "EventID cannot be empty")
	}
	if Delete.PersonID == uuid.Nil {
		return errors.NewValidationError(nil, "PersonID cannot be empty")
	}
	return nil
}
