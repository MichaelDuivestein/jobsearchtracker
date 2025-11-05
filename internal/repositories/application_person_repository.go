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

type ApplicationPersonRepository struct {
	database *sql.DB
}

func NewApplicationPersonRepository(database *sql.DB) *ApplicationPersonRepository {
	return &ApplicationPersonRepository{database: database}
}

// AssociateApplicationPerson can return ConflictError, InternalServiceError
func (repository *ApplicationPersonRepository) AssociateApplicationPerson(
	associateModel *models.AssociateApplicationPerson) (*models.ApplicationPerson, error) {

	sqlInsert := `
		INSERT INTO application_person (
			application_id, person_id, created_date
		) VALUES (?, ?, ?) 
		RETURNING application_id, person_id, created_date; `

	var createdDate string
	if associateModel.CreatedDate != nil {
		createdDate = associateModel.CreatedDate.Format(timeutil.RFC3339Milli_Write)
	} else {
		createdDate = time.Now().UTC().Format(timeutil.RFC3339Milli_Write)
	}

	row := repository.database.QueryRow(
		sqlInsert,
		associateModel.ApplicationID,
		associateModel.PersonID,
		createdDate,
	)

	if row.Err() != nil {
		if row.Err().Error() ==
			"constraint failed: UNIQUE constraint failed: application_person.application_id, application_person.person_id (1555)" {

			slog.Info(
				"application_person_repository.associateToApplication: UNIQUE constraint failed",
				"application_id", associateModel.ApplicationID,
				"person_id", associateModel.PersonID)

			return nil, internalErrors.NewConflictError(
				"ApplicationID and PersonID combination already exists in database.")
		} else if row.Err().Error() == "constraint failed: FOREIGN KEY constraint failed (787)" {
			// TODO: Use foreign key constraint names (in 0003_add_application.up.sql) once modernc.org/sqlite
			// supports it.
			slog.Info("application_person_repository.Create: FOREIGN KEY constraint failed (787)")
			return nil, internalErrors.NewValidationError(nil, "Foreign key does not exist")
		}
		return nil, row.Err()
	}

	// can return InternalServiceError
	result, err := repository.mapRow(row, "Create")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("application_person_repository.create: No result found.", "error", err.Error())
			return nil, internalErrors.NewNotFoundError("Unable to map ApplicationPerson")
		}
		return nil, err
	}

	return result, nil
}

// GetByID can return ValidationError, InternalServiceError
func (repository *ApplicationPersonRepository) GetByID(
	applicationID *uuid.UUID, personID *uuid.UUID) ([]*models.ApplicationPerson, error) {

	if (applicationID == nil || *applicationID == uuid.Nil) && (personID == nil || *personID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "applicationID and personID cannot both be empty")
	}

	var sqlString strings.Builder
	var sqlParts []string
	var sqlVars []interface{}

	sqlString.WriteString(`
		SELECT application_id, person_id, created_date 
		FROM application_person 
		WHERE `)

	applicationIDAdded := false
	if applicationID != nil && *applicationID != uuid.Nil {
		sqlParts = append(sqlParts, "application_id = ? ")
		sqlVars = append(sqlVars, applicationID)
		applicationIDAdded = true
	}

	if personID != nil && *personID != uuid.Nil {
		if applicationIDAdded {
			sqlParts = append(sqlParts, "\n\t\tAND ")
		}
		sqlParts = append(sqlParts, "person_id = ? ")
		sqlVars = append(sqlVars, personID)
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, " ", nil)
	if err != nil {
		//var message = "unable to join SQL statement string"
		slog.Error("application_person_repository.GetByID: unable to join SQL statement string", "error", err)
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

	var results []*models.ApplicationPerson
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "getByID")
		if err != nil {
			slog.Error("application_person_repository.getByID: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("application_person_repository.getByID: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading PersonCompanies from database: " + err.Error())
	}

	return results, nil
}

// GetAll can return InternalServiceError
func (repository *ApplicationPersonRepository) GetAll() ([]*models.ApplicationPerson, error) {
	sqlSelect := `
		SELECT application_id, person_id, created_date 
		FROM application_person 
		ORDER BY created_date DESC; `

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.ApplicationPerson
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "GetAll")
		if err != nil {
			slog.Error("application_person_repository.GetAll: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("application_person_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading PersonCompanies from database: " + err.Error())
	}

	return results, nil
}

// Delete can return InternalServiceError, NotFoundError
func (repository *ApplicationPersonRepository) Delete(model *models.DeleteApplicationPerson) error {
	sqlDelete := `
		DELETE FROM application_person 
		WHERE application_id = ? 
		AND person_id = ?; `

	result, err := repository.database.Exec(sqlDelete, model.ApplicationID, model.PersonID)
	if err != nil {
		slog.Error(
			"application_person_repository.Delete: Error trying to delete ApplicationPerson",
			"applicationID", model.ApplicationID,
			"personID", model.PersonID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"application_person_repository.Delete: Error trying to delete ApplicationPerson",
			"applicationID", model.ApplicationID,
			"personID", model.PersonID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError(
			"ApplicationPerson does not exist. applicationID: " + model.ApplicationID.String() +
				", personID: " + model.PersonID.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

// mapRow can return InternalServiceError
func (repository *ApplicationPersonRepository) mapRow(scanner interface{ Scan(...interface{}) error },
	methodName string) (*models.ApplicationPerson, error) {

	var result models.ApplicationPerson
	var createdDate sql.NullString

	err := scanner.Scan(&result.ApplicationID, &result.PersonID, &createdDate)

	if err != nil {
		return nil, err
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, createdDate.String)
		if err != nil {
			slog.Error("application_person_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	return &result, nil
}
