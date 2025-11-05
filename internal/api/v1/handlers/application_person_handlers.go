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
)

type ApplicationPersonHandler struct {
	applicationPersonService *services.ApplicationPersonService
}

func NewApplicationPersonHandler(applicationPersonService *services.ApplicationPersonService) *ApplicationPersonHandler {
	return &ApplicationPersonHandler{applicationPersonService: applicationPersonService}
}

// AssociateApplicationPerson associates an application with a person and returns it
//
// @Summary associate an application with a person
// @Description associate a `application` with a `person` and return it
// @Tags applicationPerson
// @Accept json
// @Produce json
// @Param application body requests.AssociateApplicationPersonRequest true "Associate Application Person request"
// @Success 201 {object} responses.ApplicationPersonResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/application-person/associate [post]
func (handler *ApplicationPersonHandler) AssociateApplicationPerson(writer http.ResponseWriter, request *http.Request) {
	var createApplicationPersonRequest requests.AssociateApplicationPersonRequest
	if err := json.NewDecoder(request.Body).Decode(&createApplicationPersonRequest); err != nil {
		slog.Info("v1.ApplicationPersonHandler.AssociateApplicationPerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createApplicationPersonModel, err := createApplicationPersonRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.ApplicationPersonHandler.AssociateApplicationPerson: Unable to convert CreateApplicationPersonRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createApplicationPersonModel == nil {
		slog.Info("v1.ApplicationPersonHandler.AssociateApplicationPerson: CreateApplicationPerson model is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	applicationPersonModel, err := handler.applicationPersonService.AssociateApplicationPerson(createApplicationPersonModel)

	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info("v1.ApplicationPersonHandler.AssociateApplicationPerson: ConflictError creating record", "error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while associating person to application"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationPersonHandler.AssociateApplicationPerson: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.ApplicationPersonHandler.AssociateApplicationPerson: ValidationError while associating person to application",
				"error", err)
		} else {
			errorMessage = "Unknown internal error while associating person to application"
			status = http.StatusInternalServerError
			slog.Error(
				"v1.ApplicationPersonHandler.AssociateApplicationPerson: Error while associating person to application",
				"error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	response := responses.NewApplicationPersonResponse(applicationPersonModel)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.PersonHandler.AssociateApplicationPerson: Unable to write response", "error", err)
		http.Error(writer, "Person created but unable to create response", http.StatusInternalServerError)

		return
	}
}

// GetApplicationPersonsByID retrieves an application matching input application UUID and/or the input person UUID.  `application-id` AND/OR `person-id` must be provided.
//
// @Summary Get applicationPersons by ID
// @Description Get `applicationPerson`s by ID
// @Tags applicationPerson
// @Produce json
// @Param application-id query string false "application ID" format(uuid)
// @Param person-id query string false "person ID" format(uuid)
// @Success 200 {array} responses.ApplicationPersonResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/application-person/get/ [get]
func (handler *ApplicationPersonHandler) GetApplicationPersonsByID(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	applicationIDString := query.Get("application-id")
	personIDString := query.Get("person-id")

	if applicationIDString == "" && personIDString == "" {
		errorMessage := "ApplicationID and/or PersonID are required"
		slog.Info("v1.ApplicationPersonHandler.GetApplicationPersonsByID: " + errorMessage)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	var applicationID, personID *uuid.UUID = nil, nil

	if applicationIDString != "" {
		applicationIDValue, err := uuid.Parse(applicationIDString)
		if err != nil || applicationIDValue == uuid.Nil {
			errorMessage := "Unable to parse ApplicationID"
			slog.Info("v1.ApplicationPersonHandler.GetApplicationPersonsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}

		applicationID = &applicationIDValue
	}

	if personIDString != "" {
		personIDValue, err := uuid.Parse(personIDString)
		if err != nil || personIDValue == uuid.Nil {
			errorMessage := "Unable to parse PersonID"
			slog.Info("v1.ApplicationPersonHandler.GetApplicationPersonsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}
		personID = &personIDValue
	}

	applicationPersons, err := handler.applicationPersonService.GetByID(applicationID, personID)
	if err != nil {
		errorMessage := "Internal service error while getting applicationPersons by ID"
		slog.Error("v1.ApplicationPersonHandler.GetApplicationPersonsByID: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	response := responses.NewApplicationPersonsResponse(applicationPersons)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		slog.Error("v1.ApplicationPersonHandler.GetApplicationPersonsByID: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.ApplicationPersonHandler.GetApplicationPersonsByID: retrieved all applications successfully")
}

// GetAllApplicationPersons retrieves all applicationPersons.
//
// @Summary Get all applicationPersons
// @Description Get all `applicationPerson`s
// @Tags applicationPerson
// @Produce json
// @Success 200 {array} responses.ApplicationPersonResponse
// @Failure 400
// @Failure 500
// @Router /v1/application-person/get/all [get]
func (handler *ApplicationPersonHandler) GetAllApplicationPersons(writer http.ResponseWriter, request *http.Request) {
	applicationPersons, err := handler.applicationPersonService.GetAll()
	if err != nil {
		errorMessage := "Internal service error while getting all applicationPersons"
		slog.Error("v1.ApplicationHandler.GetAllApplicationPersons: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	applicationPersonsResponse := responses.NewApplicationPersonsResponse(applicationPersons)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(applicationPersonsResponse)
	if err != nil {
		slog.Error("v1.ApplicationPersonHandler.GetAllApplicationPersons: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.ApplicationPersonHandler.GetAllApplicationPersons: retrieved all ApplicationPersons successfully")
}

// DeleteApplicationPerson deletes a `applicationPerson` matching input application UUID and person UUID
//
// @Summary Delete an applicationPerson by application UUID and person UUID
// @Description Delete a `applicationPerson` by application UUID and person UUID
// @Tags applicationPerson
// @Param application-id query string true "application ID" format(uuid)
// @Param person-id query string true "person ID" format(uuid)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/application-person/delete [delete]
func (handler *ApplicationPersonHandler) DeleteApplicationPerson(writer http.ResponseWriter, request *http.Request) {
	var deleteRequest requests.DeleteApplicationPersonRequest
	if err := json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
		slog.Info("v1.ApplicationPersonHandler.DeleteApplicationPerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	deleteModel, err := deleteRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.ApplicationPersonHandler.DeleteApplicationPerson: Unable to convert DeleteApplicationPersonRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if deleteModel == nil {
		slog.Info("v1.ApplicationPersonHandler.AssociateApplicationPerson: DeleteApplicationPerson is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = handler.applicationPersonService.Delete(deleteModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var notFoundErr *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while deleting ApplicationPerson"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationPersonHandler.DeleteApplicationPerson: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundErr) {
			errorMessage = err.Error()
			status = http.StatusNotFound
			slog.Info(
				"v1.ApplicationPersonHandler.DeleteApplicationPerson: NotFoundErr while deleting ApplicationPerson", "error",
				err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.ApplicationPersonHandler.DeleteApplicationPerson: ValidationError while deleting ApplicationPerson", "error",
				err)
		} else {
			errorMessage = "Unknown internal error while creating application"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationPersonHandler.DeleteApplicationPerson: Error while deleting ApplicationPerson", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
