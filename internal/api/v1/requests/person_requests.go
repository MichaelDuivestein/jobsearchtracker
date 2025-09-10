package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"

	"github.com/google/uuid"
)

type CreatePersonRequest struct {
	ID         *uuid.UUID `json:"id,omitempty"`
	Name       string     `json:"name"`
	PersonType PersonType `json:"person_type"`
	Email      *string    `json:"email,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
}

func (request *CreatePersonRequest) validate() error {
	if request.ID != nil {
		err := uuid.Validate(request.ID.String())
		if err != nil {
			message := "ID is invalid"
			slog.Info("CreatePersonRequest.validate failed: " + message)
			return internalErrors.NewValidationError(nil, message)
		} else if *request.ID == uuid.Nil {
			message := "ID is empty"
			slog.Info("CreatePersonRequest.Validate: "+message, "ID", request.ID)
			return internalErrors.NewValidationError(nil, message)
		}
	}

	if request.Name == "" {
		message := "Name is empty"
		slog.Info("CreatePersonRequest.validate failed: " + message)
		name := "Name"
		return internalErrors.NewValidationError(&name, message)
	}

	if !request.PersonType.IsValid() {
		message := "PersonType is invalid"
		slog.Info("CreatePersonRequest.validate failed: " + message)
		companyType := "PersonType"
		return internalErrors.NewValidationError(&companyType, message)
	}

	return nil
}

func (request *CreatePersonRequest) ToModel() (*models.CreatePerson, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	personType, _ := request.PersonType.ToModel()

	personModel := models.CreatePerson{
		ID:          request.ID,
		Name:        request.Name,
		PersonType:  personType,
		Email:       request.Email,
		Phone:       request.Phone,
		Notes:       request.Notes,
		CreatedDate: nil,
		UpdatedDate: nil,
	}

	return &personModel, nil
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

// ToModel can return ValidationError
func (personType PersonType) ToModel() (models.PersonType, error) {
	switch personType {
	case PersonTypeCEO:
		return models.PersonTypeCEO, nil
	case PersonTypeCTO:
		return models.PersonTypeCTO, nil
	case PersonTypeDeveloper:
		return models.PersonTypeDeveloper, nil
	case PersonTypeExternalRecruiter:
		return models.PersonTypeExternalRecruiter, nil
	case PersonTypeInternalRecruiter:
		return models.PersonTypeInternalRecruiter, nil
	case PersonTypeHR:
		return models.PersonTypeHR, nil
	case PersonTypeJobAdvertiser:
		return models.PersonTypeJobAdvertiser, nil
	case PersonTypeJobContact:
		return models.PersonTypeJobContact, nil
	case PersonTypeOther:
		return models.PersonTypeOther, nil
	case PersonTypeUnknown:
		return models.PersonTypeUnknown, nil
	default:
		slog.Info("v1.types.toModel: Invalid PersonType: '" + personType.String() + "'")
		personTypeString := "PersonType"
		return "", internalErrors.NewValidationError(
			&personTypeString,
			"invalid PersonType: '"+personType.String()+"'")
	}
}

func NewPersonType(modelPersonType *models.PersonType) (PersonType, error) {
	if modelPersonType == nil {
		slog.Info("v1.types.NewPersonType: modelPersonType is nil")
		return "", internalErrors.NewInternalServiceError(
			"Error trying to convert internal personType to external PersonType.")
	}

	switch *modelPersonType {
	case models.PersonTypeCEO:
		return PersonTypeCEO, nil
	case models.PersonTypeCTO:
		return PersonTypeCTO, nil
	case models.PersonTypeDeveloper:
		return PersonTypeDeveloper, nil
	case models.PersonTypeExternalRecruiter:
		return PersonTypeExternalRecruiter, nil
	case models.PersonTypeInternalRecruiter:
		return PersonTypeInternalRecruiter, nil
	case models.PersonTypeHR:
		return PersonTypeHR, nil
	case models.PersonTypeJobAdvertiser:
		return PersonTypeJobAdvertiser, nil
	case models.PersonTypeJobContact:
		return PersonTypeJobContact, nil
	case models.PersonTypeOther:
		return PersonTypeOther, nil
	case models.PersonTypeUnknown:
		return PersonTypeUnknown, nil
	default:
		slog.Info("v1.types.NewPersonType: Invalid modelPersonType: '" + modelPersonType.String() + "'")
		return "", internalErrors.NewInternalServiceError(
			"Error converting internal PersonType to external PersonType: '" + modelPersonType.String() + "'")
	}
}

func (personType PersonType) ToPointer() *PersonType {
	return &personType
}
