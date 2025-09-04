package responses

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewCompanyResponse tests: --------

func TestNewCompanyResponse_ShouldWork(t *testing.T) {
	notes := "some notes"
	lastContact := time.Now().AddDate(0, 0, -3)
	updatedDate := time.Now().AddDate(0, 0, -2)

	model := models.Company{
		ID:          uuid.New(),
		Name:        "Randomized Company",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       &notes,
		LastContact: &lastContact,
		CreatedDate: time.Now().AddDate(0, 0, -4),
		UpdatedDate: &updatedDate,
	}

	response, err := NewCompanyResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.CompanyType.String(), response.CompanyType.String())
	assert.Equal(t, model.Notes, response.Notes)
	assert.Equal(t, model.LastContact, response.LastContact)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Equal(t, model.UpdatedDate, response.UpdatedDate)
}

func TestNewCompanyResponse_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	model := models.Company{
		ID:          uuid.New(),
		Name:        "Yet another company name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: time.Now().AddDate(0, 0, 1),
	}

	response, err := NewCompanyResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.CompanyType.String(), response.CompanyType.String())
	assert.Nil(t, response.Notes)
	assert.Nil(t, response.LastContact)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Nil(t, response.UpdatedDate)
}

func TestNewCompanyResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewCompanyResponse(nil)
	assert.Nil(t, nilModel)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, "internal service error: Error building response: Company is nil", err.Error())
}

func TestNewCompanyResponse_ShouldReturnInternalServiceErrorIfCompanyTypeIsInvalid(t *testing.T) {
	emptyCompanyType := models.Company{
		ID:          uuid.New(),
		Name:        "Randomized Company",
		CompanyType: models.CompanyType(""),
		CreatedDate: time.Now().AddDate(0, 0, 3),
	}

	emptyResponse, err := NewCompanyResponse(&emptyCompanyType)
	assert.Nil(t, emptyResponse)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal CompanyType to external CompanyType: ''",
		err.Error())

	badDataModel := models.Company{
		ID:          uuid.New(),
		Name:        "Randomized Company",
		CompanyType: models.CompanyType("BadData"),
		CreatedDate: time.Now().AddDate(0, 0, 3),
	}

	badDataResponse, err := NewCompanyResponse(&badDataModel)
	assert.Nil(t, badDataResponse)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal CompanyType to external CompanyType: 'BadData'",
		err.Error())
}

// -------- NewCompaniesResponse tests: --------

func TestNewCompaniesResponse_ShouldWork(t *testing.T) {
	companyModels := []*models.Company{
		{
			ID:          uuid.New(),
			Name:        "CompanyOne",
			CompanyType: models.CompanyTypeConsultancy,
			CreatedDate: time.Now().AddDate(0, 0, -1),
		},
		{
			ID:          uuid.New(),
			Name:        "CompanyTwo",
			CompanyType: models.CompanyTypeEmployer,
			CreatedDate: time.Now().AddDate(0, 0, -2),
		},
	}

	companies, err := NewCompaniesResponse(companyModels)
	assert.NoError(t, err)
	assert.NotNil(t, companies)
	assert.Equal(t, len(companies), 2)
}

func TestNewCompaniesResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewCompaniesResponse(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response))
}

func TestNewCompaniesResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var companyModels []*models.Company
	response, err := NewCompaniesResponse(companyModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response))
}

func TestNewCompaniesResponse_ShouldReturnInternalServiceErrorIfOneCompanyTypeIsInvalid(t *testing.T) {
	companyModels := []*models.Company{
		{
			ID:          uuid.New(),
			Name:        "CompanyOne",
			CompanyType: models.CompanyTypeEmployer,
			CreatedDate: time.Now().AddDate(0, 0, 0),
		},
		{
			ID:          uuid.New(),
			Name:        "CompanyTwo",
			CompanyType: "",
			CreatedDate: time.Now().AddDate(0, 0, 0),
		},
	}

	companies, err := NewCompaniesResponse(companyModels)
	assert.Nil(t, companies)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal CompanyType to external CompanyType: ''",
		err.Error())
}
