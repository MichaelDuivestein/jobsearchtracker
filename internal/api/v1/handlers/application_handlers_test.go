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
			inputRequest:         testutil.ToPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body does not match CreateApplicationRequest",
			inputRequest:         testutil.ToPtr(`{"application_job_title":"test"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty\n",
		},
		{
			testName:             "body CompanyID and RecruiterID is missing",
			inputRequest:         testutil.ToPtr(`{"job_title":"Random job title", "remote_status_type": "hybrid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: CompanyID and RecruiterID cannot both be empty\n",
		},
		{
			testName:             "body JobTitle and JobAdURL is missing",
			inputRequest:         testutil.ToPtr(`{"company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "remote_status_type": "hybrid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: JobTitle and JobAdURL cannot be both be empty\n",
		},
		{
			testName:             "body RemoteStatusType is missing",
			inputRequest:         testutil.ToPtr(`{"company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "job_title":"Random job title"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'RemoteStatusType': RemoteStatusType is invalid\n",
		},
		{
			testName:             "body RemoteStatusType is invalid",
			inputRequest:         testutil.ToPtr(`{"company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "job_title":"Random job title", "remote_status_type":"Blah"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'RemoteStatusType': RemoteStatusType is invalid\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"JobTitle":"random title","remote_status_type":"hybrid"`),
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

// -------- GetAllApplications tests: --------

func TestGetAllApplications_ShouldReturnErrorIfIncludeCompanyIsInvalid(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/get/all?include_company=names", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.GetAllApplications(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_company. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

func TestGetAllApplications_ShouldReturnErrorIfIncludeRecruiterIsInvalid(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/get/all?include_recruiter=names", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.GetAllApplications(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_recruiter. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

func TestGetAllApplications_ShouldReturnErrorIfIncludePersonsIsInvalid(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/get/all?include_persons=names", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.GetAllApplications(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_persons. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

func TestGetAllApplications_ShouldReturnErrorIfIncludeEventsIsInvalid(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/get/all?include_events=names", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.GetAllApplications(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_events. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

// -------- UpdateApplication tests: --------

func TestUpdateApplication_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "body does not match UpdateApplicationRequest",
			inputRequest:         testutil.ToPtr(`{"application_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "job_title": "title"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body ID is missing",
			inputRequest:         testutil.ToPtr(`{"application_type":"Other"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body RemoteStatusType is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "job_title": "Job Title", "remote_status_type": "Blah"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error on field 'RemoteStatusType': RemoteStatusType is invalid\n",
		},
		{
			testName:             "body is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "companyID":"8abb5944-761b-447c-8a77-11ba1108ff68"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: nothing to update\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"JobTitle":"Entitled","application_type":"developer"`),
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

	applicationHandler := v1.NewApplicationHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/application/update", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			applicationHandler.UpdateApplication(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// -------- DeleteApplication tests: --------

func TestDeleteApplication_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/application/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	applicationHandler.DeleteApplication(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "application ID is empty\n", responseBodyString)
}

func TestDeleteApplication_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	applicationHandler := v1.NewApplicationHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/application/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "Some text",
	}
	request = mux.SetURLVars(request, vars)

	applicationHandler.DeleteApplication(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "application ID is not a valid UUID\n", responseBodyString)
}
