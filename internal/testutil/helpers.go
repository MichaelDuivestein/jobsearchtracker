package testutil

import (
	"jobsearchtracker/pkg/timeutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ToPtr[Type any](variable Type) *Type {
	return &variable
}

func AssertEqualFormattedDateTimes(t *testing.T, dateTime1 *time.Time, dateTime2 *time.Time) {
	formattedDateTime1 := dateTime1.Format(timeutil.RFC3339Milli_Read)
	formattedDateTime2 := dateTime2.Format(timeutil.RFC3339Milli_Read)
	assert.Equal(t, formattedDateTime1, formattedDateTime2)
}

func AssertDateTimesWithinDelta(t *testing.T, dateTime1 *time.Time, dateTime2 *time.Time, delta time.Duration) {
	diff := dateTime1.Sub(*dateTime2)
	if diff < 0 {
		diff = -diff
	}
	assert.LessOrEqual(t, diff, delta)
}
