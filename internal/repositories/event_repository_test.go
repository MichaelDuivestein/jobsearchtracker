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

func TestGetByID_ShouldReturnValidationErrorIfEventIDIsNil(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	nilEvent, err := eventRepository.GetByID(nil)
	assert.Nil(t, nilEvent)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

// -------- Update tests: --------

func TestUpdate_ShouldReturnValidationErrorIfNoEventFieldsToUpdate(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	eventToUpdate := &models.UpdateEvent{
		ID: uuid.New(),
	}
	err := eventRepository.Update(eventToUpdate)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

// -------- Delete tests: --------

func TestDelete_ShouldReturnValidationErrorIfEventIDIsNil(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	err := eventRepository.Delete(nil)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error on field 'ID': ID is nil", validationError.Error())
}

// -------- buildApplicationsCoalesceAndJoin tests: --------

func TestEventRepositoryBuildApplicationsCoalesceAndJoin_ShouldReturnNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestEventRepositoryBuildApplicationsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	expectedJoin := `
		LEFT JOIN application_event ae ON ae.event_id = e.id 
		LEFT JOIN application a ON a.id = ae.application_id `
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

func TestEventRepositoryBuildApplicationsCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	expectedJoin := `
		LEFT JOIN application_event ae ON ae.event_id = e.id 
		LEFT JOIN application a ON a.id = ae.application_id `
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

func TestEventRepositoryBuildCompaniesCoalesceAndJoin_ShouldReturnNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildCompaniesCoalesceAndJoin(models.IncludeExtraDataTypeNone)

	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestEventRepositoryBuildCompaniesCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildCompaniesCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

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
		LEFT JOIN company_event ce ON ce.event_id = e.id 
		LEFT JOIN company c ON c.id = ce.company_id `
	assert.Equal(t, expectedJoin, join)
}

func TestEventRepositoryBuildCompaniesCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildCompaniesCoalesceAndJoin(models.IncludeExtraDataTypeAll)

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
		LEFT JOIN company_event ce ON ce.event_id = e.id 
		LEFT JOIN company c ON c.id = ce.company_id `
	assert.Equal(t, expectedJoin, join)
}

// -------- buildPersonsCoalesceAndJoin tests: --------

func TestEventRepositoryBuildPersonsCoalesceAndJoin_ShouldNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)

}

func TestEventRepositoryBuildPersonsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	expectedJoin := `
		LEFT JOIN event_person ep ON ep.event_id = e.id 
		LEFT JOIN person p ON p.id = ep.person_id `
	assert.Equal(t, expectedJoin, join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', p.id
				) ORDER BY p.created_date DESC
			) FILTER (WHERE p.id IS NOT NULL),
			JSON_ARRAY()
		) as persons
`

	assert.Equal(t, expectedCoalesce, coalesce)
}

func TestEventRepositoryBuildPersonsCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
	eventRepository := NewEventRepository(nil)

	coalesce, join := eventRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	expectedJoin := `
		LEFT JOIN event_person ep ON ep.event_id = e.id 
		LEFT JOIN person p ON p.id = ep.person_id `
	assert.Equal(t, expectedJoin, join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', p.id,
					'Name', p.name,
					'PersonType', p.person_type,
					'Email', p.email,
					'Phone', p.phone,
					'Notes', p.notes,
					'CreatedDate', p.created_date,
					'UpdatedDate', p.updated_date
				) ORDER BY p.created_date DESC
			) FILTER (WHERE p.id IS NOT NULL),
			JSON_ARRAY()
		) as persons
`

	assert.Equal(t, expectedCoalesce, coalesce)
}
