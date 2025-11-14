package handlers

import (
	"errors"
	"jobsearchtracker/internal/models"
	"testing"

	internalErrors "jobsearchtracker/internal/errors"

	"github.com/stretchr/testify/assert"
)

func TestGetIncludeExtraDataParam_ShouldMapToCorrectType(t *testing.T) {
	all, err := GetExtraDataTypeParam("all")
	assert.NoError(t, err)
	assert.Equal(t, models.IncludeExtraDataTypeAll, all.String())

	ids, err := GetExtraDataTypeParam("ids")
	assert.NoError(t, err)
	assert.Equal(t, models.IncludeExtraDataTypeIDs, ids.String())

	none, err := GetExtraDataTypeParam("none")
	assert.NoError(t, err)
	assert.Equal(t, models.IncludeExtraDataTypeNone, none.String())
}

func TestGetIncludeExtraDataParam_ShouldBeCaseInsensitive(t *testing.T) {
	all, err := GetExtraDataTypeParam("ALL")
	assert.NoError(t, err)
	assert.Equal(t, models.IncludeExtraDataTypeAll, all.String())

	ids, err := GetExtraDataTypeParam("IDs")
	assert.NoError(t, err)
	assert.Equal(t, models.IncludeExtraDataTypeIDs, ids.String())

	none, err := GetExtraDataTypeParam("NoNe")
	assert.NoError(t, err)
	assert.Equal(t, models.IncludeExtraDataTypeNone, none.String())
}

func TestGetIncludeExtraDataParam_ShouldMapToNoneIfUrlParamValueIsEmpty(t *testing.T) {
	dataType, err := GetExtraDataTypeParam("")
	assert.NoError(t, err)

	assert.Equal(t, models.IncludeExtraDataTypeNone, dataType.String())
}

func TestGetIncludeExtraDataParam_ShouldReturnValidationErrIfUrlParamValueIsNotMappable(t *testing.T) {
	dataType, err := GetExtraDataTypeParam("names")
	assert.Nil(t, dataType)
	assert.Error(t, err)

	var validationError *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationError))
	assert.Equal(t, "validation error: invalid Type 'names'", validationError.Error())
}
