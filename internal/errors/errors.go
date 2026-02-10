package errors

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// NewValidationError creates a validation error with HTTP 400 status.
//
// Example:
//
//	return errors.NewValidationError("Email wajib diisi", nil)
//	return errors.NewValidationError("Format email salah", fmt.Errorf("invalid format"))
func NewValidationError(message string, internal error) *AppError {
	return &AppError{
		Code:       ErrValidation,
		Message:    message,
		HTTPStatus: fiber.StatusBadRequest,
		Internal:   internal,
		Timestamp:  time.Now(),
	}
}

// NewUnauthorizedError creates an unauthorized error with HTTP 401 status.
// Use when authentication credentials are missing or invalid.
//
// Example:
//
//	return errors.NewUnauthorizedError("Token tidak valid")
//	return errors.NewUnauthorizedError("Sesi telah berakhir")
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       ErrUnauthorized,
		Message:    message,
		HTTPStatus: fiber.StatusUnauthorized,
		Timestamp:  time.Now(),
	}
}

// NewForbiddenError creates a forbidden error with HTTP 403 status.
// Use when authenticated user lacks permission for an action.
//
// Example:
//
//	return errors.NewForbiddenError("Anda tidak memiliki akses ke modul ini")
//	return errors.NewForbiddenError("Hanya admin yang dapat menghapus data")
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       ErrForbidden,
		Message:    message,
		HTTPStatus: fiber.StatusForbidden,
		Timestamp:  time.Now(),
	}
}

// NewNotFoundError creates a not found error with HTTP 404 status.
// The resource parameter is used to generate a default message if needed.
//
// Example:
//
//	return errors.NewNotFoundError("User")
//	return errors.NewNotFoundError("Project dengan id 123")
//	return errors.NewNotFoundError("") // Uses default: "Data tidak ditemukan"
func NewNotFoundError(resource string) *AppError {
	message := "Data tidak ditemukan"
	if resource != "" {
		message = resource + " tidak ditemukan"
	}

	return &AppError{
		Code:       ErrNotFound,
		Message:    message,
		HTTPStatus: fiber.StatusNotFound,
		Timestamp:  time.Now(),
	}
}

// NewConflictError creates a conflict error with HTTP 409 status.
// Use when a resource already exists or conflicts with existing data.
//
// Example:
//
//	return errors.NewConflictError("Email sudah terdaftar")
//	return errors.NewConflictError("Nomor telepon sudah digunakan")
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrConflict,
		Message:    message,
		HTTPStatus: fiber.StatusConflict,
		Timestamp:  time.Now(),
	}
}

// NewInternalError creates an internal server error with HTTP 500 status.
// Wraps the internal error for debugging while hiding details from the user.
//
// Example:
//
//	return errors.NewInternalError(fmt.Errorf("database connection failed"))
//	return errors.NewInternalError(err)
func NewInternalError(internal error) *AppError {
	return &AppError{
		Code:       ErrInternal,
		Message:    "Terjadi kesalahan pada server",
		HTTPStatus: fiber.StatusInternalServerError,
		Internal:   internal,
		Timestamp:  time.Now(),
	}
}

// NewTusVersionError creates an error for invalid/unsupported TUS protocol version (HTTP 412).
// Used when TUS protocol version header is missing or doesn't match supported version.
//
// Example:
//
//	return errors.NewTusVersionError("1.0.0")
func NewTusVersionError(supportedVersion string) *AppError {
	return &AppError{
		Code:       ErrTusVersionMismatch,
		Message:    "Versi TUS protocol tidak didukung, gunakan " + supportedVersion,
		HTTPStatus: fiber.StatusPreconditionFailed,
		Timestamp:  time.Now(),
	}
}

// NewTusOffsetError creates an error for invalid upload offset (HTTP 409).
// Used when client sends chunk with incorrect offset (doesn't match server state).
//
// Example:
//
//	return errors.NewTusOffsetError(expectedOffset, actualOffset)
func NewTusOffsetError(expectedOffset, actualOffset int64) *AppError {
	return &AppError{
		Code:       ErrTusOffsetMismatch,
		Message:    "Offset tidak valid",
		HTTPStatus: fiber.StatusConflict,
		Timestamp:  time.Now(),
	}
}

// NewTusInactiveError creates an error for inactive upload (HTTP 423).
// Used when upload session is not active (expired, paused, or locked).
//
// Example:
//
//	return errors.NewTusInactiveError()
func NewTusInactiveError() *AppError {
	return &AppError{
		Code:       ErrTusInactive,
		Message:    "Upload tidak aktif",
		HTTPStatus: fiber.StatusLocked,
		Timestamp:  time.Now(),
	}
}

// NewTusCompletedError creates an error for already completed upload (HTTP 409).
// Used when attempting to modify an upload that's already finished.
//
// Example:
//
//	return errors.NewTusCompletedError()
func NewTusCompletedError() *AppError {
	return &AppError{
		Code:       ErrTusAlreadyCompleted,
		Message:    "Upload sudah selesai",
		HTTPStatus: fiber.StatusConflict,
		Timestamp:  time.Now(),
	}
}

// NewPayloadTooLargeError creates an error for request payload exceeding limits (HTTP 413).
// Used when file size or chunk size exceeds configured maximum.
//
// Example:
//
//	return errors.NewPayloadTooLargeError("Ukuran file melebihi batas maksimal 100MB")
func NewPayloadTooLargeError(message string) *AppError {
	if message == "" {
		message = "Ukuran data melebihi batas maksimal"
	}
	return &AppError{
		Code:       ErrPayloadTooLarge,
		Message:    message,
		HTTPStatus: fiber.StatusRequestEntityTooLarge,
		Timestamp:  time.Now(),
	}
}
