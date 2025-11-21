package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type CompanyEvent struct {
	CompanyID   uuid.UUID
	EventID     uuid.UUID
	CreatedDate time.Time
}

type AssociateCompanyEvent struct {
	CompanyID   uuid.UUID
	EventID     uuid.UUID
	CreatedDate *time.Time
}

// Validate can return ValidationError
func (companyEvent *AssociateCompanyEvent) Validate() error {
	if companyEvent.CompanyID == uuid.Nil {
		return errors.NewValidationError(nil, "CompanyID is empty")
	}

	if companyEvent.EventID == uuid.Nil {
		return errors.NewValidationError(nil, "EventID is empty")
	}

	return nil
}

type DeleteCompanyEvent struct {
	CompanyID uuid.UUID
	EventID   uuid.UUID
}

// Validate can return ValidationError
func (companyEvent *DeleteCompanyEvent) Validate() error {
	if companyEvent.CompanyID == uuid.Nil {
		return errors.NewValidationError(nil, "CompanyID cannot be empty")
	}

	if companyEvent.EventID == uuid.Nil {
		return errors.NewValidationError(nil, "EventID cannot be empty")
	}

	return nil
}
