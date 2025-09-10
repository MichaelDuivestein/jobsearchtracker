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

type PersonRepository struct {
	database *sql.DB
}

func NewPersonRepository(database *sql.DB) *PersonRepository {
	return &PersonRepository{database: database}
}

// Create can return ConflictError, InternalServiceError
func (repository *PersonRepository) Create(person *models.CreatePerson) (*models.Person, error) {
	sqlInsert :=
		"INSERT INTO person (id, name, person_type, email, phone, notes, created_date, updated_date) " +
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?) " +
			"RETURNING id, name, person_type, email, phone, notes, created_date, updated_date"

	var personID uuid.UUID
	if person.ID != nil {
		personID = *person.ID
	} else {
		personID = uuid.New()
	}

	var createdDate, updatedDate interface{}

	if person.CreatedDate != nil {
		createdDate = person.CreatedDate.Format(time.RFC3339)
	} else {
		createdDate = time.Now()
	}

	if person.UpdatedDate != nil {
		updatedDate = person.UpdatedDate.Format(time.RFC3339)
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
	result, err := repository.mapRow(row, "Create", &personID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

	sqlSelect :=
		"SELECT id, name, person_type, email, phone, notes, created_date, updated_date " +
			"FROM person " +
			"WHERE id = ?"

	row := repository.database.QueryRow(sqlSelect, id)
	result, err := repository.mapRow(row, "GetById", id)
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

	sqlSelect :=
		"SELECT id, name, person_type, email, phone, notes, created_date, updated_date " +
			"FROM person " +
			"WHERE name LIKE ?" +
			"ORDER BY name ASC"

	rows, err := repository.database.Query(sqlSelect, wildcardName)
	if err != nil {
		return nil, err
	}

	var results []*models.Person

	for rows.Next() {
		result, err := repository.mapRow(rows, "GetAllByName", nil)
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
func (repository *PersonRepository) GetAll() ([]*models.Person, error) {
	sqlSelect :=
		"SELECT id, name, person_type, email, phone, notes, created_date, updated_date " +
			"FROM person " +
			"ORDER BY name ASC"

	rows, err := repository.database.Query(sqlSelect)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var results []*models.Person

	for rows.Next() {
		result, err := repository.mapRow(rows, "GetAll", nil)
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

// mapRow can return InternalServiceError
func (repository *PersonRepository) mapRow(
	scanner interface{ Scan(...interface{}) error },
	methodName string,
	ID *uuid.UUID) (*models.Person, error) {
	var result models.Person
	var createdDate, updatedDate sql.NullString

	err := scanner.Scan(
		&result.ID,
		&result.Name,
		&result.PersonType,
		&result.Email,
		&result.Phone,
		&result.Notes,
		&createdDate,
		&updatedDate,
	)

	if err != nil {
		if err.Error() == "constraint failed: UNIQUE constraint failed: person.id (1555)" {
			var IDString string
			if ID != nil {
				IDString = ID.String()
			} else {
				IDString = "[not set]"
			}
			slog.Info(
				"person_repository.createPerson: UNIQUE constraint failed",
				"ID", IDString)
			return nil, internalErrors.NewConflictError(
				"ID already exists in database: '" + IDString + "'")
		}
		return nil, err
	}

	if createdDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, createdDate.String)
		if err != nil {
			slog.Error("person_repository."+methodName+": Error parsing createdDate",
				"createdDate", createdDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing createdDate: " + err.Error())
		}
		result.CreatedDate = timestamp
	}

	if updatedDate.Valid {
		timestamp, err := time.Parse(time.RFC3339, updatedDate.String)
		if err != nil {
			slog.Error("person_repository."+methodName+": Error parsing updatedDate",
				"updatedDate", updatedDate,
				"error", err.Error())
			return nil, internalErrors.NewInternalServiceError("Error parsing updatedDate: " + err.Error())
		}
		result.UpdatedDate = &timestamp
	}

	return &result, nil
}
