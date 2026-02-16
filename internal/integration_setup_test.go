package internal

import (
	"encoding/json"
	"fmt"
	dto "invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationMiddlewareChain tests that all middleware work together correctly
func TestIntegrationMiddlewareChain(t *testing.T) {
	// Create Fiber app with middleware chain
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"code":    code,
			})
		},
	})

	// Apply middleware chain: RequestID
	app.Use(middleware.RequestID())

	// Add test handler
	app.Get("/test", func(c *fiber.Ctx) error {
		requestID := middleware.GetRequestID(c)
		return c.JSON(fiber.Map{
			"status":     "success",
			"message":    "Test successful",
			"request_id": requestID,
		})
	})

	// Make HTTP request through middleware chain
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify request ID is propagated
	requestID := resp.Header.Get(middleware.RequestIDHeader)
	assert.NotEmpty(t, requestID, "Request ID should be present in response headers")

	// Parse response body
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "success", result["status"].(string))
	assert.Equal(t, "Test successful", result["message"])
	assert.Equal(t, requestID, result["request_id"], "Request ID should match header")
}

// TestIntegrationErrorHandling tests error handling integration with AppError
func TestIntegrationErrorHandling(t *testing.T) {
	app := fiber.New()

	// Error handler that returns AppError
	app.Get("/error/notfound", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewNotFoundError("Test Resource"))
	})

	app.Get("/error/validation", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewValidationError("Field wajib diisi", nil))
	})

	app.Get("/error/unauthorized", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewUnauthorizedError("Token tidak valid"))
	})

	app.Get("/error/forbidden", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewForbiddenError("Akses ditolak"))
	})

	app.Get("/error/conflict", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewConflictError("Data sudah ada"))
	})

	app.Get("/error/internal", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewInternalError(fmt.Errorf("database error")))
	})

	t.Run("NotFoundError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/notfound", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Test Resource tidak ditemukan", result.Message)
		assert.Equal(t, fiber.StatusNotFound, result.Code)
	})

	t.Run("ValidationError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/validation", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Field wajib diisi", result.Message)
		assert.Equal(t, fiber.StatusBadRequest, result.Code)
	})

	t.Run("UnauthorizedError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/unauthorized", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Token tidak valid", result.Message)
		assert.Equal(t, fiber.StatusUnauthorized, result.Code)
	})

	t.Run("ForbiddenError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/forbidden", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Akses ditolak", result.Message)
		assert.Equal(t, fiber.StatusForbidden, result.Code)
	})

	t.Run("ConflictError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/conflict", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Data sudah ada", result.Message)
		assert.Equal(t, fiber.StatusConflict, result.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/internal", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Terjadi kesalahan pada server", result.Message)
		assert.Equal(t, fiber.StatusInternalServerError, result.Code)
	})
}

// TestIntegrationRequestIDPropagation tests that request ID is propagated through the entire request
func TestIntegrationRequestIDPropagation(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())

	app.Get("/propagate", func(c *fiber.Ctx) error {
		requestID := middleware.GetRequestID(c)
		return c.JSON(fiber.Map{
			"request_id": requestID,
			"from":       "handler",
		})
	})

	t.Run("GeneratedRequestID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/propagate", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		requestID := resp.Header.Get(middleware.RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, requestID, result["request_id"], "Request ID in response should match header")
	})

	t.Run("CustomRequestID", func(t *testing.T) {
		customID := "custom-request-id-12345"
		req := httptest.NewRequest("GET", "/propagate", nil)
		req.Header.Set(middleware.RequestIDHeader, customID)
		resp, err := app.Test(req)
		require.NoError(t, err)

		requestID := resp.Header.Get(middleware.RequestIDHeader)
		assert.Equal(t, customID, requestID, "Should use custom request ID")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, customID, result["request_id"])
	})
}
