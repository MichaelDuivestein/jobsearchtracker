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
			inputRequest:         testutil.StringPtr(""),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "body does not match CreateCompanyRequest",
			inputRequest:         testutil.StringPtr(`{"recruiter_name":"Mark Droog"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n"},
		{
			testName:             "body Name is missing",
			inputRequest:         testutil.StringPtr(`{"company_type":"recruiter"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n"},
		{
			testName:             "body CompanyType is missing",
			inputRequest:         testutil.StringPtr(`{"name":"random company name"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'CompanyType': CompanyType is invalid\n"},
		{
			testName:             "body CompanyType is invalid",
			inputRequest:         testutil.StringPtr(`{"name":"random company name","company_type":"other"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'CompanyType': CompanyType is invalid\n"},
		{
			testName:             "body is invalid",
			inputRequest:         testutil.StringPtr(`{"recruiter_name":"Mark Droog"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error on field 'Name': Name is empty\n"},
		{
			testName:             "malformed json",
			inputRequest:         testutil.StringPtr(`"name":"random company name","company_type":"consultancy"`),
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
