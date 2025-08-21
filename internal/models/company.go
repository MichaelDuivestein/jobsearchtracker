package models

import (
	"github.com/google/uuid"
	"time"
)

type Company struct {
	ID          uuid.UUID
	Name        string
	CompanyType CompanyType
	Notes       *string
	LastContact *time.Time
	CreatedDate time.Time
	UpdatedDate *time.Time
}
