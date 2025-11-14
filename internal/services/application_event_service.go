package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"

	"github.com/google/uuid"
)

type ApplicationEventService struct {
	applicationEventRepository *repositories.ApplicationEventRepository
}

func NewApplicationEventService(
	applicationEventRepository *repositories.ApplicationEventRepository) *ApplicationEventService {

	return &ApplicationEventService{applicationEventRepository: applicationEventRepository}
}

// AssociateApplicationEvent can return ConflictError, InternalServiceError, ValidationError
func (applicationEventService *ApplicationEventService) AssociateApplicationEvent(
	AssociateApplicationEvent *models.AssociateApplicationEvent) (*models.ApplicationEvent, error) {

	if AssociateApplicationEvent == nil {
		slog.Error("event_service.AssociateApplicationEvent: model is nil")
		return nil, internalErrors.NewValidationError(nil, "AssociateApplicationEvent model is nil")
	}

	err := AssociateApplicationEvent.Validate()
	if err != nil {
		slog.Info("event_service.AssociateApplicationEvent: ApplicationEvent to create is invalid", "error", err)

		return nil, err
	}

	insertedApplicationEvent, err :=
		applicationEventService.applicationEventRepository.AssociateApplicationEvent(AssociateApplicationEvent)

	if err != nil {
		return nil, err
	}

	slog.Info(
		"event_service.AssociateApplicationEvent: Associated application to event.",
		"event.ApplicationID", insertedApplicationEvent.ApplicationID,
		"event.EventID", insertedApplicationEvent.EventID)
	return insertedApplicationEvent, nil
}

// GetByID can return ValidationError, InternalServiceError
func (applicationEventService *ApplicationEventService) GetByID(
	applicationID *uuid.UUID, eventID *uuid.UUID) ([]*models.ApplicationEvent, error) {

	if (applicationID == nil || *applicationID == uuid.Nil) && (eventID == nil || *eventID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "applicationID and eventID cannot both be empty")
	}

	applicationEvents, err := applicationEventService.applicationEventRepository.GetByID(applicationID, eventID)
	if err != nil {
		return nil, err
	}

	if applicationEvents == nil {
		slog.Info("ApplicationEventService.GetByID: Retrieved zero events")
	} else {
		slog.Info(
			"ApplicationEventService.GetByID: Retrieved " + string(rune(len(applicationEvents))) + " events")
	}

	return applicationEvents, nil
}

// GetAll can return InternalServiceError
func (applicationEventService *ApplicationEventService) GetAll() ([]*models.ApplicationEvent, error) {
	applicationEvents, err := applicationEventService.applicationEventRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if applicationEvents == nil {
		slog.Info("ApplicationEventService.GetAllApplicationEvents: Retrieved zero events")
	} else {
		slog.Info(
			"ApplicationEventService.GetAllApplicationEvents: Retrieved " + string(rune(len(applicationEvents))) + " events")
	}

	return applicationEvents, nil
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (applicationEventService *ApplicationEventService) Delete(model *models.DeleteApplicationEvent) error {
	err := model.Validate()
	if err != nil {
		return err
	}

	return applicationEventService.applicationEventRepository.Delete(model)
}
