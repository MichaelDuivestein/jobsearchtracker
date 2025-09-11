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

type ApplicationHandler struct {
	applicationService *services.ApplicationService
}

func NewApplicationHandler(applicationService *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{applicationService: applicationService}
}

func (applicationHandler *ApplicationHandler) CreateApplication(writer http.ResponseWriter, request *http.Request) {
	var createApplicationRequest requests.CreateApplicationRequest
	if err := json.NewDecoder(request.Body).Decode(&createApplicationRequest); err != nil {
		slog.Info("v1.ApplicationHandler.CreateApplication: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createApplicationModel, err := createApplicationRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.ApplicationHandler.CreateApplication: Unable to convert CreateApplicationRequest to model", "error",
			err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createApplicationModel == nil {
		slog.Info("v1.ApplicationHandler.CreateApplication: CreateApplicationModel is nil", "error", err)
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	createdApplication, err := applicationHandler.applicationService.CreateApplication(createApplicationModel)
	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info("v1.ApplicationHandler.CreateApplication: ConflictError creating application", "error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while creating application"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationHandler.CreateApplication: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.ApplicationHandler.CreateApplication: ValidationError while creating application", "error",
				err)
		} else {
			errorMessage = "Unknown internal error while creating application"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationHandler.CreateApplication: Error while creating application", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	applicationResponse, err := responses.NewApplicationResponse(createdApplication)
	if err != nil {
		slog.Error(
			"v1.ApplicationHandler.CreateApplication: Unable to convert internal model to response", "error",
			err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(applicationResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.ApplicationHandler.CreateApplication: Unable to write response", "error", err)
		http.Error(writer, "Application created but unable to create response", http.StatusInternalServerError)

		return
	}
}

func (applicationHandler *ApplicationHandler) GetApplicationByID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	applicationIDStr := vars["id"]

	if applicationIDStr == "" {
		slog.Info("v1.ApplicationHandler.GetApplicationById: application ID is empty")
		http.Error(writer, "application ID is empty", http.StatusBadRequest)
		return
	}

	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		slog.Info("v1.ApplicationHandler.GetApplicationById: application ID is not a valid UUID")
		http.Error(writer, "application ID is not a valid UUID", http.StatusBadRequest)
		return
	}

	var internalServiceError *internalErrors.InternalServiceError
	var notFoundError *internalErrors.NotFoundError
	var validationErr *internalErrors.ValidationError

	// can return InternalServiceError, NotFoundError, ValidationError
	application, err := applicationHandler.applicationService.GetApplicationById(&applicationID)
	if err != nil {
		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while retrieving application"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationHandler.GetApplicationByID: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "application not found"
			status = http.StatusNotFound
			slog.Info("v1.ApplicationHandler.GetApplicationByID: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.ApplicationHandler.GetApplicationByID: Validation error", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	applicationResponse, err := responses.NewApplicationResponse(application)
	if err != nil {
		slog.Error(
			"v1.ApplicationHandler.GetApplicationByID: Unable to convert internal model to response", "error",
			err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(applicationResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.ApplicationHandler.GetApplicationByID: Unable to write response", "error", err)
		http.Error(writer, "Application found but unable to build response", http.StatusInternalServerError)

		return
	}

	slog.Info(
		"v1.ApplicationHandler.GetApplicationByID: retrieved application successfully",
		"application.ID", application.ID.String())

	return
}

func (applicationHandler *ApplicationHandler) GetApplicationsByJobTitle(
	writer http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)
	jobTitle := vars["title"]

	if jobTitle == "" {
		slog.Info("v1.ApplicationHandler.GetApplicationByJobTitle: job title is empty")
		http.Error(writer, "job title is empty", http.StatusBadRequest)
		return
	}

	var internalServiceError *internalErrors.InternalServiceError
	var notFoundError *internalErrors.NotFoundError
	var validationErr *internalErrors.ValidationError

	applications, err := applicationHandler.applicationService.GetApplicationsByJobTitle(&jobTitle)
	if err != nil {
		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while retrieving applications"
			status = http.StatusInternalServerError
			slog.Error("v1.ApplicationHandler.GetApplicationsByJobTitle: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "No applications [partially] matching this job title found"
			status = http.StatusNotFound
			slog.Info("v1.ApplicationHandler.GetApplicationsByJobTitle: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.ApplicationHandler.GetApplicationsByJobTitle: Validation error", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	applicationsResponse, err := responses.NewApplicationsResponse(applications)
	if err != nil {
		slog.Error(
			"v1.ApplicationHandler.GetApplicationsByJobTitle: Unable to convert internal model to response", "error",
			err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(applicationsResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.ApplicationHandler.GetApplicationsByJobTitle: Unable to write response", "error", err)
		http.Error(writer, "Application found but unable to build response", http.StatusInternalServerError)

		return
	}

	slog.Info(
		"v1.ApplicationHandler.GetApplicationsByJobTitle: retrieved applications successfully",
		"jobTitle", jobTitle)

	return
}

func (applicationHandler *ApplicationHandler) GetAllApplications(writer http.ResponseWriter, request *http.Request) {
	// can return InternalServiceError
	applications, err := applicationHandler.applicationService.GetAllApplications()
	if err != nil {
		errorMessage := "Internal service error while getting all applications"
		status := http.StatusInternalServerError
		slog.Error("v1.ApplicationHandler.getAllApplications: "+errorMessage, "error", err)

		http.Error(writer, errorMessage, status)
		return
	}

	//  can return InternalServiceError
	applicationsResponse, err := responses.NewApplicationsResponse(applications)
	if err != nil {
		slog.Error(
			"v1.ApplicationHandler.GetAllApplications: Unable to convert internal model to response",
			"error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(applicationsResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.ApplicationHandler.GetAllApplications: Unable to write response", "error", err)
		http.Error(writer, "Applications retrieved but unable to create response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.ApplicationHandler.GetAllApplications: retrieved all applications successfully")

	return
}
