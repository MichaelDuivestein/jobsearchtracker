package handlers

import (
	"bytes"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- AssociateCompanyPerson tests: --------

func TestAssociateCompanyPerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "company_id is missing",
			inputRequest:         testutil.ToPtr(`{"person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: CompanyID is invalid\n"},
		{
			testName:             "company_id is empty",
			inputRequest:         testutil.ToPtr(`{"company_id": "", "person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "company_id is invalid",
			inputRequest:         testutil.ToPtr(`{"company_id": "not valid", "person_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "person_id is missing",
			inputRequest:         testutil.ToPtr(`{"company_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: PersonID is invalid\n"},
		{
			testName:             "person_id is empty",
			inputRequest:         testutil.ToPtr(`{"company_id": "06f92026-5b76-431a-909d-005ae920f4e4", "person_id": ""}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "person_id is invalid",
			inputRequest:         testutil.ToPtr(`{"company_id": "06f92026-5b76-431a-909d-005ae920f4e4", "person_id": "not valid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}
	handler := NewCompanyPersonHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest("POST", "/api/v1/company-person/associate", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.AssociateCompanyPerson(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}

}

// -------- GetCompanyPersonsByID tests: --------

func TestGetCompanyPersonsByID_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		queryParams          string
		expectedErrorMessage string
	}{
		{
			testName:             "nil companyID and nil personID",
			queryParams:          "",
			expectedErrorMessage: "CompanyID and/or PersonID are required\n",
		},
		{
			testName:             "empty companyID and empty personID",
			queryParams:          `?company_id=&person_id=`,
			expectedErrorMessage: "CompanyID and/or PersonID are required\n",
		},
		{
			testName:             "empty companyID and nil personID",
			queryParams:          `?company_id=`,
			expectedErrorMessage: "CompanyID and/or PersonID are required\n",
		},
		{
			testName:             "nil companyID and empty personID",
			queryParams:          `?person_id=`,
			expectedErrorMessage: "CompanyID and/or PersonID are required\n",
		},
		{
			testName:             "invalid companyID",
			queryParams:          `?company_id=not-valid&person_id=8b802e50-f164-4d92-9f27-8cd91167f1e8`,
			expectedErrorMessage: "CompanyID and/or PersonID are required\n",
		},
		{
			testName:             "invalid personID",
			queryParams:          `?company_id=06f92026-5b76-431a-909d-005ae920f4e4&person_id=not-valid`,
			expectedErrorMessage: "CompanyID and/or PersonID are required\n",
		},
	}

	handler := NewCompanyPersonHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			request, err := http.NewRequest(http.MethodGet, "/api/v1/company-person/get"+test.queryParams, nil)
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.GetCompanyPersonsByID(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// --------DeleteCompanyPerson tests: --------

func TestDeleteCompanyPerson_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		body                 string
		expectedResponseCode int
		expectedErrorMessage string
	}{
		{
			testName:             "empty body",
			body:                 "",
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "empty companyID and empty personID",
			body:                 `{"company_id":"", "person_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "empty companyID and nil personID",
			body:                 `"{company_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil companyID and empty personID",
			body:                 `{"person_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "invalid companyID",
			body:                 `"company_id":"not valid","person_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}"`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil companyID",
			body:                 `{"person_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}`,
			expectedErrorMessage: "validation error: CompanyID is invalid\n",
		},
		{
			testName:             "invalid personID",
			body:                 `{"company_id":"06f92026-5b76-431a-909d-005ae920f4e4","person_id":"not valid"}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil personID",
			body:                 `{"company_id":"06f92026-5b76-431a-909d-005ae920f4e4"}"`,
			expectedErrorMessage: "validation error: PersonID is invalid\n",
		},
	}
	handler := NewCompanyPersonHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			requestBody := []byte(test.body)

			request, err := http.NewRequest(http.MethodGet, "/api/v1/company-person/get", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.DeleteCompanyPerson(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}
