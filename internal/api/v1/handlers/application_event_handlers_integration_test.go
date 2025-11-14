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
	"jobsearchtracker/internal/testutil/repositoryhelpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupApplicationEventHandler(t *testing.T) (
	*handlers.ApplicationEventHandler,
	*repositories.ApplicationRepository,
	*repositories.EventRepository,
	*repositories.CompanyRepository,
	*repositories.ApplicationEventRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupApplicationEventHandlerTestContainer(t, config)

	var applicationEventHandler *handlers.ApplicationEventHandler
	err := container.Invoke(func(handler *handlers.ApplicationEventHandler) {
		applicationEventHandler = handler
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var eventRepository *repositories.EventRepository
	err = container.Invoke(func(repository *repositories.EventRepository) {
		eventRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var applicationEventRepository *repositories.ApplicationEventRepository
	err = container.Invoke(func(repository *repositories.ApplicationEventRepository) {
		applicationEventRepository = repository
	})

	return applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository
}

// -------- AssociateApplicationEvent tests: --------

func TestAssociateApplicationEvent_ShouldWork(t *testing.T) {
	applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository,
		_ := setupApplicationEventHandler(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := requests.AssociateApplicationEventRequest{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}

	requestBytes, err := json.Marshal(applicationEvent)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/application-event/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	applicationEventHandler.AssociateApplicationEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var applicationEventResponse responses.ApplicationEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&applicationEventResponse)
	assert.NoError(t, err)

	assert.Equal(t, application.ID, applicationEventResponse.ApplicationID)
	assert.Equal(t, event.ID, applicationEventResponse.EventID)
	assert.NotNil(t, applicationEventResponse.CreatedDate)
}

// -------- GetApplicationEventsByID tests: --------

func TestGetApplicationEventsByID_ShouldWork(t *testing.T) {
	applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository := setupApplicationEventHandler(t)

	_, application2ID, event1ID, _ := setupApplicationEventTestData(
		t,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository,
		false)

	queryParams := "application-id=" + application2ID.String() + "&event-id=" + event1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationEventHandler.GetApplicationEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)
}

func TestGetApplicationEventsByID_ShouldReturnAllMatchingCompanies(t *testing.T) {
	applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository := setupApplicationEventHandler(t)

	_, application2ID, event1ID, event2ID := setupApplicationEventTestData(
		t,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository,
		true)

	queryParams := "application-id=" + application2ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationEventHandler.GetApplicationEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, application2ID, response[1].ApplicationID)
	assert.Equal(t, event2ID, response[1].EventID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetApplicationEventsByID_ShouldReturnAllMatchingEvents(t *testing.T) {
	applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository := setupApplicationEventHandler(t)

	application1ID, application2ID, event1ID, _ := setupApplicationEventTestData(
		t,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository,
		true)

	queryParams := "event-id=" + event1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationEventHandler.GetApplicationEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, application1ID, response[1].ApplicationID)
	assert.Equal(t, event1ID, response[1].EventID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetApplicationEventsByID_ShouldReturnEmptyResponseIfNoMatchingApplicationEvents(t *testing.T) {
	applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository := setupApplicationEventHandler(t)

	setupApplicationEventTestData(
		t,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository,
		false)

	queryParams := "application-id=" + uuid.New().String() + "&event-id=" + uuid.New().String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-event/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationEventHandler.GetApplicationEventsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- GetAllApplicationEvents tests: --------

func TestGetAllApplicationEvents_ShouldReturnAllApplicationEvents(t *testing.T) {
	applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository := setupApplicationEventHandler(t)

	application1ID, application2ID, event1ID, event2ID := setupApplicationEventTestData(
		t,
		applicationRepository,
		eventRepository,
		companyRepository,
		applicationEventRepository,
		true)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationEventHandler.GetAllApplicationEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, event1ID, response[0].EventID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, application2ID, response[1].ApplicationID)
	assert.Equal(t, event2ID, response[1].EventID)
	assert.NotNil(t, response[1].CreatedDate)

	assert.Equal(t, application1ID, response[2].ApplicationID)
	assert.Equal(t, event1ID, response[2].EventID)
	assert.NotNil(t, response[2].CreatedDate)
}

func TestGetAllApplicationEvents_ShouldReturnNothingIfNothingInDatabase(t *testing.T) {
	applicationEventHandler, _, _, _, _ := setupApplicationEventHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationEventHandler.GetAllApplicationEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationEventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- DeleteApplicationEvent tests: --------

func TestDeleteApplicationEvent_ShouldDeleteApplicationEvent(t *testing.T) {
	applicationEventHandler,
		applicationRepository,
		eventRepository,
		companyRepository, _ := setupApplicationEventHandler(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent := requests.AssociateApplicationEventRequest{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}

	requestBytes, err := json.Marshal(applicationEvent)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/application-event/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	applicationEventHandler.AssociateApplicationEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	deleteRequest := requests.DeleteApplicationEventRequest{
		ApplicationID: application.ID,
		EventID:       event.ID,
	}

	requestBytes, err = json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err = http.NewRequest("POST", "/api/v1/application-event/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()
	applicationEventHandler.DeleteApplicationEvent(responseRecorder, request)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "", responseBodyString)
}

func TestDeleteApplicationEvent_ShouldReturnErrorIfNoMatchingApplicationEventToDelete(t *testing.T) {
	applicationEventHandler, _, _, _, _ := setupApplicationEventHandler(t)

	applicationID, eventID := uuid.New(), uuid.New()
	deleteRequest := requests.DeleteApplicationEventRequest{
		ApplicationID: applicationID,
		EventID:       eventID,
	}

	requestBytes, err := json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/application-event/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	applicationEventHandler.DeleteApplicationEvent(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"error: object not found: ApplicationEvent does not exist. applicationID: "+
			applicationID.String()+", eventID: "+eventID.String()+"\n",
		responseBodyString)
}

// -------- test helpers: --------

func setupApplicationEventTestData(
	t *testing.T,
	applicationRepository *repositories.ApplicationRepository,
	eventRepository *repositories.EventRepository,
	companyRepository *repositories.CompanyRepository,
	applicationEventRepository *repositories.ApplicationEventRepository,
	sleep bool) (
	uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	event1 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)
	event2 := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	applicationEvent1 := models.AssociateApplicationEvent{
		ApplicationID: application1.ID,
		EventID:       event1.ID,
	}
	_, err := applicationEventRepository.AssociateApplicationEvent(&applicationEvent1)
	assert.NoError(t, err)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	applicationEvent2 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event2.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent2)
	assert.NoError(t, err)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	applicationEvent3 := models.AssociateApplicationEvent{
		ApplicationID: application2.ID,
		EventID:       event1.ID,
	}
	_, err = applicationEventRepository.AssociateApplicationEvent(&applicationEvent3)
	assert.NoError(t, err)

	return application1.ID, application2.ID, event1.ID, event2.ID
}
