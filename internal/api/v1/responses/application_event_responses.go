package responses

import (
	"jobsearchtracker/internal/models"
	"time"

	"github.com/google/uuid"
)

type ApplicationEventResponse struct {
	ApplicationID uuid.UUID `json:"application_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventID       uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
	CreatedDate   time.Time `json:"created_date" example:"2025-12-31T23:59Z" extensions:"x-order=2"`
}

func NewApplicationEventResponse(model *models.ApplicationEvent) *ApplicationEventResponse {
	if model == nil {
		return nil
	}

	response := &ApplicationEventResponse{
		ApplicationID: model.ApplicationID,
		EventID:       model.EventID,
		CreatedDate:   model.CreatedDate,
	}

	return response
}

func NewApplicationEventsResponse(models []*models.ApplicationEvent) []*ApplicationEventResponse {
	if len(models) == 0 {
		return []*ApplicationEventResponse{}
	}

	var responses = make([]*ApplicationEventResponse, len(models))
	for index := range models {
		response := NewApplicationEventResponse(models[index])
		responses[index] = response
	}

	return responses
}
