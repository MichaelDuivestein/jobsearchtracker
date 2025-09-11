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

// -------- CreateApplication tests: --------

func TestCreateApplication_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			inputRequest:         testutil.StringPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body does not match CreateApplicationRequest",
			inputRequest:         testutil.StringPtr(`{"application_job_title":"test"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty\n",
		},
		{
			testName:             "body CompanyID and RecruiterID is missing",
			inputRequest:         testutil.StringPtr(`{"job_title":"Random job title", "remote_status_type": "hybrid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty\n",
		},
		{
			testName:             "body JobTitle and JobAdURL is missing",
			inputRequest:         testutil.StringPtr(`{"company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "remote_status_type": "hybrid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: JobTitle and JobAdURL cannot be both be empty\n",
		},
		{
			testName:             "body RemoteStatusType is missing",
			inputRequest:         testutil.StringPtr(`{"company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "job_title":"Random job title"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'RemoteStatusType': RemoteStatusType is invalid\n",
		},
		{
			testName:             "body RemoteStatusType is invalid",
			inputRequest:         testutil.StringPtr(`{"company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "job_title":"Random job title", "remote_status_type":"Blah"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'RemoteStatusType': RemoteStatusType is invalid\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.StringPtr(`"JobTitle":"random title","remote_status_type":"hybrid"`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}

	for _, test := range tests {
		applicationHandler := v1.NewApplicationHandler(nil)
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			applicationHandler.CreateApplication(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Contains(t, responseBodyString, test.expectedErrorMessage)
		})

	}
}

// -------- GetApplicationById tests: --------

func TestGetApplicationById_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	applicationHandler.GetApplicationByID(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "application ID is empty\n", responseBodyString)
}

func TestGetApplicationById_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "ID GOES HERE",
	}
	request = mux.SetURLVars(request, vars)

	applicationHandler.GetApplicationByID(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "application ID is not a valid UUID\n", responseBodyString)
}

// -------- GetApplicationsByJobTitle tests: --------

func TestGetApplicationsByJobTitle_ShouldReturnErrorIfNameIsEmpty(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/get/title", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "",
	}
	request = mux.SetURLVars(request, vars)

	applicationHandler.GetApplicationsByJobTitle(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "job title is empty\n", responseBodyString)
}
