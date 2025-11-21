package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	configPackage "jobsearchtracker/internal/config"
	internalErrors "jobsearchtracker/internal/errors"
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

func setupEventHandler(t *testing.T) (
	*handlers.EventHandler,
	*repositories.ApplicationRepository,
	*repositories.CompanyRepository,
	*repositories.EventRepository,
	*repositories.PersonRepository,
	*repositories.ApplicationEventRepository,
	*repositories.CompanyEventRepository,
	*repositories.EventPersonRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupEventHandlerTestContainer(t, config)

	var eventHandler *handlers.EventHandler
	err := container.Invoke(func(handler *handlers.EventHandler) {
		eventHandler = handler
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

	var applicationEventRepository *repositories.ApplicationEventRepository
	err = container.Invoke(func(repository *repositories.ApplicationEventRepository) {
		applicationEventRepository = repository
	})
	assert.NoError(t, err)

	var companyEventRepository *repositories.CompanyEventRepository
	err = container.Invoke(func(repository *repositories.CompanyEventRepository) {
		companyEventRepository = repository
	})
	assert.NoError(t, err)

	var eventPersonRepository *repositories.EventPersonRepository
	err = container.Invoke(func(repository *repositories.EventPersonRepository) {
		eventPersonRepository = repository
	})
	assert.NoError(t, err)

	return eventHandler,
		applicationRepository,
		companyRepository,
		eventRepository,
		personRepository,
		applicationEventRepository,
		companyEventRepository,
		eventPersonRepository
}

// -------- CreateEvent tests: --------

func TestCreateEvent_ShouldInsertAndReturnEvent(t *testing.T) {
	eventHandler, _, _, _, _, _, _, _ := setupEventHandler(t)

	requestBody := requests.CreateEventRequest{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   requests.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 5, 0),
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	eventHandler.CreateEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var eventResponse responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&eventResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, eventResponse.ID)
	assert.Equal(t, requestBody.EventType.String(), eventResponse.EventType.String())
	assert.Equal(t, requestBody.Description, eventResponse.Description)
	assert.Equal(t, requestBody.Notes, eventResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &requestBody.EventDate, eventResponse.EventDate)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, eventResponse.CreatedDate, time.Second)
	assert.Nil(t, eventResponse.UpdatedDate)
}

func TestCreateEvent_ShouldInsertAndReturnEventWithOnlyRequiredFields(t *testing.T) {
	eventHandler, _, _, _, _, _, _, _ := setupEventHandler(t)

	requestBody := requests.CreateEventRequest{
		EventType: requests.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 5, 0),
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.CreateEvent(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var eventResponse responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&eventResponse)
	assert.NoError(t, err)

	assert.NotNil(t, eventResponse.ID)
	assert.Equal(t, requestBody.EventType.String(), eventResponse.EventType.String())
	assert.Nil(t, eventResponse.Description)
	assert.Nil(t, eventResponse.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &requestBody.EventDate, eventResponse.EventDate)
	assert.NotNil(t, eventResponse.CreatedDate)
	assert.Nil(t, eventResponse.UpdatedDate)
}

func TestCreateEvent_ShouldReturnStatusConflictIfEventIDIsAlreadyInDB(t *testing.T) {
	eventHandler, _, _, eventRepository, _, _, _, _ := setupEventHandler(t)

	var id = uuid.New()

	repositoryhelpers.CreateEvent(t, eventRepository, &id, nil, nil)

	requestBody := requests.CreateEventRequest{
		ID:        &id,
		EventType: requests.EventTypeApplied,
		EventDate: time.Now().AddDate(0, 5, 0),
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/event/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	eventHandler.CreateEvent(responseRecorder, request)
	assert.Equal(t, http.StatusConflict, responseRecorder.Code)

	expectedError := "Conflict error on insert: ID already exists\n"
	assert.Equal(t, expectedError, responseRecorder.Body.String())
}

// -------- GetAllEvents - Base tests: --------

func TestGetAllEvents_ShouldReturnAllEvents(t *testing.T) {
	eventHandler, _, _, eventRepository, _, _, _, _ := setupEventHandler(t)

	// insert events

	createEvent1 := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	event1, err := eventRepository.Create(&createEvent1)
	assert.NoError(t, err)
	assert.NotNil(t, event1)

	event2ID := uuid.New()
	event2EventDate := time.Now().AddDate(0, 15, 0)
	repositoryhelpers.CreateEvent(t, eventRepository, &event2ID, nil, &event2EventDate)

	// get all events:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, event2ID, response[0].ID)
	assert.NotNil(t, response[0].EventType)
	testutil.AssertEqualFormattedDateTimes(t, &event2EventDate, response[0].EventDate)
	assert.NotNil(t, response[0].CreatedDate)
	assert.Nil(t, response[0].UpdatedDate)

	assert.Equal(t, event1.ID, response[1].ID)
	assert.Equal(t, event1.EventType.String(), response[1].EventType.String())
	assert.Equal(t, event1.Description, response[1].Description)
	assert.Equal(t, event1.Notes, response[1].Notes)
	testutil.AssertEqualFormattedDateTimes(t, event1.EventDate, response[1].EventDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.CreatedDate, response[1].CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, event1.UpdatedDate, response[1].UpdatedDate)
}

func TestGetAllEvents_ShouldReturnEmptyResponseIfNoEventsInDatabase(t *testing.T) {
	eventHandler, _, _, _, _, _, _, _ := setupEventHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Len(t, response, 0)
}

// -------- GetAllEvents - Applications tests: --------

func TestGetAllEvents_ShouldReturnApplicationsIfIncludeApplicationsIsSetToAll(t *testing.T) {
	eventHandler,
		applicationRepository,
		companyRepository,
		eventRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventHandler(t)

	// create events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	// add two companies

	company1ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	company2ID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	createApplication1 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &company1ID,
		RecruiterID:          &company2ID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(0),
		EstimatedCycleTime:   testutil.ToPtr(1),
		EstimatedCommuteTime: testutil.ToPtr(2),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}
	_, err := applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&company1ID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate events and applications

	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, *createApplication1.ID, event1ID, nil)
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, event1ID, nil)
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, event2ID, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_applications=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, event1ID, response[0].ID)
	assert.Len(t, *response[0].Applications, 2)

	assert.Equal(t, application2ID, (*response[0].Applications)[0].ID)

	event1Application2 := (*response[0].Applications)[1]
	assert.Equal(t, *createApplication1.ID, event1Application2.ID)
	assert.Equal(t, createApplication1.CompanyID, event1Application2.CompanyID)
	assert.Equal(t, createApplication1.RecruiterID, event1Application2.RecruiterID)
	assert.Equal(t, createApplication1.JobTitle, event1Application2.JobTitle)
	assert.Equal(t, createApplication1.JobAdURL, event1Application2.JobAdURL)
	assert.Equal(t, createApplication1.Country, event1Application2.Country)
	assert.Equal(t, createApplication1.Area, event1Application2.Area)
	assert.Equal(t, createApplication1.RemoteStatusType.String(), event1Application2.RemoteStatusType.String())
	assert.Equal(t, createApplication1.WeekdaysInOffice, event1Application2.WeekdaysInOffice)
	assert.Equal(t, createApplication1.EstimatedCycleTime, event1Application2.EstimatedCycleTime)
	assert.Equal(t, createApplication1.EstimatedCommuteTime, event1Application2.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.ApplicationDate, event1Application2.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.CreatedDate, event1Application2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createApplication1.UpdatedDate, event1Application2.UpdatedDate)

	assert.Len(t, *response[1].Applications, 1)
	assert.Equal(t, application2ID, (*response[1].Applications)[0].ID)
}

func TestGetAllEvents_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToAllAndThereAreNoApplications(t *testing.T) {
	eventHandler, applicationRepository, companyRepository, eventRepository, _, _, _, _ := setupEventHandler(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_applications=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)
	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, response[0].Applications)
}

func TestGetAllEvents_ShouldReturnApplicationIDsIfIncludeApplicationsIsSetToIDs(t *testing.T) {
	eventHandler,
		applicationRepository,
		companyRepository,
		eventRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventHandler(t)

	// create a event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	createApplication1 := models.CreateApplication{
		ID:                   testutil.ToPtr(uuid.New()),
		CompanyID:            &companyID,
		RecruiterID:          &companyID,
		JobTitle:             testutil.ToPtr("Application1JobTitle"),
		JobAdURL:             testutil.ToPtr("Application1JobAdURL"),
		Country:              testutil.ToPtr("Application1Country"),
		Area:                 testutil.ToPtr("Application1Area"),
		RemoteStatusType:     models.RemoteStatusTypeHybrid,
		WeekdaysInOffice:     testutil.ToPtr(0),
		EstimatedCycleTime:   testutil.ToPtr(1),
		EstimatedCommuteTime: testutil.ToPtr(2),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
	}
	_, err := applicationRepository.Create(&createApplication1)
	assert.NoError(t, err)

	application2ID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate event and applications

	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, *createApplication1.ID, eventID, nil)
	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, application2ID, eventID, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_applications=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, eventID, response[0].ID)
	assert.Len(t, *(response[0]).Applications, 2)

	assert.Equal(t, application2ID, (*response[0].Applications)[0].ID)

	event1Application2 := (*response[0].Applications)[1]
	assert.Equal(t, *createApplication1.ID, event1Application2.ID)
	assert.Nil(t, event1Application2.CompanyID)
	assert.Nil(t, event1Application2.RecruiterID)
	assert.Nil(t, event1Application2.JobTitle)
	assert.Nil(t, event1Application2.JobAdURL)
	assert.Nil(t, event1Application2.Country)
	assert.Nil(t, event1Application2.Area)
	assert.Nil(t, event1Application2.RemoteStatusType)
	assert.Nil(t, event1Application2.WeekdaysInOffice)
	assert.Nil(t, event1Application2.EstimatedCycleTime)
	assert.Nil(t, event1Application2.EstimatedCommuteTime)
	assert.Nil(t, event1Application2.ApplicationDate)
	assert.Nil(t, event1Application2.CreatedDate)
	assert.Nil(t, event1Application2.UpdatedDate)
}

func TestGetAllEvents_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToIDsAndThereAreNoApplications(t *testing.T) {
	eventHandler,
		applicationRepository,
		companyRepository,
		eventRepository,
		_,
		_,
		_,
		_ := setupEventHandler(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add an application

	repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_applications=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)
	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, response[0].Applications)
}

func TestGetAllEvents_ShouldReturnNoApplicationsIfIncludeApplicationsIsSetToNone(t *testing.T) {
	eventHandler,
		applicationRepository,
		companyRepository,
		eventRepository,
		_,
		applicationEventRepository,
		_,
		_ := setupEventHandler(t)

	// create a event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a company

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID

	// add two applications

	applicationID := repositoryhelpers.CreateApplication(
		t,
		applicationRepository,
		nil,
		&companyID,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate event and applications

	repositoryhelpers.AssociateApplicationEvent(t, applicationEventRepository, applicationID, eventID, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_applications=none", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, (response[0]).Applications)
}

// -------- GetAllEvents - Companies tests: --------

func TestGetAllEvents_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	eventHandler, _, companyRepository, eventRepository, _, _, companyEventRepository, _ := setupEventHandler(t)

	// create events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(2, 0, 0))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(3, 0, 0))).ID

	event3ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(4, 0, 0))).ID

	// add two companies

	createCompany1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	// associate events and companies

	Company1Event1 := models.AssociateCompanyEvent{
		CompanyID: *createCompany1.ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company1Event1)
	assert.NoError(t, err)

	Company2Event1 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event1)
	assert.NoError(t, err)

	Company2Event2 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event2)
	assert.NoError(t, err)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_companies=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, event3ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, event2ID, response[1].ID)
	assert.Len(t, *response[1].Companies, 1)
	assert.Equal(t, company2ID, (*(response[1]).Companies)[0].ID)

	assert.Equal(t, event1ID, response[2].ID)
	assert.Len(t, *response[2].Companies, 2)

	event1Company1 := (*(response[2]).Companies)[0]
	assert.Equal(t, *createCompany1.ID, event1Company1.ID)
	assert.Equal(t, createCompany1.Name, *event1Company1.Name)
	assert.Equal(t, createCompany1.CompanyType.String(), event1Company1.CompanyType.String())
	assert.Equal(t, createCompany1.Notes, event1Company1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, createCompany1.LastContact, event1Company1.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, createCompany1.CreatedDate, event1Company1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createCompany1.UpdatedDate, event1Company1.UpdatedDate)

	assert.Equal(t, company2ID, (*(response[2]).Companies)[1].ID)
}

func TestGetAllEvents_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyEventsInRepository(t *testing.T) {
	eventHandler, _, companyRepository, eventRepository, _, _, _, _ := setupEventHandler(t)

	eventID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(2, 0, 0))).ID

	// add two companies

	repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_companies=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, response[0].Companies)
}

func TestGetAllEvents_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	eventHandler, _, companyRepository, eventRepository, _, _, companyEventRepository, _ := setupEventHandler(t)

	// create events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(2, 0, 0))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(3, 0, 0))).ID

	event3ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(4, 0, 0))).ID

	// add two companies

	createCompany1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&createCompany1)
	assert.NoError(t, err)

	company2ID := repositoryhelpers.CreateCompany(
		t,
		companyRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	// associate events and companies

	Company1Event1 := models.AssociateCompanyEvent{
		CompanyID: *createCompany1.ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company1Event1)
	assert.NoError(t, err)

	Company2Event1 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event1ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event1)
	assert.NoError(t, err)

	Company2Event2 := models.AssociateCompanyEvent{
		CompanyID: company2ID,
		EventID:   event2ID,
	}
	_, err = companyEventRepository.AssociateCompanyEvent(&Company2Event2)
	assert.NoError(t, err)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_companies=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, event3ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, event2ID, response[1].ID)
	assert.Len(t, *(response[1]).Companies, 1)
	assert.Equal(t, company2ID, (*(response[1]).Companies)[0].ID)

	assert.Equal(t, event1ID, response[2].ID)
	assert.Len(t, *response[2].Companies, 2)

	event1Company1 := (*(response[2]).Companies)[0]
	assert.Equal(t, *createCompany1.ID, event1Company1.ID)
	assert.Nil(t, event1Company1.Name)
	assert.Nil(t, event1Company1.CompanyType)
	assert.Nil(t, event1Company1.Notes)
	assert.Nil(t, event1Company1.LastContact)
	assert.Nil(t, event1Company1.CreatedDate)
	assert.Nil(t, event1Company1.UpdatedDate)

	event1Company2 := (*(response[2]).Companies)[1]
	assert.Equal(t, company2ID, event1Company2.ID)
}

func TestGetAllEvents_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyEventsInRepository(t *testing.T) {
	eventHandler, _, companyRepository, eventRepository, _, _, _, _ := setupEventHandler(t)

	eventID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(2, 0, 0))).ID

	// add a company

	repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_companies=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, response[0].Companies)
}

func TestGetAllEvents_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	eventHandler, _, companyRepository, eventRepository, _, _, companyEventRepository, _ := setupEventHandler(t)

	eventID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(2, 0, 0))).ID

	// add a company and associate it to an event

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	repositoryhelpers.AssociateCompanyEvent(t, companyEventRepository, companyID, eventID, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_companies=none", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, response[0].Companies)
}

// -------- GetAllEvents - Persons tests: --------

func TestGetAllEvents_ShouldReturnPersonsIfIncludePersonsIsSetToAll(t *testing.T) {
	eventHandler,
		_,
		_,
		eventRepository,
		personRepository,
		_,
		_,
		eventPersonRepository := setupEventHandler(t)

	// create events

	event1ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 3))).ID

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 2))).ID

	// add two persons

	var person1Type models.PersonType = models.PersonTypeJobContact
	createPerson1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1Name",
		PersonType:  person1Type,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err := personRepository.Create(&createPerson1)
	assert.NoError(t, err)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate events and persons

	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event1ID, *createPerson1.ID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event1ID, person2ID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event2ID, person2ID, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_persons=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, event1ID, response[0].ID)
	assert.Len(t, *response[0].Persons, 2)

	assert.Equal(t, person2ID, (*response[0].Persons)[0].ID)

	event1Person2 := (*response[0].Persons)[1]
	assert.Equal(t, *createPerson1.ID, event1Person2.ID)
	assert.Equal(t, createPerson1.Name, *event1Person2.Name)
	assert.Equal(t, createPerson1.PersonType.String(), event1Person2.PersonType.String())
	assert.Equal(t, createPerson1.Email, event1Person2.Email)
	assert.Equal(t, createPerson1.Phone, event1Person2.Phone)
	assert.Equal(t, createPerson1.Notes, event1Person2.Notes)
	testutil.AssertEqualFormattedDateTimes(t, createPerson1.CreatedDate, event1Person2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, createPerson1.UpdatedDate, event1Person2.UpdatedDate)

	assert.Len(t, *response[1].Persons, 1)
	assert.Equal(t, person2ID, (*response[1].Persons)[0].ID)
}

func TestGetAllEvents_ShouldReturnNoPersonsIfIncludePersonsIsSetToAllAndThereAreNoPersons(t *testing.T) {
	eventHandler, _, _, eventRepository, personRepository, _, _, _ := setupEventHandler(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_persons=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)
	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, response[0].Persons)
}

func TestGetAllEvents_ShouldReturnPersonIDsIfIncludePersonsIsSetToIDs(t *testing.T) {
	eventHandler,
		_,
		_,
		eventRepository,
		personRepository,
		_,
		_,
		eventPersonRepository := setupEventHandler(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add two persons

	var person1Type models.PersonType = models.PersonTypeJobContact
	createPerson1 := models.CreatePerson{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Person1Name",
		PersonType:  person1Type,
		Email:       testutil.ToPtr("Person1Email"),
		Phone:       testutil.ToPtr("Person1Phone"),
		Notes:       testutil.ToPtr("Person1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 2)),
	}
	_, err := personRepository.Create(&createPerson1)
	assert.NoError(t, err)

	person2ID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate event and persons

	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, *createPerson1.ID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, person2ID, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_persons=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, eventID, response[0].ID)
	assert.Len(t, *(response[0]).Persons, 2)

	assert.Equal(t, person2ID, (*response[0].Persons)[0].ID)

	event1Person2 := (*response[0].Persons)[1]
	assert.Equal(t, *createPerson1.ID, event1Person2.ID)
	assert.Nil(t, event1Person2.Name)
	assert.Nil(t, event1Person2.PersonType)
	assert.Nil(t, event1Person2.Email)
	assert.Nil(t, event1Person2.Phone)
	assert.Nil(t, event1Person2.Notes)
	assert.Nil(t, event1Person2.CreatedDate)
	assert.Nil(t, event1Person2.UpdatedDate)
}

func TestGetAllEvents_ShouldReturnNoPersonsIfIncludePersonsIsSetToIDsAndThereAreNoPersons(t *testing.T) {
	eventHandler,
		_,
		_,
		eventRepository,
		personRepository,
		_,
		_,
		_ := setupEventHandler(t)

	// create event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add a person

	repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_persons=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)
	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, response[0].Persons)
}

func TestGetAllEvents_ShouldReturnNoPersonsIfIncludePersonsIsSetToNone(t *testing.T) {
	eventHandler,
		_,
		_,
		eventRepository,
		personRepository,
		_,
		_,
		eventPersonRepository := setupEventHandler(t)

	// create an event

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID

	// add two persons

	personID := repositoryhelpers.CreatePerson(
		t,
		personRepository,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 5))).ID

	// associate event and persons

	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, personID, nil)

	// get all events

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/event/get/all?include_persons=none", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.GetAllEvents(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.EventResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, eventID, response[0].ID)
	assert.Nil(t, (response[0]).Persons)
}

// -------- UpdateEvent tests: --------

func TestUpdateEvent_ShouldUpdateEvent(t *testing.T) {
	eventHandler, _, _, eventRepository, _, _, _, _ := setupEventHandler(t)

	// create an event

	createEvent := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("EventDescription"),
		Notes:       testutil.ToPtr("EventNotes"),
		EventDate:   time.Now().AddDate(0, 12, 0),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 13, 0)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 14, 0)),
	}
	event, err := eventRepository.Create(&createEvent)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	// update the event

	var updatedEventType requests.EventType = requests.EventTypeOffer
	updateBody := requests.UpdateEventRequest{
		ID:          *createEvent.ID,
		EventType:   &updatedEventType,
		Description: testutil.ToPtr("Updated Description"),
		Notes:       testutil.ToPtr("Updated Notes"),
		EventDate:   testutil.ToPtr(time.Now().AddDate(4, 0, 0)),
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/event/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now()
	eventHandler.UpdateEvent(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the event by ID

	updatedEvent, err := eventRepository.GetByID(createEvent.ID)
	assert.NoError(t, err)

	assert.Equal(t, updateBody.ID, updatedEvent.ID)
	assert.Equal(t, updateBody.EventType.String(), updatedEvent.EventType.String())
	assert.Equal(t, updateBody.Description, updatedEvent.Description)
	assert.Equal(t, updateBody.Notes, updatedEvent.Notes)
	testutil.AssertEqualFormattedDateTimes(t, updateBody.EventDate, updatedEvent.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, createEvent.CreatedDate, updatedEvent.CreatedDate)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, updatedEvent.UpdatedDate, time.Second)

}

func TestUpdateEvent_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	eventHandler, _, _, eventRepository, _, _, _, _ := setupEventHandler(t)

	// create an event

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	// update the event
	updateBody := requests.UpdateEventRequest{
		ID: event.ID,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/event/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	eventHandler.UpdateEvent(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(
		t,
		"Unable to convert request to internal model: validation error: nothing to update\n",
		responseBodyString)
}

// -------- DeleteEvent tests: --------

func TestDeleteEvent_ShouldDeleteEvent(t *testing.T) {
	eventHandler, _, _, eventRepository, _, _, _, _ := setupEventHandler(t)

	// insert an event

	event := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil)

	// delete the event

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/event/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": event.ID.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	eventHandler.DeleteEvent(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusOK, deleteResponseRecorder.Code)

	// try to get the event

	nilEvent, err := eventRepository.GetByID(&event.ID)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var notFoundError *internalErrors.NotFoundError
	assert.True(t, errors.As(err, &notFoundError))
	assert.Equal(t, "error: object not found: ID: '"+event.ID.String()+"'", notFoundError.Error())
}

func TestDeleteEvent_ShouldReturnStatusNotFoundIfEventDoesNotExist(t *testing.T) {
	eventHandler, _, _, _, _, _, _, _ := setupEventHandler(t)

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/event/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": uuid.New().String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	eventHandler.DeleteEvent(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusNotFound, deleteResponseRecorder.Code)
}
