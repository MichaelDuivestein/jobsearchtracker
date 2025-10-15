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
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: nothing to update", validationErr.Error())
}

// -------- buildApplicationsCoalesceAndJoin tests: --------

func TestBuildApplicationsCoalesceAndJoin_ShouldReturnEmptyStringsIfIncludeExtraDataTypeIsNone(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeNone)
	assert.Equal(t, "", coalesce)
	assert.Equal(t, "", join)

}

func TestBuildApplicationsCoalesceAndJoin_ShouldBuildWithOnlyIDsIfIncludeExtraDataTypeIsIDs(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeIDs)

	assert.Equal(t, "\t\tLEFT JOIN application a ON (c.id = a.company_id OR c.id = a.recruiter_id) \n", join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				JSON_OBJECT(
					'ID', a.id,
					'CompanyID', a.company_id,
					'RecruiterID', a.recruiter_id
				)
			) FILTER (WHERE a.id IS NOT NULL),
			JSON_ARRAY()
		) as applications
	`

	assert.Equal(t, expectedCoalesce, coalesce)
}

func TestBuildApplicationsCoalesceAndJoin_ShouldBuildWithAllColumnsIfIncludeExtraDataTypeIsAll(t *testing.T) {
	companyRepository := NewCompanyRepository(nil)

	coalesce, join := companyRepository.buildApplicationsCoalesceAndJoin(models.IncludeExtraDataTypeAll)

	assert.Equal(t, "\t\tLEFT JOIN application a ON (c.id = a.company_id OR c.id = a.recruiter_id) \n", join)

	expectedCoalesce := `
		COALESCE(
			JSON_GROUP_ARRAY(
				JSON_OBJECT(
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
				)
			) FILTER (WHERE a.id IS NOT NULL),
			JSON_ARRAY()
		) as applications
	`

	assert.Equal(t, expectedCoalesce, coalesce)
}
