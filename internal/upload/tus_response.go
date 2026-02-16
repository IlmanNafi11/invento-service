package upload

import (
	"invento-service/internal/httputil"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendTusInitiateResponse(c *fiber.Ctx, uploadID string, uploadURL string, fileSize int64) error {
	SetTusResponseHeaders(c, 0, fileSize)
	SetTusLocationHeader(c, uploadURL)

	response := map[string]interface{}{
		"status":  "success",
		"message": "Upload berhasil diinisiasi",
		"code":    fiber.StatusCreated,
		"data": map[string]interface{}{
			"upload_id":  uploadID,
			"upload_url": uploadURL,
			"offset":     0,
			"length":     fileSize,
		},
		"timestamp": time.Now(),
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func SendTusChunkResponse(c *fiber.Ctx, newOffset int64) error {
	c.Set(HeaderTusResumable, TusVersion)
	c.Set(HeaderUploadOffset, strconv.FormatInt(newOffset, 10))

	return c.SendStatus(fiber.StatusNoContent)
}

func SendTusHeadResponse(c *fiber.Ctx, offset int64, length int64) error {
	c.Set(HeaderTusResumable, TusVersion)
	c.Set(HeaderUploadOffset, strconv.FormatInt(offset, 10))
	c.Set(HeaderUploadLength, strconv.FormatInt(length, 10))

	return c.SendStatus(fiber.StatusOK)
}

func SendTusDeleteResponse(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func SendTusSlotResponse(c *fiber.Ctx, available bool, message string, queueLength int, activeCount int, maxConcurrent int) error {
	response := map[string]interface{}{
		"status":  "success",
		"message": "Pengecekan slot upload berhasil",
		"code":    fiber.StatusOK,
		"data": map[string]interface{}{
			"available":      available,
			"message":        message,
			"queue_length":   queueLength,
			"active_count":   activeCount,
			"active_upload":  activeCount > 0,
			"max_concurrent": maxConcurrent,
		},
		"timestamp": time.Now(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// SendTusErrorResponse sends a TUS protocol error response with proper headers.
// Sets Tus-Resumable header and returns the specified HTTP status code.
//
// Parameters:
//   - c: Fiber context
//   - statusCode: HTTP status code to return (e.g., 400, 404, 403, 409, 413, 429, 500)
//   - tusVersion: TUS protocol version (e.g., "1.0.0")
//
// Usage:
//
//	return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
func SendTusErrorResponse(c *fiber.Ctx, statusCode int, tusVersion string) error {
	c.Set(HeaderTusResumable, tusVersion)
	return c.SendStatus(statusCode)
}

// SendTusErrorResponseWithOffset sends a TUS error response with Upload-Offset header.
// Use when the upload has progressed but an error occurred.
//
// Parameters:
//   - c: Fiber context
//   - statusCode: HTTP status code to return (typically 409 Conflict for offset mismatch)
//   - tusVersion: TUS protocol version
//   - offset: Current upload offset to include in response
//
// Usage:
//
//	return upload.SendTusErrorResponseWithOffset(c, fiber.StatusConflict, ctrl.config.Upload.TusVersion, newOffset)
func SendTusErrorResponseWithOffset(c *fiber.Ctx, statusCode int, tusVersion string, offset int64) error {
	c.Set(HeaderTusResumable, tusVersion)
	c.Set(HeaderUploadOffset, strconv.FormatInt(offset, 10))
	return c.SendStatus(statusCode)
}

// SendTusErrorResponseWithLength sends a TUS error response with Upload-Length header.
// Use when file size validation fails.
//
// Parameters:
//   - c: Fiber context
//   - statusCode: HTTP status code to return (typically 413 Payload Too Large)
//   - tusVersion: TUS protocol version
//   - length: Upload length that was being processed
//
// Usage:
//
//	return upload.SendTusErrorResponseWithLength(c, fiber.StatusRequestEntityTooLarge, ctrl.config.Upload.TusVersion, fileSize)
func SendTusErrorResponseWithLength(c *fiber.Ctx, statusCode int, tusVersion string, length int64) error {
	c.Set(HeaderTusResumable, tusVersion)
	c.Set(HeaderUploadLength, strconv.FormatInt(length, 10))
	return c.SendStatus(statusCode)
}

// SendTusValidationErrorResponse sends a TUS validation error with JSON response.
// Use for REST endpoints that return JSON instead of protocol-only responses.
//
// Parameters:
//   - c: Fiber context
//   - message: Validation error message in Indonesian
//
// Usage:
//
//	if uploadMetadata == "" {
//	    return upload.SendTusValidationErrorResponse(c, "Metadata upload wajib diisi")
//	}
func SendTusValidationErrorResponse(c *fiber.Ctx, message string) error {
	return httputil.SendBadRequestResponse(c, message)
}

// SendTusNotFoundErrorResponse sends a TUS not found error with JSON response.
// Use when upload ID is not found.
//
// Parameters:
//   - c: Fiber context
//   - message: Error message (uses default if empty)
//
// Usage:
//
//	if err != nil && errors.Is(err, ErrUploadNotFound) {
//	    return upload.SendTusNotFoundErrorResponse(c, "Upload tidak ditemukan")
//	}
func SendTusNotFoundErrorResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Data tidak ditemukan"
	}
	return httputil.SendNotFoundResponse(c, message)
}

// SendTusForbiddenErrorResponse sends a TUS forbidden error with JSON response.
// Use when user lacks access to the upload.
//
// Parameters:
//   - c: Fiber context
//
// Usage:
//
//	if err != nil && errors.Is(err, ErrAccessDenied) {
//	    return upload.SendTusForbiddenErrorResponse(c)
//	}
func SendTusForbiddenErrorResponse(c *fiber.Ctx) error {
	return httputil.SendForbiddenResponse(c)
}

// SendTusConflictErrorResponse sends a TUS conflict error with JSON response.
// Use when upload is in an invalid state (e.g., already completed, locked).
//
// Parameters:
//   - c: Fiber context
//   - message: Error message in Indonesian
//
// Usage:
//
//	if err != nil && errors.Is(err, ErrUploadAlreadyCompleted) {
//	    return upload.SendTusConflictErrorResponse(c, "Upload sudah selesai")
//	}
func SendTusConflictErrorResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Data sudah ada"
	}
	return httputil.SendConflictResponse(c, message)
}

// SendTusPayloadTooLargeErrorResponse sends a TUS payload too large error.
// Use when file size exceeds maximum allowed.
//
// Parameters:
//   - c: Fiber context
//   - message: Error message (uses default if empty)
//
// Usage:
//
//	if fileSize > ctrl.config.Upload.MaxFileSize {
//	    return upload.SendTusPayloadTooLargeErrorResponse(c, "Ukuran file melebihi batas maksimal")
//	}
func SendTusPayloadTooLargeErrorResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Ukuran data melebihi batas maksimal"
	}
	return httputil.SendPayloadTooLargeResponse(c, message)
}

// SendTusTooManyRequestsErrorResponse sends a TUS too many requests error.
// Use when upload queue is full.
//
// Parameters:
//   - c: Fiber context
//   - message: Error message (uses default if empty)
//
// Usage:
//
//	if err != nil && errors.Is(err, ErrUploadQueueFull) {
//	    return upload.SendTusTooManyRequestsErrorResponse(c, "Antrian upload penuh")
//	}
func SendTusTooManyRequestsErrorResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Terlalu banyak permintaan, silakan coba lagi nanti"
	}
	return httputil.SendTooManyRequestsResponse(c, message)
}

// SendTusInternalErrorResponse sends a TUS internal server error with JSON response.
// Use for unexpected errors during upload processing.
//
// Parameters:
//   - c: Fiber context
//
// Usage:
//
//	if err != nil {
//	    return upload.SendTusInternalErrorResponse(c)
//	}
func SendTusInternalErrorResponse(c *fiber.Ctx) error {
	return httputil.SendInternalServerErrorResponse(c)
}
