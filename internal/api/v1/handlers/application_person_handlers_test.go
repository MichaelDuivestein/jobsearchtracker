package handlers

import (
	"bytes"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- AssociateApplicationPerson tests: --------

func TestAssociateApplicationPerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			inputRequest:         testutil.ToPtr(`{"person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: ApplicationID is invalid\n"},
		{
			testName:             "application_id is empty",
			inputRequest:         testutil.ToPtr(`{"application_id": "", "person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "application_id is invalid",
			inputRequest:         testutil.ToPtr(`{"application_id": "not valid", "person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "person_id is missing",
			inputRequest:         testutil.ToPtr(`{"application_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: PersonID is invalid\n"},
		{
			testName:             "person_id is empty",
			inputRequest:         testutil.ToPtr(`{"application_id": "06f92026-5b76-431a-909d-005ae920f4e4", "person_id": ""}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "person_id is invalid",
			inputRequest:         testutil.ToPtr(`{"application_id": "06f92026-5b76-431a-909d-005ae920f4e4", "person_id": "not valid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}
	handler := NewApplicationPersonHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest("POST", "/api/v1/application-person/associate", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.AssociateApplicationPerson(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}

}

// -------- GetApplicationPersonsByID tests: --------

func TestGetApplicationPersonsByID_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		queryParams          string
		expectedErrorMessage string
	}{
		{
			testName:             "nil applicationID and nil personID",
			queryParams:          "",
			expectedErrorMessage: "ApplicationID and/or PersonID are required\n",
		},
		{
			testName:             "empty applicationID and empty personID",
			queryParams:          `?application_id=&person_id=`,
			expectedErrorMessage: "ApplicationID and/or PersonID are required\n",
		},
		{
			testName:             "empty applicationID and nil personID",
			queryParams:          `?application_id=`,
			expectedErrorMessage: "ApplicationID and/or PersonID are required\n",
		},
		{
			testName:             "nil applicationID and empty personID",
			queryParams:          `?person_id=`,
			expectedErrorMessage: "ApplicationID and/or PersonID are required\n",
		},
		{
			testName:             "invalid applicationID",
			queryParams:          `?application_id=not-valid&person_id=8b802e50-f164-4d92-9f27-8cd91167f1e8`,
			expectedErrorMessage: "ApplicationID and/or PersonID are required\n",
		},
		{
			testName:             "invalid personID",
			queryParams:          `?application_id=06f92026-5b76-431a-909d-005ae920f4e4&person_id=not-valid`,
			expectedErrorMessage: "ApplicationID and/or PersonID are required\n",
		},
	}

	handler := NewApplicationPersonHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			request, err := http.NewRequest(http.MethodGet, "/api/v1/application-person/get"+test.queryParams, nil)
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.GetApplicationPersonsByID(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// --------DeleteApplicationPerson tests: --------

func TestDeleteApplicationPerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "empty applicationID and empty personID",
			body:                 `{"application_id":"", "person_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "empty applicationID and nil personID",
			body:                 `"{application_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil applicationID and empty personID",
			body:                 `{"person_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "invalid applicationID",
			body:                 `"application_id":"not valid","person_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}"`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil applicationID",
			body:                 `{"person_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}`,
			expectedErrorMessage: "validation error: ApplicationID is invalid\n",
		},
		{
			testName:             "invalid personID",
			body:                 `{"application_id":"06f92026-5b76-431a-909d-005ae920f4e4","person_id":"not valid"}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil personID",
			body:                 `{"application_id":"06f92026-5b76-431a-909d-005ae920f4e4"}"`,
			expectedErrorMessage: "validation error: PersonID is invalid\n",
		},
	}
	handler := NewApplicationPersonHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			requestBody := []byte(test.body)

			request, err := http.NewRequest(http.MethodGet, "/api/v1/application-person/get", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.DeleteApplicationPerson(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}
