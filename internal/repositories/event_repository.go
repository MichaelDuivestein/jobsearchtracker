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

type EventRepository struct {
	database *sql.DB
}

func NewEventRepository(database *sql.DB) *EventRepository {
	return &EventRepository{database: database}
}

// Create can return ConflictError, InternalServiceError
func (repository *EventRepository) Create(event *models.CreateEvent) (*models.Event, error) {
	sqlInsert := `
		INSERT INTO event (
			id, event_type, description, notes, event_date, created_date, updated_date
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, event_type, description, notes, event_date, created_date, updated_date, null, null, null`

	var eventID uuid.UUID
	if event.ID != nil {
		eventID = *event.ID
	} else {
		eventID = uuid.New()
	}

	var eventDate, createdDate, updatedDate interface{}

	eventDate = event.EventDate.Format(timeutil.RFC3339Milli_Write)

	if event.CreatedDate != nil {
		createdDate = event.CreatedDate.Format(timeutil.RFC3339Milli_Write)
	} else {
		createdDate = time.Now().Format(timeutil.RFC3339Milli_Write)
	}

	if event.UpdatedDate != nil {
		updatedDate = event.UpdatedDate.Format(timeutil.RFC3339Milli_Write)
	}

	row := repository.database.QueryRow(
		sqlInsert,
		eventID,
		event.EventType,
		event.Description,
		event.Notes,
		eventDate,
		createdDate,
		updatedDate,
	)

	result, err := repository.mapRow(row, "Create")
	if err != nil {
		if err.Error() == "constraint failed: UNIQUE constraint failed: event.id (1555)" {
			slog.Info(
				"event_repository.CreateEvent: UNIQUE constraint failed",
				"ID", eventID)
			return nil, internalErrors.NewConflictError(
				"ID already exists in database: '" + eventID.String() + "'")
		}
		return nil, err
	}

	return result, nil
}

// GetById can return InternalServiceError, NotFoundError, ValidationError
func (repository *EventRepository) GetByID(id *uuid.UUID) (*models.Event, error) {
	if id == nil {
		slog.Info("event_repository.GetById: ID is nil")
		var id = "ID"
		return nil, internalErrors.NewValidationError(&id, "ID is nil")
	}

	sqlSelect := `
		SELECT id, event_type, description, notes, event_date, created_date, updated_date, null, null, null
		FROM event
		WHERE id = ? `

	row := repository.database.QueryRow(sqlSelect, id)
	result, err := repository.mapRow(row, "GetById")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("event_repository.GetById: No result found for ID", "ID", id, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + id.String() + "'")
		}
		return nil, err
	}

	return result, err
}

// GetAll can return InternalServiceError
func (repository *EventRepository) GetAll(
	includeApplications models.IncludeExtraDataType,
	includeCompanies models.IncludeExtraDataType,
	includePersons models.IncludeExtraDataType) ([]*models.Event, error) {
	sqlSelect := `
		SELECT e.id, e.event_type, e.description, e.notes, e.event_date, e.created_date, e.updated_date, %s, %s, %s
		FROM event e %s %s %s
		GROUP BY e.ID
		ORDER BY e.event_date DESC`

	applicationsCoalesceString, applicationsJoinString :=
		repository.buildApplicationsCoalesceAndJoin(includeApplications)
	companiesCoalesceString, companiesJoinString := repository.buildCompaniesCoalesceAndJoin(includeCompanies)
	personsCoalesceString, personsJoinString := repository.buildPersonsCoalesceAndJoin(includePersons)

	sqlSelect = fmt.Sprintf(
		sqlSelect,
		applicationsCoalesceString,
		companiesCoalesceString,
		personsCoalesceString,
		applicationsJoinString,
		companiesJoinString,
		personsJoinString)

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.Event

	for rows.Next() {
		result, err := repository.mapRow(rows, "GetAll")
		if err != nil {
			slog.Error("event_repository.GetAll: mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("error processing event data" + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("event_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError("Error reading events from database: " + err.Error())
	}

	return results, nil
}

// Update can return InternalServiceError, ValidationError
func (repository *EventRepository) Update(event *models.UpdateEvent) error {
	var sqlString strings.Builder
	var sqlParts []string
	var sqlVars []interface{}

	sqlString.WriteString(`
		UPDATE event SET 
			updated_date = ?, 
			`)
	sqlVars = append(sqlVars, time.Now().Format(timeutil.RFC3339Milli_Write))

	updateItemCount := 0

	if event.EventType != nil {
		sqlParts = append(sqlParts, "event_type = ?")
		sqlVars = append(sqlVars, *event.EventType)
		updateItemCount++
	}

	if event.Description != nil {
		sqlParts = append(sqlParts, "description = ?")
		sqlVars = append(sqlVars, *event.Description)
		updateItemCount++
	}

	if event.Notes != nil {
		sqlParts = append(sqlParts, "notes = ?")
		sqlVars = append(sqlVars, *event.Notes)
		updateItemCount++
	}

	if event.EventDate != nil {
		sqlParts = append(sqlParts, "event_date = ?")
		sqlVars = append(sqlVars, event.EventDate.Format(timeutil.RFC3339Milli_Write))
		updateItemCount++
	}

	if updateItemCount == 0 {
		slog.Info("event_repository.Update: nothing to update", "id", event.ID)
		return internalErrors.NewValidationError(nil, "nothing to update")
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, ", \n\t\t\t", nil)
	if err != nil {
		var message = "unable to join SQL statement string"
		slog.Error("event_repository.Update: unable to join SQL statement string", "error", err)
		return internalErrors.NewInternalServiceError(message)
	}

	sqlString.WriteString(sqlPayload)

	sqlString.WriteString(`
		WHERE id = ? `)
	sqlVars = append(sqlVars, event.ID)

	_, err = repository.database.Exec(
		sqlString.String(),
		sqlVars...,
	)

	if err != nil {
		slog.Error("event_repository.Update: unable to update event", "id", event.ID, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	return err
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (repository *EventRepository) Delete(id *uuid.UUID) error {
	if id == nil {
		slog.Error("event_repository.Delete: ID is nil")
		id := "ID"
		return internalErrors.NewValidationError(&id, "ID is nil")
	}

	sqlDelete := "DELETE FROM event WHERE id = ?"

	result, err := repository.database.Exec(sqlDelete, id)
	if err != nil {
		slog.Error("event_repository.Delete: Error trying to delete event", "id", id, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("event_repository.Delete: Error trying to delete event", "id", id, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError("event does not exist. ID: " + id.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

// mapRow can return InternalServiceError
func (repository *EventRepository) mapRow(
	scanner interface{ Scan(...interface{}) error },
	methodName string) (*models.Event, error) {

	var result models.Event
	var eventDate, createdDate, updatedDate, applicationsString, companiesString, personsString sql.NullString

	err := scanner.Scan(
		&result.ID,
		&result.EventType,
		&result.Description,
		&result.Notes,
		&eventDate,
		&createdDate,
		&updatedDate,
		&applicationsString,
		&companiesString,
		&personsString)

	if err != nil {
		return nil, err
	}

	if eventDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, eventDate.String)
		if err != nil {
			slog.Error("event_repository."+methodName+": Error parsing eventDate in mapRow",
				"eventDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing eventDate: " + err.Error())
		}
		result.EventDate = &timestamp
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, createdDate.String)
		if err != nil {
			slog.Error("event_repository."+methodName+": Error parsing createdDate in mapRow",
				"createdDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = &timestamp
	}

	if updatedDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, updatedDate.String)
		if err != nil {
			slog.Error("event_repository."+methodName+": Error parsing updatedDate in mapRow",
				"updatedDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing updatedDate: " + err.Error())
		}
		result.UpdatedDate = &timestamp
	}

	if applicationsString.Valid {
		var applications []*models.Application
		if err = json.Unmarshal([]byte(applicationsString.String), &applications); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing applications: " + err.Error())
		}

		if len(applications) > 0 {
			result.Applications = &applications
		}
	}

	if companiesString.Valid {
		var companies []*models.Company
		if err = json.Unmarshal([]byte(companiesString.String), &companies); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing companies: " + err.Error())
		}

		if len(companies) > 0 {
			result.Companies = &companies
		}
	}

	if personsString.Valid {
		var persons []*models.Person
		if err = json.Unmarshal([]byte(personsString.String), &persons); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing persons: " + err.Error())
		}

		if len(persons) > 0 {
			result.Persons = &persons
		}
	}

	return &result, nil
}

func (repository *EventRepository) buildApplicationsCoalesceAndJoin(
	includeApplications models.IncludeExtraDataType) (string, string) {

	if includeApplications == models.IncludeExtraDataTypeNone {
		return "null \n", ""
	}

	coalesceString := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', a.id%s
				) ORDER BY a.created_date DESC
			) FILTER (WHERE a.id IS NOT NULL),
			JSON_ARRAY()
		) as applications
		`

	allColumns := ""
	if includeApplications == models.IncludeExtraDataTypeAll {
		allColumns = `,
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
					'UpdatedDate', a.updated_date`
	}
	coalesceString = fmt.Sprintf(coalesceString, allColumns)

	joinString := `
		LEFT JOIN application_event ae ON ae.event_id = e.id 
		LEFT JOIN application a ON a.id = ae.application_id `

	return coalesceString, joinString
}

func (repository *EventRepository) buildCompaniesCoalesceAndJoin(
	includeCompanies models.IncludeExtraDataType) (string, string) {

	if includeCompanies == models.IncludeExtraDataTypeNone {
		return "null \n", ""
	}

	coalesceString := `
		COALESCE(
			JSON_GROUP_ARRAY(
				DISTINCT JSON_OBJECT(
					'ID', c.id%s
				) ORDER BY c.created_date DESC
			) FILTER (WHERE c.id IS NOT NULL),
			JSON_ARRAY()
		) as companies`

	allColumns := ""
	if includeCompanies == models.IncludeExtraDataTypeAll {
		allColumns = `, 
					'Name', c.name, 
					'CompanyType', c.company_type, 
					'Notes', c.notes, 
					'LastContact', c.last_contact, 
					'CreatedDate', c.created_date, 
					'UpdatedDate', c.updated_date `
	}
	coalesceString = fmt.Sprintf(coalesceString, allColumns)

	joinString := `
		LEFT JOIN company_event ce ON ce.event_id = e.id 
		LEFT JOIN company c ON c.id = ce.company_id `

	return coalesceString, joinString
}

func (repository *EventRepository) buildPersonsCoalesceAndJoin(
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

	joinString := `
		LEFT JOIN event_person ep ON ep.event_id = e.id 
		LEFT JOIN person p ON p.id = ep.person_id `

	return coalesceString, joinString
}
