package responses

import (
	"jobsearchtracker/internal/models"
	"time"

	"github.com/google/uuid"
)

type CompanyPersonResponse struct {
	CompanyID   uuid.UUID `json:"company_id"`
	PersonID    uuid.UUID `json:"person_id"`
	CreatedDate time.Time `json:"created_date"`
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
