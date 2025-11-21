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

type CompanyEventHandler struct {
	companyEventService *services.CompanyEventService
}

func NewCompanyEventHandler(companyEventService *services.CompanyEventService) *CompanyEventHandler {
	return &CompanyEventHandler{companyEventService: companyEventService}
}

// AssociateCompanyEvent associates a company with an event and returns it
//
// @Summary associate a company with an event
// @Description associate a `company` with a `event` and return it
// @Tags companyEvent
// @Accept json
// @Produce json
// @Param company body requests.AssociateCompanyEventRequest true "Associate Company Event request"
// @Success 201 {object} responses.CompanyEventResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company-event/associate [post]
func (handler *CompanyEventHandler) AssociateCompanyEvent(writer http.ResponseWriter, request *http.Request) {
	var createCompanyEventRequest requests.AssociateCompanyEventRequest
	if err := json.NewDecoder(request.Body).Decode(&createCompanyEventRequest); err != nil {
		slog.Info("v1.CompanyEventHandler.AssociateCompanyEvent: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createCompanyEventModel, err := createCompanyEventRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.CompanyEventHandler.AssociateCompanyEvent: Unable to convert CreateCompanyEventRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createCompanyEventModel == nil {
		slog.Info("v1.CompanyEventHandler.AssociateCompanyEvent: CreateCompanyEvent model is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	companyEventModel, err := handler.companyEventService.AssociateCompanyEvent(createCompanyEventModel)

	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info("v1.CompanyEventHandler.AssociateCompanyEvent: ConflictError creating record", "error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while associating event to company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyEventHandler.AssociateCompanyEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.CompanyEventHandler.AssociateCompanyEvent: ValidationError while associating event to company",
				"error", err)
		} else {
			errorMessage = "Unknown internal error while associating event to company"
			status = http.StatusInternalServerError
			slog.Error(
				"v1.CompanyEventHandler.AssociateCompanyEvent: Error while associating event to company",
				"error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	response := responses.NewCompanyEventResponse(companyEventModel)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.EventHandler.AssociateCompanyEvent: Unable to write response", "error", err)
		http.Error(writer, "Event created but unable to create response", http.StatusInternalServerError)

		return
	}
}

// GetCompanyEventsByID retrieves a company matching input company UUID and/or the input event UUID. `company-id` AND/OR `event-id` must be provided.
//
// @Summary Get companyEvents by ID
// @Description Get `companyEvent`s by ID
// @Tags companyEvent
// @Produce json
// @Param company-id query string false "company ID" format(uuid)
// @Param event-id query string false "event ID" format(uuid)
// @Success 200 {array} responses.CompanyEventResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company-event/get/ [get]
func (handler *CompanyEventHandler) GetCompanyEventsByID(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	companyIDString := query.Get("company-id")
	eventIDString := query.Get("event-id")

	if companyIDString == "" && eventIDString == "" {
		errorMessage := "CompanyID and/or EventID are required"
		slog.Info("v1.CompanyEventHandler.GetCompanyEventsByID: " + errorMessage)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	var companyID, eventID *uuid.UUID = nil, nil

	if companyIDString != "" {
		companyIDValue, err := uuid.Parse(companyIDString)
		if err != nil || companyIDValue == uuid.Nil {
			errorMessage := "Unable to parse CompanyID"
			slog.Info("v1.CompanyEventHandler.GetCompanyEventsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}

		companyID = &companyIDValue
	}

	if eventIDString != "" {
		eventIDValue, err := uuid.Parse(eventIDString)
		if err != nil || eventIDValue == uuid.Nil {
			errorMessage := "Unable to parse EventID"
			slog.Info("v1.CompanyEventHandler.GetCompanyEventsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}
		eventID = &eventIDValue
	}

	companyEvents, err := handler.companyEventService.GetByID(companyID, eventID)
	if err != nil {
		errorMessage := "Internal service error while getting companyEvents by ID"
		slog.Error("v1.CompanyEventHandler.GetCompanyEventsByID: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	response := responses.NewCompanyEventsResponse(companyEvents)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		slog.Error("v1.CompanyEventHandler.GetCompanyEventsByID: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.CompanyEventHandler.GetCompanyEventsByID: retrieved all companies successfully")
}

// GetAllCompanyEvents retrieves all companyEvents.
//
// @Summary Get all companyEvents
// @Description Get all `companyEvent`s
// @Tags companyEvent
// @Produce json
// @Success 200 {array} responses.CompanyEventResponse
// @Failure 400
// @Failure 500
// @Router /v1/company-event/get/all [get]
func (handler *CompanyEventHandler) GetAllCompanyEvents(writer http.ResponseWriter, request *http.Request) {
	companyEvents, err := handler.companyEventService.GetAll()
	if err != nil {
		errorMessage := "Internal service error while getting all companyEvents"
		slog.Error("v1.CompanyHandler.GetAllCompanyEvents: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	companyEventsResponse := responses.NewCompanyEventsResponse(companyEvents)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(companyEventsResponse)
	if err != nil {
		slog.Error("v1.CompanyEventHandler.GetAllCompanyEvents: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.CompanyEventHandler.GetAllCompanyEvents: retrieved all CompanyEvents successfully")
}

// DeleteCompanyEvent deletes a `companyEvent` matching input company UUID and event UUID
//
// @Summary Delete a companyEvent by company UUID and event UUID
// @Description Delete a `companyEvent` by company UUID and event UUID
// @Tags companyEvent
// @Param company-id query string true "company ID" format(uuid)
// @Param event-id query string true "event ID" format(uuid)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company-event/delete [delete]
func (handler *CompanyEventHandler) DeleteCompanyEvent(writer http.ResponseWriter, request *http.Request) {
	var deleteRequest requests.DeleteCompanyEventRequest
	if err := json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
		slog.Info("v1.CompanyEventHandler.DeleteCompanyEvent: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	deleteModel, err := deleteRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.CompanyEventHandler.DeleteCompanyEvent: Unable to convert DeleteCompanyEventRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if deleteModel == nil {
		slog.Info("v1.CompanyEventHandler.AssociateCompanyEvent: DeleteCompanyEvent is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = handler.companyEventService.Delete(deleteModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var notFoundErr *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while deleting CompanyEvent"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyEventHandler.DeleteCompanyEvent: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundErr) {
			errorMessage = err.Error()
			status = http.StatusNotFound
			slog.Info(
				"v1.CompanyEventHandler.DeleteCompanyEvent: NotFoundErr while deleting CompanyEvent", "error",
				err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.CompanyEventHandler.DeleteCompanyEvent: ValidationError while deleting CompanyEvent", "error",
				err)
		} else {
			errorMessage = "Unknown internal error while creating company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyEventHandler.DeleteCompanyEvent: Error while deleting CompanyEvent", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
}
