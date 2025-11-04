package requests

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -------- IncludeExtraDataType tests: --------

func TestNewIncludeExtraDataType_ShouldReturnIncludeExtraDataType(t *testing.T) {
	all, err := NewIncludeExtraDataType("all")
	assert.NoError(t, err)
	var expectedAll IncludeExtraDataType = IncludeExtraDataTypeAll
	assert.Equal(t, expectedAll, all)

	ids, err := NewIncludeExtraDataType("ids")
	assert.NoError(t, err)
	var expectedIDs IncludeExtraDataType = IncludeExtraDataTypeIDs
	assert.Equal(t, expectedIDs, ids)

	none, err := NewIncludeExtraDataType("none")
	assert.NoError(t, err)
	var expectedNone IncludeExtraDataType = IncludeExtraDataTypeNone
	assert.Equal(t, expectedNone, none)
}

func TestNewIncludeExtraDataType_ShouldReturnErrorForWrongValue(t *testing.T) {
	emptyType, err := NewIncludeExtraDataType("")
	assert.NotNil(t, emptyType)
	assert.NotNil(t, err)

	assert.Equal(t, "", emptyType.String())

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: invalid Type ''", validationError.Error())

	namesType, err := NewIncludeExtraDataType("names")
	assert.NotNil(t, namesType)
	assert.NotNil(t, err)

	assert.Equal(t, "", namesType.String())

	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: invalid Type 'names'", validationError.Error())
}

func TestIncludeExtraDataTypeIsValid_ShouldReturnTrue(t *testing.T) {
	all := IncludeExtraDataType(IncludeExtraDataTypeAll)
	assert.True(t, all.IsValid())

	ids := IncludeExtraDataType(IncludeExtraDataTypeIDs)
	assert.True(t, ids.IsValid())

	none := IncludeExtraDataType(IncludeExtraDataTypeNone)
	assert.True(t, none.IsValid())
}

func TestIncludeExtraDataTypeIsValid_ShouldReturnFalseOnInvalidType(t *testing.T) {
	empty := IncludeExtraDataType("")
	assert.False(t, empty.IsValid())

	name := IncludeExtraDataType("name")
	assert.False(t, name.IsValid())
}

func TestIncludeExtraDataTypeToModel_ShouldConvertToModel(t *testing.T) {
	all := IncludeExtraDataType(IncludeExtraDataTypeAll)
	modelAll, err := all.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, modelAll)
	assert.Equal(t, models.IncludeExtraDataTypeAll, modelAll.String())

	ids := IncludeExtraDataType(IncludeExtraDataTypeIDs)
	modelIDs, err := ids.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, modelIDs)
	assert.Equal(t, models.IncludeExtraDataTypeIDs, modelIDs.String())

	none := IncludeExtraDataType(IncludeExtraDataTypeNone)
	modelNone, err := none.ToModel()
	assert.NoError(t, err)
	assert.NotNil(t, modelNone)
	assert.Equal(t, models.IncludeExtraDataTypeNone, modelNone.String())
}

func TestIncludeExtraDataTypeToModel_ShouldReturnValidationErrorOnInvalidType(t *testing.T) {
	empty := IncludeExtraDataType("")
	emptyModel, err := empty.ToModel()
	assert.NotNil(t, emptyModel)
	assert.NotNil(t, err)

	assert.Equal(t, "", emptyModel.String())

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'IncludeExtraDataType': invalid IncludeExtraDataType: ''",
		validationError.Error())

	name := IncludeExtraDataType("name")
	nameModel, err := name.ToModel()
	assert.NotNil(t, nameModel)
	assert.NotNil(t, err)

	assert.Equal(t, "", nameModel.String())

	assert.True(t, errors.As(err, &validationError))
	assert.Equal(
		t,
		"validation error on field 'IncludeExtraDataType': invalid IncludeExtraDataType: 'name'",
		validationError.Error())
}
