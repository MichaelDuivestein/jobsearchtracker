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
		builder.WriteString(substring)

		// don't write the separator after the last index
		if index != len(*stringParts)-1 {
			builder.WriteString(separator)
		}
	}

	if postfix != nil {
		builder.WriteString(*postfix)
	}

	return builder.String(), nil
}
