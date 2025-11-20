package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/utils"
	"jobsearchtracker/pkg/timeutil"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ApplicationRepository struct {
	database *sql.DB
}

func NewApplicationRepository(database *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{database: database}
}

// Create can return ConflictError, InternalServiceError, ValidationError
func (repository *ApplicationRepository) Create(application *models.CreateApplication) (*models.Application, error) {
	sqlInsert := `
		INSERT INTO application (
	 		id, company_id, recruiter_id, job_title, job_ad_url, country, area, remote_status_type, weekdays_in_office, 
			estimated_cycle_time, estimated_commute_time, application_date, created_date, updated_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) 
	  	RETURNING 
			id, company_id, recruiter_id, job_title, job_ad_url, country, area, remote_status_type, 
		    weekdays_in_office, estimated_cycle_time, estimated_commute_time, application_date, created_date, 
		    updated_date, null, null, null, null; `

	var applicationID uuid.UUID
	if application.ID != nil {
		applicationID = *application.ID
	} else {
		applicationID = uuid.New()
	}

	var applicationDate, createdDate, updatedDate interface{}

	if application.ApplicationDate != nil {
		applicationDate = application.ApplicationDate.Format(timeutil.RFC3339Milli_Write)
	}

	if application.CreatedDate != nil {
		createdDate = application.CreatedDate.Format(timeutil.RFC3339Milli_Write)
	} else {
		createdDate = time.Now().Format(timeutil.RFC3339Milli_Write)
	}

	if application.UpdatedDate != nil {
		updatedDate = application.UpdatedDate.Format(timeutil.RFC3339Milli_Write)
	}

	row := repository.database.QueryRow(
		sqlInsert,
		applicationID,
		application.CompanyID,
		application.RecruiterID,
		application.JobTitle,
		application.JobAdURL,
		application.Country,
		application.Area,
		application.RemoteStatusType,
		application.WeekdaysInOffice,
		application.EstimatedCycleTime,
		application.EstimatedCommuteTime,
		applicationDate,
		createdDate,
		updatedDate,
	)

	result, err := repository.mapRow(row, "Create")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("application_repository.Create: No result found for ID",
				"ID", applicationID,
				"error", err.Error())
		} else if err.Error() == "constraint failed: CHECK constraint failed: company_reference_not_null (275)" {
			slog.Error("application_repository.Create: CHECK constraint failed: company_reference_not_null")
			return nil, internalErrors.NewValidationError(nil, "CompanyID and RecruiterID cannot both be empty")
		} else if err.Error() == "constraint failed: CHECK constraint failed: job_title_job_url_not_null (275)" {
			slog.Error("application_repository.Create: CHECK constraint failed: job_title_job_url_not_null")
			return nil, internalErrors.NewValidationError(nil, "JobTitle and JobAdURL cannot both be empty")
		} else if err.Error() == "constraint failed: UNIQUE constraint failed: application.id (1555)" {
			slog.Info(
				"application_repository.createApplication: UNIQUE constraint failed",
				"ID", applicationID.String())
			return nil, internalErrors.NewConflictError(
				"ID already exists in database: '" + applicationID.String() + "'")
		} else if err.Error() == "constraint failed: FOREIGN KEY constraint failed (787)" {
			// TODO: Use foreign key constraint names (in 0003_add_application.up.sql) once modernc.org/sqlite
			// supports it.
			slog.Info("application_repository.Create: FOREIGN KEY constraint failed (787)")
			return nil, internalErrors.NewValidationError(nil, "Foreign key does not exist")
		}
		return nil, err
	}

	return result, err
}

// GetById can return InternalServiceError, NotFoundError, ValidationError
func (repository *ApplicationRepository) GetById(id *uuid.UUID) (*models.Application, error) {
	if id == nil {
		slog.Info("application_repository.GetById: ID is nil")
		return nil, internalErrors.NewValidationError(nil, "ID is nil")
	}

	sqlSelect := `
		SELECT id, company_id, recruiter_id, job_title, job_ad_url, country, area, remote_status_type, 
		   weekdays_in_office, estimated_cycle_time, estimated_commute_time, application_date, created_date, 
		   updated_date, null, null, null, null 
		FROM application 
		WHERE id = ? `

	row := repository.database.QueryRow(sqlSelect, id)

	// can return ConflictError, InternalServiceError
	result, err := repository.mapRow(row, "GetById")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("application_repository.GetById: No result found for ID", "ID", id, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + id.String() + "'")
		}
		return nil, err
	}

	return result, nil
}

// GetAllByJobTitle can return InternalServiceError, NotFoundError, ValidationError
func (repository *ApplicationRepository) GetAllByJobTitle(jobTitle *string) ([]*models.Application, error) {
	if jobTitle == nil {
		slog.Info("application_repository.GetAllByJobTitle: JobTitle is nil")
		return nil, internalErrors.NewValidationError(nil, "JobTitle is nil")
	}

	sqlSelect := `
		SELECT id, company_id, recruiter_id, job_title, job_ad_url, country, area, remote_status_type, 
		   weekdays_in_office, estimated_cycle_time, estimated_commute_time, application_date, created_date, 
		   updated_date, null, null, null, null 
		FROM application 
		WHERE job_title LIKE ? 
		ORDER BY updated_Date DESC `

	wildcardJobTitle := "%" + *jobTitle + "%"
	rows, err := repository.database.Query(sqlSelect, wildcardJobTitle)
	if err != nil {
		return nil, err
	}

	var results []*models.Application

	for rows.Next() {
		// can return ConflictError, InternalServiceError
		result, err := repository.mapRow(rows, "GetAllByJobTitle")
		if err != nil {
			slog.Error("application_repository.GetAllByJobTitle: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing application data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("application_repository.GetAllByJobTitle: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError("Error reading applications from database: " + err.Error())
	}

	if len(results) == 0 {
		slog.Info("application_repository.GetAllByJobTitle: No result found for JobTitle", "JobTitle", jobTitle)
		return nil, internalErrors.NewNotFoundError("JobTitle: '" + *jobTitle + "'")
	}

	return results, nil
}

// GetAll can return InternalServiceError
func (repository *ApplicationRepository) GetAll(
	includeCompany models.IncludeExtraDataType,
	includeRecruiter models.IncludeExtraDataType,
	includePersons models.IncludeExtraDataType,
	includeEvents models.IncludeExtraDataType) ([]*models.Application, error) {

	sqlSelect := `
		SELECT a.id, a.company_id, a.recruiter_id, a.job_title, a.job_ad_url, a.country, a.area, a.remote_status_type, 
			a.weekdays_in_office, a.estimated_cycle_time, a.estimated_commute_time, a.application_date, a.created_date, 
			a.updated_date, %s, %s, %s, %s
		FROM application a %s %s %s %s
		GROUP BY a.id
		ORDER BY a.created_date DESC `

	companyCoalesceString, companyJoinString := repository.buildCompanyCoalesceAndJoin(includeCompany)
	recruiterCoalesceString, recruiterJoinString := repository.buildRecruiterCoalesceAndJoin(includeRecruiter)
	personsCoalesceString, personsJoinString := repository.buildPersonsCoalesceAndJoin(includePersons)
	eventsCoalesceString, eventsJoinString := repository.buildEventsCoalesceAndJoin(includeEvents)

	sqlSelect = fmt.Sprintf(
		sqlSelect,
		companyCoalesceString,
		recruiterCoalesceString,
		personsCoalesceString,
		eventsCoalesceString,
		companyJoinString,
		recruiterJoinString,
		personsJoinString,
		eventsJoinString)

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.Application
	for rows.Next() {
		// can return ConflictError, InternalServiceError
		result, err := repository.mapRow(rows, "GetAll")
		if err != nil {
			slog.Error("application_repository.GetAll: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing application data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("application_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError("Error reading applications from database: " + err.Error())
	}

	return results, nil
}

// Update can return InternalServiceError, ValidationError
func (repository *ApplicationRepository) Update(application *models.UpdateApplication) error {
	var sqlString strings.Builder
	var sqlParts []string
	var sqlVars []interface{}

	sqlString.WriteString(`
		UPDATE application SET
			updated_date = ?, 
			`)
	sqlVars = append(sqlVars, time.Now().Format(timeutil.RFC3339Milli_Write))

	updateItemCount := 0

	if application.CompanyID != nil {
		sqlParts = append(sqlParts, "company_id = ?")
		sqlVars = append(sqlVars, *application.CompanyID)
		updateItemCount++
	}

	if application.RecruiterID != nil {
		sqlParts = append(sqlParts, "recruiter_id = ?")
		sqlVars = append(sqlVars, *application.RecruiterID)
		updateItemCount++
	}

	if application.JobTitle != nil {
		sqlParts = append(sqlParts, "job_title = ?")
		sqlVars = append(sqlVars, *application.JobTitle)
		updateItemCount++
	}

	if application.JobAdURL != nil {
		sqlParts = append(sqlParts, "job_ad_url = ?")
		sqlVars = append(sqlVars, *application.JobAdURL)
		updateItemCount++
	}

	if application.Country != nil {
		sqlParts = append(sqlParts, "country = ?")
		sqlVars = append(sqlVars, *application.Country)
		updateItemCount++
	}

	if application.Area != nil {
		sqlParts = append(sqlParts, "area = ?")
		sqlVars = append(sqlVars, *application.Area)
		updateItemCount++
	}

	if application.RemoteStatusType != nil {
		sqlParts = append(sqlParts, "remote_status_type = ?")
		sqlVars = append(sqlVars, *application.RemoteStatusType)
		updateItemCount++
	}

	if application.WeekdaysInOffice != nil {
		sqlParts = append(sqlParts, "weekdays_in_office = ?")
		sqlVars = append(sqlVars, *application.WeekdaysInOffice)
		updateItemCount++
	}

	if application.EstimatedCycleTime != nil {
		sqlParts = append(sqlParts, "estimated_cycle_time = ?")
		sqlVars = append(sqlVars, *application.EstimatedCycleTime)
		updateItemCount++
	}

	if application.EstimatedCommuteTime != nil {
		sqlParts = append(sqlParts, "estimated_commute_time = ?")
		sqlVars = append(sqlVars, *application.EstimatedCommuteTime)
		updateItemCount++
	}

	if application.ApplicationDate != nil {
		sqlParts = append(sqlParts, "application_date = ?")
		sqlVars = append(sqlVars, application.ApplicationDate.Format(timeutil.RFC3339Milli_Write))
		updateItemCount++
	}

	if updateItemCount == 0 {
		slog.Info("application_repository.Update: nothing to update", "id", application.ID)
		return internalErrors.NewValidationError(nil, "nothing to update")
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, ", \n\t\t\t", nil)
	if err != nil {
		var message = "unable to join SQL statement string"
		slog.Error("application_repository.Update: unable to join SQL statement string", "error", err)
		return internalErrors.NewInternalServiceError(message)
	}

	sqlString.WriteString(sqlPayload)

	sqlString.WriteString(`
		WHERE id = ? `)
	sqlVars = append(sqlVars, application.ID)

	_, err = repository.database.Exec(
		sqlString.String(),
		sqlVars...,
	)

	if err != nil {
		slog.Error(
			"application_repository.Update: unable to update application",
			"id", application.ID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	return err
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (repository *ApplicationRepository) Delete(id *uuid.UUID) error {
	if id == nil {
		slog.Error("application_repository.Delete: ID is nil")
		return internalErrors.NewValidationError(nil, "ID is nil")
	}

	sqlDelete := "DELETE FROM application WHERE id = ?"

	result, err := repository.database.Exec(sqlDelete, id)
	if err != nil {
		slog.Error(
			"application_repository.Delete: Error trying to delete application",
			"id", id,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"application_repository.Delete: Error trying to delete application",
			"id", id,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError("Application does not exist. ID: " + id.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

func (repository *ApplicationRepository) mapRow(
	scanner interface{ Scan(...interface{}) error }, methodName string) (*models.Application, error) {

	var result models.Application
	var applicationDate,
		createdDate,
		updatedDate,
		companyString,
		recruiterString,
		personsString,
		eventsString sql.NullString

	err := scanner.Scan(
		&result.ID,
		&result.CompanyID,
		&result.RecruiterID,
		&result.JobTitle,
		&result.JobAdURL,
		&result.Country,
		&result.Area,
		&result.RemoteStatusType,
		&result.WeekdaysInOffice,
		&result.EstimatedCycleTime,
		&result.EstimatedCommuteTime,
		&applicationDate,
		&createdDate,
		&updatedDate,
		&companyString,
		&recruiterString,
		&personsString,
		&eventsString,
	)

	if err != nil {
		return nil, err
	}

	if applicationDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, applicationDate.String)
		if err != nil {
			slog.Error(
				"application_repository."+methodName+": Error parsing applicationDate",
				"applicationDate", applicationDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing applicationDate: " + err.Error())
		}
		result.ApplicationDate = &timestamp
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, createdDate.String)
		if err != nil {
			slog.Error("application_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = &timestamp
	}

	if updatedDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, updatedDate.String)
		if err != nil {
			slog.Error("application_repository."+methodName+": Error parsing updatedDate",
				"updatedDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing updatedDate: " + err.Error())
		}
		result.UpdatedDate = &timestamp
	}

	if companyString.Valid {
		var company *models.Company
		if err := json.NewDecoder(strings.NewReader(companyString.String)).Decode(&company); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing company: " + err.Error())
		}
		if company != nil {
			result.Company = company
		}
	}

	if recruiterString.Valid {
		var recruiter *models.Company
		if err := json.NewDecoder(strings.NewReader(recruiterString.String)).Decode(&recruiter); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing recruiter: " + err.Error())
		}
		if recruiter != nil {
			result.Recruiter = recruiter
		}
	}

	if personsString.Valid {
		var persons []*models.Person
		if err := json.NewDecoder(strings.NewReader(personsString.String)).Decode(&persons); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing persons: " + err.Error())
		}

		if len(persons) > 0 {
			result.Persons = &persons
		}
	}

	if eventsString.Valid {
		var events []*models.Event
		if err := json.NewDecoder(strings.NewReader(eventsString.String)).Decode(&events); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing events: " + err.Error())
		}

		if len(events) > 0 {
			result.Events = &events
		}
	}

	return &result, nil
}

func (repository *ApplicationRepository) buildCompanyCoalesceAndJoin(
	includeCompany models.IncludeExtraDataType) (string, string) {

	if includeCompany == models.IncludeExtraDataTypeNone {
		return "null \n", ""
	}

	coalesceString := `
		CASE 
			WHEN c.id IS NOT NULL THEN JSON_OBJECT(
				'ID', c.id%s
			)
			ELSE NULL
		END as company`

	allColumns := ""
	if includeCompany == models.IncludeExtraDataTypeAll {
		allColumns = `,
				'Name', c.name, 
				'CompanyType', c.company_type,  
				'Notes', c.notes, 
				'LastContact', c.last_contact, 
				'CreatedDate', c.created_date, 
				'UpdatedDate', c.updated_date`
	}
	coalesceString = fmt.Sprintf(coalesceString, allColumns)

	joinString := "\n\t\tLEFT JOIN company c ON (a.company_id = c.id)"

	return coalesceString, joinString
}

func (repository *ApplicationRepository) buildRecruiterCoalesceAndJoin(
	includeRecruiter models.IncludeExtraDataType) (string, string) {

	if includeRecruiter == models.IncludeExtraDataTypeNone {
		return "null \n", ""
	}

	coalesceString := `
		CASE 
			WHEN r.id IS NOT NULL THEN JSON_OBJECT(
				'ID', r.id%s
			)
			ELSE NULL
		END as recruiter`

	allColumns := ""
	if includeRecruiter == models.IncludeExtraDataTypeAll {
		allColumns = `,
				'Name', r.name, 
				'CompanyType', r.company_type,  
				'Notes', r.notes, 
				'LastContact', r.last_contact, 
				'CreatedDate', r.created_date, 
				'UpdatedDate', r.updated_date`
	}
	coalesceString = fmt.Sprintf(coalesceString, allColumns)

	joinString := "\n\t\tLEFT JOIN company r ON (a.recruiter_id = r.id)"

	return coalesceString, joinString
}

func (repository *ApplicationRepository) buildPersonsCoalesceAndJoin(
	includePersons models.IncludeExtraDataType) (string, string) {

	if includePersons == models.IncludeExtraDataTypeNone {
		return "null \n", ""
	}

	coalesceString := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', p.id%s
				) ORDER BY p.created_date DESC
			) FILTER (WHERE p.id IS NOT NULL),
			JSON_ARRAY()
		) as persons
`

	allColumns := ""
	if includePersons == models.IncludeExtraDataTypeAll {
		allColumns = `,
					'Name', p.name,
					'PersonType', p.person_type,
					'Email', p.email,
					'Phone', p.phone,
					'Notes', p.notes,
					'CreatedDate', p.created_date,
					'UpdatedDate', p.updated_date`

	}
	coalesceString = fmt.Sprintf(coalesceString, allColumns)

	joinString :=
		`LEFT JOIN application_person ap ON (ap.application_id = a.id)
		LEFT JOIN person p ON (ap.person_id = p.id)
`

	return coalesceString, joinString
}

func (repository *ApplicationRepository) buildEventsCoalesceAndJoin(
	includeEvents models.IncludeExtraDataType) (string, string) {

	if includeEvents == models.IncludeExtraDataTypeNone {
		return "null \n", ""
	}

	coalesceString := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', e.id%s
				) ORDER BY e.event_date DESC
			) FILTER (WHERE e.id IS NOT NULL),
			JSON_ARRAY()
		) as events
`

	allColumns := ""
	if includeEvents == models.IncludeExtraDataTypeAll {
		allColumns = `,
					'EventType', e.event_type,
					'Description', e.description,
					'Notes', e.notes,
					'EventDate', e.event_date,
					'CreatedDate', e.created_date,
					'UpdatedDate', e.updated_date`

	}
	coalesceString = fmt.Sprintf(coalesceString, allColumns)

	joinString :=
		`LEFT JOIN application_event ae ON (ae.application_id = a.id)
		LEFT JOIN event e ON (ae.event_id = e.id)
`

	return coalesceString, joinString
}
