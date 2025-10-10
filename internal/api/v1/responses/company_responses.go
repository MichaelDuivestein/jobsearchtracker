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
	ID           uuid.UUID               `json:"id" swaggertype:"string" format:"uuid" example:"123e4567-e89b-12d3-a456-426614174000" extensions:"x-order=0"`
	Name         string                  `json:"name" example:"CompanyName AB" extensions:"x-order=1"`
	CompanyType  requests.CompanyType    `json:"company_type" example:"employer" extensions:"x-order=2"`
	Notes        *string                 `json:"notes" example:"Notes go here" extensions:"x-order=3"`
	LastContact  *time.Time              `json:"last_contact" example:"2025-12-31T23:59Z"  extensions:"x-order=4"`
	CreatedDate  time.Time               `json:"created_date" example:"2025-12-31T23:59Z"  extensions:"x-order=5"`
	UpdatedDate  *time.Time              `json:"updated_date" example:"2025-12-31T23:59Z"  extensions:"x-order=6"`
	Applications *[]*ApplicationResponse `json:"applications" extensions:"x-order=7"`
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

	var applications []*ApplicationResponse = nil
	if companyModel.Applications != nil && len(*companyModel.Applications) >= 0 {
		applications, err = NewApplicationsResponse(*companyModel.Applications)
		if err != nil {
			return nil, err
		}
	}

	companyResponse := CompanyResponse{
		ID:           companyModel.ID,
		Name:         companyModel.Name,
		CompanyType:  companyType,
		Notes:        companyModel.Notes,
		LastContact:  companyModel.LastContact,
		CreatedDate:  companyModel.CreatedDate,
		UpdatedDate:  companyModel.UpdatedDate,
		Applications: &applications,
	}

	return &companyResponse, nil
}

// NewCompaniesResponse can return InternalServiceError
func NewCompaniesResponse(companyModels []*models.Company) ([]*CompanyResponse, error) {
	if len(companyModels) == 0 {
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
