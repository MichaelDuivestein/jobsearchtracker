package responses

import (
	"jobsearchtracker/internal/api/v1/requests"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type EventDTO struct {
	ID          uuid.UUID           `json:"id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventType   *requests.EventType `json:"event_type,omitempty" example:"interviewCompleted" extensions:"x-order=2"`
	Description *string             `json:"description,omitempty" example:"Event Description" extensions:"x-order=2"`
	Notes       *string             `json:"notes,omitempty" example:"Notes go here" extensions:"x-order=3"`
	EventDate   *time.Time          `json:"event_date,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=4"`
	CreatedDate *time.Time          `json:"created_date,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=5"`
	UpdatedDate *time.Time          `json:"updated_date,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=6"`
}

func NewEventDTO(eventModel *models.Event) (*EventDTO, error) {
	if eventModel == nil {
		slog.Error("responses.EventDTO: Event is nil")
		return nil, internalErrors.NewInternalServiceError("Error building DTO: Event is nil")
	}

	var eventType *requests.EventType = nil
	if eventModel.EventType != nil {
		nonNilEventType, err := requests.NewEventType(eventModel.EventType)
		if err != nil {
			return nil, err
		}
		eventType = &nonNilEventType
	}

	eventDto := EventDTO{
		ID:          eventModel.ID,
		EventType:   eventType,
		Description: eventModel.Description,
		Notes:       eventModel.Notes,
		EventDate:   eventModel.EventDate,
		CreatedDate: eventModel.CreatedDate,
		UpdatedDate: eventModel.UpdatedDate,
	}

	return &eventDto, nil
}

func NewEventDTOs(events []*models.Event) ([]*EventDTO, error) {
	if len(events) == 0 {
		return []*EventDTO{}, nil
	}

	var eventDTOs = make([]*EventDTO, len(events))
	for index := range events {
		eventDTO, err := NewEventDTO(events[index])
		if err != nil {
			return nil, err
		}
		eventDTOs[index] = eventDTO
	}
	return eventDTOs, nil
}

type EventResponse struct {
	EventDTO
}

func NewEventResponse(eventModel *models.Event) (*EventResponse, error) {
	if eventModel == nil {
		slog.Error("responses.EventResponse: EventModel is nil")
		return nil, internalErrors.NewInternalServiceError("Error building response: Event is nil")
	}

	eventDto, err := NewEventDTO(eventModel)
	if err != nil {
		return nil, err
	}

	eventResponse := EventResponse{
		EventDTO: *eventDto,
	}

	return &eventResponse, nil
}

func NewEventsResponse(events []*models.Event) ([]*EventResponse, error) {
	if len(events) == 0 {
		return []*EventResponse{}, nil
	}

	var eventResponses = make([]*EventResponse, len(events))
	for index := range events {
		eventResponse, err := NewEventResponse(events[index])
		if err != nil {
			return nil, err
		}
		eventResponses[index] = eventResponse
	}
	return eventResponses, nil
}
