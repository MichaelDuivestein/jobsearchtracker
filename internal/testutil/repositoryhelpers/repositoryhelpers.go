package repositoryhelpers

import (
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func CreateApplication(
	t *testing.T,
	applicationRepository *repositories.ApplicationRepository,
	applicationID *uuid.UUID,
	companyID *uuid.UUID,
	recruiterID *uuid.UUID,
	createdDate *time.Time) *models.Application {

	application := models.CreateApplication{
		ID:               applicationID,
		CompanyID:        companyID,
		RecruiterID:      recruiterID,
		JobTitle:         testutil.ToPtr("JobTitle"),
		RemoteStatusType: models.RemoteStatusTypeHybrid,
		CreatedDate:      createdDate,
	}
	insertedApplication, err := applicationRepository.Create(&application)
	assert.NoError(t, err)

	return insertedApplication
}

func AssociateApplicationPerson(
	t *testing.T,
	repository *repositories.ApplicationPersonRepository,
	applicationID uuid.UUID,
	personID uuid.UUID,
	createdDate *time.Time) *models.ApplicationPerson {

	model := models.AssociateApplicationPerson{
		ApplicationID: applicationID,
		PersonID:      personID,
		CreatedDate:   createdDate,
	}

	associatedApplicationPerson, err := repository.AssociateApplicationPerson(&model)
	assert.NoError(t, err)

	return associatedApplicationPerson
}

func CreateCompany(
	t *testing.T,
	companyRepository *repositories.CompanyRepository,
	companyID *uuid.UUID,
	createdDate *time.Time) *models.Company {

	company := models.CreateCompany{
		ID:          companyID,
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: createdDate,
	}

	insertedCompany, err := companyRepository.Create(&company)
	assert.NoError(t, err)

	return insertedCompany
}

func CreateEvent(
	t *testing.T,
	repository *repositories.EventRepository,
	eventID *uuid.UUID,
	eventType *models.EventType,
	eventDate *time.Time) *models.Event {

	eventIDToUse := eventID
	if eventIDToUse == nil {
		eventIDToUse = testutil.ToPtr(uuid.New())
	}

	eventTypeToUse := eventType
	if eventTypeToUse == nil {
		var eventTypeApplied models.EventType = models.EventTypeApplied
		eventTypeToUse = testutil.ToPtr(eventTypeApplied)
	}

	eventDateToUse := eventDate
	if eventDateToUse == nil {
		eventDateToUse = &time.Time{}
	}

	event := models.CreateEvent{
		ID:        eventIDToUse,
		EventType: *eventTypeToUse,
		EventDate: *eventDateToUse,
	}
	insertedEvent, err := repository.Create(&event)
	assert.NoError(t, err)

	return insertedEvent
}

func CreatePerson(
	t *testing.T,
	repository *repositories.PersonRepository,
	personID *uuid.UUID,
	createdDate *time.Time) *models.Person {

	person := models.CreatePerson{
		ID:          personID,
		Name:        "PersonName",
		PersonType:  models.PersonTypeUnknown,
		CreatedDate: createdDate,
	}

	insertedPerson, err := repository.Create(&person)
	assert.NoError(t, err)

	return insertedPerson
}
