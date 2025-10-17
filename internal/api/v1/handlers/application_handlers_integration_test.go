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
	"jobsearchtracker/internal/testutil"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
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

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	requestBody := requests.CreateApplicationRequest{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            testutil.ToPtr(company.ID),
		RecruiterID:          testutil.ToPtr(recruiter.ID),
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     requests.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(9),
		EstimatedCycleTime:   testutil.ToPtr(8),
		EstimatedCommuteTime: testutil.ToPtr(7),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -9)),
	}
	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	applicationHandler.CreateApplication(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var applicationResponse responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&applicationResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, applicationResponse.ID)
	assert.Equal(t, requestBody.CompanyID, applicationResponse.CompanyID)
	assert.Equal(t, requestBody.RecruiterID, applicationResponse.RecruiterID)
	assert.Equal(t, requestBody.JobTitle, applicationResponse.JobTitle)
	assert.Equal(t, requestBody.JobAdURL, applicationResponse.JobAdURL)
	assert.Equal(t, requestBody.Country, applicationResponse.Country)
	assert.Equal(t, requestBody.Area, applicationResponse.Area)
	assert.Equal(t, requestBody.RemoteStatusType.String(), applicationResponse.RemoteStatusType.String())
	assert.Equal(t, requestBody.WeekdaysInOffice, applicationResponse.WeekdaysInOffice)
	assert.Equal(t, requestBody.EstimatedCycleTime, applicationResponse.EstimatedCycleTime)
	assert.Equal(t, requestBody.EstimatedCommuteTime, applicationResponse.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, requestBody.ApplicationDate, applicationResponse.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, &createdDateApproximation, applicationResponse.CreatedDate)
	assert.Nil(t, applicationResponse.UpdatedDate)
}

func TestCreateApplication_ShouldReturnStatusConflictIfApplicationIDIsDuplicate(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	applicationID := uuid.New()
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	firstRequestBody := requests.CreateApplicationRequest{
		ID:               &applicationID,
		RecruiterID:      testutil.ToPtr(recruiter.ID),
		JobTitle:         testutil.ToPtr("First Job Title"),
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

	secondRequestBody := requests.CreateApplicationRequest{
		ID:               &applicationID,
		RecruiterID:      testutil.ToPtr(recruiter.ID),
		JobTitle:         testutil.ToPtr("Second Job Title"),
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

	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	requestBody := requests.CreateApplicationRequest{
		ID:          testutil.ToPtr(uuid.New()),
		CompanyID:   testutil.ToPtr(uuid.New()),
		RecruiterID: testutil.ToPtr(recruiter.ID),
		JobTitle:    testutil.ToPtr("Job Title"),
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

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	requestBody := requests.CreateApplicationRequest{
		ID:          testutil.ToPtr(uuid.New()),
		CompanyID:   testutil.ToPtr(company.ID),
		RecruiterID: testutil.ToPtr(uuid.New()),
		JobTitle:    testutil.ToPtr("Job Title"),
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
	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	requestBody := requests.CreateApplicationRequest{
		ID:                   &id,
		CompanyID:            testutil.ToPtr(company.ID),
		RecruiterID:          testutil.ToPtr(recruiter.ID),
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("country"),
		Area:                 testutil.ToPtr("area"),
		RemoteStatusType:     requests.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(6),
		EstimatedCycleTime:   testutil.ToPtr(7),
		EstimatedCommuteTime: testutil.ToPtr(8),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -20)),
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
	assert.Equal(t, requestBody.RemoteStatusType.String(), response.RemoteStatusType.String())
	assert.Equal(t, *requestBody.WeekdaysInOffice, *response.WeekdaysInOffice)
	assert.Equal(t, *requestBody.EstimatedCycleTime, *response.EstimatedCycleTime)
	assert.Equal(t, *requestBody.EstimatedCommuteTime, *response.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, requestBody.ApplicationDate, response.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, createdDateApproximation, response.CreatedDate)
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

	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	requestBody := requests.CreateApplicationRequest{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      testutil.ToPtr(recruiter.ID),
		JobTitle:         testutil.ToPtr("Software Engineer"),
		RemoteStatusType: requests.RemoteStatusTypeOffice,
	}
	insertApplication(t, applicationHandler, requestBody)

	// get the application by full job title:

	firstGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/title", nil)
	assert.NoError(t, err)
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
	assert.Len(t, firstResponse, 1)

	assert.Equal(t, *requestBody.ID, firstResponse[0].ID)
	assert.Equal(t, requestBody.JobTitle, firstResponse[0].JobTitle)

	// get the application by partial name:

	secondGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/title", nil)
	assert.NoError(t, err)
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
	assert.Len(t, secondResponse, 1)

	assert.Equal(t, *requestBody.ID, secondResponse[0].ID)
	assert.Equal(t, requestBody.JobTitle, secondResponse[0].JobTitle)
}

func TestGetApplicationsByJobTitle_ShouldReturnApplications(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// insert two applications:

	firstCompany := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	firstRequestBody := requests.CreateApplicationRequest{
		ID:               testutil.ToPtr(uuid.New()),
		CompanyID:        testutil.ToPtr(firstCompany.ID),
		JobTitle:         testutil.ToPtr("GoLang Software Engineer"),
		RemoteStatusType: requests.RemoteStatusTypeHybrid,
	}
	insertApplication(t, applicationHandler, firstRequestBody)

	secondRecruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	secondRequestBody := requests.CreateApplicationRequest{
		ID:               testutil.ToPtr(uuid.New()),
		RecruiterID:      testutil.ToPtr(secondRecruiter.ID),
		JobTitle:         testutil.ToPtr("Backend Developer (golang)"),
		RemoteStatusType: requests.RemoteStatusTypeUnknown,
	}
	insertApplication(t, applicationHandler, secondRequestBody)

	// Get applications by name:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/get/title", nil)
	assert.NoError(t, err)
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
	assert.Len(t, response, 2)

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

	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	firstID := uuid.New()
	firstRequestBody := requests.CreateApplicationRequest{
		ID:               &firstID,
		RecruiterID:      testutil.ToPtr(recruiter.ID),
		JobTitle:         testutil.ToPtr("Software Engineer 1"),
		RemoteStatusType: requests.RemoteStatusTypeOffice,
	}
	insertApplication(t, applicationHandler, firstRequestBody)

	secondID := uuid.New()
	secondRequestBody := requests.CreateApplicationRequest{
		ID:               &secondID,
		RecruiterID:      testutil.ToPtr(recruiter.ID),
		JobTitle:         testutil.ToPtr("Software Engineer 2"),
		RemoteStatusType: requests.RemoteStatusTypeRemote,
	}
	insertApplication(t, applicationHandler, secondRequestBody)

	thirdID := uuid.New()
	thirdRequestBody := requests.CreateApplicationRequest{
		ID:               &thirdID,
		RecruiterID:      testutil.ToPtr(recruiter.ID),
		JobTitle:         testutil.ToPtr("Software Engineer 3"),
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
	assert.Len(t, response, 3)

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

	assert.Len(t, response, 0)
}

// -------- UpdateApplication tests: --------

func TestUpdateApplication_ShouldUpdateApplication(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// create an application

	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	id := uuid.New()
	createRequest := requests.CreateApplicationRequest{
		ID:                   &id,
		CompanyID:            testutil.ToPtr(company.ID),
		RecruiterID:          testutil.ToPtr(recruiter.ID),
		JobTitle:             testutil.ToPtr("Job Title"),
		JobAdURL:             testutil.ToPtr("Job Ad URL"),
		Country:              testutil.ToPtr("Some Country"),
		Area:                 testutil.ToPtr("Some Area"),
		RemoteStatusType:     requests.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(9),
		EstimatedCycleTime:   testutil.ToPtr(8),
		EstimatedCommuteTime: testutil.ToPtr(7),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 6)),
	}
	_, createdDateApproximation := insertApplication(t, applicationHandler, createRequest)

	// update the application

	newCompany := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	newRecruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	var newRemoteStatusType requests.RemoteStatusType = requests.RemoteStatusTypeOffice
	updateBody := requests.UpdateApplicationRequest{
		ID:                   id,
		CompanyID:            testutil.ToPtr(newCompany.ID),
		RecruiterID:          testutil.ToPtr(newRecruiter.ID),
		JobTitle:             testutil.ToPtr("New Job Title"),
		JobAdURL:             testutil.ToPtr("New Job Ad URL"),
		Country:              testutil.ToPtr("New Country"),
		Area:                 testutil.ToPtr("New Area"),
		RemoteStatusType:     &newRemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 40)),
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/application/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now()
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
	assert.Equal(t, newCompany.ID, *getApplicationResponse.CompanyID)
	assert.Equal(t, newRecruiter.ID, *getApplicationResponse.RecruiterID)
	assert.Equal(t, updateBody.JobTitle, getApplicationResponse.JobTitle)
	assert.Equal(t, updateBody.JobAdURL, getApplicationResponse.JobAdURL)
	assert.Equal(t, updateBody.Country, getApplicationResponse.Country)
	assert.Equal(t, updateBody.Area, getApplicationResponse.Area)
	assert.Equal(t, updateBody.RemoteStatusType.String(), getApplicationResponse.RemoteStatusType.String())
	assert.Equal(t, updateBody.WeekdaysInOffice, getApplicationResponse.WeekdaysInOffice)
	assert.Equal(t, updateBody.EstimatedCycleTime, getApplicationResponse.EstimatedCycleTime)
	assert.Equal(t, updateBody.EstimatedCommuteTime, getApplicationResponse.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, updateBody.ApplicationDate, getApplicationResponse.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, createdDateApproximation, getApplicationResponse.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, &updatedDateApproximation, getApplicationResponse.UpdatedDate)
}

func TestUpdateApplication_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	applicationHandler, companyRepository := setupApplicationHandler(t)

	// create an application

	id := uuid.New()
	recruiter := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)

	createRequest := requests.CreateApplicationRequest{
		ID:               &id,
		RecruiterID:      testutil.ToPtr(recruiter.ID),
		JobAdURL:         testutil.ToPtr("Job Ad URL"),
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
	company := repositoryhelpers.CreateCompany(t, companyRepository, testutil.ToPtr(uuid.New()), nil)
	requestBody := requests.CreateApplicationRequest{
		ID:               &id,
		CompanyID:        testutil.ToPtr(company.ID),
		JobTitle:         testutil.ToPtr("JobTitle"),
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

func insertApplication(
	t *testing.T, applicationHandler *handlers.ApplicationHandler, requestBody requests.CreateApplicationRequest) (
	*responses.ApplicationResponse, *time.Time) {

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	createRequest, err := http.NewRequest(http.MethodPost, "/api/v1/application/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	applicationHandler.CreateApplication(responseRecorder, createRequest)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var createApplicationResponse responses.ApplicationResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&createApplicationResponse)
	assert.NoError(t, err)

	return &createApplicationResponse, &createdDateApproximation
}
