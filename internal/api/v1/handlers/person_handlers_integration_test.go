package handlers_test

import (
	"bytes"
	"encoding/json"
	"jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	configPackage "jobsearchtracker/internal/config"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupPersonHandler(t *testing.T) *handlers.PersonHandler {
	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupPersonHandlerTestContainer(t, config)

	var personHandler *handlers.PersonHandler
	err := container.Invoke(func(personHand *handlers.PersonHandler) {
		personHandler = personHand
	})
	assert.NoError(t, err)

	return personHandler
}

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldInsertAndReturnPerson(t *testing.T) {
	personHandler := setupPersonHandler(t)

	id := uuid.New()
	email := "e@ma.il"
	phone := "456908"
	notes := "Notes appeared here"
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "random person name",
		PersonType: requests.PersonTypeHR,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	personHandler.CreatePerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var personResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&personResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, personResponse.ID)
	assert.Equal(t, requestBody.Name, personResponse.Name)
	assert.Equal(t, requestBody.PersonType, personResponse.PersonType)
	assert.Equal(t, requestBody.Email, personResponse.Email)
	assert.Equal(t, requestBody.Phone, personResponse.Phone)
	assert.Equal(t, requestBody.Notes, personResponse.Notes)

	insertedPersonCreatedDate := personResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, insertedPersonCreatedDate)

	assert.Nil(t, personResponse.UpdatedDate)
}

func TestCreatePerson_ShouldReturnStatusConflictIfPersonIDIsDuplicate(t *testing.T) {
	personHandler := setupPersonHandler(t)

	personID := uuid.New()

	firstRequestBody := requests.CreatePersonRequest{
		ID:         &personID,
		Name:       "first Person Name",
		PersonType: requests.PersonTypeOther,
	}

	firstRequestBytes, err := json.Marshal(firstRequestBody)
	assert.NoError(t, err)

	firstRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(firstRequestBytes))
	assert.NoError(t, err)

	firstResponseRecorder := httptest.NewRecorder()

	personHandler.CreatePerson(firstResponseRecorder, firstRequest)
	assert.Equal(t, http.StatusCreated, firstResponseRecorder.Code)

	var firstPersonResponse responses.PersonResponse
	err = json.NewDecoder(firstResponseRecorder.Body).Decode(&firstPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, personID, firstPersonResponse.ID)

	secondRequestBody := requests.CreatePersonRequest{
		ID:         &personID,
		Name:       "second Person Name",
		PersonType: requests.PersonTypeCEO,
	}

	secondRequestBytes, err := json.Marshal(secondRequestBody)
	assert.NoError(t, err)

	secondRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(secondRequestBytes))
	assert.NoError(t, err)

	secondResponseRecorder := httptest.NewRecorder()

	personHandler.CreatePerson(secondResponseRecorder, secondRequest)
	assert.Equal(t, http.StatusConflict, secondResponseRecorder.Code)

	expectedError := "Conflict error on insert: ID already exists\n"
	assert.Equal(t, expectedError, secondResponseRecorder.Body.String())
}

// -------- GetPersonById tests: --------

func TestGetPersonById_ShouldReturnPerson(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// insert a person:

	id := uuid.New()
	email := "Email here"
	phone := "456908"
	notes := "Notes appeared here"
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "random person name",
		PersonType: requests.PersonTypeDeveloper,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	_, createdDateApproximation := insertPerson(t, personHandler, requestBody)

	// get the person:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, response.ID)
	assert.Equal(t, requestBody.Name, response.Name)
	assert.Equal(t, requestBody.PersonType, response.PersonType)
	assert.Equal(t, requestBody.Email, response.Email)
	assert.Equal(t, requestBody.Phone, response.Phone)
	assert.Equal(t, requestBody.Notes, response.Notes)

	insertedPersonCreatedDate := response.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, insertedPersonCreatedDate)

	assert.Nil(t, response.UpdatedDate)
}

func TestGetPersonById_ShouldReturnNotFoundIfPersonDoesNotExist(t *testing.T) {
	personHandler := setupPersonHandler(t)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": uuid.New().String(),
	}
	request = mux.SetURLVars(request, vars)

	personHandler.GetPersonByID(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	firstResponseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, firstResponseBodyString, "Person not found\n")
}

// -------- Test helpers: --------

func insertPerson(
	t *testing.T, personHandler *handlers.PersonHandler, requestBody requests.CreatePersonRequest) (
	*responses.PersonResponse, *string) {

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	createRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	personHandler.CreatePerson(responseRecorder, createRequest)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var createPersonResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&createPersonResponse)
	assert.NoError(t, err)

	return &createPersonResponse, &createdDateApproximation
}
