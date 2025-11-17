package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"

	"github.com/google/uuid"
)

type EventPersonService struct {
	eventPersonRepository *repositories.EventPersonRepository
}

func NewEventPersonService(eventPersonRepository *repositories.EventPersonRepository) *EventPersonService {
	return &EventPersonService{eventPersonRepository: eventPersonRepository}
}

// AssociateEventPerson can return ConflictError, InternalServiceError, ValidationError
func (eventPersonService *EventPersonService) AssociateEventPerson(
	AssociateEventPerson *models.AssociateEventPerson) (*models.EventPerson, error) {

	if AssociateEventPerson == nil {
		slog.Error("person_service.AssociateEventPerson: model is nil")
		return nil, internalErrors.NewValidationError(nil, "AssociateEventPerson model is nil")
	}

	err := AssociateEventPerson.Validate()
	if err != nil {
		slog.Info("person_service.AssociateEventPerson: EventPerson to create is invalid", "error", err)

		return nil, err
	}

	insertedEventPerson, err :=
		eventPersonService.eventPersonRepository.AssociateEventPerson(AssociateEventPerson)

	if err != nil {
		return nil, err
	}

	slog.Info(
		"person_service.AssociateEventPerson: Associated event to person.",
		"person.EventID", insertedEventPerson.EventID,
		"person.PersonID", insertedEventPerson.PersonID)
	return insertedEventPerson, nil
}

// GetByID can return ValidationError, InternalServiceError
func (eventPersonService *EventPersonService) GetByID(
	eventID *uuid.UUID, personID *uuid.UUID) ([]*models.EventPerson, error) {

	if (eventID == nil || *eventID == uuid.Nil) && (personID == nil || *personID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "eventID and personID cannot both be empty")
	}

	eventPersons, err := eventPersonService.eventPersonRepository.GetByID(eventID, personID)
	if err != nil {
		return nil, err
	}

	if eventPersons == nil {
		slog.Info("EventPersonService.GetByID: Retrieved zero persons")
	} else {
		slog.Info(
			"EventPersonService.GetByID: Retrieved " + string(rune(len(eventPersons))) + " persons")
	}

	return eventPersons, nil
}

// GetAll can return InternalServiceError
func (eventPersonService *EventPersonService) GetAll() ([]*models.EventPerson, error) {
	eventPersons, err := eventPersonService.eventPersonRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if eventPersons == nil {
		slog.Info("EventPersonService.GetAllEventPersons: Retrieved zero persons")
	} else {
		slog.Info(
			"EventPersonService.GetAllEventPersons: Retrieved " + string(rune(len(eventPersons))) + " persons")
	}

	return eventPersons, nil
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (eventPersonService *EventPersonService) Delete(model *models.DeleteEventPerson) error {
	err := model.Validate()
	if err != nil {
		return err
	}

	return eventPersonService.eventPersonRepository.Delete(model)
}
