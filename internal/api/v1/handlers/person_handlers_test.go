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
			inputRequest:         testutil.StringPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body does not match CreatePersonRequest",
			inputRequest:         testutil.StringPtr("{\"person_name\":\"test\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n",
		},
		{
			testName:             "body Name is missing",
			inputRequest:         testutil.StringPtr("{\"person_type\":\"CEO\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n",
		},
		{
			testName:             "body PersonType is missing",
			inputRequest:         testutil.StringPtr("{\"name\":\"test\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'PersonType': PersonType is invalid\n",
		},
		{
			testName:             "body PersonType is invalid",
			inputRequest:         testutil.StringPtr("{\"name\":\"test\", \"person_type\":\"Blah\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'PersonType': PersonType is invalid\n",
		},
		{
			testName:             "body is invalid",
			inputRequest:         testutil.StringPtr("{\"person_name\":\"test\"}"),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.StringPtr(`"name":"random name","person_type":"Other"`),
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
