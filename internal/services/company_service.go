package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type CompanyService struct {
	companyRepository *repositories.CompanyRepository
}

func NewCompanyService(companyRepository *repositories.CompanyRepository) *CompanyService {
	return &CompanyService{companyRepository: companyRepository}
}

// CreateCompany can return ConflictError, InternalServiceError, ValidationError
func (companyService *CompanyService) CreateCompany(company *models.CreateCompany) (*models.Company, error) {
	if company == nil {
		slog.Error("company_service.CreateCompany: company is nil")
		return nil, internalErrors.NewValidationError(nil, "CreateCompany is nil")
	}

	err := company.Validate()
	if err != nil {
		var companyId string
		if company.ID != nil {
			companyId = company.ID.String()
		} else {
			companyId = "[not set]"
		}

		slog.Info("company_service.CreateCompany: Company to create is invalid. ", "ID", companyId, "error", err)
		return nil, err
	}

	if company.CreatedDate == nil {
		createdDate := time.Now()
		company.CreatedDate = &createdDate
	} else if company.CreatedDate.IsZero() {
		createdDate := time.Now()
		company.CreatedDate = &createdDate
		slog.Info(
			"company_service.createCompany: company.CreatedDate is zero. Setting to '" +
				company.CreatedDate.String() + "'")
	}

	if company.LastContact != nil && company.LastContact.IsZero() {
		company.LastContact = company.CreatedDate
		slog.Info(
			"company_service.createCompany: company.LastContact is zero. Setting to CreatedDate: '" +
				company.LastContact.String() + "'")
	}

	// can return ConflictError, InternalServiceError
	insertedCompany, err := companyService.companyRepository.Create(company)
	if err != nil {
		return nil, err
	}

	slog.Info("CompanyService.CreateCompany: inserted company.", "company.ID", insertedCompany.ID.String())
	return insertedCompany, nil
}

// GetCompanyById can return InternalServiceError, NotFoundError, ValidationError
func (companyService *CompanyService) GetCompanyById(
	companyId *uuid.UUID) (*models.Company, error) {

	if companyId == nil {
		companyIdString := "company ID"
		err := internalErrors.NewValidationError(&companyIdString, "companyId is required")
		slog.Info("CompanyService.GetCompanyById: Failed to get company", "error", err)
		return nil, err
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	company, err := companyService.companyRepository.GetById(companyId)
	if err != nil {
		return nil, err
	}

	slog.Info("CompanyService.GetCompanyById: Retrieved company.", "company.ID", company.ID.String())
	return company, nil
}

// GetCompaniesByName can return InternalServiceError, NotFoundError, ValidationError
func (companyService *CompanyService) GetCompaniesByName(companyName *string) ([]*models.Company, error) {
	if companyName == nil {
		err := internalErrors.NewValidationError(nil, "companyName is required")
		slog.Info("companyService.GetCompanyByName: Failed to get company", "error", err)
		return nil, err
	}
	if *companyName == "" {
		err := internalErrors.NewValidationError(nil, "companyName is required")
		slog.Info("companyService.GetCompanyByName: Failed to get company", "error", err)
		return nil, err
	}

	companies, err := companyService.companyRepository.GetAllByName(companyName)
	if err != nil {
		return nil, err
	}

	if companies == nil {
		slog.Info("CompanyService.GetAllCompanies: Retrieved zero companies")
	} else {
		slog.Info("CompanyService.GetAllCompanies: Retrieved " + string(rune(len(companies))) + " companies")
	}

	return companies, nil
}

// GetAllCompanies can return InternalServiceError
func (companyService *CompanyService) GetAllCompanies(
	includeApplications models.IncludeExtraDataType,
	includePersons models.IncludeExtraDataType,
	includeEvents models.IncludeExtraDataType) ([]*models.Company, error) {

	// can return InternalServiceError
	companies, err := companyService.companyRepository.GetAll(includeApplications, includePersons, includeEvents)
	if err != nil {
		return nil, err
	}
	if len(companies) == 0 {
		slog.Info("CompanyService.GetAllCompanies: Retrieved zero companies")
		return nil, nil
	}

	slog.Info("CompanyService.GetAllCompanies: Retrieved " + string(rune(len(companies))) + " companies")
	return companies, nil
}

// UpdateCompany can return InternalServiceError, ValidationError
func (companyService *CompanyService) UpdateCompany(company *models.UpdateCompany) error {
	if company == nil {
		slog.Error("CompanyService.UpdateCompany: company is nil")
		return internalErrors.NewValidationError(nil, "UpdateCompany model is nil")
	}

	// can return ValidationError
	err := company.Validate()
	if err != nil {
		slog.Info("CompanyService.UpdateCompany: UpdateCompany model is invalid. ", "error", err)
		return err
	}

	// can return InternalServiceError, ValidationError
	err = companyService.companyRepository.Update(company)
	if err != nil {
		slog.Error("CompanyService.Update: Error updating company", "error", err)
	}

	return err
}

// DeleteCompany can return InternalServiceError, NotFoundError, ValidationError
func (companyService *CompanyService) DeleteCompany(companyId *uuid.UUID) error {
	if companyId == nil {
		companyIdString := "company ID"
		err := internalErrors.NewValidationError(&companyIdString, "companyId is required")
		slog.Info("CompanyService.DeleteCompany: Error: companyID is nil", "error", err)
		return err
	}

	// can return InternalServiceError, ValidationError
	err := companyService.companyRepository.Delete(companyId)
	if err != nil {
		slog.Error("CompanyService.DeleteCompany: Error deleting company", "error", err)
	}

	return err
}
