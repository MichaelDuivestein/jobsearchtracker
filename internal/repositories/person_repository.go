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

type PersonRepository struct {
	database *sql.DB
}

func NewPersonRepository(database *sql.DB) *PersonRepository {
	return &PersonRepository{database: database}
}

// Create can return ConflictError, InternalServiceError
func (repository *PersonRepository) Create(person *models.CreatePerson) (*models.Person, error) {
	sqlInsert := `
		INSERT INTO person (
			id, name, person_type, email, phone, notes, created_date, updated_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, name, person_type, email, phone, notes, created_date, updated_date, null, null`

	var personID uuid.UUID
	if person.ID != nil {
		personID = *person.ID
	} else {
		personID = uuid.New()
	}

	var createdDate, updatedDate interface{}

	if person.CreatedDate != nil {
		createdDate = person.CreatedDate.Format(timeutil.RFC3339Milli_Write)
	} else {
		createdDate = time.Now().Format(timeutil.RFC3339Milli_Write)
	}

	if person.UpdatedDate != nil {
		updatedDate = person.UpdatedDate.Format(timeutil.RFC3339Milli_Write)
	}

	row := repository.database.QueryRow(
		sqlInsert,
		personID,
		person.Name,
		person.PersonType,
		person.Email,
		person.Phone,
		person.Notes,
		createdDate,
		updatedDate,
	)

	// can return ConflictError, InternalServiceError
	result, err := repository.mapRow(row, "Create")
	if err != nil {
		if err.Error() == "constraint failed: UNIQUE constraint failed: person.id (1555)" {
			slog.Info(
				"person_repository.CreatePerson: UNIQUE constraint failed",
				"ID", personID)
			return nil, internalErrors.NewConflictError(
				"ID already exists in database: '" + personID.String() + "'")
		} else if errors.Is(err, sql.ErrNoRows) {
			slog.Info("person_repository.create: No result found for ID", "ID", personID, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + personID.String() + "'")
		}
		return nil, err
	}

	return result, nil
}

// GetById can return InternalServiceError, NotFoundError, ValidationError
func (repository *PersonRepository) GetById(id *uuid.UUID) (*models.Person, error) {
	if id == nil {
		slog.Info("person_repository.GetById: ID is nil")
		var id = "ID"
		return nil, internalErrors.NewValidationError(&id, "ID is nil")
	}

	sqlSelect := `
		SELECT id, name, person_type, email, phone, notes, created_date, updated_date, null, null 
		FROM person
		WHERE id = ? `

	row := repository.database.QueryRow(sqlSelect, id)
	result, err := repository.mapRow(row, "GetById")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("person_repository.GetById: No result found for ID", "ID", id, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + id.String() + "'")
		}
		return nil, err
	}

	return result, err
}

// GetAllByName can return InternalServiceError, NotFoundError, ValidationError
func (repository *PersonRepository) GetAllByName(name *string) ([]*models.Person, error) {
	if name == nil {
		slog.Info("person_repository.name: Name is nil")
		var id = "Name"
		return nil, internalErrors.NewValidationError(&id, "Name is nil")
	}

	wildcardName := "%" + *name + "%"

	sqlSelect := `
		SELECT id, name, person_type, email, phone, notes, created_date, updated_date, null, null
		FROM person
		WHERE name LIKE ?
		ORDER BY name ASC `

	rows, err := repository.database.Query(sqlSelect, wildcardName)
	if err != nil {
		return nil, err
	}

	var results []*models.Person

	for rows.Next() {
		result, err := repository.mapRow(rows, "GetAllByName")
		if err != nil {
			slog.Error("person_repository.GetAllByName: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("person_repository.GetAllByName: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError("Error reading persons from database: " + err.Error())
	}

	if len(results) == 0 {
		slog.Info("person_repository.GetByName: No result found for Name", "Name", name)
		return nil, internalErrors.NewNotFoundError("Name: '" + *name + "'")
	}

	return results, nil
}

// GetAll can return InternalServiceError
func (repository *PersonRepository) GetAll(
	includeCompanies models.IncludeExtraDataType,
	includeEvents models.IncludeExtraDataType) ([]*models.Person, error) {
	sqlSelect := `
        SELECT p.id, p.name, p.person_type, p.email, p.phone, p.notes, p.created_date, p.updated_date, %s, %s
        FROM person p %s %s
        GROUP BY p.id
        ORDER BY p.created_date DESC; `

	companiesCoalesceString, companiesJoinString := repository.buildCompaniesCoalesceAndJoin(includeCompanies)
	eventsCoalesceString, eventsJoinString := repository.buildEventsCoalesceAndJoin(includeEvents)

	sqlSelect = fmt.Sprintf(
		sqlSelect,
		companiesCoalesceString,
		eventsCoalesceString,
		companiesJoinString,
		eventsJoinString)

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.Person

	for rows.Next() {
		result, err := repository.mapRow(rows, "GetAll")
		if err != nil {
			slog.Error("person_repository.GetAll: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("person_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError("Error reading persons from database: " + err.Error())
	}

	return results, nil
}

// Update can return InternalServiceError, ValidationError
func (repository *PersonRepository) Update(person *models.UpdatePerson) error {
	var sqlString strings.Builder
	var sqlParts []string
	var sqlVars []interface{}

	sqlString.WriteString(`
		UPDATE person SET 
			updated_date = ?, 
			`)
	sqlVars = append(sqlVars, time.Now().Format(timeutil.RFC3339Milli_Write))

	updateItemCount := 0

	if person.Name != nil {
		sqlParts = append(sqlParts, "name = ?")
		sqlVars = append(sqlVars, *person.Name)
		updateItemCount++
	}

	if person.PersonType != nil {
		sqlParts = append(sqlParts, "person_type = ?")
		sqlVars = append(sqlVars, *person.PersonType)
		updateItemCount++
	}

	if person.Email != nil {
		sqlParts = append(sqlParts, "email = ?")
		sqlVars = append(sqlVars, *person.Email)
		updateItemCount++
	}

	if person.Phone != nil {
		sqlParts = append(sqlParts, "phone = ?")
		sqlVars = append(sqlVars, *person.Phone)
		updateItemCount++
	}

	if person.Notes != nil {
		sqlParts = append(sqlParts, "notes = ?")
		sqlVars = append(sqlVars, *person.Notes)
		updateItemCount++
	}

	if updateItemCount == 0 {
		slog.Info("person_repository.Update: nothing to update", "id", person.ID)
		return internalErrors.NewValidationError(nil, "nothing to update")
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, ", \n\t\t\t", nil)
	if err != nil {
		var message = "unable to join SQL statement string"
		slog.Error("person_repository.Update: unable to join SQL statement string", "error", err)
		return internalErrors.NewInternalServiceError(message)
	}

	sqlString.WriteString(sqlPayload)

	sqlString.WriteString(`
		WHERE id = ? `)
	sqlVars = append(sqlVars, person.ID)

	_, err = repository.database.Exec(
		sqlString.String(),
		sqlVars...,
	)

	if err != nil {
		slog.Error("person_repository.Update: unable to update person", "id", person.ID, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	return err
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (repository *PersonRepository) Delete(id *uuid.UUID) error {
	if id == nil {
		slog.Error("person_repository.Delete: ID is nil")
		id := "ID"
		return internalErrors.NewValidationError(&id, "ID is nil")
	}

	sqlDelete := "DELETE FROM person WHERE id = ?"

	result, err := repository.database.Exec(sqlDelete, id)
	if err != nil {
		slog.Error("person_repository.Delete: Error trying to delete person", "id", id, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("person_repository.Delete: Error trying to delete person", "id", id, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError("Person does not exist. ID: " + id.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

// mapRow can return InternalServiceError
func (repository *PersonRepository) mapRow(
	scanner interface{ Scan(...interface{}) error },
	methodName string) (*models.Person, error) {

	var result models.Person
	var createdDate, updatedDate, companiesString, eventsString sql.NullString

	err := scanner.Scan(
		&result.ID,
		&result.Name,
		&result.PersonType,
		&result.Email,
		&result.Phone,
		&result.Notes,
		&createdDate,
		&updatedDate,
		&companiesString,
		&eventsString,
	)

	if err != nil {
		return nil, err
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, createdDate.String)
		if err != nil {
			slog.Error("person_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = &timestamp
	}

	if updatedDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, updatedDate.String)
		if err != nil {
			slog.Error("person_repository."+methodName+": Error parsing updatedDate",
				"updatedDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing updatedDate: " + err.Error())
		}
		result.UpdatedDate = &timestamp
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

	if eventsString.Valid {
		var events []*models.Event
		if err = json.Unmarshal([]byte(eventsString.String), &events); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing events: " + err.Error())
		}

		if len(events) > 0 {
			result.Events = &events
		}
	}

	return &result, nil
}

func (repository *PersonRepository) buildCompaniesCoalesceAndJoin(
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
		LEFT JOIN company_person cp ON cp.person_id = p.id 
		LEFT JOIN company c ON c.id = cp.company_id `

	return coalesceString, joinString
}

func (repository *PersonRepository) buildEventsCoalesceAndJoin(
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
		) as events`

	allColumns := ""
	if includeEvents == models.IncludeExtraDataTypeAll {
		allColumns = `, 
					'EventType', e.event_type, 
					'Description', e.description,  
					'Notes', e.notes, 
					'EventDate', e.event_date, 
					'CreatedDate', e.created_date, 
					'UpdatedDate', e.updated_date `
	}
	coalesceString = fmt.Sprintf(coalesceString, allColumns)

	joinString := `
		LEFT JOIN event_person ep ON ep.person_id = p.id 
		LEFT JOIN event e ON e.id = ep.event_id `

	return coalesceString, joinString
}
