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

type EventHandler struct {
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// CreateEvent creates an event and returns it
//
// @Summary create an event
// @Description create an `event` and return it
// @Tags event
// @Accept json
// @Produce json
// @Param event body requests.CreateEventRequest true "Create Event request"
// @Success 201 {object} responses.EventResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/event/new [post]
func (eventHandler *EventHandler) CreateEvent(writer http.ResponseWriter, request *http.Request) {
	var createEventRequest requests.CreateEventRequest
	if err := json.NewDecoder(request.Body).Decode(&createEventRequest); err != nil {
		slog.Info("v1.EventHandler.CreateEvent: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createEventModel, err := createEventRequest.ToModel()
	if err != nil {
		slog.Info("v1.EventHandler.CreateEvent: Unable to convert CreateEventRequest to model", "error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createEventModel == nil {
		slog.Info("v1.EventHandler.CreateEvent: CreateEventModel is nil", "error", err)
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	createdEvent, err := eventHandler.eventService.CreateEvent(createEventModel)

	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info("v1.EventHandler.CreateEvent: ConflictError creating event", "error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while creating event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventHandler.CreateEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.EventHandler.CreateEvent: ValidationError while creating event", "error", err)
		} else {
			errorMessage = "Unknown internal error while creating event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventHandler.CreateEvent: Error while creating event", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	eventResponse, err := responses.NewEventResponse(createdEvent)
	if err != nil {
		slog.Error("v1.EventHandler.CreateEvent: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(eventResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.EventHandler.CreateEvent: Unable to write response", "error", err)
		http.Error(writer, "Event created but unable to create response", http.StatusInternalServerError)

		return
	}
}

// GetEventByID retrieves an event matching input UUID
//
// @Summary Get an event by ID
// @Description Get an `event` by ID
// @Tags event
// @Produce json
// @Param id path string true "Event ID" format(uuid)
// @Success 200 {object} responses.EventResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/event/get/id/{id} [get]
func (eventHandler *EventHandler) GetEventByID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	eventIDStr := vars["id"]

	if eventIDStr == "" {
		slog.Info("v1.EventHandler.GetEventById: event ID is empty")
		http.Error(writer, "event ID is empty", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		slog.Info("v1.EventHandler.GetEventById: event ID is not a valid UUID")
		http.Error(writer, "event ID is not a valid UUID", http.StatusBadRequest)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	event, err := eventHandler.eventService.GetEventByID(&eventID)
	if err != nil {
		var internalServiceError *internalErrors.InternalServiceError
		var notFoundError *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while retrieving event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventHandler.GetEventByID: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "event not found"
			status = http.StatusNotFound
			slog.Info("v1.EventHandler.GetEventByID: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.EventHandler.GetEventByID: Validation error", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	eventResponse, err := responses.NewEventResponse(event)
	if err != nil {
		slog.Error("v1.EventHandler.GetEventByID: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(eventResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.EventHandler.GetEventByID: Unable to write response", "error", err)
		http.Error(writer, "Event found but unable to build response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.EventHandler.GetEventByID: retrieved event successfully", "event.ID", event.ID.String())
}

// GetAllEvents retrieves all events.
//
// @Summary Get all events
// @Description Get all `event`s
// @Description - include_applications=all: Returns `application`s with all fields
// @Description - include_applications=ids: Returns `application`s with only `id`, `application_id`, and `recruiter_id`
// @Description - include_applications=none: No `application` data included (default)
// @Description - include_companies=all: Returns `company`s with all fields
// @Description - include_companies=ids: Returns `company`s with only `id`
// @Description - include_companies=none: No `company` data included (default)
// @Description - include_persons=all: Returns `person`s with all fields
// @Description - include_persons=ids: Returns `person`s with only `id`
// @Description - include_persons=none: No `person` data included (default)
// @Tags event
// @Produce json
// @Success 200 {array} responses.EventResponse
// @Failure 400
// @Failure 500
// @Router /v1/event/get/all [get]
func (eventHandler *EventHandler) GetAllEvents(writer http.ResponseWriter, request *http.Request) {

	query := request.URL.Query()

	includeApplications, err := GetExtraDataTypeParam(query.Get("include_applications"))
	if err != nil {
		slog.Error("v1.CompanyHandler.GetAllCompanies: Could not parse include_applications param", "error", err)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(
			writer,
			"Invalid value for include_applications. Accepted params are 'all', 'ids', and 'none'",
			status)
		return
	}

	includeCompanies, err := GetExtraDataTypeParam(query.Get("include_companies"))
	if err != nil {
		slog.Error("v1.CompanyHandler.GetAllCompanies: Could not parse include_companies param", "error", err)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(
			writer,
			"Invalid value for include_companies. Accepted params are 'all', 'ids', and 'none'",
			status)
		return
	}

	includePersons, err := GetExtraDataTypeParam(query.Get("include_persons"))
	if err != nil {
		slog.Error("v1.CompanyHandler.GetAllCompanies: Could not parse include_persons param", "error", err)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(
			writer,
			"Invalid value for include_persons. Accepted params are 'all', 'ids', and 'none'",
			status)
		return
	}

	// can return InternalServiceError
	events, err := eventHandler.eventService.GetAllEvents(
		*includeApplications,
		*includeCompanies,
		*includePersons)
	if err != nil {
		errorMessage := "Internal service error while getting all events"
		status := http.StatusInternalServerError
		slog.Error("v1.EventHandler.getAllEvents: "+errorMessage, "error", err)

		http.Error(writer, errorMessage, status)
		return
	}

	//  can return InternalServiceError
	eventsResponse, err := responses.NewEventsResponse(events)
	if err != nil {
		slog.Error("v1.EventHandler.GetAllEvents: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(eventsResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.EventHandler.GetAllEvents: Unable to write response", "error", err)
		http.Error(writer, "Events retrieved but unable to create response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.EventHandler.GetAllEvents: retrieved all events successfully")
}

// UpdateEvent updates an event
//
// @Summary update an event
// @Description update an `event`
// @Tags event
// @Accept json
// @Produce json
// @Param event body requests.UpdateEventRequest true "Update Event Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/event/update [post]
func (eventHandler *EventHandler) UpdateEvent(writer http.ResponseWriter, request *http.Request) {
	var updateEventRequest requests.UpdateEventRequest
	if err := json.NewDecoder(request.Body).Decode(&updateEventRequest); err != nil {
		slog.Info("v1.EventHandler.UpdateEvent: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	updateEventModel, err := updateEventRequest.ToModel()
	if err != nil {
		slog.Info("v1.EventHandler.UpdateEvent: Unable to convert UpdateEventRequest to model", "error", err)
		http.Error(writer, "Unable to convert request to internal model: "+err.Error(), http.StatusBadRequest)

		return
	}

	if updateEventModel == nil {
		slog.Error(
			"v1.EventHandler.UpdateEvent: updateEventModel is nil after attempting to convert request to internal model")
		http.Error(writer, "Unable to convert request to model: Internal model is nil.", http.StatusBadRequest)
		return
	}

	// can return InternalServiceError, ValidationError
	err = eventHandler.eventService.UpdateEvent(updateEventModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while updating event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventHandler.UpdateEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.EventHandler.UpdateEvent: ValidationError while updating event", "error", err)
		} else {
			errorMessage = "Unknown internal error while updating event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventHandler.UpdateEvent: Error while updating event", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}

// DeleteEvent deletes an `event` matching input UUID
//
// @Summary Delete an event by ID
// @Description Delete an `event` by ID
// @Tags event
// @Param id path string true "Event ID" format(uuid)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/event/delete/{id} [delete]
func (eventHandler *EventHandler) DeleteEvent(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	eventIDStr := vars["id"]

	if eventIDStr == "" {
		slog.Info("v1.EventHandler.DeleteEvent: event ID is empty")
		http.Error(writer, "event ID is empty", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		slog.Info("v1.EventHandler.DeleteEvent: event ID is not a valid UUID")
		http.Error(writer, "event ID is not a valid UUID", http.StatusBadRequest)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = eventHandler.eventService.DeleteEvent(&eventID)
	if err != nil {
		var internalServiceError *internalErrors.InternalServiceError
		var notFoundError *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while deleting event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventHandler.DeleteEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "Event not found"
			status = http.StatusNotFound
			slog.Info("v1.EventHandler.DeleteEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.EventHandler.DeleteEvent: ValidationError while deleting event", "error", err)
		} else {
			errorMessage = "Unknown internal error while creating event"
			status = http.StatusInternalServerError
			slog.Error("v1.EventHandler.DeleteEvent: Error while deleting event", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
