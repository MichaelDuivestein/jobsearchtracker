package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ToPtr[Type any](variable Type) *Type {
	return &variable
}

func AssertEqualFormattedDateTimes(t *testing.T, dateTime1 *time.Time, dateTime2 *time.Time) {
	formattedDateTime1 := dateTime1.Format(time.RFC3339)
	formattedDateTime2 := dateTime2.Format(time.RFC3339)
	assert.Equal(t, formattedDateTime1, formattedDateTime2)
}
