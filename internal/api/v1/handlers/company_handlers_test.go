package handlers_test

import (
	"bytes"
	v1 "jobsearchtracker/internal/api/v1/handlers"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// -------- CreateCompany tests: --------

func TestCreateCompany_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		inputRequest         *string
		expectedResponseCode int
		expectedErrorMessage string
	}{
		{
			testName:             "body is nil",
			inputRequest:         nil,
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "body is empty",
			inputRequest:         testutil.ToPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "body does not match CreateCompanyRequest",
			inputRequest:         testutil.ToPtr(`{"recruiter_name":"Mark Droog"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n"},
		{
			testName:             "body Name is missing",
			inputRequest:         testutil.ToPtr(`{"company_type":"recruiter"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n"},
		{
			testName:             "body CompanyType is missing",
			inputRequest:         testutil.ToPtr(`{"name":"random company name"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'CompanyType': CompanyType is invalid\n"},
		{
			testName:             "body CompanyType is invalid",
			inputRequest:         testutil.ToPtr(`{"name":"random company name","company_type":"other"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'CompanyType': CompanyType is invalid\n"},
		{
			testName:             "body is invalid",
			inputRequest:         testutil.ToPtr(`{"recruiter_name":"Mark Droog"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n"},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"name":"random company name","company_type":"consultancy"`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}
	for _, test := range tests {
		companyHandler := v1.NewCompanyHandler(nil)
		t.Run(test.testName, func(t *testing.T) {

			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/company/new", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			companyHandler.CreateCompany(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code, "CreateCompany returned wrong status code")

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString, "Unexpected response body")
		})
	}
}

// -------- GetCompanyById tests: --------

func TestGetCompanyById_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	companyHandler := v1.NewCompanyHandler(nil)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	companyHandler.GetCompanyById(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code, "GetCompanyById returned wrong status code")

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "company ID is empty\n", responseBodyString, "CreateCompany returned wrong error message in body")
}

func TestGetCompanyById_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	companyHandler := v1.NewCompanyHandler(nil)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/id", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "ID GOES HERE",
	}
	request = mux.SetURLVars(request, vars)

	companyHandler.GetCompanyById(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code, "GetCompanyById returned wrong status code")

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "company ID is not a valid UUID\n", responseBodyString, "CreateCompany returned wrong error message in body")
}

// -------- GetCompaniesByName tests: --------

func TestGetCompaniesByName_ShouldReturnErrorIfNameIsEmpty(t *testing.T) {
	companyHandler := v1.NewCompanyHandler(nil)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/name", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"name": "",
	}
	request = mux.SetURLVars(request, vars)

	companyHandler.GetCompaniesByName(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "company Name is empty\n", responseBodyString)
}

// -------- GetAllCompanies tests: --------
func TestGetAllCompanies_ShouldReturnErrorIfIncludeAllCompaniesIsInvalid(t *testing.T) {
	companyHandler := v1.NewCompanyHandler(nil)

	request, err := http.NewRequest(http.MethodGet, "/api/v1/company/get/all?include_applications=maybe", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	companyHandler.GetAllCompanies(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(
		t,
		"Invalid value for include_applications. Accepted params are 'all', 'ids', and 'none'\n",
		responseBodyString)
}

// -------- UpdateCompany tests: --------

func TestUpdateCompany_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		inputRequest         *string
		expectedResponseCode int
		expectedErrorMessage string
	}{
		{
			testName:             "body is nil",
			inputRequest:         nil,
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body is empty",
			inputRequest:         testutil.ToPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body does not match UpdateCompanyRequest",
			inputRequest:         testutil.ToPtr(`{"company_id": "8abb5944-761b-447c-8a77-11ba1108ff68", "notes": "Notes"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body ID is missing",
			inputRequest:         testutil.ToPtr(`{"company_type":"recruiter"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: ID is empty\n",
		},
		{
			testName:             "body CompanyType is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "name":"random company name","company_type":"other"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error on field 'CompanyType': CompanyType is invalid\n",
		},
		{
			testName:             "body is invalid",
			inputRequest:         testutil.ToPtr(`{"id": "8abb5944-761b-447c-8a77-11ba1108ff68", "recruiter_name":"Mark Droog"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: nothing to update\n",
		},
		{
			testName:             "malformed json",
			inputRequest:         testutil.ToPtr(`"name":"random company name","company_type":"consultancy"`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "body contains no fields to update",
			inputRequest:         testutil.ToPtr(`{"id":"8abb5944-761b-447c-8a77-11ba1108ff68"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "Unable to convert request to internal model: validation error: nothing to update\n",
		},
	}

	companyHandler := v1.NewCompanyHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest(http.MethodPost, "/api/v1/company/update", bytes.NewBuffer(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			companyHandler.UpdateCompany(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// -------- DeleteCompany tests: --------

func TestDeleteCompany_ShouldReturnErrorIfIdIsEmpty(t *testing.T) {
	companyHandler := v1.NewCompanyHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/company/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "",
	}
	request = mux.SetURLVars(request, vars)

	companyHandler.DeleteCompany(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "company ID is empty\n", responseBodyString)
}

func TestDeleteCompany_ShouldReturnErrorIfIdIsNotUUID(t *testing.T) {
	companyHandler := v1.NewCompanyHandler(nil)

	request, err := http.NewRequest(http.MethodDelete, "/api/v1/company/delete", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "Some text",
	}
	request = mux.SetURLVars(request, vars)

	companyHandler.DeleteCompany(responseRecorder, request)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

	responseBodyString := responseRecorder.Body.String()
	assert.Equal(t, "company ID is not a valid UUID\n", responseBodyString)
}
