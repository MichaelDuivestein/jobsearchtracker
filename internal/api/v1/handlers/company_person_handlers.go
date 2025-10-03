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

type CompanyPersonHandler struct {
	companyPersonService *services.CompanyPersonService
}

func NewCompanyPersonHandler(companyPersonService *services.CompanyPersonService) *CompanyPersonHandler {
	return &CompanyPersonHandler{companyPersonService: companyPersonService}
}

func (handler *CompanyPersonHandler) AssociateCompanyPerson(writer http.ResponseWriter, request *http.Request) {
	var createCompanyPersonRequest requests.AssociateCompanyPersonRequest
	if err := json.NewDecoder(request.Body).Decode(&createCompanyPersonRequest); err != nil {
		slog.Info("v1.CompanyPersonHandler.AssociateCompanyPerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createCompanyPersonModel, err := createCompanyPersonRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.CompanyPersonHandler.AssociateCompanyPerson: Unable to convert CreateCompanyPersonRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if createCompanyPersonModel == nil {
		slog.Info("v1.CompanyPersonHandler.AssociateCompanyPerson: CreateCompanyPerson model is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	companyPersonModel, err := handler.companyPersonService.AssociateCompanyPerson(createCompanyPersonModel)

	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info("v1.CompanyPersonHandler.AssociateCompanyPerson: ConflictError creating record", "error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while associating person to company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyPersonHandler.AssociateCompanyPerson: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.CompanyPersonHandler.AssociateCompanyPerson: ValidationError while associating person to company",
				"error", err)
		} else {
			errorMessage = "Unknown internal error while associating person to company"
			status = http.StatusInternalServerError
			slog.Error(
				"v1.CompanyPersonHandler.AssociateCompanyPerson: Error while associating person to company",
				"error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	response := responses.NewCompanyPersonResponse(companyPersonModel)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.PersonHandler.AssociateCompanyPerson: Unable to write response", "error", err)
		http.Error(writer, "Person created but unable to create response", http.StatusInternalServerError)

		return
	}
}

func (handler *CompanyPersonHandler) GetCompanyPersonsByID(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	companyIDString := query.Get("company-id")
	personIDString := query.Get("person-id")

	if companyIDString == "" && personIDString == "" {
		errorMessage := "CompanyID and/or PersonID are required"
		slog.Info("v1.CompanyPersonHandler.GetCompanyPersonsByID: " + errorMessage)

		status := http.StatusBadRequest
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	var companyID, personID *uuid.UUID = nil, nil

	if companyIDString != "" {
		companyIDValue, err := uuid.Parse(companyIDString)
		if err != nil || companyIDValue == uuid.Nil {
			errorMessage := "Unable to parse CompanyID"
			slog.Info("v1.CompanyPersonHandler.GetCompanyPersonsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}

		companyID = &companyIDValue
	}

	if personIDString != "" {
		personIDValue, err := uuid.Parse(personIDString)
		if err != nil || personIDValue == uuid.Nil {
			errorMessage := "Unable to parse PersonID"
			slog.Info("v1.CompanyPersonHandler.GetCompanyPersonsByID: " + errorMessage)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(writer, errorMessage, status)
			return
		}
		personID = &personIDValue
	}

	companyPersons, err := handler.companyPersonService.GetByID(companyID, personID)
	if err != nil {
		errorMessage := "Internal service error while getting companyPersons by ID"
		slog.Error("v1.CompanyPersonHandler.GetCompanyPersonsByID: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	response := responses.NewCompanyPersonsResponse(companyPersons)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		slog.Error("v1.CompanyPersonHandler.GetCompanyPersonsByID: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.CompanyPersonHandler.GetCompanyPersonsByID: retrieved all companies successfully")

	return
}

func (handler *CompanyPersonHandler) GetAllCompanyPersons(writer http.ResponseWriter, request *http.Request) {
	companyPersons, err := handler.companyPersonService.GetAll()
	if err != nil {
		errorMessage := "Internal service error while getting all companyPersons"
		slog.Error("v1.CompanyHandler.GetAllCompanyPersons: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	companyPersonsResponse := responses.NewCompanyPersonsResponse(companyPersons)

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(companyPersonsResponse)
	if err != nil {
		slog.Error("v1.CompanyPersonHandler.GetAllCompanyPersons: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.CompanyPersonHandler.GetAllCompanyPersons: retrieved all CompanyPersons successfully")

	return
}

// DeleteCompanyPerson can return InternalServiceError, ValidationError
func (handler *CompanyPersonHandler) DeleteCompanyPerson(writer http.ResponseWriter, request *http.Request) {
	var deleteRequest requests.DeleteCompanyPersonRequest
	if err := json.NewDecoder(request.Body).Decode(&deleteRequest); err != nil {
		slog.Info("v1.CompanyPersonHandler.DeleteCompanyPerson: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	deleteModel, err := deleteRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.CompanyPersonHandler.DeleteCompanyPerson: Unable to convert DeleteCompanyPersonRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if deleteModel == nil {
		slog.Info("v1.CompanyPersonHandler.AssociateCompanyPerson: DeleteCompanyPerson is nil")
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = handler.companyPersonService.Delete(deleteModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var notFoundErr *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while deleting CompanyPerson"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyPersonHandler.DeleteCompanyPerson: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundErr) {
			errorMessage = err.Error()
			status = http.StatusNotFound
			slog.Info(
				"v1.CompanyPersonHandler.DeleteCompanyPerson: NotFoundErr while deleting CompanyPerson", "error",
				err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info(
				"v1.CompanyPersonHandler.DeleteCompanyPerson: ValidationError while deleting CompanyPerson", "error",
				err)
		} else {
			errorMessage = "Unknown internal error while creating company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyPersonHandler.DeleteCompanyPerson: Error while deleting CompanyPerson", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}
