package errors

import (
	"fmt"
	"time"
)

// AppError is a standardized error type for application-level errors.
// It provides machine-readable error codes, user-facing messages (Indonesian),
// HTTP status codes, and supports error wrapping for debugging.
//
// Example usage:
//
//	err := errors.NewValidationError("Email tidak valid", fmt.Errorf("invalid email format"))
//	if err != nil {
//	    return err // Can be handled by helper.SendAppError()
//	}
type AppError struct {
	Code       string    // Machine-readable error code (e.g., "VALIDATION_ERROR")
	Message    string    // User-facing message in Indonesian
	HTTPStatus int       // HTTP status code to return
	Internal   error     // Wrapped internal error for debugging (optional)
	Timestamp  time.Time // When the error occurred
}

// Error returns the error message, implementing the error interface.
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped internal error, supporting error unwrapping.
// This allows using errors.Is() and errors.As() with the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Internal
}
