package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- CreateEventRequest tests: --------

func TestCreateEventRequestValidate_ShouldValidateRequest(t *testing.T) {
	request := CreateEventRequest{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 1, 0),
	}
	err := request.validate()
	assert.NoError(t, err)
}

func TestCreateEventRequestValidate_ShouldValidateRequestWithOnlyRequiredFields(t *testing.T) {
	request := CreateEventRequest{
		EventType: EventTypeApplied,
		EventDate: time.Now().AddDate(0, 1, 0),
	}
	err := request.validate()
	assert.NoError(t, err)
}

func TestCreateEventRequestValidate_ShouldReturnValidationErrors(t *testing.T) {
	tests := []struct {
		testName             string
		id                   *uuid.UUID
		eventType            EventType
		eventDate            time.Time
		expectedErrorMessage string
	}{
		{
			"Nil ID",
			testutil.ToPtr(uuid.Nil),
			EventTypeApplied,
			time.Now(),
			"validation error on field 'id': event ID is empty. It should either be 'nil' or a valid UUID"},
		{
			"Empty UUID",
			testutil.ToPtr(uuid.UUID{}),
			EventTypeApplied,
			time.Now(),
			"validation error on field 'id': event ID is empty. It should either be 'nil' or a valid UUID"},
		{
			"Empty EventType",
			testutil.ToPtr(uuid.New()),
			"validation error on field 'eventType': event type is invalid",
			time.Now(),
			"validation error on field 'eventType': event type is invalid"},
		{
			"invalid EventType",
			testutil.ToPtr(uuid.New()),
			"Invalid",
			time.Now(),
			"validation error on field 'eventType': event type is invalid"},
		{
			"Empty eventDate",
			testutil.ToPtr(uuid.New()),
			EventTypeApplied,
			time.Time{},
			"validation error on field 'eventDate': event date is zero. It should be a recent date"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			idToUse := test.id
			if idToUse == nil {
				idToUse = testutil.ToPtr(uuid.New())
			}

			request := CreateEventRequest{
				ID:        idToUse,
				EventType: test.eventType,
				EventDate: test.eventDate,
			}
			err := request.validate()
			assert.Error(t, err)

			var validationError *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationError))
			assert.Equal(t, test.expectedErrorMessage, validationError.Error())
		})
	}
}

func TestCreateEventRequestToModel_ShouldConvertToModel(t *testing.T) {
	request := CreateEventRequest{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 1, 0),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.ID, model.ID)
	assert.Equal(t, request.EventType.String(), model.EventType.String())
	assert.Equal(t, request.Description, model.Description)
	assert.Equal(t, request.Notes, model.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &request.EventDate, &model.EventDate)
	assert.Nil(t, model.CreatedDate)
	assert.Nil(t, model.UpdatedDate)
}

func TestCreateEventRequestToModel_ShouldConvertToModelWithOnlyRequiredFields(t *testing.T) {
	request := CreateEventRequest{
		EventType: EventTypeApplied,
		EventDate: time.Now().AddDate(0, 1, 0),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Nil(t, request.ID)
	assert.Equal(t, request.EventType.String(), model.EventType.String())
	assert.Nil(t, model.Description)
	assert.Nil(t, model.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &request.EventDate, &model.EventDate)
	assert.Nil(t, model.CreatedDate)
	assert.Nil(t, model.UpdatedDate)
}

// -------- UpdateEventRequest tests: --------

func TestUpdateEventRequestValidate_ShouldValidateRequest(t *testing.T) {
	var eventType EventType = EventTypeApplied
	request := UpdateEventRequest{
		ID:          uuid.New(),
		EventType:   &eventType,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	err := request.validate()
	assert.NoError(t, err)
}

func TestUpdateEventRequestValidate_ShouldReturnErrorIfNothingToUpdate(t *testing.T) {
	request := UpdateEventRequest{
		ID: uuid.New(),
	}
	err := request.validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

func TestUpdateEventRequestValidate_ShouldReturnValidationErrorIfEventTypeIsEmpty(t *testing.T) {
	var emptyEventType EventType = ""
	request := UpdateEventRequest{
		ID:        uuid.New(),
		EventType: testutil.ToPtr(emptyEventType),
	}
	err := request.validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'EventType': EventType is invalid", validationError.Error())
}

func TestUpdateEventRequestValidate_ShouldReturnValidationErrorIfEventTypeIsInvalid(t *testing.T) {
	var emptyEventType EventType = "Trouble"
	request := UpdateEventRequest{
		ID:        uuid.New(),
		EventType: testutil.ToPtr(emptyEventType),
	}
	err := request.validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'EventType': EventType is invalid", validationError.Error())
}

func TestUpdateEventRequestValidate_ShouldReturnValidationErrorIfEventDateIsEmpty(t *testing.T) {
	request := UpdateEventRequest{
		ID:        uuid.New(),
		EventDate: &time.Time{},
	}
	err := request.validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'eventDate': event date is zero. It should either be `nil` or a recent date",
		validationError.Error())
}

func TestUpdateEventRequestValidate_ShouldReturnValidationErrorIfNothingToUpdate(t *testing.T) {
	request := UpdateEventRequest{
		ID: uuid.New(),
	}
	err := request.validate()
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

func TestUpdateEventRequestToModel_ShouldConvertToModel(t *testing.T) {
	var eventType EventType = EventTypeApplied
	request := UpdateEventRequest{
		ID:          uuid.New(),
		EventType:   &eventType,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}

	model, err := request.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, request.ID, model.ID)
	assert.Equal(t, request.EventType.String(), model.EventType.String())
	assert.Equal(t, request.Description, model.Description)
	assert.Equal(t, request.Notes, model.Notes)
	testutil.AssertEqualFormattedDateTimes(t, request.EventDate, model.EventDate)
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

func TestEventTypeToModel_ShouldConvertToModel(t *testing.T) {
	tests := []struct {
		testName               string
		eventType              EventType
		expectedModelEventType models.EventType
	}{
		{"EventTypeApplied", EventTypeApplied, models.EventTypeApplied},
		{"EventTypeCallBooked", EventTypeCallBooked, models.EventTypeCallBooked},
		{"EventTypeCallCompleted", EventTypeCallCompleted, models.EventTypeCallCompleted},
		{"EventTypeCodeTestCompleted", EventTypeCodeTestCompleted, models.EventTypeCodeTestCompleted},
		{"EventTypeCodeTestReceived", EventTypeCodeTestReceived, models.EventTypeCodeTestReceived},
		{"EventTypeInterviewBooked", EventTypeInterviewBooked, models.EventTypeInterviewBooked},
		{"EventTypeInterviewCompleted", EventTypeInterviewCompleted, models.EventTypeInterviewCompleted},
		{"EventTypePaused", EventTypePaused, models.EventTypePaused},
		{"EventTypeOffer", EventTypeOffer, models.EventTypeOffer},
		{"EventTypeOther", EventTypeOther, models.EventTypeOther},
		{"EventTypeRecruiterInterviewBooked", EventTypeRecruiterInterviewBooked, models.EventTypeRecruiterInterviewBooked},
		{"EventTypeRecruiterInterviewCompleted", EventTypeRecruiterInterviewCompleted, models.EventTypeRecruiterInterviewCompleted},
		{"EventTypeRejected", EventTypeRejected, models.EventTypeRejected},
		{"EventTypeSigned", EventTypeSigned, models.EventTypeSigned},
		{"EventTypeWithdrew", EventTypeWithdrew, models.EventTypeWithdrew},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			modelEventType, err := test.eventType.ToModel()
			assert.NoError(t, err)
			assert.NotNil(t, modelEventType)
			assert.Equal(t, test.expectedModelEventType.String(), modelEventType.String())
			assert.Equal(t, test.eventType.String(), modelEventType.String())
		})
	}
}

func TestEventTypeToModel_ShouldReturnValidationErrorOnEmptyEventType(t *testing.T) {
	empty := EventType("")
	emptyModel, err := empty.ToModel()
	assert.Error(t, err)
	assert.NotNil(t, emptyModel)

	assert.Equal(t, "", emptyModel.String())

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'EventType': invalid EventType: ''", validationError.Error())
}

func TestEventTypeToModel_ShouldReturnValidationErrorOnInvalidEventType(t *testing.T) {
	empty := EventType("Unknown")
	emptyModel, err := empty.ToModel()
	assert.Error(t, err)
	assert.NotNil(t, emptyModel)

	assert.Equal(t, "", emptyModel.String())

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'EventType': invalid EventType: 'Unknown'", validationError.Error())
}

func TestNewEventType_ShouldConvertFromModel(t *testing.T) {
	tests := []struct {
		testName       string
		modelEventType models.EventType
		eventType      EventType
	}{
		{"EventTypeApplied", models.EventTypeApplied, EventTypeApplied},
		{"EventTypeCallBooked", models.EventTypeCallBooked, EventTypeCallBooked},
		{"EventTypeCallCompleted", models.EventTypeCallCompleted, EventTypeCallCompleted},
		{"EventTypeCodeTestCompleted", models.EventTypeCodeTestCompleted, EventTypeCodeTestCompleted},
		{"EventTypeCodeTestReceived", models.EventTypeCodeTestReceived, EventTypeCodeTestReceived},
		{"EventTypeInterviewBooked", models.EventTypeInterviewBooked, EventTypeInterviewBooked},
		{"EventTypeInterviewCompleted", models.EventTypeInterviewCompleted, EventTypeInterviewCompleted},
		{"EventTypePaused", models.EventTypePaused, EventTypePaused},
		{"EventTypeOffer", models.EventTypeOffer, EventTypeOffer},
		{"EventTypeOther", models.EventTypeOther, EventTypeOther},
		{"EventTypeRecruiterInterviewBooked", models.EventTypeRecruiterInterviewBooked, EventTypeRecruiterInterviewBooked},
		{"EventTypeRecruiterInterviewCompleted", models.EventTypeRecruiterInterviewCompleted, EventTypeRecruiterInterviewCompleted},
		{"EventTypeRejected", models.EventTypeRejected, EventTypeRejected},
		{"EventTypeSigned", models.EventTypeSigned, EventTypeSigned},
		{"EventTypeWithdrew", models.EventTypeWithdrew, EventTypeWithdrew},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			eventType, err := NewEventType(&test.modelEventType)
			assert.NoError(t, err)
			assert.NotNil(t, eventType)
			assert.Equal(t, test.eventType.String(), eventType.String())
			assert.Equal(t, test.modelEventType.String(), eventType.String())
		})
	}
}

func TestNewEventType_ShouldReturnInternalServiceErrorOnNilEventType(t *testing.T) {
	eventType, err := NewEventType(nil)
	assert.NotNil(t, eventType)
	assert.Error(t, err)

	assert.Equal(t, "", eventType.String())

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error trying to convert internal eventType to external EventType.",
		internalServiceError.Error())
}

func TestNewEventType_ShouldReturnInternalServiceErrorOnEmptyEventType(t *testing.T) {
	var eventType models.EventType = ""

	model, err := NewEventType(&eventType)
	assert.NotNil(t, model)
	assert.Error(t, err)

	assert.Equal(t, "", model.String())

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: ''",
		internalServiceError.Error())
}

func TestNewEventType_ShouldReturnInternalServiceErrorOnInvalidEventType(t *testing.T) {
	var eventType models.EventType = "Broken"

	model, err := NewEventType(&eventType)
	assert.NotNil(t, model)
	assert.Error(t, err)

	assert.Equal(t, "", model.String())

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(
		t,
		"internal service error: Error converting internal EventType to external EventType: 'Broken'",
		internalServiceError.Error())
}
