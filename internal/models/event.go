package models

import (
	"jobsearchtracker/internal/errors"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID
	EventType   *EventType
	Description *string
	Notes       *string
	EventDate   *time.Time
	CreatedDate *time.Time
	UpdatedDate *time.Time
}

type CreateEvent struct {
	ID          *uuid.UUID
	EventType   EventType
	Description *string
	Notes       *string
	EventDate   time.Time
	CreatedDate *time.Time
	UpdatedDate *time.Time
}

func (event CreateEvent) Validate() error {
	if event.ID != nil && *event.ID == uuid.Nil {
		name := "id"
		return errors.NewValidationError(
			&name,
			"event ID is empty. It should either be 'nil' or a valid UUID")
	}

	if !event.EventType.isValid() {
		name := "eventType"
		return errors.NewValidationError(
			&name,
			"event type is invalid")
	}

	if event.EventDate.IsZero() {
		updatedDate := "eventDate"
		return errors.NewValidationError(
			&updatedDate,
			"event date is zero. It should be a recent date")
	}

	if event.CreatedDate != nil && event.CreatedDate.IsZero() {
		createdDate := "createdDate"
		return errors.NewValidationError(
			&createdDate,
			"created date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	if event.UpdatedDate != nil && event.UpdatedDate.IsZero() {
		updatedDate := "updatedDate"
		return errors.NewValidationError(
			&updatedDate,
			"updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil")
	}

	return nil
}

type UpdateEvent struct {
	ID          uuid.UUID
	EventType   *EventType
	Description *string
	Notes       *string
	EventDate   *time.Time
}

func (event UpdateEvent) Validate() error {
	if event.ID == uuid.Nil {
		name := "id"
		return errors.NewValidationError(
			&name,
			"event ID is empty. It should either be 'nil' or a valid UUID")
	}

	if event.EventType != nil && !event.EventType.isValid() {
		name := "eventType"
		return errors.NewValidationError(
			&name,
			"event type is invalid")
	}

	if event.EventDate != nil && event.EventDate.IsZero() {
		updatedDate := "eventDate"
		return errors.NewValidationError(
			&updatedDate,
			"event date is zero. It should either be 'nil' or a recent date")
	}

	return nil
}

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
