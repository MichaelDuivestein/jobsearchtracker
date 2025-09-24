package handlers_test

import (
	"bytes"
	"encoding/json"
	"jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	configPackage "jobsearchtracker/internal/config"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupCompanyHandler(t *testing.T) *handlers.CompanyHandler {
	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyHandlerTestContainer(t, config)

	var companyHandler *handlers.CompanyHandler
	err := container.Invoke(func(companyHand *handlers.CompanyHandler) {
		companyHandler = companyHand
	})
	assert.NoError(t, err)

	return companyHandler
}

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldReturnCompany(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	id := uuid.New()
	notes := "Not a lot of notes for this company"
	lastContact := time.Now().AddDate(0, 0, -1)

	requestBody := requests.CreateCompanyRequest{
		ID:          &id,
		Name:        "random company name",
		CompanyType: requests.CompanyTypeConsultancy,
		Notes:       &notes,
		LastContact: &lastContact,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.Nil(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyHandler.CreateCompany(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var companyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, companyResponse.ID)
	assert.Equal(t, requestBody.Name, companyResponse.Name)
	assert.Equal(t, requestBody.CompanyType, companyResponse.CompanyType)
	assert.Equal(t, requestBody.Notes, companyResponse.Notes)

	companyResponseLastContact := companyResponse.LastContact.Format(time.RFC3339)
	requestBodyToInsertLastContact := requestBody.LastContact.Format(time.RFC3339)
	assert.Equal(t, requestBodyToInsertLastContact, companyResponseLastContact)

	companyResponseCreatedDate := companyResponse.CreatedDate.Format(time.RFC3339)
	now := time.Now().Format(time.RFC3339)
	assert.Equal(t, now, companyResponseCreatedDate)

	assert.Nil(t, companyResponse.UpdatedDate)
}

func TestCreateCompany_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	requestBody := requests.CreateCompanyRequest{
		Name:        "random company name",
		CompanyType: requests.CompanyTypeRecruiter,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	companyHandler.CreateCompany(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	var responseBodyString = responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var companyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyResponse)
	assert.NoError(t, err)

	assert.Equal(t, requestBody.Name, companyResponse.Name)
	assert.Equal(t, requestBody.CompanyType, companyResponse.CompanyType)

	assert.Nil(t, companyResponse.Notes)
	assert.Nil(t, companyResponse.LastContact)

	insertedCompanyCreatedDate := companyResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, insertedCompanyCreatedDate)

	assert.Nil(t, companyResponse.UpdatedDate)
}

func TestCreateCompany_ShouldReturnStatusConflict_IfCompanyIDIsDuplicate(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	companyID := uuid.New()

	firstRequestBody := requests.CreateCompanyRequest{
		ID:          &companyID,
		Name:        "First Company",
		CompanyType: requests.CompanyTypeRecruiter,
	}

	firstRequestBytes, err := json.Marshal(firstRequestBody)
	assert.NoError(t, err)

	firstRequest, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(firstRequestBytes))
	assert.NoError(t, err)

	firstResponseRecorder := httptest.NewRecorder()

	companyHandler.CreateCompany(firstResponseRecorder, firstRequest)
	assert.Equal(t, http.StatusCreated, firstResponseRecorder.Code)

	var firstCompanyResponse responses.CompanyResponse
	err = json.NewDecoder(firstResponseRecorder.Body).Decode(&firstCompanyResponse)
	assert.NoError(t, err)

	assert.Equal(t, companyID, firstCompanyResponse.ID)

	secondRequestBody := requests.CreateCompanyRequest{
		ID:          &companyID,
		Name:        "Second Company",
		CompanyType: requests.CompanyTypeEmployer,
	}

	secondRequestBytes, err := json.Marshal(secondRequestBody)
	assert.NoError(t, err)

	secondRequest, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(secondRequestBytes))
	assert.NoError(t, err)

	secondResponseRecorder := httptest.NewRecorder()

	companyHandler.CreateCompany(secondResponseRecorder, secondRequest)
	assert.Equal(t, http.StatusConflict, secondResponseRecorder.Code)

	assert.Equal(t, "Conflict error on insert: ID already exists\n", secondResponseRecorder.Body.String())
}

// -------- GetCompanyById tests: --------

func TestGetCompanyById_ShouldReturnCompany(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// Insert the company:

	id := uuid.New()
	notes := "Not a lot of notes for this company"
	lastContact := time.Now().AddDate(0, 0, -1)
	requestBody := requests.CreateCompanyRequest{
		ID:          &id,
		Name:        "random company name",
		CompanyType: requests.CompanyTypeConsultancy,
		Notes:       &notes,
		LastContact: &lastContact,
	}

	_, createdDateApproximation := insertCompany(t, companyHandler, requestBody)

	// Get the company:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	companyHandler.GetCompanyById(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var getCompanyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&getCompanyResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, getCompanyResponse.ID)
	assert.Equal(t, requestBody.Name, getCompanyResponse.Name)
	assert.Equal(t, requestBody.CompanyType, getCompanyResponse.CompanyType)
	assert.Equal(t, requestBody.Notes, getCompanyResponse.Notes)

	companyResponseLastContact := getCompanyResponse.LastContact.Format(time.RFC3339)
	requestBodyToInsertLastContact := requestBody.LastContact.Format(time.RFC3339)
	assert.Equal(t, requestBodyToInsertLastContact, companyResponseLastContact)

	companyResponseCreatedDate := getCompanyResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, companyResponseCreatedDate)

	assert.Nil(t, getCompanyResponse.UpdatedDate)
}

func TestGetCompanyById_ShouldReturnNotFoundIfCompanyDoesNotExist(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// Get a company that doesn't exist

	firstGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	firstGetVars := map[string]string{
		"id": uuid.New().String(),
	}
	firstGetRequest = mux.SetURLVars(firstGetRequest, firstGetVars)

	companyHandler.GetCompanyById(responseRecorder, firstGetRequest)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	firstResponseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, firstResponseBodyString)

	// Insert a company

	id := uuid.New()
	notes := "Not a lot of notes for this company"
	lastContact := time.Now().AddDate(0, 0, -1)
	requestBody := requests.CreateCompanyRequest{
		ID:          &id,
		Name:        "random company name",
		CompanyType: requests.CompanyTypeConsultancy,
		Notes:       &notes,
		LastContact: &lastContact,
	}
	insertCompany(t, companyHandler, requestBody)

	// Get another company that doesn't exist

	secondGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()

	secondGetVars := map[string]string{
		"id": uuid.New().String(),
	}
	secondGetRequest = mux.SetURLVars(secondGetRequest, secondGetVars)

	companyHandler.GetCompanyById(responseRecorder, secondGetRequest)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	secondResponseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, secondResponseBodyString)
}

// -------- GetCompanyByName tests: --------

func TestGetCompaniesByName_ShouldReturnCompany(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// Insert a company:

	id := uuid.New()
	notes := "Notes appeared here"
	lastContact := time.Now().AddDate(0, 1, 0)
	requestBody := requests.CreateCompanyRequest{
		ID:          &id,
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       &notes,
		LastContact: &lastContact,
	}
	insertCompany(t, companyHandler, requestBody)

	// Get the company by full name:

	firstGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "CompanyName",
	}
	firstGetRequest = mux.SetURLVars(firstGetRequest, vars)

	companyHandler.GetCompaniesByName(responseRecorder, firstGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var firstResponse []responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&firstResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(firstResponse), 1)

	assert.Equal(t, *requestBody.ID, firstResponse[0].ID)
	assert.Equal(t, requestBody.Name, firstResponse[0].Name)

	// Get the company by partial name:

	secondGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	responseRecorder = httptest.NewRecorder()

	vars = map[string]string{
		"name": "pany",
	}
	secondGetRequest = mux.SetURLVars(secondGetRequest, vars)

	companyHandler.GetCompaniesByName(responseRecorder, secondGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString = responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var secondResponse []responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&secondResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(secondResponse), 1)

	assert.Equal(t, *requestBody.ID, secondResponse[0].ID)
	assert.Equal(t, requestBody.Name, secondResponse[0].Name)
}

func TestGetCompaniesByName_ShouldReturnCompanies(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// Insert two companies:

	firstID := uuid.New()
	firstNotes := "Noteworthy stuff"
	firstLastContact := time.Now().AddDate(0, 1, 0)

	firstRequestBody := requests.CreateCompanyRequest{
		ID:          &firstID,
		Name:        "Duck Watchers",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       &firstNotes,
		LastContact: &firstLastContact,
	}
	insertCompany(t, companyHandler, firstRequestBody)

	secondID := uuid.New()
	secondNotes := "More Noteworthy stuff"
	secondLastContact := time.Now().AddDate(0, 1, 0)
	secondRequestBody := requests.CreateCompanyRequest{
		ID:          &secondID,
		Name:        "Duck farm",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       &secondNotes,
		LastContact: &secondLastContact,
	}
	insertCompany(t, companyHandler, secondRequestBody)

	// Get companies by name:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "duck",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	companyHandler.GetCompaniesByName(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, len(response), 2)

	assert.Equal(t, *firstRequestBody.ID, response[0].ID)
	assert.Equal(t, firstRequestBody.Name, response[0].Name)

	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, secondRequestBody.Name, response[1].Name)

}

func TestGetCompaniesByName_ShouldReturnNotFoundIfNoCompaniesMatchingName(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "Florist",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	companyHandler.GetCompaniesByName(responseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(t, "No people [partially] matching this name found\n", responseBodyString)
}

// -------- GetAllCompanies tests: --------

func TestGetAllCompanies_ShouldReturnAllCompanies(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// create 2 companies

	company1Id := uuid.New()
	company1Notes := "First Company Notes"
	company1LastContact := time.Now().AddDate(-1, 0, 0)
	request1Body := requests.CreateCompanyRequest{
		ID:          &company1Id,
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       &company1Notes,
		LastContact: &company1LastContact,
	}
	insertCompany(t, companyHandler, request1Body)

	company2Id := uuid.New()
	company2Notes := "Second Company notes"
	company2LastContact := time.Now().AddDate(-1, 0, 0)
	request2Body := requests.CreateCompanyRequest{
		ID:          &company2Id,
		Name:        "company2Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       &company2Notes,
		LastContact: &company2LastContact,
	}
	insertCompany(t, companyHandler, request2Body)

	// GetAllCompanies:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyHandler.GetAllCompanies(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response))

	assert.Equal(t, company2Id, response[0].ID)
	assert.Equal(t, company1Id, response[1].ID)
}

func TestGetAllCompanies_ShouldReturnEmptyResponseIfNoCompaniesInDatabase(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyHandler.GetAllCompanies(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response))
}

// -------- Update tests: --------

func TestUpdateCompany_ShouldUpdateCompany(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// create a company

	id := uuid.New()
	notes := "Notes here"
	lastContact := time.Now().AddDate(0, 2, 0)
	createRequest := requests.CreateCompanyRequest{
		ID:          &id,
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       &notes,
		LastContact: &lastContact,
	}
	_, createdDateApproximation := insertCompany(t, companyHandler, createRequest)

	// update the company

	updatedName := "Updated Name"
	var updatedCompanyType requests.CompanyType = models.CompanyTypeConsultancy
	updatedNotes := "Updated Notes"
	updatedLastContact := time.Now().AddDate(0, 0, -4)
	updateBody := requests.UpdateCompanyRequest{
		ID:          id,
		Name:        &updatedName,
		CompanyType: &updatedCompanyType,
		Notes:       &updatedNotes,
		LastContact: &updatedLastContact,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/company/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	companyHandler.UpdateCompany(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the company by ID

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": id.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	companyHandler.GetCompanyById(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var getCompanyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&getCompanyResponse)
	assert.NoError(t, err)

	assert.Equal(t, id, getCompanyResponse.ID)
	assert.Equal(t, updatedName, getCompanyResponse.Name)
	assert.Equal(t, updatedCompanyType, getCompanyResponse.CompanyType)
	assert.Equal(t, updatedNotes, *getCompanyResponse.Notes)

	companyResponseLastContact := getCompanyResponse.LastContact.Format(time.RFC3339)
	updatedLastContactString := updatedLastContact.Format(time.RFC3339)
	assert.Equal(t, updatedLastContactString, companyResponseLastContact)

	companyResponseCreatedDate := getCompanyResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, companyResponseCreatedDate)

	companyResponseUpdatedDate := getCompanyResponse.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, companyResponseUpdatedDate)
}

func TestUpdateCompany_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// create a company

	id := uuid.New()
	notes := "Notes"
	lastContact := time.Now().AddDate(0, 0, -1)
	createRequest := requests.CreateCompanyRequest{
		ID:          &id,
		Name:        "Nameless Company",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       &notes,
		LastContact: &lastContact,
	}
	insertCompany(t, companyHandler, createRequest)

	// update the company

	updateBody := requests.UpdateCompanyRequest{
		ID: id,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/company/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyHandler.UpdateCompany(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(
		t,
		"Unable to convert request to internal model: validation error: nothing to update\n",
		responseBodyString)
}

// -------- DeleteCompany tests: --------

func TestDeleteCompany_ShouldDeleteCompany(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	// insert the company

	id := uuid.New()
	notes := "Noting things"
	lastContact := time.Now().AddDate(0, 0, 0)
	requestBody := requests.CreateCompanyRequest{
		ID:          &id,
		Name:        "Keeping company",
		CompanyType: requests.CompanyTypeConsultancy,
		Notes:       &notes,
		LastContact: &lastContact,
	}
	insertCompany(t, companyHandler, requestBody)

	// delete the company

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/company/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	companyHandler.DeleteCompany(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusOK, deleteResponseRecorder.Code)

	// try to get the company

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	getResponseRecorder := httptest.NewRecorder()
	getRequest = mux.SetURLVars(getRequest, vars)

	companyHandler.GetCompanyById(getResponseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, getResponseRecorder.Code)
}

func TestDeleteCompany_ShouldReturnStatusNotFoundIfCompanyDoesNotExist(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	id := uuid.New()

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/company/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	companyHandler.DeleteCompany(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusNotFound, deleteResponseRecorder.Code)
}

// -------- Test helpers: --------

func insertCompany(
	t *testing.T, companyHandler *handlers.CompanyHandler, requestBody requests.CreateCompanyRequest) (
	*responses.CompanyResponse, *string) {
	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	createRequest, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err, "Failed to create request")

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	companyHandler.CreateCompany(responseRecorder, createRequest)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var createCompanyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&createCompanyResponse)
	assert.NoError(t, err)

	return &createCompanyResponse, &createdDateApproximation
}
