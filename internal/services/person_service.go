package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type PersonService struct {
	personRepository *repositories.PersonRepository
}

func NewPersonService(personRepository *repositories.PersonRepository) *PersonService {
	return &PersonService{personRepository: personRepository}
}

// CreatePerson can return ConflictError, InternalServiceError, ValidationError
func (personService *PersonService) CreatePerson(person *models.CreatePerson) (*models.Person, error) {
	if person == nil {
		slog.Error("person_service.CreatePerson: person is nil")
		return nil, internalErrors.NewValidationError(nil, "CreatePerson is nil")
	}

	err := person.Validate()
	if err != nil {
		var personID string
		if person.ID != nil {
			personID = person.ID.String()
		} else {
			personID = "[not set]"
		}
		slog.Info("person_service.CreatePerson: Person to create is invalid. ", "ID", personID, "error", err)
		return nil, err
	}

	if person.CreatedDate == nil {
		createdDate := time.Now()
		person.CreatedDate = &createdDate
	} else if person.CreatedDate.IsZero() {
		createdDate := time.Now()
		person.CreatedDate = &createdDate
		slog.Info(
			"person_service.createPerson: person.CreatedDate is zero. Setting to '" + person.CreatedDate.String() + "'")
	}

	insertedPerson, err := personService.personRepository.Create(person)
	if err != nil {
		return nil, err
	}

	slog.Info("person_service.createPerson: Inserted person.", "person.ID", insertedPerson.ID)
	return insertedPerson, nil
}

// GetPersonById can return ConflictError, InternalServiceError, NewValidationError
func (personService *PersonService) GetPersonById(personId *uuid.UUID) (*models.Person, error) {
	if personId == nil {
		personIdString := "person ID"
		err := internalErrors.NewValidationError(&personIdString, "personId is required")
		slog.Info("personService.GetPersonById: Failed to get person", "error", err)
		return nil, err
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	person, err := personService.personRepository.GetById(personId)
	if err != nil {
		return nil, err
	}

	slog.Info("PersonService.GetPersonById: Retrieved person.", "person.ID", person.ID.String())

	return person, nil
}

// GetPersonsByName can return InternalServiceError, NotFoundError, ValidationError
func (personService *PersonService) GetPersonsByName(personName *string) ([]*models.Person, error) {
	if personName == nil {
		personNameString := "personName"
		err := internalErrors.NewValidationError(&personNameString, "personName is required")
		slog.Info("personService.GetPersonByName: Failed to get person", "error", err)
		return nil, err
	}

	persons, err := personService.personRepository.GetAllByName(personName)
	if err != nil {
		return nil, err
	}

	if persons == nil {
		slog.Info("PersonService.GetAllPersons: Retrieved zero persons")
	} else {
		slog.Info("PersonService.GetAllPersons: Retrieved " + string(rune(len(persons))) + " persons")
	}

	return persons, nil
}

// GetAllPersons can return InternalServiceError
func (personService *PersonService) GetAllPersons() ([]*models.Person, error) {
	// can return InternalServiceError
	persons, err := personService.personRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if persons == nil {
		slog.Info("PersonService.GetAllPersons: Retrieved zero persons")
	} else {
		slog.Info("PersonService.GetAllPersons: Retrieved " + string(rune(len(persons))) + " persons")
	}

	return persons, nil
}
