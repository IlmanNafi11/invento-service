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

// TestNewValidationError tests the ValidationError constructor.
func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		internalErr    error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "validation error with internal error",
			message:        "Email tidak valid",
			internalErr:    fmt.Errorf("invalid email format"),
			expectedCode:   ErrValidation,
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "validation error without internal error",
			message:        "Nama wajib diisi",
			internalErr:    nil,
			expectedCode:   ErrValidation,
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewValidationError(tt.message, tt.internalErr)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.message, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Equal(t, tt.internalErr, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewUnauthorizedError tests the UnauthorizedError constructor.
func TestNewUnauthorizedError(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "unauthorized error with custom message",
			message:        "Token tidak valid",
			expectedCode:   ErrUnauthorized,
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "unauthorized error with session expired message",
			message:        "Sesi telah berakhir",
			expectedCode:   ErrUnauthorized,
			expectedStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewUnauthorizedError(tt.message)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.message, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Nil(t, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewForbiddenError tests the ForbiddenError constructor.
func TestNewForbiddenError(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "forbidden error for module access",
			message:        "Anda tidak memiliki akses ke modul ini",
			expectedCode:   ErrForbidden,
			expectedStatus: fiber.StatusForbidden,
		},
		{
			name:           "forbidden error for admin only action",
			message:        "Hanya admin yang dapat menghapus data",
			expectedCode:   ErrForbidden,
			expectedStatus: fiber.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewForbiddenError(tt.message)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.message, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Nil(t, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewNotFoundError tests the NotFoundError constructor.
func TestNewNotFoundError(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		expectedCode   string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "not found with resource name",
			resource:       "User",
			expectedCode:   ErrNotFound,
			expectedStatus: fiber.StatusNotFound,
			expectedMsg:    "User tidak ditemukan",
		},
		{
			name:           "not found with detailed resource",
			resource:       "Project dengan id 123",
			expectedCode:   ErrNotFound,
			expectedStatus: fiber.StatusNotFound,
			expectedMsg:    "Project dengan id 123 tidak ditemukan",
		},
		{
			name:           "not found with empty resource",
			resource:       "",
			expectedCode:   ErrNotFound,
			expectedStatus: fiber.StatusNotFound,
			expectedMsg:    "Data tidak ditemukan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewNotFoundError(tt.resource)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Nil(t, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewConflictError tests the ConflictError constructor.
func TestNewConflictError(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "conflict error for duplicate email",
			message:        "Email sudah terdaftar",
			expectedCode:   ErrConflict,
			expectedStatus: fiber.StatusConflict,
		},
		{
			name:           "conflict error for duplicate phone",
			message:        "Nomor telepon sudah digunakan",
			expectedCode:   ErrConflict,
			expectedStatus: fiber.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewConflictError(tt.message)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.message, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Nil(t, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewInternalError tests the InternalError constructor.
func TestNewInternalError(t *testing.T) {
	tests := []struct {
		name           string
		internalErr    error
		expectedCode   string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "internal error with database error",
			internalErr:    fmt.Errorf("database connection failed"),
			expectedCode:   ErrInternal,
			expectedStatus: fiber.StatusInternalServerError,
			expectedMsg:    "Terjadi kesalahan pada server",
		},
		{
			name:           "internal error with wrapped error",
			internalErr:    fmt.Errorf("context deadline exceeded"),
			expectedCode:   ErrInternal,
			expectedStatus: fiber.StatusInternalServerError,
			expectedMsg:    "Terjadi kesalahan pada server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewInternalError(tt.internalErr)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Equal(t, tt.internalErr, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewTusVersionError tests the TusVersionError constructor.
func TestNewTusVersionError(t *testing.T) {
	tests := []struct {
		name             string
		supportedVersion string
		expectedCode     string
		expectedStatus   int
		expectedMsg      string
	}{
		{
			name:             "TUS version error with version 1.0.0",
			supportedVersion: "1.0.0",
			expectedCode:     ErrTusVersionMismatch,
			expectedStatus:   fiber.StatusPreconditionFailed,
			expectedMsg:      "Versi TUS protocol tidak didukung, gunakan 1.0.0",
		},
		{
			name:             "TUS version error with version 1.1.0",
			supportedVersion: "1.1.0",
			expectedCode:     ErrTusVersionMismatch,
			expectedStatus:   fiber.StatusPreconditionFailed,
			expectedMsg:      "Versi TUS protocol tidak didukung, gunakan 1.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTusVersionError(tt.supportedVersion)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Nil(t, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewTusOffsetError tests the TusOffsetError constructor.
func TestNewTusOffsetError(t *testing.T) {
	tests := []struct {
		name           string
		expectedOffset int64
		actualOffset   int64
		expectedCode   string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "TUS offset mismatch with different offsets",
			expectedOffset: 1024,
			actualOffset:   512,
			expectedCode:   ErrTusOffsetMismatch,
			expectedStatus: fiber.StatusConflict,
			expectedMsg:    "Upload offset tidak sesuai. Diharapkan: 1024, diterima: 512",
		},
		{
			name:           "TUS offset mismatch at zero",
			expectedOffset: 0,
			actualOffset:   100,
			expectedCode:   ErrTusOffsetMismatch,
			expectedStatus: fiber.StatusConflict,
			expectedMsg:    "Upload offset tidak sesuai. Diharapkan: 0, diterima: 100",
		},
		{
			name:           "TUS offset mismatch with large values",
			expectedOffset: 1000000000,
			actualOffset:   500000000,
			expectedCode:   ErrTusOffsetMismatch,
			expectedStatus: fiber.StatusConflict,
			expectedMsg:    "Upload offset tidak sesuai. Diharapkan: 1000000000, diterima: 500000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTusOffsetError(tt.expectedOffset, tt.actualOffset)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Nil(t, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestNewTusInactiveError tests the TusInactiveError constructor.
func TestNewTusInactiveError(t *testing.T) {
	result := NewTusInactiveError()

	assert.NotNil(t, result)
	assert.Equal(t, ErrTusInactive, result.Code)
	assert.Equal(t, "Upload tidak aktif", result.Message)
	assert.Equal(t, fiber.StatusLocked, result.HTTPStatus)
	assert.Nil(t, result.Internal)
	assert.False(t, result.Timestamp.IsZero())
}

// TestNewTusCompletedError tests the TusCompletedError constructor.
func TestNewTusCompletedError(t *testing.T) {
	result := NewTusCompletedError()

	assert.NotNil(t, result)
	assert.Equal(t, ErrTusAlreadyCompleted, result.Code)
	assert.Equal(t, "Upload sudah selesai", result.Message)
	assert.Equal(t, fiber.StatusConflict, result.HTTPStatus)
	assert.Nil(t, result.Internal)
	assert.False(t, result.Timestamp.IsZero())
}

// TestNewPayloadTooLargeError tests the PayloadTooLargeError constructor.
func TestNewPayloadTooLargeError(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		expectedCode   string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "payload too large with custom message",
			message:        "Ukuran file melebihi batas maksimal 100MB",
			expectedCode:   ErrPayloadTooLarge,
			expectedStatus: fiber.StatusRequestEntityTooLarge,
			expectedMsg:    "Ukuran file melebihi batas maksimal 100MB",
		},
		{
			name:           "payload too large with chunk size message",
			message:        "Ukuran chunk melebihi 5MB",
			expectedCode:   ErrPayloadTooLarge,
			expectedStatus: fiber.StatusRequestEntityTooLarge,
			expectedMsg:    "Ukuran chunk melebihi 5MB",
		},
		{
			name:           "payload too large with empty message uses default",
			message:        "",
			expectedCode:   ErrPayloadTooLarge,
			expectedStatus: fiber.StatusRequestEntityTooLarge,
			expectedMsg:    "Ukuran data melebihi batas maksimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPayloadTooLargeError(tt.message)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCode, result.Code)
			assert.Equal(t, tt.expectedMsg, result.Message)
			assert.Equal(t, tt.expectedStatus, result.HTTPStatus)
			assert.Nil(t, result.Internal)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
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
