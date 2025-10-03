package responses

import (
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewCompanyPersonResponse tests: --------

func TestNewCompanyPersonResponse_ShouldWork(t *testing.T) {
	model := models.CompanyPerson{
		PersonID:    uuid.New(),
		CompanyID:   uuid.New(),
		CreatedDate: time.Now().AddDate(1, 2, 3),
	}

	response := NewCompanyPersonResponse(&model)
	assert.NotNil(t, response)

	assert.Equal(t, response.CompanyID, model.CompanyID)
	assert.Equal(t, response.PersonID.String(), model.PersonID.String())
	assert.Equal(t, response.CreatedDate, model.CreatedDate)
}

func TestNewCompanyPersonResponse_ReturnNilIfModelIsNil(t *testing.T) {
	response := NewCompanyPersonResponse(nil)
	assert.Nil(t, response)
}

// -------- NewCompanyPersonsResponse tests: --------

func TestNewCompanyPersonsResponse_ShouldWork(t *testing.T) {
	companyPersonModels := []*models.CompanyPerson{
		{
			PersonID:    uuid.New(),
			CompanyID:   uuid.New(),
			CreatedDate: time.Now().AddDate(1, 2, 3),
		},
		{
			PersonID:    uuid.New(),
			CompanyID:   uuid.New(),
			CreatedDate: time.Now().AddDate(4, 5, 6),
		},
	}

	response := NewCompanyPersonsResponse(companyPersonModels)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(response))

	assert.Equal(t, response[0].CompanyID, companyPersonModels[0].CompanyID)
	assert.Equal(t, response[0].PersonID, companyPersonModels[0].PersonID)
	assert.Equal(t, response[0].CreatedDate, companyPersonModels[0].CreatedDate)

	assert.Equal(t, response[1].CompanyID, companyPersonModels[1].CompanyID)
	assert.Equal(t, response[1].PersonID, companyPersonModels[1].PersonID)
	assert.Equal(t, response[1].CreatedDate, companyPersonModels[1].CreatedDate)
}

func TestNewCompanyPersonsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	response := NewCompanyPersonsResponse([]*models.CompanyPerson{})
	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response))
}

func TestNewCompanyPersonsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response := NewCompanyPersonsResponse(nil)
	assert.NotNil(t, response)
	assert.Equal(t, 0, len(response))
}
