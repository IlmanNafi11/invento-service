package http

import (
	"bytes"
	"io"

	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"

	"github.com/gofiber/fiber/v2"
)

func getTusAuthContext(c *fiber.Ctx) (userID string, email string, role string, err error) {
	uid, ok := c.Locals("user_id").(string)
	if !ok || uid == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	userEmail, ok := c.Locals("user_email").(string)
	if !ok || userEmail == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	userRole, ok := c.Locals("user_role").(string)
	if !ok || userRole == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	return uid, userEmail, userRole, nil
}

func validateTusHeaders(c *fiber.Ctx, tusVersion string) error {
	if c.Get(helper.HeaderTusResumable) != tusVersion {
		return apperrors.NewTusVersionError(tusVersion)
	}

	if c.Method() == fiber.MethodPatch && c.Get(helper.HeaderContentType) != helper.TusContentType {
		return apperrors.NewValidationError("Content-Type harus application/offset+octet-stream", nil)
	}

	return nil
}

func parseChunkRequest(c *fiber.Ctx) (offset int64, chunkSize int64, body io.Reader, err error) {
	tusHeaders, err := helper.GetTusHeaders(c)
	if err != nil {
		return 0, 0, nil, err
	}

	if tusHeaders.UploadOffset < 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Upload-Offset tidak valid")
	}
	if tusHeaders.ContentLength <= 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Content-Length tidak valid")
	}
	if err := helper.ValidateChunkSize(tusHeaders.ContentLength); err != nil {
		return 0, 0, nil, err
	}

	bodyBytes := c.Body()
	if bodyBytes == nil || len(bodyBytes) == 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Request body kosong")
	}
	if int64(len(bodyBytes)) != tusHeaders.ContentLength {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Ukuran chunk tidak sesuai dengan Content-Length")
	}

	return tusHeaders.UploadOffset, tusHeaders.ContentLength, bytes.NewReader(bodyBytes), nil
}

func handleTusChunkError(c *fiber.Ctx, err error, tusVersion string) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*apperrors.AppError); ok {
		switch appErr.Code {
		case apperrors.ErrTusOffsetMismatch:
			return helper.SendTusErrorResponse(c, fiber.StatusConflict, tusVersion)
		case apperrors.ErrNotFound:
			return helper.SendTusErrorResponse(c, fiber.StatusNotFound, tusVersion)
		case apperrors.ErrForbidden:
			return helper.SendTusErrorResponse(c, fiber.StatusForbidden, tusVersion)
		case apperrors.ErrTusInactive:
			return helper.SendTusErrorResponse(c, fiber.StatusLocked, tusVersion)
		case apperrors.ErrTusAlreadyCompleted:
			return helper.SendTusErrorResponse(c, fiber.StatusConflict, tusVersion)
		case apperrors.ErrPayloadTooLarge:
			return helper.SendTusErrorResponse(c, fiber.StatusRequestEntityTooLarge, tusVersion)
		default:
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, tusVersion)
		}
	}

	return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, tusVersion)
}

func handleTusUsecaseError(c *fiber.Ctx, err error, tusVersion string) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*apperrors.AppError); ok {
		if c.Method() == fiber.MethodPatch || c.Method() == fiber.MethodHead || c.Method() == fiber.MethodDelete {
			return helper.SendTusErrorResponse(c, appErr.HTTPStatus, tusVersion)
		}
		return helper.SendAppError(c, appErr)
	}

	if c.Method() == fiber.MethodPatch || c.Method() == fiber.MethodHead || c.Method() == fiber.MethodDelete {
		return helper.SendTusErrorResponse(c, fiber.StatusInternalServerError, tusVersion)
	}

	return helper.SendInternalServerErrorResponse(c)
}
