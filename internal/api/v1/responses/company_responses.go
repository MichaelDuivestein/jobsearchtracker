package responses

import (
	"jobsearchtracker/internal/api/v1/requests"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type CompanyResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	CompanyType requests.CompanyType `json:"company_type"`
	Notes       *string              `json:"notes"`
	LastContact *time.Time           `json:"last_contact"`
	CreatedDate time.Time            `json:"created_date"`
	UpdatedDate *time.Time           `json:"updated_date"`
}

// NewCompanyResponse can return InternalServiceError
func NewCompanyResponse(companyModel *models.Company) (*CompanyResponse, error) {
	if companyModel == nil {
		slog.Error("responses.NewCompanyResponse: Company is nil")
		return nil, internalErrors.NewInternalServiceError("Error building response: Company is nil")
	}

	// can return InternalServerError
	companyType, err := requests.NewCompanyType(&companyModel.CompanyType)
	if err != nil {
		return nil, err
	}

	companyResponse := CompanyResponse{
		ID:          companyModel.ID,
		Name:        companyModel.Name,
		CompanyType: companyType,
		Notes:       companyModel.Notes,
		LastContact: companyModel.LastContact,
		CreatedDate: companyModel.CreatedDate,
		UpdatedDate: companyModel.UpdatedDate,
	}

	return &companyResponse, nil
}

// NewCompaniesResponse can return InternalServiceError
func NewCompaniesResponse(companyModels []*models.Company) ([]*CompanyResponse, error) {
	if companyModels == nil || len(companyModels) == 0 {
		return []*CompanyResponse{}, nil
	}

	var companyResponses = make([]*CompanyResponse, len(companyModels))
	for index := range companyModels {
		companyResponse, err := NewCompanyResponse(companyModels[index])
		if err != nil {
			return nil, err
		}
		companyResponses[index] = companyResponse

	}
	return companyResponses, nil
}
