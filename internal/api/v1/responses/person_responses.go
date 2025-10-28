package responses

import (
	"jobsearchtracker/internal/api/v1/requests"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type PersonDTO struct {
	ID          uuid.UUID            `json:"id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	Name        *string              `json:"name,omitempty" example:"CompanyName AB" extensions:"x-order=1"`
	PersonType  *requests.PersonType `json:"person_type,omitempty" example:"internalRecruiter" extensions:"x-order=2"`
	Email       *string              `json:"email,omitempty" example:"name@domain.com" extensions:"x-order=3"`
	Phone       *string              `json:"phone,omitempty" example:"+46123456789" extensions:"x-order=4"`
	Notes       *string              `json:"notes,omitempty" example:"Notes go here" extensions:"x-order=5"`
	CreatedDate *time.Time           `json:"created_date,omitempty" example:"2025-12-31T23:59Z"  extensions:"x-order=6"`
	UpdatedDate *time.Time           `json:"updated_date,omitempty" example:"2025-12-31T23:59Z"  extensions:"x-order=7"`
}

// NewPersonDTO can return InternalServerError
func NewPersonDTO(personModel *models.Person) (*PersonDTO, error) {
	if personModel == nil {
		slog.Error("responses.NewPersonDTO: Person is nil")
		return nil, internalErrors.NewInternalServiceError("Error building DTO: Person is nil")
	}

	var personType *requests.PersonType = nil
	if personModel.PersonType != nil {
		nonNilPersonType, err := requests.NewPersonType(personModel.PersonType)
		if err != nil {
			return nil, err
		}
		personType = &nonNilPersonType
	}

	personDTO := PersonDTO{
		ID:          personModel.ID,
		Name:        personModel.Name,
		PersonType:  personType,
		Email:       personModel.Email,
		Phone:       personModel.Phone,
		Notes:       personModel.Notes,
		CreatedDate: personModel.CreatedDate,
		UpdatedDate: personModel.UpdatedDate,
	}

	return &personDTO, nil
}

func NewPersonDTOs(persons []*models.Person) ([]*PersonDTO, error) {
	if len(persons) == 0 {
		return []*PersonDTO{}, nil
	}

	var personDTOs = make([]*PersonDTO, len(persons))
	for index := range persons {
		personDTO, err := NewPersonDTO(persons[index])
		if err != nil {
			return nil, err
		}
		personDTOs[index] = personDTO
	}
	return personDTOs, nil
}

type PersonResponse struct {
	PersonDTO
	Companies *[]*CompanyDTO `json:"companies" extensions:"x-order=8"`
}

// NewPersonResponse can return InternalServerError
func NewPersonResponse(personModel *models.Person) (*PersonResponse, error) {
	if personModel == nil {
		slog.Error("responses.NewPersonResponse: Person is nil")
		return nil, internalErrors.NewInternalServiceError("Error building response: Person is nil")
	}

	// can return InternalServerError
	personDTO, err := NewPersonDTO(personModel)
	if err != nil {
		return nil, err
	}

	var companies []*CompanyDTO
	if personModel.Companies != nil {
		// can return InternalServerError
		companies, err = NewCompanyDTOs(*personModel.Companies)
		if err != nil {
			return nil, err
		}
	}

	personResponse := PersonResponse{
		PersonDTO: *personDTO,
		Companies: &companies,
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
