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

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "body does not match CreatePersonRequest",
			inputRequest:         testutil.ToPtr("{\"person_name\":\"test\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n",
		},
		{
			testName:             "body Name is missing",
			inputRequest:         testutil.ToPtr("{\"person_type\":\"CEO\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n",
		},
		{
			testName:             "body PersonType is missing",
			inputRequest:         testutil.ToPtr("{\"name\":\"test\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'PersonType': PersonType is invalid\n",
		},
		{
			testName:             "body PersonType is invalid",
			inputRequest:         testutil.ToPtr("{\"name\":\"test\", \"person_type\":\"Blah\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'PersonType': PersonType is invalid\n",
		},
		{
			testName:             "body is invalid",
			inputRequest:         testutil.ToPtr("{\"person_name\":\"test\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"name":"random name","person_type":"Other"`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}

	for _, test := range tests {
		personHandler := v1.NewPersonHandler(nil)
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			personHandler.CreatePerson(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Contains(t, responseBodyString, test.expectedErrorMessage)
		})

	}
}

// -------- GetPersonById tests: --------

func TestGetPersonById_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	personHandler := v1.NewPersonHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	personHandler.GetPersonByID(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "person ID is empty\n", responseBodyString)
}

func TestGetPersonById_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	personHandler := v1.NewPersonHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "ID GOES HERE",
	}
	request = mux.SetURLVars(request, vars)

	personHandler.GetPersonByID(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "person ID is not a valid UUID\n", responseBodyString)
}

// -------- GetPersonsByName tests: --------

func TestGetPersonsByName_ShouldReturnErrorIfNameIsEmpty(t *testing.T) {
	personHandler := v1.NewPersonHandler(nil)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/person/get/name", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "",
	}
	request = mux.SetURLVars(request, vars)

	personHandler.GetPersonsByName(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "person Name is empty\n", responseBodyString)
}

// -------- UpdatePerson tests: --------

func TestUpdatePerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "body does not match UpdatePersonRequest",
			inputRequest:         testutil.ToPtr(`{"person_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "notes": "Notes"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body ID is missing",
			inputRequest:         testutil.ToPtr(`{"person_type":"Other"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body PersonType is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "name":"random company name","person_type":"Blah"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error on field 'PersonType': PersonType is invalid\n",
		},
		{
			testName:             "body is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "person_name":"Mark Droog"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: nothing to update\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"name":"random company name","person_type":"developer"`),
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

	personHandler := v1.NewPersonHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/person/update", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			personHandler.UpdatePerson(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// -------- DeletePerson tests: --------

func TestDeletePerson_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	personHandler := v1.NewPersonHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	personHandler.DeletePerson(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "person ID is empty\n", responseBodyString)
}

func TestDeletePerson_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	personHandler := v1.NewPersonHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "Some text",
	}
	request = mux.SetURLVars(request, vars)

	personHandler.DeletePerson(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "person ID is not a valid UUID\n", responseBodyString)
}
