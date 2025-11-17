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

type EventPersonHandler struct {
	eventPersonService *services.EventPersonService
}

func NewEventPersonHandler(eventPersonService *services.EventPersonService) *EventPersonHandler {
	return &EventPersonHandler{eventPersonService: eventPersonService}
}

// AssociateEventPerson associates an event with a person and returns it
//
// @Summary associate an event with a person
// @Description associate an `event` with a `person` and return it
// @Tags eventPerson
// @Accept json
// @Produce json
// @Param event body requests.AssociateEventPersonRequest true "Associate Event Person request"
// @Success 201 {object} responses.EventPersonResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/event-person/associate [post]
func (handler *EventPersonHandler) AssociateEventPerson(writer http.ResponseWriter, request *http.Request) {
	var createEventPersonRequest requests.AssociateEventPersonRequest
	if err := json.NewDecoder(request.Body).Decode(&createEventPersonRequest); err != nil {
		slog.Info("v1.EventPersonHandler.AssociateEventPerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createEventPersonModel, err := createEventPersonRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.EventPersonHandler.AssociateEventPerson: Unable to convert CreateEventPersonRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createEventPersonModel == nil {
		slog.Info("v1.EventPersonHandler.AssociateEventPerson: CreateEventPerson model is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	eventPersonModel, err := handler.eventPersonService.AssociateEventPerson(createEventPersonModel)

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
				"v1.EventPersonHandler.AssociateEventPerson: ConflictError creating record",
				"error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while associating person to event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventPersonHandler.AssociateEventPerson: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.EventPersonHandler.AssociateEventPerson: ValidationError while associating person to event",
				"error", err)
		} else {
			errorMessage = "Unknown internal error while associating person to event"
			status = http.StatusInternalServerError
			slog.Error(
				"v1.EventPersonHandler.AssociateEventPerson: Error while associating person to event",
				"error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	response := responses.NewEventPersonResponse(eventPersonModel)

	writer.Header().Set("Content-Type", "event/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.PersonHandler.AssociateEventPerson: Unable to write response", "error", err)
		http.Error(writer, "Person created but unable to create response", http.StatusInternalServerError)

		return
	}
}

// GetEventPersonsByID retrieves an event matching input event UUID and/or the input person UUID. `event-id` AND/OR `person-id` must be provided.
//
// @Summary Get eventPersons by ID
// @Description Get `eventPerson`s by ID
// @Tags eventPerson
// @Produce json
// @Param event-id query string false "event ID" format(uuid)
// @Param person-id query string false "person ID" format(uuid)
// @Success 200 {array} responses.EventPersonResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/event-person/get/ [get]
func (handler *EventPersonHandler) GetEventPersonsByID(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	eventIDString := query.Get("event-id")
	personIDString := query.Get("person-id")

	if eventIDString == "" && personIDString == "" {
		errorMessage := "EventID and/or PersonID are required"
		slog.Info("v1.EventPersonHandler.GetEventPersonsByID: " + errorMessage)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	var eventID, personID *uuid.UUID = nil, nil

	if eventIDString != "" {
		eventIDValue, err := uuid.Parse(eventIDString)
		if err != nil || eventIDValue == uuid.Nil {
			errorMessage := "Unable to parse EventID"
			slog.Info("v1.EventPersonHandler.GetEventPersonsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}

		eventID = &eventIDValue
	}

	if personIDString != "" {
		personIDValue, err := uuid.Parse(personIDString)
		if err != nil || personIDValue == uuid.Nil {
			errorMessage := "Unable to parse PersonID"
			slog.Info("v1.EventPersonHandler.GetEventPersonsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}
		personID = &personIDValue
	}

	eventPersons, err := handler.eventPersonService.GetByID(eventID, personID)
	if err != nil {
		errorMessage := "Internal service error while getting eventPersons by ID"
		slog.Error("v1.EventPersonHandler.GetEventPersonsByID: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	response := responses.NewEventPersonsResponse(eventPersons)

	writer.Header().Set("Content-Type", "event/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		slog.Error("v1.EventPersonHandler.GetEventPersonsByID: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.EventPersonHandler.GetEventPersonsByID: retrieved all events successfully")
}

// GetAllEventPersons retrieves all eventPersons.
//
// @Summary Get all eventPersons
// @Description Get all `eventPerson`s
// @Tags eventPerson
// @Produce json
// @Success 200 {array} responses.EventPersonResponse
// @Failure 400
// @Failure 500
// @Router /v1/event-person/get/all [get]
func (handler *EventPersonHandler) GetAllEventPersons(writer http.ResponseWriter, request *http.Request) {
	eventPersons, err := handler.eventPersonService.GetAll()
	if err != nil {
		errorMessage := "Internal service error while getting all eventPersons"
		slog.Error("v1.EventHandler.GetAllEventPersons: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	eventPersonsResponse := responses.NewEventPersonsResponse(eventPersons)

	writer.Header().Set("Content-Type", "event/json")
	err = json.NewEncoder(writer).Encode(eventPersonsResponse)
	if err != nil {
		slog.Error("v1.EventPersonHandler.GetAllEventPersons: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.EventPersonHandler.GetAllEventPersons: retrieved all EventPersons successfully")
}

// DeleteEventPerson deletes a `eventPerson` matching input event UUID and person UUID
//
// @Summary Delete an eventPerson by event UUID and person UUID
// @Description Delete a `eventPerson` by event UUID and person UUID
// @Tags eventPerson
// @Param event-id query string true "event ID" format(uuid)
// @Param person-id query string true "person ID" format(uuid)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/event-person/delete [delete]
func (handler *EventPersonHandler) DeleteEventPerson(writer http.ResponseWriter, request *http.Request) {
	var deleteRequest requests.DeleteEventPersonRequest
	if err := json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
		slog.Info("v1.EventPersonHandler.DeleteEventPerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	deleteModel, err := deleteRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.EventPersonHandler.DeleteEventPerson: Unable to convert DeleteEventPersonRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if deleteModel == nil {
		slog.Info("v1.EventPersonHandler.AssociateEventPerson: DeleteEventPerson is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = handler.eventPersonService.Delete(deleteModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var notFoundErr *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while deleting EventPerson"
			status = http.StatusInternalServerError
			slog.Error("v1.EventPersonHandler.DeleteEventPerson: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundErr) {
			errorMessage = err.Error()
			status = http.StatusNotFound
			slog.Info(
				"v1.EventPersonHandler.DeleteEventPerson: NotFoundErr while deleting EventPerson", "error",
				err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.EventPersonHandler.DeleteEventPerson: ValidationError while deleting EventPerson", "error",
				err)
		} else {
			errorMessage = "Unknown internal error while creating event"
			status = http.StatusInternalServerError
			slog.Error(
				"v1.EventPersonHandler.DeleteEventPerson: Error while deleting EventPerson",
				"error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
