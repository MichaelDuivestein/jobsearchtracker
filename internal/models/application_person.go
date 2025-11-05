package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type ApplicationPerson struct {
	ApplicationID uuid.UUID
	PersonID      uuid.UUID
	CreatedDate   time.Time
}

type AssociateApplicationPerson struct {
	ApplicationID uuid.UUID
	PersonID      uuid.UUID
	CreatedDate   *time.Time
}

// Validate can return ValidationError
func (applicationPerson *AssociateApplicationPerson) Validate() error {
	if applicationPerson.ApplicationID == uuid.Nil {
		return errors.NewValidationError(nil, "ApplicationID is empty")
	}

	if applicationPerson.PersonID == uuid.Nil {
		return errors.NewValidationError(nil, "PersonID is empty")
	}

	return nil
}

type DeleteApplicationPerson struct {
	ApplicationID uuid.UUID
	PersonID      uuid.UUID
}

// Validate can return ValidationError
func (applicationPerson *DeleteApplicationPerson) Validate() error {
	if applicationPerson.ApplicationID == uuid.Nil {
		return errors.NewValidationError(nil, "ApplicationID cannot be empty")
	}

	if applicationPerson.PersonID == uuid.Nil {
		return errors.NewValidationError(nil, "PersonID cannot be empty")
	}

	return nil
}
