package http

import (
	"bytes"
	"fiber-boiler-plate/config"
	base "fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// TusController handles TUS (Resumable Upload Protocol) v1.0.0 uploads.
// See tus_controller_doc.go for protocol documentation and integration examples.
type TusController struct {
	base       *base.BaseController
	tusUsecase usecase.TusUploadUsecase
	config     *config.Config
	validator  *validator.Validate
}

// NewTusController creates a new TUS controller instance.
func NewTusController(tusUsecase usecase.TusUploadUsecase, cfg *config.Config) *TusController {
	return &TusController{
		base:       base.NewBaseController("", nil),
		tusUsecase: tusUsecase,
		config:     cfg,
		validator:  validator.New(),
	}
}

// ========================================================================
// Helper Endpoints (Standard REST responses)
// ========================================================================

// CheckUploadSlot checks if an upload slot is available for the user.
// This endpoint uses standard REST response format.
//
// @Summary Check upload slot availability
// @Description Check if an upload slot is available for the current user
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.SuccessResponse "Slot availability info"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/upload/slot/check [get]
func (ctrl *TusController) CheckUploadSlot(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	result, err := ctrl.tusUsecase.CheckUploadSlot(userID)
	if err != nil {
		return ctrl.base.SendInternalError(c)
	}

	return helper.SendTusSlotResponse(
		c,
		result.Available,
		result.Message,
		result.QueueLength,
		result.ActiveUpload,
		result.MaxConcurrent,
	)
}

// ResetUploadQueue resets the user's upload queue.
// This endpoint uses standard REST response format.
//
// @Summary Reset upload queue
// @Description Reset the user's upload queue
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.SuccessResponse "Queue reset successfully"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/upload/queue/reset [post]
func (ctrl *TusController) ResetUploadQueue(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	err := ctrl.tusUsecase.ResetUploadQueue(userID)
	if err != nil {
		return ctrl.base.SendInternalError(c)
	}

	return ctrl.base.SendSuccess(c, nil, "Queue upload berhasil direset")
}

// ========================================================================
// TUS Protocol Endpoints - Project Upload
// ========================================================================

// InitiateUpload initiates a new TUS upload session for a project.
//
// @Summary Initiate TUS upload
// @Description Initiate a new resumable upload session using TUS protocol
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Length header int true "Total file size in bytes"
// @Param Upload-Metadata header string true "Base64 encoded metadata (nama_project, kategori, semester)"
// @Success 201 {object} domain.SuccessResponse "Upload initiated successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid headers or metadata"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 413 {object} domain.ErrorResponse "File size exceeds limit"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/upload [post]
func (ctrl *TusController) InitiateUpload(c *fiber.Ctx) error {
	// Get authenticated user context
	userID, userEmail, userRole, err := ctrl.getTusAuthContext(c)
	if err != nil {
		return err
	}

	// Validate TUS protocol headers
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion == "" {
		return ctrl.base.SendBadRequest(c, "Header Tus-Resumable wajib diisi")
	}

	if tusVersion != ctrl.config.Upload.TusVersion {
		appErr := apperrors.NewTusVersionError(ctrl.config.Upload.TusVersion)
		return helper.SendAppError(c, appErr)
	}

	// Parse and validate upload length
	fileSize, err := ctrl.parseUploadLength(c)
	if err != nil {
		return ctrl.base.SendBadRequest(c, "Header Upload-Length tidak valid")
	}

	// Parse and validate metadata
	metadata, err := ctrl.parseUploadMetadata(c.Get(helper.HeaderUploadMetadata))
	if err != nil {
		return ctrl.base.SendBadRequest(c, "Format Upload-Metadata tidak valid")
	}

	if !ctrl.base.ValidateStruct(c, metadata) {
		return nil
	}

	// Initiate upload
	result, err := ctrl.tusUsecase.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	// Set TUS response headers
	helper.SetTusResponseHeaders(c, 0, fileSize)
	helper.SetTusLocationHeader(c, result.UploadURL)

	return helper.SendTusInitiateResponse(c, result.UploadID, result.UploadURL, fileSize)
}

// UploadChunk uploads a chunk of data for an existing upload session.
//
// @Summary Upload chunk (TUS PATCH)
// @Description Upload a chunk of data for an existing upload session
// @Tags Upload
// @Accept application/offset+octet-stream
// @Security BearerAuth
// @Param id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Offset header int true "Current upload offset"
// @Param Content-Length header int true "Chunk size in bytes"
// @Success 204 "Chunk uploaded successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 409 {object} domain.ErrorResponse "Offset mismatch"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/upload/{id} [patch]
func (ctrl *TusController) UploadChunk(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		helper.SetTusResponseHeaders(c, 0, 0)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Validate upload ID
	uploadID := c.Params("id")
	if uploadID == "" {
		return helper.BuildTusErrorResponse(c, fiber.StatusBadRequest, 0)
	}

	// Validate TUS protocol headers
	if err := ctrl.validateTusHeaders(c); err != nil {
		return helper.BuildTusErrorResponse(c, fiber.StatusPreconditionFailed, 0)
	}

	// Parse and validate chunk data
	offset, _, bodyReader, err := ctrl.parseChunkRequest(c)
	if err != nil {
		return helper.BuildTusErrorResponse(c, fiber.StatusBadRequest, 0)
	}

	// Handle chunk upload
	newOffset, err := ctrl.tusUsecase.HandleChunk(uploadID, userID, offset, bodyReader)
	if err != nil {
		return ctrl.handleTusChunkError(c, err, newOffset)
	}

	// Return success with new offset
	return helper.SendTusChunkResponse(c, newOffset)
}

// GetUploadStatus retrieves the current upload status (TUS HEAD request).
//
// @Summary Get upload status (TUS HEAD)
// @Description Retrieve current upload offset and length
// @Tags Upload
// @Security BearerAuth
// @Param id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 200 "Upload status with Upload-Offset and Upload-Length headers"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/upload/{id} [head]
func (ctrl *TusController) GetUploadStatus(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	// Validate upload ID
	uploadID := c.Params("id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	// Validate TUS version
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion != ctrl.config.Upload.TusVersion {
		appErr := apperrors.NewTusVersionError(ctrl.config.Upload.TusVersion)
		return helper.SendAppError(c, appErr)
	}

	// Get upload status
	offset, length, err := ctrl.tusUsecase.GetUploadStatus(uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	// Return TUS HEAD response
	return helper.SendTusHeadResponse(c, offset, length)
}

// GetUploadInfo retrieves detailed information about an upload (REST endpoint).
//
// @Summary Get upload info
// @Description Retrieve detailed upload information including progress percentage
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upload ID"
// @Success 200 {object} domain.SuccessResponse "Upload info retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid upload ID"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/upload/{id}/info [get]
func (ctrl *TusController) GetUploadInfo(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	// Validate upload ID
	uploadID := c.Params("id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	// Get upload info
	result, err := ctrl.tusUsecase.GetUploadInfo(uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	return ctrl.base.SendSuccess(c, result, "Informasi upload berhasil didapat")
}

// CancelUpload cancels an active upload session.
//
// @Summary Cancel upload (TUS DELETE)
// @Description Cancel and clean up an active upload session
// @Tags Upload
// @Security BearerAuth
// @Param id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 204 "Upload cancelled successfully"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/upload/{id} [delete]
func (ctrl *TusController) CancelUpload(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	// Validate upload ID
	uploadID := c.Params("id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	// Validate TUS version
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion != ctrl.config.Upload.TusVersion {
		appErr := apperrors.NewTusVersionError(ctrl.config.Upload.TusVersion)
		return helper.SendAppError(c, appErr)
	}

	// Cancel upload
	err := ctrl.tusUsecase.CancelUpload(uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	return helper.SendTusDeleteResponse(c)
}

// ========================================================================
// TUS Protocol Endpoints - Project Update Upload
// ========================================================================

// InitiateProjectUpdateUpload initiates a new TUS upload session for project update.
//
// @Summary Initiate project update upload
// @Description Initiate a new resumable upload session for updating a project
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Length header int true "Total file size in bytes"
// @Param Upload-Metadata header string false "Base64 encoded metadata"
// @Success 201 {object} domain.SuccessResponse "Upload initiated successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid headers or project ID"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - no access to project"
// @Failure 404 {object} domain.ErrorResponse "Project not found"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 413 {object} domain.ErrorResponse "File size exceeds limit"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/project/{id}/upload [post]
func (ctrl *TusController) InitiateProjectUpdateUpload(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	// Parse project ID
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}

	// Validate TUS protocol headers
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion == "" {
		return ctrl.base.SendBadRequest(c, "Header Tus-Resumable wajib diisi")
	}

	if tusVersion != ctrl.config.Upload.TusVersion {
		appErr := apperrors.NewTusVersionError(ctrl.config.Upload.TusVersion)
		return helper.SendAppError(c, appErr)
	}

	// Parse and validate upload length
	fileSize, err := ctrl.parseUploadLength(c)
	if err != nil {
		return ctrl.base.SendBadRequest(c, "Header Upload-Length tidak valid")
	}

	// Parse and validate metadata (optional for project updates)
	uploadMetadata := c.Get(helper.HeaderUploadMetadata)
	var metadata domain.TusUploadInitRequest
	if uploadMetadata != "" {
		metadata, err = ctrl.parseUploadMetadata(uploadMetadata)
		if err != nil {
			return ctrl.base.SendBadRequest(c, "Format Upload-Metadata tidak valid")
		}
		if !ctrl.base.ValidateStruct(c, metadata) {
			return nil
		}
	}

	// Initiate upload
	result, err := ctrl.tusUsecase.InitiateProjectUpdateUpload(projectID, userID, fileSize, metadata)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	// Set TUS response headers
	helper.SetTusResponseHeaders(c, 0, fileSize)
	helper.SetTusLocationHeader(c, result.UploadURL)

	return helper.SendTusInitiateResponse(c, result.UploadID, result.UploadURL, fileSize)
}

// UploadProjectUpdateChunk uploads a chunk for project update.
//
// @Summary Upload project update chunk (TUS PATCH)
// @Description Upload a chunk of data for a project update upload session
// @Tags Upload
// @Accept application/offset+octet-stream
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Offset header int true "Current upload offset"
// @Param Content-Length header int true "Chunk size in bytes"
// @Success 204 "Chunk uploaded successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 409 {object} domain.ErrorResponse "Offset mismatch"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/project/{id}/upload/{upload_id} [patch]
func (ctrl *TusController) UploadProjectUpdateChunk(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		helper.SetTusResponseHeaders(c, 0, 0)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Parse project ID
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		helper.SetTusResponseHeaders(c, 0, 0)
		return err
	}

	// Validate upload ID
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.BuildTusErrorResponse(c, fiber.StatusBadRequest, 0)
	}

	// Validate TUS protocol headers
	if err := ctrl.validateTusHeaders(c); err != nil {
		return helper.BuildTusErrorResponse(c, fiber.StatusPreconditionFailed, 0)
	}

	// Parse and validate chunk data
	offset, _, bodyReader, err := ctrl.parseChunkRequest(c)
	if err != nil {
		return helper.BuildTusErrorResponse(c, fiber.StatusBadRequest, 0)
	}

	// Handle chunk upload
	newOffset, err := ctrl.tusUsecase.HandleProjectUpdateChunk(projectID, uploadID, userID, offset, bodyReader)
	if err != nil {
		return ctrl.handleTusChunkError(c, err, newOffset)
	}

	// Return success with new offset
	return helper.SendTusChunkResponse(c, newOffset)
}

// GetProjectUpdateUploadStatus retrieves project update upload status (TUS HEAD).
//
// @Summary Get project update upload status (TUS HEAD)
// @Description Retrieve current upload offset and length for project update
// @Tags Upload
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 200 "Upload status with Upload-Offset and Upload-Length headers"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/project/{id}/upload/{upload_id} [head]
func (ctrl *TusController) GetProjectUpdateUploadStatus(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	// Parse project ID
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}

	// Validate upload ID
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	// Validate TUS version
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion != ctrl.config.Upload.TusVersion {
		appErr := apperrors.NewTusVersionError(ctrl.config.Upload.TusVersion)
		return helper.SendAppError(c, appErr)
	}

	// Get upload status
	offset, length, err := ctrl.tusUsecase.GetProjectUpdateUploadStatus(projectID, uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	// Return TUS HEAD response
	return helper.SendTusHeadResponse(c, offset, length)
}

// GetProjectUpdateUploadInfo retrieves project update upload info (REST endpoint).
//
// @Summary Get project update upload info
// @Description Retrieve detailed upload information for project update
// @Tags Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} domain.SuccessResponse "Upload info retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid upload ID"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/project/{id}/upload/{upload_id}/info [get]
func (ctrl *TusController) GetProjectUpdateUploadInfo(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	// Parse project ID
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}

	// Validate upload ID
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	// Get upload info
	result, err := ctrl.tusUsecase.GetProjectUpdateUploadInfo(projectID, uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	return ctrl.base.SendSuccess(c, result, "Informasi update upload berhasil didapat")
}

// CancelProjectUpdateUpload cancels project update upload.
//
// @Summary Cancel project update upload (TUS DELETE)
// @Description Cancel and clean up an active project update upload session
// @Tags Upload
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 204 "Upload cancelled successfully"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Upload not found"
// @Failure 412 {object} domain.ErrorResponse "Unsupported TUS version"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/project/{id}/upload/{upload_id} [delete]
func (ctrl *TusController) CancelProjectUpdateUpload(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	// Parse project ID
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}

	// Validate upload ID
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	// Validate TUS version
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion != ctrl.config.Upload.TusVersion {
		appErr := apperrors.NewTusVersionError(ctrl.config.Upload.TusVersion)
		return helper.SendAppError(c, appErr)
	}

	// Cancel upload
	err = ctrl.tusUsecase.CancelProjectUpdateUpload(projectID, uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err, 0)
	}

	return helper.SendTusDeleteResponse(c)
}

// ========================================================================
// Helper Methods
// ========================================================================

// getTusAuthContext extracts authenticated user context (ID, email, role).
// Returns error and sends response if authentication fails.
func (ctrl *TusController) getTusAuthContext(c *fiber.Ctx) (string, string, string, error) {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	userEmail := ctrl.base.GetAuthenticatedUserEmail(c)
	if userEmail == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	userRole := ctrl.base.GetAuthenticatedUserRole(c)
	if userRole == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	return userID, userEmail, userRole, nil
}

// validateTusHeaders validates TUS protocol headers.
// Returns error if headers are invalid.
func (ctrl *TusController) validateTusHeaders(c *fiber.Ctx) error {
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion == "" || tusVersion != ctrl.config.Upload.TusVersion {
		return fiber.NewError(fiber.StatusPreconditionFailed, "invalid TUS version")
	}

	contentType := c.Get(helper.HeaderContentType)
	if contentType != helper.TusContentType {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "invalid content type")
	}

	return nil
}

// parseUploadLength parses and validates the Upload-Length header.
func (ctrl *TusController) parseUploadLength(c *fiber.Ctx) (int64, error) {
	uploadLengthStr := c.Get(helper.HeaderUploadLength)
	if uploadLengthStr == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Upload-Length header required")
	}

	fileSize, err := strconv.ParseInt(uploadLengthStr, 10, 64)
	if err != nil || fileSize <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid Upload-Length")
	}

	return fileSize, nil
}

// parseUploadMetadata parses TUS Upload-Metadata header.
// Format: "key1 base64value1,key2 base64value2"
func (ctrl *TusController) parseUploadMetadata(metadataHeader string) (domain.TusUploadInitRequest, error) {
	var metadata domain.TusUploadInitRequest

	if metadataHeader == "" {
		return metadata, fiber.NewError(fiber.StatusBadRequest, "Upload-Metadata header required")
	}

	metadataMap := helper.ParseTusMetadata(metadataHeader)

	if namaProject, ok := metadataMap["nama_project"]; ok {
		metadata.NamaProject = namaProject
	}

	if kategori, ok := metadataMap["kategori"]; ok {
		metadata.Kategori = kategori
	} else {
		metadata.Kategori = "website"
	}

	if semesterStr, ok := metadataMap["semester"]; ok {
		semester, err := strconv.Atoi(semesterStr)
		if err == nil {
			metadata.Semester = semester
		}
	}

	return metadata, nil
}

// parseChunkRequest parses and validates chunk upload request.
// Returns offset, chunk size, and body reader.
func (ctrl *TusController) parseChunkRequest(c *fiber.Ctx) (offset int64, chunkSize int64, bodyReader *bytes.Reader, err error) {
	// Parse offset
	offsetStr := c.Get(helper.HeaderUploadOffset)
	if offsetStr == "" {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Upload-Offset header required")
	}

	offset, err = strconv.ParseInt(offsetStr, 10, 64)
	if err != nil || offset < 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "invalid Upload-Offset")
	}

	// Parse chunk size
	contentLengthStr := c.Get(helper.HeaderContentLength)
	if contentLengthStr == "" {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Content-Length header required")
	}

	chunkSize, err = strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil || chunkSize <= 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "invalid Content-Length")
	}

	// Validate chunk size against maximum
	if chunkSize > int64(ctrl.config.Upload.ChunkSize*2) {
		return 0, 0, nil, fiber.NewError(fiber.StatusRequestEntityTooLarge, "chunk size exceeds maximum")
	}

	// Read and validate body
	bodyBytes := c.Body()
	if bodyBytes == nil || len(bodyBytes) == 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "request body is empty")
	}

	if int64(len(bodyBytes)) != chunkSize {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "body size doesn't match Content-Length")
	}

	bodyReader = bytes.NewReader(bodyBytes)
	return offset, chunkSize, bodyReader, nil
}

// handleTusUsecaseError handles usecase errors for TUS operations.
// Uses type-safe AppError checking to map to appropriate HTTP responses.
func (ctrl *TusController) handleTusUsecaseError(c *fiber.Ctx, err error, offset int64) error {
	if err == nil {
		return nil
	}

	// Handle typed AppError
	if appErr, ok := err.(*apperrors.AppError); ok {
		switch appErr.Code {
		case apperrors.ErrNotFound:
			return ctrl.base.SendNotFound(c, appErr.Message)
		case apperrors.ErrForbidden:
			return ctrl.base.SendForbidden(c)
		case apperrors.ErrPayloadTooLarge:
			return helper.SendAppError(c, appErr)
		case apperrors.ErrTusAlreadyCompleted:
			return helper.SendAppError(c, appErr)
		case apperrors.ErrTusVersionMismatch:
			return helper.SendAppError(c, appErr)
		default:
			// For any other AppError, use the AppError's status
			return helper.SendAppError(c, appErr)
		}
	}

	// Fallback for unexpected errors
	return ctrl.base.SendInternalError(c)
}

// handleTusChunkError handles errors during chunk upload.
// Uses type-safe AppError checking to return TUS protocol-compliant error responses.
func (ctrl *TusController) handleTusChunkError(c *fiber.Ctx, err error, offset int64) error {
	if err == nil {
		return nil
	}

	helper.SetTusResponseHeaders(c, 0, 0)

	// Handle typed AppError
	if appErr, ok := err.(*apperrors.AppError); ok {
		switch appErr.Code {
		case apperrors.ErrNotFound:
			return helper.BuildTusErrorResponse(c, fiber.StatusNotFound, 0)
		case apperrors.ErrForbidden:
			return helper.BuildTusErrorResponse(c, fiber.StatusForbidden, 0)
		case apperrors.ErrTusOffsetMismatch:
			helper.SetTusOffsetHeader(c, offset)
			return helper.BuildTusErrorResponse(c, fiber.StatusConflict, offset)
		case apperrors.ErrTusInactive:
			return helper.BuildTusErrorResponse(c, fiber.StatusLocked, 0)
		case apperrors.ErrTusAlreadyCompleted:
			return helper.BuildTusErrorResponse(c, fiber.StatusConflict, 0)
		default:
			return helper.BuildTusErrorResponse(c, appErr.HTTPStatus, 0)
		}
	}

	// Fallback for unexpected errors
	return helper.BuildTusErrorResponse(c, fiber.StatusInternalServerError, 0)
}
