package http

import (
	"bytes"
	"fiber-boiler-plate/config"
	base "fiber-boiler-plate/internal/controller/base"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type TusModulController struct {
	base            *base.BaseController
	tusModulUsecase usecase.TusModulUsecase
	config          *config.Config
	validator       *validator.Validate
}

func NewTusModulController(tusModulUsecase usecase.TusModulUsecase, cfg *config.Config, baseCtrl *base.BaseController) *TusModulController {
	return &TusModulController{
		base:            baseCtrl,
		tusModulUsecase: tusModulUsecase,
		config:          cfg,
		validator:       validator.New(),
	}
}

func (ctrl *TusModulController) CheckUploadSlot(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	response, err := ctrl.tusModulUsecase.CheckModulUploadSlot(userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.base.SendInternalError(c)
	}

	return ctrl.base.SendSuccess(c, response, "Status slot upload berhasil didapat")
}

func (ctrl *TusModulController) InitiateUpload(c *fiber.Ctx) error {
	userID, _, _, err := ctrl.getTusAuthContext(c)
	if err != nil {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion != ctrl.config.Upload.TusVersion {
		return helper.SendTusErrorResponse(c, fiber.StatusPreconditionFailed, ctrl.config.Upload.TusVersion)
	}

	uploadLengthStr := c.Get(helper.HeaderUploadLength)
	if uploadLengthStr == "" {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Length wajib diisi")
	}

	fileSize, err := strconv.ParseInt(uploadLengthStr, 10, 64)
	if err != nil || fileSize <= 0 {
		return helper.SendTusValidationErrorResponse(c, "Ukuran file tidak valid")
	}

	uploadMetadata := c.Get(helper.HeaderUploadMetadata)
	if uploadMetadata == "" {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Metadata wajib diisi")
	}

	result, err := ctrl.tusModulUsecase.InitiateModulUpload(userID, fileSize, uploadMetadata)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
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

	helper.SetTusResponseHeaders(c, 0, fileSize)
	helper.SetTusLocationHeader(c, result.UploadURL)

	return ctrl.base.SendCreated(c, result, "Upload modul berhasil diinisiasi")
}

func (ctrl *TusModulController) UploadChunk(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	if err := ctrl.validateTusHeaders(c); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
		}
		return helper.SendTusErrorResponse(c, fiber.StatusPreconditionFailed, ctrl.config.Upload.TusVersion)
	}

	tusHeaders := helper.GetTusHeaders(c)
	if tusHeaders.UploadOffset < 0 {
		return helper.SendTusValidationErrorResponse(c, "Upload-Offset tidak valid")
	}
	if tusHeaders.ContentLength <= 0 {
		return helper.SendTusValidationErrorResponse(c, "Content-Length tidak valid")
	}
	if err := helper.ValidateChunkSize(tusHeaders.ContentLength); err != nil {
		return helper.SendTusErrorResponseWithLength(c, fiber.StatusRequestEntityTooLarge, ctrl.config.Upload.TusVersion, tusHeaders.ContentLength)
	}

	bodyBytes := c.Body()
	if bodyBytes == nil || len(bodyBytes) == 0 {
		return helper.SendTusValidationErrorResponse(c, "Request body kosong")
	}
	if int64(len(bodyBytes)) != tusHeaders.ContentLength {
		return helper.SendTusValidationErrorResponse(c, "Ukuran chunk tidak sesuai dengan Content-Length")
	}

	bodyReader := bytes.NewReader(bodyBytes)
	newOffset, err := ctrl.tusModulUsecase.HandleModulChunk(uploadID, userID, tusHeaders.UploadOffset, bodyReader)
	if err != nil {
		return ctrl.handleTusChunkError(c, err, newOffset)
	}

	return helper.SendTusChunkResponse(c, newOffset)
}

func (ctrl *TusModulController) GetUploadStatus(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	offset, length, err := ctrl.tusModulUsecase.GetModulUploadStatus(uploadID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	return helper.SendTusHeadResponse(c, offset, length)
}

func (ctrl *TusModulController) GetUploadInfo(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "Upload ID tidak valid")
	}

	info, err := ctrl.tusModulUsecase.GetModulUploadInfo(uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err)
	}

	return ctrl.base.SendSuccess(c, info, "Informasi upload berhasil didapat")
}

func (ctrl *TusModulController) CancelUpload(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	err := ctrl.tusModulUsecase.CancelModulUpload(uploadID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	return helper.SendTusDeleteResponse(c)
}

func (ctrl *TusModulController) InitiateModulUpdateUpload(c *fiber.Ctx) error {
	userID, _, _, err := ctrl.getTusAuthContext(c)
	if err != nil {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	modulID, err := ctrl.base.ParsePathID(c)
	if err != nil {
		return helper.SendTusErrorResponse(c, fiber.StatusBadRequest, ctrl.config.Upload.TusVersion)
	}

	tusHeaders := helper.GetTusHeaders(c)
	if tusHeaders.TusResumable != ctrl.config.Upload.TusVersion {
		return helper.SendTusErrorResponse(c, fiber.StatusPreconditionFailed, ctrl.config.Upload.TusVersion)
	}
	if tusHeaders.UploadLength <= 0 {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Length wajib diisi")
	}
	if tusHeaders.UploadMetadata == "" {
		return helper.SendTusValidationErrorResponse(c, "Header Upload-Metadata wajib diisi")
	}

	result, err := ctrl.tusModulUsecase.InitiateModulUpdateUpload(modulID, userID, tusHeaders.UploadLength, tusHeaders.UploadMetadata)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
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

	helper.SetTusResponseHeaders(c, 0, tusHeaders.UploadLength)
	helper.SetTusLocationHeader(c, result.UploadURL)

	return ctrl.base.SendCreated(c, result, "Update upload modul berhasil diinisiasi")
}

func (ctrl *TusModulController) UploadModulUpdateChunk(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	if err := ctrl.validateTusHeaders(c); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
		}
		return helper.SendTusErrorResponse(c, fiber.StatusPreconditionFailed, ctrl.config.Upload.TusVersion)
	}

	tusHeaders := helper.GetTusHeaders(c)
	if tusHeaders.UploadOffset < 0 {
		return helper.SendTusValidationErrorResponse(c, "Upload-Offset tidak valid")
	}
	if tusHeaders.ContentLength <= 0 {
		return helper.SendTusValidationErrorResponse(c, "Content-Length tidak valid")
	}
	if err := helper.ValidateChunkSize(tusHeaders.ContentLength); err != nil {
		return helper.SendTusErrorResponseWithLength(c, fiber.StatusRequestEntityTooLarge, ctrl.config.Upload.TusVersion, tusHeaders.ContentLength)
	}

	bodyBytes := c.Body()
	if bodyBytes == nil || len(bodyBytes) == 0 {
		return helper.SendTusValidationErrorResponse(c, "Request body kosong")
	}
	if int64(len(bodyBytes)) != tusHeaders.ContentLength {
		return helper.SendTusValidationErrorResponse(c, "Ukuran chunk tidak sesuai dengan Content-Length")
	}

	bodyReader := bytes.NewReader(bodyBytes)
	newOffset, err := ctrl.tusModulUsecase.HandleModulUpdateChunk(uploadID, userID, tusHeaders.UploadOffset, bodyReader)
	if err != nil {
		return ctrl.handleTusChunkError(c, err, newOffset)
	}

	return helper.SendTusChunkResponse(c, newOffset)
}

func (ctrl *TusModulController) GetModulUpdateUploadStatus(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	offset, length, err := ctrl.tusModulUsecase.GetModulUploadStatus(uploadID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	return helper.SendTusHeadResponse(c, offset, length)
}

func (ctrl *TusModulController) GetModulUpdateUploadInfo(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return ctrl.base.SendBadRequest(c, "Upload ID tidak valid")
	}

	info, err := ctrl.tusModulUsecase.GetModulUploadInfo(uploadID, userID)
	if err != nil {
		return ctrl.handleTusUsecaseError(c, err)
	}

	return ctrl.base.SendSuccess(c, info, "Informasi update upload berhasil didapat")
}

func (ctrl *TusModulController) CancelModulUpdateUpload(c *fiber.Ctx) error {
	userID := ctrl.base.GetAuthenticatedUserID(c)
	if userID == "" {
		return helper.SendTusErrorResponse(c, fiber.StatusUnauthorized, ctrl.config.Upload.TusVersion)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendTusValidationErrorResponse(c, "Upload ID wajib diisi")
	}

	err := ctrl.tusModulUsecase.CancelModulUpload(uploadID, userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
		}
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
	}

	return helper.SendTusDeleteResponse(c)
}

func (ctrl *TusModulController) getTusAuthContext(c *fiber.Ctx) (string, string, string, error) {
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

func (ctrl *TusModulController) validateTusHeaders(c *fiber.Ctx) error {
	tusVersion := c.Get(helper.HeaderTusResumable)
	if tusVersion == "" || tusVersion != ctrl.config.Upload.TusVersion {
		return apperrors.NewTusVersionMismatchError(ctrl.config.Upload.TusVersion)
	}

	contentType := c.Get(helper.HeaderContentType)
	if contentType != helper.TusContentType {
		return apperrors.NewValidationError("Content-Type harus application/offset+octet-stream", nil)
	}

	return nil
}

func (ctrl *TusModulController) handleTusUsecaseError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*apperrors.AppError); ok {
		return helper.SendAppError(c, appErr)
	}

	return ctrl.base.SendInternalError(c)
}

func (ctrl *TusModulController) handleTusChunkError(c *fiber.Ctx, err error, offset int64) error {
	if err == nil {
		return nil
	}

	helper.SetTusResponseHeaders(c, 0, 0)

	if appErr, ok := err.(*apperrors.AppError); ok {
		switch appErr.Code {
		case apperrors.ErrNotFound:
			return helper.SendTusErrorResponse(c, fiber.StatusNotFound, ctrl.config.Upload.TusVersion)
		case apperrors.ErrForbidden:
			return helper.SendTusErrorResponse(c, fiber.StatusForbidden, ctrl.config.Upload.TusVersion)
		case apperrors.ErrTusOffsetMismatch:
			return helper.SendTusErrorResponseWithOffset(c, fiber.StatusConflict, ctrl.config.Upload.TusVersion, offset)
		case apperrors.ErrTusInactive:
			return helper.SendTusErrorResponse(c, fiber.StatusLocked, ctrl.config.Upload.TusVersion)
		case apperrors.ErrTusAlreadyCompleted:
			return helper.SendTusErrorResponse(c, fiber.StatusConflict, ctrl.config.Upload.TusVersion)
		default:
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, ctrl.config.Upload.TusVersion)
		}
	}

	return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, ctrl.config.Upload.TusVersion)
}
