package errors

import "fmt"

type ConflictError struct {
	Message string
}

func NewConflictError(field string) *ConflictError {
	return &ConflictError{field}
}

func (err *ConflictError) Error() string {
	return fmt.Sprintf("conflict error on insert: %s", err.Message)
}

type InternalServiceError struct {
	Message string
}

func NewInternalServiceError(message string) *InternalServiceError {
	return &InternalServiceError{message}
}

func (err *InternalServiceError) Error() string {
	return fmt.Sprintf("internal service error: %s", err.Message)
}

type NotFoundError struct {
	message string
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message}
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("error: object not found: %s", err.message)
}

type ValidationError struct {
	Field   *string
	Message string
}

func NewValidationError(field *string, message string) *ValidationError {
	return &ValidationError{field, message}
}

func (err *ValidationError) Error() string {
	if err.Field == nil {
		return fmt.Sprintf("validation error: %s", err.Message)
	}
	return fmt.Sprintf("validation error on field '%s': %s", *err.Field, err.Message)
}
