package http

import (
	"invento-service/config"
	base "invento-service/internal/controller/base"
	"invento-service/internal/dto"
	"invento-service/internal/upload"
	"invento-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type TusModulController struct {
	base            *base.BaseController
	tusModulUsecase usecase.TusModulUsecase
	config          *config.Config
}

func NewTusModulController(tusModulUsecase usecase.TusModulUsecase, cfg *config.Config, baseCtrl *base.BaseController) *TusModulController {
	return &TusModulController{
		base:            baseCtrl,
		tusModulUsecase: tusModulUsecase,
		config:          cfg,
	}
}

// CheckUploadSlot handles GET /api/v1/modul/upload/check-slot - Check modul upload slot availability
// @Summary Check modul upload slot
// @Description Check if a modul upload slot is available for the authenticated user
// @Tags TUS Modul Upload
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.TusModulUploadSlotResponse} "Slot status"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/upload/check-slot [get]
func (ctrl *TusModulController) CheckUploadSlot(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	response, err := ctrl.tusModulUsecase.CheckModulUploadSlot(ctx, userID)
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return ctrl.base.SendSuccess(c, response, "Status slot upload berhasil didapat")
}

// InitiateUpload handles POST /api/v1/modul/upload/ - Initiate new modul upload
// @Summary Initiate modul upload
// @Description Start a new TUS resumable upload for a module file
// @Tags TUS Modul Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Length header int true "Total file size in bytes"
// @Param Upload-Metadata header string true "Upload metadata (judul, deskripsi)"
// @Success 201 {object} dto.SuccessResponse{data=dto.TusModulUploadResponse} "Upload initiated"
// @Header 201 {string} Location "Upload URL"
// @Header 201 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 409 {object} dto.ErrorResponse "No upload slot available"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/upload/ [post]
func (ctrl *TusModulController) InitiateUpload(c *fiber.Ctx) error {
	return ctrl.initiateUpload(c, nil)
}

// InitiateModulUpdateUpload handles POST /api/v1/modul/{id}/upload - Initiate modul update upload
// @Summary Initiate modul update upload
// @Description Start a new TUS resumable upload to update an existing module file
// @Tags TUS Modul Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Modul ID (UUID)"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Length header int true "Total file size in bytes"
// @Param Upload-Metadata header string true "Upload metadata (judul, deskripsi)"
// @Success 201 {object} dto.SuccessResponse{data=dto.TusModulUploadResponse} "Update upload initiated"
// @Header 201 {string} Location "Upload URL"
// @Header 201 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Modul not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/{id}/upload [post]
func (ctrl *TusModulController) InitiateModulUpdateUpload(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.initiateUpload(c, &modulID)
}

func (ctrl *TusModulController) initiateUpload(c *fiber.Ctx, modulID *string) error {
	ctx := c.UserContext()
	userID, _, _, err := getTusAuthContext(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	if err = validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	tusHeaders, err := upload.GetTusHeaders(c)
	if err != nil {
		return upload.SendTusValidationErrorResponse(c, err.Error())
	}
	if tusHeaders.UploadLength <= 0 {
		return upload.SendTusValidationErrorResponse(c, "Header Upload-Length wajib diisi")
	}
	if tusHeaders.UploadMetadata == "" {
		return upload.SendTusValidationErrorResponse(c, "Header Upload-Metadata wajib diisi")
	}

	var result *dto.TusModulUploadResponse
	if modulID == nil {
		result, err = ctrl.tusModulUsecase.InitiateModulUpload(ctx, userID, tusHeaders.UploadLength, tusHeaders.UploadMetadata)
	} else {
		result, err = ctrl.tusModulUsecase.InitiateModulUpdateUpload(ctx, *modulID, userID, tusHeaders.UploadLength, tusHeaders.UploadMetadata)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	upload.SetTusResponseHeaders(c, 0, tusHeaders.UploadLength)
	upload.SetTusLocationHeader(c, result.UploadURL)

	message := "Upload modul berhasil diinisiasi"
	if modulID != nil {
		message = "Update upload modul berhasil diinisiasi"
	}

	return ctrl.base.SendCreated(c, result, message)
}

// UploadChunk handles PATCH /api/v1/modul/upload/{upload_id} - Upload modul file chunk
// @Summary Upload modul file chunk
// @Description Upload a chunk of data for a TUS resumable modul upload
// @Tags TUS Modul Upload
// @Accept application/offset+octet-stream
// @Security BearerAuth
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
// @Router /modul/upload/{upload_id} [patch]
func (ctrl *TusModulController) UploadChunk(c *fiber.Ctx) error {
	return ctrl.uploadChunk(c, nil)
}

// UploadModulUpdateChunk handles PATCH /api/v1/modul/{id}/update/{upload_id} - Upload modul update chunk
// @Summary Upload modul update file chunk
// @Description Upload a chunk of data for a TUS resumable modul update upload
// @Tags TUS Modul Upload
// @Accept application/offset+octet-stream
// @Security BearerAuth
// @Param id path string true "Modul ID (UUID)"
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
// @Router /modul/{id}/update/{upload_id} [patch]
func (ctrl *TusModulController) UploadModulUpdateChunk(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.uploadChunk(c, &modulID)
}

func (ctrl *TusModulController) uploadChunk(c *fiber.Ctx, modulID *string) error {
	ctx := c.UserContext()
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return upload.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return upload.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	offset, _, bodyReader, err := parseChunkRequest(c)
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	var newOffset int64
	if modulID == nil {
		newOffset, err = ctrl.tusModulUsecase.HandleModulChunk(ctx, uploadID, userID, offset, bodyReader)
	} else {
		newOffset, err = ctrl.tusModulUsecase.HandleModulUpdateChunk(ctx, *modulID, uploadID, userID, offset, bodyReader)
	}
	if err != nil {
		return handleTusChunkError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusChunkResponse(c, newOffset)
}

// GetUploadStatus handles HEAD /api/v1/modul/upload/{upload_id} - Get modul upload status
// @Summary Get modul upload status
// @Description Get the current offset and length of a TUS modul upload via HEAD request
// @Tags TUS Modul Upload
// @Security BearerAuth
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 200 "Upload status"
// @Header 200 {string} Upload-Offset "Current byte offset"
// @Header 200 {string} Upload-Length "Total file size"
// @Header 200 {string} Tus-Resumable "TUS protocol version"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/upload/{upload_id} [head]
func (ctrl *TusModulController) GetUploadStatus(c *fiber.Ctx) error {
	return ctrl.getUploadStatus(c, nil)
}

// GetModulUpdateUploadStatus handles HEAD /api/v1/modul/{id}/update/{upload_id} - Get modul update upload status
// @Summary Get modul update upload status
// @Description Get the current offset and length of a TUS modul update upload via HEAD request
// @Tags TUS Modul Upload
// @Security BearerAuth
// @Param id path string true "Modul ID (UUID)"
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
// @Router /modul/{id}/update/{upload_id} [head]
func (ctrl *TusModulController) GetModulUpdateUploadStatus(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.getUploadStatus(c, &modulID)
}

func (ctrl *TusModulController) getUploadStatus(c *fiber.Ctx, modulID *string) error {
	ctx := c.UserContext()
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return upload.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return upload.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	var (
		offset int64
		length int64
		err    error
	)
	if modulID == nil {
		offset, length, err = ctrl.tusModulUsecase.GetModulUploadStatus(ctx, uploadID, userID)
	} else {
		offset, length, err = ctrl.tusModulUsecase.GetModulUpdateUploadStatus(ctx, *modulID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusHeadResponse(c, offset, length)
}

// GetUploadInfo handles GET /api/v1/modul/upload/:upload_id - Get modul upload info
// @Summary Get modul upload info
// @Description Mendapatkan informasi detail upload modul berdasarkan upload ID
// @Tags TUS Modul Upload
// @Produce json
// @Security BearerAuth
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.TusModulUploadInfoResponse} "Informasi upload berhasil didapat"
// @Failure 400 {object} dto.ErrorResponse "Upload ID tidak valid"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/upload/{upload_id} [get]
func (ctrl *TusModulController) GetUploadInfo(c *fiber.Ctx) error {
	return ctrl.getUploadInfo(c, nil)
}

// GetModulUpdateUploadInfo handles GET /api/v1/modul/:id/update/:upload_id - Get modul update upload info
// @Summary Get modul update upload info
// @Description Mendapatkan informasi detail upload update modul berdasarkan modul ID dan upload ID
// @Tags TUS Modul Upload
// @Produce json
// @Security BearerAuth
// @Param id path string true "Modul ID (UUID)"
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.TusModulUploadInfoResponse} "Informasi update upload berhasil didapat"
// @Failure 400 {object} dto.ErrorResponse "ID modul tidak valid"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/{id}/update/{upload_id} [get]
func (ctrl *TusModulController) GetModulUpdateUploadInfo(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return ctrl.base.SendBadRequest(c, "ID modul tidak valid")
	}
	return ctrl.getUploadInfo(c, &modulID)
}

func (ctrl *TusModulController) getUploadInfo(c *fiber.Ctx, modulID *string) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "Upload ID tidak valid")
	}

	var (
		info *dto.TusModulUploadInfoResponse
		err  error
	)
	ctx := c.UserContext()
	if modulID == nil {
		info, err = ctrl.tusModulUsecase.GetModulUploadInfo(ctx, uploadID, userID)
	} else {
		info, err = ctrl.tusModulUsecase.GetModulUpdateUploadInfo(ctx, *modulID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	message := "Informasi upload berhasil didapat"
	if modulID != nil {
		message = "Informasi update upload berhasil didapat"
	}

	return ctrl.base.SendSuccess(c, info, message)
}

// CancelUpload handles DELETE /api/v1/modul/upload/:upload_id - Cancel modul upload
// @Summary Cancel modul upload
// @Description Membatalkan upload modul yang sedang berlangsung berdasarkan upload ID
// @Tags TUS Modul Upload
// @Produce json
// @Security BearerAuth
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 204 "Upload berhasil dibatalkan"
// @Failure 400 {object} dto.ErrorResponse "Upload ID tidak valid"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/upload/{upload_id} [delete]
func (ctrl *TusModulController) CancelUpload(c *fiber.Ctx) error {
	return ctrl.cancelUpload(c, nil)
}

// CancelModulUpdateUpload handles DELETE /api/v1/modul/:id/update/:upload_id - Cancel modul update upload
// @Summary Cancel modul update upload
// @Description Membatalkan upload update modul yang sedang berlangsung berdasarkan modul ID dan upload ID
// @Tags TUS Modul Upload
// @Produce json
// @Security BearerAuth
// @Param id path string true "Modul ID (UUID)"
// @Param upload_id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 204 "Upload update berhasil dibatalkan"
// @Failure 400 {object} dto.ErrorResponse "ID modul tidak valid"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /modul/{id}/update/{upload_id} [delete]
func (ctrl *TusModulController) CancelModulUpdateUpload(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.cancelUpload(c, &modulID)
}

func (ctrl *TusModulController) cancelUpload(c *fiber.Ctx, modulID *string) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return upload.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return upload.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	ctx := c.UserContext()
	var err error
	if modulID == nil {
		err = ctrl.tusModulUsecase.CancelModulUpload(ctx, uploadID, userID)
	} else {
		err = ctrl.tusModulUsecase.CancelModulUpdateUpload(ctx, *modulID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusDeleteResponse(c)
}
