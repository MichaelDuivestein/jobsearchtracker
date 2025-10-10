package handlers

import (
	"encoding/json"
	"errors"
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/api/v1/responses"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/services"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PersonHandler struct {
	personService *services.PersonService
}

func NewPersonHandler(personService *services.PersonService) *PersonHandler {
	return &PersonHandler{personService: personService}
}

// CreatePerson creates a person and returns it
//
// @Summary create a person
// @Description create a `person` and return it
// @Tags person
// @Accept json
// @Produce json
// @Param person body requests.CreatePersonRequest true "Create Person request"
// @Success 201 {object} responses.PersonResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/person/new [post]
func (personHandler *PersonHandler) CreatePerson(writer http.ResponseWriter, request *http.Request) {
	var createPersonRequest requests.CreatePersonRequest
	if err := json.NewDecoder(request.Body).Decode(&createPersonRequest); err != nil {
		slog.Info("v1.PersonHandler.CreatePerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createPersonModel, err := createPersonRequest.ToModel()
	if err != nil {
		slog.Info("v1.PersonHandler.CreatePerson: Unable to convert CreatePersonRequest to model", "error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createPersonModel == nil {
		slog.Info("v1.PersonHandler.CreatePerson: CreatePersonModel is nil", "error", err)
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	createdPerson, err := personHandler.personService.CreatePerson(createPersonModel)
	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info("v1.PersonHandler.CreatePerson: ConflictError creating person", "error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while creating person"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.CreatePerson: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.PersonHandler.CreatePerson: ValidationError while creating person", "error", err)
		} else {
			errorMessage = "Unknown internal error while creating person"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.CreatePerson: Error while creating person", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	personResponse, err := responses.NewPersonResponse(createdPerson)
	if err != nil {
		slog.Error("v1.PersonHandler.CreatePerson: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(personResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.PersonHandler.CreatePerson: Unable to write response", "error", err)
		http.Error(writer, "Person created but unable to create response", http.StatusInternalServerError)

		return
	}
}

// GetPersonByID retrieves a person matching input UUID
//
// @Summary Get a person by ID
// @Description Get a `person` by ID
// @Tags person
// @Produce json
// @Param id path string true "Person ID" format(uuid)
// @Success 200 {object} responses.PersonResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/person/get/id/{id} [get]
func (personHandler *PersonHandler) GetPersonByID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	personIDStr := vars["id"]

	if personIDStr == "" {
		slog.Info("v1.PersonHandler.GetPersonById: person ID is empty")
		http.Error(writer, "person ID is empty", http.StatusBadRequest)
		return
	}

	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		slog.Info("v1.PersonHandler.GetPersonById: person ID is not a valid UUID")
		http.Error(writer, "person ID is not a valid UUID", http.StatusBadRequest)
		return
	}

	var internalServiceError *internalErrors.InternalServiceError
	var notFoundError *internalErrors.NotFoundError
	var validationErr *internalErrors.ValidationError

	// can return InternalServiceError, NotFoundError, ValidationError
	person, err := personHandler.personService.GetPersonById(&personID)
	if err != nil {
		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while retrieving person"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.GetPersonByID: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "person not found"
			status = http.StatusNotFound
			slog.Info("v1.PersonHandler.GetPersonByID: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.PersonHandler.GetPersonByID: Validation error", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	personResponse, err := responses.NewPersonResponse(person)
	if err != nil {
		slog.Error("v1.PersonHandler.GetPersonByID: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(personResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.PersonHandler.GetPersonByID: Unable to write response", "error", err)
		http.Error(writer, "Person found but unable to build response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.PersonHandler.GetPersonByID: retrieved person successfully", "person.ID", person.ID.String())
}

// GetPersonsByName retrieves `person`s which fully, or partially, match the input name
//
// @Summary Get persons by name
// @Description Get `person`s which fully, or partially, match the input name
// @Tags person
// @Produce json
// @Param name path string true "Person Name"
// @Success 200 {array} responses.PersonResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/person/get/name/{name} [get]
func (personHandler *PersonHandler) GetPersonsByName(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	personName := vars["name"]

	if personName == "" {
		slog.Info("v1.PersonHandler.GetPersonByName: person Name is empty")
		http.Error(writer, "person Name is empty", http.StatusBadRequest)
		return
	}

	var internalServiceError *internalErrors.InternalServiceError
	var notFoundError *internalErrors.NotFoundError
	var validationErr *internalErrors.ValidationError

	persons, err := personHandler.personService.GetPersonsByName(&personName)
	if err != nil {
		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while retrieving persons"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.GetPersonsByName: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "No people [partially] matching this name found"
			status = http.StatusNotFound
			slog.Info("v1.PersonHandler.GetPersonsByName: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.PersonHandler.GetPersonsByName: Validation error", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	personsResponse, err := responses.NewPersonsResponse(persons)
	if err != nil {
		slog.Error("v1.PersonHandler.GetPersonsByName: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(personsResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.PersonHandler.GetPersonsByName: Unable to write response", "error", err)
		http.Error(writer, "Person found but unable to build response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.PersonHandler.GetPersonsByName: retrieved persons successfully", "name", personName)
}

// GetAllPersons retrieves all persons.
//
// @Summary Get all persons
// @Description Get all `person`s
// @Tags person
// @Produce json
// @Success 200 {array} responses.PersonResponse
// @Failure 400
// @Failure 500
// @Router /v1/person/get/all [get]
func (personHandler *PersonHandler) GetAllPersons(writer http.ResponseWriter, request *http.Request) {
	// can return InternalServiceError
	persons, err := personHandler.personService.GetAllPersons()
	if err != nil {
		errorMessage := "Internal service error while getting all persons"
		status := http.StatusInternalServerError
		slog.Error("v1.PersonHandler.getAllPersons: "+errorMessage, "error", err)

		http.Error(writer, errorMessage, status)
		return
	}

	//  can return InternalServiceError
	personsResponse, err := responses.NewPersonsResponse(persons)
	if err != nil {
		slog.Error("v1.PersonHandler.GetAllPersons: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(personsResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.PersonHandler.GetAllPersons: Unable to write response", "error", err)
		http.Error(writer, "Persons retrieved but unable to create response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.PersonHandler.GetAllPersons: retrieved all persons successfully")
}

// UpdatePerson updates a person
//
// @Summary update a person
// @Description update a `person`
// @Tags person
// @Accept json
// @Produce json
// @Param person body requests.UpdatePersonRequest true "Update Company Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/person/update [post]
func (personHandler *PersonHandler) UpdatePerson(writer http.ResponseWriter, request *http.Request) {
	var updatePersonRequest requests.UpdatePersonRequest
	if err := json.NewDecoder(request.Body).Decode(&updatePersonRequest); err != nil {
		slog.Info("v1.PersonHandler.UpdatePerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	updatePersonModel, err := updatePersonRequest.ToModel()
	if err != nil {
		slog.Info("v1.PersonHandler.UpdatePerson: Unable to convert UpdatePersonRequest to model", "error", err)
		http.Error(writer, "Unable to convert request to internal model: "+err.Error(), http.StatusBadRequest)

		return
	}

	if updatePersonModel == nil {
		slog.Error(
			"v1.PersonHandler.UpdatePerson: updatePersonModel is nil after attempting to convert request to internal model")
		http.Error(writer, "Unable to convert request to model: Internal model is nil.", http.StatusBadRequest)
		return
	}

	// can return InternalServiceError, ValidationError
	err = personHandler.personService.UpdatePerson(updatePersonModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while updating person"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.UpdatePerson: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.PersonHandler.UpdatePerson: ValidationError while updating person", "error", err)
		} else {
			errorMessage = "Unknown internal error while updating person"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.UpdatePerson: Error while updating person", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}

// DeletePerson deletes a `person` matching input UUID
//
// @Summary Delete a person by ID
// @Description Delete a `person` by ID
// @Tags person
// @Param id path string true "Person ID" format(uuid)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/person/delete/{id} [delete]
func (personHandler *PersonHandler) DeletePerson(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	personIDStr := vars["id"]

	if personIDStr == "" {
		slog.Info("v1.PersonHandler.DeletePerson: person ID is empty")
		http.Error(writer, "person ID is empty", http.StatusBadRequest)
		return
	}

	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		slog.Info("v1.PersonHandler.DeletePerson: person ID is not a valid UUID")
		http.Error(writer, "person ID is not a valid UUID", http.StatusBadRequest)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = personHandler.personService.DeletePerson(&personID)
	if err != nil {
		var internalServiceError *internalErrors.InternalServiceError
		var notFoundError *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while deleting person"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.DeletePerson: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "Person not found"
			status = http.StatusNotFound
			slog.Info("v1.PersonHandler.DeletePerson: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.PersonHandler.DeletePerson: ValidationError while deleting person", "error", err)
		} else {
			errorMessage = "Unknown internal error while creating person"
			status = http.StatusInternalServerError
			slog.Error("v1.PersonHandler.DeletePerson: Error while deleting person", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
