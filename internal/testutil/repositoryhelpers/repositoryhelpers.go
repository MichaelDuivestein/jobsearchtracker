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
