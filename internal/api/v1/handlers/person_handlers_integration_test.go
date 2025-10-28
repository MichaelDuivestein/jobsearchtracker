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
	*repositories.CompanyPersonRepository) {

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

	var companyPersonRepository *repositories.CompanyPersonRepository
	err = container.Invoke(func(repository *repositories.CompanyPersonRepository) {
		companyPersonRepository = repository
	})
	assert.NoError(t, err)

	return personHandler, companyRepository, companyPersonRepository
}

// -------- CreatePerson tests: --------

func TestCreatePerson_ShouldInsertAndReturnPerson(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

	id := uuid.New()
	email := "e@ma.il"
	phone := "456908"
	notes := "Notes appeared here"
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "random person name",
		PersonType: requests.PersonTypeHR,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
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

	insertedPersonCreatedDate := personResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, createdDateApproximation, insertedPersonCreatedDate)

	assert.Nil(t, personResponse.UpdatedDate)
}

func TestCreatePerson_ShouldReturnStatusConflictIfPersonIDIsDuplicate(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

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
	personHandler, _, _ := setupPersonHandler(t)

	// insert a person:

	id := uuid.New()
	email := "Email here"
	phone := "456908"
	notes := "Notes appeared here"
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "random person name",
		PersonType: requests.PersonTypeDeveloper,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	_, createdDateApproximation := insertPerson(t, personHandler, requestBody)

	// get the person:

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
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

	insertedPersonCreatedDate := response.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, insertedPersonCreatedDate)

	assert.Nil(t, response.UpdatedDate)
}

func TestGetPersonById_ShouldReturnNotFoundIfPersonDoesNotExist(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

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
	personHandler, _, _ := setupPersonHandler(t)

	// Insert a person:

	id := uuid.New()
	email := "Email here"
	phone := "456908"
	notes := "Notes appeared here"
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "PersonName",
		PersonType: requests.PersonTypeDeveloper,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, requestBody)

	// Get the person by full name:

	firstGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	assert.NoError(t, err)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "PersonName",
	}
	firstGetRequest = mux.SetURLVars(firstGetRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, firstGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var firstResponse []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&firstResponse)
	assert.NoError(t, err)
	assert.Len(t, firstResponse, 1)

	assert.Equal(t, *requestBody.ID, firstResponse[0].ID)
	assert.Equal(t, requestBody.Name, *firstResponse[0].Name)

	// Get the person by partial name:

	secondGetRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/name", nil)
	assert.NoError(t, err)
	responseRecorder = httptest.NewRecorder()

	vars = map[string]string{
		"name": "son",
	}
	secondGetRequest = mux.SetURLVars(secondGetRequest, vars)

	personHandler.GetPersonsByName(responseRecorder, secondGetRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString = responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var secondResponse []responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&secondResponse)
	assert.NoError(t, err)
	assert.Len(t, secondResponse, 1)

	assert.Equal(t, *requestBody.ID, secondResponse[0].ID)
	assert.Equal(t, requestBody.Name, *secondResponse[0].Name)
}

func TestGetPersonsByName_ShouldReturnPersons(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

	// Insert two persons:

	firstID := uuid.New()
	email := "Jane email"
	phone := "345345"
	notes := "Jane Notes"
	firstRequestBody := requests.CreatePersonRequest{
		ID:         &firstID,
		Name:       "Jane Smith",
		PersonType: requests.PersonTypeJobAdvertiser,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, firstRequestBody)

	secondID := uuid.New()
	email = "Sarah email"
	phone = "4567856654"
	notes = "Sara Notes"
	secondRequestBody := requests.CreatePersonRequest{
		ID:         &secondID,
		Name:       "Sarah Janesson",
		PersonType: requests.PersonTypeCEO,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
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
	personHandler, _, _ := setupPersonHandler(t)

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

// -------- GetAllPersons tests: --------

func TestGetAllPersons_ShouldReturnAllPersons(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

	// insert persons

	firstID := uuid.New()
	email := "Person1 Email"
	phone := "1111111"
	notes := "Person1 Notes"
	firstRequestBody := requests.CreatePersonRequest{
		ID:         &firstID,
		Name:       "Person1",
		PersonType: requests.PersonTypeCTO,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, firstRequestBody)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	secondID := uuid.New()
	email = "Person2 Email"
	phone = "222222"
	notes = "Person2 Notes"
	secondRequestBody := requests.CreatePersonRequest{
		ID:         &secondID,
		Name:       "Person2",
		PersonType: requests.PersonTypeInternalRecruiter,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, secondRequestBody)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	thirdID := uuid.New()
	email = "Person3 Email"
	phone = "333333"
	notes = "Person3 Notes"
	thirdRequestBody := requests.CreatePersonRequest{
		ID:         &thirdID,
		Name:       "Person3",
		PersonType: requests.PersonTypeJobAdvertiser,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, thirdRequestBody)

	// get all persons:

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

	assert.NotNil(t, response)
	assert.Len(t, response, 3)

	assert.Equal(t, *thirdRequestBody.ID, response[0].ID)
	assert.Equal(t, *secondRequestBody.ID, response[1].ID)
	assert.Equal(t, *firstRequestBody.ID, response[2].ID)
}

func TestGetAllPersons_ShouldReturnEmptyResponseIfNoPersonsInDatabase(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

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

func TestGetAll_ShouldReturnCompaniesIfIncludeCompaniesIsSetToAll(t *testing.T) {
	personHandler, companyRepository, companyPersonRepository := setupPersonHandler(t)

	// setup persons
	person1ID := uuid.New()
	person1 := requests.CreatePersonRequest{
		ID:         &person1ID,
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person2ID := uuid.New()
	person2 := requests.CreatePersonRequest{
		ID:         &person2ID,
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person3ID := uuid.New()
	person3 := requests.CreatePersonRequest{
		ID:         &person3ID,
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: company1ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person2ID,
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

	assert.Equal(t, person3ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, person2ID, response[1].ID)
	assert.Len(t, *response[1].Companies, 1)
	assert.Equal(t, company2ID, (*(response[1]).Companies)[0].ID)

	assert.Equal(t, person1ID, response[2].ID)
	assert.Len(t, *response[2].Companies, 2)

	person1Company1 := (*(response[2]).Companies)[0]
	assert.Equal(t, company1ID, person1Company1.ID)
	assert.Equal(t, company1.Name, *person1Company1.Name)
	assert.Equal(t, company1.CompanyType.String(), person1Company1.CompanyType.String())
	assert.Equal(t, company1.Notes, person1Company1.Notes)
	testutil.AssertEqualFormattedDateTimes(t, company1.LastContact, person1Company1.LastContact)
	testutil.AssertEqualFormattedDateTimes(t, company1.CreatedDate, person1Company1.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, company1.UpdatedDate, person1Company1.UpdatedDate)

	person1Company2 := (*(response[2]).Companies)[1]
	assert.Equal(t, company2ID, person1Company2.ID)
	assert.Equal(t, company2.Name, *person1Company2.Name)
	assert.Equal(t, company2.CompanyType.String(), person1Company2.CompanyType.String())
	assert.Nil(t, person1Company2.Notes)
	assert.Nil(t, person1Company2.LastContact)
	assert.NotNil(t, person1Company2.CreatedDate)
	assert.Nil(t, person1Company2.UpdatedDate)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToAllAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personHandler, companyRepository, _ := setupPersonHandler(t)

	// setup persons
	person1ID := uuid.New()
	person1 := requests.CreatePersonRequest{
		ID:         &person1ID,
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person2ID := uuid.New()
	person2 := requests.CreatePersonRequest{
		ID:         &person2ID,
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person3ID := uuid.New()
	person3 := requests.CreatePersonRequest{
		ID:         &person3ID,
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
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

	assert.Equal(t, person3ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, person2ID, response[1].ID)
	assert.Nil(t, response[1].Companies)

	assert.Equal(t, person1ID, response[2].ID)
	assert.Nil(t, response[2].Companies)
}

func TestGetAllPerson_ShouldReturnCompanyIDsIfIncludeCompaniesIsSetToIDs(t *testing.T) {
	personHandler, companyRepository, companyPersonRepository := setupPersonHandler(t)

	// setup persons
	person1ID := uuid.New()
	person1 := requests.CreatePersonRequest{
		ID:         &person1ID,
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person2ID := uuid.New()
	person2 := requests.CreatePersonRequest{
		ID:         &person2ID,
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person3ID := uuid.New()
	person3 := requests.CreatePersonRequest{
		ID:         &person3ID,
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: company1ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person2ID,
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

	assert.Equal(t, person3ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, person2ID, response[1].ID)
	assert.Len(t, *(response[1]).Companies, 1)
	assert.Equal(t, company2ID, (*(response[1]).Companies)[0].ID)

	assert.Equal(t, person1ID, response[2].ID)
	assert.Len(t, *response[2].Companies, 2)

	person1Company1 := (*(response[2]).Companies)[0]
	assert.Equal(t, company1ID, person1Company1.ID)
	assert.Nil(t, person1Company1.Name)
	assert.Nil(t, person1Company1.CompanyType)
	assert.Nil(t, person1Company1.Notes)
	assert.Nil(t, person1Company1.LastContact)
	assert.Nil(t, person1Company1.CreatedDate)
	assert.Nil(t, person1Company1.UpdatedDate)

	person1Company2 := (*(response[2]).Companies)[1]
	assert.Equal(t, company2ID, person1Company2.ID)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToIDsAndThereAreNoCompanyPersonsInRepository(t *testing.T) {
	personHandler, companyRepository, _ := setupPersonHandler(t)

	// setup persons
	person1ID := uuid.New()
	person1 := requests.CreatePersonRequest{
		ID:         &person1ID,
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person2ID := uuid.New()
	person2 := requests.CreatePersonRequest{
		ID:         &person2ID,
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person3ID := uuid.New()
	person3 := requests.CreatePersonRequest{
		ID:         &person3ID,
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
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

	assert.Equal(t, person3ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, person2ID, response[1].ID)
	assert.Nil(t, response[1].Companies)

	assert.Equal(t, person1ID, response[2].ID)
	assert.Nil(t, response[2].Companies)
}

func TestGetAllPerson_ShouldReturnNoCompaniesIfIncludeCompaniesIsSetToNone(t *testing.T) {
	personHandler, companyRepository, companyPersonRepository := setupPersonHandler(t)

	// setup persons
	person1ID := uuid.New()
	person1 := requests.CreatePersonRequest{
		ID:         &person1ID,
		Name:       "Person1",
		PersonType: models.PersonTypeDeveloper,
	}
	insertPerson(t, personHandler, person1)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person2ID := uuid.New()
	person2 := requests.CreatePersonRequest{
		ID:         &person2ID,
		Name:       "Person2",
		PersonType: models.PersonTypeCTO,
	}
	insertPerson(t, personHandler, person2)

	// a sleep is needed in order to ensure the order of the records.
	//There needs to be a minimum of 1 second between inserts.
	time.Sleep(1000 * time.Millisecond)

	person3ID := uuid.New()
	person3 := requests.CreatePersonRequest{
		ID:         &person3ID,
		Name:       "person3",
		PersonType: models.PersonTypeHR,
	}
	insertPerson(t, personHandler, person3)

	// add two companies

	company1ID := uuid.New()
	company1 := models.CreateCompany{
		ID:          &company1ID,
		Name:        "Company1Name",
		CompanyType: requests.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("Company1Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 5)),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, 4)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 3)),
	}
	_, err := companyRepository.Create(&company1)
	assert.NoError(t, err)

	company2ID := uuid.New()
	company2 := models.CreateCompany{
		ID:          &company2ID,
		Name:        "Company2Name",
		CompanyType: requests.CompanyTypeConsultancy,
	}
	_, err = companyRepository.Create(&company2)
	assert.NoError(t, err)

	// associate persons and companies

	Company1person1 := models.AssociateCompanyPerson{
		CompanyID: company1ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company1person1)
	assert.NoError(t, err)

	Company2person1 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person1ID,
	}
	_, err = companyPersonRepository.AssociateCompanyPerson(&Company2person1)
	assert.NoError(t, err)

	Company2person2 := models.AssociateCompanyPerson{
		CompanyID: company2ID,
		PersonID:  person2ID,
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

	assert.Equal(t, person3ID, response[0].ID)
	assert.Nil(t, response[0].Companies)

	assert.Equal(t, person2ID, response[1].ID)
	assert.Nil(t, response[1].Companies)

	assert.Equal(t, person1ID, response[2].ID)
	assert.Nil(t, response[2].Companies)
}

// -------- UpdatePerson tests: --------

func TestUpdatePerson_ShouldUpdatePerson(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

	// create a person

	id := uuid.New()
	email := "Person Email"
	phone := "2345345"
	notes := "Notes"
	createRequest := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "Person Name",
		PersonType: requests.PersonTypeUnknown,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	_, createdDateApproximation := insertPerson(t, personHandler, createRequest)

	// update the person

	updatedName := "updated person name"
	var updatedPersonType requests.PersonType = requests.PersonTypeJobContact
	updatedEmail := "updated email"
	updatedPhone := "46584566745"
	updatedNotes := "updated notes"

	updateBody := requests.UpdatePersonRequest{
		ID:         id,
		Name:       &updatedName,
		PersonType: &updatedPersonType,
		Email:      &updatedEmail,
		Phone:      &updatedPhone,
		Notes:      &updatedNotes,
	}

	requestBytes, err := json.Marshal(updateBody)
	assert.NoError(t, err)

	updateRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/update", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	updatedDateApproximation := time.Now().Format(time.RFC3339)
	personHandler.UpdatePerson(responseRecorder, updateRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	// get the person by ID

	getRequest, err := http.NewRequest(http.MethodGet, "/api/v1/person/get/id", nil)
	assert.NoError(t, err)

	vars := map[string]string{
		"id": id.String(),
	}
	getRequest = mux.SetURLVars(getRequest, vars)

	personHandler.GetPersonByID(responseRecorder, getRequest)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var getPersonResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&getPersonResponse)
	assert.NoError(t, err)

	assert.Equal(t, id, getPersonResponse.ID)
	assert.Equal(t, updatedName, *getPersonResponse.Name)
	assert.Equal(t, updatedPersonType.String(), getPersonResponse.PersonType.String())
	assert.Equal(t, updatedEmail, *getPersonResponse.Email)
	assert.Equal(t, updatedPhone, *getPersonResponse.Phone)
	assert.Equal(t, updatedNotes, *getPersonResponse.Notes)

	personResponseCreatedDate := getPersonResponse.CreatedDate.Format(time.RFC3339)
	assert.Equal(t, *createdDateApproximation, personResponseCreatedDate)

	personResponseUpdatedDate := getPersonResponse.UpdatedDate.Format(time.RFC3339)
	assert.Equal(t, updatedDateApproximation, personResponseUpdatedDate)
}

func TestUpdatePerson_ShouldReturnBadRequestIfNothingToUpdate(t *testing.T) {
	personHandler, _, _ := setupPersonHandler(t)

	// create a person
	id := uuid.New()
	email := "Person Email"
	phone := "2345345"
	notes := "Notes"
	createRequest := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "Person Name",
		PersonType: requests.PersonTypeUnknown,
		Email:      &email,
		Phone:      &phone,
		Notes:      &notes,
	}
	insertPerson(t, personHandler, createRequest)

	// update the person
	updateBody := requests.UpdatePersonRequest{
		ID: id,
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
	personHandler, _, _ := setupPersonHandler(t)

	// insert a person

	id := uuid.New()
	requestBody := requests.CreatePersonRequest{
		ID:         &id,
		Name:       "Person Name",
		PersonType: requests.PersonTypeDeveloper,
	}

	insertPerson(t, personHandler, requestBody)

	// delete the person

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
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
	personHandler, _, _ := setupPersonHandler(t)

	id := uuid.New()

	deleteRequest, err := http.NewRequest(http.MethodDelete, "/api/v1/person/delete/", nil)
	assert.NoError(t, err)

	deleteResponseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": id.String(),
	}
	deleteRequest = mux.SetURLVars(deleteRequest, vars)

	personHandler.DeletePerson(deleteResponseRecorder, deleteRequest)
	assert.Equal(t, http.StatusNotFound, deleteResponseRecorder.Code)
}

// -------- Test helpers: --------

func insertPerson(
	t *testing.T, personHandler *handlers.PersonHandler, requestBody requests.CreatePersonRequest) (
	*responses.PersonResponse, *string) {

	requestBytes, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	createRequest, err := http.NewRequest(http.MethodPost, "/api/v1/person/new", bytes.NewBuffer(requestBytes))
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	createdDateApproximation := time.Now().Format(time.RFC3339)
	personHandler.CreatePerson(responseRecorder, createRequest)
	assert.Equal(t, http.StatusCreated, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.NotEmpty(t, responseBodyString)

	var createPersonResponse responses.PersonResponse
	err = json.NewDecoder(responseRecorder.Body).Decode(&createPersonResponse)
	assert.NoError(t, err)

	return &createPersonResponse, &createdDateApproximation
}
