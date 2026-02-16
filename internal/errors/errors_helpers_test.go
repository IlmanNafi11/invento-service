package errors

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestNewValidationError tests the ValidationError constructor.
func TestNewValidationError(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
