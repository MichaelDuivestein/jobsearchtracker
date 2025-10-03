package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type CompanyPerson struct {
	CompanyID   uuid.UUID
	PersonID    uuid.UUID
	CreatedDate time.Time
}

type AssociateCompanyPerson struct {
	CompanyID   uuid.UUID
	PersonID    uuid.UUID
	CreatedDate *time.Time
}

// Validate can return ValidationError
func (companyPerson *AssociateCompanyPerson) Validate() error {
	if companyPerson.CompanyID == uuid.Nil {
		return errors.NewValidationError(nil, "CompanyID is empty")
	}

	if companyPerson.PersonID == uuid.Nil {
		return errors.NewValidationError(nil, "PersonID is empty")
	}

	return nil
}

type DeleteCompanyPerson struct {
	CompanyID uuid.UUID
	PersonID  uuid.UUID
}

// Validate can return ValidationError
func (companyPerson *DeleteCompanyPerson) Validate() error {
	if companyPerson.CompanyID == uuid.Nil {
		return errors.NewValidationError(nil, "CompanyID cannot be empty")
	}

	if companyPerson.PersonID == uuid.Nil {
		return errors.NewValidationError(nil, "PersonID cannot be empty")
	}

	return nil
}
