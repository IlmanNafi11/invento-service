package http

import (
	"bytes"
	"io"

	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"invento-service/internal/upload"

	"github.com/gofiber/fiber/v2"
)

func getTusAuthContext(c *fiber.Ctx) (userID string, email string, role string, err error) {
	uid, ok := c.Locals(middleware.LocalsKeyUserID).(string)
	if !ok || uid == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	userEmail, ok := c.Locals(middleware.LocalsKeyUserEmail).(string)
	if !ok || userEmail == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	userRole, ok := c.Locals(middleware.LocalsKeyUserRole).(string)
	if !ok || userRole == "" {
		return "", "", "", fiber.ErrUnauthorized
	}

	return uid, userEmail, userRole, nil
}

func validateTusHeaders(c *fiber.Ctx, tusVersion string) error {
	if c.Get(upload.HeaderTusResumable) != tusVersion {
		return apperrors.NewTusVersionError(tusVersion)
	}

	if c.Method() == fiber.MethodPatch && c.Get(upload.HeaderContentType) != upload.TusContentType {
		return apperrors.NewValidationError("Content-Type harus application/offset+octet-stream", nil)
	}

	return nil
}

func parseChunkRequest(c *fiber.Ctx) (offset int64, chunkSize int64, body io.Reader, err error) {
	tusHeaders, err := upload.GetTusHeaders(c)
	if err != nil {
		return 0, 0, nil, err
	}

	if tusHeaders.UploadOffset < 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Upload-Offset tidak valid")
	}
	if tusHeaders.ContentLength <= 0 {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, "Content-Length tidak valid")
	}
	if err := upload.ValidateChunkSize(tusHeaders.ContentLength); err != nil {
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
			return upload.SendTusErrorResponse(c, fiber.StatusConflict, tusVersion)
		case apperrors.ErrNotFound:
			return upload.SendTusErrorResponse(c, fiber.StatusNotFound, tusVersion)
		case apperrors.ErrForbidden:
			return upload.SendTusErrorResponse(c, fiber.StatusForbidden, tusVersion)
		case apperrors.ErrTusInactive:
			return upload.SendTusErrorResponse(c, fiber.StatusLocked, tusVersion)
		case apperrors.ErrTusAlreadyCompleted:
			return upload.SendTusErrorResponse(c, fiber.StatusConflict, tusVersion)
		case apperrors.ErrPayloadTooLarge:
			return upload.SendTusErrorResponse(c, fiber.StatusRequestEntityTooLarge, tusVersion)
		default:
			return upload.SendTusErrorResponse(c, appErr.HTTPStatus, tusVersion)
		}
	}

	return upload.SendTusErrorResponse(c, fiber.StatusInternalServerError, tusVersion)
}

func handleTusUsecaseError(c *fiber.Ctx, err error, tusVersion string) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*apperrors.AppError); ok {
		if c.Method() == fiber.MethodPatch || c.Method() == fiber.MethodHead || c.Method() == fiber.MethodDelete {
			return upload.SendTusErrorResponse(c, appErr.HTTPStatus, tusVersion)
		}
		return httputil.SendAppError(c, appErr)
	}

	if c.Method() == fiber.MethodPatch || c.Method() == fiber.MethodHead || c.Method() == fiber.MethodDelete {
		return upload.SendTusErrorResponse(c, fiber.StatusInternalServerError, tusVersion)
	}

	return httputil.SendInternalServerErrorResponse(c)
}
