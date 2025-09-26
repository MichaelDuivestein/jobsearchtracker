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

// -------- GetApplicationByJobTitle tests: --------

func TestGetApplicationsByJobTitle_ShouldReturnApplication(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// Insert an application:

	id := uuid.New()
	recruiterID := createCompany(t, companyRepository)
	jobTitle := "Software Engineer"

	requestBody := requests.CreateApplicationRequest{
		ID:               &id,
		RecruiterID:      recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: requests.RemoteStatusTypeOffice,
	}
	insertApplication(t, applicationHandler, requestBody)

	// get the application by full job title:

	firstGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/title", nil)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"title": "Software Engineer",
	}
	firstGetRequest = mux.SetURLVars(firstGetRequest, vars)

	applicationHandler.GetApplicationsByJobTitle(responseRecorder, firstGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var firstResponse []responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&firstResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(firstResponse), 1)

	assert.Equal(t, *requestBody.ID, firstResponse[0].ID)
	assert.Equal(t, requestBody.JobTitle, firstResponse[0].JobTitle)

	// get the application by partial name:

	secondGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/title", nil)
	responseRecorder = httptest.NewRecorder()

	vars = map[string]string{
		"title": "eng",
	}
	secondGetRequest = mux.SetURLVars(secondGetRequest, vars)

	applicationHandler.GetApplicationsByJobTitle(responseRecorder, secondGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString = responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var secondResponse []responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&secondResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(secondResponse), 1)

	assert.Equal(t, *requestBody.ID, secondResponse[0].ID)
	assert.Equal(t, requestBody.JobTitle, secondResponse[0].JobTitle)
}

func TestGetApplicationsByJobTitle_ShouldReturnApplications(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// insert two applications:

	firstID := uuid.New()
	firstCompanyID := createCompany(t, companyRepository)
	firstJobTitle := "GoLang Software Engineer"

	firstRequestBody := requests.CreateApplicationRequest{
		ID:               &firstID,
		CompanyID:        firstCompanyID,
		JobTitle:         &firstJobTitle,
		RemoteStatusType: requests.RemoteStatusTypeHybrid,
	}
	insertApplication(t, applicationHandler, firstRequestBody)

	secondID := uuid.New()
	secondRecruiterID := createCompany(t, companyRepository)
	secondJobTitle := "Backend Developer (golang)"
	secondRequestBody := requests.CreateApplicationRequest{
		ID:               &secondID,
		RecruiterID:      secondRecruiterID,
		JobTitle:         &secondJobTitle,
		RemoteStatusType: requests.RemoteStatusTypeUnknown,
	}
	insertApplication(t, applicationHandler, secondRequestBody)

	// Get applications by name:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/title", nil)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"title": "go",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	applicationHandler.GetApplicationsByJobTitle(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response))

	assert.Equal(t, *firstRequestBody.ID, response[0].ID)
	assert.Equal(t, firstRequestBody.JobTitle, response[0].JobTitle)

	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, secondRequestBody.JobTitle, response[1].JobTitle)

}

func TestGetApplicationsByJobTitle_ShouldReturnNotFoundIfNoApplicationsMatchingJobTitle(t *testing.T) {
	applicationHandler, _ := setupApplicationHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/title", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"title": "Developer",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	applicationHandler.GetApplicationsByJobTitle(responseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(t, "No applications [partially] matching this job title found\n", responseBodyString)
}

// -------- GetAllApplications tests: --------

func TestGetAllApplications_ShouldReturnAllApplications(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// insert applications

	recruiterID := createCompany(t, companyRepository)
	jobTitle := "Software Engineer"

	firstID := uuid.New()
	firstRequestBody := requests.CreateApplicationRequest{
		ID:               &firstID,
		RecruiterID:      recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: requests.RemoteStatusTypeOffice,
	}
	insertApplication(t, applicationHandler, firstRequestBody)

	secondID := uuid.New()
	secondRequestBody := requests.CreateApplicationRequest{
		ID:               &secondID,
		RecruiterID:      recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: requests.RemoteStatusTypeRemote,
	}
	insertApplication(t, applicationHandler, secondRequestBody)

	thirdID := uuid.New()
	thirdRequestBody := requests.CreateApplicationRequest{
		ID:               &thirdID,
		RecruiterID:      recruiterID,
		JobTitle:         &jobTitle,
		RemoteStatusType: requests.RemoteStatusTypeHybrid,
	}
	insertApplication(t, applicationHandler, thirdRequestBody)

	// GetAllApplications:
	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.GetAllApplications(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Equal(t, len(response), 3)

	assert.Equal(t, *firstRequestBody.ID, response[0].ID)
	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, *thirdRequestBody.ID, response[2].ID)
}

func TestGetAllApplications_ShouldReturnEmptyResponseIfNoApplicationsInDatabase(t *testing.T) {
	applicationHandler, _ := setupApplicationHandler(t)

	// GetAllApplications:
	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.GetAllApplications(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(response))
}

// -------- UpdateApplication tests: --------

func TestUpdateApplication_ShouldUpdateApplication(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// create an application

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
	applicationDate := time.Now().AddDate(0, 0, 6)
	createRequest := requests.CreateApplicationRequest{
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
	_, createdDateApproximation := insertApplication(t, applicationHandler, createRequest)

	// update the application

	newCompanyID := createCompany(t, companyRepository)
	newRecruiterID := createCompany(t, companyRepository)

	newJobTitle := "New Job Title"
	newJobAdURL := "New Job Ad URL"
	newCountry := "New Country"
	newArea := "New Area"
	var newRemoteStatusType requests.RemoteStatusType = requests.RemoteStatusTypeOffice
	newWeekdaysInOffice := 1
	newEstimatedCycleTime := 2
	newEstimatedCommuteTime := 3
	newApplicationDate := time.Now().AddDate(0, 0, 40)

	updateBody := requests.UpdateApplicationRequest{
		ID:                   id,
		CompanyID:            newCompanyID,
		RecruiterID:          newRecruiterID,
		JobTitle:             &newJobTitle,
		JobAdURL:             &newJobAdURL,
		Country:              &newCountry,
		Area:                 &newArea,
		RemoteStatusType:     &newRemoteStatusType,
		WeekdaysInOffice:     &newWeekdaysInOffice,
		EstimatedCycleTime:   &newEstimatedCycleTime,
		EstimatedCommuteTime: &newEstimatedCommuteTime,
		ApplicationDate:      &newApplicationDate,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/application/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	applicationHandler.UpdateApplication(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the application by ID

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/id", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": id.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	applicationHandler.GetApplicationByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var getApplicationResponse responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&getApplicationResponse)
	assert.NoError(t, err)

	assert.Equal(t, id, getApplicationResponse.ID)
	assert.Equal(t, newCompanyID.String(), getApplicationResponse.CompanyID.String())
	assert.Equal(t, newRecruiterID.String(), getApplicationResponse.RecruiterID.String())
	assert.Equal(t, newJobTitle, *getApplicationResponse.JobTitle)
	assert.Equal(t, newJobAdURL, *getApplicationResponse.JobAdURL)
	assert.Equal(t, newCountry, *getApplicationResponse.Country)
	assert.Equal(t, newArea, *getApplicationResponse.Area)
	assert.Equal(t, newRemoteStatusType, *getApplicationResponse.RemoteStatusType)
	assert.Equal(t, newWeekdaysInOffice, *getApplicationResponse.WeekdaysInOffice)
	assert.Equal(t, newEstimatedCycleTime, *getApplicationResponse.EstimatedCycleTime)
	assert.Equal(t, newEstimatedCommuteTime, *getApplicationResponse.EstimatedCommuteTime)

	retrievedApplicationDate := getApplicationResponse.ApplicationDate.Format(time.RFC3339)
	expectedApplicationDate := newApplicationDate.Format(time.RFC3339)
	assert.Equal(t, expectedApplicationDate, retrievedApplicationDate)

	applicationResponseCreatedDate := getApplicationResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, applicationResponseCreatedDate)

	applicationResponseUpdatedDate := getApplicationResponse.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, applicationResponseUpdatedDate)
}

func TestUpdateApplication_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// create an application
	id := uuid.New()
	recruiterID := createCompany(t, companyRepository)
	jobAdURL := "Job Ad URL"

	createRequest := requests.CreateApplicationRequest{
		ID:               &id,
		RecruiterID:      recruiterID,
		JobAdURL:         &jobAdURL,
		RemoteStatusType: requests.RemoteStatusTypeOffice,
	}
	insertApplication(t, applicationHandler, createRequest)

	// update the application

	updateBody := requests.UpdateApplicationRequest{
		ID: id,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/application/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationHandler.UpdateApplication(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(
		t,
		"Unable to convert request to internal model: validation error: nothing to update\n",
		responseBodyString)
}

// -------- DeleteApplication tests: --------

func TestDeleteApplication_ShouldDeleteApplication(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// insert an application

	id := uuid.New()
	companyID := createCompany(t, companyRepository)
	jobTitle := "JobTitle"
	requestBody := requests.CreateApplicationRequest{
		ID:               &id,
		CompanyID:        companyID,
		JobTitle:         &jobTitle,
		RemoteStatusType: requests.RemoteStatusTypeHybrid,
	}

	insertApplication(t, applicationHandler, requestBody)

	// delete the application

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/application/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	applicationHandler.DeleteApplication(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusOK, deleteResponseRecorder.Code)

	// try to get the application
	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/id", nil)
	assert.NoError(t, err)

	getResponseRecorder := httptest.NewRecorder()
	getRequest = mux.SetURLVars(getRequest, vars)

	applicationHandler.GetApplicationByID(getResponseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, getResponseRecorder.Code, "GetApplicationByID returned wrong status code")
}

func TestDeleteApplication_ShouldReturnStatusNotFoundIfApplicationDoesNotExist(t *testing.T) {
	applicationHandler, _ := setupApplicationHandler(t)

	id := uuid.New()

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/application/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	applicationHandler.DeleteApplication(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusNotFound, deleteResponseRecorder.Code)
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
