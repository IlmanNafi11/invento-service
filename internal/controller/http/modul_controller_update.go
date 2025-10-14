package http

import (
	"bytes"
	"fiber-boiler-plate/internal/helper"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// InitiateModulUpdateUpload handles POST /modul/:id/upload
func (ctrl *ModulController) InitiateModulUpdateUpload(c *fiber.Ctx) error {
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

	modulIDStr := c.Params("id")
	modulID, err := strconv.Atoi(modulIDStr)
	if err != nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
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

	result, err := ctrl.tusModulUsecase.InitiateModulUpdateUpload(uint(modulID), userID, fileSize, uploadMetadata)
	if err != nil {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		if strings.Contains(err.Error(), "tidak ditemukan") {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return c.SendStatus(fiber.StatusForbidden)
		}
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

	return helper.SendSuccessResponse(c, fiber.StatusCreated, "Update upload modul berhasil diinisiasi", result)
}

// UploadModulUpdateChunk handles PATCH /modul/:id/update/:upload_id
func (ctrl *ModulController) UploadModulUpdateChunk(c *fiber.Ctx) error {
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

	newOffset, err := ctrl.tusModulUsecase.HandleModulUpdateChunk(uploadID, userID, offset, bodyReader)
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

// GetModulUpdateUploadStatus handles HEAD /modul/:id/update/:upload_id
func (ctrl *ModulController) GetModulUpdateUploadStatus(c *fiber.Ctx) error {
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

// GetModulUpdateUploadInfo handles GET /modul/:id/update/:upload_id
func (ctrl *ModulController) GetModulUpdateUploadInfo(c *fiber.Ctx) error {
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

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Informasi update upload berhasil didapat", info)
}

// CancelModulUpdateUpload handles DELETE /modul/:id/update/:upload_id
func (ctrl *ModulController) CancelModulUpdateUpload(c *fiber.Ctx) error {
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
