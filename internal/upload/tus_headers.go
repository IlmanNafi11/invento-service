package upload

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const (
	HeaderTusResumable   = "Tus-Resumable"
	HeaderUploadOffset   = "Upload-Offset"
	HeaderUploadLength   = "Upload-Length"
	HeaderUploadMetadata = "Upload-Metadata"
	HeaderContentType    = "Content-Type"
	HeaderContentLength  = "Content-Length"
	HeaderLocation       = "Location"

	TusVersion         = "1.0.0"
	TusContentType     = "application/offset+octet-stream"
	DefaultChunkSize   = 1048576
	MaxChunkSize       = 2097152
	MaxProjectFileSize = 524288000
	MaxModulFileSize   = 52428800
)

type TusHeaders struct {
	TusResumable   string
	UploadOffset   int64
	UploadLength   int64
	UploadMetadata string
	ContentType    string
	ContentLength  int64
	Location       string
}

func GetTusHeaders(c *fiber.Ctx) (TusHeaders, error) {
	headers := TusHeaders{
		TusResumable:   c.Get(HeaderTusResumable),
		UploadMetadata: c.Get(HeaderUploadMetadata),
		ContentType:    c.Get(HeaderContentType),
	}

	if offsetStr := c.Get(HeaderUploadOffset); offsetStr != "" {
		offset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			return TusHeaders{}, fmt.Errorf("header %s tidak valid: %w", HeaderUploadOffset, err)
		}
		headers.UploadOffset = offset
	}

	if lengthStr := c.Get(HeaderUploadLength); lengthStr != "" {
		length, err := strconv.ParseInt(lengthStr, 10, 64)
		if err != nil {
			return TusHeaders{}, fmt.Errorf("header %s tidak valid: %w", HeaderUploadLength, err)
		}
		headers.UploadLength = length
	}

	if contentLengthStr := c.Get(HeaderContentLength); contentLengthStr != "" {
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			return TusHeaders{}, fmt.Errorf("header %s tidak valid: %w", HeaderContentLength, err)
		}
		headers.ContentLength = contentLength
	}

	return headers, nil
}

func SetTusResponseHeaders(c *fiber.Ctx, offset int64, length int64) {
	c.Set(HeaderTusResumable, TusVersion)
	c.Set(HeaderUploadOffset, strconv.FormatInt(offset, 10))

	if length > 0 {
		c.Set(HeaderUploadLength, strconv.FormatInt(length, 10))
	}
}

func SetTusLocationHeader(c *fiber.Ctx, location string) {
	c.Set(HeaderLocation, location)
}

func SetTusOffsetHeader(c *fiber.Ctx, offset int64) {
	c.Set(HeaderUploadOffset, strconv.FormatInt(offset, 10))
}

func ValidateChunkSize(size int64) error {
	if size <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "ukuran chunk tidak valid")
	}

	if size > MaxChunkSize {
		maxSizeMB := MaxChunkSize / (1024 * 1024)
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, strconv.FormatInt(int64(maxSizeMB), 10)+" MB")
	}

	return nil
}

func BuildTusErrorResponse(c *fiber.Ctx, statusCode int, offset int64) error {
	c.Set(HeaderTusResumable, TusVersion)

	if statusCode == fiber.StatusConflict && offset >= 0 {
		c.Set(HeaderUploadOffset, strconv.FormatInt(offset, 10))
	}

	return c.SendStatus(statusCode)
}
