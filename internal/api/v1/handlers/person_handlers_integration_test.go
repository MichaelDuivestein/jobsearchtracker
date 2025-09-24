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

// -------- GetPersonByName tests: --------

func TestGetPersonsByName_ShouldReturnPerson(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// Insert a person:

	id := uuid.New()
	email := "Email here"
	phone := "456908"
	notes := "Notes appeared here"
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "PersonName",
		PersonType: requests.PersonTypeDeveloper,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, requestBody)

	// Get the person by full name:

	firstGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "PersonName",
	}
	firstGetRequest = mux.SetURLVars(firstGetRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, firstGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var firstResponse []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&firstResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(firstResponse), 1)

	assert.Equal(t, *requestBody.ID, firstResponse[0].ID)
	assert.Equal(t, requestBody.Name, firstResponse[0].Name)

	// Get the person by partial name:

	secondGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	responseRecorder = httptest.NewRecorder()

	vars = map[string]string{
		"name": "son",
	}
	secondGetRequest = mux.SetURLVars(secondGetRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, secondGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString = responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var secondResponse []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&secondResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(secondResponse), 1)

	assert.Equal(t, *requestBody.ID, secondResponse[0].ID)
	assert.Equal(t, requestBody.Name, secondResponse[0].Name)
}

func TestGetPersonsByName_ShouldReturnPersons(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// Insert two persons:

	firstID := uuid.New()
	email := "Jane email"
	phone := "345345"
	notes := "Jane Notes"
	firstRequestBody := requests.CreatePersonRequest{
		ID:         &firstID,
		Name:       "Jane Smith",
		PersonType: requests.PersonTypeJobAdvertiser,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, firstRequestBody)

	secondID := uuid.New()
	email = "Sarah email"
	phone = "4567856654"
	notes = "Sara Notes"
	secondRequestBody := requests.CreatePersonRequest{
		ID:         &secondID,
		Name:       "Sarah Janesson",
		PersonType: requests.PersonTypeCEO,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, secondRequestBody)

	// Get persons by name:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "Jane",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, len(response), 2)

	assert.Equal(t, *firstRequestBody.ID, response[0].ID)
	assert.Equal(t, firstRequestBody.Name, response[0].Name)

	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, secondRequestBody.Name, response[1].Name)

}

func TestGetPersonsByName_ShouldReturnNotFoundIfNoPersonsMatchingName(t *testing.T) {
	personHandler := setupPersonHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "Steve",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(t, "No people [partially] matching this name found\n", responseBodyString)
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

// -------- UpdatePerson tests: --------

func TestUpdatePerson_ShouldUpdatePerson(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// create a person
	id := uuid.New()
	email := "Person Email"
	phone := "2345345"
	notes := "Notes"
	createRequest := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "Person Name",
		PersonType: requests.PersonTypeUnknown,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	_, createdDateApproximation := insertPerson(t, personHandler, createRequest)

	// update the person

	updatedName := "updated person name"
	var updatedPersonType requests.PersonType = requests.PersonTypeJobContact
	updatedEmail := "updated email"
	updatedPhone := "46584566745"
	updatedNotes := "updated notes"

	updateBody := requests.UpdatePersonRequest{
		ID:         id,
		Name:       &updatedName,
		PersonType: &updatedPersonType,
		Email:      &updatedEmail,
		Phone:      &updatedPhone,
		Notes:      &updatedNotes,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	personHandler.UpdatePerson(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the person by ID

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": id.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var getPersonResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&getPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, id, getPersonResponse.ID)
	assert.Equal(t, updatedName, getPersonResponse.Name)
	assert.Equal(t, updatedPersonType, getPersonResponse.PersonType)
	assert.Equal(t, updatedEmail, *getPersonResponse.Email)
	assert.Equal(t, updatedPhone, *getPersonResponse.Phone)
	assert.Equal(t, updatedNotes, *getPersonResponse.Notes)

	personResponseCreatedDate := getPersonResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, personResponseCreatedDate)

	personResponseUpdatedDate := getPersonResponse.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, personResponseUpdatedDate)
}

func TestUpdatePerson_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// create a person
	id := uuid.New()
	email := "Person Email"
	phone := "2345345"
	notes := "Notes"
	createRequest := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "Person Name",
		PersonType: requests.PersonTypeUnknown,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, createRequest)

	// update the person
	updateBody := requests.UpdatePersonRequest{
		ID: id,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.UpdatePerson(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(
		t,
		"Unable to convert request to internal model: validation error: nothing to update\n",
		responseBodyString)
}

// -------- DeletePerson tests: --------

func TestDeletePerson_ShouldDeletePerson(t *testing.T) {
	personHandler := setupPersonHandler(t)

	// insert a person

	id := uuid.New()
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "Person Name",
		PersonType: requests.PersonTypeDeveloper,
	}

	insertPerson(t, personHandler, requestBody)

	// delete the person

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	personHandler.DeletePerson(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusOK, deleteResponseRecorder.Code)

	// try to get the person
	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	getResponseRecorder := httptest.NewRecorder()
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonByID(getResponseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, getResponseRecorder.Code, "GetPersonByID returned wrong status code")
}

func TestDeletePerson_ShouldReturnStatusNotFoundIfPersonDoesNotExist(t *testing.T) {
	personHandler := setupPersonHandler(t)

	id := uuid.New()

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	personHandler.DeletePerson(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusNotFound, deleteResponseRecorder.Code)
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
