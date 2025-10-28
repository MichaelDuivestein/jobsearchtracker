package repositories

import (
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
                JSON_OBJECT(
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
                JSON_OBJECT(
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
