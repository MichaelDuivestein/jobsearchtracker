package handlers_test

import (
	"bytes"
	"encoding/json"
	"jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	configPackage "jobsearchtracker/internal/config"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupApplicationHandler(t *testing.T) (*handlers.ApplicationHandler, *repositories.CompanyRepository) {
	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupApplicationHandlerTestContainer(t, config)

	var applicationHandler *handlers.ApplicationHandler
	err := container.Invoke(func(applicationHand *handlers.ApplicationHandler) {
		applicationHandler = applicationHand
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	return applicationHandler, companyRepository
}

// -------- CreateApplication tests: --------

func TestCreateApplication_ShouldInsertAndReturnApplication(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	companyID := createCompany(t, companyRepository)
	recruiterID := createCompany(t, companyRepository)

	id := uuid.New()
	jobTitle := "Job Title"
	jobAdURL := "Job Ad URL"
	country := "Some Country"
	area := "Some Area"
	weekdaysInOffice := 9
	estimatedCycleTime := 8
	estimatedCommuteTime := 7
	applicationDate := time.Now().AddDate(0, 0, -9)

	requestBody := requests.CreateApplicationRequest{
		ID:                   &id,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     requests.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	applicationHandler.CreateApplication(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var applicationResponse responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&applicationResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, applicationResponse.ID)
	assert.Equal(t, *requestBody.CompanyID, *applicationResponse.CompanyID)
	assert.Equal(t, *requestBody.RecruiterID, *applicationResponse.RecruiterID)
	assert.Equal(t, *requestBody.JobTitle, *applicationResponse.JobTitle)
	assert.Equal(t, *requestBody.JobAdURL, *applicationResponse.JobAdURL)
	assert.Equal(t, *requestBody.Country, *applicationResponse.Country)
	assert.Equal(t, *requestBody.Area, *applicationResponse.Area)
	assert.Equal(t, requests.RemoteStatusTypeHybrid, applicationResponse.RemoteStatusType.String())
	assert.Equal(t, *requestBody.WeekdaysInOffice, *applicationResponse.WeekdaysInOffice)
	assert.Equal(t, *requestBody.EstimatedCycleTime, *applicationResponse.EstimatedCycleTime)
	assert.Equal(t, *requestBody.EstimatedCommuteTime, *applicationResponse.EstimatedCommuteTime)

	applicationToInsertApplicationDate := applicationDate.Format(time.RFC3339)
	applicationResponseApplicationDate := applicationResponse.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertApplicationDate, applicationResponseApplicationDate)

	applicationResponseCreatedDate := applicationResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, applicationResponseCreatedDate)

	assert.Nil(t, applicationResponse.UpdatedDate)
}

func TestCreateApplication_ShouldReturnStatusConflictIfApplicationIDIsDuplicate(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	applicationID := uuid.New()
	recruiterID := createCompany(t, companyRepository)
	firstJobTitle := "First Job Title"

	firstRequestBody := requests.CreateApplicationRequest{
		ID:               &applicationID,
		RecruiterID:      recruiterID,
		JobTitle:         &firstJobTitle,
		RemoteStatusType: requests.RemoteStatusTypeHybrid,
	}

	firstRequestBytes, err := json.Marshal(firstRequestBody)
	assert.NoError(t, err)

	firstRequest, err :=
		http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(firstRequestBytes))
	assert.NoError(t, err)

	firstResponseRecorder := httptest.NewRecorder()

	applicationHandler.CreateApplication(firstResponseRecorder, firstRequest)
	assert.Equal(t, http.StatusCreated, firstResponseRecorder.Code)

	var firstApplicationResponse responses.ApplicationResponse
	err = json.NewDecoder(firstResponseRecorder.Body).Decode(&firstApplicationResponse)
	assert.NoError(t, err)

	assert.Equal(t, applicationID, firstApplicationResponse.ID)

	secondJobTitle := "Second Job Title"
	secondRequestBody := requests.CreateApplicationRequest{
		ID:               &applicationID,
		RecruiterID:      recruiterID,
		JobTitle:         &secondJobTitle,
		RemoteStatusType: models.RemoteStatusTypeRemote,
	}

	secondRequestBytes, err := json.Marshal(secondRequestBody)
	assert.NoError(t, err)

	secondRequest, err :=
		http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(secondRequestBytes))
	assert.NoError(t, err)

	secondResponseRecorder := httptest.NewRecorder()

	applicationHandler.CreateApplication(secondResponseRecorder, secondRequest)
	assert.Equal(t, http.StatusConflict, secondResponseRecorder.Code)

	expectedError := "Conflict error on insert: ID already exists\n"
	assert.Equal(t, expectedError, secondResponseRecorder.Body.String())
}

func TestCreateApplication_ShouldReturnErrorIfCompanyIDDoesNotExistInCompany(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	applicationID := uuid.New()
	companyID := uuid.New()
	recruiterID := createCompany(t, companyRepository)
	jobTitle := "Job Title"

	requestBody := requests.CreateApplicationRequest{
		ID:          &applicationID,
		CompanyID:   &companyID,
		RecruiterID: recruiterID,
		JobTitle:    &jobTitle,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.CreateApplication(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"validation error on field 'RemoteStatusType': RemoteStatusType is invalid\n",
		responseBodyString)
}

func TestCreateApplication_ShouldReturnErrorIfRecruiterIDDoesNotExistInCompany(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	applicationID := uuid.New()
	companyID := createCompany(t, companyRepository)
	recruiterID := uuid.New()
	jobTitle := "Job Title"

	requestBody := requests.CreateApplicationRequest{
		ID:          &applicationID,
		CompanyID:   companyID,
		RecruiterID: &recruiterID,
		JobTitle:    &jobTitle,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.CreateApplication(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"validation error on field 'RemoteStatusType': RemoteStatusType is invalid\n",
		responseBodyString)
}

// -------- GetApplicationById tests: --------

func TestGetApplicationById_ShouldReturnApplication(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// Insert an application:

	id := uuid.New()
	companyID := createCompany(t, companyRepository)
	recruiterID := createCompany(t, companyRepository)
	jobTitle := "Job Title"
	jobAdURL := "job Ad URL"
	country := "country"
	area := "area"
	weekdaysInOffice := 6
	estimatedCycleTime := 7
	estimatedCommuteTime := 8
	applicationDate := time.Now().AddDate(0, 0, -20)
	requestBody := requests.CreateApplicationRequest{
		ID:                   &id,
		CompanyID:            companyID,
		RecruiterID:          recruiterID,
		JobTitle:             &jobTitle,
		JobAdURL:             &jobAdURL,
		Country:              &country,
		Area:                 &area,
		RemoteStatusType:     requests.RemoteStatusTypeOffice,
		WeekdaysInOffice:     &weekdaysInOffice,
		EstimatedCycleTime:   &estimatedCycleTime,
		EstimatedCommuteTime: &estimatedCommuteTime,
		ApplicationDate:      &applicationDate,
	}
	_, createdDateApproximation := insertApplication(t, applicationHandler, requestBody)

	// Get the application:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	applicationHandler.GetApplicationByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, response.ID)
	assert.Equal(t, *requestBody.CompanyID, *response.CompanyID)
	assert.Equal(t, *requestBody.RecruiterID, *response.RecruiterID)
	assert.Equal(t, *requestBody.JobTitle, *response.JobTitle)
	assert.Equal(t, *requestBody.JobAdURL, *response.JobAdURL)
	assert.Equal(t, *requestBody.Country, *response.Country)
	assert.Equal(t, *requestBody.Area, *response.Area)
	assert.Equal(t, requests.RemoteStatusTypeOffice, response.RemoteStatusType.String())
	assert.Equal(t, *requestBody.WeekdaysInOffice, *response.WeekdaysInOffice)
	assert.Equal(t, *requestBody.EstimatedCycleTime, *response.EstimatedCycleTime)
	assert.Equal(t, *requestBody.EstimatedCommuteTime, *response.EstimatedCommuteTime)

	applicationToInsertApplicationDate := applicationDate.Format(time.RFC3339)
	responseApplicationDate := response.ApplicationDate.Format(time.RFC3339)
	assert.Equal(t, applicationToInsertApplicationDate, responseApplicationDate)

	responseCreatedDate := response.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, responseCreatedDate)

	assert.Nil(t, response.UpdatedDate)
}

func TestGetApplicationById_ShouldReturnNotFoundIfApplicationDoesNotExist(t *testing.T) {
	applicationHandler, _ := setupApplicationHandler(t)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": uuid.New().String(),
	}
	request = mux.SetURLVars(request, vars)

	applicationHandler.GetApplicationByID(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	firstResponseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, firstResponseBodyString, "Application not found\n")
}

// -------- Test helpers: --------

func createCompany(t *testing.T, companyRepository *repositories.CompanyRepository) *uuid.UUID {

	id := uuid.New()
	company := models.CreateCompany{
		ID:          &id,
		Name:        "Example Company Name",
		CompanyType: models.CompanyTypeEmployer,
	}

	insertedCompany, err := companyRepository.Create(&company)
	assert.NoError(t, err)

	return &insertedCompany.ID
}

func insertApplication(
	t *testing.T, applicationHandler *handlers.ApplicationHandler, requestBody requests.CreateApplicationRequest) (
	*responses.ApplicationResponse, *string) {

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	createRequest, err := http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	applicationHandler.CreateApplication(responseRecorder, createRequest)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var createApplicationResponse responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&createApplicationResponse)
	assert.NoError(t, err)

	return &createApplicationResponse, &createdDateApproximation
}
