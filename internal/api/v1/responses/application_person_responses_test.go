package responses

import (
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewApplicationPersonResponse tests: --------

func TestNewApplicationPersonResponse_ShouldWork(t *testing.T) {
	model := models.ApplicationPerson{
		PersonID:      uuid.New(),
		ApplicationID: uuid.New(),
		CreatedDate:   time.Now().AddDate(1, 2, 3),
	}

	response := NewApplicationPersonResponse(&model)
	assert.NotNil(t, response)

	assert.Equal(t, response.ApplicationID, model.ApplicationID)
	assert.Equal(t, response.PersonID.String(), model.PersonID.String())
	assert.Equal(t, response.CreatedDate, model.CreatedDate)
}

func TestNewApplicationPersonResponse_ReturnNilIfModelIsNil(t *testing.T) {
	response := NewApplicationPersonResponse(nil)
	assert.Nil(t, response)
}

// -------- NewApplicationPersonsResponse tests: --------

func TestNewApplicationPersonsResponse_ShouldWork(t *testing.T) {
	ApplicationPersonModels := []*models.ApplicationPerson{
		{
			PersonID:      uuid.New(),
			ApplicationID: uuid.New(),
			CreatedDate:   time.Now().AddDate(1, 2, 3),
		},
		{
			PersonID:      uuid.New(),
			ApplicationID: uuid.New(),
			CreatedDate:   time.Now().AddDate(4, 5, 6),
		},
	}

	response := NewApplicationPersonsResponse(ApplicationPersonModels)
	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, response[0].ApplicationID, ApplicationPersonModels[0].ApplicationID)
	assert.Equal(t, response[0].PersonID, ApplicationPersonModels[0].PersonID)
	assert.Equal(t, response[0].CreatedDate, ApplicationPersonModels[0].CreatedDate)

	assert.Equal(t, response[1].ApplicationID, ApplicationPersonModels[1].ApplicationID)
	assert.Equal(t, response[1].PersonID, ApplicationPersonModels[1].PersonID)
	assert.Equal(t, response[1].CreatedDate, ApplicationPersonModels[1].CreatedDate)
}

func TestNewApplicationPersonsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	response := NewApplicationPersonsResponse([]*models.ApplicationPerson{})
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewApplicationPersonsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response := NewApplicationPersonsResponse(nil)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
