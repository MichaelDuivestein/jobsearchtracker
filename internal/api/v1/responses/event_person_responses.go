package responses

import (
	"jobsearchtracker/internal/models"
	"time"

	"github.com/google/uuid"
)

type EventPersonResponse struct {
	EventID     uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID    uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
	CreatedDate time.Time `json:"created_date" example:"2025-12-31T23:59Z" extensions:"x-order=2"`
}

func NewEventPersonResponse(model *models.EventPerson) *EventPersonResponse {
	if model == nil {
		return nil
	}

	response := &EventPersonResponse{
		EventID:     model.EventID,
		PersonID:    model.PersonID,
		CreatedDate: model.CreatedDate,
	}

	return response
}

func NewEventPersonsResponse(models []*models.EventPerson) []*EventPersonResponse {
	if len(models) == 0 {
		return []*EventPersonResponse{}
	}

	var responses = make([]*EventPersonResponse, len(models))
	for index := range models {
		response := NewEventPersonResponse(models[index])
		responses[index] = response
	}

	return responses
}
