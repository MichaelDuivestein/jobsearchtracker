package models

import (
	internalErrors "jobsearchtracker/internal/errors"
)

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
