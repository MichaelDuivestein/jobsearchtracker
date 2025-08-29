package repositories

import (
	"database/sql"
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
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

	row := repository.database.QueryRow(
		sqlInsert,
		companyID,
		company.Name,
		company.CompanyType,
		company.Notes,
		lastContact,
		createdDate,
		updatedDate,
	)

	// can return ConflictError, InternalServiceError
	result, err := repository.mapRow(row, "Create", &companyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("company_repository.GetById: No result found for ID", "ID", companyID, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + companyID.String() + "'")
		}
		return nil, err
	}

	return result, err
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

	row := repository.database.QueryRow(sqlSelect, id)

	// can return ConflictError, InternalServiceError
	result, err := repository.mapRow(row, "GetById", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("company_repository.GetById: No result found for ID", "ID", id, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + id.String() + "'")
		}
		return nil, err
	}

	return result, err
}

// mapRow can return ConflictError, InternalServiceError
func (repository *CompanyRepository) mapRow(scanner interface{ Scan(...interface{}) error }, methodName string, ID *uuid.UUID) (*models.Company, error) {
	var result models.Company
	var lastContact, createdDate, updatedDate sql.NullString

	err := scanner.Scan(
		&result.ID,
		&result.Name,
		&result.CompanyType,
		&result.Notes,
		&lastContact,
		&createdDate,
		&updatedDate,
	)

	if err != nil {
		if err.Error() == "constraint failed: UNIQUE constraint failed: company.id (1555)" {
			var IDString string
			if ID != nil {
				IDString = ID.String()
			} else {
				IDString = "[not set]"
			}
			slog.Info(
				"company_repository.createCompany: UNIQUE constraint failed",
				"ID", IDString)
			return nil, internalErrors.NewConflictError(
				"ID already exists in database: '" + IDString + "'")
		}

		return nil, err
	}

	if lastContact.Valid {
		timestamp, err := time.Parse(time.RFC3339, lastContact.String)
		if err != nil {
			slog.Error(
				"company_repository."+methodName+": Error parsing lastContact",
				"lastContact", lastContact,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing lastContact: " + err.Error())
		}
		result.LastContact = &timestamp
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, createdDate.String)
		if err != nil {
			slog.Error("company_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	if updatedDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, updatedDate.String)
		if err != nil {
			slog.Error("company_repository."+methodName+": Error parsing updatedDate",
				"updatedDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing updatedDate: " + err.Error())
		}
		result.UpdatedDate = &timestamp
	}

	return &result, nil
}
