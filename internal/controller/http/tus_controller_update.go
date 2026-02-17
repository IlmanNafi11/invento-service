package http

import (
	"invento-service/internal/upload"

	"github.com/gofiber/fiber/v2"
)

// InitiateProjectUpdateUpload handles POST /api/v1/project/{id}/upload - Initiate project update upload
// @Summary Initiate project update upload
// @Description Start a new TUS resumable upload to update an existing project file
// @Tags TUS Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Length header int true "Total file size in bytes"
// @Param Upload-Metadata header string false "Upload metadata (nama_project, kategori, semester)"
// @Success 201 {object} dto.SuccessResponse{data=dto.TusUploadResponse} "Update upload initiated"
// @Header 201 {string} Location "Upload URL"
// @Header 201 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Project not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/{id}/upload [post]
func (ctrl *TusController) InitiateProjectUpdateUpload(c *fiber.Ctx) error {
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}

	return ctrl.initiateUpload(c, &projectID)
}

// UploadProjectUpdateChunk handles PATCH /api/v1/project/{id}/update/{upload_id} - Upload project update chunk
// @Summary Upload project update file chunk
// @Description Upload a chunk of data for a TUS resumable project update upload
// @Tags TUS Upload
// @Accept application/offset+octet-stream
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Offset header int true "Byte offset for this chunk"
// @Param Content-Type header string true "Content type" default(application/offset+octet-stream)
// @Success 204 "Chunk uploaded"
// @Header 204 {string} Upload-Offset "New byte offset after chunk"
// @Header 204 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 409 {object} dto.ErrorResponse "Offset mismatch"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/{id}/update/{upload_id} [patch]
func (ctrl *TusController) UploadProjectUpdateChunk(c *fiber.Ctx) error {
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.uploadChunk(c, &projectID)
}

// GetProjectUpdateUploadStatus handles HEAD /api/v1/project/{id}/update/{upload_id} - Get project update upload status
// @Summary Get project update upload status
// @Description Get the current offset and length of a TUS project update upload via HEAD request
// @Tags TUS Upload
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 200 "Upload status"
// @Header 200 {string} Upload-Offset "Current byte offset"
// @Header 200 {string} Upload-Length "Total file size"
// @Header 200 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/{id}/update/{upload_id} [head]
func (ctrl *TusController) GetProjectUpdateUploadStatus(c *fiber.Ctx) error {
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}
	return ctrl.getUploadStatus(c, &projectID)
}

// GetProjectUpdateUploadInfo handles GET /api/v1/project/{id}/update/{upload_id} - Get project update upload info
// @Summary Get project update upload info
// @Description Get detailed information about a TUS project update upload
// @Tags TUS Upload
// @Produce json
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.TusUploadInfoResponse} "Upload info"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/{id}/update/{upload_id} [get]
func (ctrl *TusController) GetProjectUpdateUploadInfo(c *fiber.Ctx) error {
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}
	return ctrl.getUploadInfo(c, &projectID)
}

// CancelProjectUpdateUpload handles DELETE /api/v1/project/{id}/update/{upload_id} - Cancel project update upload
// @Summary Cancel project update upload
// @Description Cancel and clean up a TUS project update upload
// @Tags TUS Upload
// @Security BearerAuth
// @Param id path int true "Project ID"
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 204 "Upload cancelled"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/{id}/update/{upload_id} [delete]
func (ctrl *TusController) CancelProjectUpdateUpload(c *fiber.Ctx) error {
	projectID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return err
	}
	return ctrl.cancelUpload(c, &projectID)
}
