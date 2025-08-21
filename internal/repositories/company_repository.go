package repositories

import (
	"database/sql"
	"github.com/google/uuid"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"
)

type CompanyRepository struct {
	database *sql.DB
}

func NewCompanyRepository(database *sql.DB) *CompanyRepository {
	return &CompanyRepository{database: database}
}

// Create can return ConflictError, InternalServiceError
func (repository *CompanyRepository) Create(company *models.CreateCompany) (*models.Company, error) {
	sqlInsert :=
		"INSERT INTO company (id, name, company_type, notes, last_contact, created_date, updated_date) " +
			"VALUES (?, ?, ?, ?, ?, ?, ?) " +
			"RETURNING id, name, company_type, notes, last_contact, created_date, updated_date"

	var companyID uuid.UUID
	if company.ID != nil {
		companyID = *company.ID
	} else {
		companyID = uuid.New()
	}

	var lastContact, createdDate, updatedDate interface{}

	if company.LastContact != nil {
		lastContact = company.LastContact.Format(time.RFC3339)
	}

	if company.CreatedDate != nil {
		createdDate = company.CreatedDate.Format(time.RFC3339)
	} else {
		createdDate = time.Now()
	}

	if company.UpdatedDate != nil {
		updatedDate = company.UpdatedDate.Format(time.RFC3339)
	}

	var result models.Company
	var lastContactStr, createdDateStr, updatedDateStr sql.NullString

	err := repository.database.QueryRow(
		sqlInsert,
		companyID,
		company.Name,
		company.CompanyType,
		company.Notes,
		lastContact,
		createdDate,
		updatedDate,
	).Scan(
		&result.ID, &result.Name, &result.CompanyType, &result.Notes,
		&lastContactStr, &createdDateStr, &updatedDateStr,
	)

	if err != nil {
		if err.Error() == "constraint failed: UNIQUE constraint failed: company.id (1555)" {
			slog.Info("company_repository.createCompany: UNIQUE constraint failed", "companyID", companyID.String())
			return nil, internalErrors.NewConflictError("companyID already exists in database: '" + companyID.String() + "'")
		}
		slog.Error("company_repository.CreateCompany: error trying to insert", "error", err, "companyID", companyID.String())
		return nil, internalErrors.NewInternalServiceError(err.Error())
	}

	parsedTime, err := parseTimeFromDB(lastContactStr)
	if err != nil {
		slog.Error("company_repository.CreateCompany: error trying to parse lastContact", "lastContact", lastContactStr, "companyID", companyID.String())
		return nil, internalErrors.NewInternalServiceError(err.Error())
	}
	result.LastContact = parsedTime

	parsedTime, err = parseTimeFromDB(createdDateStr)
	if err != nil {
		slog.Error("company_repository.CreateCompany: error trying to parse createdDate", "createdDate", lastContactStr, "companyID", companyID.String())
		return nil, internalErrors.NewInternalServiceError(err.Error())
	}
	result.CreatedDate = *parsedTime

	parsedTime, err = parseTimeFromDB(updatedDateStr)
	if err != nil {
		slog.Error("company_repository.CreateCompany: error trying to parse updatedDate", "updatedDate", lastContactStr, "companyID", companyID.String())
		return nil, internalErrors.NewInternalServiceError(err.Error())
	}
	result.UpdatedDate = parsedTime

	return &result, nil
}

// GetById can return InternalServiceError, NotFoundError, ValidationError
func (repository *CompanyRepository) GetById(id *uuid.UUID) (*models.Company, error) {
	if id == nil {
		slog.Info("company_repository.GetById: ID is nil")
		var id = "ID"
		return nil, internalErrors.NewValidationError(&id, "ID is nil")
	}

	sqlSelect :=
		"SELECT id, name, company_type, notes, last_contact, created_date, updated_date " +
			"FROM company " +
			"WHERE id = ?"

	var result models.Company
	var lastContact, createdDate, updatedDate sql.NullString

	err := repository.database.QueryRow(sqlSelect, id).Scan(
		&result.ID,
		&result.Name,
		&result.CompanyType,
		&result.Notes,
		&lastContact,
		&createdDate,
		&updatedDate,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			slog.Info("company_repository.GetById: No result found for ID", "ID", id, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + id.String() + "'")
		}
		return nil, err
	}

	if lastContact.Valid {
		timestamp, err := time.Parse(time.RFC3339, lastContact.String)
		if err != nil {
			slog.Error("company_repository.GetById: Error parsing lastContact", "lastContact", lastContact, "error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing lastContact: " + err.Error())
		}
		result.LastContact = &timestamp
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, createdDate.String)
		if err != nil {
			slog.Error("company_repository.GetById: Error parsing createdDate", "createdDate", createdDate, "error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	if updatedDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, updatedDate.String)
		if err != nil {
			slog.Error("company_repository.GetById: Error parsing updatedDate", "updatedDate", updatedDate, "error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing updatedDate: " + err.Error())
		}
		result.UpdatedDate = &timestamp
	}

	return &result, nil
}

func parseTimeFromDB(timeString sql.NullString) (*time.Time, error) {
	if !timeString.Valid {
		return nil, nil
	}

	parsedTime, err := time.Parse(time.RFC3339, timeString.String)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}
