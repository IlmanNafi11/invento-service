package helper

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendTusInitiateResponse(c *fiber.Ctx, uploadID string, uploadURL string, fileSize int64) error {
	SetTusResponseHeaders(c, 0, fileSize)
	SetTusLocationHeader(c, uploadURL)

	response := map[string]interface{}{
		"success": true,
		"message": "Upload berhasil diinisiasi",
		"code":    fiber.StatusCreated,
		"data": map[string]interface{}{
			"upload_id":  uploadID,
			"upload_url": uploadURL,
			"offset":     0,
			"length":     fileSize,
		},
		"timestamp": time.Now(),
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func SendTusChunkResponse(c *fiber.Ctx, newOffset int64) error {
	c.Set(HeaderTusResumable, TusVersion)
	c.Set(HeaderUploadOffset, strconv.FormatInt(newOffset, 10))

	return c.SendStatus(fiber.StatusNoContent)
}

func SendTusHeadResponse(c *fiber.Ctx, offset int64, length int64) error {
	c.Set(HeaderTusResumable, TusVersion)
	c.Set(HeaderUploadOffset, strconv.FormatInt(offset, 10))
	c.Set(HeaderUploadLength, strconv.FormatInt(length, 10))

	return c.SendStatus(fiber.StatusOK)
}

func SendTusDeleteResponse(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func SendTusSlotResponse(c *fiber.Ctx, available bool, message string, queueLength int, activeUpload bool, maxConcurrent int) error {
	response := map[string]interface{}{
		"success": true,
		"message": "Pengecekan slot upload berhasil",
		"code":    fiber.StatusOK,
		"data": map[string]interface{}{
			"available":      available,
			"message":        message,
			"queue_length":   queueLength,
			"active_upload":  activeUpload,
			"max_concurrent": maxConcurrent,
		},
		"timestamp": time.Now(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func SendTusModulSlotResponse(c *fiber.Ctx, available bool, message string, queueLength int, maxQueue int) error {
	response := map[string]interface{}{
		"success": true,
		"message": "Status slot upload berhasil didapat",
		"code":    fiber.StatusOK,
		"data": map[string]interface{}{
			"available":    available,
			"message":      message,
			"queue_length": queueLength,
			"max_queue":    maxQueue,
		},
		"timestamp": time.Now(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
