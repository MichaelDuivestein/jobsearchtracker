package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type Person struct {
	ID          uuid.UUID
	Name        string
	PersonType  PersonType
	Email       *string
	Phone       *string
	Notes       *string
	CreatedDate time.Time
	UpdatedDate *time.Time
}

type CreatePerson struct {
	ID          *uuid.UUID
	Name        string
	PersonType  PersonType
	Email       *string
	Phone       *string
	Notes       *string
	CreatedDate *time.Time
	UpdatedDate *time.Time
}

// Validate can return NewValidationError
func (person *CreatePerson) Validate() error {
	emptyString := ""

	if person.Name == "" {
		name := "Name"
		return errors.NewValidationError(&name, "person name is empty")
	}

	if !person.PersonType.IsValid() {
		companyType := "PersonType"
		return errors.NewValidationError(&companyType, "person type is invalid")
	}

	if person.Email != nil && *person.Email == emptyString {
		name := "email"
		return errors.NewValidationError(&name, "person email is empty")
	}

	if person.Phone != nil && *person.Phone == emptyString {
		name := "phone"
		return errors.NewValidationError(
			&name,
			"person phone is empty. It should either be nil or a valid phone number")
	}

	if person.UpdatedDate != nil && person.UpdatedDate.IsZero() {
		updatedDate := "UpdatedDate"
		return errors.NewValidationError(
			&updatedDate,
			"updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	return nil
}

type UpdatePerson struct {
	ID         uuid.UUID
	Name       *string
	PersonType *PersonType
	Email      *string
	Phone      *string
	Notes      *string
}

// Validate can return ValidationError
func (updatePerson *UpdatePerson) Validate() error {
	if updatePerson.Name == nil && updatePerson.PersonType == nil && updatePerson.Email == nil && updatePerson.Phone == nil && updatePerson.Notes == nil {
		return errors.NewValidationError(nil, "nothing to update")
	}
	return nil
}

type PersonType string

const (
	PersonTypeCEO               = "CEO"
	PersonTypeCTO               = "CTO"
	PersonTypeDeveloper         = "developer"
	PersonTypeExternalRecruiter = "externalRecruiter"
	PersonTypeInternalRecruiter = "internalRecruiter"
	PersonTypeHR                = "HR"
	PersonTypeJobAdvertiser     = "jobAdvertiser"
	PersonTypeJobContact        = "jobContact"
	PersonTypeOther             = "other"
	PersonTypeUnknown           = "unknown"
)

func (personType PersonType) IsValid() bool {
	switch personType {
	case PersonTypeCEO, PersonTypeCTO, PersonTypeDeveloper, PersonTypeExternalRecruiter, PersonTypeInternalRecruiter,
		PersonTypeHR, PersonTypeJobAdvertiser, PersonTypeJobContact, PersonTypeOther, PersonTypeUnknown:
		return true
	}
	return false
}

func (personType PersonType) String() string { return string(personType) }
