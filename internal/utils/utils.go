package utils

import (
	"errors"
	"strings"
)

func JoinToString(stringParts *[]string, prefix *string, separator string, postfix *string) (string, error) {
	if stringParts == nil {
		return "", errors.New("nothing to join")
	}
	if len(*stringParts) == 0 {
		return "", errors.New("nothing to join")
	}

	var builder strings.Builder

	if prefix != nil {
		builder.WriteString(*prefix)
	}

	for index, substring := range *stringParts {
		if index > 0 {
			builder.WriteString(separator)
		}

		builder.WriteString(substring)
	}

	if postfix != nil {
		builder.WriteString(*postfix)
	}

	return builder.String(), nil
}
