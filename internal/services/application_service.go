package services

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/repositories"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type ApplicationService struct {
	applicationRepository *repositories.ApplicationRepository
}

func NewApplicationService(applicationRepository *repositories.ApplicationRepository) *ApplicationService {
	return &ApplicationService{applicationRepository: applicationRepository}
}

// CreateApplication can return  ConflictError, InternalServiceError, ValidationError
func (applicationService *ApplicationService) CreateApplication(
	application *models.CreateApplication) (*models.Application, error) {
	if application == nil {
		slog.Error("application_service.CreateApplication: application is nil")
		return nil, internalErrors.NewValidationError(nil, "CreateApplication is nil")
	}

	err := application.Validate()
	if err != nil {
		var applicationID string
		if application.ID != nil {
			applicationID = application.ID.String()
		} else {
			applicationID = "[not set]"
		}
		slog.Info(
			"company_service.CreateApplication: Application to create is invalid. ",
			"ID", applicationID,
			"error", err)
		return nil, err
	}

	if application.CreatedDate == nil {
		createdDate := time.Now()
		application.CreatedDate = &createdDate
	} else if application.CreatedDate.IsZero() {
		createdDate := time.Now()
		application.CreatedDate = &createdDate
		slog.Info(
			"application_service.createApplication: application.CreatedDate is zero. Setting to '" +
				application.CreatedDate.String() + "'")
	}

	// can return ConflictError, InternalServiceError, ValidationError
	insertedApplication, err := applicationService.applicationRepository.Create(application)
	if err != nil {
		return nil, err
	}

	slog.Info("application_service.createApplication: Inserted application.", "application.ID", insertedApplication.ID)
	return insertedApplication, nil
}

// GetApplicationById can return  ConflictError, InternalServiceError, NewValidationError
func (applicationService *ApplicationService) GetApplicationById(
	applicationId *uuid.UUID) (*models.Application, error) {

	if applicationId == nil {
		applicationIdString := "application ID"
		err := internalErrors.NewValidationError(&applicationIdString, "applicationId is required")
		slog.Info("applicationService.GetApplicationById: Failed to get application", "error", err)
		return nil, err
	}

	// can return InternalServiceError, NotFoundError, ValidationError
	application, err := applicationService.applicationRepository.GetById(applicationId)
	if err != nil {
		return nil, err
	}

	slog.Info("ApplicationService.GetApplicationById: Retrieved company.", "company.ID", application.ID.String())

	return application, nil
}

// GetApplicationsByJobTitle can return InternalServiceError, NotFoundError, ValidationError
func (applicationService *ApplicationService) GetApplicationsByJobTitle(
	applicationJobTitle *string) ([]*models.Application, error) {
	if applicationJobTitle == nil || *applicationJobTitle == "" {
		applicationJobTitleString := "applicationJobTitle"
		err := internalErrors.NewValidationError(&applicationJobTitleString, "applicationJobTitle is required")
		slog.Info("applicationService.GetApplicationByJobTitle: Failed to get applications", "error", err)
		return nil, err
	}

	applications, err := applicationService.applicationRepository.GetAllByJobTitle(applicationJobTitle)
	if err != nil {
		return nil, err
	}

	if applications == nil {
		slog.Info("ApplicationService.GetApplicationsByJobTitle: Retrieved zero applications")
	} else {
		slog.Info(
			"ApplicationService.GetApplicationsByJobTitle: Retrieved " +
				string(rune(len(applications))) +
				" applications")
	}

	return applications, nil
}

// GetAllApplications can return InternalServiceError
func (applicationService *ApplicationService) GetAllApplications() ([]*models.Application, error) {
	// can return InternalServiceError
	applications, err := applicationService.applicationRepository.GetAll()
	if err != nil {
		return nil, err
	}

	if applications == nil {
		slog.Info("ApplicationService.GetAllApplications: Retrieved zero applications")
	} else {
		slog.Info(
			"ApplicationService.GetAllApplications: Retrieved " + string(rune(len(applications))) + " applications")
	}

	return applications, nil
}

// UpdateApplication can return InternalServiceError, ValidationError
func (applicationService *ApplicationService) UpdateApplication(application *models.UpdateApplication) error {
	if application == nil {
		slog.Error("ApplicationService.UpdateApplication: UpdateApplication is nil")
		return internalErrors.NewValidationError(nil, "UpdateApplication model is nil")
	}

	// can return ValidationError
	err := application.Validate()
	if err != nil {
		slog.Info("ApplicationService.UpdateApplication: UpdateApplication model is invalid. ", "error", err)
		return err
	}

	// can return InternalServiceError, ValidationError
	err = applicationService.applicationRepository.Update(application)
	if err != nil {
		slog.Error("ApplicationService.UpdateApplication: Error updating application", "error", err)
	}

	return err
}

// DeleteApplication can return InternalServiceError, NotFoundError, ValidationError
func (applicationService *ApplicationService) DeleteApplication(applicationId *uuid.UUID) error {
	if applicationId == nil {
		applicationIdString := "application ID"
		err := internalErrors.NewValidationError(&applicationIdString, "applicationId is required")
		slog.Info("ApplicationService.DeleteApplication: Error deleting application", "error", err)
		return err
	}

	// can return InternalServiceError, ValidationError
	err := applicationService.applicationRepository.Delete(applicationId)
	if err != nil {
		slog.Error("ApplicationService.DeleteApplication: Error deleting application", "error", err)
	}

	return err
}
