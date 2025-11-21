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

// -------- buildApplicationsCoalesceAndJoin tests: --------

func TestPersonRepositoryBuildApplicationsCoalesceAndJoin_ShouldReturnNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestPersonRepositoryBuildApplicationsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	expectedJoin := `
		LEFT JOIN application_person ap ON ap.person_id = p.id 
		LEFT JOIN application a ON a.id = ap.application_id `
	assert.Equal(t, expectedJoin, join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', a.id
				) ORDER BY a.created_date DESC
			) FILTER (WHERE a.id IS NOT NULL),
			JSON_ARRAY()
		) as applications
		`
	assert.Equal(t, expectedCoalesce, coalesce)
}

func TestPersonRepositoryBuildApplicationsCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	expectedJoin := `
		LEFT JOIN application_person ap ON ap.person_id = p.id 
		LEFT JOIN application a ON a.id = ap.application_id `
	assert.Equal(t, expectedJoin, join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', a.id,
					'CompanyID', a.company_id,
					'RecruiterID', a.recruiter_id,
					'JobTitle', a.job_title,
					'JobAdURL', a.job_ad_url,
					'Country', a.country,
					'Area', a.area,
					'RemoteStatusType', a.remote_status_type,
					'WeekdaysInOffice', a.weekdays_in_office,
					'EstimatedCycleTime', a.estimated_cycle_time,
					'EstimatedCommuteTime', a.estimated_commute_time,
					'ApplicationDate', a.application_date,
					'CreatedDate', a.created_date,
					'UpdatedDate', a.updated_date
				) ORDER BY a.created_date DESC
			) FILTER (WHERE a.id IS NOT NULL),
			JSON_ARRAY()
		) as applications
		`

	assert.Equal(t, expectedCoalesce, coalesce)
}

// -------- buildCompaniesCoalesceAndJoin tests: --------

func TestPersonRepositoryBuildCompaniesCoalesceAndJoin_ShouldReturnNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildCompaniesCoalesceAndJoin(models.IncludeExtraDataTypeNone)

	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestPersonRepositoryBuildCompaniesCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
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

func TestPersonRepositoryBuildCompaniesCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
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

// -------- buildEventsCoalesceAndJoin tests: --------

func TestPersonRepositoryBuildEventsCoalesceAndJoin_ShouldReturnNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeNone)

	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestPersonRepositoryBuildEventsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', e.id
				) ORDER BY e.event_date DESC
			) FILTER (WHERE e.id IS NOT NULL),
			JSON_ARRAY()
		) as events`
	assert.Equal(t, expectedCoalesce, coalesce)

	expectedJoin := `
		LEFT JOIN event_person ep ON ep.person_id = p.id 
		LEFT JOIN event e ON e.id = ep.event_id `

	assert.Equal(t, expectedJoin, join)
}

func TestPersonRepositoryBuildEventsCoalesceAndJoin_ShouldBuildWithAllColumnsIfIncludeExtraDataTypeIsAll(t *testing.T) {
	personRepository := NewPersonRepository(nil)

	coalesce, join := personRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', e.id, 
					'EventType', e.event_type, 
					'Description', e.description,  
					'Notes', e.notes, 
					'EventDate', e.event_date, 
					'CreatedDate', e.created_date, 
					'UpdatedDate', e.updated_date 
				) ORDER BY e.event_date DESC
			) FILTER (WHERE e.id IS NOT NULL),
			JSON_ARRAY()
		) as events`
	assert.Equal(t, expectedCoalesce, coalesce)

	expectedJoin := `
		LEFT JOIN event_person ep ON ep.person_id = p.id 
		LEFT JOIN event e ON e.id = ep.event_id `

	assert.Equal(t, expectedJoin, join)
}
