package responses

import (
	"jobsearchtracker/internal/models"
	"time"

	"github.com/google/uuid"
)

type CompanyPersonResponse struct {
	CompanyID   uuid.UUID `json:"company_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	PersonID    uuid.UUID `json:"person_id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=1"`
	CreatedDate time.Time `json:"created_date" example:"2025-12-31T23:59Z" extensions:"x-order=2"`
}

func NewCompanyPersonResponse(model *models.CompanyPerson) *CompanyPersonResponse {
	if model == nil {
		return nil
	}

	response := &CompanyPersonResponse{
		CompanyID:   model.CompanyID,
		PersonID:    model.PersonID,
		CreatedDate: model.CreatedDate,
	}

	return response
}

func NewCompanyPersonsResponse(models []*models.CompanyPerson) []*CompanyPersonResponse {
	if len(models) == 0 {
		return []*CompanyPersonResponse{}
	}

	var responses = make([]*CompanyPersonResponse, len(models))
	for index := range models {
		response := NewCompanyPersonResponse(models[index])
		responses[index] = response
	}

	return responses
}
