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

type ApplicationEventHandler struct {
	applicationEventService *services.ApplicationEventService
}

func NewApplicationEventHandler(applicationEventService *services.ApplicationEventService) *ApplicationEventHandler {
	return &ApplicationEventHandler{applicationEventService: applicationEventService}
}

// AssociateApplicationEvent associates an application with an event and returns it
//
// @Summary associate an application with an event
// @Description associate an `application` with a `event` and return it
// @Tags applicationEvent
// @Accept json
// @Produce json
// @Param application body requests.AssociateApplicationEventRequest true "Associate Application Event request"
// @Success 201 {object} responses.ApplicationEventResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/application-event/associate [post]
func (handler *ApplicationEventHandler) AssociateApplicationEvent(writer http.ResponseWriter, request *http.Request) {
	var createApplicationEventRequest requests.AssociateApplicationEventRequest
	if err := json.NewDecoder(request.Body).Decode(&createApplicationEventRequest); err != nil {
		slog.Info("v1.ApplicationEventHandler.AssociateApplicationEvent: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createApplicationEventModel, err := createApplicationEventRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.ApplicationEventHandler.AssociateApplicationEvent: Unable to convert CreateApplicationEventRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createApplicationEventModel == nil {
		slog.Info("v1.ApplicationEventHandler.AssociateApplicationEvent: CreateApplicationEvent model is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	applicationEventModel, err := handler.applicationEventService.AssociateApplicationEvent(createApplicationEventModel)

	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info(
				"v1.ApplicationEventHandler.AssociateApplicationEvent: ConflictError creating record",
				"error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while associating event to application"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationEventHandler.AssociateApplicationEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.ApplicationEventHandler.AssociateApplicationEvent: ValidationError while associating event to application",
				"error", err)
		} else {
			errorMessage = "Unknown internal error while associating event to application"
			status = http.StatusInternalServerError
			slog.Error(
				"v1.ApplicationEventHandler.AssociateApplicationEvent: Error while associating event to application",
				"error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	response := responses.NewApplicationEventResponse(applicationEventModel)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.EventHandler.AssociateApplicationEvent: Unable to write response", "error", err)
		http.Error(writer, "Event created but unable to create response", http.StatusInternalServerError)

		return
	}
}

// GetApplicationEventsByID retrieves an application matching input application UUID and/or the input event UUID. `application-id` AND/OR `event-id` must be provided.
//
// @Summary Get applicationEvents by ID
// @Description Get `applicationEvent`s by ID
// @Tags applicationEvent
// @Produce json
// @Param application-id query string false "application ID" format(uuid)
// @Param event-id query string false "event ID" format(uuid)
// @Success 200 {array} responses.ApplicationEventResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/application-event/get/ [get]
func (handler *ApplicationEventHandler) GetApplicationEventsByID(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	applicationIDString := query.Get("application-id")
	eventIDString := query.Get("event-id")

	if applicationIDString == "" && eventIDString == "" {
		errorMessage := "ApplicationID and/or EventID are required"
		slog.Info("v1.ApplicationEventHandler.GetApplicationEventsByID: " + errorMessage)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	var applicationID, eventID *uuid.UUID = nil, nil

	if applicationIDString != "" {
		applicationIDValue, err := uuid.Parse(applicationIDString)
		if err != nil || applicationIDValue == uuid.Nil {
			errorMessage := "Unable to parse ApplicationID"
			slog.Info("v1.ApplicationEventHandler.GetApplicationEventsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}

		applicationID = &applicationIDValue
	}

	if eventIDString != "" {
		eventIDValue, err := uuid.Parse(eventIDString)
		if err != nil || eventIDValue == uuid.Nil {
			errorMessage := "Unable to parse EventID"
			slog.Info("v1.ApplicationEventHandler.GetApplicationEventsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}
		eventID = &eventIDValue
	}

	applicationEvents, err := handler.applicationEventService.GetByID(applicationID, eventID)
	if err != nil {
		errorMessage := "Internal service error while getting applicationEvents by ID"
		slog.Error("v1.ApplicationEventHandler.GetApplicationEventsByID: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	response := responses.NewApplicationEventsResponse(applicationEvents)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		slog.Error("v1.ApplicationEventHandler.GetApplicationEventsByID: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.ApplicationEventHandler.GetApplicationEventsByID: retrieved all applications successfully")
}

// GetAllApplicationEvents retrieves all applicationEvents.
//
// @Summary Get all applicationEvents
// @Description Get all `applicationEvent`s
// @Tags applicationEvent
// @Produce json
// @Success 200 {array} responses.ApplicationEventResponse
// @Failure 400
// @Failure 500
// @Router /v1/application-event/get/all [get]
func (handler *ApplicationEventHandler) GetAllApplicationEvents(writer http.ResponseWriter, request *http.Request) {
	applicationEvents, err := handler.applicationEventService.GetAll()
	if err != nil {
		errorMessage := "Internal service error while getting all applicationEvents"
		slog.Error("v1.ApplicationHandler.GetAllApplicationEvents: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	applicationEventsResponse := responses.NewApplicationEventsResponse(applicationEvents)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(applicationEventsResponse)
	if err != nil {
		slog.Error("v1.ApplicationEventHandler.GetAllApplicationEvents: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.ApplicationEventHandler.GetAllApplicationEvents: retrieved all ApplicationEvents successfully")
}

// DeleteApplicationEvent deletes a `applicationEvent` matching input application UUID and event UUID
//
// @Summary Delete an applicationEvent by application UUID and event UUID
// @Description Delete a `applicationEvent` by application UUID and event UUID
// @Tags applicationEvent
// @Param application-id query string true "application ID" format(uuid)
// @Param event-id query string true "event ID" format(uuid)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/application-event/delete [delete]
func (handler *ApplicationEventHandler) DeleteApplicationEvent(writer http.ResponseWriter, request *http.Request) {
	var deleteRequest requests.DeleteApplicationEventRequest
	if err := json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
		slog.Info("v1.ApplicationEventHandler.DeleteApplicationEvent: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	deleteModel, err := deleteRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.ApplicationEventHandler.DeleteApplicationEvent: Unable to convert DeleteApplicationEventRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if deleteModel == nil {
		slog.Info("v1.ApplicationEventHandler.AssociateApplicationEvent: DeleteApplicationEvent is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = handler.applicationEventService.Delete(deleteModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var notFoundErr *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while deleting ApplicationEvent"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationEventHandler.DeleteApplicationEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundErr) {
			errorMessage = err.Error()
			status = http.StatusNotFound
			slog.Info(
				"v1.ApplicationEventHandler.DeleteApplicationEvent: NotFoundErr while deleting ApplicationEvent", "error",
				err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.ApplicationEventHandler.DeleteApplicationEvent: ValidationError while deleting ApplicationEvent", "error",
				err)
		} else {
			errorMessage = "Unknown internal error while creating application"
			status = http.StatusInternalServerError
			slog.Error(
				"v1.ApplicationEventHandler.DeleteApplicationEvent: Error while deleting ApplicationEvent",
				"error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
