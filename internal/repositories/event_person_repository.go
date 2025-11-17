package repositories

import (
	"database/sql"
	"errors"
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

type EventPersonRepository struct {
	database *sql.DB
}

func NewEventPersonRepository(database *sql.DB) *EventPersonRepository {
	return &EventPersonRepository{database: database}
}

// AssociateEventPerson can return ConflictError, InternalServiceError
func (repository *EventPersonRepository) AssociateEventPerson(
	associateModel *models.AssociateEventPerson) (*models.EventPerson, error) {

	sqlInsert := `
		INSERT INTO event_person (
			event_id, person_id, created_date
		) VALUES (?, ?, ?) 
		RETURNING event_id, person_id, created_date; `

	var createdDate string
	if associateModel.CreatedDate != nil {
		createdDate = associateModel.CreatedDate.Format(timeutil.RFC3339Milli_Write)
	} else {
		createdDate = time.Now().UTC().Format(timeutil.RFC3339Milli_Write)
	}

	row := repository.database.QueryRow(
		sqlInsert,
		associateModel.EventID,
		associateModel.PersonID,
		createdDate,
	)

	if row.Err() != nil {
		if row.Err().Error() ==
			"constraint failed: UNIQUE constraint failed: event_person.event_id, event_person.person_id (1555)" {

			slog.Info(
				"event_person_repository.associateToEvent: UNIQUE constraint failed",
				"event_id", associateModel.EventID,
				"person_id", associateModel.PersonID)

			return nil, internalErrors.NewConflictError(
				"EventID and PersonID combination already exists in database.")
		} else if row.Err().Error() == "constraint failed: FOREIGN KEY constraint failed (787)" {
			// TODO: Use foreign key constraint names (in 0003_add_application.up.sql) once modernc.org/sqlite
			// supports it.
			slog.Info("event_person_repository.Create: FOREIGN KEY constraint failed (787)")
			return nil, internalErrors.NewValidationError(nil, "Foreign key does not exist")
		}
		return nil, row.Err()
	}

	// can return InternalServiceError
	result, err := repository.mapRow(row, "Create")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("event_person_repository.create: No result found.", "error", err.Error())
			return nil, internalErrors.NewNotFoundError("Unable to map EventPerson")
		}
		return nil, err
	}

	return result, nil
}

// GetByID can return ValidationError, InternalServiceError
func (repository *EventPersonRepository) GetByID(
	eventID *uuid.UUID, personID *uuid.UUID) ([]*models.EventPerson, error) {

	if (eventID == nil || *eventID == uuid.Nil) && (personID == nil || *personID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "eventID and personID cannot both be empty")
	}

	var sqlString strings.Builder
	var sqlParts []string
	var sqlVars []interface{}

	sqlString.WriteString(`
		SELECT event_id, person_id, created_date 
		FROM event_person 
		WHERE `)

	eventIDAdded := false
	if eventID != nil && *eventID != uuid.Nil {
		sqlParts = append(sqlParts, "event_id = ? ")
		sqlVars = append(sqlVars, eventID)
		eventIDAdded = true
	}

	if personID != nil && *personID != uuid.Nil {
		if eventIDAdded {
			sqlParts = append(sqlParts, "\n\t\tAND ")
		}
		sqlParts = append(sqlParts, "person_id = ? ")
		sqlVars = append(sqlVars, personID)
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, " ", nil)
	if err != nil {
		//var message = "unable to join SQL statement string"
		slog.Error("event_person_repository.GetByID: unable to join SQL statement string", "error", err)
		//return internalErrors.NewInternalServiceError(message)
	}

	sqlString.WriteString(sqlPayload)

	sqlString.WriteString("\n\t\tORDER BY created_date DESC ")

	rows, err := repository.database.Query(sqlString.String(), sqlVars...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.EventPerson
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "getByID")
		if err != nil {
			slog.Error("event_person_repository.getByID: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("event_person_repository.getByID: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading PersonCompanies from database: " + err.Error())
	}

	return results, nil
}

// GetAll can return InternalServiceError
func (repository *EventPersonRepository) GetAll() ([]*models.EventPerson, error) {
	sqlSelect := `
		SELECT event_id, person_id, created_date 
		FROM event_person 
		ORDER BY created_date DESC; `

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.EventPerson
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "GetAll")
		if err != nil {
			slog.Error("event_person_repository.GetAll: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("event_person_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading PersonCompanies from database: " + err.Error())
	}

	return results, nil
}

// Delete can return InternalServiceError, NotFoundError
func (repository *EventPersonRepository) Delete(model *models.DeleteEventPerson) error {
	sqlDelete := `
		DELETE FROM event_person 
		WHERE event_id = ? 
		AND person_id = ?; `

	result, err := repository.database.Exec(sqlDelete, model.EventID, model.PersonID)
	if err != nil {
		slog.Error(
			"event_person_repository.Delete: Error trying to delete EventPerson",
			"eventID", model.EventID,
			"personID", model.PersonID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"event_person_repository.Delete: Error trying to delete EventPerson",
			"eventID", model.EventID,
			"personID", model.PersonID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError(
			"EventPerson does not exist. eventID: " + model.EventID.String() +
				", personID: " + model.PersonID.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

// mapRow can return InternalServiceError
func (repository *EventPersonRepository) mapRow(scanner interface{ Scan(...interface{}) error },
	methodName string) (*models.EventPerson, error) {

	var result models.EventPerson
	var createdDate sql.NullString

	err := scanner.Scan(&result.EventID, &result.PersonID, &createdDate)

	if err != nil {
		return nil, err
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, createdDate.String)
		if err != nil {
			slog.Error("event_person_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	return &result, nil
}
