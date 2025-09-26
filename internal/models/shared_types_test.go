package models

import (
	"errors"
	internalErrors "jobsearchtracker/internal/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	var validationErr *internalErrors.ValidationError
	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: invalid Type ''", validationErr.Error())

	namesType, err := NewIncludeExtraDataType("names")
	assert.NotNil(t, namesType)
	assert.NotNil(t, err)

	assert.Equal(t, "", namesType.String())

	assert.True(t, errors.As(err, &validationErr))
	assert.Equal(t, "validation error: invalid Type 'names'", validationErr.Error())
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
