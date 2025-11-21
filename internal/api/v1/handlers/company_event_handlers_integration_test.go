package handlers_test

import (
	"bytes"
	"encoding/json"
	"jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	configPackage "jobsearchtracker/internal/config"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil/dependencyinjection"
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupCompanyEventHandler(t *testing.T) (
	*handlers.CompanyEventHandler, *repositories.CompanyRepository, *repositories.EventRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupCompanyEventHandlerTestContainer(t, config)

	var companyEventHandler *handlers.CompanyEventHandler
	err := container.Invoke(func(handler *handlers.CompanyEventHandler) {
		companyEventHandler = handler
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

	return companyEventHandler, companyRepository, eventRepository
}

// -------- AssociateCompanyEvent tests: --------

func TestAssociateCompanyEvent_ShouldWork(t *testing.T) {
	companyEventHandler, companyRepository, eventRepository := setupCompanyEventHandler(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent := requests.AssociateCompanyEventRequest{
		CompanyID: company.ID,
		EventID:   event.ID,
	}

	requestBytes, err := json.Marshal(companyEvent)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/company-event/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	companyEventHandler.AssociateCompanyEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var companyEventResponse responses.CompanyEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyEventResponse)
	assert.NoError(t, err)

	assert.Equal(t, company.ID, companyEventResponse.CompanyID)
	assert.Equal(t, event.ID, companyEventResponse.EventID)
	assert.NotNil(t, companyEventResponse.CreatedDate)
}

// -------- GetCompanyEventsByID tests: --------

func TestGetCompanyEventsByID_ShouldWork(t *testing.T) {
	companyEventHandler, companyRepository, eventRepository := setupCompanyEventHandler(t)

	_, company2ID, event1ID, _ :=
		setupCompanyEventTestData(t, companyEventHandler, companyRepository, eventRepository, false)

	queryParams := "company-id=" + company2ID.String() + "&event-id=" + event1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyEventHandler.GetCompanyEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)
}

func TestGetCompanyEventsByID_ShouldReturnAllMatchingCompanies(t *testing.T) {
	companyEventHandler, companyRepository, eventRepository := setupCompanyEventHandler(t)

	_, company2ID, event1ID, event2ID :=
		setupCompanyEventTestData(t, companyEventHandler, companyRepository, eventRepository, true)

	queryParams := "company-id=" + company2ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyEventHandler.GetCompanyEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, company2ID, response[1].CompanyID)
	assert.Equal(t, event2ID, response[1].EventID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetCompanyEventsByID_ShouldReturnAllMatchingEvents(t *testing.T) {
	companyEventHandler, companyRepository, eventRepository := setupCompanyEventHandler(t)

	company1ID, company2ID, event1ID, _ :=
		setupCompanyEventTestData(t, companyEventHandler, companyRepository, eventRepository, true)

	queryParams := "event-id=" + event1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyEventHandler.GetCompanyEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, company1ID, response[1].CompanyID)
	assert.Equal(t, event1ID, response[1].EventID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetCompanyEventsByID_ShouldReturnEmptyResponseIfNoMatchingCompanyEvents(t *testing.T) {
	companyEventHandler, companyRepository, eventRepository := setupCompanyEventHandler(t)

	setupCompanyEventTestData(t, companyEventHandler, companyRepository, eventRepository, false)

	queryParams := "company-id=" + uuid.New().String() + "&event-id=" + uuid.New().String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyEventHandler.GetCompanyEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- GetAllCompanyEvents tests: --------

func TestGetAllCompanyEvents_ShouldReturnAllCompanyEvents(t *testing.T) {
	companyEventHandler, companyRepository, eventRepository := setupCompanyEventHandler(t)

	company1ID, company2ID, event1ID, event2ID :=
		setupCompanyEventTestData(t, companyEventHandler, companyRepository, eventRepository, true)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyEventHandler.GetAllCompanyEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, company2ID, response[1].CompanyID)
	assert.Equal(t, event2ID, response[1].EventID)
	assert.NotNil(t, response[1].CreatedDate)

	assert.Equal(t, company1ID, response[2].CompanyID)
	assert.Equal(t, event1ID, response[2].EventID)
	assert.NotNil(t, response[2].CreatedDate)
}

func TestGetAllCompanyEvents_ShouldReturnNothingIfNothingInDatabase(t *testing.T) {
	companyEventHandler, _, _ := setupCompanyEventHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyEventHandler.GetAllCompanyEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- DeleteCompanyEvent tests: --------

func TestDeleteCompanyEvent_ShouldDeleteCompanyEvent(t *testing.T) {
	companyEventHandler, companyRepository, eventRepository := setupCompanyEventHandler(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent := requests.AssociateCompanyEventRequest{
		CompanyID: company.ID,
		EventID:   event.ID,
	}

	requestBytes, err := json.Marshal(companyEvent)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/company-event/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	companyEventHandler.AssociateCompanyEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	deleteRequest := requests.DeleteCompanyEventRequest{
		CompanyID: company.ID,
		EventID:   event.ID,
	}

	requestBytes, err = json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err = http.NewRequest("POST", "/api/v1/company-event/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()
	companyEventHandler.DeleteCompanyEvent(responseRecorder, request)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "", responseBodyString)
}

func TestDeleteCompanyEvent_ShouldReturnErrorIfNoMatchingCompanyEventToDelete(t *testing.T) {
	companyEventHandler, _, _ := setupCompanyEventHandler(t)

	companyID, eventID := uuid.New(), uuid.New()
	deleteRequest := requests.DeleteCompanyEventRequest{
		CompanyID: companyID,
		EventID:   eventID,
	}

	requestBytes, err := json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/company-event/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	companyEventHandler.DeleteCompanyEvent(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"error: object not found: CompanyEvent does not exist. companyID: "+
			companyID.String()+", eventID: "+eventID.String()+"\n",
		responseBodyString)
}

// -------- test helpers: --------

func setupCompanyEventTestData(
	t *testing.T,
	companyEventHandler *handlers.CompanyEventHandler,
	companyRepository *repositories.CompanyRepository,
	eventRepository *repositories.EventRepository,
	sleep bool) (
	uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	companyEvent1 := requests.AssociateCompanyEventRequest{
		CompanyID: company1.ID,
		EventID:   event1.ID,
	}
	requestBytes, err := json.Marshal(companyEvent1)
	assert.NoError(t, err)
	request, err := http.NewRequest("POST", "/api/v1/company-event/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder := httptest.NewRecorder()
	companyEventHandler.AssociateCompanyEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	companyEvent2 := requests.AssociateCompanyEventRequest{
		CompanyID: company2.ID,
		EventID:   event2.ID,
	}
	requestBytes, err = json.Marshal(companyEvent2)
	assert.NoError(t, err)
	request, err = http.NewRequest("POST", "/api/v1/company-event/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder = httptest.NewRecorder()
	companyEventHandler.AssociateCompanyEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	companyEvent3 := requests.AssociateCompanyEventRequest{
		CompanyID: company2.ID,
		EventID:   event1.ID,
	}
	requestBytes, err = json.Marshal(companyEvent3)
	assert.NoError(t, err)
	request, err = http.NewRequest("POST", "/api/v1/company-event/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder = httptest.NewRecorder()
	companyEventHandler.AssociateCompanyEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	return company1.ID, company2.ID, event1.ID, event2.ID
}
