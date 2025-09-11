package testutil

import (
	"time"

	"github.com/google/uuid"
)

func StringPtr(input string) *string {
	return &input
}

func UUIDPtr(input uuid.UUID) *uuid.UUID {
	return &input
}

func TimePtr(date time.Time) *time.Time {
	return &date
}
