package models

import (
	"github.com/google/uuid"
	internalErrors "jobsearchtracker/internal/errors"
	"time"
)

type CreateCompany struct {
	ID          *uuid.UUID
	Name        string
	CompanyType CompanyType
	Notes       *string
	LastContact *time.Time
	CreatedDate *time.Time
	UpdatedDate *time.Time
}

func (company *CreateCompany) IsValid() error {
	if company.Name == "" {
		name := "Name"
		return internalErrors.NewValidationError(&name, "company name is empty")
	}

	if !company.CompanyType.IsValid() {
		companyType := "companyType"
		return internalErrors.NewValidationError(&companyType, "company type is invalid")
	}

	if company.UpdatedDate != nil && company.UpdatedDate.IsZero() {
		updatedDate := "UpdatedDate"
		return internalErrors.NewValidationError(&updatedDate, "updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	return nil
}
