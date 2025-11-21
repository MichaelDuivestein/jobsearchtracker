package handlers_test

import (
	"bytes"
	v1 "jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// -------- CreateEvent tests: --------

func TestCreateEvent_ShouldRespondWithBadStatusRequest(t *testing.T) {
	tests := []struct {
		testName             string
		inputRequest         *string
		expectedResponseCode int
		expectedErrorMessage string
	}{
		{
			testName:             "body is nil",
			inputRequest:         nil,
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body is empty",
			inputRequest:         testutil.ToPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body does not match CreateEventRequest",
			inputRequest:         testutil.ToPtr(`{"type":"applied"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'eventType': event type is invalid\n",
		},
		{
			testName:             "body EventType is missing",
			inputRequest:         testutil.ToPtr(`{"event_date":"2025-11-12T13:56:41.012Z"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'eventType': event type is invalid\n",
		},
		{
			testName:             "body EventType is invalid",
			inputRequest:         testutil.ToPtr(`{"event_type":"NotSure", "event_date":"2025-11-12T13:56:41.012Z"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'eventType': event type is invalid\n",
		},
		{
			testName:             "body EventDate is missing",
			inputRequest:         testutil.ToPtr(`{"event_type":"applied"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'eventDate': event date is zero. It should be a recent date\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"event_type":"applied","event_date":"2025-11-12T13:56:41.012Z"`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}
	for _, test := range tests {
		eventHandler := v1.NewEventHandler(nil)
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/event/new", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			eventHandler.CreateEvent(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Contains(t, responseBodyString, test.expectedErrorMessage)
		})
	}
}

// -------- GetEventById tests: --------

func TestGetEventById_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	eventHandler := v1.NewEventHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	eventHandler.GetEventByID(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "event ID is empty\n", responseBodyString)
}

func TestGetEventById_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	eventHandler := v1.NewEventHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "ID GOES HERE",
	}
	request = mux.SetURLVars(request, vars)

	eventHandler.GetEventByID(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "event ID is not a valid UUID\n", responseBodyString)
}

// -------- GetAllEvents tests: --------

func TestGetAllEvents_ShouldReturnErrorIfIncludeApplicationsIsInvalid(t *testing.T) {
	eventHandler := v1.NewEventHandler(nil)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_applications=maybe", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_applications. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

func TestGetAllEvents_ShouldReturnErrorIfIncludeCompaniesIsInvalid(t *testing.T) {
	eventHandler := v1.NewEventHandler(nil)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_companies=maybe", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_companies. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

func TestGetAllEvents_ShouldReturnErrorIfIncludePersonsIsInvalid(t *testing.T) {
	eventHandler := v1.NewEventHandler(nil)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_persons=maybe", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_persons. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

// -------- UpdateEvent tests: --------

func TestUpdateEvent_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		inputRequest         *string
		expectedResponseCode int
		expectedErrorMessage string
	}{
		{
			testName:             "body is nil",
			inputRequest:         nil,
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body is empty",
			inputRequest:         testutil.ToPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body does not match UpdateEventRequest",
			inputRequest:         testutil.ToPtr(`{"event_id":"8abb5944-761b-447c-8a77-11ba1108ff68", "Notes":"SomeNotes"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body ID is missing",
			inputRequest:         testutil.ToPtr(`{"Notes":"New Notes"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body EventType is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "event_type":"Blah"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error on field 'EventType': EventType is invalid\n",
		},
		{
			testName:             "body EventType is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "event_Type":"ghosted"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error on field 'EventType': EventType is invalid\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"ID":"8abb5944-761b-447c-8a77-11ba1108ff68","event_type":"applied"`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body contains no fields to update",
			inputRequest:         testutil.ToPtr(`{"id":"8abb5944-761b-447c-8a77-11ba1108ff68"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: nothing to update\n",
		},
	}

	eventHandler := v1.NewEventHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/event/update", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			eventHandler.UpdateEvent(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// -------- DeleteEvent tests: --------

func TestDeleteEvent_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	eventHandler := v1.NewEventHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/event/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	eventHandler.DeleteEvent(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "event ID is empty\n", responseBodyString)
}

func TestDeleteEvent_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	eventHandler := v1.NewEventHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/event/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "Some text",
	}
	request = mux.SetURLVars(request, vars)

	eventHandler.DeleteEvent(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "event ID is not a valid UUID\n", responseBodyString)
}
