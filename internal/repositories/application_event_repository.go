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

type ApplicationEventRepository struct {
	database *sql.DB
}

func NewApplicationEventRepository(database *sql.DB) *ApplicationEventRepository {
	return &ApplicationEventRepository{database: database}
}

// AssociateApplicationEvent can return ConflictError, InternalServiceError
func (repository *ApplicationEventRepository) AssociateApplicationEvent(
	associateModel *models.AssociateApplicationEvent) (*models.ApplicationEvent, error) {

	sqlInsert := `
		INSERT INTO application_event (
			application_id, event_id, created_date
		) VALUES (?, ?, ?) 
		RETURNING application_id, event_id, created_date; `

	var createdDate string
	if associateModel.CreatedDate != nil {
		createdDate = associateModel.CreatedDate.Format(timeutil.RFC3339Milli_Write)
	} else {
		createdDate = time.Now().UTC().Format(timeutil.RFC3339Milli_Write)
	}

	row := repository.database.QueryRow(
		sqlInsert,
		associateModel.ApplicationID,
		associateModel.EventID,
		createdDate,
	)

	if row.Err() != nil {
		if row.Err().Error() ==
			"constraint failed: UNIQUE constraint failed: application_event.application_id, application_event.event_id (1555)" {

			slog.Info(
				"application_event_repository.associateToApplication: UNIQUE constraint failed",
				"application_id", associateModel.ApplicationID,
				"event_id", associateModel.EventID)

			return nil, internalErrors.NewConflictError(
				"ApplicationID and EventID combination already exists in database.")
		} else if row.Err().Error() == "constraint failed: FOREIGN KEY constraint failed (787)" {
			// TODO: Use foreign key constraint names (in 0003_add_application.up.sql) once modernc.org/sqlite
			// supports it.
			slog.Info("application_event_repository.Create: FOREIGN KEY constraint failed (787)")
			return nil, internalErrors.NewValidationError(nil, "Foreign key does not exist")
		}
		return nil, row.Err()
	}

	// can return InternalServiceError
	result, err := repository.mapRow(row, "Create")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("application_event_repository.create: No result found.", "error", err.Error())
			return nil, internalErrors.NewNotFoundError("Unable to map ApplicationEvent")
		}
		return nil, err
	}

	return result, nil
}

// GetByID can return ValidationError, InternalServiceError
func (repository *ApplicationEventRepository) GetByID(
	applicationID *uuid.UUID, eventID *uuid.UUID) ([]*models.ApplicationEvent, error) {

	if (applicationID == nil || *applicationID == uuid.Nil) && (eventID == nil || *eventID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "applicationID and eventID cannot both be empty")
	}

	var sqlString strings.Builder
	var sqlParts []string
	var sqlVars []interface{}

	sqlString.WriteString(`
		SELECT application_id, event_id, created_date 
		FROM application_event 
		WHERE `)

	applicationIDAdded := false
	if applicationID != nil && *applicationID != uuid.Nil {
		sqlParts = append(sqlParts, "application_id = ? ")
		sqlVars = append(sqlVars, applicationID)
		applicationIDAdded = true
	}

	if eventID != nil && *eventID != uuid.Nil {
		if applicationIDAdded {
			sqlParts = append(sqlParts, "\n\t\tAND ")
		}
		sqlParts = append(sqlParts, "event_id = ? ")
		sqlVars = append(sqlVars, eventID)
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, " ", nil)
	if err != nil {
		//var message = "unable to join SQL statement string"
		slog.Error("application_event_repository.GetByID: unable to join SQL statement string", "error", err)
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

	var results []*models.ApplicationEvent
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "getByID")
		if err != nil {
			slog.Error("application_event_repository.getByID: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing event data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("application_event_repository.getByID: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading EventCompanies from database: " + err.Error())
	}

	return results, nil
}

// GetAll can return InternalServiceError
func (repository *ApplicationEventRepository) GetAll() ([]*models.ApplicationEvent, error) {
	sqlSelect := `
		SELECT application_id, event_id, created_date 
		FROM application_event 
		ORDER BY created_date DESC; `

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.ApplicationEvent
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "GetAll")
		if err != nil {
			slog.Error("application_event_repository.GetAll: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing event data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("application_event_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading EventCompanies from database: " + err.Error())
	}

	return results, nil
}

// Delete can return InternalServiceError, NotFoundError
func (repository *ApplicationEventRepository) Delete(model *models.DeleteApplicationEvent) error {
	sqlDelete := `
		DELETE FROM application_event 
		WHERE application_id = ? 
		AND event_id = ?; `

	result, err := repository.database.Exec(sqlDelete, model.ApplicationID, model.EventID)
	if err != nil {
		slog.Error(
			"application_event_repository.Delete: Error trying to delete ApplicationEvent",
			"applicationID", model.ApplicationID,
			"eventID", model.EventID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"application_event_repository.Delete: Error trying to delete ApplicationEvent",
			"applicationID", model.ApplicationID,
			"eventID", model.EventID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError(
			"ApplicationEvent does not exist. applicationID: " + model.ApplicationID.String() +
				", eventID: " + model.EventID.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

// mapRow can return InternalServiceError
func (repository *ApplicationEventRepository) mapRow(scanner interface{ Scan(...interface{}) error },
	methodName string) (*models.ApplicationEvent, error) {

	var result models.ApplicationEvent
	var createdDate sql.NullString

	err := scanner.Scan(&result.ApplicationID, &result.EventID, &createdDate)

	if err != nil {
		return nil, err
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, createdDate.String)
		if err != nil {
			slog.Error("application_event_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	return &result, nil
}
