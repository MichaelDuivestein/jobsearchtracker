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

func setupCompanyHandler(t *testing.T) (
	*handlers.CompanyHandler,
	*repositories.ApplicationRepository,
	*repositories.CompanyRepository,
	*repositories.EventRepository,
	*repositories.PersonRepository,
	*repositories.CompanyEventRepository,
	*repositories.CompanyPersonRepository) {
	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupCompanyHandlerTestContainer(t, config)

	var companyHandler *handlers.CompanyHandler
	err := container.Invoke(func(handler *handlers.CompanyHandler) {
		companyHandler = handler
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	var companyEventRepository *repositories.CompanyEventRepository
	err = container.Invoke(func(repository *repositories.CompanyEventRepository) {
		companyEventRepository = repository
	})
	assert.NoError(t, err)

	var companyPersonRepository *repositories.CompanyPersonRepository
	err = container.Invoke(func(repository *repositories.CompanyPersonRepository) {
		companyPersonRepository = repository
	})
	assert.NoError(t, err)

	return companyHandler,
		applicationRepository,
		companyRepository,
		eventRepository,
		personRepository,
		companyEventRepository,
		companyPersonRepository
}

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldInsertAndReturnReturnCompany(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	requestBody := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "random company name",
		CompanyType: requests.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("Not a lot of notes for this company"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
	}
	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	companyHandler.CreateCompany(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var companyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, companyResponse.ID)
	assert.Equal(t, requestBody.Name, *companyResponse.Name)
	assert.Equal(t, requestBody.CompanyType.String(), companyResponse.CompanyType.String())
	assert.Equal(t, requestBody.Notes, companyResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, requestBody.LastContact, companyResponse.LastContact)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, companyResponse.CreatedDate, time.Second)
	assert.Nil(t, companyResponse.UpdatedDate)
}

func TestCreateCompany_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	requestBody := requests.CreateCompanyRequest{
		Name:        "random company name",
		CompanyType: requests.CompanyTypeRecruiter,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	companyHandler.CreateCompany(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	var responseBodyString = responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var companyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyResponse)
	assert.NoError(t, err)

	assert.Equal(t, requestBody.Name, *companyResponse.Name)
	assert.Equal(t, requestBody.CompanyType.String(), companyResponse.CompanyType.String())

	assert.Nil(t, companyResponse.Notes)
	assert.Nil(t, companyResponse.LastContact)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, companyResponse.CreatedDate, time.Second)
	assert.Nil(t, companyResponse.UpdatedDate)
}

func TestCreateCompany_ShouldReturnStatusConflict_IfCompanyIDIsDuplicate(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

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
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	// Insert the company:

	requestBody := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "random company name",
		CompanyType: requests.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("Not a lot of notes for this company"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
	}
	_, createdDateApproximation := insertCompany(t, companyHandler, requestBody)

	// Get the company:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": requestBody.ID.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	companyHandler.GetCompanyById(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var companyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, companyResponse.ID)
	assert.Equal(t, requestBody.Name, *companyResponse.Name)
	assert.Equal(t, requestBody.CompanyType.String(), companyResponse.CompanyType.String())
	assert.Equal(t, requestBody.Notes, companyResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, requestBody.LastContact, companyResponse.LastContact)
	testutil.AssertDateTimesWithinDelta(t, createdDateApproximation, companyResponse.CreatedDate, time.Second)
	assert.Nil(t, companyResponse.UpdatedDate)
}

func TestGetCompanyById_ShouldReturnNotFoundIfCompanyDoesNotExist(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

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
}

// -------- GetCompanyByName tests: --------

func TestGetCompaniesByName_ShouldReturnCompany(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	// Insert a company:

	requestBody := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("Notes appeared here"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	insertCompany(t, companyHandler, requestBody)

	// Get the company by full name:

	firstGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	assert.NoError(t, err)
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
	assert.Len(t, firstResponse, 1)

	assert.Equal(t, *requestBody.ID, firstResponse[0].ID)
	assert.Equal(t, requestBody.Name, *firstResponse[0].Name)

	// Get the company by partial name:

	secondGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	assert.NoError(t, err)
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
	assert.Len(t, secondResponse, 1)

	assert.Equal(t, *requestBody.ID, secondResponse[0].ID)
	assert.Equal(t, requestBody.Name, *secondResponse[0].Name)
}

func TestGetCompaniesByName_ShouldReturnCompanies(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	// Insert two companies:

	firstRequestBody := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Duck Watchers",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Noteworthy stuff"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	insertCompany(t, companyHandler, firstRequestBody)

	secondRequestBody := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Duck farm",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("More Noteworthy stuff"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 1, 0)),
	}
	insertCompany(t, companyHandler, secondRequestBody)

	// Get companies by name:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	assert.NoError(t, err)
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
	assert.Len(t, response, 2)

	assert.Equal(t, *firstRequestBody.ID, response[0].ID)
	assert.Equal(t, firstRequestBody.Name, *response[0].Name)

	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, secondRequestBody.Name, *response[1].Name)
}

func TestGetCompaniesByName_ShouldReturnNotFoundIfNoCompaniesMatchingName(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

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
	assert.Equal(t, "No companies [partially] matching this name found\n", responseBodyString)
}

// -------- GetAllCompanies  - Base tests: --------

func TestGetAllCompanies_ShouldReturnAllCompanies(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	// create 2 companies

	request1Body := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "company1Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("First Company Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-2, 0, 0)),
	}
	insertCompany(t, companyHandler, request1Body)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	request2Body := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "company2Name",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("Second Company notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(-1, 0, 0)),
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
	assert.Len(t, response, 2)

	assert.Equal(t, *request2Body.ID, response[0].ID)
	assert.Equal(t, *request1Body.ID, response[1].ID)
}

func TestGetAllCompanies_ShouldReturnEmptyResponseIfNoCompaniesInDatabase(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

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
	assert.Len(t, response, 0)
}

// -------- GetAllCompanies - Applications tests: --------

func TestGetAllCompanies_ShouldReturnCompaniesWithApplicationsIfIncludeApplicationsIsAll(t *testing.T) {
	companyHandler, applicationRepository, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup applications
	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	createApplication2 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		RecruiterID:          &companyID,
		JobTitle:             testutil.ToPtr("Application2JobTitle"),
		JobAdURL:             testutil.ToPtr("Application2JobAdURL"),
		Country:              testutil.ToPtr("Application2Country"),
		Area:                 testutil.ToPtr("Application2Area"),
		RemoteStatusType:     models.RemoteStatusTypeRemote,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationRepository.Create(&createApplication2)
	assert.NoError(t, err)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_applications=all", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.NotNil(t, retrievedCompany.Applications)
	assert.Len(t, *retrievedCompany.Applications, 2)

	assert.Equal(t, application1ID, (*retrievedCompany.Applications)[0].ID)
	assert.Equal(t, companyID, *(*retrievedCompany.Applications)[0].CompanyID)
	assert.Nil(t, (*retrievedCompany.Applications)[0].RecruiterID)

	retrievedApplication2 := (*retrievedCompany.Applications)[1]
	assert.Equal(t, *createApplication2.ID, retrievedApplication2.ID)
	assert.Nil(t, retrievedApplication2.CompanyID)
	assert.Equal(t, companyID, *retrievedApplication2.RecruiterID)
	assert.Equal(t, createApplication2.JobTitle, retrievedApplication2.JobTitle)
	assert.Equal(t, createApplication2.JobAdURL, retrievedApplication2.JobAdURL)
	assert.Equal(t, createApplication2.Country, retrievedApplication2.Country)
	assert.Equal(t, createApplication2.Area, retrievedApplication2.Area)
	assert.Equal(t, createApplication2.RemoteStatusType.String(), retrievedApplication2.RemoteStatusType.String())
	assert.Equal(t, *createApplication2.WeekdaysInOffice, *retrievedApplication2.WeekdaysInOffice)
	assert.Equal(t, *createApplication2.EstimatedCycleTime, *retrievedApplication2.EstimatedCycleTime)
	assert.Equal(t, *createApplication2.EstimatedCommuteTime, *retrievedApplication2.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, createApplication2.ApplicationDate, retrievedApplication2.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication2.CreatedDate, retrievedApplication2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication2.UpdatedDate, retrievedApplication2.UpdatedDate)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsAllAndThereAreNoApplications(t *testing.T) {
	companyHandler, _, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_applications=all", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Applications)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithApplicationIDsIfIncludeApplicationsIsIDs(t *testing.T) {
	companyHandler, applicationRepository, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup applications

	application1ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	).ID

	application2 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		RecruiterID:          &companyID,
		JobTitle:             testutil.ToPtr("Application2JobTitle"),
		JobAdURL:             testutil.ToPtr("Application2JobAdURL"),
		Country:              testutil.ToPtr("Application2Country"),
		Area:                 testutil.ToPtr("Application2Area"),
		RemoteStatusType:     models.RemoteStatusTypeRemote,
		WeekdaysInOffice:     testutil.ToPtr(1),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(3),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
	}
	_, err := applicationRepository.Create(&application2)
	assert.NoError(t, err)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_applications=ids", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.NotNil(t, retrievedCompany.Applications)
	assert.Len(t, *retrievedCompany.Applications, 2)

	application := (*retrievedCompany.Applications)[0]
	assert.Equal(t, *application2.ID, application.ID)
	assert.Nil(t, application.CompanyID)
	assert.Equal(t, companyID, *application.RecruiterID)
	assert.Nil(t, application.JobTitle)
	assert.Nil(t, application.JobAdURL)
	assert.Nil(t, application.Country)
	assert.Nil(t, application.Area)
	assert.Nil(t, application.RemoteStatusType)
	assert.Nil(t, application.WeekdaysInOffice)
	assert.Nil(t, application.EstimatedCycleTime)
	assert.Nil(t, application.EstimatedCommuteTime)
	assert.Nil(t, application.ApplicationDate)
	assert.Nil(t, application.CreatedDate)
	assert.Nil(t, application.UpdatedDate)

	assert.Equal(t, application1ID, (*retrievedCompany.Applications)[1].ID)
	assert.Equal(t, companyID, *(*retrievedCompany.Applications)[1].CompanyID)
	assert.Nil(t, (*retrievedCompany.Applications)[1].RecruiterID)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsIDsAndThereAreNoApplications(t *testing.T) {
	companyHandler, _, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_applications=ids", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Applications)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoApplicationsIfIncludeApplicationsIsNone(t *testing.T) {
	companyHandler, applicationRepository, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup applications

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil,
		nil)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_applications=none", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Applications)
}

// -------- GetAllCompanies - Events tests: --------

func TestGetAllCompanies_ShouldReturnCompaniesWithEventsIfIncludeEventsIsAll(t *testing.T) {
	companyHandler, _, companyRepository, eventRepository, _, companyEventRepository, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	createEvent2 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   time.Now().AddDate(0, 0, 1),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}
	_, err := eventRepository.Create(&createEvent2)
	assert.NoError(t, err)

	// associate company to events
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event1ID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, *createEvent2.ID, nil)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_events=all", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.NotNil(t, retrievedCompany.Events)
	assert.Len(t, *retrievedCompany.Events, 2)

	assert.Equal(t, event1ID, (*retrievedCompany.Events)[0].ID)

	event := (*retrievedCompany.Events)[1]
	assert.Equal(t, *createEvent2.ID, event.ID)
	assert.Equal(t, createEvent2.EventType.String(), event.EventType.String())
	assert.Equal(t, createEvent2.Description, event.Description)
	assert.Equal(t, createEvent2.Notes, event.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &createEvent2.EventDate, event.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent2.CreatedDate, event.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent2.UpdatedDate, event.UpdatedDate)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoEventsIfIncludeEventsIsAllAndThereAreNoEvents(t *testing.T) {
	companyHandler, _, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_events=all", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Events)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithEventIDsIfIncludeEventsIsIDs(t *testing.T) {
	companyHandler, _, companyRepository, eventRepository, _, companyEventRepository, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	createEvent2 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("Event2Description"),
		Notes:       testutil.ToPtr("Event2Notes"),
		EventDate:   time.Now().AddDate(0, 0, 1),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}

	_, err := eventRepository.Create(&createEvent2)
	assert.NoError(t, err)

	// setup companyEvents
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, event1ID, nil)
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, *createEvent2.ID, nil)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_events=ids", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.NotNil(t, retrievedCompany.Events)
	assert.Len(t, *retrievedCompany.Events, 2)

	assert.Equal(t, event1ID, (*retrievedCompany.Events)[0].ID)

	event := (*retrievedCompany.Events)[1]
	assert.Equal(t, *createEvent2.ID, event.ID)
	assert.Nil(t, event.EventType)
	assert.Nil(t, event.Description)
	assert.Nil(t, event.Notes)
	assert.Nil(t, event.EventDate)
	assert.Nil(t, event.CreatedDate)
	assert.Nil(t, event.UpdatedDate)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoEventsIfIncludeEventsIsIDsAndThereAreNoEvents(t *testing.T) {
	companyHandler, _, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_events=ids", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Events)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoEventsIfIncludeEventsIsNone(t *testing.T) {
	companyHandler, _, companyRepository, eventRepository, _, companyEventRepository, _ := setupCompanyHandler(t)

	// setup company
	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup event and associate it to company

	eventID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_events=none", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Events)
}

// -------- GetAllCompanies - Persons tests: --------

func TestGetAllCompanies_ShouldReturnCompaniesWithPersonsIfIncludePersonsIsAll(t *testing.T) {
	companyHandler, _, companyRepository, _, personRepository, _, companyPersonRepository := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup persons

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	createPerson2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person 2 Name",
		PersonType:  models.PersonTypeCTO,
		Email:       testutil.ToPtr("Person 2 Email"),
		Phone:       testutil.ToPtr("Person 2 Phone"),
		Notes:       testutil.ToPtr("Person 2 Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err := personRepository.Create(&createPerson2)
	assert.NoError(t, err)

	// associate company to persons
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person1ID, nil)
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, *createPerson2.ID, nil)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_persons=all", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.NotNil(t, retrievedCompany.Persons)
	assert.Len(t, *retrievedCompany.Persons, 2)

	assert.Equal(t, person1ID, (*retrievedCompany.Persons)[0].ID)

	person := (*retrievedCompany.Persons)[1]
	assert.Equal(t, *createPerson2.ID, person.ID)
	assert.Equal(t, createPerson2.Name, *person.Name)
	assert.Equal(t, createPerson2.PersonType.String(), person.PersonType.String())
	assert.Equal(t, createPerson2.Email, person.Email)
	assert.Equal(t, createPerson2.Phone, person.Phone)
	assert.Equal(t, createPerson2.Notes, person.Notes)
	testutil.AssertEqualFormattedDateTimes(t, createPerson2.CreatedDate, person.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createPerson2.UpdatedDate, person.UpdatedDate)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoPersonsIfIncludePersonsIsAllAndThereAreNoPersons(t *testing.T) {
	companyHandler, _, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_persons=all", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Persons)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithPersonIDsIfIncludePersonsIsIDs(t *testing.T) {
	companyHandler, _, companyRepository, _, personRepository, _, companyPersonRepository := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup persons

	person1ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID

	person2 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person 2 Name",
		PersonType:  models.PersonTypeCTO,
		Email:       testutil.ToPtr("Person 2 Email"),
		Phone:       testutil.ToPtr("Person 2 Phone"),
		Notes:       testutil.ToPtr("Person 2 Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 1)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
	}
	_, err := personRepository.Create(&person2)
	assert.NoError(t, err)

	// setup companyPersons
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, person1ID, nil)
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, *person2.ID, nil)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_persons=ids", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.NotNil(t, retrievedCompany.Persons)
	assert.Len(t, *retrievedCompany.Persons, 2)

	assert.Equal(t, person1ID, (*retrievedCompany.Persons)[0].ID)

	person := (*retrievedCompany.Persons)[1]
	assert.Equal(t, *person2.ID, person.ID)
	assert.Nil(t, person.Name)
	assert.Nil(t, person.PersonType)
	assert.Nil(t, person.Email)
	assert.Nil(t, person.Phone)
	assert.Nil(t, person.Notes)
	assert.Nil(t, person.CreatedDate)
	assert.Nil(t, person.UpdatedDate)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoPersonsIfIncludePersonsIsIDsAndThereAreNoPersons(t *testing.T) {
	companyHandler, _, companyRepository, _, _, _, _ := setupCompanyHandler(t)

	// setup company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_persons=ids", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Persons)
}

func TestGetAllCompanies_ShouldReturnCompaniesWithNoPersonsIfIncludePersonsIsNone(t *testing.T) {
	companyHandler, _, companyRepository, _, personRepository, _, companyPersonRepository := setupCompanyHandler(t)

	// setup company
	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// setup person and associate it to company

	personID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	).ID
	repositoryhelpers.AssociateCompanyPerson(t, companyPersonRepository, companyID, personID, nil)

	// get all companies

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_persons=none", nil)
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
	assert.Len(t, response, 1)

	assert.Equal(t, companyID, response[0].ID)
	retrievedCompany := response[0]

	assert.Nil(t, retrievedCompany.Persons)
}

// -------- Update tests: --------

func TestUpdateCompany_ShouldUpdateCompany(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	// create a company

	createRequest := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "companyName",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Notes here"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 2, 0)),
	}
	_, createdDateApproximation := insertCompany(t, companyHandler, createRequest)

	// update the company

	var updatedCompanyType requests.CompanyType = requests.CompanyTypeConsultancy
	updateBody := requests.UpdateCompanyRequest{
		ID:          *createRequest.ID,
		Name:        testutil.ToPtr("Updated Name"),
		CompanyType: &updatedCompanyType,
		Notes:       testutil.ToPtr("Updated Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -4)),
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/company/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now()
	companyHandler.UpdateCompany(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the company by ID

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": createRequest.ID.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	companyHandler.GetCompanyById(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var getCompanyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&getCompanyResponse)
	assert.NoError(t, err)

	assert.Equal(t, updateBody.ID, getCompanyResponse.ID)
	assert.Equal(t, *updateBody.Name, *getCompanyResponse.Name)
	assert.Equal(t, updatedCompanyType.String(), getCompanyResponse.CompanyType.String())
	assert.Equal(t, *updateBody.Notes, *getCompanyResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, updateBody.LastContact, getCompanyResponse.LastContact)
	testutil.AssertDateTimesWithinDelta(t, createdDateApproximation, getCompanyResponse.CreatedDate, time.Second)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, getCompanyResponse.UpdatedDate, time.Second)
}

func TestUpdateCompany_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	// create a company

	createRequest := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Nameless Company",
		CompanyType: models.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("Notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
	}
	insertCompany(t, companyHandler, createRequest)

	// update the company

	updateBody := requests.UpdateCompanyRequest{
		ID: *createRequest.ID,
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
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

	// insert the company

	requestBody := requests.CreateCompanyRequest{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Keeping company",
		CompanyType: requests.CompanyTypeConsultancy,
		Notes:       testutil.ToPtr("Noting things"),
		LastContact: testutil.ToPtr(time.Now()),
	}
	insertCompany(t, companyHandler, requestBody)

	// delete the company

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/company/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": requestBody.ID.String(),
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
	companyHandler, _, _, _, _, _, _ := setupCompanyHandler(t)

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
	t *testing.T,
	companyHandler *handlers.CompanyHandler,
	requestBody requests.CreateCompanyRequest) (*responses.CompanyResponse, *time.Time) {

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	createRequest, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err, "Failed to create request")

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	companyHandler.CreateCompany(responseRecorder, createRequest)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var createCompanyResponse responses.CompanyResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&createCompanyResponse)
	assert.NoError(t, err)

	return &createCompanyResponse, &createdDateApproximation
}
