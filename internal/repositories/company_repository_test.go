package repositories

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- Update tests: --------

func TestUpdate_ShouldReturnValidationErrorIfNoCompanyFieldsToUpdate(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	updateModel := models.UpdateCompany{
		ID: uuid.New(),
	}
	err := companyRepository.Update(&updateModel)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: nothing to update", validationError.Error())
}

// -------- buildApplicationsCoalesceAndJoin tests: --------

func TestCompanyRepositoryBuildApplicationsCoalesceAndJoin_ShouldReturnNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestCompanyRepositoryBuildApplicationsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(t, "\n\t\tLEFT JOIN application a ON (c.id = a.company_id OR c.id = a.recruiter_id) \n", join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', a.id,
					'CompanyID', a.company_id,
					'RecruiterID', a.recruiter_id
				) ORDER BY a.created_date DESC
			) FILTER (WHERE a.id IS NOT NULL),
			JSON_ARRAY()
		) as applications
	`

	assert.Equal(t, expectedCoalesce, coalesce)
}

func TestCompanyRepositoryBuildApplicationsCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(t, "\n\t\tLEFT JOIN application a ON (c.id = a.company_id OR c.id = a.recruiter_id) \n", join)

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

// -------- buildEventsCoalesceAndJoin tests: --------

func TestCompanyRepositoryBuildEventsCoalesceAndJoin_ShouldNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestCompanyRepositoryBuildEventsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(
		t,
		"LEFT JOIN company_event ce ON (ce.company_id = c.id)\n\t\tLEFT JOIN event e ON (ce.event_id = e.id)\n",
		join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', e.id
				) ORDER BY e.event_date DESC
			) FILTER (WHERE e.id IS NOT NULL),
			JSON_ARRAY()
		) as events
`

	assert.Equal(t, expectedCoalesce, coalesce)
}

func TestCompanyRepositoryBuildEventsCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(t, "LEFT JOIN company_event ce ON (ce.company_id = c.id)\n\t\tLEFT JOIN event e ON (ce.event_id = e.id)\n", join)

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
		) as events
`

	assert.Equal(t, expectedCoalesce, coalesce)
}

// -------- buildPersonsCoalesceAndJoin tests: --------

func TestCompanyRepositoryBuildPersonsCoalesceAndJoin_ShouldNullTextAndEmptyStringIfIncludeExtraDataTypeIsNone(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)

}

func TestCompanyRepositoryBuildPersonsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(
		t,
		"LEFT JOIN company_person cp ON (cp.company_id = c.id)\n\t\tLEFT JOIN person p ON (cp.person_id = p.id)\n",
		join)

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

func TestCompanyRepositoryBuildPersonsCoalesceAndJoin_ShouldBuildWithAllColumnsIncludeExtraDataTypeIsAll(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(t, "LEFT JOIN company_person cp ON (cp.company_id = c.id)\n\t\tLEFT JOIN person p ON (cp.person_id = p.id)\n", join)

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
