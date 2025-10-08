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

type CompanyHandler struct {
	companyService *services.CompanyService
}

func NewCompanyHandler(companyService *services.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService}
}

// CreateCompany creates a company and returns it
//
// @Summary create a company
// @Description create a `company` and return it
// @Tags company
// @Accept json
// @Produce json
// @Param company body requests.CreateCompanyRequest true "Create Company request"
// @Success 201 {object} responses.CompanyResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company/new [post]
func (companyHandler *CompanyHandler) CreateCompany(writer http.ResponseWriter, request *http.Request) {
	var createCompanyRequest requests.CreateCompanyRequest
	if err := json.NewDecoder(request.Body).Decode(&createCompanyRequest); err != nil {
		slog.Info("v1.CompanyHandler.CreateCompany: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	createCompanyModel, err := createCompanyRequest.ToModel()
	if err != nil {
		slog.Info(
			"v1.CompanyHandler.CreateCompany: Unable to convert CreateCompanyRequest to model",
			"error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	if createCompanyModel == nil {
		slog.Error("v1.CompanyHandler.CreateCompany: createCompanyModel is nil", "error", err)
		http.Error(writer,
			"Unable to convert request to internal model: Internal model is nil",
			http.StatusInternalServerError)
		return
	}

	// can return ConflictError, InternalServiceError, ValidationError
	createdCompany, err := companyHandler.companyService.CreateCompany(createCompanyModel)
	if err != nil {
		var conflictErr *internalErrors.ConflictError
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &conflictErr) {
			errorMessage = "Conflict error on insert: ID already exists"
			status = http.StatusConflict
			slog.Info("v1.CompanyHandler.CreateCompany: ConflictError creating company", "error", err)
		} else if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while creating company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.CreateCompany: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.CompanyHandler.CreateCompany: ValidationError while creating company", "error", err)
		} else {
			errorMessage = "Unknown internal error while creating company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.CreateCompany: Error while creating company", "error", err)
		}
		http.Error(writer, errorMessage, status)
		return
	}

	// can return InternalServiceError
	companyResponse, err := responses.NewCompanyResponse(createdCompany)
	if err != nil {
		slog.Error("v1.CompanyHandler.CreateCompany: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(companyResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.CompanyHandler.CreateCompany: Unable to write response", "error", err)
		http.Error(writer, "Company created but unable to create response", http.StatusInternalServerError)

		return
	}
}

// GetCompanyById retrieves a company matching input UUID
//
// @Summary Get a company by ID
// @Description Get a `company` by ID
// @Tags company
// @Produce json
// @Param id path string true "Company ID" format(uuid)
// @Success 200 {object} responses.CompanyResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company/get/id/{id} [get]
func (companyHandler *CompanyHandler) GetCompanyById(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	companyIDStr := vars["id"]

	if companyIDStr == "" {
		slog.Info("v1.CompanyHandler.GetCompanyById: company ID is empty")
		http.Error(writer, "company ID is empty", http.StatusBadRequest)
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		slog.Info("v1.CompanyHandler.GetCompanyById: Company ID is not a valid UUID")
		http.Error(writer, "company ID is not a valid UUID", http.StatusBadRequest)
		return
	}

	var internalServiceError *internalErrors.InternalServiceError
	var notFoundError *internalErrors.NotFoundError
	var validationErr *internalErrors.ValidationError

	// can return InternalServiceError, NotFoundError, ValidationError
	company, err := companyHandler.companyService.GetCompanyById(&companyID)
	if err != nil {
		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while retrieving company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.GetCompanyById: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "Company not found"
			status = http.StatusNotFound
			slog.Info("v1.CompanyHandler.GetCompanyById: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.CompanyHandler.GetCompanyById: Validation error", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	companyResponse, err := responses.NewCompanyResponse(company)
	if err != nil {
		slog.Error("v1.CompanyHandler.GetCompanyById: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(companyResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.CompanyHandler.GetCompanyById: Unable to write response", "error", err)
		http.Error(writer, "Company found but unable to build response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.CompanyHandler.GetCompanyById: retrieved company successfully", "company.ID", company.ID.String())

	return
}

// GetCompaniesByName retrieves `company`s which fully, or partially, match the input name
//
// @Summary Get companies by name
// @Description Get `company`s which fully, or partially, match the input name
// @Tags company
// @Produce json
// @Param name path string true "Company Name"
// @Success 200 {array} responses.CompanyResponse
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company/get/name/{name} [get]
func (companyHandler *CompanyHandler) GetCompaniesByName(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	companyName := vars["name"]

	if companyName == "" {
		slog.Info("v1.CompanyHandler.GetCompanyByName: company Name is empty")
		http.Error(writer, "company Name is empty", http.StatusBadRequest)
		return
	}

	companies, err := companyHandler.companyService.GetCompaniesByName(&companyName)
	if err != nil {
		var internalServiceError *internalErrors.InternalServiceError
		var notFoundError *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceError) {
			errorMessage = "Internal service error while retrieving companies"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.GetCompaniesByName: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "No people [partially] matching this name found"
			status = http.StatusNotFound
			slog.Info("v1.CompanyHandler.GetCompaniesByName: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.CompanyHandler.GetCompaniesByName: Validation error", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	// can return InternalServiceError
	companiesResponse, err := responses.NewCompaniesResponse(companies)
	if err != nil {
		slog.Error("v1.CompanyHandler.GetCompaniesByName: Unable to convert internal model to response", "error", err)
		http.Error(writer, "Error: Unable to convert internal model to response", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(companiesResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("v1.CompanyHandler.GetCompaniesByName: Unable to write response", "error", err)
		http.Error(writer, "Company found but unable to build response", http.StatusInternalServerError)

		return
	}

	slog.Info("v1.CompanyHandler.GetCompaniesByName: retrieved companies successfully", "name", companyName)

	return
}

// GetAllCompanies retrieves all companies.
//
// @Summary Get all companies
// @Description Get all `company`s
// @Description - include_applications=all: Returns `application`s with all fields
// @Description - include_applications=ids: Returns `application`s with only `id`, `application_id`, and `recruiter_id`
// @Description - include_applications=none: No `application` data included (default)
// @Tags company
// @Produce json
// @Param include_applications query string false "string enums" Enums(all, ids, none)
// @Success 200 {array} responses.CompanyResponse
// @Failure 400
// @Failure 500
// @Router /v1/company/get/all [get]
func (companyHandler *CompanyHandler) GetAllCompanies(writer http.ResponseWriter, request *http.Request) {

	query := request.URL.Query()
	includeApplicationsString := query.Get("include_applications")

	var includeApplicationsType requests.IncludeExtraDataType
	if includeApplicationsString == "" {
		includeApplicationsType = requests.IncludeExtraDataTypeNone
	} else {
		var err error

		// can return ValidationError
		includeApplicationsType, err = requests.NewIncludeExtraDataType(includeApplicationsString)

		if err != nil {
			slog.Error("v1.CompanyHandler.CreateCompany: Could not parse include_applications param", "error", err)

			status := http.StatusBadRequest
			writer.WriteHeader(status)
			http.Error(
				writer,
				"Invalid value for include_applications. Accepted params are 'all', 'ids', and 'none'",
				status)
			return
		}
	}

	includeApplicationsTypeModel, err := includeApplicationsType.ToModel()
	if err != nil {
		slog.Error(
			"v1.CompanyHandler.CreateCompany: For include_applications, unable to convert request to model",
			"error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "For include_applications, unable to convert request to model", status)
		return
	}

	// can return InternalServiceError
	companies, err := companyHandler.companyService.GetAllCompanies(includeApplicationsTypeModel)
	if err != nil {
		errorMessage := "Internal service error while getting all companies"
		slog.Error("v1.CompanyHandler.GetAllCompanies: "+errorMessage, "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, errorMessage, status)
		return
	}

	// can return InternalServiceError
	companiesResponse, err := responses.NewCompaniesResponse(companies)
	if err != nil {
		slog.Error("v1.CompanyHandler.GetAllCompanies: Unable to convert internal model to response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(http.StatusInternalServerError)
		http.Error(writer, "Error: Unable to convert internal model to response", status)
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(companiesResponse)
	if err != nil {
		slog.Error("v1.CompanyHandler.GetAllCompanies: Unable to write response", "error", err)

		status := http.StatusInternalServerError
		writer.WriteHeader(status)
		http.Error(writer, "Companies retrieved but unable to create response", status)

		return
	}

	slog.Info("v1.CompanyHandler.GetAllCompanies: retrieved all companies successfully")

	return
}

// UpdateCompany updates a company
//
// @Summary update a company
// @Description update a `company`
// @Tags company
// @Accept json
// @Produce json
// @Param company body requests.UpdateCompanyRequest true "Update Company Request"
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company/update [post]
func (companyHandler *CompanyHandler) UpdateCompany(writer http.ResponseWriter, request *http.Request) {
	var updateCompanyRequest requests.UpdateCompanyRequest
	if err := json.NewDecoder(request.Body).Decode(&updateCompanyRequest); err != nil {
		slog.Info("v1.CompanyHandler.UpdateCompany: invalid request body", "error", err)
		http.Error(writer, "invalid request body: Unable to parse JSON", http.StatusBadRequest)
		return
	}

	// can return ValidationError
	updateCompanyModel, err := updateCompanyRequest.ToModel()
	if err != nil {
		slog.Info("v1.CompanyHandler.UpdateCompany: Unable to convert UpdateCompanyRequest to model", "error", err)
		http.Error(writer, "Unable to convert request to internal model: "+err.Error(), http.StatusBadRequest)

		return
	}
	if updateCompanyModel == nil {
		slog.Error(
			"v1.CompanyHandler.UpdateCompany: updateCompanyModel is nil after attempting to convert request to internal model")
		http.Error(writer, "Unable to convert request to model", http.StatusBadRequest)
		return
	}

	// can return InternalServiceError, ValidationError
	err = companyHandler.companyService.UpdateCompany(updateCompanyModel)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while updating company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.UpdateCompany: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.CompanyHandler.UpdateCompany: ValidationError while updating company", "error", err)
		} else {
			errorMessage = "Unknown internal error while updating company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.UpdateCompany: Error while updating company", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}

// DeleteCompany deletes a `company` matching input UUID
//
// @Summary Delete a company by ID
// @Description Delete a `company` by ID
// @Tags company
// @Param id path string true "Company ID" format(uuid)
// @Success 200
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /v1/company/delete/{id} [delete]
func (companyHandler *CompanyHandler) DeleteCompany(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	companyIDStr := vars["id"]

	if companyIDStr == "" {
		errorMessage := "company ID is empty"
		slog.Info(errorMessage)
		http.Error(writer, errorMessage, http.StatusBadRequest)
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		errorMessage := "company ID is not a valid UUID"
		slog.Info(errorMessage)
		http.Error(writer, errorMessage, http.StatusBadRequest)
		return
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	err = companyHandler.companyService.DeleteCompany(&companyID)
	if err != nil {
		var internalServiceErr *internalErrors.InternalServiceError
		var notFoundError *internalErrors.NotFoundError
		var validationErr *internalErrors.ValidationError

		var errorMessage string
		var status int

		if errors.As(err, &internalServiceErr) {
			errorMessage = "Internal service error while deleting company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.DeleteCompany: "+errorMessage, "error", err)
		} else if errors.As(err, &notFoundError) {
			errorMessage = "Company not found"
			status = http.StatusNotFound
			slog.Info("v1.CompanyHandler.DeleteCompany: "+errorMessage, "error", err)
		} else if errors.As(err, &validationErr) {
			errorMessage = err.Error()
			status = http.StatusBadRequest
			slog.Info("v1.CompanyHandler.DeleteCompany: ValidationError while deleting company", "error", err)
		} else {
			errorMessage = "Unknown internal error while creating company"
			status = http.StatusInternalServerError
			slog.Error("v1.CompanyHandler.DeleteCompany: Error while deleting company", "error", err)
		}
		http.Error(writer, errorMessage, status)

		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}
