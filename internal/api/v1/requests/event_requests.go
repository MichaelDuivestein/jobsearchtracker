package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type CreateEventRequest struct {
	ID          *uuid.UUID `json:"id,omitempty" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventType   EventType  `json:"event_type" example:"interviewCompleted" extensions:"x-order=2"`
	Description *string    `json:"description,omitempty" example:"Event Description" extensions:"x-order=2"`
	Notes       *string    `json:"notes,omitempty" example:"Notes go here" extensions:"x-order=3"`
	EventDate   time.Time  `json:"event_date" example:"2025-12-31T23:59Z" extensions:"x-order=4"`
}

func (request *CreateEventRequest) validate() error {
	if request.ID != nil && *request.ID == uuid.Nil {
		name := "id"
		return internalErrors.NewValidationError(
			&name,
			"event ID is empty. It should either be 'nil' or a valid UUID")
	}

	if !request.EventType.isValid() {
		name := "eventType"
		return internalErrors.NewValidationError(
			&name,
			"event type is invalid")
	}

	if request.EventDate.IsZero() {
		updatedDate := "eventDate"
		return internalErrors.NewValidationError(
			&updatedDate,
			"event date is zero. It should be a recent date")
	}

	return nil
}

func (request *CreateEventRequest) ToModel() (*models.CreateEvent, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	eventType, _ := request.EventType.ToModel()

	eventModel := models.CreateEvent{
		ID:          request.ID,
		EventType:   eventType,
		Description: request.Description,
		Notes:       request.Notes,
		EventDate:   request.EventDate,
	}

	return &eventModel, nil
}

type UpdateEventRequest struct {
	ID          uuid.UUID  `json:"id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventType   *EventType `json:"event_type,omitempty" example:"codeTestCompleted" extensions:"x-order=2"`
	Description *string    `json:"description,omitempty" example:"Event Description" extensions:"x-order=2"`
	Notes       *string    `json:"notes,omitempty" example:"Notes go here" extensions:"x-order=3"`
	EventDate   *time.Time `json:"event_date,omitempty" example:"2025-12-31T23:59Z" extensions:"x-order=4"`
}

func (request *UpdateEventRequest) validate() error {
	err := uuid.Validate(request.ID.String())
	if err != nil {
		message := "ID is invalid"
		slog.Info("UpdateEventRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}
	if request.ID == uuid.Nil {
		message := "ID is empty"
		slog.Info("UpdateEventRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventType == nil && request.Description == nil && request.Notes == nil && request.EventDate == nil {
		message := "nothing to update"
		slog.Info("UpdateEventRequest.Validate: "+message, "ID", request.ID)
		return internalErrors.NewValidationError(nil, message)
	}

	if request.EventType != nil && !request.EventType.isValid() {
		message := "EventType is invalid"

		slog.Info("UpdateEventRequest.Validate: "+message, "ID", request.ID)

		personType := "EventType"
		return internalErrors.NewValidationError(&personType, message)
	}

	if request.Description != nil && *request.Description == "" {
		message := "Description is invalid"
		slog.Info("UpdateEventRequest.Validate: "+message, "ID", request.ID)

		companyType := "Description"
		return internalErrors.NewValidationError(&companyType, message)
	}

	if request.Notes != nil && *request.Notes == "" {
		message := "Notes is invalid"
		slog.Info("UpdateEventRequest.Validate: "+message, "ID", request.ID)

		companyType := "Notes"
		return internalErrors.NewValidationError(&companyType, message)
	}

	if request.EventDate != nil && request.EventDate.IsZero() {
		updatedDate := "eventDate"
		return internalErrors.NewValidationError(
			&updatedDate,
			"event date is zero. It should either be `nil` or a recent date")
	}

	return nil
}

func (request *UpdateEventRequest) ToModel() (*models.UpdateEvent, error) {
	err := request.validate()
	if err != nil {
		return nil, err
	}

	eventType, _ := request.EventType.ToModel()

	eventModel := models.UpdateEvent{
		ID:          request.ID,
		EventType:   &eventType,
		Description: request.Description,
		Notes:       request.Notes,
		EventDate:   request.EventDate,
	}

	return &eventModel, nil
}

// EventType represents the type of event.
//
// @enum applied,callBooked,callCompleted,codeTestCompleted,codeTestReceived,interviewBooked,interviewCompleted,paused,offer,other,recruiterInterviewBooked,recruiterInterviewCompleted,rejected,signed,withdrew
type EventType string

const (
	EventTypeApplied                     = "applied"
	EventTypeCallBooked                  = "callBooked"
	EventTypeCallCompleted               = "callCompleted"
	EventTypeCodeTestCompleted           = "codeTestCompleted"
	EventTypeCodeTestReceived            = "codeTestReceived"
	EventTypeInterviewBooked             = "interviewBooked"
	EventTypeInterviewCompleted          = "interviewCompleted"
	EventTypePaused                      = "paused"
	EventTypeOffer                       = "offer"
	EventTypeOther                       = "other"
	EventTypeRecruiterInterviewBooked    = "recruiterInterviewBooked"
	EventTypeRecruiterInterviewCompleted = "recruiterInterviewCompleted"
	EventTypeRejected                    = "rejected"
	EventTypeSigned                      = "signed"
	EventTypeWithdrew                    = "withdrew"
)

func (eventType EventType) isValid() bool {
	switch eventType {
	case EventTypeApplied, EventTypeCallBooked, EventTypeCallCompleted, EventTypeCodeTestCompleted,
		EventTypeCodeTestReceived, EventTypeInterviewBooked, EventTypeInterviewCompleted, EventTypePaused,
		EventTypeOffer, EventTypeOther, EventTypeRecruiterInterviewBooked, EventTypeRecruiterInterviewCompleted,
		EventTypeRejected, EventTypeSigned, EventTypeWithdrew:
		return true
	}
	return false
}

func (eventType EventType) String() string { return string(eventType) }

func (eventType EventType) ToModel() (models.EventType, error) {
	switch eventType {
	case EventTypeApplied:
		return models.EventTypeApplied, nil
	case EventTypeCallBooked:
		return models.EventTypeCallBooked, nil
	case EventTypeCallCompleted:
		return models.EventTypeCallCompleted, nil
	case EventTypeCodeTestCompleted:
		return models.EventTypeCodeTestCompleted, nil
	case EventTypeCodeTestReceived:
		return models.EventTypeCodeTestReceived, nil
	case EventTypeInterviewBooked:
		return models.EventTypeInterviewBooked, nil
	case EventTypeInterviewCompleted:
		return models.EventTypeInterviewCompleted, nil
	case EventTypePaused:
		return models.EventTypePaused, nil
	case EventTypeOffer:
		return models.EventTypeOffer, nil
	case EventTypeOther:
		return models.EventTypeOther, nil
	case EventTypeRecruiterInterviewBooked:
		return models.EventTypeRecruiterInterviewBooked, nil
	case EventTypeRecruiterInterviewCompleted:
		return models.EventTypeRecruiterInterviewCompleted, nil
	case EventTypeRejected:
		return models.EventTypeRejected, nil
	case EventTypeSigned:
		return models.EventTypeSigned, nil
	case EventTypeWithdrew:
		return models.EventTypeWithdrew, nil
	default:
		slog.Info("v1.types.toModel: Invalid EventType: '" + eventType.String() + "'")
		personTypeString := "EventType"
		return "", internalErrors.NewValidationError(
			&personTypeString,
			"invalid EventType: '"+eventType.String()+"'")
	}
}

func NewEventType(modelEventType *models.EventType) (EventType, error) {
	if modelEventType == nil {
		slog.Info("v1.types.NewEventType: modelEventType is nil")
		return "", internalErrors.NewInternalServiceError(
			"Error trying to convert internal eventType to external EventType.")
	}

	switch *modelEventType {
	case models.EventTypeApplied:
		return EventTypeApplied, nil
	case models.EventTypeCallBooked:
		return EventTypeCallBooked, nil
	case models.EventTypeCallCompleted:
		return EventTypeCallCompleted, nil
	case models.EventTypeCodeTestCompleted:
		return EventTypeCodeTestCompleted, nil
	case models.EventTypeCodeTestReceived:
		return EventTypeCodeTestReceived, nil
	case models.EventTypeInterviewBooked:
		return EventTypeInterviewBooked, nil
	case models.EventTypeInterviewCompleted:
		return EventTypeInterviewCompleted, nil
	case models.EventTypePaused:
		return EventTypePaused, nil
	case models.EventTypeOffer:
		return EventTypeOffer, nil
	case models.EventTypeOther:
		return EventTypeOther, nil
	case models.EventTypeRecruiterInterviewBooked:
		return EventTypeRecruiterInterviewBooked, nil
	case models.EventTypeRecruiterInterviewCompleted:
		return EventTypeRecruiterInterviewCompleted, nil
	case models.EventTypeRejected:
		return EventTypeRejected, nil
	case models.EventTypeSigned:
		return EventTypeSigned, nil
	case models.EventTypeWithdrew:
		return EventTypeWithdrew, nil
	default:
		slog.Info("v1.types.NewEventType: Invalid modelEventType: '" + modelEventType.String() + "'")
		return "", internalErrors.NewInternalServiceError(
			"Error converting internal EventType to external EventType: '" + modelEventType.String() + "'")
	}
}
