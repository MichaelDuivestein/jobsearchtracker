package responses

import (
	"jobsearchtracker/internal/api/v1/requests"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type PersonResponse struct {
	ID          uuid.UUID           `json:"id" extensions:"x-order=0" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	Name        string              `json:"name" extensions:"x-order=1" example:"CompanyName AB" extensions:"x-order=1"`
	PersonType  requests.PersonType `json:"person_type" extensions:"x-order=2" example:"internalRecruiter" extensions:"x-order=2"`
	Email       *string             `json:"email,omitempty" example:"name@domain.com" extensions:"x-order=3"`
	Phone       *string             `json:"phone,omitempty" example:"+46123456789" extensions:"x-order=4"`
	Notes       *string             `json:"notes,omitempty" example:"Notes go here" extensions:"x-order=5"`
	CreatedDate time.Time           `json:"created_date" extensions:"x-order=6" example:"2025-12-31T23:59Z"  extensions:"x-order=6"`
	UpdatedDate *time.Time          `json:"updated_date,omitempty" extensions:"x-order=7" example:"2025-12-31T23:59Z"  extensions:"x-order=7"`
}

// NewPersonResponse can return InternalServerError
func NewPersonResponse(personModel *models.Person) (*PersonResponse, error) {
	if personModel == nil {
		slog.Error("responses.NewPersonResponse: Person is nil")
		return nil, internalErrors.NewInternalServiceError("Error building response: Person is nil")
	}

	var personType *requests.PersonType = nil
	if personModel.PersonType != nil {
		nonNilPersonType, err := requests.NewPersonType(personModel.PersonType)
		if err != nil {
			return nil, err
		}
		personType = &nonNilPersonType
	}

	personResponse := PersonResponse{
		ID:          personModel.ID,
		Name:        personModel.Name,
		PersonType:  personType,
		Email:       personModel.Email,
		Phone:       personModel.Phone,
		Notes:       personModel.Notes,
		CreatedDate: personModel.CreatedDate,
		UpdatedDate: personModel.UpdatedDate,
	}

	return &personResponse, nil
}

func NewPersonsResponse(persons []*models.Person) ([]*PersonResponse, error) {
	if len(persons) == 0 {
		return []*PersonResponse{}, nil
	}

	var personResponses = make([]*PersonResponse, len(persons))
	for index := range persons {
		personResponse, err := NewPersonResponse(persons[index])
		if err != nil {
			return nil, err
		}
		personResponses[index] = personResponse
	}
	return personResponses, nil
}
