package repositories

import (
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- buildCompanyCoalesceAndJoin tests: --------

func TestBuildCompanyCoalesceAndJoin_ShouldReturnEmptyStringsIfIncludeExtraDataTypeIsNone(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildCompanyCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestBuildCompanyCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildCompanyCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(t, "\n\t\tLEFT JOIN company c ON (a.company_id = c.id)", join)

	expectedCoalesce := `
		CASE 
			WHEN c.id IS NOT NULL THEN JSON_OBJECT(
				'ID', c.id
			)
			ELSE NULL
		END as company`
	assert.Equal(t, expectedCoalesce, coalesce)
}

func TestBuildCompanyCoalesceAndJoin_ShouldBuildWithAllColumnsIfIncludeExtraDataTypeIsAll(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildCompanyCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(t, "\n\t\tLEFT JOIN company c ON (a.company_id = c.id)", join)

	expectedCoalesce := `
		CASE 
			WHEN c.id IS NOT NULL THEN JSON_OBJECT(
				'ID', c.id,
				'Name', c.name, 
				'CompanyType', c.company_type,  
				'Notes', c.notes, 
				'LastContact', c.last_contact, 
				'CreatedDate', c.created_date, 
				'UpdatedDate', c.updated_date
			)
			ELSE NULL
		END as company`
	assert.Equal(t, expectedCoalesce, coalesce)
}

// -------- buildRecruiterCoalesceAndJoin tests: --------

func TestBuildRecruiterCoalesceAndJoin_ShouldReturnEmpryStringsIfIncludeExtraDataTypeIsNone(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildRecruiterCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestBuildRecruiterCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildRecruiterCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(t, "\n\t\tLEFT JOIN company r ON (a.recruiter_id = r.id)", join)

	expectedCoalesce := `
		CASE 
			WHEN r.id IS NOT NULL THEN JSON_OBJECT(
				'ID', r.id
			)
			ELSE NULL
		END as recruiter`
	assert.Equal(t, expectedCoalesce, coalesce)
}

func TestBuildRecruiterCoalesceAndJoin_ShouldBuildWithAllColumnsIfIncludeExtraDataTypeIsAll(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildRecruiterCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(t, "\n\t\tLEFT JOIN company r ON (a.recruiter_id = r.id)", join)

	expectedCoalesce := `
		CASE 
			WHEN r.id IS NOT NULL THEN JSON_OBJECT(
				'ID', r.id,
				'Name', r.name, 
				'CompanyType', r.company_type,  
				'Notes', r.notes, 
				'LastContact', r.last_contact, 
				'CreatedDate', r.created_date, 
				'UpdatedDate', r.updated_date
			)
			ELSE NULL
		END as recruiter`
	assert.Equal(t, expectedCoalesce, coalesce)
}

// -------- buildPersonsCoalesceAndJoin tests: --------

func TestApplicationRepositoryBuildPersonsCoalesceAndJoin_ShouldReturnEmptyStringsIfIncludeExtraDataTypeIsNone(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestApplicationRepositoryBuildPersonsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(
		t,
		"LEFT JOIN application_person ap ON (ap.application_id = a.id)\n\t\tLEFT JOIN person p ON (ap.person_id = p.id)\n",
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

func TestApplicationRepositoryBuildPersonsCoalesceAndJoin_ShouldBuildWithAllColumnsIfIncludeExtraDataTypeIsAll(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildPersonsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(t, "LEFT JOIN application_person ap ON (ap.application_id = a.id)\n\t\tLEFT JOIN person p ON (ap.person_id = p.id)\n", join)

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

// -------- buildEventsCoalesceAndJoin tests: --------

func TestBuildEventsCoalesceAndJoin_ShouldReturnEmptyStringsIfIncludeExtraDataTypeIsNone(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestApplicationRepositoryBuildEventsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(
		t,
		"LEFT JOIN application_event ae ON (ae.application_id = a.id)\n\t\tLEFT JOIN event e ON (ae.event_id = e.id)\n",
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

func TestApplicationRepositoryBuildEventsCoalesceAndJoin_ShouldBuildWithAllColumnsIfIncludeExtraDataTypeIsAll(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildEventsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(
		t,
		"LEFT JOIN application_event ae ON (ae.application_id = a.id)\n\t\tLEFT JOIN event e ON (ae.event_id = e.id)\n",
		join)

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
