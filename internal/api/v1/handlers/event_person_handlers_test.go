package handlers

import (
	"bytes"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- AssociateEventPerson tests: --------

func TestAssociateEventPerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "body is empty",
			inputRequest:         testutil.ToPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "event_id is missing",
			inputRequest:         testutil.ToPtr(`{"person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: EventID is invalid\n"},
		{
			testName:             "event_id is empty",
			inputRequest:         testutil.ToPtr(`{"event_id": "", "person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "event_id is invalid",
			inputRequest:         testutil.ToPtr(`{"event_id": "not valid", "person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "person_id is missing",
			inputRequest:         testutil.ToPtr(`{"event_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: PersonID is invalid\n"},
		{
			testName:             "person_id is empty",
			inputRequest:         testutil.ToPtr(`{"event_id": "06f92026-5b76-431a-909d-005ae920f4e4", "person_id": ""}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "person_id is invalid",
			inputRequest:         testutil.ToPtr(`{"event_id": "06f92026-5b76-431a-909d-005ae920f4e4", "person_id": "not valid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}
	handler := NewEventPersonHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest("POST", "/api/v1/event-person/associate", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.AssociateEventPerson(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}

}

// -------- GetEventPersonsByID tests: --------

func TestGetEventPersonsByID_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		queryParams          string
		expectedErrorMessage string
	}{
		{
			testName:             "nil eventID and nil personID",
			queryParams:          "",
			expectedErrorMessage: "EventID and/or PersonID are required\n",
		},
		{
			testName:             "empty eventID and empty personID",
			queryParams:          `?event_id=&person_id=`,
			expectedErrorMessage: "EventID and/or PersonID are required\n",
		},
		{
			testName:             "empty eventID and nil personID",
			queryParams:          `?event_id=`,
			expectedErrorMessage: "EventID and/or PersonID are required\n",
		},
		{
			testName:             "nil eventID and empty personID",
			queryParams:          `?person_id=`,
			expectedErrorMessage: "EventID and/or PersonID are required\n",
		},
		{
			testName:             "invalid eventID",
			queryParams:          `?event_id=not-valid&person_id=8b802e50-f164-4d92-9f27-8cd91167f1e8`,
			expectedErrorMessage: "EventID and/or PersonID are required\n",
		},
		{
			testName:             "invalid personID",
			queryParams:          `?event_id=06f92026-5b76-431a-909d-005ae920f4e4&person_id=not-valid`,
			expectedErrorMessage: "EventID and/or PersonID are required\n",
		},
	}

	handler := NewEventPersonHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			request, err := http.NewRequest(http.MethodGet, "/api/v1/event-person/get"+test.queryParams, nil)
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.GetEventPersonsByID(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// --------DeleteEventPerson tests: --------

func TestDeleteEventPerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		body                 string
		expectedResponseCode int
		expectedErrorMessage string
	}{
		{
			testName:             "empty body",
			body:                 "",
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "empty eventID and empty personID",
			body:                 `{"event_id":"", "person_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "empty eventID and nil personID",
			body:                 `"{event_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil eventID and empty personID",
			body:                 `{"person_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "invalid eventID",
			body:                 `"event_id":"not valid","person_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}"`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil eventID",
			body:                 `{"person_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}`,
			expectedErrorMessage: "validation error: EventID is invalid\n",
		},
		{
			testName:             "invalid personID",
			body:                 `{"event_id":"06f92026-5b76-431a-909d-005ae920f4e4","person_id":"not valid"}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil personID",
			body:                 `{"event_id":"06f92026-5b76-431a-909d-005ae920f4e4"}"`,
			expectedErrorMessage: "validation error: PersonID is invalid\n",
		},
	}
	handler := NewEventPersonHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			requestBody := []byte(test.body)

			request, err := http.NewRequest(
				http.MethodGet, "/api/v1/event-person/get",
				bytes.NewReader(requestBody))

			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.DeleteEventPerson(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}
