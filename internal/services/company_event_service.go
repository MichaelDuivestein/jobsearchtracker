package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"

	"github.com/google/uuid"
)

type CompanyEventService struct {
	companyEventRepository *repositories.CompanyEventRepository
}

func NewCompanyEventService(
	companyEventRepository *repositories.CompanyEventRepository) *CompanyEventService {

	return &CompanyEventService{companyEventRepository: companyEventRepository}
}

// AssociateCompanyEvent can return ConflictError, InternalServiceError, ValidationError
func (companyEventService *CompanyEventService) AssociateCompanyEvent(
	AssociateCompanyEvent *models.AssociateCompanyEvent) (*models.CompanyEvent, error) {

	if AssociateCompanyEvent == nil {
		slog.Error("event_service.AssociateCompanyEvent: model is nil")
		return nil, internalErrors.NewValidationError(nil, "AssociateCompanyEvent model is nil")
	}

	err := AssociateCompanyEvent.Validate()
	if err != nil {
		slog.Info("event_service.AssociateCompanyEvent: CompanyEvent to create is invalid", "error", err)

		return nil, err
	}

	insertedCompanyEvent, err :=
		companyEventService.companyEventRepository.AssociateCompanyEvent(AssociateCompanyEvent)

	if err != nil {
		return nil, err
	}

	slog.Info(
		"event_service.AssociateCompanyEvent: Associated company to event.",
		"event.CompanyID", insertedCompanyEvent.CompanyID,
		"event.EventID", insertedCompanyEvent.EventID)
	return insertedCompanyEvent, nil
}

// GetByID can return ValidationError, InternalServiceError
func (companyEventService *CompanyEventService) GetByID(
	companyID *uuid.UUID, eventID *uuid.UUID) ([]*models.CompanyEvent, error) {

	if (companyID == nil || *companyID == uuid.Nil) && (eventID == nil || *eventID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "companyID and eventID cannot both be empty")
	}

	companyEvents, err := companyEventService.companyEventRepository.GetByID(companyID, eventID)
	if err != nil {
		return nil, err
	}

	if companyEvents == nil {
		slog.Info("CompanyEventService.GetByID: Retrieved zero events")
	} else {
		slog.Info(
			"CompanyEventService.GetByID: Retrieved " + string(rune(len(companyEvents))) + " events")
	}

	return companyEvents, nil
}

// GetAll can return InternalServiceError
func (companyEventService *CompanyEventService) GetAll() ([]*models.CompanyEvent, error) {
	companyEvents, err := companyEventService.companyEventRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if companyEvents == nil {
		slog.Info("CompanyEventService.GetAllCompanyEvents: Retrieved zero events")
	} else {
		slog.Info(
			"CompanyEventService.GetAllCompanyEvents: Retrieved " + string(rune(len(companyEvents))) + " events")
	}

	return companyEvents, nil
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (companyEventService *CompanyEventService) Delete(model *models.DeleteCompanyEvent) error {
	err := model.Validate()
	if err != nil {
		return err
	}

	return companyEventService.companyEventRepository.Delete(model)
}
