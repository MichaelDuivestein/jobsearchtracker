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

func setupCompanyPersonHandler(t *testing.T) (
	*handlers.CompanyPersonHandler, *repositories.CompanyRepository, *repositories.PersonRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}

	container := dependencyinjection.SetupCompanyPersonHandlerTestContainer(t, config)

	var companyPersonHandler *handlers.CompanyPersonHandler
	err := container.Invoke(func(handler *handlers.CompanyPersonHandler) {
		companyPersonHandler = handler
	})
	assert.NoError(t, err)

	var companyRepository *repositories.CompanyRepository
	err = container.Invoke(func(repository *repositories.CompanyRepository) {
		companyRepository = repository
	})
	assert.NoError(t, err)

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	return companyPersonHandler, companyRepository, personRepository
}

// -------- AssociateCompanyPerson tests: --------

func TestAssociateCompanyPerson_ShouldWork(t *testing.T) {
	companyPersonHandler, companyRepository, personRepository := setupCompanyPersonHandler(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := requests.AssociateCompanyPersonRequest{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}

	requestBytes, err := json.Marshal(companyPerson)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/company-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	companyPersonHandler.AssociateCompanyPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var companyPersonResponse responses.CompanyPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&companyPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, company.ID, companyPersonResponse.CompanyID)
	assert.Equal(t, person.ID, companyPersonResponse.PersonID)
	assert.NotNil(t, companyPersonResponse.CreatedDate)
}

// -------- GetCompanyPersonsByID tests: --------

func TestGetCompanyPersonsByID_ShouldWork(t *testing.T) {
	companyPersonHandler, companyRepository, personRepository := setupCompanyPersonHandler(t)

	_, company2ID, person1ID, _ :=
		setupTestData(t, companyPersonHandler, companyRepository, personRepository, false)

	queryParams := "company-id=" + company2ID.String() + "&person-id=" + person1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyPersonHandler.GetCompanyPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)
}

func TestGetCompanyPersonsByID_ShouldReturnAllMatchingCompanies(t *testing.T) {
	companyPersonHandler, companyRepository, personRepository := setupCompanyPersonHandler(t)

	_, company2ID, person1ID, person2ID :=
		setupTestData(t, companyPersonHandler, companyRepository, personRepository, true)

	queryParams := "company-id=" + company2ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyPersonHandler.GetCompanyPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, company2ID, response[1].CompanyID)
	assert.Equal(t, person2ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetCompanyPersonsByID_ShouldReturnAllMatchingPersons(t *testing.T) {
	companyPersonHandler, companyRepository, personRepository := setupCompanyPersonHandler(t)

	company1ID, company2ID, person1ID, _ :=
		setupTestData(t, companyPersonHandler, companyRepository, personRepository, true)

	queryParams := "person-id=" + person1ID.String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyPersonHandler.GetCompanyPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, company1ID, response[1].CompanyID)
	assert.Equal(t, person1ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)
}

func TestGetCompanyPersonsByID_ShouldReturnEmptyResponseIfNoMatchingCompanyPersons(t *testing.T) {
	companyPersonHandler, companyRepository, personRepository := setupCompanyPersonHandler(t)

	setupTestData(t, companyPersonHandler, companyRepository, personRepository, false)

	queryParams := "company-id=" + uuid.New().String() + "&person-id=" + uuid.New().String()

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-person/get/?"+queryParams, nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyPersonHandler.GetCompanyPersonsByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- GetAllCompanyPersons tests: --------

func TestGetAllCompanyPersons_ShouldReturnAllCompanyPersons(t *testing.T) {
	companyPersonHandler, companyRepository, personRepository := setupCompanyPersonHandler(t)

	company1ID, company2ID, person1ID, person2ID :=
		setupTestData(t, companyPersonHandler, companyRepository, personRepository, true)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyPersonHandler.GetAllCompanyPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, company2ID, response[0].CompanyID)
	assert.Equal(t, person1ID, response[0].PersonID)
	assert.NotNil(t, response[0].CreatedDate)

	assert.Equal(t, company2ID, response[1].CompanyID)
	assert.Equal(t, person2ID, response[1].PersonID)
	assert.NotNil(t, response[1].CreatedDate)

	assert.Equal(t, company1ID, response[2].CompanyID)
	assert.Equal(t, person1ID, response[2].PersonID)
	assert.NotNil(t, response[2].CreatedDate)
}

func TestGetAllCompanyPersons_ShouldReturnNothingIfNothingInDatabase(t *testing.T) {
	companyPersonHandler, _, _ := setupCompanyPersonHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/company/company-person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyPersonHandler.GetAllCompanyPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.CompanyPersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

// -------- DeleteCompanyPerson tests: --------

func TestDeleteCompanyPerson_ShouldDeleteCompanyPerson(t *testing.T) {
	companyPersonHandler, companyRepository, personRepository := setupCompanyPersonHandler(t)

	company := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson := requests.AssociateCompanyPersonRequest{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}

	requestBytes, err := json.Marshal(companyPerson)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/company-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	companyPersonHandler.AssociateCompanyPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	deleteRequest := requests.DeleteCompanyPersonRequest{
		CompanyID: company.ID,
		PersonID:  person.ID,
	}

	requestBytes, err = json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err = http.NewRequest("POST", "/api/v1/company-person/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder = httptest.NewRecorder()
	companyPersonHandler.DeleteCompanyPerson(responseRecorder, request)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "", responseBodyString)
}

func TestDeleteCompanyPerson_ShouldReturnErrorIfNoMatchingCompanyPersonToDelete(t *testing.T) {
	companyPersonHandler, _, _ := setupCompanyPersonHandler(t)

	companyID, personID := uuid.New(), uuid.New()
	deleteRequest := requests.DeleteCompanyPersonRequest{
		CompanyID: companyID,
		PersonID:  personID,
	}

	requestBytes, err := json.Marshal(deleteRequest)
	assert.NoError(t, err)

	request, err := http.NewRequest("POST", "/api/v1/company-person/delete", bytes.NewReader(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	companyPersonHandler.DeleteCompanyPerson(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"error: object not found: CompanyPerson does not exist. companyID: "+
			companyID.String()+", personID: "+personID.String()+"\n",
		responseBodyString)
}

// -------- test helpers: --------

func setupTestData(
	t *testing.T,
	companyPersonHandler *handlers.CompanyPersonHandler,
	companyRepository *repositories.CompanyRepository,
	personRepository *repositories.PersonRepository,
	sleep bool) (
	uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {

	company1 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	company2 := repositoryhelpers.CreateCompany(t, companyRepository, nil, nil)
	person1 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)
	person2 := repositoryhelpers.CreatePerson(t, personRepository, nil, nil)

	companyPerson1 := requests.AssociateCompanyPersonRequest{
		CompanyID: company1.ID,
		PersonID:  person1.ID,
	}
	requestBytes, err := json.Marshal(companyPerson1)
	assert.NoError(t, err)
	request, err := http.NewRequest("POST", "/api/v1/company-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder := httptest.NewRecorder()
	companyPersonHandler.AssociateCompanyPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	companyPerson2 := requests.AssociateCompanyPersonRequest{
		CompanyID: company2.ID,
		PersonID:  person2.ID,
	}
	requestBytes, err = json.Marshal(companyPerson2)
	assert.NoError(t, err)
	request, err = http.NewRequest("POST", "/api/v1/company-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder = httptest.NewRecorder()
	companyPersonHandler.AssociateCompanyPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	if sleep {
		// a sleep is needed in order to ensure the order of the records.
		//There needs to be a minimum of 10 milliseconds between inserts.
		time.Sleep(10 * time.Millisecond)
	}

	companyPerson3 := requests.AssociateCompanyPersonRequest{
		CompanyID: company2.ID,
		PersonID:  person1.ID,
	}
	requestBytes, err = json.Marshal(companyPerson3)
	assert.NoError(t, err)
	request, err = http.NewRequest("POST", "/api/v1/company-person/associate", bytes.NewReader(requestBytes))
	assert.NoError(t, err)
	responseRecorder = httptest.NewRecorder()
	companyPersonHandler.AssociateCompanyPerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	return company1.ID, company2.ID, person1.ID, person2.ID
}
