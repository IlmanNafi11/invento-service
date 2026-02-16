package errors

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestAppError_Error tests the Error() method implementation.
func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error with internal error",
			appError: &AppError{
				Code:       ErrValidation,
				Message:    "Test message",
				HTTPStatus: fiber.StatusBadRequest,
				Internal:   fmt.Errorf("internal error"),
				Timestamp:  time.Now(),
			},
			expected: "[" + ErrValidation + "] Test message: internal error",
		},
		{
			name: "error without internal error",
			appError: &AppError{
				Code:       ErrNotFound,
				Message:    "Resource not found",
				HTTPStatus: fiber.StatusNotFound,
				Internal:   nil,
				Timestamp:  time.Now(),
			},
			expected: "[" + ErrNotFound + "] Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAppError_Unwrap tests the Unwrap() method implementation.
func TestAppError_Unwrap(t *testing.T) {
	internalErr := fmt.Errorf("internal error")
	appErr := &AppError{
		Code:       ErrInternal,
		Message:    "Server error",
		HTTPStatus: fiber.StatusInternalServerError,
		Internal:   internalErr,
		Timestamp:  time.Now(),
	}

	result := appErr.Unwrap()
	assert.Equal(t, internalErr, result)
}

// TestAppError_Unwrap_Nil tests Unwrap() when internal error is nil.
func TestAppError_Unwrap_Nil(t *testing.T) {
	appErr := &AppError{
		Code:       ErrValidation,
		Message:    "Validation error",
		HTTPStatus: fiber.StatusBadRequest,
		Internal:   nil,
		Timestamp:  time.Now(),
	}

	result := appErr.Unwrap()
	assert.Nil(t, result)
}

// TestErrorWrapping tests error wrapping and unwrapping with errors.Is and errors.As.
func TestErrorWrapping(t *testing.T) {
	// Create a wrapped error for testing
	wrappedErr := fmt.Errorf("wrapped error")

	tests := []struct {
		name        string
		appError    *AppError
		wrappedErr  error
		shouldMatch bool
	}{
		{
			name: "validation error with wrapped error",
			appError: NewValidationError(
				"Validation failed",
				wrappedErr,
			),
			wrappedErr:  wrappedErr,
			shouldMatch: true,
		},
		{
			name: "internal error with wrapped standard error",
			appError: NewInternalError(
				fmt.Errorf("database error"),
			),
			wrappedErr:  nil,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test errors.Is for unwrapping
			if tt.wrappedErr != nil {
				isMatch := errors.Is(tt.appError, tt.wrappedErr)
				assert.Equal(t, tt.shouldMatch, isMatch)
			}

			// Test that Unwrap returns the internal error
			if tt.appError.Internal != nil {
				unwrapped := tt.appError.Unwrap()
				assert.Equal(t, tt.appError.Internal, unwrapped)
			}
		})
	}
}

// TestErrorAs tests error type assertion with errors.As.
func TestErrorAs(t *testing.T) {
	appErr := NewValidationError("test message", fmt.Errorf("internal"))

	// Test errors.As can extract AppError
	var extractedErr *AppError
	assert.True(t, errors.As(appErr, &extractedErr))
	assert.Equal(t, appErr.Code, extractedErr.Code)
	assert.Equal(t, appErr.Message, extractedErr.Message)
}

// TestHTTPStatusCodes verifies all HTTP status codes are correct.
func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		constructor    func() *AppError
		expectedStatus int
	}{
		{
			name: "ValidationError returns 400",
			constructor: func() *AppError {
				return NewValidationError("test", nil)
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "UnauthorizedError returns 401",
			constructor: func() *AppError {
				return NewUnauthorizedError("test")
			},
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name: "ForbiddenError returns 403",
			constructor: func() *AppError {
				return NewForbiddenError("test")
			},
			expectedStatus: fiber.StatusForbidden,
		},
		{
			name: "NotFoundError returns 404",
			constructor: func() *AppError {
				return NewNotFoundError("test")
			},
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name: "ConflictError returns 409",
			constructor: func() *AppError {
				return NewConflictError("test")
			},
			expectedStatus: fiber.StatusConflict,
		},
		{
			name: "InternalError returns 500",
			constructor: func() *AppError {
				return NewInternalError(fmt.Errorf("test"))
			},
			expectedStatus: fiber.StatusInternalServerError,
		},
		{
			name: "TusVersionError returns 412",
			constructor: func() *AppError {
				return NewTusVersionError("1.0.0")
			},
			expectedStatus: fiber.StatusPreconditionFailed,
		},
		{
			name: "TusOffsetError returns 409",
			constructor: func() *AppError {
				return NewTusOffsetError(100, 50)
			},
			expectedStatus: fiber.StatusConflict,
		},
		{
			name: "TusInactiveError returns 423",
			constructor: func() *AppError {
				return NewTusInactiveError()
			},
			expectedStatus: fiber.StatusLocked,
		},
		{
			name: "TusCompletedError returns 409",
			constructor: func() *AppError {
				return NewTusCompletedError()
			},
			expectedStatus: fiber.StatusConflict,
		},
		{
			name: "PayloadTooLargeError returns 413",
			constructor: func() *AppError {
				return NewPayloadTooLargeError("test")
			},
			expectedStatus: fiber.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.constructor()
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
		})
	}
}

// TestErrorCodes verifies all error codes match constants.
func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() *AppError
		expectedCode string
	}{
		{
			name: "ValidationError code",
			constructor: func() *AppError {
				return NewValidationError("test", nil)
			},
			expectedCode: ErrValidation,
		},
		{
			name: "UnauthorizedError code",
			constructor: func() *AppError {
				return NewUnauthorizedError("test")
			},
			expectedCode: ErrUnauthorized,
		},
		{
			name: "ForbiddenError code",
			constructor: func() *AppError {
				return NewForbiddenError("test")
			},
			expectedCode: ErrForbidden,
		},
		{
			name: "NotFoundError code",
			constructor: func() *AppError {
				return NewNotFoundError("test")
			},
			expectedCode: ErrNotFound,
		},
		{
			name: "ConflictError code",
			constructor: func() *AppError {
				return NewConflictError("test")
			},
			expectedCode: ErrConflict,
		},
		{
			name: "InternalError code",
			constructor: func() *AppError {
				return NewInternalError(fmt.Errorf("test"))
			},
			expectedCode: ErrInternal,
		},
		{
			name: "TusVersionError code",
			constructor: func() *AppError {
				return NewTusVersionError("1.0.0")
			},
			expectedCode: ErrTusVersionMismatch,
		},
		{
			name: "TusOffsetError code",
			constructor: func() *AppError {
				return NewTusOffsetError(100, 50)
			},
			expectedCode: ErrTusOffsetMismatch,
		},
		{
			name: "TusInactiveError code",
			constructor: func() *AppError {
				return NewTusInactiveError()
			},
			expectedCode: ErrTusInactive,
		},
		{
			name: "TusCompletedError code",
			constructor: func() *AppError {
				return NewTusCompletedError()
			},
			expectedCode: ErrTusAlreadyCompleted,
		},
		{
			name: "PayloadTooLargeError code",
			constructor: func() *AppError {
				return NewPayloadTooLargeError("test")
			},
			expectedCode: ErrPayloadTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.constructor()
			assert.Equal(t, tt.expectedCode, result.Code)
		})
	}
}

// TestTimestampSet verifies timestamp is set on all errors.
func TestTimestampSet(t *testing.T) {
	before := time.Now()

	// Test all constructors set timestamps
	constructors := []func() *AppError{
		func() *AppError { return NewValidationError("test", nil) },
		func() *AppError { return NewUnauthorizedError("test") },
		func() *AppError { return NewForbiddenError("test") },
		func() *AppError { return NewNotFoundError("test") },
		func() *AppError { return NewConflictError("test") },
		func() *AppError { return NewInternalError(fmt.Errorf("test")) },
		func() *AppError { return NewTusVersionError("1.0.0") },
		func() *AppError { return NewTusOffsetError(100, 50) },
		func() *AppError { return NewTusInactiveError() },
		func() *AppError { return NewTusCompletedError() },
		func() *AppError { return NewPayloadTooLargeError("test") },
	}

	for _, fn := range constructors {
		result := fn()
		assert.False(t, result.Timestamp.IsZero(), "Timestamp should be set")
		assert.True(t, result.Timestamp.After(before) || result.Timestamp.Equal(before),
			"Timestamp should be current or after before time")
	}
}
