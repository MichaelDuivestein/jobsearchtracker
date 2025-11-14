package responses

import (
	"jobsearchtracker/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewApplicationEventResponse tests: --------

func TestNewApplicationEventResponse_ShouldWork(t *testing.T) {
	model := models.ApplicationEvent{
		EventID:       uuid.New(),
		ApplicationID: uuid.New(),
		CreatedDate:   time.Now().AddDate(1, 2, 3),
	}

	response := NewApplicationEventResponse(&model)
	assert.NotNil(t, response)

	assert.Equal(t, response.ApplicationID, model.ApplicationID)
	assert.Equal(t, response.EventID.String(), model.EventID.String())
	assert.Equal(t, response.CreatedDate, model.CreatedDate)
}

func TestNewApplicationEventResponse_ReturnNilIfModelIsNil(t *testing.T) {
	response := NewApplicationEventResponse(nil)
	assert.Nil(t, response)
}

// -------- NewApplicationEventsResponse tests: --------

func TestNewApplicationEventsResponse_ShouldWork(t *testing.T) {
	ApplicationEventModels := []*models.ApplicationEvent{
		{
			EventID:       uuid.New(),
			ApplicationID: uuid.New(),
			CreatedDate:   time.Now().AddDate(1, 2, 3),
		},
		{
			EventID:       uuid.New(),
			ApplicationID: uuid.New(),
			CreatedDate:   time.Now().AddDate(4, 5, 6),
		},
	}

	response := NewApplicationEventsResponse(ApplicationEventModels)
	assert.NotNil(t, response)
	assert.Len(t, response, 2)

	assert.Equal(t, response[0].ApplicationID, ApplicationEventModels[0].ApplicationID)
	assert.Equal(t, response[0].EventID, ApplicationEventModels[0].EventID)
	assert.Equal(t, response[0].CreatedDate, ApplicationEventModels[0].CreatedDate)

	assert.Equal(t, response[1].ApplicationID, ApplicationEventModels[1].ApplicationID)
	assert.Equal(t, response[1].EventID, ApplicationEventModels[1].EventID)
	assert.Equal(t, response[1].CreatedDate, ApplicationEventModels[1].CreatedDate)
}

func TestNewApplicationEventsResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	response := NewApplicationEventsResponse([]*models.ApplicationEvent{})
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewApplicationEventsResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response := NewApplicationEventsResponse(nil)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
