package handlers

import (
	"bytes"
	"jobsearchtracker/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- AssociateCompanyEvent tests: --------

func TestAssociateCompanyEvent_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			inputRequest:         testutil.ToPtr(`{"event_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: CompanyID is invalid\n"},
		{
			testName:             "company_id is empty",
			inputRequest:         testutil.ToPtr(`{"company_id": "", "event_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "company_id is invalid",
			inputRequest:         testutil.ToPtr(`{"company_id": "not valid", "event_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "event_id is missing",
			inputRequest:         testutil.ToPtr(`{"company_id": "8b802e50-f164-4d92-9f27-8cd91167f1e8"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "validation error: EventID is invalid\n"},
		{
			testName:             "event_id is empty",
			inputRequest:         testutil.ToPtr(`{"company_id": "06f92026-5b76-431a-909d-005ae920f4e4", "event_id": ""}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
		{
			testName:             "event_id is invalid",
			inputRequest:         testutil.ToPtr(`{"company_id": "06f92026-5b76-431a-909d-005ae920f4e4", "event_id": "not valid"}`),
			expectedResponseCode: http.StatusBadRequest,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n"},
	}
	handler := NewCompanyEventHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var requestBody []byte
			if test.inputRequest != nil {
				requestBody = []byte(*test.inputRequest)
			} else {
				requestBody = nil
			}

			request, err := http.NewRequest("POST", "/api/v1/company-event/associate", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.AssociateCompanyEvent(responseRecorder, request)
			assert.Equal(t, test.expectedResponseCode, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}

}

// -------- GetCompanyEventsByID tests: --------

func TestGetCompanyEventsByID_ShouldRespondWithBadRequestStatus(t *testing.T) {
	tests := []struct {
		testName             string
		queryParams          string
		expectedErrorMessage string
	}{
		{
			testName:             "nil companyID and nil eventID",
			queryParams:          "",
			expectedErrorMessage: "CompanyID and/or EventID are required\n",
		},
		{
			testName:             "empty companyID and empty eventID",
			queryParams:          `?company_id=&event_id=`,
			expectedErrorMessage: "CompanyID and/or EventID are required\n",
		},
		{
			testName:             "empty companyID and nil eventID",
			queryParams:          `?company_id=`,
			expectedErrorMessage: "CompanyID and/or EventID are required\n",
		},
		{
			testName:             "nil companyID and empty eventID",
			queryParams:          `?event_id=`,
			expectedErrorMessage: "CompanyID and/or EventID are required\n",
		},
		{
			testName:             "invalid companyID",
			queryParams:          `?company_id=not-valid&event_id=8b802e50-f164-4d92-9f27-8cd91167f1e8`,
			expectedErrorMessage: "CompanyID and/or EventID are required\n",
		},
		{
			testName:             "invalid eventID",
			queryParams:          `?company_id=06f92026-5b76-431a-909d-005ae920f4e4&event_id=not-valid`,
			expectedErrorMessage: "CompanyID and/or EventID are required\n",
		},
	}

	handler := NewCompanyEventHandler(nil)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			request, err := http.NewRequest(http.MethodGet, "/api/v1/company-event/get"+test.queryParams, nil)
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.GetCompanyEventsByID(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}

// --------DeleteCompanyEvent tests: --------

func TestDeleteCompanyEvent_ShouldRespondWithBadRequestStatus(t *testing.T) {
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
			testName:             "empty companyID and empty eventID",
			body:                 `{"company_id":"", "event_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "empty companyID and nil eventID",
			body:                 `"{company_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil companyID and empty eventID",
			body:                 `{"event_id":""}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "invalid companyID",
			body:                 `"company_id":"not valid","event_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}"`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil companyID",
			body:                 `{"event_id":"8b802e50-f164-4d92-9f27-8cd91167f1e8"}`,
			expectedErrorMessage: "validation error: CompanyID is invalid\n",
		},
		{
			testName:             "invalid eventID",
			body:                 `{"company_id":"06f92026-5b76-431a-909d-005ae920f4e4","event_id":"not valid"}`,
			expectedErrorMessage: "invalid request body: Unable to parse JSON\n",
		},
		{
			testName:             "nil eventID",
			body:                 `{"company_id":"06f92026-5b76-431a-909d-005ae920f4e4"}"`,
			expectedErrorMessage: "validation error: EventID is invalid\n",
		},
	}
	handler := NewCompanyEventHandler(nil)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {

			requestBody := []byte(test.body)

			request, err := http.NewRequest(http.MethodGet, "/api/v1/company-event/get", bytes.NewReader(requestBody))
			assert.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler.DeleteCompanyEvent(responseRecorder, request)
			assert.Equal(t, http.StatusBadRequest, responseRecorder.Code)

			responseBodyString := responseRecorder.Body.String()
			assert.Equal(t, test.expectedErrorMessage, responseBodyString)
		})
	}
}
