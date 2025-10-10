package requests

import (
	internalErrors "jobsearchtracker/internal/errors"
	"jobsearchtracker/internal/models"
	"log/slog"
)

// IncludeExtraDataType represents how much additional data to send. "all" will return all data, "ids" will return only IDs, "none" will return no extra data.
//
// @enum all,ids,none
type IncludeExtraDataType string

const (
	IncludeExtraDataTypeAll  = "all"
	IncludeExtraDataTypeIDs  = "ids"
	IncludeExtraDataTypeNone = "none"
)

func NewIncludeExtraDataType(includeExtraDataType string) (IncludeExtraDataType, error) {
	switch includeExtraDataType {
	case IncludeExtraDataTypeAll:
		return IncludeExtraDataTypeAll, nil
	case IncludeExtraDataTypeIDs:
		return IncludeExtraDataTypeIDs, nil
	case IncludeExtraDataTypeNone:
		return IncludeExtraDataTypeNone, nil
	default:
		return "", internalErrors.NewValidationError(nil, "invalid Type '"+includeExtraDataType+"'")
	}
}

func (includeExtraDataType IncludeExtraDataType) IsValid() bool {
	switch includeExtraDataType {
	case IncludeExtraDataTypeAll, IncludeExtraDataTypeIDs, IncludeExtraDataTypeNone:
		return true
	}
	return false
}

func (includeExtraDataType IncludeExtraDataType) String() string {
	return string(includeExtraDataType)
}

func (includeExtraDataType IncludeExtraDataType) ToModel() (models.IncludeExtraDataType, error) {
	switch includeExtraDataType {
	case IncludeExtraDataTypeAll:
		return models.IncludeExtraDataTypeAll, nil
	case IncludeExtraDataTypeIDs:
		return models.IncludeExtraDataTypeIDs, nil
	case IncludeExtraDataTypeNone:
		return models.IncludeExtraDataTypeNone, nil
	default:
		slog.Info("v1.types.toModel: Invalid IncludeExtraDataType: '" + includeExtraDataType.String() + "'")
		includeExtraDataTypeString := "IncludeExtraDataType"
		return "", internalErrors.NewValidationError(
			&includeExtraDataTypeString, "invalid IncludeExtraDataType: '"+includeExtraDataType.String()+"'")
	}
}
