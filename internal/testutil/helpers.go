package testutil

import "time"

func StringPtr(input string) *string {
	return &input
}

func TimePtr(date time.Time) *time.Time {
	return &date
}
