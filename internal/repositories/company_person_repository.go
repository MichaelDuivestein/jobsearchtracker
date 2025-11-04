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

type CompanyPersonRepository struct {
	database *sql.DB
}

func NewCompanyPersonRepository(database *sql.DB) *CompanyPersonRepository {
	return &CompanyPersonRepository{database: database}
}

// AssociateCompanyPerson can return ConflictError, InternalServiceError
func (repository *CompanyPersonRepository) AssociateCompanyPerson(
	associateModel *models.AssociateCompanyPerson) (*models.CompanyPerson, error) {

	sqlInsert := `
		INSERT INTO company_person (
			company_id, person_id, created_date
		) VALUES (?, ?, ?) 
		RETURNING company_id, person_id, created_date; `

	var createdDate string
	if associateModel.CreatedDate != nil {
		createdDate = associateModel.CreatedDate.Format(timeutil.RFC3339Milli_Write)
	} else {
		createdDate = time.Now().UTC().Format(timeutil.RFC3339Milli_Write)
	}

	row := repository.database.QueryRow(
		sqlInsert,
		associateModel.CompanyID,
		associateModel.PersonID,
		createdDate,
	)

	if row.Err() != nil {
		if row.Err().Error() ==
			"constraint failed: UNIQUE constraint failed: company_person.company_id, company_person.person_id (1555)" {

			slog.Info(
				"company_person_repository.associateToCompany: UNIQUE constraint failed",
				"company_id", associateModel.CompanyID,
				"person_id", associateModel.PersonID)

			return nil, internalErrors.NewConflictError(
				"CompanyID and PersonID combination already exists in database.")
		} else if row.Err().Error() == "constraint failed: FOREIGN KEY constraint failed (787)" {
			// TODO: Use foreign key constraint names (in 0003_add_application.up.sql) once modernc.org/sqlite
			// supports it.
			slog.Info("company_person_repository.Create: FOREIGN KEY constraint failed (787)")
			return nil, internalErrors.NewValidationError(nil, "Foreign key does not exist")
		}
		return nil, row.Err()
	}

	// can return InternalServiceError
	result, err := repository.mapRow(row, "Create")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("company_person_repository.create: No result found.", "error", err.Error())
			return nil, internalErrors.NewNotFoundError("Unable to map CompanyPerson")
		}
		return nil, err
	}

	return result, nil
}

// GetByID can return ValidationError, InternalServiceError
func (repository *CompanyPersonRepository) GetByID(
	companyID *uuid.UUID, personID *uuid.UUID) ([]*models.CompanyPerson, error) {

	if (companyID == nil || *companyID == uuid.Nil) && (personID == nil || *personID == uuid.Nil) {
		return nil, internalErrors.NewValidationError(nil, "companyID and personID cannot both be empty")
	}

	var sqlString strings.Builder
	var sqlParts []string
	var sqlVars []interface{}

	sqlString.WriteString(`
		SELECT company_id, person_id, created_date 
		FROM company_person 
		WHERE `)

	companyIDAdded := false
	if companyID != nil && *companyID != uuid.Nil {
		sqlParts = append(sqlParts, "company_id = ? ")
		sqlVars = append(sqlVars, companyID)
		companyIDAdded = true
	}

	if personID != nil && *personID != uuid.Nil {
		if companyIDAdded {
			sqlParts = append(sqlParts, "\n\t\tAND ")
		}
		sqlParts = append(sqlParts, "person_id = ? ")
		sqlVars = append(sqlVars, personID)
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, " ", nil)
	if err != nil {
		//var message = "unable to join SQL statement string"
		slog.Error("company_person_repository.GetByID: unable to join SQL statement string", "error", err)
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

	var results []*models.CompanyPerson
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "getByID")
		if err != nil {
			slog.Error("company_person_repository.getByID: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("company_person_repository.getByID: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading PersonCompanies from database: " + err.Error())
	}

	return results, nil
}

// GetAll can return InternalServiceError
func (repository *CompanyPersonRepository) GetAll() ([]*models.CompanyPerson, error) {
	sqlSelect := `
		SELECT company_id, person_id, created_date 
		FROM company_person 
		ORDER BY created_date DESC; `

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.CompanyPerson
	for rows.Next() {
		// mapRow can return InternalServiceError
		result, err := repository.mapRow(rows, "GetAll")
		if err != nil {
			slog.Error("company_person_repository.GetAll: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing person data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("company_person_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError(
			"Error reading PersonCompanies from database: " + err.Error())
	}

	return results, nil
}

// Delete can return InternalServiceError, NotFoundError
func (repository *CompanyPersonRepository) Delete(model *models.DeleteCompanyPerson) error {
	sqlDelete := `
		DELETE FROM company_person 
		WHERE company_id = ? 
		AND person_id = ?; `

	result, err := repository.database.Exec(sqlDelete, model.CompanyID, model.PersonID)
	if err != nil {
		slog.Error(
			"company_person_repository.Delete: Error trying to delete CompanyPerson",
			"companyID", model.CompanyID,
			"personID", model.PersonID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"company_person_repository.Delete: Error trying to delete CompanyPerson",
			"companyID", model.CompanyID,
			"personID", model.PersonID,
			"error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError(
			"CompanyPerson does not exist. companyID: " + model.CompanyID.String() +
				", personID: " + model.PersonID.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

// mapRow can return InternalServiceError
func (repository *CompanyPersonRepository) mapRow(scanner interface{ Scan(...interface{}) error },
	methodName string) (*models.CompanyPerson, error) {

	var result models.CompanyPerson
	var createdDate sql.NullString

	err := scanner.Scan(&result.CompanyID, &result.PersonID, &createdDate)

	if err != nil {
		return nil, err
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(timeutil.RFC3339Milli_Read, createdDate.String)
		if err != nil {
			slog.Error("company_person_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	return &result, nil
}
