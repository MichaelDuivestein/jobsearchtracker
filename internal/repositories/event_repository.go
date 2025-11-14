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
		RETURNING id, event_type, description, notes, event_date, created_date, updated_date`

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
		SELECT id, event_type, description, notes, event_date, created_date, updated_date
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
func (repository *EventRepository) GetAll() ([]*models.Event, error) {
	sqlSelect := `
		SELECT id, event_type, description, notes, event_date, created_date, updated_date
		FROM event 
		ORDER BY event_date DESC`

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
	var eventDate, createdDate, updatedDate sql.NullString

	err := scanner.Scan(
		&result.ID,
		&result.EventType,
		&result.Description,
		&result.Notes,
		&eventDate,
		&createdDate,
		&updatedDate)
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

	return &result, nil
}
