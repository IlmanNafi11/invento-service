package http

import (
	"invento-service/config"
	"invento-service/internal/dto"
	"invento-service/internal/upload"
	"invento-service/internal/usecase"
	"strconv"

	base "invento-service/internal/controller/base"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type TusController struct {
	base       *base.BaseController
	tusUsecase usecase.TusUploadUsecase
	config     *config.Config
	validator  *validator.Validate
}

func NewTusController(tusUsecase usecase.TusUploadUsecase, cfg *config.Config, baseCtrl ...*base.BaseController) *TusController {
	resolvedBase := base.NewBaseController("", nil)
	if len(baseCtrl) > 0 && baseCtrl[0] != nil {
		resolvedBase = baseCtrl[0]
	}

	return &TusController{
		base:       resolvedBase,
		tusUsecase: tusUsecase,
		config:     cfg,
		validator:  validator.New(),
	}
}

// CheckUploadSlot handles GET /api/v1/project/upload/check-slot - Check upload slot availability
// @Summary Check project upload slot
// @Description Check if an upload slot is available for the authenticated user
// @Tags TUS Upload
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.TusUploadSlotResponse} "Slot status"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 409 {object} dto.ErrorResponse "No slot available"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/upload/check-slot [get]
func (ctrl *TusController) CheckUploadSlot(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	ctx := c.UserContext()
	result, err := ctrl.tusUsecase.CheckUploadSlot(ctx, userID)
	if err != nil {
		return ctrl.base.SendInternalError(c)
	}

	return upload.SendTusSlotResponse(
		c,
		result.Available,
		result.Message,
		result.QueueLength,
		map[bool]int{true: 1, false: 0}[result.ActiveUpload],
		result.MaxConcurrent,
	)
}

// ResetUploadQueue handles POST /api/v1/project/upload/reset-queue - Reset upload queue
// @Summary Reset project upload queue
// @Description Reset the upload queue for the authenticated user
// @Tags TUS Upload
// @Produce json
// @Security BearerAuth
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 200 {object} dto.SuccessResponse "Queue reset berhasil"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/upload/reset-queue [post]
func (ctrl *TusController) ResetUploadQueue(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	ctx := c.UserContext()
	if err := ctrl.tusUsecase.ResetUploadQueue(ctx, userID); err != nil {
		return ctrl.base.SendInternalError(c)
	}

	return ctrl.base.SendSuccess(c, nil, "Queue upload berhasil direset")
}

// InitiateUpload handles POST /api/v1/project/upload/ - Initiate new project upload
// @Summary Initiate project upload
// @Description Start a new TUS resumable upload for a project file
// @Tags TUS Upload
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Param Upload-Length header int true "Total file size in bytes"
// @Param Upload-Metadata header string true "Upload metadata (nama_project, kategori, semester)"
// @Success 201 {object} dto.SuccessResponse{data=dto.TusUploadResponse} "Upload initiated"
// @Header 201 {string} Location "Upload URL"
// @Header 201 {string} Tus-Resumable "TUS protocol version"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 409 {object} dto.ErrorResponse "No upload slot available"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/upload/ [post]
func (ctrl *TusController) InitiateUpload(c *fiber.Ctx) error {
	return ctrl.initiateUpload(c, nil)
}

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

func (ctrl *TusController) initiateUpload(c *fiber.Ctx, projectID *uint) error {
	userID, userEmail, userRole, err := getTusAuthContext(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	if err = validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	fileSize, err := ctrl.parseUploadLength(c)
	if err != nil {
		return ctrl.base.SendBadRequest(c, "Header Upload-Length tidak valid")
	}

	uploadMetadata := c.Get(upload.HeaderUploadMetadata)
	metadata := dto.TusUploadInitRequest{}
	if projectID == nil || uploadMetadata != "" {
		metadata, err = ctrl.parseUploadMetadata(uploadMetadata)
		if err != nil {
			return ctrl.base.SendBadRequest(c, "Format Upload-Metadata tidak valid")
		}
		if !ctrl.base.ValidateStruct(c, metadata) {
			return nil
		}
	}

	var result *dto.TusUploadResponse
	ctx := c.UserContext()
	if projectID == nil {
		result, err = ctrl.tusUsecase.InitiateUpload(ctx, userID, userEmail, userRole, fileSize, metadata)
	} else {
		result, err = ctrl.tusUsecase.InitiateProjectUpdateUpload(ctx, *projectID, userID, fileSize, metadata)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	upload.SetTusResponseHeaders(c, 0, fileSize)
	upload.SetTusLocationHeader(c, result.UploadURL)
	return upload.SendTusInitiateResponse(c, result.UploadID, result.UploadURL, fileSize)
}

// UploadChunk handles PATCH /api/v1/project/upload/{id} - Upload file chunk
// @Summary Upload project file chunk
// @Description Upload a chunk of data for a TUS resumable project upload
// @Tags TUS Upload
// @Accept application/offset+octet-stream
// @Security BearerAuth
// @Param id path string true "Upload ID"
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
// @Router /project/upload/{id} [patch]
func (ctrl *TusController) UploadChunk(c *fiber.Ctx) error {
	return ctrl.uploadChunk(c, nil)
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

func (ctrl *TusController) uploadChunk(c *fiber.Ctx, projectID *uint) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return upload.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("id")
	if projectID != nil {
		uploadID = c.Params("upload_id")
	}
	if uploadID == "" {
		return upload.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	offset, bodyReader, err := parseChunkRequest(c)
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	var newOffset int64
	ctx := c.UserContext()
	if projectID == nil {
		newOffset, err = ctrl.tusUsecase.HandleChunk(ctx, uploadID, userID, offset, bodyReader)
	} else {
		newOffset, err = ctrl.tusUsecase.HandleProjectUpdateChunk(ctx, *projectID, uploadID, userID, offset, bodyReader)
	}
	if err != nil {
		return handleTusChunkError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusChunkResponse(c, newOffset)
}

// GetUploadStatus handles HEAD /api/v1/project/upload/{id} - Get upload status
// @Summary Get project upload status
// @Description Get the current offset and length of a TUS project upload via HEAD request
// @Tags TUS Upload
// @Security BearerAuth
// @Param id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 200 "Upload status"
// @Header 200 {string} Upload-Offset "Current byte offset"
// @Header 200 {string} Upload-Length "Total file size"
// @Header 200 {string} Tus-Resumable "TUS protocol version"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/upload/{id} [head]
func (ctrl *TusController) GetUploadStatus(c *fiber.Ctx) error {
	return ctrl.getUploadStatus(c, nil)
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

func (ctrl *TusController) getUploadStatus(c *fiber.Ctx, projectID *uint) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	uploadID := c.Params("id")
	if projectID != nil {
		uploadID = c.Params("upload_id")
	}
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	var (
		offset int64
		length int64
		err    error
	)
	ctx := c.UserContext()
	if projectID == nil {
		offset, length, err = ctrl.tusUsecase.GetUploadStatus(ctx, uploadID, userID)
	} else {
		offset, length, err = ctrl.tusUsecase.GetProjectUpdateUploadStatus(ctx, *projectID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusHeadResponse(c, offset, length)
}

// GetUploadInfo handles GET /api/v1/project/upload/{id} - Get upload info
// @Summary Get project upload info
// @Description Get detailed information about a TUS project upload
// @Tags TUS Upload
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upload ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.TusUploadInfoResponse} "Upload info"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/upload/{id} [get]
func (ctrl *TusController) GetUploadInfo(c *fiber.Ctx) error {
	return ctrl.getUploadInfo(c, nil)
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

func (ctrl *TusController) getUploadInfo(c *fiber.Ctx, projectID *uint) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	uploadID := c.Params("id")
	if projectID != nil {
		uploadID = c.Params("upload_id")
	}
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	var (
		result *dto.TusUploadInfoResponse
		err    error
	)
	ctx := c.UserContext()
	if projectID == nil {
		result, err = ctrl.tusUsecase.GetUploadInfo(ctx, uploadID, userID)
	} else {
		result, err = ctrl.tusUsecase.GetProjectUpdateUploadInfo(ctx, *projectID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	message := "Informasi upload berhasil didapat"
	if projectID != nil {
		message = "Informasi update upload berhasil didapat"
	}

	return ctrl.base.SendSuccess(c, result, message)
}

// CancelUpload handles DELETE /api/v1/project/upload/{id} - Cancel upload
// @Summary Cancel project upload
// @Description Cancel and clean up a TUS project upload
// @Tags TUS Upload
// @Security BearerAuth
// @Param id path string true "Upload ID"
// @Param Tus-Resumable header string true "TUS protocol version" default(1.0.0)
// @Success 204 "Upload cancelled"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Upload not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /project/upload/{id} [delete]
func (ctrl *TusController) CancelUpload(c *fiber.Ctx) error {
	return ctrl.cancelUpload(c, nil)
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

func (ctrl *TusController) cancelUpload(c *fiber.Ctx, projectID *uint) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	uploadID := c.Params("id")
	if projectID != nil {
		uploadID = c.Params("upload_id")
	}
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "ID upload tidak valid")
	}

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	ctx := c.UserContext()
	var err error
	if projectID == nil {
		err = ctrl.tusUsecase.CancelUpload(ctx, uploadID, userID)
	} else {
		err = ctrl.tusUsecase.CancelProjectUpdateUpload(ctx, *projectID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusDeleteResponse(c)
}

func (ctrl *TusController) parseUploadLength(c *fiber.Ctx) (int64, error) {
	uploadLengthStr := c.Get(upload.HeaderUploadLength)
	if uploadLengthStr == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Upload-Length header required")
	}

	fileSize, err := strconv.ParseInt(uploadLengthStr, 10, 64)
	if err != nil || fileSize <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid Upload-Length")
	}

	return fileSize, nil
}

func (ctrl *TusController) parseUploadMetadata(metadataHeader string) (dto.TusUploadInitRequest, error) {
	var metadata dto.TusUploadInitRequest

	if metadataHeader == "" {
		return metadata, fiber.NewError(fiber.StatusBadRequest, "Upload-Metadata header required")
	}

	metadataMap := upload.ParseTusMetadata(metadataHeader)
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
