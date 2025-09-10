package handlers_test

import (
	"bytes"
	"encoding/json"
	v1 "jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
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

// -------- GetAllPersons tests: --------

func TestGetAllPersons_ShouldReturnAllPersons(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// insert persons

	firstID := uuid.New()
	email := "Person1 Email"
	phone := "1111111"
	notes := "Person1 Notes"
	firstRequestBody := requests.CreatePersonRequest{
		ID:         &firstID,
		Name:       "Person1",
		PersonType: requests.PersonTypeCTO,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, firstRequestBody)

	secondID := uuid.New()
	email = "Person2 Email"
	phone = "222222"
	notes = "Person2 Notes"
	secondRequestBody := requests.CreatePersonRequest{
		ID:         &secondID,
		Name:       "Person2",
		PersonType: requests.PersonTypeInternalRecruiter,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, secondRequestBody)

	thirdID := uuid.New()
	email = "Person3 Email"
	phone = "333333"
	notes = "Person3 Notes"
	thirdRequestBody := requests.CreatePersonRequest{
		ID:         &thirdID,
		Name:       "Person3",
		PersonType: requests.PersonTypeJobAdvertiser,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, thirdRequestBody)

	// GetAllPersons:
	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Equal(t, len(response), 3)

	assert.Equal(t, *firstRequestBody.ID, response[0].ID)
	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, *thirdRequestBody.ID, response[2].ID)
}

func TestGetAllPersons_ShouldReturnEmptyResponseIfNoPersonsInDatabase(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// GetAllPersons:
	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(response))
}
