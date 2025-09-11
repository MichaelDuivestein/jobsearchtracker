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

type ApplicationRepository struct {
	database *sql.DB
}

func NewApplicationRepository(database *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{database: database}
}

// Create can return ConflictError, InternalServiceError, ValidationError
func (repository *ApplicationRepository) Create(application *models.CreateApplication) (*models.Application, error) {
	sqlInsert := "INSERT INTO application (id, company_id, recruiter_id, job_title, job_ad_url, country, area, " +
		"remote_status_type, weekdays_in_office, estimated_cycle_time, estimated_commute_time, application_date, " +
		"created_date, updated_date)" +
		"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)" +
		"RETURNING id, company_id, recruiter_id, job_title, job_ad_url, country, area, remote_status_type, " +
		"weekdays_in_office, estimated_cycle_time, estimated_commute_time, application_date, created_date, " +
		"updated_date"

	var applicationID uuid.UUID
	if application.ID != nil {
		applicationID = *application.ID
	} else {
		applicationID = uuid.New()
	}

	var applicationDate, createdDate, updatedDate interface{}

	if application.ApplicationDate != nil {
		applicationDate = application.ApplicationDate.Format(time.RFC3339)
	}

	if application.CreatedDate != nil {
		createdDate = application.CreatedDate.Format(time.RFC3339)
	} else {
		createdDate = time.Now()
	}

	if application.UpdatedDate != nil {
		updatedDate = application.UpdatedDate.Format(time.RFC3339)
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

	result, err := repository.mapRow(row, "Create", &applicationID)
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

	sqlSelect := "SELECT id, company_id, recruiter_id, job_title, job_ad_url, country, area, remote_status_type, " +
		"weekdays_in_office, estimated_cycle_time, estimated_commute_time, application_date, created_date, " +
		"updated_date " +
		"FROM application " +
		"WHERE id = ?"

	row := repository.database.QueryRow(sqlSelect, id)

	// can return ConflictError, InternalServiceError
	result, err := repository.mapRow(row, "GetById", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("application_repository.GetById: No result found for ID", "ID", id, "error", err.Error())
			return nil, internalErrors.NewNotFoundError("ID: '" + id.String() + "'")
		}
		return nil, err
	}

	return result, nil
}

func (repository *ApplicationRepository) mapRow(
	scanner interface{ Scan(...interface{}) error }, methodName string, ID *uuid.UUID) (*models.Application, error) {

	var result models.Application
	var applicationDate, createdDate, updatedDate sql.NullString

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
	)

	if err != nil {
		if err.Error() == "constraint failed: UNIQUE constraint failed: application.id (1555)" {
			var IDString string
			if ID != nil {
				IDString = ID.String()
			} else {
				IDString = "[not set]"
			}
			slog.Info(
				"application_repository.createApplication: UNIQUE constraint failed",
				"ID", IDString)
			return nil, internalErrors.NewConflictError(
				"ID already exists in database: '" + IDString + "'")
		}
		return nil, err
	}

	if applicationDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, applicationDate.String)
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
		timestamp, err := time.Parse(time.RFC3339, createdDate.String)
		if err != nil {
			slog.Error("application_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	if updatedDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, updatedDate.String)
		if err != nil {
			slog.Error("application_repository."+methodName+": Error parsing updatedDate",
				"updatedDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing updatedDate: " + err.Error())
		}
		result.UpdatedDate = &timestamp
	}

	return &result, nil
}
