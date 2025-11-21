package responses

import (
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewCompanyEventResponse tests: --------

func TestNewCompanyEventResponse_ShouldWork(t *testing.T) {
	model := models.CompanyEvent{
		EventID:     uuid.New(),
		CompanyID:   uuid.New(),
		CreatedDate: time.Now().AddDate(1, 2, 3),
	}

	response := NewCompanyEventResponse(&model)
	assert.NotNil(t, response)

	assert.Equal(t, response.CompanyID, model.CompanyID)
	assert.Equal(t, response.EventID.String(), model.EventID.String())
	assert.Equal(t, response.CreatedDate, model.CreatedDate)
}

func TestNewCompanyEventResponse_ReturnNilIfModelIsNil(t *testing.T) {
	response := NewCompanyEventResponse(nil)
	assert.Nil(t, response)
}

// -------- NewCompanyEventsResponse tests: --------

func TestNewCompanyEventsResponse_ShouldWork(t *testing.T) {
	CompanyEventModels := []*models.CompanyEvent{
		{
			EventID:     uuid.New(),
			CompanyID:   uuid.New(),
			CreatedDate: time.Now().AddDate(1, 2, 3),
		},
		{
			EventID:     uuid.New(),
			CompanyID:   uuid.New(),
			CreatedDate: time.Now().AddDate(4, 5, 6),
		},
	}

	response := NewCompanyEventsResponse(CompanyEventModels)
	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, response[0].CompanyID, CompanyEventModels[0].CompanyID)
	assert.Equal(t, response[0].EventID, CompanyEventModels[0].EventID)
	assert.Equal(t, response[0].CreatedDate, CompanyEventModels[0].CreatedDate)

	assert.Equal(t, response[1].CompanyID, CompanyEventModels[1].CompanyID)
	assert.Equal(t, response[1].EventID, CompanyEventModels[1].EventID)
	assert.Equal(t, response[1].CreatedDate, CompanyEventModels[1].CreatedDate)
}

func TestNewCompanyEventsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	response := NewCompanyEventsResponse([]*models.CompanyEvent{})
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewCompanyEventsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response := NewCompanyEventsResponse(nil)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
