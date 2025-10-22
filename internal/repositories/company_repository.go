package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/utils"
	"log/slog"
	"strconv"
	"strings"
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
			"RETURNING id, name, company_type, notes, last_contact, created_date, updated_date, null, null"

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
		createdDate = time.Now().Format(time.RFC3339)
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
			slog.Info("company_repository.Create: No result found for ID", "ID", companyID, "error", err.Error())
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
		"SELECT id, name, company_type, notes, last_contact, created_date, updated_date, null, null " +
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

// GetAllByName can return InternalServiceError, NotFoundError, ValidationError
func (repository *CompanyRepository) GetAllByName(name *string) ([]*models.Company, error) {
	if name == nil {
		slog.Info("company_repository.GetAllByName: name is nil")
		return nil, internalErrors.NewValidationError(nil, "name is nil")
	}

	sqlSelect :=
		"SELECT id, name, company_type, notes, last_contact, created_date, updated_date, null, null " +
			"FROM company " +
			"WHERE name LIKE ? " +
			"ORDER BY name ASC"

	wildcardName := "%" + *name + "%"
	rows, err := repository.database.Query(sqlSelect, wildcardName)
	if err != nil {
		return nil, err
	}

	var results []*models.Company

	for rows.Next() {
		// can return ConflictError, InternalServiceError
		result, err := repository.mapRow(rows, "GetAllByName", nil)
		if err != nil {
			slog.Error("company_repository.GetAllByName: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing company data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("company_repository.GetAllByName: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError("Error reading companies from database: " + err.Error())
	}

	if len(results) == 0 {
		slog.Info("company_repository.GetAllByName: No result found for Name", "Name", name)
		return nil, internalErrors.NewNotFoundError("Name: '" + *name + "'")
	}

	return results, nil
}

// GetAll can return InternalServiceError
func (repository *CompanyRepository) GetAll(
	includeApplications models.IncludeExtraDataType,
	includePersons models.IncludeExtraDataType) ([]*models.Company, error) {

	sqlSelect := `
		SELECT c.id, c.name, c.company_type, c.notes, c.last_contact, c.created_date, c.updated_date, %s, %s
		FROM company c
		%s %s
		GROUP BY c.id, c.name, c.company_type
		ORDER BY c.created_date DESC;
		`

	applicationsCoalesceString := "null \n"
	applicationsJoinString := ""
	if includeApplications != models.IncludeExtraDataTypeNone {
		applicationsCoalesceString, applicationsJoinString =
			repository.buildApplicationsCoalesceAndJoin(includeApplications)
	}

	personsCoalesceString := "null \n"
	personsJoinString := ""
	if includePersons != models.IncludeExtraDataTypeNone {
		personsCoalesceString, personsJoinString =
			repository.buildPersonsCoalesceAndJoin(includePersons)
	}

	sqlSelect = fmt.Sprintf(
		sqlSelect,
		applicationsCoalesceString,
		personsCoalesceString,
		applicationsJoinString,
		personsJoinString)

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.Company
	for rows.Next() {
		// can return ConflictError, InternalServiceError
		result, err := repository.mapRow(rows, "GetAll", nil)
		if err != nil {
			slog.Error("company_repository.GetAll: Error mapping row", "error", err)
			return nil, internalErrors.NewInternalServiceError("Error processing company data: " + err.Error())
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if err = rows.Err(); err != nil {
		slog.Error("company_repository.GetAll: Error iterating rows", "error", err)
		return nil, internalErrors.NewInternalServiceError("Error reading companies from database: " + err.Error())
	}

	return results, nil
}

// Update can return InternalServiceError, ValidationError
func (repository *CompanyRepository) Update(company *models.UpdateCompany) error {
	var sqlParts []string
	var sqlVars []interface{}

	var sqlString strings.Builder
	sqlString.WriteString("UPDATE company SET ")
	sqlString.WriteString("updated_date = ?, ")
	sqlVars = append(sqlVars, time.Now().Format(time.RFC3339))

	updateItemCount := 0

	if company.Name != nil {
		sqlParts = append(sqlParts, "name = ?")
		sqlVars = append(sqlVars, *company.Name)
		updateItemCount++
	}

	if company.CompanyType != nil {
		sqlParts = append(sqlParts, "company_type = ?")
		sqlVars = append(sqlVars, *company.CompanyType)
		updateItemCount++
	}

	if company.Notes != nil {
		sqlParts = append(sqlParts, "notes = ?")
		sqlVars = append(sqlVars, *company.Notes)
		updateItemCount++
	}

	if company.LastContact != nil {
		sqlParts = append(sqlParts, "last_contact = ?")
		sqlVars = append(sqlVars, company.LastContact.Format(time.RFC3339))
		updateItemCount++
	}

	if updateItemCount == 0 {
		slog.Info("company_repository.Update: nothing to update", "id", company.ID)
		return internalErrors.NewValidationError(nil, "nothing to update")
	}

	sqlPayload, err := utils.JoinToString(&sqlParts, nil, ", ", nil)
	if err != nil {
		var message = "unable to join SQL statement string"
		slog.Error("company_repository.Update: unable to join SQL statement string", "error", err)
		return internalErrors.NewInternalServiceError(message)
	}

	sqlString.WriteString(sqlPayload)

	sqlString.WriteString(" WHERE id = ?")
	sqlVars = append(sqlVars, company.ID)

	_, err = repository.database.Exec(
		sqlString.String(),
		sqlVars...,
	)

	if err != nil {
		slog.Error("company_repository.Update: unable to update company", "id", company.ID, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	return err
}

// Delete can return InternalServiceError, NotFoundError, ValidationError
func (repository *CompanyRepository) Delete(id *uuid.UUID) error {
	if id == nil {
		slog.Error("company_repository.Delete: ID is nil")
		id := "ID"
		return internalErrors.NewValidationError(&id, "ID is nil")
	}

	sqlDelete := "DELETE FROM company WHERE id = ?"

	result, err := repository.database.Exec(sqlDelete, id)
	if err != nil {
		slog.Error("company_repository.Delete: Error trying to delete company", "id", id, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("company_repository.Delete: Error trying to delete company", "id", id, "error", err.Error())
		return internalErrors.NewInternalServiceError(err.Error())
	}
	if rowsAffected == 0 {
		return internalErrors.NewNotFoundError("Company does not exist. ID: " + id.String())
	} else if rowsAffected > 1 {
		return internalErrors.NewInternalServiceError(
			"Unexpected number of rows affected: " + strconv.FormatInt(rowsAffected, 10))
	}

	return nil
}

// internal functions

// mapRow can return ConflictError, InternalServiceError
func (repository *CompanyRepository) mapRow(
	scanner interface{ Scan(...interface{}) error }, methodName string, ID *uuid.UUID) (*models.Company, error) {

	var result models.Company
	var lastContact, createdDate, updatedDate, applicationsString, personsString sql.NullString

	err := scanner.Scan(
		&result.ID,
		&result.Name,
		&result.CompanyType,
		&result.Notes,
		&lastContact,
		&createdDate,
		&updatedDate,
		&applicationsString,
		&personsString,
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
		result.CreatedDate = &timestamp
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

	if applicationsString.Valid {
		var applications []*models.Application
		if err := json.NewDecoder(strings.NewReader(applicationsString.String)).Decode(&applications); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing applications: " + err.Error())
		}

		if len(applications) > 0 {
			result.Applications = &applications
		}
	}

	if personsString.Valid {
		var persons []*models.Person
		if err := json.NewDecoder(strings.NewReader(personsString.String)).Decode(&persons); err != nil {
			return nil, internalErrors.NewInternalServiceError("Error parsing persons: " + err.Error())
		}

		if len(persons) > 0 {
			result.Persons = &persons
		}
	}

	return &result, nil
}

func (repository *CompanyRepository) buildApplicationsCoalesceAndJoin(
	includeApplications models.IncludeExtraDataType) (string, string) {

	if includeApplications == models.IncludeExtraDataTypeNone {
		return "", ""
	}

	coalesceString := `
		COALESCE(
			JSON_GROUP_ARRAY(
				JSON_OBJECT(
					'ID', a.id,
					'CompanyID', a.company_id,
					'RecruiterID', a.recruiter_id%s
				) ORDER BY a.created_date DESC
			) FILTER (WHERE a.id IS NOT NULL),
			JSON_ARRAY()
		) as applications
	`

	allColumns := ""
	if includeApplications == models.IncludeExtraDataTypeAll {
		allColumns = `,
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

	joinString := "		LEFT JOIN application a ON (c.id = a.company_id OR c.id = a.recruiter_id) \n"

	return coalesceString, joinString
}

func (repository *CompanyRepository) buildPersonsCoalesceAndJoin(
	includePersons models.IncludeExtraDataType) (string, string) {

	if includePersons == models.IncludeExtraDataTypeNone {
		return "", ""
	}

	coalesceString := `
		COALESCE(
			JSON_GROUP_ARRAY(
				JSON_OBJECT(
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

	joinString :=
		`LEFT JOIN company_person cp ON (cp.company_id = c.id)
		LEFT JOIN person p ON (cp.person_id = p.id)
`

	return coalesceString, joinString
}
