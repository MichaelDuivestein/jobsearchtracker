package handlers

import (
	"bytes"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- AssociateApplicationEvent tests: --------

func TestAssociateApplicationEvent_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "application_id is missing",
			inputRequest:         testutil.ToPtr(`{"event_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: ApplicationID is invalid\n"},
		{
			testName:             "application_id is empty",
			inputRequest:         testutil.ToPtr(`{"application_id": "", "event_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "application_id is invalid",
			inputRequest:         testutil.ToPtr(`{"application_id": "not valid", "event_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "event_id is missing",
			inputRequest:         testutil.ToPtr(`{"application_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: EventID is invalid\n"},
		{
			testName:             "event_id is empty",
			inputRequest:         testutil.ToPtr(`{"application_id": "06f92026-5b76-431a-909d-005ae920f4e4", "event_id": ""}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "event_id is invalid",
			inputRequest:         testutil.ToPtr(`{"application_id": "06f92026-5b76-431a-909d-005ae920f4e4", "event_id": "not valid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}
	handler := NewApplicationEventHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest("POST", "/api/v1/application-event/associate", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.AssociateApplicationEvent(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}

}

// -------- GetApplicationEventsByID tests: --------

func TestGetApplicationEventsByID_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		queryParams          string
		expectedErrorMessage string
	}{
		{
			testName:             "nil applicationID and nil eventID",
			queryParams:          "",
			expectedErrorMessage: "ApplicationID and/or EventID are required\n",
		},
		{
			testName:             "empty applicationID and empty eventID",
			queryParams:          `?application_id=&event_id=`,
			expectedErrorMessage: "ApplicationID and/or EventID are required\n",
		},
		{
			testName:             "empty applicationID and nil eventID",
			queryParams:          `?application_id=`,
			expectedErrorMessage: "ApplicationID and/or EventID are required\n",
		},
		{
			testName:             "nil applicationID and empty eventID",
			queryParams:          `?event_id=`,
			expectedErrorMessage: "ApplicationID and/or EventID are required\n",
		},
		{
			testName:             "invalid applicationID",
			queryParams:          `?application_id=not-valid&event_id=8b802e50-f164-4d92-9f27-8cd91167f1e8`,
			expectedErrorMessage: "ApplicationID and/or EventID are required\n",
		},
		{
			testName:             "invalid eventID",
			queryParams:          `?application_id=06f92026-5b76-431a-909d-005ae920f4e4&event_id=not-valid`,
			expectedErrorMessage: "ApplicationID and/or EventID are required\n",
		},
	}

	handler := NewApplicationEventHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			request, err := http.NewRequest(http.MethodGet, "/api/v1/application-event/get"+test.queryParams, nil)
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.GetApplicationEventsByID(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// --------DeleteApplicationEvent tests: --------

func TestDeleteApplicationEvent_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "empty applicationID and empty eventID",
			body:                 `{"application_id":"", "event_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "empty applicationID and nil eventID",
			body:                 `"{application_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil applicationID and empty eventID",
			body:                 `{"event_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "invalid applicationID",
			body:                 `"application_id":"not valid","event_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}"`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil applicationID",
			body:                 `{"event_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}`,
			expectedErrorMessage: "validation error: ApplicationID is invalid\n",
		},
		{
			testName:             "invalid eventID",
			body:                 `{"application_id":"06f92026-5b76-431a-909d-005ae920f4e4","event_id":"not valid"}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil eventID",
			body:                 `{"application_id":"06f92026-5b76-431a-909d-005ae920f4e4"}"`,
			expectedErrorMessage: "validation error: EventID is invalid\n",
		},
	}
	handler := NewApplicationEventHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			requestBody := []byte(test.body)

			request, err := http.NewRequest(
				http.MethodGet, "/api/v1/application-event/get",
				bytes.NewReader(requestBody))

			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.DeleteApplicationEvent(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}
