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

func setupApplicationPersonHandler(t *testing.T) (
	*handlers.ApplicationPersonHandler,
	*repositories.ApplicationRepository,
	*repositories.PersonRepository,
	*repositories.CompanyRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupApplicationPersonHandlerTestContainer(t, config)

	var applicationPersonHandler *handlers.ApplicationPersonHandler
	err := container.Invoke(func(handler *handlers.ApplicationPersonHandler) {
		applicationPersonHandler = handler
	})
	assert.NoError(t, err)

	var applicationRepository *repositories.ApplicationRepository
	err = container.Invoke(func(repository *repositories.ApplicationRepository) {
		applicationRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	return applicationPersonHandler, applicationRepository, personRepository, companyRepository
}

// -------- AssociateApplicationPerson tests: --------

func TestAssociateApplicationPerson_ShouldWork(t *testing.T) {
	applicationPersonHandler, applicationRepository, personRepository, companyRepository := setupApplicationPersonHandler(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := requests.AssociateApplicationPersonRequest{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}

	requestBytes, err := json.Marshal(applicationPerson)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/application-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	applicationPersonHandler.AssociateApplicationPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var applicationPersonResponse responses.ApplicationPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&applicationPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, application.ID, applicationPersonResponse.ApplicationID)
	assert.Equal(t, person.ID, applicationPersonResponse.PersonID)
	assert.NotNil(t, applicationPersonResponse.CreatedDate)
}

// -------- GetApplicationPersonsByID tests: --------

func TestGetApplicationPersonsByID_ShouldWork(t *testing.T) {
	applicationPersonHandler, applicationRepository, personRepository, companyRepository := setupApplicationPersonHandler(t)

	_, application2ID, person1ID, _ :=
		setupApplicationPersonTestData(t, applicationPersonHandler, applicationRepository, personRepository, companyRepository, false)

	queryParams := "application-id=" + application2ID.String() + "&person-id=" + person1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationPersonHandler.GetApplicationPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)
}

func TestGetApplicationPersonsByID_ShouldReturnAllMatchingCompanies(t *testing.T) {
	applicationPersonHandler, applicationRepository, personRepository, companyRepository := setupApplicationPersonHandler(t)

	_, application2ID, person1ID, person2ID :=
		setupApplicationPersonTestData(t, applicationPersonHandler, applicationRepository, personRepository, companyRepository, true)

	queryParams := "application-id=" + application2ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationPersonHandler.GetApplicationPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, application2ID, response[1].ApplicationID)
	assert.Equal(t, person2ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetApplicationPersonsByID_ShouldReturnAllMatchingPersons(t *testing.T) {
	applicationPersonHandler, applicationRepository, personRepository, companyRepository := setupApplicationPersonHandler(t)

	application1ID, application2ID, person1ID, _ :=
		setupApplicationPersonTestData(t, applicationPersonHandler, applicationRepository, personRepository, companyRepository, true)

	queryParams := "person-id=" + person1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationPersonHandler.GetApplicationPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, application1ID, response[1].ApplicationID)
	assert.Equal(t, person1ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetApplicationPersonsByID_ShouldReturnEmptyResponseIfNoMatchingApplicationPersons(t *testing.T) {
	applicationPersonHandler, applicationRepository, personRepository, companyRepository := setupApplicationPersonHandler(t)

	setupApplicationPersonTestData(t, applicationPersonHandler, applicationRepository, personRepository, companyRepository, false)

	queryParams := "application-id=" + uuid.New().String() + "&person-id=" + uuid.New().String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationPersonHandler.GetApplicationPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- GetAllApplicationPersons tests: --------

func TestGetAllApplicationPersons_ShouldReturnAllApplicationPersons(t *testing.T) {
	applicationPersonHandler, applicationRepository, personRepository, companyRepository := setupApplicationPersonHandler(t)

	application1ID, application2ID, person1ID, person2ID :=
		setupApplicationPersonTestData(t, applicationPersonHandler, applicationRepository, personRepository, companyRepository, true)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationPersonHandler.GetAllApplicationPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, application2ID, response[0].ApplicationID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, application2ID, response[1].ApplicationID)
	assert.Equal(t, person2ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)

	assert.Equal(t, application1ID, response[2].ApplicationID)
	assert.Equal(t, person1ID, response[2].PersonID)
	assert.NotNil(t, response[2].CreatedDate)
}

func TestGetAllApplicationPersons_ShouldReturnNothingIfNothingInDatabase(t *testing.T) {
	applicationPersonHandler, _, _, _ := setupApplicationPersonHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/application/application-person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	applicationPersonHandler.GetAllApplicationPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.ApplicationPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- DeleteApplicationPerson tests: --------

func TestDeleteApplicationPerson_ShouldDeleteApplicationPerson(t *testing.T) {
	applicationPersonHandler, applicationRepository, personRepository, companyRepository := setupApplicationPersonHandler(t)

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson := requests.AssociateApplicationPersonRequest{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}

	requestBytes, err := json.Marshal(applicationPerson)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/application-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	applicationPersonHandler.AssociateApplicationPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	deleteRequest := requests.DeleteApplicationPersonRequest{
		ApplicationID: application.ID,
		PersonID:      person.ID,
	}

	requestBytes, err = json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err = http.NewRequest("POST", "/api/v1/application-person/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()
	applicationPersonHandler.DeleteApplicationPerson(responseRecorder, request)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "", responseBodyString)
}

func TestDeleteApplicationPerson_ShouldReturnErrorIfNoMatchingApplicationPersonToDelete(t *testing.T) {
	applicationPersonHandler, _, _, _ := setupApplicationPersonHandler(t)

	applicationID, personID := uuid.New(), uuid.New()
	deleteRequest := requests.DeleteApplicationPersonRequest{
		ApplicationID: applicationID,
		PersonID:      personID,
	}

	requestBytes, err := json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/application-person/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	applicationPersonHandler.DeleteApplicationPerson(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"error: object not found: ApplicationPerson does not exist. applicationID: "+
			applicationID.String()+", personID: "+personID.String()+"\n",
		responseBodyString)
}

// -------- test helpers: --------

func setupApplicationPersonTestData(
	t *testing.T,
	applicationPersonHandler *handlers.ApplicationPersonHandler,
	applicationRepository *repositories.ApplicationRepository,
	personRepository *repositories.PersonRepository,
	companyRepository *repositories.CompanyRepository,
	sleep bool) (
	uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {

	companyID := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil).ID
	application1 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	application2 := repositoryhelpers.CreateApplication(t, applicationRepository, nil, &companyID, nil, nil)
	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	applicationPerson1 := requests.AssociateApplicationPersonRequest{
		ApplicationID: application1.ID,
		PersonID:      person1.ID,
	}
	requestBytes, err := json.Marshal(applicationPerson1)
	assert.NoError(t, err)
	request, err := http.NewRequest("POST", "/api/v1/application-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder := httptest.NewRecorder()
	applicationPersonHandler.AssociateApplicationPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	applicationPerson2 := requests.AssociateApplicationPersonRequest{
		ApplicationID: application2.ID,
		PersonID:      person2.ID,
	}
	requestBytes, err = json.Marshal(applicationPerson2)
	assert.NoError(t, err)
	request, err = http.NewRequest("POST", "/api/v1/application-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder = httptest.NewRecorder()
	applicationPersonHandler.AssociateApplicationPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	applicationPerson3 := requests.AssociateApplicationPersonRequest{
		ApplicationID: application2.ID,
		PersonID:      person1.ID,
	}
	requestBytes, err = json.Marshal(applicationPerson3)
	assert.NoError(t, err)
	request, err = http.NewRequest("POST", "/api/v1/application-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder = httptest.NewRecorder()
	applicationPersonHandler.AssociateApplicationPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	return application1.ID, application2.ID, person1.ID, person2.ID
}
