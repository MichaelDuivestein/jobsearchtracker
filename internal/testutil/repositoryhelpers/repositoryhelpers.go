package repositoryhelpers

import (
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
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
