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
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	PersonType  requests.PersonType `json:"person_type"`
	Email       *string             `json:"email"`
	Phone       *string             `json:"phone"`
	Notes       *string             `json:"notes"`
	CreatedDate time.Time           `json:"created_date"`
	UpdatedDate *time.Time          `json:"updated_date"`
}

// NewPersonResponse can return InternalServerError
func NewPersonResponse(personModel *models.Person) (*PersonResponse, error) {
	if personModel == nil {
		slog.Error("responses.NewPersonResponse: Person is nil")
		return nil, internalErrors.NewInternalServiceError("Error building response: Person is nil")
	}

	personType, err := requests.NewPersonType(&personModel.PersonType)
	if err != nil {
		return nil, err
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
	if persons == nil || len(persons) == 0 {
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
