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
	assert.Nil(t, err, "Failed to marshal create company request")

	request, err := http.NewRequest(http.MethodPost, "/api/v1/companies/new", bytes.NewBuffer(requestBytes))
	assert.Nil(t, err, "Failed to create request")

	responseRecorder := httptest.NewRecorder()

	companyHandler.CreateCompany(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code, "expected response code to be 201")

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString, "CreateCompany returned empty body")

	var companyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyResponse)
	assert.Nil(t, err, "Failed to unmarshal create company response")

	assert.Equal(t, *requestBody.ID, companyResponse.ID, "companyResponse.ID should be the same as requestBody.ID")
	assert.Equal(t, requestBody.Name, companyResponse.Name, "companyResponse.Name should be the same as requestBody.Name")
	assert.Equal(t, requestBody.CompanyType, companyResponse.CompanyType, "companyResponse.requestBodyType should be the same as requestBody.requestBodyType")
	assert.Equal(t, requestBody.Notes, companyResponse.Notes, "companyResponse.Notes should be the same as requestBody.Notes")

	companyResponseLastContact := companyResponse.LastContact.Format(time.RFC3339)
	requestBodyToInsertLastContact := requestBody.LastContact.Format(time.RFC3339)
	assert.Equal(t, requestBodyToInsertLastContact, companyResponseLastContact, "companyResponse.LastContact should be the same as requestBody.LastContact")

	companyResponseCreatedDate := companyResponse.CreatedDate.Format(time.RFC3339)
	now := time.Now().Format(time.RFC3339)
	assert.Equal(t, now, companyResponseCreatedDate, "companyResponse.CreatedDate should be the same as now")

	assert.Nil(t, companyResponse.UpdatedDate, "companyResponse.UpdatedDate should be nil")
}

func TestCreateCompany_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	companyHandler := setupCompanyHandler(t)

	requestBody := requests.CreateCompanyRequest{
		Name:        "random company name",
		CompanyType: requests.CompanyTypeRecruiter,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.Nil(t, err, "Failed to marshal create company request")

	request, err := http.NewRequest(http.MethodPost, "/api/v1/companies/new", bytes.NewBuffer(requestBytes))
	assert.Nil(t, err, "Failed to create request")

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	companyHandler.CreateCompany(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code, "expected response code to be 201")

	var responseBodyString = responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString, "CreateCompany returned empty body")

	var companyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyResponse)
	assert.Nil(t, err, "Failed to unmarshal create company response")

	assert.Equal(t, requestBody.Name, companyResponse.Name, "companyResponse.Name should be the same as requestBody.Name")
	assert.Equal(t, requestBody.CompanyType, companyResponse.CompanyType, "companyResponse.requestBodyType should be the same as requestBody.requestBodyType")

	assert.Nil(t, companyResponse.Notes, "inserted company.Notes should be nil, but got '%s'", companyResponse.Notes)
	assert.Nil(t, companyResponse.LastContact, "inserted company.LastContact should be nil, but got '%s'", companyResponse.LastContact)

	insertedCompanyCreatedDate := companyResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, insertedCompanyCreatedDate, "insertedCompany.CreatedDate should be the same as '%s'", createdDateApproximation)

	assert.Nil(t, companyResponse.UpdatedDate, "companyResponse.UpdatedDate should be nil")
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
	assert.Nil(t, err, "Failed to marshal create first company request")

	firstRequest, err := http.NewRequest(http.MethodPost, "/api/v1/companies/new", bytes.NewBuffer(firstRequestBytes))
	assert.Nil(t, err, "Failed to create second request")

	firstResponseRecorder := httptest.NewRecorder()

	companyHandler.CreateCompany(firstResponseRecorder, firstRequest)
	assert.Equal(t, http.StatusCreated, firstResponseRecorder.Code, "expected response code to be 201")

	var firstCompanyResponse responses.CompanyResponse
	err = json.NewDecoder(firstResponseRecorder.Body).Decode(&firstCompanyResponse)
	assert.Nil(t, err, "Failed to unmarshal first create company response")

	assert.Equal(t, companyID, firstCompanyResponse.ID, "firstCompanyResponse.ID should be the same as companyID")

	secondRequestBody := requests.CreateCompanyRequest{
		ID:          &companyID,
		Name:        "Second Company",
		CompanyType: requests.CompanyTypeEmployer,
	}

	secondRequestBytes, err := json.Marshal(secondRequestBody)
	assert.Nil(t, err, "Failed to marshal create second company request")

	secondRequest, err := http.NewRequest(http.MethodPost, "/api/v1/companies/new", bytes.NewBuffer(secondRequestBytes))
	assert.Nil(t, err, "Failed to create second request")

	secondResponseRecorder := httptest.NewRecorder()

	companyHandler.CreateCompany(secondResponseRecorder, secondRequest)
	assert.Equal(t, http.StatusConflict, secondResponseRecorder.Code, "expected response code to be 400")

	expectedError := "Conflict error on insert: ID already exists\n"
	assert.Equal(t, expectedError, secondResponseRecorder.Body.String(), "secondCompanyResponse returned wrong error in body")
}
