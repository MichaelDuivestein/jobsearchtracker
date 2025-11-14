package models

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreateEvent.Validate tests: --------

func TestCreateEventValidate_ShouldReturnNilIfCreateEventIsValid(t *testing.T) {
	event := CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   EventTypeApplied,
		Description: testutil.ToPtr("description"),
		Notes:       testutil.ToPtr("notes"),
		EventDate:   time.Now().AddDate(0, 1, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 3, 0)),
	}
	err := event.Validate()
	assert.NoError(t, err)
}

func TestCreateEventValidate_ShouldReturnNilIfOnlyRequiredFieldsAreFilled(t *testing.T) {
	event := CreateEvent{
		EventType: EventTypeApplied,
		EventDate: time.Now().AddDate(0, 1, 0),
	}
	err := event.Validate()
	assert.NoError(t, err)
}

func TestCreateEventValidate_ShouldReturnValidationErrorOnNilEventType(t *testing.T) {
	event := CreateEvent{
		EventDate: time.Now().AddDate(0, 1, 0),
	}
	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'eventType': event type is invalid", validationError.Error())
}

func TestCreateEventValidate_ShouldReturnValidationErrorOnInvalidEventType(t *testing.T) {
	var eventType EventType = "broken"
	event := CreateEvent{
		EventType: eventType,
		EventDate: time.Now().AddDate(0, 1, 0),
	}
	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'eventType': event type is invalid", validationError.Error())
}

func TestCreateEventValidate_ShouldReturnValidationErrorOnNilEventDate(t *testing.T) {
	event := CreateEvent{
		EventType: EventTypeApplied,
	}
	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'eventDate': event date is zero. It should be a recent date", validationError.Error())
}

func TestCreateEventValidate_ShouldReturnValidationErrorOnEmptyCreatedDate(t *testing.T) {
	event := CreateEvent{
		EventType:   EventTypeApplied,
		EventDate:   time.Now().AddDate(0, 1, 0),
		CreatedDate: &time.Time{},
	}
	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'createdDate': created date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

func TestCreateEventValidate_ShouldReturnValidationErrorOnEmptyUpdatedDate(t *testing.T) {
	event := CreateEvent{
		EventType:   EventTypeApplied,
		EventDate:   time.Now().AddDate(0, 1, 0),
		UpdatedDate: &time.Time{},
	}
	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'updatedDate': updated date is zero. It should either be 'nil' or a recent date. Given that this is an insert, it is recommended to use nil",
		validationError.Error())
}

// -------- UpdateEvent.Validate tests: --------

func TestUpdateEventValidate_ShouldReturnNilIfUpdateEventIsValid(t *testing.T) {
	var eventType EventType = EventTypeApplied
	event := UpdateEvent{
		ID:          uuid.New(),
		EventType:   testutil.ToPtr(eventType),
		Description: testutil.ToPtr("description"),
		Notes:       testutil.ToPtr("notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	err := event.Validate()
	assert.NoError(t, err)
}

func TestUpdateEventValidate_ShouldReturnValidationErrorIfIDIsEmpty(t *testing.T) {
	var eventType EventType = EventTypeApplied
	event := UpdateEvent{
		ID:          uuid.UUID{},
		EventType:   testutil.ToPtr(eventType),
		Description: testutil.ToPtr("description"),
		Notes:       testutil.ToPtr("notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'id': event ID is empty. It should either be 'nil' or a valid UUID",
		validationError.Error())
}

func TestUpdateEventValidate_ShouldReturnValidationErrorIfUpdateEventTypeIsNil(t *testing.T) {
	var eventType EventType = "broken"
	event := UpdateEvent{
		ID:        uuid.New(),
		EventType: testutil.ToPtr(eventType),
	}
	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'eventType': event type is invalid", validationError.Error())
}

func TestUpdateEventValidate_ShouldReturnValidationErrorIfEventDateTypeIsEmpty(t *testing.T) {
	event := UpdateEvent{
		ID:        uuid.New(),
		EventDate: &time.Time{},
	}

	err := event.Validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'eventDate': event date is zero. It should either be 'nil' or a recent date",
		validationError.Error())
}

// -------- EventType tests: --------

func TestEventTypeIsValid_ShouldReturnTrue(t *testing.T) {
	applied := EventType(EventTypeApplied)
	assert.True(t, applied.isValid())

	callBooked := EventType(EventTypeCallBooked)
	assert.True(t, callBooked.isValid())

	callCompleted := EventType(EventTypeCallCompleted)
	assert.True(t, callCompleted.isValid())

	codeTestCompleted := EventType(EventTypeCodeTestCompleted)
	assert.True(t, codeTestCompleted.isValid())

	codeTestReceived := EventType(EventTypeCodeTestReceived)
	assert.True(t, codeTestReceived.isValid())

	interviewBooked := EventType(EventTypeInterviewBooked)
	assert.True(t, interviewBooked.isValid())

	interviewCompleted := EventType(EventTypeInterviewCompleted)
	assert.True(t, interviewCompleted.isValid())

	paused := EventType(EventTypePaused)
	assert.True(t, paused.isValid())

	offer := EventType(EventTypeOffer)
	assert.True(t, offer.isValid())

	other := EventType(EventTypeOther)
	assert.True(t, other.isValid())

	recruiterInterviewBooked := EventType(EventTypeRecruiterInterviewBooked)
	assert.True(t, recruiterInterviewBooked.isValid())

	recruiterInterviewCompleted := EventType(EventTypeRecruiterInterviewCompleted)
	assert.True(t, recruiterInterviewCompleted.isValid())

	rejected := EventType(EventTypeRejected)
	assert.True(t, rejected.isValid())

	signed := EventType(EventTypeSigned)
	assert.True(t, signed.isValid())

	withdrew := EventType(EventTypeWithdrew)
	assert.True(t, withdrew.isValid())
}

func TestEventTypeIsValid_ShouldReturnFalseOnEmptyEventType(t *testing.T) {
	var eventType EventType = ""
	assert.False(t, eventType.isValid())
}

func TestEventTypeIsValid_ShouldReturnFalseOnInvalidEventType(t *testing.T) {
	var eventType EventType = "unknown"
	assert.False(t, eventType.isValid())
}
