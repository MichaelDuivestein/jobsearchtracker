package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID           uuid.UUID
	Name         *string
	CompanyType  *CompanyType
	Notes        *string
	LastContact  *time.Time
	CreatedDate  *time.Time
	UpdatedDate  *time.Time
	Applications *[]*Application
	Persons      *[]*Person
	Events       *[]*Event
}

type CreateCompany struct {
	ID          *uuid.UUID
	Name        string
	CompanyType CompanyType
	Notes       *string
	LastContact *time.Time
	CreatedDate *time.Time
	UpdatedDate *time.Time
}

func (company *CreateCompany) Validate() error {
	if company.Name == "" {
		name := "Name"
		return errors.NewValidationError(&name, "company name is empty")
	}

	if !company.CompanyType.IsValid() {
		companyType := "CompanyType"
		return errors.NewValidationError(&companyType, "company type is invalid")
	}

	if company.UpdatedDate != nil && company.UpdatedDate.IsZero() {
		updatedDate := "UpdatedDate"
		return errors.NewValidationError(
			&updatedDate,
			"updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	return nil
}

type UpdateCompany struct {
	ID          uuid.UUID
	Name        *string
	CompanyType *CompanyType
	Notes       *string
	LastContact *time.Time
}

// Validate can return ValidationError
func (updateCompany *UpdateCompany) Validate() error {
	if updateCompany.Name == nil &&
		updateCompany.CompanyType == nil &&
		updateCompany.Notes == nil &&
		updateCompany.LastContact == nil {

		return errors.NewValidationError(nil, "nothing to update")
	}
	return nil
}

type CompanyType string

const (
	CompanyTypeEmployer    = "employer"
	CompanyTypeRecruiter   = "recruiter"
	CompanyTypeConsultancy = "consultancy"
)

func (companyType CompanyType) IsValid() bool {
	switch companyType {
	case CompanyTypeEmployer, CompanyTypeRecruiter, CompanyTypeConsultancy:
		return true
	}
	return false
}

func (companyType CompanyType) String() string {
	return string(companyType)
}
