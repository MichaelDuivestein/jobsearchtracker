package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"

	"github.com/google/uuid"
)

type AssociateCompanyPersonRequest struct {
	CompanyID uuid.UUID `json:"company_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID  uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *AssociateCompanyPersonRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("CreateCompanyPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.CompanyID == uuid.Nil {
		message := "CompanyID is invalid"
		slog.Info("CreateCompanyPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.PersonID == uuid.Nil {
		message := "PersonID is invalid"
		slog.Info("CreateCompanyPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *AssociateCompanyPersonRequest) ToModel() (*models.AssociateCompanyPerson, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.AssociateCompanyPerson{
		CompanyID: request.CompanyID,
		PersonID:  request.PersonID,
	}

	return &model, nil
}

type DeleteCompanyPersonRequest struct {
	CompanyID uuid.UUID `json:"company_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID  uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
}

// validate can return ValidationError
func (request *DeleteCompanyPersonRequest) validate() error {
	if request == nil {
		message := "request is nil"
		slog.Info("DeleteCompanyPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.CompanyID == uuid.Nil {
		message := "CompanyID is invalid"
		slog.Info("DeleteCompanyPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.PersonID == uuid.Nil {
		message := "PersonID is invalid"
		slog.Info("DeleteCompanyPersonRequest.validate failed: " + message)
		return internalErrors.NewValidationError(nil, message)
	}

	return nil
}

// ToModel can return ValidationError
func (request *DeleteCompanyPersonRequest) ToModel() (*models.DeleteCompanyPerson, error) {
	if request == nil {
		return nil, nil
	}

	err := request.validate()
	if err != nil {
		return nil, err
	}

	model := models.DeleteCompanyPerson{
		CompanyID: request.CompanyID,
		PersonID:  request.PersonID,
	}

	return &model, nil
}
