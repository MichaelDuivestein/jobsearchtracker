package repositories

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- GetById tests: --------

func TestGetById_ShouldReturnValidationErrorIfPersonIDIsNil(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	person, err := personRepository.GetById(nil)
	assert.Nil(t, person)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

// -------- GetAllByName tests: --------

func TestGetAllByName_ShouldReturnValidationErrorIfPersonNameIsNil(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	retrievedPersons, err := personRepository.GetAllByName(nil)
	assert.Nil(t, retrievedPersons)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'Name': Name is nil", validationError.Error())
}

// -------- Update tests: --------

func TestUpdate_ShouldReturnValidationErrorIfNoPersonFieldsToUpdate(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	personToUpdate := models.UpdatePerson{
		ID: uuid.New(),
	}

	err := personRepository.Update(&personToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfPersonIDIsNil(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	err := personRepository.Delete(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

// -------- buildCompaniesCoalesceAndJoin tests: --------

func TestBuildCompaniesCoalesceAndJoin_ShouldReturnEmptyStringsIfIncludeExtraDataTypeIsNone(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildCompaniesCoalesceAndJoin(models.IncludeExtraDataTypeNone)

	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestBuildCompaniesCoalesceAndJoin_ShouldReturnEmptyStringsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildCompaniesCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', c.id
				) ORDER BY c.created_date DESC
			) FILTER (WHERE c.id IS NOT NULL),
			JSON_ARRAY()
		) as companies`
	assert.Equal(t, expectedCoalesce, coalesce)

	expectedJoin := `
		LEFT JOIN company_person cp ON cp.person_id = p.id 
		LEFT JOIN company c ON c.id = cp.company_id `

	assert.Equal(t, expectedJoin, join)
}

func TestBuildCompaniesCoalesceAndJoin_ShouldReturnEmptyStringsIfIncludeExtraDataTypeIsAll(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildCompaniesCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', c.id, 
					'Name', c.name, 
					'CompanyType', c.company_type, 
					'Notes', c.notes, 
					'LastContact', c.last_contact, 
					'CreatedDate', c.created_date, 
					'UpdatedDate', c.updated_date 
				) ORDER BY c.created_date DESC
			) FILTER (WHERE c.id IS NOT NULL),
			JSON_ARRAY()
		) as companies`
	assert.Equal(t, expectedCoalesce, coalesce)

	expectedJoin := `
		LEFT JOIN company_person cp ON cp.person_id = p.id 
		LEFT JOIN company c ON c.id = cp.company_id `

	assert.Equal(t, expectedJoin, join)
}
