package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"

	"github.com/google/uuid"
)

type ApplicationPersonService struct {
	applicationPersonRepository *repositories.ApplicationPersonRepository
}

func NewApplicationPersonService(applicationPersonRepository *repositories.ApplicationPersonRepository) *ApplicationPersonService {
	return &ApplicationPersonService{applicationPersonRepository: applicationPersonRepository}
}

// AssociateApplicationPerson can return ConflictError, InternalServiceError, ValidationError
func (applicationPersonService *ApplicationPersonService) AssociateApplicationPerson(
	AssociateApplicationPerson *models.AssociateApplicationPerson) (*models.ApplicationPerson, error) {

	if AssociateApplicationPerson == nil {
		slog.Error("person_service.AssociateApplicationPerson: model is nil")
		return nil, internalErrors.NewValidationError(nil, "AssociateApplicationPerson model is nil")
	}

	err := AssociateApplicationPerson.Validate()
	if err != nil {
		slog.Info("person_service.AssociateApplicationPerson: ApplicationPerson to create is invalid", "error", err)

		return nil, err
	}

	insertedApplicationPerson, err :=
		applicationPersonService.applicationPersonRepository.AssociateApplicationPerson(AssociateApplicationPerson)

	if err != nil {
		return nil, err
	}

	slog.Info(
		"person_service.AssociateApplicationPerson: Associated application to person.",
		"person.ApplicationID", insertedApplicationPerson.ApplicationID,
		"person.PersonID", insertedApplicationPerson.PersonID)
	return insertedApplicationPerson, nil
}

// GetByID can return ValidationError, InternalServiceError
func (applicationPersonService *ApplicationPersonService) GetByID(
	applicationID *uuid.UUID, personID *uuid.UUID) ([]*models.ApplicationPerson, error) {

	if (applicationID == nil || *applicationID == uuid.Nil) && (personID == nil || *personID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "applicationID and personID cannot both be empty")
	}

	applicationPersons, err := applicationPersonService.applicationPersonRepository.GetByID(applicationID, personID)
	if err != nil {
		return nil, err
	}

	if applicationPersons == nil {
		slog.Info("ApplicationPersonService.GetByID: Retrieved zero persons")
	} else {
		slog.Info(
			"ApplicationPersonService.GetByID: Retrieved " + string(rune(len(applicationPersons))) + " persons")
	}

	return applicationPersons, nil
}

// GetAll can return InternalServiceError
func (applicationPersonService *ApplicationPersonService) GetAll() ([]*models.ApplicationPerson, error) {
	applicationPersons, err := applicationPersonService.applicationPersonRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if applicationPersons == nil {
		slog.Info("ApplicationPersonService.GetAllApplicationPersons: Retrieved zero persons")
	} else {
		slog.Info(
			"ApplicationPersonService.GetAllApplicationPersons: Retrieved " + string(rune(len(applicationPersons))) + " persons")
	}

	return applicationPersons, nil
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (applicationPersonService *ApplicationPersonService) Delete(model *models.DeleteApplicationPerson) error {
	err := model.Validate()
	if err != nil {
		return err
	}

	return applicationPersonService.applicationPersonRepository.Delete(model)
}
