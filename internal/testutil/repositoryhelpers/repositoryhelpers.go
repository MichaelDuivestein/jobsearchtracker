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

func AssociateApplicationEvent(
	t *testing.T,
	repository *repositories.ApplicationEventRepository,
	applicationID uuid.UUID,
	eventID uuid.UUID,
	createdDate *time.Time) *models.ApplicationEvent {

	model := models.AssociateApplicationEvent{
		ApplicationID: applicationID,
		EventID:       eventID,
		CreatedDate:   createdDate,
	}

	associatedApplicationEvent, err := repository.AssociateApplicationEvent(&model)
	assert.NoError(t, err)

	return associatedApplicationEvent
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

func AssociateCompanyEvent(
	t *testing.T,
	repository *repositories.CompanyEventRepository,
	companyID uuid.UUID,
	eventID uuid.UUID,
	createdDate *time.Time) *models.CompanyEvent {

	model := models.AssociateCompanyEvent{
		CompanyID:   companyID,
		EventID:     eventID,
		CreatedDate: createdDate,
	}

	associatedCompanyEvent, err := repository.AssociateCompanyEvent(&model)
	assert.NoError(t, err)

	return associatedCompanyEvent
}

func AssociateCompanyPerson(
	t *testing.T,
	repository *repositories.CompanyPersonRepository,
	companyID uuid.UUID,
	personID uuid.UUID,
	createdDate *time.Time) *models.CompanyPerson {

	model := models.AssociateCompanyPerson{
		CompanyID:   companyID,
		PersonID:    personID,
		CreatedDate: createdDate,
	}

	associatedCompanyPerson, err := repository.AssociateCompanyPerson(&model)
	assert.NoError(t, err)

	return associatedCompanyPerson
}

func AssociateEventPerson(
	t *testing.T,
	repository *repositories.EventPersonRepository,
	eventID uuid.UUID,
	personID uuid.UUID,
	createdDate *time.Time) *models.EventPerson {

	model := models.AssociateEventPerson{
		EventID:     eventID,
		PersonID:    personID,
		CreatedDate: createdDate,
	}

	associatedEventPerson, err := repository.AssociateEventPerson(&model)
	assert.NoError(t, err)

	return associatedEventPerson
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
