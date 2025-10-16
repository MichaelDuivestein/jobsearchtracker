package responses

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"jobsearchtracker/internal/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// -------- NewCompanyDTO tests: --------

func TestNewCompanyDTO_ShouldWork(t *testing.T) {
	model := models.Company{
		ID:          uuid.New(),
		Name:        "Randomized Company",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		CreatedDate: time.Now().AddDate(0, 0, -4),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}

	dto, err := NewCompanyDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, dto)

	assert.Equal(t, model.ID, dto.ID)
	assert.Equal(t, model.Name, dto.Name)
	assert.Equal(t, model.CompanyType.String(), dto.CompanyType.String())
	assert.Equal(t, model.Notes, dto.Notes)
	assert.Equal(t, model.LastContact, dto.LastContact)
	assert.Equal(t, model.CreatedDate, dto.CreatedDate)
	assert.Equal(t, model.UpdatedDate, dto.UpdatedDate)
}

func TestNewCompanyDTO_ShouldWorkWithOnlyRequiredFields(t *testing.T) {
	model := models.Company{
		ID:          uuid.New(),
		Name:        "Yet another company name",
		CompanyType: models.CompanyTypeConsultancy,
		CreatedDate: time.Now().AddDate(0, 0, 1),
	}

	dto, err := NewCompanyDTO(&model)
	assert.NoError(t, err)
	assert.NotNil(t, dto)

	assert.Equal(t, model.ID, dto.ID)
	assert.Equal(t, model.Name, dto.Name)
	assert.Equal(t, model.CompanyType.String(), dto.CompanyType.String())
	assert.Nil(t, dto.Notes)
	assert.Nil(t, dto.LastContact)
	assert.Equal(t, model.CreatedDate, dto.CreatedDate)
	assert.Nil(t, dto.UpdatedDate)
}

func TestNewCompanyDTO_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilDTO, err := NewCompanyDTO(nil)
	assert.Nil(t, nilDTO)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, "internal service error: Error building DTO: Company is nil", internalServiceErr.Error())
}

func TestNewCompanyDTO_ShouldReturnInternalServiceErrorIfCompanyTypeIsInvalid(t *testing.T) {
	emptyCompanyType := models.Company{
		ID:          uuid.New(),
		Name:        "Randomized Company",
		CompanyType: models.CompanyType(""),
		CreatedDate: time.Now().AddDate(0, 0, 3),
	}

	nilDTO, err := NewCompanyDTO(&emptyCompanyType)
	assert.Nil(t, nilDTO)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal CompanyType to external CompanyType: ''",
		internalServiceErr.Error())

	badCompanyType := models.Company{
		ID:          uuid.New(),
		Name:        "Randomized Company",
		CompanyType: models.CompanyType("BadData"),
		CreatedDate: time.Now().AddDate(0, 0, 3),
	}

	badDataResponse, err := NewCompanyResponse(&badCompanyType)
	assert.Nil(t, badDataResponse)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal CompanyType to external CompanyType: 'BadData'",
		internalServiceErr.Error())
}

// -------- NewCompaniesDTO tests: --------

func TestNewCompaniesDTO_ShouldWork(t *testing.T) {
	companyModels := []*models.Company{
		{
			ID:          uuid.New(),
			Name:        "CompanyOne",
			CompanyType: models.CompanyTypeConsultancy,
			CreatedDate: time.Now().AddDate(0, 0, -1),
		},
		{
			ID:          uuid.New(),
			Name:        "CompanyTwo",
			CompanyType: models.CompanyTypeEmployer,
			CreatedDate: time.Now().AddDate(0, 0, -2),
		},
	}

	companyDTOs, err := NewCompanyDTOs(companyModels)
	assert.NoError(t, err)
	assert.NotNil(t, companyDTOs)
	assert.Len(t, companyDTOs, 2)
}

func TestNewCompaniesDTO_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	emptyDTOs, err := NewCompanyDTOs(nil)
	assert.NoError(t, err)
	assert.NotNil(t, emptyDTOs)
	assert.Len(t, emptyDTOs, 0)
}

func TestNewCompaniesDTO_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var companyModels []*models.Company
	emptyDTOs, err := NewCompanyDTOs(companyModels)
	assert.NoError(t, err)
	assert.NotNil(t, emptyDTOs)
	assert.Len(t, emptyDTOs, 0)
}

func TestNewCompaniesDTO_ShouldReturnInternalServiceErrorIfOneCompanyTypeIsInvalid(t *testing.T) {
	companyModels := []*models.Company{
		{
			ID:          uuid.New(),
			Name:        "CompanyOne",
			CompanyType: models.CompanyTypeEmployer,
			CreatedDate: time.Now().AddDate(0, 0, 0),
		},
		{
			ID:          uuid.New(),
			Name:        "CompanyTwo",
			CompanyType: "",
			CreatedDate: time.Now().AddDate(0, 0, 0),
		},
	}

	nilDTOs, err := NewCompanyDTOs(companyModels)
	assert.Nil(t, nilDTOs)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t,
		"internal service error: Error converting internal CompanyType to external CompanyType: ''",
		err.Error())
}

// -------- NewCompanyResponse tests: --------

func TestNewCompanyResponse_ShouldWork(t *testing.T) {
	model := models.Company{
		ID:          uuid.New(),
		Name:        "Randomized Company",
		CompanyType: models.CompanyTypeEmployer,
		Notes:       testutil.ToPtr("some notes"),
		LastContact: testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		CreatedDate: time.Now().AddDate(0, 0, -4),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}

	response, err := NewCompanyResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.CompanyType.String(), response.CompanyType.String())
	assert.Equal(t, model.Notes, response.Notes)
	assert.Equal(t, model.LastContact, response.LastContact)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Equal(t, model.UpdatedDate, response.UpdatedDate)
}

func TestNewCompanyResponse_ShouldReturnInternalServiceErrorIfModelIsNil(t *testing.T) {
	nilModel, err := NewCompanyResponse(nil)
	assert.Nil(t, nilModel)
	assert.NotNil(t, err)

	var internalServiceErr *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceErr))

	assert.Equal(t, "internal service error: Error building response: Company is nil", internalServiceErr.Error())
}

func TestNewCompanyResponse_ShouldHandleApplications(t *testing.T) {
	companyId := uuid.New()

	application1 := models.Application{
		ID:        uuid.New(),
		CompanyID: &companyId,
	}

	var application2RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeOffice
	application2 := models.Application{
		ID:                   uuid.New(),
		RecruiterID:          &companyId,
		JobTitle:             testutil.ToPtr("Application2JobTitle"),
		JobAdURL:             testutil.ToPtr("Application2JobAdURL"),
		Country:              testutil.ToPtr("Application2Country"),
		Area:                 testutil.ToPtr("Application2Area"),
		RemoteStatusType:     &application2RemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(3),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(1),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
	}
	applications := []*models.Application{
		&application1,
		&application2,
	}

	model := models.Company{
		ID:           companyId,
		Name:         "Randomized Company",
		CompanyType:  models.CompanyTypeEmployer,
		Notes:        testutil.ToPtr("some notes"),
		LastContact:  testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		Applications: &applications,
		CreatedDate:  time.Now().AddDate(0, 0, -4),
		UpdatedDate:  testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
	}

	response, err := NewCompanyResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.CompanyType.String(), response.CompanyType.String())
	assert.Equal(t, model.Notes, response.Notes)
	assert.Equal(t, model.LastContact, response.LastContact)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Equal(t, model.UpdatedDate, response.UpdatedDate)

	assert.Equal(t, application1.ID, (*response.Applications)[0].ID)
	assert.Equal(t, companyId, *(*response.Applications)[0].CompanyID)
	assert.Nil(t, (*response.Applications)[0].RecruiterID)

	returnedApplication2 := (*response.Applications)[1]
	assert.Equal(t, application2.ID, returnedApplication2.ID)
	assert.Nil(t, returnedApplication2.CompanyID)
	assert.Equal(t, companyId, *returnedApplication2.RecruiterID)
	assert.Equal(t, "Application2JobTitle", *returnedApplication2.JobTitle)
	assert.Equal(t, "Application2JobAdURL", *returnedApplication2.JobAdURL)
	assert.Equal(t, "Application2Country", *returnedApplication2.Country)
	assert.Equal(t, "Application2Area", *returnedApplication2.Area)
	assert.Equal(t, models.RemoteStatusTypeOffice, returnedApplication2.RemoteStatusType.String())
	assert.Equal(t, 3, *returnedApplication2.WeekdaysInOffice)
	assert.Equal(t, 2, *returnedApplication2.EstimatedCycleTime)
	assert.Equal(t, 1, *returnedApplication2.EstimatedCommuteTime)
	testutil.AssertEqualFormattedDateTimes(t, application2.ApplicationDate, returnedApplication2.ApplicationDate)
	testutil.AssertEqualFormattedDateTimes(t, application2.CreatedDate, returnedApplication2.CreatedDate)
	testutil.AssertEqualFormattedDateTimes(t, application2.UpdatedDate, returnedApplication2.UpdatedDate)
}

func TestNewCompanyResponse_ShouldHandlePersons(t *testing.T) {
	companyID := uuid.New()

	person1 := models.Person{
		ID: uuid.New(),
	}

	var person2Type models.PersonType = models.PersonTypeHR
	person2 := models.Person{
		ID:          uuid.New(),
		Name:        testutil.ToPtr("Person2Name"),
		PersonType:  &person2Type,
		Email:       testutil.ToPtr("Person2Email"),
		Phone:       testutil.ToPtr("Person2Phone"),
		Notes:       testutil.ToPtr("Person2Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 7)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 12)),
	}

	persons := []*models.Person{
		&person1,
		&person2,
	}

	model := models.Company{
		ID:          companyID,
		Name:        "CompanyName",
		CompanyType: models.CompanyTypeEmployer,
		CreatedDate: time.Now().AddDate(0, 0, 4),
		Persons:     &persons,
	}

	response, err := NewCompanyResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.CompanyType.String(), response.CompanyType.String())
	assert.Nil(t, response.Notes)
	assert.Nil(t, response.LastContact)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Nil(t, response.UpdatedDate)
	assert.Len(t, *response.Applications, 0)
	assert.Len(t, *response.Persons, 2)

	assert.Equal(t, person1.ID, (*response.Persons)[0].ID)

	assert.Equal(t, person2.ID, (*response.Persons)[1].ID)
	assert.Equal(t, "Person2Name", *(*response.Persons)[1].Name)
	assert.Equal(t, person2Type.String(), (*response.Persons)[1].PersonType.String())
	assert.Equal(t, "Person2Email", *(*response.Persons)[1].Email)
	assert.Equal(t, "Person2Phone", *(*response.Persons)[1].Phone)
	assert.Equal(t, "Person2Notes", *(*response.Persons)[1].Notes)
	testutil.AssertEqualFormattedDateTimes(t, person2.CreatedDate, (*response.Persons)[1].CreatedDate)
}

func TestNewCompanyResponse_ShouldHandleApplicationsAndPersons(t *testing.T) {
	companyId := uuid.New()

	application1 := models.Application{
		ID:        uuid.New(),
		CompanyID: &companyId,
	}

	var application2RemoteStatusType models.RemoteStatusType = models.RemoteStatusTypeOffice
	application2 := models.Application{
		ID:                   uuid.New(),
		RecruiterID:          &companyId,
		JobTitle:             testutil.ToPtr("Application2JobTitle"),
		JobAdURL:             testutil.ToPtr("Application2JobAdURL"),
		Country:              testutil.ToPtr("Application2Country"),
		Area:                 testutil.ToPtr("Application2Area"),
		RemoteStatusType:     &application2RemoteStatusType,
		WeekdaysInOffice:     testutil.ToPtr(3),
		EstimatedCycleTime:   testutil.ToPtr(2),
		EstimatedCommuteTime: testutil.ToPtr(1),
		ApplicationDate:      testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		CreatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		UpdatedDate:          testutil.ToPtr(time.Now().AddDate(0, 0, -1)),
	}
	applications := []*models.Application{
		&application1,
		&application2,
	}

	person1ID := uuid.New()
	person1 := models.Person{
		ID: person1ID,
	}

	person2ID := uuid.New()
	var person2Type models.PersonType = models.PersonTypeHR
	person2 := models.Person{
		ID:          person2ID,
		Name:        testutil.ToPtr("Person2Name"),
		PersonType:  &person2Type,
		Email:       testutil.ToPtr("Person2Email"),
		Phone:       testutil.ToPtr("Person2Phone"),
		Notes:       testutil.ToPtr("Person2Notes"),
		CreatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 7)),
		UpdatedDate: testutil.ToPtr(time.Now().AddDate(0, 0, 12)),
	}

	persons := []*models.Person{
		&person1,
		&person2,
	}

	model := models.Company{
		ID:           companyId,
		Name:         "Randomized Company",
		CompanyType:  models.CompanyTypeEmployer,
		Notes:        testutil.ToPtr("some notes"),
		LastContact:  testutil.ToPtr(time.Now().AddDate(0, 0, -3)),
		CreatedDate:  time.Now().AddDate(0, 0, -4),
		UpdatedDate:  testutil.ToPtr(time.Now().AddDate(0, 0, -2)),
		Applications: &applications,
		Persons:      &persons,
	}

	response, err := NewCompanyResponse(&model)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, model.ID, response.ID)
	assert.Equal(t, model.Name, response.Name)
	assert.Equal(t, model.CompanyType.String(), response.CompanyType.String())
	assert.Equal(t, model.Notes, response.Notes)
	assert.Equal(t, model.LastContact, response.LastContact)
	assert.Equal(t, model.CreatedDate, response.CreatedDate)
	assert.Equal(t, model.UpdatedDate, response.UpdatedDate)
	assert.Len(t, *response.Applications, 2)
	assert.Len(t, *response.Persons, 2)

	assert.Equal(t, application1.ID, (*response.Applications)[0].ID)
	assert.Equal(t, companyId, *(*response.Applications)[0].CompanyID)
	assert.Nil(t, (*response.Applications)[0].RecruiterID)

	application := (*response.Applications)[1]
	assert.Equal(t, application2.ID, application.ID)
	assert.Nil(t, application.CompanyID)
	assert.Equal(t, companyId, *application.RecruiterID)

	assert.Equal(t, person1ID, (*response.Persons)[0].ID)
	assert.Equal(t, person2ID, (*response.Persons)[1].ID)
}

// -------- NewCompaniesResponse tests: --------

func TestNewCompaniesResponse_ShouldWork(t *testing.T) {
	companyModels := []*models.Company{
		{
			ID:          uuid.New(),
			Name:        "CompanyOne",
			CompanyType: models.CompanyTypeConsultancy,
			CreatedDate: time.Now().AddDate(0, 0, -1),
		},
		{
			ID:          uuid.New(),
			Name:        "CompanyTwo",
			CompanyType: models.CompanyTypeEmployer,
			CreatedDate: time.Now().AddDate(0, 0, -2),
		},
	}

	companies, err := NewCompaniesResponse(companyModels)
	assert.NoError(t, err)
	assert.NotNil(t, companies)
	assert.Len(t, companies, 2)
}

func TestNewCompaniesResponse_ShouldReturnEmptySliceIfModelIsNil(t *testing.T) {
	response, err := NewCompaniesResponse(nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}

func TestNewCompaniesResponse_ShouldReturnEmptySliceIfModelIsEmpty(t *testing.T) {
	var companyModels []*models.Company
	response, err := NewCompaniesResponse(companyModels)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 0)
}
