package handlers

import (
	"jobsearchtracker/internal/api/v1/requests"
	"jobsearchtracker/internal/models"
	"strings"
)

func GetExtraDataTypeParam(urlParamValue string) (*models.IncludeExtraDataType, error) {
	var includeExtraDataType requests.IncludeExtraDataType
	if urlParamValue == "" {
		includeExtraDataType = requests.IncludeExtraDataTypeNone
	} else {
		var err error

		// can return ValidationError
		includeExtraDataType, err = requests.NewIncludeExtraDataType(strings.ToLower(urlParamValue))

		if err != nil {
			return nil, err
		}
	}
	includeApplicationsTypeModel, err := includeExtraDataType.ToModel()

	return &includeApplicationsTypeModel, err
}
