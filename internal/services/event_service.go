package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type EventService struct {
	eventRepository *repositories.EventRepository
}

func NewEventService(eventRepository *repositories.EventRepository) *EventService {
	return &EventService{eventRepository: eventRepository}
}

// CreateEvent can return ConflictError, InternalServiceError, ValidationError
func (eventService *EventService) CreateEvent(event *models.CreateEvent) (*models.Event, error) {
	if event == nil {
		slog.Error("event_service.CreateEvent: event is nil")
		return nil, internalErrors.NewValidationError(nil, "CreateEvent is nil")
	}

	err := event.Validate()
	if err != nil {
		var eventID string
		if event.ID != nil {
			eventID = event.ID.String()
		} else {
			eventID = "[not set]"
		}
		slog.Info("event_service.CreateEvent: event to create is invalid. ", "ID", eventID, "error", err)
		return nil, err
	}

	if event.CreatedDate == nil {
		createdDate := time.Now()
		event.CreatedDate = &createdDate
	} else if event.CreatedDate.IsZero() {
		createdDate := time.Now()
		event.CreatedDate = &createdDate
		slog.Info(
			"event_service.CreateEvent: event.CreatedDate is zero. Setting to '" + event.CreatedDate.String() + "'")
	}

	insertedEvent, err := eventService.eventRepository.Create(event)
	if err != nil {
		return nil, err
	}

	slog.Info("event_service.CreateEvent: Inserted event.", "event.ID", insertedEvent.ID)
	return insertedEvent, nil
}

// GetEventByID can return ConflictError, InternalServiceError, NewValidationError
func (eventService *EventService) GetEventByID(eventID *uuid.UUID) (*models.Event, error) {
	if eventID == nil {
		eventIDString := "event ID"
		err := internalErrors.NewValidationError(&eventIDString, "eventID is required")
		slog.Info("eventService.GetEvenByID: Failed to get Event", "error", err)
		return nil, err
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	event, err := eventService.eventRepository.GetByID(eventID)
	if err != nil {
		return nil, err
	}

	slog.Info("eventService.GetEvenByID: Retrieved event.", "event.ID", event.ID.String())

	return event, nil
}

// GetAllEvents can return InternalServiceError
func (eventService *EventService) GetAllEvents() ([]*models.Event, error) {
	events, err := eventService.eventRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if events == nil {
		slog.Info("event_service.GetAllEvents: Retrieved Zero events")
	} else {
		slog.Info("event_service.GetAllEvents: Retrieved " + string(rune(len(events))) + " events")
	}

	return events, nil
}

// UpdateEvent can return InternalServiceError, ValidationError
func (eventService *EventService) UpdateEvent(event *models.UpdateEvent) error {
	if event == nil {
		slog.Error("EventService.UpdateEvent: UpdateEvent is nil")
		return internalErrors.NewValidationError(nil, "UpdateEvent model is nil")
	}

	// can return ValidationError
	err := event.Validate()
	if err != nil {
		slog.Info("EventService.UpdateEvent: UpdateEvent model is invalid. ", "error", err)
		return err
	}

	// can return InternalServiceError, ValidationError
	err = eventService.eventRepository.Update(event)
	if err != nil {
		slog.Error("EventService.UpdateEvent: Error updating event", "error", err)
	}

	return err
}

// DeleteEvent can return InternalServiceError, NotFoundError, ValidationError
func (eventService *EventService) DeleteEvent(eventID *uuid.UUID) error {
	if eventID == nil {
		eventIDString := "event ID"
		err := internalErrors.NewValidationError(&eventIDString, "eventID is required")
		slog.Info("eventService.DeleteEvent: Failed to delete event", "error", err)
		return err
	}

	err := eventService.eventRepository.Delete(eventID)
	if err != nil {
		slog.Error("EventService.DeleteEvent: Error deleting event", "error", err)
	}

	return err
}
