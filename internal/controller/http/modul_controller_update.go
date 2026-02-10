package http

import (
	"bytes"
	"fiber-boiler-plate/internal/helper"
	apperrors "fiber-boiler-plate/internal/errors"

	"github.com/gofiber/fiber/v2"
)

// Note: These methods are extensions of ModulController for modul file updates.
// They share the same receiver type and use the embedded base controller.

// InitiateModulUpdateUpload handles POST /modul/:id/upload
// @Summary Initiate modul update upload (TUS POST)
// @Description Start a new TUS upload for updating an existing module file
// @Tags Modul Update
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID to update"
// @Header 200 {string} Tus-Resumable "TUS protocol version"
// @Header 201 {string} Location "Upload URL for subsequent PATCH requests"
// @Header 201 {string} Upload-Offset "Initial offset (0)"
// @Param Tus-Resumable header string true "TUS protocol version (1.0.0)"
// @Param Upload-Length header string true "Total file size in bytes"
// @Param Upload-Metadata header string true "Base64-encoded metadata (filename, etc.)"
// @Success 201 {object} domain.SuccessResponse{data=map[string]interface{}}
// @Failure 400 {string} string "TUS error response"
// @Failure 401 {string} string "TUS error response"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Module not found"
// @Failure 412 {string} string "TUS version mismatch"
// @Failure 413 {string} string "File too large"
// @Failure 429 {string} string "Upload queue full"
// @Failure 500 {string} string "Server error"
// @Router /api/v1/modul/{id}/upload [post]
func (ctrl *ModulController) InitiateModulUpdateUpload(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	// Parse module ID from path
	modulID, err := ctrl.ParsePathID(c)
	if err != nil {
		return helper.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}

	// Get TUS headers
	tusHeaders := helper.GetTusHeaders(c)

	// Validate TUS version
	if tusHeaders.TusResumable != ctrl.config.Upload.TusVersion {
		return helper.SendTusErrorResponse(c, fiber.StatusPreconditionFailed, ctrl.config.Upload.TusVersion)
	}

	// Validate Upload-Length
	if tusHeaders.UploadLength <= 0 {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Length wajib diisi")
	}

	// Validate Upload-Metadata
	if tusHeaders.UploadMetadata == "" {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Metadata wajib diisi")
	}

	// Call usecase
	result, err := ctrl.tusModulUsecase.InitiateModulUpdateUpload(modulID, userID, tusHeaders.UploadLength, tusHeaders.UploadMetadata)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.HTTPStatus {
			case fiber.StatusNotFound:
				return helper.SendTusErrorResponse(c, fiber.StatusNotFound, ctrl.config.Upload.TusVersion)
			case fiber.StatusForbidden:
				return helper.SendTusErrorResponse(c, fiber.StatusForbidden, ctrl.config.Upload.TusVersion)
			case fiber.StatusTooManyRequests:
				return helper.SendTusTooManyRequestsErrorResponse(c, appErr.Message)
			case fiber.StatusBadRequest:
				return helper.SendTusValidationErrorResponse(c, appErr.Message)
			case fiber.StatusRequestEntityTooLarge:
				return helper.SendTusPayloadTooLargeErrorResponse(c, appErr.Message)
			default:
				return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
			}
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	// Set TUS response headers
	helper.SetTusResponseHeaders(c, 0, tusHeaders.UploadLength)
	helper.SetTusLocationHeader(c, result.UploadURL)

	return helper.SendSuccessResponse(c, helper.StatusCreated, "Update upload modul berhasil diinisiasi", result)
}

// UploadModulUpdateChunk handles PATCH /modul/:id/update/:upload_id
// @Summary Upload update chunk (TUS PATCH)
// @Description Upload a chunk of file data for updating an existing module
// @Tags Modul Update
// @Accept application/offset+octet-stream
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID being updated"
// @Param upload_id path string true "Upload ID from InitiateModulUpdateUpload"
// @Param Tus-Resumable header string true "TUS protocol version (1.0.0)"
// @Param Upload-Offset header string true "Current offset in bytes"
// @Param Content-Type header string true "Must be application/offset+octet-stream"
// @Param Content-Length header string true "Chunk size in bytes"
// @Success 204 {string} string "No content (upload continuing)"
// @Header 204 {string} Tus-Resumable "TUS protocol version"
// @Header 204 {string} Upload-Offset "New offset after this chunk"
// @Failure 400 {string} string "Invalid chunk or offset"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Upload not found"
// @Failure 409 {string} string "Offset mismatch"
// @Failure 412 {string} string "TUS version mismatch"
// @Failure 413 {string} string "Chunk too large"
// @Failure 415 {string} string "Invalid content type"
// @Failure 500 {string} string "Server error"
// @Router /api/v1/modul/{id}/update/{upload_id} [patch]
func (ctrl *ModulController) UploadModulUpdateChunk(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	// Get upload ID from path
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	// Get TUS headers
	tusHeaders := helper.GetTusHeaders(c)

	// Validate TUS version
	if tusHeaders.TusResumable != ctrl.config.Upload.TusVersion {
		return helper.SendTusErrorResponse(c, fiber.StatusPreconditionFailed, ctrl.config.Upload.TusVersion)
	}

	// Validate Content-Type
	if tusHeaders.ContentType != helper.TusContentType {
		return helper.SendTusErrorResponse(c, fiber.StatusUnsupportedMediaType, ctrl.config.Upload.TusVersion)
	}

	// Validate offset
	if tusHeaders.UploadOffset < 0 {
		return helper.SendTusValidationErrorResponse(c, "Upload-Offset tidak valid")
	}

	// Validate chunk size
	if tusHeaders.ContentLength <= 0 {
		return helper.SendTusValidationErrorResponse(c, "Content-Length tidak valid")
	}

	// Validate chunk size limit
	if err := helper.ValidateChunkSize(tusHeaders.ContentLength); err != nil {
		return helper.SendTusErrorResponseWithLength(c, fiber.StatusRequestEntityTooLarge, ctrl.config.Upload.TusVersion, tusHeaders.ContentLength)
	}

	// Get request body
	bodyBytes := c.Body()
	if bodyBytes == nil || len(bodyBytes) == 0 {
		return helper.SendTusValidationErrorResponse(c, "Request body kosong")
	}

	if int64(len(bodyBytes)) != tusHeaders.ContentLength {
		return helper.SendTusValidationErrorResponse(c, "Ukuran chunk tidak sesuai dengan Content-Length")
	}

	bodyReader := bytes.NewReader(bodyBytes)

	// Call usecase
	newOffset, err := ctrl.tusModulUsecase.HandleModulUpdateChunk(uploadID, userID, tusHeaders.UploadOffset, bodyReader)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.HTTPStatus {
			case fiber.StatusNotFound:
				return helper.SendTusErrorResponse(c, fiber.StatusNotFound, ctrl.config.Upload.TusVersion)
			case fiber.StatusForbidden:
				return helper.SendTusErrorResponse(c, fiber.StatusForbidden, ctrl.config.Upload.TusVersion)
			case fiber.StatusConflict:
				return helper.SendTusErrorResponseWithOffset(c, fiber.StatusConflict, ctrl.config.Upload.TusVersion, newOffset)
			case fiber.StatusLocked:
				return helper.SendTusErrorResponse(c, fiber.StatusLocked, ctrl.config.Upload.TusVersion)
			default:
				return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
			}
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	// Send TUS chunk response
	return helper.SendTusChunkResponse(c, newOffset)
}

// GetModulUpdateUploadStatus handles HEAD /modul/:id/update/:upload_id
// @Summary Get update upload status (TUS HEAD)
// @Description Check the current progress of a modul update upload
// @Tags Modul Update
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID being updated"
// @Param upload_id path string true "Upload ID"
// @Success 200 {string} string "TUS headers with current offset and length"
// @Header 200 {string} Tus-Resumable "TUS protocol version"
// @Header 200 {string} Upload-Offset "Current offset in bytes"
// @Header 200 {string} Upload-Length "Total file length"
// @Failure 400 {string} string "Invalid upload ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Upload not found"
// @Failure 500 {string} string "Server error"
// @Router /api/v1/modul/{id}/update/{upload_id} [head]
func (ctrl *ModulController) GetModulUpdateUploadStatus(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	// Get upload ID from path
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	// Call usecase
	offset, length, err := ctrl.tusModulUsecase.GetModulUploadStatus(uploadID, userID)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.HTTPStatus {
			case fiber.StatusNotFound:
				return helper.SendTusErrorResponse(c, fiber.StatusNotFound, ctrl.config.Upload.TusVersion)
			case fiber.StatusForbidden:
				return helper.SendTusErrorResponse(c, fiber.StatusForbidden, ctrl.config.Upload.TusVersion)
			default:
				return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
			}
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	// Send TUS HEAD response
	return helper.SendTusHeadResponse(c, offset, length)
}

// GetModulUpdateUploadInfo handles GET /modul/:id/update/:upload_id
// @Summary Get update upload information
// @Description Get detailed update upload info as JSON (not part of TUS protocol)
// @Tags Modul Update
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID being updated"
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} domain.SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/modul/{id}/update/{upload_id} [get]
func (ctrl *ModulController) GetModulUpdateUploadInfo(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Get upload ID from path
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return ctrl.SendBadRequest(c, "Upload ID tidak valid")
	}

	// Call usecase
	info, err := ctrl.tusModulUsecase.GetModulUploadInfo(uploadID, userID)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, info, "Informasi update upload berhasil didapat")
}

// CancelModulUpdateUpload handles DELETE /modul/:id/update/:upload_id
// @Summary Cancel update upload (TUS DELETE)
// @Description Cancel an ongoing modul update upload and clean up resources
// @Tags Modul Update
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID being updated"
// @Param upload_id path string true "Upload ID"
// @Success 204 {string} string "No content (upload cancelled)"
// @Header 204 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {string} string "Invalid upload ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Upload not found"
// @Failure 409 {string} string "Upload already completed"
// @Failure 500 {string} string "Server error"
// @Router /api/v1/modul/{id}/update/{upload_id} [delete]
func (ctrl *ModulController) CancelModulUpdateUpload(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	// Get upload ID from path
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	// Call usecase
	err := ctrl.tusModulUsecase.CancelModulUpload(uploadID, userID)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			switch appErr.HTTPStatus {
			case fiber.StatusNotFound:
				return helper.SendTusErrorResponse(c, fiber.StatusNotFound, ctrl.config.Upload.TusVersion)
			case fiber.StatusForbidden:
				return helper.SendTusErrorResponse(c, fiber.StatusForbidden, ctrl.config.Upload.TusVersion)
			case fiber.StatusConflict:
				return helper.SendTusErrorResponse(c, fiber.StatusConflict, ctrl.config.Upload.TusVersion)
			default:
				return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
			}
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	// Send TUS delete response
	return helper.SendTusDeleteResponse(c)
}
