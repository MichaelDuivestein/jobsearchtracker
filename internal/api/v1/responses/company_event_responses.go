package responses

import (
	"jobsearchtracker/internal/models"
	"time"

	"github.com/google/uuid"
)

type CompanyEventResponse struct {
	CompanyID   uuid.UUID `json:"company_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	EventID     uuid.UUID `json:"event_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
	CreatedDate time.Time `json:"created_date" example:"2025-12-31T23:59Z" extensions:"x-order=2"`
}

func NewCompanyEventResponse(model *models.CompanyEvent) *CompanyEventResponse {
	if model == nil {
		return nil
	}

	response := &CompanyEventResponse{
		CompanyID:   model.CompanyID,
		EventID:     model.EventID,
		CreatedDate: model.CreatedDate,
	}

	return response
}

func NewCompanyEventsResponse(models []*models.CompanyEvent) []*CompanyEventResponse {
	if len(models) == 0 {
		return []*CompanyEventResponse{}
	}

	var responses = make([]*CompanyEventResponse, len(models))
	for index := range models {
		response := NewCompanyEventResponse(models[index])
		responses[index] = response
	}

	return responses
}
