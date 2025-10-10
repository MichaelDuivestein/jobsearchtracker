package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type CreateCompanyRequest struct {
	ID          *uuid.UUID  `json:"id,omitempty" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	Name        string      `json:"name" example:"CompanyName AB" extensions:"x-order=1"`
	CompanyType CompanyType `json:"company_type" example:"employer" extensions:"x-order=2"`
	Notes       *string     `json:"notes,omitempty" example:"Notes go here" extensions:"x-order=3"`
	LastContact *time.Time  `json:"last_contact,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=4"`
}

// Validate can return ValidationError
func (request *CreateCompanyRequest) Validate() error {
	if request.ID != nil {
		err := uuid.Validate(request.ID.String())
		if err != nil {
			message := "ID is invalid"
			slog.Info("CreateCompanyRequest.validate failed: " + message)
			return internalErrors.NewValidationError(nil, message)
		} else if *request.ID == uuid.Nil {
			message := "ID is empty"
			slog.Info("CreateCompanyRequest.Validate: "+message, "ID", request.ID)
			return internalErrors.NewValidationError(nil, message)
		}
	}

	if request.Name == "" {
		message := "Name is empty"
		slog.Info("CreateCompanyRequest.validate failed: " + message)
		name := "Name"
		return internalErrors.NewValidationError(&name, message)

	}
	if !request.CompanyType.IsValid() {
		message := "CompanyType is invalid"
		slog.Info("CreateCompanyRequest.validate failed: " + message)
		companyType := "CompanyType"
		return internalErrors.NewValidationError(&companyType, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *CreateCompanyRequest) ToModel() (*models.CreateCompany, error) {
	err := request.Validate()
	if err != nil {
		return nil, err
	}

	companyType, _ := request.CompanyType.ToModel()

	companyModel := models.CreateCompany{
		ID:          request.ID,
		Name:        request.Name,
		CompanyType: companyType,
		Notes:       request.Notes,
		LastContact: request.LastContact,
		CreatedDate: nil,
		UpdatedDate: nil,
	}

	return &companyModel, nil
}

type UpdateCompanyRequest struct {
	ID          uuid.UUID    `json:"id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	Name        *string      `json:"name,omitempty" example:"CompanyName AB" extensions:"x-order=1"`
	CompanyType *CompanyType `json:"company_type,omitempty" example:"employer" extensions:"x-order=2"`
	Notes       *string      `json:"notes,omitempty" example:"Notes go here" extensions:"x-order=3"`
	LastContact *time.Time   `json:"last_contact,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=4"`
}

// Validate can return ValidationError
func (request *UpdateCompanyRequest) Validate() error {
	err := uuid.Validate(request.ID.String())
	if err != nil {
		message := "ID is invalid"
		slog.Info("UpdateCompanyRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}
	if request.ID == uuid.Nil {
		message := "ID is empty"
		slog.Info("UpdateCompanyRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.Name == nil && request.CompanyType == nil && request.Notes == nil && request.LastContact == nil {
		message := "nothing to update"
		slog.Info("UpdateCompanyRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.Name != nil && *request.Name == "" {
		message := "Name is invalid"
		slog.Info("UpdateCompanyRequest.Validate: "+message, "ID", request.ID)

		companyType := "Name"
		return internalErrors.NewValidationError(&companyType, message)
	}

	if request.CompanyType != nil && !request.CompanyType.IsValid() {
		message := "CompanyType is invalid"
		slog.Info("UpdateCompanyRequest.Validate: "+message, "ID", request.ID)

		companyType := "CompanyType"
		return internalErrors.NewValidationError(&companyType, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *UpdateCompanyRequest) ToModel() (*models.UpdateCompany, error) {
	// can return ValidationError
	err := request.Validate()
	if err != nil {
		slog.Info("validate updateCompanyRequest failed", "error", err)
		return nil, err
	}

	var companyType *models.CompanyType
	if request.CompanyType != nil {
		// can return ValidationError
		tempCompanyType, _ := request.CompanyType.ToModel()
		companyType = &tempCompanyType
	} else {
		companyType = nil
	}

	updateModel := models.UpdateCompany{
		ID:          request.ID,
		Name:        request.Name,
		CompanyType: companyType,
		Notes:       request.Notes,
		LastContact: request.LastContact,
	}

	return &updateModel, nil
}

// CompanyType represents the type of company
//
// @enum employer,recruiter,consultancy
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

// ToModel can return ValidationError
func (companyType CompanyType) ToModel() (models.CompanyType, error) {
	switch companyType {
	case CompanyTypeEmployer:
		return models.CompanyTypeEmployer, nil
	case CompanyTypeRecruiter:
		return models.CompanyTypeRecruiter, nil
	case CompanyTypeConsultancy:
		return models.CompanyTypeConsultancy, nil
	default:
		slog.Info("v1.types.toModel: Invalid CompanyType: '" + companyType.String() + "'")
		companyTypeString := "CompanyType"
		return "", internalErrors.NewValidationError(
			&companyTypeString,
			"invalid CompanyType: '"+companyType.String()+"'")
	}
}

// NewCompanyType can return InternalServerError
func NewCompanyType(modelCompanyType *models.CompanyType) (CompanyType, error) {
	if modelCompanyType == nil {
		slog.Info("v1.types.NewCompanyType: modelCompanyType is nil")
		return "",
			internalErrors.NewInternalServiceError(
				"Error trying to convert internal companyType to external CompanyType.")
	}

	switch *modelCompanyType {
	case models.CompanyTypeEmployer:
		return CompanyTypeEmployer, nil
	case models.CompanyTypeRecruiter:
		return CompanyTypeRecruiter, nil
	case models.CompanyTypeConsultancy:
		return CompanyTypeConsultancy, nil
	default:
		slog.Error("v1.types.NewCompanyType: Invalid modelCompanyType: '" + modelCompanyType.String() + "'")
		return "",
			internalErrors.NewInternalServiceError(
				"Error converting internal CompanyType to external CompanyType: '" + modelCompanyType.String() + "'")
	}
}

func (companyType CompanyType) ToPointer() *CompanyType {
	return &companyType
}
