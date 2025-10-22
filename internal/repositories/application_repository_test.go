package repositories

import (
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- buildCompanyCoalesceAndJoin tests: --------

func TestBuildCompanyCoalesceAndJoin_ShouldReturnEmprtStringsIfIncludeExtraDataTypeIsNone(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildCompanyCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "null \n", coalesce)
	assert.Equal(t, "", join)
}

func TestBuildCompanyCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	applicationRepository := NewApplicationRepository(nil)

	coalesce, join := applicationRepository.buildCompanyCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(t, "\n        LEFT JOIN company c ON (a.company_id = c.id)", join)

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

	assert.Equal(t, "\n        LEFT JOIN company c ON (a.company_id = c.id)", join)

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
