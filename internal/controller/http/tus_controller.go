package http

import (
	"bytes"
	"encoding/base64"
	"fiber-boiler-plate/config"
	base "fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"
	"strings"

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
		base:       base.NewBaseController(nil, nil),
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
// Response: JSON with slot availability information
func (ctrl *TusController) CheckUploadSlot(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// Response: JSON success message
func (ctrl *TusController) ResetUploadQueue(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// TUS Protocol: POST /upload
// Request Headers:
//   - Tus-Resumable: 1.0.0
//   - Upload-Length: [total file size]
//   - Upload-Metadata: [base64 encoded metadata]
//
// Response: 201 Created with Location header and JSON body
func (ctrl *TusController) InitiateUpload(c *fiber.Ctx) error {
	// Get authenticated user context
	userID, userEmail, userRole, err := ctrl.getTusAuthContext(c)
	if err != nil {
		return err
	}

	// Validate TUS protocol headers
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Header Tus-Resumable wajib diisi", nil)
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
// TUS Protocol: PATCH /upload/{id}
// Request Headers:
//   - Tus-Resumable: 1.0.0
//   - Upload-Offset: [current offset]
//   - Content-Type: application/offset+octet-stream
//   - Content-Length: [chunk size]
//
// Response: 204 No Content with Upload-Offset header
func (ctrl *TusController) UploadChunk(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// TUS Protocol: HEAD /upload/{id}
// Response: 200 OK with Upload-Offset and Upload-Length headers
func (ctrl *TusController) GetUploadStatus(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// This is NOT a TUS protocol endpoint - it returns standard JSON response.
// Use this for getting upload metadata, progress percentage, etc.
//
// Response: JSON with upload information
func (ctrl *TusController) GetUploadInfo(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// TUS Protocol: DELETE /upload/{id}
// Response: 204 No Content
func (ctrl *TusController) CancelUpload(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// TUS Protocol: POST /project/{id}/upload
// Response: 201 Created with Location header and JSON body
func (ctrl *TusController) InitiateProjectUpdateUpload(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Header Tus-Resumable wajib diisi", nil)
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
// TUS Protocol: PATCH /project/{id}/upload/{upload_id}
// Response: 204 No Content with Upload-Offset header
func (ctrl *TusController) UploadProjectUpdateChunk(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// TUS Protocol: HEAD /project/{id}/upload/{upload_id}
// Response: 200 OK with Upload-Offset and Upload-Length headers
func (ctrl *TusController) GetProjectUpdateUploadStatus(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// This is NOT a TUS protocol endpoint - it returns standard JSON response.
//
// Response: JSON with upload information
func (ctrl *TusController) GetProjectUpdateUploadInfo(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
// TUS Protocol: DELETE /project/{id}/upload/{upload_id}
// Response: 204 No Content
func (ctrl *TusController) CancelProjectUpdateUpload(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
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
func (ctrl *TusController) getTusAuthContext(c *fiber.Ctx) (uint, string, string, error) {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == 0 {
		return 0, "", "", fiber.ErrUnauthorized
	}

	userEmail := ctrl.base.GetAuthenticatedUserEmail(c)
	if userEmail == "" {
		return 0, "", "", fiber.ErrUnauthorized
	}

	userRole := ctrl.base.GetAuthenticatedUserRole(c)
	if userRole == "" {
		return 0, "", "", fiber.ErrUnauthorized
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

	pairs := strings.Split(metadataHeader, ",")
	metadataMap := make(map[string]string)

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), " ", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		valueB64 := parts[1]

		value, err := base64.StdEncoding.DecodeString(valueB64)
		if err != nil {
			return metadata, err
		}

		metadataMap[key] = string(value)
	}

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
// Converts string-based errors to appropriate HTTP responses.
func (ctrl *TusController) handleTusUsecaseError(c *fiber.Ctx, err error, offset int64) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "tidak ditemukan"):
		return ctrl.base.SendNotFound(c, errMsg)
	case strings.Contains(errMsg, "tidak memiliki akses"):
		return ctrl.base.SendForbidden(c)
	case strings.Contains(errMsg, "melebihi batas maksimal"):
		appErr := apperrors.NewPayloadTooLargeError(errMsg)
		return helper.SendAppError(c, appErr)
	case strings.Contains(errMsg, "sudah selesai"):
		appErr := apperrors.NewTusCompletedError()
		return helper.SendAppError(c, appErr)
	default:
		return ctrl.base.SendInternalError(c)
	}
}

// handleTusChunkError handles errors during chunk upload.
// Returns TUS protocol-compliant error responses with headers.
func (ctrl *TusController) handleTusChunkError(c *fiber.Ctx, err error, offset int64) error {
	if err == nil {
		return nil
	}

	helper.SetTusResponseHeaders(c, 0, 0)

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "tidak ditemukan"):
		return helper.BuildTusErrorResponse(c, fiber.StatusNotFound, 0)
	case strings.Contains(errMsg, "tidak memiliki akses"):
		return helper.BuildTusErrorResponse(c, fiber.StatusForbidden, 0)
	case strings.Contains(errMsg, "offset tidak valid"):
		helper.SetTusOffsetHeader(c, offset)
		return helper.BuildTusErrorResponse(c, fiber.StatusConflict, offset)
	case strings.Contains(errMsg, "tidak aktif"):
		return helper.BuildTusErrorResponse(c, fiber.StatusLocked, 0)
	default:
		return helper.BuildTusErrorResponse(c, fiber.StatusInternalServerError, 0)
	}
}
