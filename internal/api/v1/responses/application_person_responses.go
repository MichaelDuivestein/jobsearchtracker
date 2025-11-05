package responses

import (
	"jobsearchtracker/internal/models"
	"time"

	"github.com/google/uuid"
)

type ApplicationPersonResponse struct {
	ApplicationID uuid.UUID `json:"application_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID      uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
	CreatedDate   time.Time `json:"created_date" example:"2025-12-31T23:59Z" extensions:"x-order=2"`
}

func NewApplicationPersonResponse(model *models.ApplicationPerson) *ApplicationPersonResponse {
	if model == nil {
		return nil
	}

	response := &ApplicationPersonResponse{
		ApplicationID: model.ApplicationID,
		PersonID:      model.PersonID,
		CreatedDate:   model.CreatedDate,
	}

	return response
}

func NewApplicationPersonsResponse(models []*models.ApplicationPerson) []*ApplicationPersonResponse {
	if len(models) == 0 {
		return []*ApplicationPersonResponse{}
	}

	var responses = make([]*ApplicationPersonResponse, len(models))
	for index := range models {
		response := NewApplicationPersonResponse(models[index])
		responses[index] = response
	}

	return responses
}
