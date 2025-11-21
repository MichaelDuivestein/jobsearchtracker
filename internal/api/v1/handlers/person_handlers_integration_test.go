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

func setupPersonHandler(t *testing.T) (
	*handlers.PersonHandler,
	*repositories.CompanyRepository,
	*repositories.EventRepository,
	*repositories.PersonRepository,
	*repositories.CompanyPersonRepository,
	*repositories.EventPersonRepository) {

	config := configPackage.Config{
		DatabaseMigrationsPath:               "../../../../migrations",
		IsDatabaseMigrationsPathAbsolutePath: false,
	}
	container := dependencyinjection.SetupPersonHandlerTestContainer(t, config)

	var personHandler *handlers.PersonHandler
	err := container.Invoke(func(personHand *handlers.PersonHandler) {
		personHandler = personHand
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

	var personRepository *repositories.PersonRepository
	err = container.Invoke(func(repository *repositories.PersonRepository) {
		personRepository = repository
	})
	assert.NoError(t, err)

	var companyPersonRepository *repositories.CompanyPersonRepository
	err = container.Invoke(func(repository *repositories.CompanyPersonRepository) {
		companyPersonRepository = repository
	})
	assert.NoError(t, err)

	var eventPersonRepository *repositories.EventPersonRepository
	err = container.Invoke(func(repository *repositories.EventPersonRepository) {
		eventPersonRepository = repository
	})
	assert.NoError(t, err)

	return personHandler,
		companyRepository,
		eventRepository,
		personRepository,
		companyPersonRepository,
		eventPersonRepository
}

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldInsertAndReturnPerson(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	requestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "random person name",
		PersonType: requests.PersonTypeHR,
		Email:      testutil.ToPtr("e@ma.il"),
		Phone:      testutil.ToPtr("456908"),
		Notes:      testutil.ToPtr("Notes appeared here"),
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	personHandler.CreatePerson(responseRecorder, request)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var personResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&personResponse)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, personResponse.ID)
	assert.Equal(t, requestBody.Name, *personResponse.Name)
	assert.Equal(t, requestBody.PersonType.String(), personResponse.PersonType.String())
	assert.Equal(t, requestBody.Email, personResponse.Email)
	assert.Equal(t, requestBody.Phone, personResponse.Phone)
	assert.Equal(t, requestBody.Notes, personResponse.Notes)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, personResponse.CreatedDate, time.Second)
	assert.Nil(t, personResponse.UpdatedDate)
}

func TestCreatePerson_ShouldReturnStatusConflictIfPersonIDIsDuplicate(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	personID := uuid.New()

	firstRequestBody := requests.CreatePersonRequest{
		ID:         &personID,
		Name:       "first Person Name",
		PersonType: requests.PersonTypeOther,
	}

	firstRequestBytes, err := json.Marshal(firstRequestBody)
	assert.NoError(t, err)

	firstRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(firstRequestBytes))
	assert.NoError(t, err)

	firstResponseRecorder := httptest.NewRecorder()

	personHandler.CreatePerson(firstResponseRecorder, firstRequest)
	assert.Equal(t, http.StatusCreated, firstResponseRecorder.Code)

	var firstPersonResponse responses.PersonResponse
	err = json.NewDecoder(firstResponseRecorder.Body).Decode(&firstPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, personID, firstPersonResponse.ID)

	secondRequestBody := requests.CreatePersonRequest{
		ID:         &personID,
		Name:       "second Person Name",
		PersonType: requests.PersonTypeCEO,
	}

	secondRequestBytes, err := json.Marshal(secondRequestBody)
	assert.NoError(t, err)

	secondRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(secondRequestBytes))
	assert.NoError(t, err)

	secondResponseRecorder := httptest.NewRecorder()

	personHandler.CreatePerson(secondResponseRecorder, secondRequest)
	assert.Equal(t, http.StatusConflict, secondResponseRecorder.Code)

	expectedError := "Conflict error on insert: ID already exists\n"
	assert.Equal(t, expectedError, secondResponseRecorder.Body.String())
}

// -------- GetPersonById tests: --------

func TestGetPersonById_ShouldReturnPerson(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	// insert a person:

	requestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "random person name",
		PersonType: requests.PersonTypeDeveloper,
		Email:      testutil.ToPtr("Email here"),
		Phone:      testutil.ToPtr("456908"),
		Notes:      testutil.ToPtr("Notes appeared here"),
	}
	_, createdDateApproximation := insertPerson(t, personHandler, requestBody)

	// get the person:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": requestBody.ID.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, *requestBody.ID, response.ID)
	assert.Equal(t, requestBody.Name, *response.Name)
	assert.Equal(t, requestBody.PersonType.String(), response.PersonType.String())
	assert.Equal(t, requestBody.Email, response.Email)
	assert.Equal(t, requestBody.Phone, response.Phone)
	assert.Equal(t, requestBody.Notes, response.Notes)
	testutil.AssertDateTimesWithinDelta(t, createdDateApproximation, response.CreatedDate, time.Second)
	assert.Nil(t, response.UpdatedDate)
}

func TestGetPersonById_ShouldReturnNotFoundIfPersonDoesNotExist(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": uuid.New().String(),
	}
	request = mux.SetURLVars(request, vars)

	personHandler.GetPersonByID(responseRecorder, request)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	firstResponseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, firstResponseBodyString, "Person not found\n")
}

// -------- GetPersonByName tests: --------

func TestGetPersonsByName_ShouldReturnPerson(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	// Insert a person:

	requestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "PersonName",
		PersonType: requests.PersonTypeDeveloper,
		Email:      testutil.ToPtr("Email here"),
		Phone:      testutil.ToPtr("456908"),
		Notes:      testutil.ToPtr("Notes appeared here"),
	}
	insertPerson(t, personHandler, requestBody)

	// Get the person by full name:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	assert.NoError(t, err)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "PersonName",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	createdDateApproximation := time.Now()
	personHandler.GetPersonsByName(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)

	assert.Equal(t, *requestBody.ID, response[0].ID)
	assert.Equal(t, requestBody.Name, *response[0].Name)
	assert.Equal(t, requestBody.PersonType.String(), response[0].PersonType.String())
	assert.Equal(t, requestBody.Email, response[0].Email)
	assert.Equal(t, requestBody.Phone, response[0].Phone)
	assert.Equal(t, requestBody.Notes, response[0].Notes)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, response[0].CreatedDate, time.Second)
	assert.Nil(t, response[0].UpdatedDate)
}

func TestGetPersonsByName_ShouldReturnPersons(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	// Insert two persons:

	firstRequestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Jane Smith",
		PersonType: requests.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, firstRequestBody)

	secondRequestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Sarah Janesson",
		PersonType: requests.PersonTypeCEO,
	}
	insertPerson(t, personHandler, secondRequestBody)

	// Get persons by name:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	assert.NoError(t, err)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "Jane",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	assert.Equal(t, *firstRequestBody.ID, response[0].ID)
	assert.Equal(t, firstRequestBody.Name, *response[0].Name)

	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, secondRequestBody.Name, *response[1].Name)
}

func TestGetPersonsByName_ShouldReturnNotFoundIfNoPersonsMatchingName(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "Steve",
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(t, "No people [partially] matching this name found\n", responseBodyString)
}

// -------- GetAllPersons - base tests: --------

func TestGetAllPersons_ShouldReturnAllPersons(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	// insert persons

	firstRequestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person1",
		PersonType: requests.PersonTypeCTO,
		Email:      testutil.ToPtr("Person1 Email"),
		Phone:      testutil.ToPtr("1111111"),
		Notes:      testutil.ToPtr("Person1 Notes"),
	}
	insertPerson(t, personHandler, firstRequestBody)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	secondRequestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person2",
		PersonType: requests.PersonTypeInternalRecruiter,
	}
	insertPerson(t, personHandler, secondRequestBody)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	thirdRequestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person3",
		PersonType: requests.PersonTypeJobAdvertiser,
	}
	insertPerson(t, personHandler, thirdRequestBody)

	// get all persons:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, *thirdRequestBody.ID, response[0].ID)
	assert.Equal(t, *secondRequestBody.ID, response[1].ID)

	assert.Equal(t, *firstRequestBody.ID, response[2].ID)
	assert.Equal(t, firstRequestBody.Name, *response[2].Name)
	assert.Equal(t, firstRequestBody.PersonType.String(), response[2].PersonType.String())
	assert.Equal(t, firstRequestBody.Email, response[2].Email)
	assert.Equal(t, firstRequestBody.Phone, response[2].Phone)
	assert.Equal(t, firstRequestBody.Notes, response[2].Notes)
	testutil.AssertDateTimesWithinDelta(t, &createdDateApproximation, response[2].CreatedDate, time.Second)
	assert.Nil(t, response[2].UpdatedDate)
}

func TestGetAllPersons_ShouldReturnEmptyResponseIfNoPersonsInDatabase(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Len(t, response, 0)
}

// -------- GetAllPersons - companies tests: --------

func TestGetAll_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	personHandler, companyRepository, _, _, companyPersonRepository, _ := setupPersonHandler(t)

	// setup persons
	person1 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person2 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person3 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: *company1.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_companies=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, *person3.ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, *person2.ID, response[1].ID)
	assert.Len(t, *response[1].Companies, 1)
	assert.Equal(t, *company2.ID, (*(response[1]).Companies)[0].ID)

	assert.Equal(t, *person1.ID, response[2].ID)
	assert.Len(t, *response[2].Companies, 2)

	person1Company1 := (*(response[2]).Companies)[0]
	assert.Equal(t, *company1.ID, person1Company1.ID)
	assert.Equal(t, company1.Name, *person1Company1.Name)
	assert.Equal(t, company1.CompanyType.String(), person1Company1.CompanyType.String())
	assert.Equal(t, company1.Notes, person1Company1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, company1.LastContact, person1Company1.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company1.CreatedDate, person1Company1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company1.UpdatedDate, person1Company1.UpdatedDate)

	assert.Equal(t, *company2.ID, (*(response[2]).Companies)[1].ID)
}

func TestGetAllPersons_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personHandler, companyRepository, _, _, _, _ := setupPersonHandler(t)

	// setup persons
	person1 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person2 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person3 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_companies=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, *person3.ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, *person2.ID, response[1].ID)
	assert.Nil(t, response[1].Companies)

	assert.Equal(t, *person1.ID, response[2].ID)
	assert.Nil(t, response[2].Companies)
}

func TestGetAllPersons_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	personHandler, companyRepository, _, _, companyPersonRepository, _ := setupPersonHandler(t)

	// setup persons
	person1 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person2 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person3 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: *company1.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_companies=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, *person3.ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, *person2.ID, response[1].ID)
	assert.Len(t, *(response[1]).Companies, 1)
	assert.Equal(t, *company2.ID, (*(response[1]).Companies)[0].ID)

	assert.Equal(t, *person1.ID, response[2].ID)
	assert.Len(t, *response[2].Companies, 2)

	person1Company1 := (*(response[2]).Companies)[0]
	assert.Equal(t, *company1.ID, person1Company1.ID)
	assert.Nil(t, person1Company1.Name)
	assert.Nil(t, person1Company1.CompanyType)
	assert.Nil(t, person1Company1.Notes)
	assert.Nil(t, person1Company1.LastContact)
	assert.Nil(t, person1Company1.CreatedDate)
	assert.Nil(t, person1Company1.UpdatedDate)

	person1Company2 := (*(response[2]).Companies)[1]
	assert.Equal(t, *company2.ID, person1Company2.ID)
}

func TestGetAllPersons_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personHandler, companyRepository, _, _, _, _ := setupPersonHandler(t)

	// setup persons
	person1 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person2 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person3 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_companies=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, *person3.ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, *person2.ID, response[1].ID)
	assert.Nil(t, response[1].Companies)

	assert.Equal(t, *person1.ID, response[2].ID)
	assert.Nil(t, response[2].Companies)
}

func TestGetAllPersons_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	personHandler, companyRepository, _, _, companyPersonRepository, _ := setupPersonHandler(t)

	// setup persons
	person1 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person2 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 10 milliseconds between inserts.
	time.Sleep(10 * time.Millisecond)

	person3 := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2 := models.CreateCompany{
		ID:          testutil.ToPtr(uuid.New()),
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: *company1.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person1.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: *company2.ID,
		PersonID:  *person2.ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person2)
	assert.NoError(t, err)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_companies=none", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, *person3.ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, *person2.ID, response[1].ID)
	assert.Nil(t, response[1].Companies)

	assert.Equal(t, *person1.ID, response[2].ID)
	assert.Nil(t, response[2].Companies)
}

// -------- GetAllPersons - events tests: --------

func TestGetAllPersons_ShouldReturnEventsIfIncludeEventsIsSetToAll(t *testing.T) {
	personHandler, _, eventRepository, personRepository, _, eventPersonRepository := setupPersonHandler(t)

	// create person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add two events and associate them to the person

	event1ToInsert := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   time.Now().AddDate(0, 0, 5),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := eventRepository.Create(&event1ToInsert)
	assert.NoError(t, err)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 6))).ID

	// associate persons and events

	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, *event1ToInsert.ID, personID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event2ID, personID, nil)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_events=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	retrievedPerson := response[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.NotNil(t, retrievedPerson.Events)
	assert.Len(t, *retrievedPerson.Events, 2)

	assert.Equal(t, event2ID, *(*retrievedPerson.Events)[0].ID)

	event2 := (*retrievedPerson.Events)[1]
	assert.Equal(t, event1ToInsert.ID, event2.ID)
	assert.Equal(t, event1ToInsert.EventType.String(), event2.EventType.String())
	assert.Equal(t, event1ToInsert.Description, event2.Description)
	assert.Equal(t, event1ToInsert.Notes, event2.Notes)
	testutil.AssertEqualFormattedDateTimes(t, &event1ToInsert.EventDate, event2.EventDate)
	testutil.AssertEqualFormattedDateTimes(t, event1ToInsert.CreatedDate, event2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, event1ToInsert.UpdatedDate, event2.UpdatedDate)
}

func TestGetAllPersons_ShouldReturnNoEventsIfIncludeEventsIsSetToAllAndThereAreNoEventPersonsInRepository(t *testing.T) {
	personHandler, _, _, personRepository, _, _ := setupPersonHandler(t)

	// create person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_events=all", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	retrievedPerson := response[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.Nil(t, retrievedPerson.Events)
}

func TestGetAllPersons_ShouldReturnEventIDsIfIncludeEventsIsSetToIDs(t *testing.T) {
	personHandler, _, eventRepository, personRepository, _, eventPersonRepository := setupPersonHandler(t)

	// create person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add two events and associate them to the person

	event1ToInsert := models.CreateEvent{
		ID:          testutil.ToPtr(uuid.New()),
		EventType:   models.EventTypeApplied,
		Description: testutil.ToPtr("Event1Description"),
		Notes:       testutil.ToPtr("Event1Notes"),
		EventDate:   time.Now().AddDate(0, 0, 5),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := eventRepository.Create(&event1ToInsert)
	assert.NoError(t, err)

	event2ID := repositoryhelpers.CreateEvent(
		t,
		eventRepository,
		nil,
		nil,
		testutil.ToPtr(time.Now().AddDate(0, 0, 6))).ID

	// associate persons and events

	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, *event1ToInsert.ID, personID, nil)
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, event2ID, personID, nil)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_events=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	retrievedPerson := response[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.NotNil(t, retrievedPerson.Events)
	assert.Len(t, *retrievedPerson.Events, 2)

	assert.Equal(t, event2ID, *(*retrievedPerson.Events)[0].ID)

	event2 := (*retrievedPerson.Events)[1]
	assert.Equal(t, event1ToInsert.ID, event2.ID)
	assert.Nil(t, event2.EventType)
	assert.Nil(t, event2.Description)
	assert.Nil(t, event2.Notes)
	assert.Nil(t, event2.EventDate)
	assert.Nil(t, event2.CreatedDate)
	assert.Nil(t, event2.UpdatedDate)
}

func TestGetAllPersons_ShouldReturnNoEventsIfIncludeEventsIsSetToIDsAndThereAreNoEventPersonsInRepository(t *testing.T) {
	personHandler, _, _, personRepository, _, _ := setupPersonHandler(t)

	// create person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_events=ids", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	retrievedPerson := response[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.Nil(t, retrievedPerson.Events)
}

func TestGetAllPersons_ShouldReturnNoEventsIfIncludeEventsIsSetToNone(t *testing.T) {
	personHandler, _, eventRepository, personRepository, _, eventPersonRepository := setupPersonHandler(t)

	// create person

	personID := repositoryhelpers.CreatePerson(t, personRepository, nil, nil).ID

	// add an event and associate it to the person

	eventID := repositoryhelpers.CreateEvent(t, eventRepository, nil, nil, nil).ID
	repositoryhelpers.AssociateEventPerson(t, eventPersonRepository, eventID, personID, nil)

	// get all persons

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/all?include_events=none", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.GetAllPersons(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var response []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.NotNil(t, response)
	assert.Len(t, response, 1)

	retrievedPerson := response[0]
	assert.Equal(t, personID, retrievedPerson.ID)
	assert.Nil(t, retrievedPerson.Events)
}

// -------- UpdatePerson tests: --------

func TestUpdatePerson_ShouldUpdatePerson(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	// create a person

	createRequest := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person Name",
		PersonType: requests.PersonTypeUnknown,
		Email:      testutil.ToPtr("Person Email"),
		Phone:      testutil.ToPtr("2345345"),
		Notes:      testutil.ToPtr("Notes"),
	}
	_, createdDateApproximation := insertPerson(t, personHandler, createRequest)

	// update the person

	var updatedPersonType requests.PersonType = requests.PersonTypeJobContact
	updateBody := requests.UpdatePersonRequest{
		ID:         *createRequest.ID,
		Name:       testutil.ToPtr("updated person name"),
		PersonType: &updatedPersonType,
		Email:      testutil.ToPtr("updated email"),
		Phone:      testutil.ToPtr("46584566745"),
		Notes:      testutil.ToPtr("updated notes"),
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now()
	personHandler.UpdatePerson(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the person by ID

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": createRequest.ID.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var getPersonResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&getPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, updateBody.ID, getPersonResponse.ID)
	assert.Equal(t, updateBody.Name, getPersonResponse.Name)
	assert.Equal(t, updateBody.PersonType.String(), getPersonResponse.PersonType.String())
	assert.Equal(t, updateBody.Email, getPersonResponse.Email)
	assert.Equal(t, updateBody.Phone, getPersonResponse.Phone)
	assert.Equal(t, updateBody.Notes, getPersonResponse.Notes)
	testutil.AssertDateTimesWithinDelta(t, createdDateApproximation, getPersonResponse.CreatedDate, time.Second)
	testutil.AssertDateTimesWithinDelta(t, &updatedDateApproximation, getPersonResponse.UpdatedDate, time.Second)
}

func TestUpdatePerson_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	// create a person
	createRequest := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person Name",
		PersonType: requests.PersonTypeUnknown,
	}
	insertPerson(t, personHandler, createRequest)

	// update the person
	updateBody := requests.UpdatePersonRequest{
		ID: *createRequest.ID,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	personHandler.UpdatePerson(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)
	assert.Equal(
		t,
		"Unable to convert request to internal model: validation error: nothing to update\n",
		responseBodyString)
}

// -------- DeletePerson tests: --------

func TestDeletePerson_ShouldDeletePerson(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	// insert a person

	requestBody := requests.CreatePersonRequest{
		ID:         testutil.ToPtr(uuid.New()),
		Name:       "Person Name",
		PersonType: requests.PersonTypeDeveloper,
	}

	insertPerson(t, personHandler, requestBody)

	// delete the person

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": requestBody.ID.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	personHandler.DeletePerson(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusOK, deleteResponseRecorder.Code)

	// try to get the person
	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	getResponseRecorder := httptest.NewRecorder()
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonByID(getResponseRecorder, getRequest)
	assert.Equal(t, http.StatusNotFound, getResponseRecorder.Code, "GetPersonByID returned wrong status code")
}

func TestDeletePerson_ShouldReturnStatusNotFoundIfPersonDoesNotExist(t *testing.T) {
	personHandler, _, _, _, _, _ := setupPersonHandler(t)

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": uuid.New().String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	personHandler.DeletePerson(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusNotFound, deleteResponseRecorder.Code)
}

// -------- Test helpers: --------

func insertPerson(
	t *testing.T, personHandler *handlers.PersonHandler, requestBody requests.CreatePersonRequest) (
	*responses.PersonResponse, *time.Time) {

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	createRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now()
	personHandler.CreatePerson(responseRecorder, createRequest)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var createPersonResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&createPersonResponse)
	assert.NoError(t, err)

	return &createPersonResponse, &createdDateApproximation
}
