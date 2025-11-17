package responses

import (
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewEventPersonResponse tests: --------

func TestNewEventPersonResponse_ShouldWork(t *testing.T) {
	model := models.EventPerson{
		PersonID:    uuid.New(),
		EventID:     uuid.New(),
		CreatedDate: time.Now().AddDate(1, 2, 3),
	}

	response := NewEventPersonResponse(&model)
	assert.NotNil(t, response)

	assert.Equal(t, response.EventID, model.EventID)
	assert.Equal(t, response.PersonID.String(), model.PersonID.String())
	assert.Equal(t, response.CreatedDate, model.CreatedDate)
}

func TestNewEventPersonResponse_ReturnNilIfModelIsNil(t *testing.T) {
	response := NewEventPersonResponse(nil)
	assert.Nil(t, response)
}

// -------- NewEventPersonsResponse tests: --------

func TestNewEventPersonsResponse_ShouldWork(t *testing.T) {
	eventPersonModels := []*models.EventPerson{
		{
			PersonID:    uuid.New(),
			EventID:     uuid.New(),
			CreatedDate: time.Now().AddDate(1, 2, 3),
		},
		{
			PersonID:    uuid.New(),
			EventID:     uuid.New(),
			CreatedDate: time.Now().AddDate(4, 5, 6),
		},
	}

	response := NewEventPersonsResponse(eventPersonModels)
	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, response[0].EventID, eventPersonModels[0].EventID)
	assert.Equal(t, response[0].PersonID, eventPersonModels[0].PersonID)
	assert.Equal(t, response[0].CreatedDate, eventPersonModels[0].CreatedDate)

	assert.Equal(t, response[1].EventID, eventPersonModels[1].EventID)
	assert.Equal(t, response[1].PersonID, eventPersonModels[1].PersonID)
	assert.Equal(t, response[1].CreatedDate, eventPersonModels[1].CreatedDate)
}

func TestNewEventPersonsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	response := NewEventPersonsResponse([]*models.EventPerson{})
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewEventPersonsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response := NewEventPersonsResponse(nil)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
