package http

import (
	"bytes"
	"encoding/base64"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type TusController struct {
	tusUsecase usecase.TusUploadUsecase
	config     *config.Config
	validator  *validator.Validate
}

func NewTusController(tusUsecase usecase.TusUploadUsecase, cfg *config.Config) *TusController {
	return &TusController{
		tusUsecase: tusUsecase,
		config:     cfg,
		validator:  validator.New(),
	}
}

func (ctrl *TusController) CheckUploadSlot(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	result, err := ctrl.tusUsecase.CheckUploadSlot(userID)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Pengecekan slot upload berhasil", result)
}

func (ctrl *TusController) ResetUploadQueue(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	err := ctrl.tusUsecase.ResetUploadQueue(userID)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Queue upload berhasil direset", nil)
}

func (ctrl *TusController) InitiateUpload(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	userEmailVal := c.Locals("user_email")
	if userEmailVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userEmail, ok := userEmailVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	userRoleVal := c.Locals("user_role")
	if userRoleVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userRole, ok := userRoleVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	tusVersion := c.Get("Tus-Resumable")
	if tusVersion == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Header Tus-Resumable wajib diisi", nil)
	}

	if tusVersion != ctrl.config.Upload.TusVersion {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Versi Tus tidak didukung, gunakan %s", ctrl.config.Upload.TusVersion), nil)
	}

	uploadLengthStr := c.Get("Upload-Length")
	if uploadLengthStr == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Header Upload-Length wajib diisi", nil)
	}

	fileSize, err := strconv.ParseInt(uploadLengthStr, 10, 64)
	if err != nil || fileSize <= 0 {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Header Upload-Length tidak valid", nil)
	}

	uploadMetadata := c.Get("Upload-Metadata")
	if uploadMetadata == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Header Upload-Metadata wajib diisi", nil)
	}

	metadata, err := ctrl.parseMetadata(uploadMetadata)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format Upload-Metadata tidak valid", nil)
	}

	if err := ctrl.validator.Struct(metadata); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Data validasi tidak valid", nil)
	}

	result, err := ctrl.tusUsecase.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)
	if err != nil {
		if strings.Contains(err.Error(), "melebihi batas maksimal") {
			return helper.SendErrorResponse(c, fiber.StatusRequestEntityTooLarge, err.Error(), nil)
		}
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
	c.Set("Location", result.UploadURL)
	c.Set("Upload-Offset", "0")

	return helper.SendSuccessResponse(c, fiber.StatusCreated, "Upload project berhasil diinisiasi", result)
}

func (ctrl *TusController) UploadChunk(c *fiber.Ctx) error {
	log.Printf("=== UPLOAD CHUNK CONTROLLER START ===")
	log.Printf("Upload ID from params: %s", c.Params("id"))
	
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		log.Printf("ERROR: user_id not found in context")
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		log.Printf("ERROR: user_id type assertion failed")
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	log.Printf("User ID: %d", userID)

	uploadID := c.Params("id")
	if uploadID == "" {
		log.Printf("ERROR: Upload ID is empty")
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	tusVersion := c.Get("Tus-Resumable")
	if tusVersion == "" {
		log.Printf("DEBUG PATCH: Missing Tus-Resumable header. Headers: Content-Type=%s, Upload-Offset=%s, Content-Length=%s\n", 
			c.Get("Content-Type"), c.Get("Upload-Offset"), c.Get("Content-Length"))
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}
	
	if tusVersion != ctrl.config.Upload.TusVersion {
		log.Printf("DEBUG PATCH: Invalid Tus-Resumable version: got %s, expected %s\n", tusVersion, ctrl.config.Upload.TusVersion)
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusPreconditionFailed)
	}

	contentType := c.Get("Content-Type")
	if contentType != "application/offset+octet-stream" {
		log.Printf("DEBUG PATCH: Invalid Content-Type: got '%s', expected 'application/offset+octet-stream'\n", contentType)
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusUnsupportedMediaType)
	}

	offsetStr := c.Get("Upload-Offset")
	if offsetStr == "" {
		log.Printf("DEBUG PATCH: Missing Upload-Offset header\n")
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
		log.Printf("ERROR: Missing Content-Length header")
		log.Printf("Available headers:")
		c.Request().Header.VisitAll(func(key, value []byte) {
			log.Printf("  %s: %s", string(key), string(value))
		})
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}
	log.Printf("Content-Length: %s", contentLengthStr)

	chunkSize, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil || chunkSize <= 0 {
		log.Printf("ERROR: Invalid Content-Length: %s (error: %v)", contentLengthStr, err)
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}
	log.Printf("Chunk size parsed: %d bytes", chunkSize)

	if chunkSize > ctrl.config.Upload.ChunkSize*2 {
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusRequestEntityTooLarge)
	}

	// Get request body as byte slice
	bodyBytes := c.Body()
	if bodyBytes == nil || len(bodyBytes) == 0 {
		log.Printf("ERROR: Request body is empty")
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}
	
	if int64(len(bodyBytes)) != chunkSize {
		log.Printf("ERROR: Body size mismatch - Content-Length: %d, Actual: %d", chunkSize, len(bodyBytes))
		c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
		return c.SendStatus(fiber.StatusBadRequest)
	}
	
	log.Printf("Body size OK: %d bytes, creating reader...", len(bodyBytes))
	
	// Create a reader from the byte slice
	bodyReader := bytes.NewReader(bodyBytes)
	log.Printf("Reader created, calling usecase.HandleChunk...")

	newOffset, err := ctrl.tusUsecase.HandleChunk(uploadID, userID, offset, bodyReader, chunkSize)
	if err != nil {
		log.Printf("ERROR from HandleChunk: %v", err)
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

	log.Printf("SUCCESS: Chunk uploaded, new offset: %d", newOffset)
	log.Printf("=== UPLOAD CHUNK CONTROLLER END ===")
	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *TusController) GetUploadStatus(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	uploadID := c.Params("id")
	if uploadID == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID upload tidak valid", nil)
	}

	tusVersion := c.Get("Tus-Resumable")
	if tusVersion != ctrl.config.Upload.TusVersion {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Versi Tus tidak didukung", nil)
	}

	offset, length, err := ctrl.tusUsecase.GetUploadStatus(uploadID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			return helper.SendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	c.Set("Tus-Resumable", ctrl.config.Upload.TusVersion)
	c.Set("Upload-Offset", fmt.Sprintf("%d", offset))
	c.Set("Upload-Length", fmt.Sprintf("%d", length))

	return c.SendStatus(fiber.StatusOK)
}

func (ctrl *TusController) GetUploadInfo(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	uploadID := c.Params("id")
	if uploadID == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID upload tidak valid", nil)
	}

	result, err := ctrl.tusUsecase.GetUploadInfo(uploadID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Informasi upload berhasil didapat", result)
}

func (ctrl *TusController) CancelUpload(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	uploadID := c.Params("id")
	if uploadID == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID upload tidak valid", nil)
	}

	tusVersion := c.Get("Tus-Resumable")
	if tusVersion != ctrl.config.Upload.TusVersion {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Versi Tus tidak didukung", nil)
	}

	err := ctrl.tusUsecase.CancelUpload(uploadID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			return helper.SendErrorResponse(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if strings.Contains(err.Error(), "tidak memiliki akses") {
			return helper.SendForbiddenResponse(c)
		}
		if strings.Contains(err.Error(), "sudah selesai") {
			return helper.SendErrorResponse(c, fiber.StatusConflict, err.Error(), nil)
		}
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (ctrl *TusController) parseMetadata(metadataHeader string) (domain.TusUploadInitRequest, error) {
	var metadata domain.TusUploadInitRequest

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
