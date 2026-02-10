package http

import (
	"bytes"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ModulController handles module management endpoints.
// Embeds BaseController for common functionality.
type ModulController struct {
	*base.BaseController
	modulUsecase    usecase.ModulUsecase
	tusModulUsecase usecase.TusModulUsecase
	config          *config.Config
}

// NewModulController creates a new modul controller instance.
func NewModulController(
	modulUsecase usecase.ModulUsecase,
	tusModulUsecase usecase.TusModulUsecase,
	config *config.Config,
	baseCtrl *base.BaseController,
) *ModulController {
	return &ModulController{
		BaseController:  baseCtrl,
		modulUsecase:    modulUsecase,
		tusModulUsecase: tusModulUsecase,
		config:          config,
	}
}

// GetList retrieves a paginated list of modules for the authenticated user.
// @Summary Get list of modules
// @Description Retrieve paginated list of modules with optional filtering
// @Tags Modul
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search keyword"
// @Param filter_type query string false "Filter by file type"
// @Param filter_semester query int false "Filter by semester (1-8)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} domain.SuccessResponse{data=domain.ModulListData}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/modul [get]
func (ctrl *ModulController) GetList(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse query parameters
	var params domain.ModulListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	// Set defaults for pagination
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Call usecase
	result, err := ctrl.modulUsecase.GetList(userID, params.Search, params.FilterType, params.FilterSemester, params.Page, params.Limit)
	if err != nil {
		// Check if AppError
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Daftar modul berhasil diambil", result)
}

// Delete deletes a module by ID.
// @Summary Delete a module
// @Description Permanently delete a module owned by the authenticated user
// @Tags Modul
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/modul/{id} [delete]
func (ctrl *ModulController) Delete(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse path parameter using base controller
	modulID, err := ctrl.ParsePathID(c)
	if err != nil {
		return nil // Error response already sent
	}

	// Call usecase
	err = ctrl.modulUsecase.Delete(modulID, userID)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Modul berhasil dihapus")
}

// CheckUploadSlot checks if an upload slot is available for the user.
// @Summary Check upload slot availability
// @Description Check if the user has an available upload slot (queue status)
// @Tags Modul
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.SuccessResponse{data=map[string]interface{}}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/modul/upload/slot [get]
func (ctrl *ModulController) CheckUploadSlot(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Call usecase
	response, err := ctrl.tusModulUsecase.CheckModulUploadSlot(userID)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Status slot upload berhasil didapat", response)
}

// InitiateUpload initiates a new TUS upload for a module file.
// @Summary Initiate TUS upload
// @Description Start a new resumable upload using TUS protocol
// @Tags Modul Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Header 200 {string} Tus-Resumable "TUS protocol version"
// @Header 201 {string} Location "Upload URL for subsequent PATCH requests"
// @Header 201 {string} Upload-Offset "Initial offset (0)"
// @Param Tus-Resumable header string true "TUS protocol version (1.0.0)"
// @Param Upload-Length header string true "Total file size in bytes"
// @Param Upload-Metadata header string true "Base64-encoded metadata (filename, etc.)"
// @Success 201 {object} domain.SuccessResponse{data=map[string]interface{}}
// @Failure 400 {string} string "TUS error response"
// @Failure 401 {string} string "TUS error response"
// @Failure 412 {string} string "TUS version mismatch"
// @Failure 413 {string} string "File too large"
// @Failure 429 {string} string "Upload queue full"
// @Failure 500 {string} string "Server error"
// @Router /api/v1/modul/upload [post]
func (ctrl *ModulController) InitiateUpload(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	// Validate TUS version
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion != ctrl.config.Upload.TusVersion {
		return helper.SendTusErrorResponse(c, fiber.StatusPreconditionFailed, ctrl.config.Upload.TusVersion)
	}

	// Parse Upload-Length header
	uploadLengthStr := c.Get(helper.HeaderUploadLength)
	if uploadLengthStr == "" {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Length wajib diisi")
	}

	fileSize, err := strconv.ParseInt(uploadLengthStr, 10, 64)
	if err != nil || fileSize <= 0 {
		return helper.SendTusValidationErrorResponse(c, "Ukuran file tidak valid")
	}

	// Parse Upload-Metadata header
	uploadMetadata := c.Get(helper.HeaderUploadMetadata)
	if uploadMetadata == "" {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Metadata wajib diisi")
	}

	// Call usecase
	result, err := ctrl.tusModulUsecase.InitiateModulUpload(userID, fileSize, uploadMetadata)
	if err != nil {
		// Handle typed errors with appropriate TUS responses
		if appErr, ok := err.(*apperrors.AppError); ok {
			// Map AppError HTTP status to TUS responses
			switch appErr.HTTPStatus {
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
	helper.SetTusResponseHeaders(c, 0, fileSize)
	helper.SetTusLocationHeader(c, result.UploadURL)

	return helper.SendSuccessResponse(c, helper.StatusCreated, "Upload modul berhasil diinisiasi", result)
}

// UploadChunk handles PATCH request to upload a chunk of file data.
// @Summary Upload chunk (TUS PATCH)
// @Description Upload a chunk of file data as part of a resumable upload
// @Tags Modul Upload
// @Accept application/offset+octet-stream
// @Produce json
// @Security BearerAuth
// @Param upload_id path string true "Upload ID from InitiateUpload"
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
// @Router /api/v1/modul/upload/{upload_id} [patch]
func (ctrl *ModulController) UploadChunk(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	// Get upload ID from path
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	// Get TUS headers using helper
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
	newOffset, err := ctrl.tusModulUsecase.HandleModulChunk(uploadID, userID, tusHeaders.UploadOffset, bodyReader)
	if err != nil {
		// Handle typed errors with appropriate TUS responses
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

	// Send TUS chunk response with new offset
	return helper.SendTusChunkResponse(c, newOffset)
}

// GetUploadStatus retrieves the current status of an upload (TUS HEAD).
// @Summary Get upload status (TUS HEAD)
// @Description Check the current progress of an upload using TUS HEAD request
// @Tags Modul Upload
// @Accept json
// @Produce json
// @Security BearerAuth
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
// @Router /api/v1/modul/upload/{upload_id} [head]
func (ctrl *ModulController) GetUploadStatus(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
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
		// Handle typed errors with appropriate TUS responses
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

// GetUploadInfo retrieves detailed information about an upload (REST helper).
// @Summary Get upload information
// @Description Get detailed upload info as JSON (not part of TUS protocol)
// @Tags Modul Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} domain.SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/modul/upload/{upload_id}/info [get]
func (ctrl *ModulController) GetUploadInfo(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
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

	return ctrl.SendSuccess(c, info, "Informasi upload berhasil didapat")
}

// CancelUpload cancels an ongoing upload (TUS DELETE).
// @Summary Cancel upload (TUS DELETE)
// @Description Cancel an ongoing upload and clean up resources
// @Tags Modul Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param upload_id path string true "Upload ID"
// @Success 204 {string} string "No content (upload cancelled)"
// @Header 204 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {string} string "Invalid upload ID"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 404 {string} string "Upload not found"
// @Failure 409 {string} string "Upload already completed"
// @Failure 500 {string} string "Server error"
// @Router /api/v1/modul/upload/{upload_id} [delete]
func (ctrl *ModulController) CancelUpload(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
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
		// Handle typed errors with appropriate TUS responses
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

// UpdateMetadata updates a module's metadata (name and semester).
// @Summary Update module metadata
// @Description Update module name and/or semester for a module owned by the user
// @Tags Modul
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID"
// @Param request body domain.ModulUpdateRequest true "Update request"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/modul/{id} [patch]
func (ctrl *ModulController) UpdateMetadata(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse path parameter using base controller
	modulID, err := ctrl.ParsePathID(c)
	if err != nil {
		return nil // Error response already sent
	}

	// Parse request body
	var req domain.ModulUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	// Validate using base controller
	if !ctrl.ValidateStruct(c, req) {
		return nil // Validation error response already sent
	}

	// Call usecase
	err = ctrl.modulUsecase.UpdateMetadata(modulID, userID, req)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Metadata modul berhasil diperbarui")
}

// Download downloads one or more modules as a ZIP file.
// @Summary Download modules
// @Description Download selected modules as a ZIP archive
// @Tags Modul
// @Accept json
// @Produce application/zip
// @Security BearerAuth
// @Param request body domain.ModulDownloadRequest true "Download request with module IDs"
// @Success 200 {file} binary "ZIP file containing modules"
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /api/v1/modul/download [post]
func (ctrl *ModulController) Download(c *fiber.Ctx) error {
	// Get authenticated user ID using base controller
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == 0 {
		return nil
	}

	// Parse request body
	var req domain.ModulDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	// Validate IDs
	if len(req.IDs) == 0 {
		return ctrl.SendBadRequest(c, "ID modul tidak boleh kosong")
	}

	// Validate using base controller
	if !ctrl.ValidateStruct(c, req) {
		return nil // Validation error response already sent
	}

	// Call usecase
	filePath, err := ctrl.modulUsecase.Download(userID, req.IDs)
	if err != nil {
		// Handle typed errors
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return c.Download(filePath)
}
