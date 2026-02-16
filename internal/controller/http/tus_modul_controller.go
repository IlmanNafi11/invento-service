package http

import (
	"invento-service/config"
	base "invento-service/internal/controller/base"
	"invento-service/internal/domain"
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

func (ctrl *TusModulController) CheckUploadSlot(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	response, err := ctrl.tusModulUsecase.CheckModulUploadSlot(userID)
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return ctrl.base.SendSuccess(c, response, "Status slot upload berhasil didapat")
}

func (ctrl *TusModulController) InitiateUpload(c *fiber.Ctx) error {
	return ctrl.initiateUpload(c, nil)
}

func (ctrl *TusModulController) InitiateModulUpdateUpload(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.initiateUpload(c, &modulID)
}

func (ctrl *TusModulController) initiateUpload(c *fiber.Ctx, modulID *string) error {
	userID, _, _, err := getTusAuthContext(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	if err := validateTusHeaders(c, ctrl.config.Upload.TusVersion); err != nil {
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

	var result *domain.TusModulUploadResponse
	if modulID == nil {
		result, err = ctrl.tusModulUsecase.InitiateModulUpload(userID, tusHeaders.UploadLength, tusHeaders.UploadMetadata)
	} else {
		result, err = ctrl.tusModulUsecase.InitiateModulUpdateUpload(*modulID, userID, tusHeaders.UploadLength, tusHeaders.UploadMetadata)
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

func (ctrl *TusModulController) UploadChunk(c *fiber.Ctx) error {
	return ctrl.uploadChunk(c, nil)
}

func (ctrl *TusModulController) UploadModulUpdateChunk(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.uploadChunk(c, &modulID)
}

func (ctrl *TusModulController) uploadChunk(c *fiber.Ctx, modulID *string) error {
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
		newOffset, err = ctrl.tusModulUsecase.HandleModulChunk(uploadID, userID, offset, bodyReader)
	} else {
		newOffset, err = ctrl.tusModulUsecase.HandleModulUpdateChunk(*modulID, uploadID, userID, offset, bodyReader)
	}
	if err != nil {
		return handleTusChunkError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusChunkResponse(c, newOffset)
}

func (ctrl *TusModulController) GetUploadStatus(c *fiber.Ctx) error {
	return ctrl.getUploadStatus(c, nil)
}

func (ctrl *TusModulController) GetModulUpdateUploadStatus(c *fiber.Ctx) error {
	modulID, err := ctrl.base.ParsePathUUID(c)
	if err != nil {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}
	return ctrl.getUploadStatus(c, &modulID)
}

func (ctrl *TusModulController) getUploadStatus(c *fiber.Ctx, modulID *string) error {
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
		offset, length, err = ctrl.tusModulUsecase.GetModulUploadStatus(uploadID, userID)
	} else {
		offset, length, err = ctrl.tusModulUsecase.GetModulUpdateUploadStatus(*modulID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusHeadResponse(c, offset, length)
}

func (ctrl *TusModulController) GetUploadInfo(c *fiber.Ctx) error {
	return ctrl.getUploadInfo(c, nil)
}

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
		info *domain.TusModulUploadInfoResponse
		err  error
	)
	if modulID == nil {
		info, err = ctrl.tusModulUsecase.GetModulUploadInfo(uploadID, userID)
	} else {
		info, err = ctrl.tusModulUsecase.GetModulUpdateUploadInfo(*modulID, uploadID, userID)
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

func (ctrl *TusModulController) CancelUpload(c *fiber.Ctx) error {
	return ctrl.cancelUpload(c, nil)
}

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

	var err error
	if modulID == nil {
		err = ctrl.tusModulUsecase.CancelModulUpload(uploadID, userID)
	} else {
		err = ctrl.tusModulUsecase.CancelModulUpdateUpload(*modulID, uploadID, userID)
	}
	if err != nil {
		return handleTusUsecaseError(c, err, ctrl.config.Upload.TusVersion)
	}

	return upload.SendTusDeleteResponse(c)
}
