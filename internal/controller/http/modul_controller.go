package http

import (
	"bytes"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ModulController struct {
	modulUsecase    usecase.ModulUsecase
	tusModulUsecase usecase.TusModulUsecase
	config          *config.Config
}

func NewModulController(modulUsecase usecase.ModulUsecase, tusModulUsecase usecase.TusModulUsecase, config *config.Config) *ModulController {
	return &ModulController{
		modulUsecase:    modulUsecase,
		tusModulUsecase: tusModulUsecase,
		config:          config,
	}
}



func (ctrl *ModulController) GetList(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var params domain.ModulListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Parameter query tidak valid", nil)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	result, err := ctrl.modulUsecase.GetList(userID, params.Search, params.FilterType, params.FilterSemester, params.Page, params.Limit)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar modul: "+err.Error(), nil)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar modul berhasil diambil", result)
}



func (ctrl *ModulController) Delete(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	modulID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID modul tidak valid", nil)
	}

	err = ctrl.modulUsecase.Delete(uint(modulID), userID)
	if err != nil {
		if err.Error() == "modul tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke modul ini" {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Modul berhasil dihapus", nil)
}

func (ctrl *ModulController) CheckUploadSlot(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	response, err := ctrl.tusModulUsecase.CheckModulUploadSlot(userID)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Status slot upload berhasil didapat", response)
}

func (ctrl *ModulController) InitiateUpload(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	tusVersion := c.Get("Tus-Resumable")
	if tusVersion != ctrl.config.Upload.TusVersion {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusPreconditionFailed)
	}

	uploadLengthStr := c.Get("Upload-Length")
	if uploadLengthStr == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	fileSize, err := strconv.ParseInt(uploadLengthStr, 10, 64)
	if err != nil || fileSize <= 0 {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	uploadMetadata := c.Get("Upload-Metadata")
	if uploadMetadata == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	result, err := ctrl.tusModulUsecase.InitiateModulUpload(userID, fileSize, uploadMetadata)
	if err != nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		if strings.Contains(err.Error(), "antrian penuh") {
			return c.SendStatus(fiber.StatusTooManyRequests)
		}
		if strings.Contains(err.Error(), "metadata") {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if strings.Contains(err.Error(), "ukuran file melebihi") {
			return c.SendStatus(fiber.StatusRequestEntityTooLarge)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
	c.Set("Location", result.UploadURL)
	c.Set("Upload-Offset", "0")

	return helper.SendSuccessResponse(c, fiber.StatusCreated, "Upload modul berhasil diinisiasi", result)
}

func (ctrl *ModulController) UploadChunk(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	tusVersion := c.Get("Tus-Resumable")
	if tusVersion == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if tusVersion != ctrl.config.Upload.TusVersion {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusPreconditionFailed)
	}

	contentType := c.Get("Content-Type")
	if contentType != "application/offset+octet-stream" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnsupportedMediaType)
	}

	offsetStr := c.Get("Upload-Offset")
	if offsetStr == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil || offset < 0 {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	contentLengthStr := c.Get("Content-Length")
	if contentLengthStr == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	chunkSize, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil || chunkSize <= 0 {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if chunkSize > ctrl.config.Upload.ChunkSize*2 {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusRequestEntityTooLarge)
	}

	bodyBytes := c.Body()
	if bodyBytes == nil || len(bodyBytes) == 0 {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if int64(len(bodyBytes)) != chunkSize {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	bodyReader := bytes.NewReader(bodyBytes)

	newOffset, err := ctrl.tusModulUsecase.HandleModulChunk(uploadID, userID, offset, bodyReader)
	if err != nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)

		if strings.Contains(err.Error(), "tidak ditemukan") {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return c.SendStatus(fiber.StatusForbidden)
		}
		if strings.Contains(err.Error(), "offset tidak valid") {
			c.Set("Upload-Offset", fmt.Sprintf("%d", newOffset))
			return c.SendStatus(fiber.StatusConflict)
		}
		if strings.Contains(err.Error(), "tidak aktif") {
			return c.SendStatus(fiber.StatusLocked)
		}

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
	c.Set("Upload-Offset", fmt.Sprintf("%d", newOffset))

	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *ModulController) GetUploadStatus(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	offset, length, err := ctrl.tusModulUsecase.GetModulUploadStatus(uploadID, userID)
	if err != nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		if strings.Contains(err.Error(), "tidak ditemukan") {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
	c.Set("Upload-Offset", fmt.Sprintf("%d", offset))
	c.Set("Upload-Length", fmt.Sprintf("%d", length))

	return c.SendStatus(fiber.StatusOK)
}

func (ctrl *ModulController) GetUploadInfo(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Upload ID tidak valid", nil)
	}

	info, err := ctrl.tusModulUsecase.GetModulUploadInfo(uploadID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Informasi upload berhasil didapat", info)
}

func (ctrl *ModulController) CancelUpload(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	uploadID := c.Params("upload_id")
	if uploadID == "" {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	err := ctrl.tusModulUsecase.CancelModulUpload(uploadID, userID)
	if err != nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		if strings.Contains(err.Error(), "tidak ditemukan") {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return c.SendStatus(fiber.StatusForbidden)
		}
		if strings.Contains(err.Error(), "sudah selesai") {
			return c.SendStatus(fiber.StatusConflict)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *ModulController) UpdateMetadata(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	modulID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID modul tidak valid", nil)
	}

	var req domain.ModulUpdateMetadataRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	modul, err := ctrl.modulUsecase.GetByID(uint(modulID), userID)
	if err != nil {
		if err.Error() == "modul tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke modul ini" {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	modulDomain := &domain.Modul{
		ID:       modul.ID,
		UserID:   userID,
		NamaFile: req.NamaFile,
		Tipe:     modul.Tipe,
		Ukuran:   modul.Ukuran,
		Semester: req.Semester,
		PathFile: modul.PathFile,
	}

	if err := ctrl.modulUsecase.UpdateMetadataOnly(modulDomain); err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	modul.NamaFile = req.NamaFile
	modul.Semester = req.Semester

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Metadata modul berhasil diperbarui", modul)
}

func (ctrl *ModulController) Download(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var req domain.ModulDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if len(req.IDs) == 0 {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID modul tidak boleh kosong", nil)
	}

	filePath, err := ctrl.modulUsecase.Download(userID, req.IDs)
	if err != nil {
		if err.Error() == "modul tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return c.Download(filePath)
}
