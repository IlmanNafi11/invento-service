package http

import (
	"invento-service/config"
	base "invento-service/internal/controller/base"
	"invento-service/internal/dto"
	"invento-service/internal/upload"
	"invento-service/internal/usecase"
	"strconv"

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

func (ctrl *TusController) InitiateUpload(c *fiber.Ctx) error {
	return ctrl.initiateUpload(c, nil)
}

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

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
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

func (ctrl *TusController) UploadChunk(c *fiber.Ctx) error {
	return ctrl.uploadChunk(c, nil)
}

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

	offset, _, bodyReader, err := parseChunkRequest(c)
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

func (ctrl *TusController) GetUploadStatus(c *fiber.Ctx) error {
	return ctrl.getUploadStatus(c, nil)
}

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

func (ctrl *TusController) GetUploadInfo(c *fiber.Ctx) error {
	return ctrl.getUploadInfo(c, nil)
}

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

func (ctrl *TusController) CancelUpload(c *fiber.Ctx) error {
	return ctrl.cancelUpload(c, nil)
}

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
