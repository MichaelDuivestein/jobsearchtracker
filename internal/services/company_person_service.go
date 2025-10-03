package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"

	"github.com/google/uuid"
)

type CompanyPersonService struct {
	companyPersonRepository *repositories.CompanyPersonRepository
}

func NewCompanyPersonService(companyPersonRepository *repositories.CompanyPersonRepository) *CompanyPersonService {
	return &CompanyPersonService{companyPersonRepository: companyPersonRepository}
}

// AssociateCompanyPerson can return ConflictError, InternalServiceError, ValidationError
func (companyPersonService *CompanyPersonService) AssociateCompanyPerson(
	AssociateCompanyPerson *models.AssociateCompanyPerson) (*models.CompanyPerson, error) {

	if AssociateCompanyPerson == nil {
		slog.Error("person_service.AssociateCompanyPerson: model is nil")
		return nil, internalErrors.NewValidationError(nil, "AssociateCompanyPerson model is nil")
	}

	err := AssociateCompanyPerson.Validate()
	if err != nil {
		slog.Info("person_service.AssociateCompanyPerson: CompanyPerson to create is invalid", "error", err)

		return nil, err
	}

	insertedCompanyPerson, err :=
		companyPersonService.companyPersonRepository.AssociateCompanyPerson(AssociateCompanyPerson)

	if err != nil {
		return nil, err
	}

	slog.Info(
		"person_service.AssociateCompanyPerson: Associated company to person.",
		"person.CompanyID", insertedCompanyPerson.CompanyID,
		"person.PersonID", insertedCompanyPerson.PersonID)
	return insertedCompanyPerson, nil
}

// GetByID can return ValidationError, InternalServiceError
func (companyPersonService *CompanyPersonService) GetByID(
	companyID *uuid.UUID, personID *uuid.UUID) ([]*models.CompanyPerson, error) {

	if (companyID == nil || *companyID == uuid.Nil) && (personID == nil || *personID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "companyID and personID cannot both be empty")
	}

	companyPersons, err := companyPersonService.companyPersonRepository.GetByID(companyID, personID)
	if err != nil {
		return nil, err
	}

	if companyPersons == nil {
		slog.Info("CompanyPersonService.GetByID: Retrieved zero persons")
	} else {
		slog.Info(
			"CompanyPersonService.GetByID: Retrieved " + string(rune(len(companyPersons))) + " persons")
	}

	return companyPersons, nil
}

// GetAll can return InternalServiceError
func (companyPersonService *CompanyPersonService) GetAll() ([]*models.CompanyPerson, error) {
	companyPersons, err := companyPersonService.companyPersonRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if companyPersons == nil {
		slog.Info("CompanyPersonService.GetAllCompanyPersons: Retrieved zero persons")
	} else {
		slog.Info(
			"CompanyPersonService.GetAllCompanyPersons: Retrieved " + string(rune(len(companyPersons))) + " persons")
	}

	return companyPersons, nil
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (companyPersonService *CompanyPersonService) Delete(model *models.DeleteCompanyPerson) error {
	err := model.Validate()
	if err != nil {
		return err
	}

	return companyPersonService.companyPersonRepository.Delete(model)
}
