package httputil

import (
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/version"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendSuccessResponse(c *fiber.Ctx, code int, message string, data interface{}) error {
	// Add version headers
	c.Set("X-API-Version", version.CurrentAPIVersion)
	c.Set("X-API-Deprecated", strconv.FormatBool(false))

	response := dto.SuccessResponse{
		BaseResponse: dto.BaseResponse{
			Status:  "success",
			Message: message,
			Code:    code,
		},
		Data:      data,
		Timestamp: time.Now(),
	}
	return c.Status(code).JSON(response)
}

func SendErrorResponse(c *fiber.Ctx, code int, message string, errors interface{}) error {
	// Add version headers
	c.Set("X-API-Version", version.CurrentAPIVersion)
	c.Set("X-API-Deprecated", strconv.FormatBool(false))

	response := dto.ErrorResponse{
		BaseResponse: dto.BaseResponse{
			Status:  "error",
			Message: message,
			Code:    code,
		},
		Errors:    errors,
		Timestamp: time.Now(),
	}
	return c.Status(code).JSON(response)
}

func SendListResponse(c *fiber.Ctx, code int, message string, items interface{}, pagination dto.PaginationData) error {
	listData := dto.ListData{
		Items:      items,
		Pagination: pagination,
	}
	return SendSuccessResponse(c, code, message, listData)
}

func SendValidationErrorResponse(c *fiber.Ctx, validationErrors []dto.ValidationError) error {
	return SendErrorResponse(c, fiber.StatusBadRequest, "Data validasi tidak valid", validationErrors)
}

func SendBadRequestResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Request tidak valid"
	}
	return SendErrorResponse(c, fiber.StatusBadRequest, message, nil)
}

func SendUnauthorizedResponse(c *fiber.Ctx) error {
	return SendErrorResponse(c, fiber.StatusUnauthorized, "Tidak memiliki akses", nil)
}

func SendForbiddenResponse(c *fiber.Ctx) error {
	return SendErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", nil)
}

func SendNotFoundResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Data tidak ditemukan"
	}
	return SendErrorResponse(c, fiber.StatusNotFound, message, nil)
}

func SendConflictResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Data sudah ada"
	}
	return SendErrorResponse(c, fiber.StatusConflict, message, nil)
}

func SendPayloadTooLargeResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Ukuran data melebihi batas maksimal"
	}
	return SendErrorResponse(c, fiber.StatusRequestEntityTooLarge, message, nil)
}

func SendTooManyRequestsResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "Terlalu banyak permintaan, silakan coba lagi nanti"
	}
	return SendErrorResponse(c, fiber.StatusTooManyRequests, message, nil)
}

func SendInternalServerErrorResponse(c *fiber.Ctx) error {
	return SendErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", nil)
}

// SendAppError handles *AppError types by mapping them to appropriate HTTP responses.
// This function provides a bridge between the new error handling system and the Fiber response helpers.
//
// Example usage in controllers:
//
//	if err != nil {
//	    if appErr, ok := err.(*apperrors.AppError); ok {
//	        return helper.SendAppError(c, appErr)
//	    }
//	    return helper.SendInternalServerErrorResponse(c)
//	}
//
// Migration path from string matching:
// OLD: switch err.Error() { case "user tidak ditemukan": ... }
// NEW: return errors.NewNotFoundError("User")
func SendAppError(c *fiber.Ctx, err *apperrors.AppError) error {
	return SendErrorResponse(c, err.HTTPStatus, err.Message, nil)
}
