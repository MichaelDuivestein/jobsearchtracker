package requests

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

// -------- CreateCompanyRequest tests: --------

func TestCreateCompanyRequestValidate_ShouldValidateRequest(t *testing.T) {
	id := uuid.New()
	notes := "No notes here!"
	lastContact := time.Now().AddDate(0, 0, -3)

	request := CreateCompanyRequest{
		ID:          &id,
		Name:        "A random company",
		CompanyType: CompanyTypeEmployer,
		Notes:       &notes,
		LastContact: &lastContact,
	}

	err := request.Validate()
	assert.NoError(t, err)
}

func TestCreateCompanyRequestValidate_ShouldReturnValidationError(t *testing.T) {
	tests := []struct {
		testName             string
		companyName          string
		companyType          CompanyType
		expectedErrorMessage string
	}{
		{
			"Empty Name",
			"",
			CompanyTypeRecruiter,
			"validation error on field 'Name': Name is empty"},
		{
			"Empty CompanyType",
			"John Smith",
			"",
			"validation error on field 'CompanyType': CompanyType is invalid"},
		{
			"Invalid CompanyType", "Jane Snow",
			"Spammer",
			"validation error on field 'CompanyType': CompanyType is invalid"},
		{
			"Empty Name and CompanyType", "",
			"",
			"validation error on field 'Name': Name is empty"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			id := uuid.New()
			notes := "No notes here!"
			lastContact := time.Now().AddDate(0, 0, -3)

			request := CreateCompanyRequest{
				ID:          &id,
				Name:        test.companyName,
				CompanyType: test.companyType,
				Notes:       &notes,
				LastContact: &lastContact,
			}

			err := request.Validate()
			assert.NotNil(t, err, "err should not be nil")

			var validationErr *internalErrors.ValidationError
			assert.True(t, errors.As(err, &validationErr))

			assert.Equal(t, test.expectedErrorMessage, err.Error())
		})
	}
}

func TestCreateCompanyRequestToModel_ShouldConvertToModel(t *testing.T) {
	id := uuid.New()
	notes := "No notes here!"
	lastContact := time.Now().AddDate(0, 0, -3)

	request := CreateCompanyRequest{
		ID:          &id,
		Name:        "A random company",
		CompanyType: CompanyTypeEmployer,
		Notes:       &notes,
		LastContact: &lastContact,
	}

	model, err := request.ToModel()
	assert.Nil(t, err, "Error on CreateCompanyRequest.ToModel(): '%s'.", err)
	assert.NotNil(t, model, "model is nil")

	assert.Equal(t, *request.ID, *model.ID, "model.ID should be the same as request.ID")
	assert.Equal(t, request.Name, model.Name, "model.Name should be the same as request.Name")
	assert.Equal(t, request.CompanyType.String(), model.CompanyType.String(), "model.CompanyType should be the same value as request.CompanyType")
	assert.Equal(t, request.Notes, model.Notes, "model.Notes should be the same as request.Notes")

	modelLastContact := model.LastContact.Format(time.RFC3339)
	requestLastContact := request.LastContact.Format(time.RFC3339)
	assert.Equal(t, requestLastContact, modelLastContact, "model.LastContact should be the same as request..LastContact")

	assert.Nil(t, model.CreatedDate, "model.CreatedDate should be nil, but got '%s'", model.CreatedDate)
	assert.Nil(t, model.UpdatedDate, "model.UpdatedDate should be nil, but got '%s'", model.UpdatedDate)
}

func TestCreateCompanyRequestToModel_ShouldConvertToModelWithNilValues(t *testing.T) {

	request := CreateCompanyRequest{
		Name:        "Another company",
		CompanyType: CompanyTypeEmployer,
	}

	model, err := request.ToModel()
	assert.Nil(t, err, "Error on CreateCompanyRequest.ToModel(): '%s'.", err)
	assert.NotNil(t, model, "model is nil")

	assert.Nil(t, model.ID, "Expected ID to be nil")
	assert.Equal(t, request.Name, model.Name, "model.Name should be the same as request.Name")
	assert.Equal(t, request.CompanyType.String(), model.CompanyType.String(), "model.CompanyType should be the same as request.CompanyType")
	assert.Nil(t, model.Notes, "Expected Notes to be nil")
	assert.Nil(t, model.ID, "Expected ID to be nil")
	assert.Nil(t, model.LastContact, "model.LastContact should be nil, but got '%s'", model.LastContact)
	assert.Nil(t, model.CreatedDate, "model.CreatedDate should be nil, but got '%s'", model.CreatedDate)
	assert.Nil(t, model.UpdatedDate, "model.UpdatedDate should be nil, but got '%s'", model.UpdatedDate)
}

// -------- UpdateCompanyRequest tests: --------

func TestUpdateCompanyRequestValidate_ShouldValidateRequest(t *testing.T) {
	id := uuid.New()
	name := "Some big corp"
	var companyType CompanyType = CompanyTypeConsultancy
	notes := "The quick brown fox"
	lastContact := time.Now().AddDate(0, 0, -3)

	request := UpdateCompanyRequest{
		ID:          id,
		Name:        &name,
		CompanyType: &companyType,
		Notes:       &notes,
		LastContact: &lastContact,
	}

	err := request.Validate()
	assert.NoError(t, err)
}

func TestUpdateCompanyRequestValidate_ShouldReturnValidationErrorIfNothingToUpdate(t *testing.T) {
	id := uuid.New()

	request := UpdateCompanyRequest{
		ID: id,
	}

	err := request.Validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error: nothing to update", validationErr.Error())
}

func TestUpdateCompanyRequestValidate_ShouldReturnValidationErrorIfCompanyTypeIsInvalid(t *testing.T) {
	id := uuid.New()
	var fakeCompanyType CompanyType = "something that should never happen"

	request := UpdateCompanyRequest{
		ID:          id,
		CompanyType: &fakeCompanyType,
	}

	err := request.Validate()
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error on field 'CompanyType': CompanyType is invalid", validationErr.Error())
}

func TestUpdateCompanyRequestValidate_ShouldValidatePartialModels(t *testing.T) {
	tests := []struct {
		testName      string
		updateRequest *UpdateCompanyRequest
	}{
		{
			testName: "only Name",
			updateRequest: &UpdateCompanyRequest{
				ID:   uuid.New(),
				Name: testutil.StringPtr("SmallCorp"),
			},
		},
		{
			testName: "only CompanyType",
			updateRequest: &UpdateCompanyRequest{
				ID:          uuid.New(),
				CompanyType: CompanyType(CompanyTypeConsultancy).ToPointer(),
			},
		},
		{
			testName: "only Notes",
			updateRequest: &UpdateCompanyRequest{
				ID:    uuid.New(),
				Notes: testutil.StringPtr("Variable Notes"),
			},
		},
		{
			testName: "only LastContact",
			updateRequest: &UpdateCompanyRequest{
				ID:          uuid.New(),
				LastContact: testutil.TimePtr(time.Now().AddDate(0, 0, 2)),
			},
		},
		{
			testName: "Name and CompanyType",
			updateRequest: &UpdateCompanyRequest{
				ID:          uuid.New(),
				Name:        testutil.StringPtr("MediumCorp"),
				CompanyType: CompanyType(CompanyTypeEmployer).ToPointer(),
			},
		},
		{
			testName: "Notes and LastContact",
			updateRequest: &UpdateCompanyRequest{
				ID:          uuid.New(),
				Notes:       testutil.StringPtr("Variable Notes"),
				LastContact: testutil.TimePtr(time.Now()),
			},
		},
		{
			testName: "Name and CompanyType and LastContact",
			updateRequest: &UpdateCompanyRequest{
				ID:          uuid.New(),
				Name:        testutil.StringPtr("MediumCorp"),
				CompanyType: CompanyType(CompanyTypeRecruiter).ToPointer(),
				LastContact: testutil.TimePtr(time.Now().AddDate(0, -1, 0)),
			},
		},
		{
			testName: "CompanyType and LastContact and Notes",
			updateRequest: &UpdateCompanyRequest{
				ID:          uuid.New(),
				Name:        testutil.StringPtr("Small business"),
				CompanyType: CompanyType(CompanyTypeEmployer).ToPointer(),
				LastContact: testutil.TimePtr(time.Now().AddDate(0, 0, 3)),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := test.updateRequest.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestUpdateCompanyRequestToModel_ShouldConvertToModel(t *testing.T) {
	name := "Nameless"
	var companyType CompanyType = CompanyTypeRecruiter

	notes := "Something unimportant"
	lastContact := time.Now().AddDate(-1, 0, 0)

	updateRequest := UpdateCompanyRequest{
		ID:          uuid.New(),
		Name:        &name,
		CompanyType: &companyType,
		Notes:       &notes,
		LastContact: &lastContact,
	}

	model, err := updateRequest.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, updateRequest.ID, model.ID)
	assert.Equal(t, *updateRequest.Name, *model.Name)
	assert.Equal(t, updateRequest.CompanyType.String(), model.CompanyType.String())
	assert.Equal(t, *updateRequest.Notes, *model.Notes)
	assert.Equal(t, *updateRequest.LastContact, *model.LastContact)
}

func TestUpdateCompanyRequestToModel_ShouldConvertToModelWithNilValues(t *testing.T) {
	lastContact := time.Now().AddDate(0, -2, 0)

	updateRequest := UpdateCompanyRequest{
		ID:          uuid.New(),
		LastContact: &lastContact,
	}

	model, err := updateRequest.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, model)

	assert.Equal(t, updateRequest.ID, model.ID)
	assert.Nil(t, model.Name)
	assert.Nil(t, model.CompanyType)
	assert.Nil(t, model.Notes)
	assert.Equal(t, *updateRequest.LastContact, *model.LastContact)
}

func TestUpdateCompanyRequestToModel_ShouldReturnValidationErrorIfNothingToUpdate(t *testing.T) {
	updateRequest := UpdateCompanyRequest{
		ID: uuid.New(),
	}

	model, err := updateRequest.ToModel()
	assert.Nil(t, model)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "validation error: nothing to update", err.Error())
}

// -------- CompanyType tests: --------

func TestCompanyTypeIsValid_ShouldReturnTrue(t *testing.T) {
	employer := CompanyType(CompanyTypeEmployer)
	assert.True(t, employer.IsValid())

	recruiter := CompanyType(CompanyTypeRecruiter)
	assert.True(t, recruiter.IsValid())

	consultancy := CompanyType(CompanyTypeConsultancy)
	assert.True(t, consultancy.IsValid())
}

func TestCompanyTypeIsValid_ShouldReturnFalseOnInvalidCompanyType(t *testing.T) {
	empty := CompanyType("")
	assert.False(t, empty.IsValid())

	spammer := CompanyType("Spammer")
	assert.False(t, spammer.IsValid())
}

func TestCompanyTypeToModel_ShouldConvertToModel(t *testing.T) {
	employer := CompanyType(CompanyTypeEmployer)
	modelEmployer, err := employer.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, modelEmployer)
	assert.Equal(t, models.CompanyTypeEmployer, modelEmployer.String())

	recruiter := CompanyType(CompanyTypeRecruiter)
	modelRecruiter, err := recruiter.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, modelRecruiter)
	assert.Equal(t, models.CompanyTypeRecruiter, modelRecruiter.String())

	consultancy := CompanyType(CompanyTypeConsultancy)
	modelConsultancy, err := consultancy.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, modelConsultancy)
	assert.Equal(t, models.CompanyTypeConsultancy, modelConsultancy.String())
}

func TestCompanyTypeToModel_ShouldReturnValidationErrorOnInvalidCompanyType(t *testing.T) {
	empty := CompanyType("")
	emptyModel, err := empty.ToModel()
	assert.NotNil(t, emptyModel)
	assert.NotNil(t, err)

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "", emptyModel.String())
	assert.Equal(t, "validation error on field 'CompanyType': invalid CompanyType: ''", err.Error())

	scammer := CompanyType("scammer")
	scammerModel, err := scammer.ToModel()
	assert.NotNil(t, scammerModel)
	assert.NotNil(t, err)

	assert.True(t, errors.As(err, &validationErr))

	assert.Equal(t, "", scammerModel.String())
	assert.Equal(t, "validation error on field 'CompanyType': invalid CompanyType: 'scammer'", err.Error())
}

func TestNewCompanyType_ShouldConvertFromModel(t *testing.T) {
	employer := models.CompanyType(models.CompanyTypeEmployer)
	v1Employer, err := NewCompanyType(&employer)
	assert.NoError(t, err)
	assert.NotNil(t, v1Employer)

	assert.Equal(t, CompanyTypeEmployer, v1Employer.String())

	recruiter := models.CompanyType(models.CompanyTypeRecruiter)
	v1Recruiter, err := NewCompanyType(&recruiter)
	assert.NoError(t, err)
	assert.NotNil(t, v1Recruiter)
	assert.Equal(t, CompanyTypeRecruiter, v1Recruiter.String())

	consultancy := models.CompanyType(models.CompanyTypeConsultancy)
	v1Consultancy, err := NewCompanyType(&consultancy)
	assert.NoError(t, err)
	assert.NotNil(t, v1Consultancy)
	assert.Equal(t, CompanyTypeConsultancy, v1Consultancy.String())
}

func TestNewCompanyType_ShouldReturnInternalServiceErrorOnNilCompanyType(t *testing.T) {
	companyType, err := NewCompanyType(nil)
	assert.NotNil(t, companyType)
	assert.NotNil(t, err)

	assert.Equal(t, "", companyType.String())
	assert.Equal(t, "internal service error: Error trying to convert internal companyType to external CompanyType.", err.Error())
}

func TestNewCompanyType_ShouldReturnInternalServiceErrorOnInvalidCompanyType(t *testing.T) {
	empty := models.CompanyType("")
	emptyV1, err := NewCompanyType(&empty)
	assert.NotNil(t, emptyV1)
	assert.NotNil(t, err)
	assert.Equal(t, "", emptyV1.String())

	var internalServiceError *internalErrors.InternalServiceError
	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, "internal service error: Error converting internal CompanyType to external CompanyType: ''", err.Error())

	scammer := models.CompanyType("scammer")
	scammerV1, err := NewCompanyType(&scammer)
	assert.NotNil(t, scammerV1)
	assert.NotNil(t, err)
	assert.Equal(t, "", scammerV1.String())

	assert.True(t, errors.As(err, &internalServiceError))
	assert.Equal(t, "internal service error: Error converting internal CompanyType to external CompanyType: 'scammer'", err.Error())
}
