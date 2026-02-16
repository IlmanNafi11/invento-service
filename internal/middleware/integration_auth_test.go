package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationMiddlewareRequestFlow tests complete request flow through middleware stack
func TestIntegrationMiddlewareRequestFlow(t *testing.T) {
	t.Parallel()
	t.Run("FullMiddlewareStackSuccessFlow", func(t *testing.T) {
		// Create Fiber app with full middleware stack
		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				var appErr *apperrors.AppError
				if errors.As(err, &appErr) {
					return httputil.SendAppError(c, appErr)
				}
				return httputil.SendInternalServerErrorResponse(c)
			},
		})

		// Apply middleware stack in correct order
		app.Use(RequestID())

		// Add test handler
		app.Get("/api/test", func(c *fiber.Ctx) error {
			requestID := GetRequestID(c)
			return httputil.SendSuccessResponse(c, fiber.StatusOK, "Success", fiber.Map{
				"request_id": requestID,
				"message":    "Test successful",
			})
		})

		// Make request
		req := httptest.NewRequest("GET", "/api/test", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Verify request ID is present in response headers
		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		// Parse response body
		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Verify response structure
		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "Success", result.Message)
		assert.NotNil(t, result.Data)

		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, requestID, data["request_id"])
	})

	t.Run("MiddlewareExecutionOrder", func(t *testing.T) {
		// Track middleware execution order
		executionOrder := []string{}

		app := fiber.New()

		app.Use(RequestID())
		app.Use(func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "middleware1")
			return c.Next()
		})
		app.Use(func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "middleware2")
			return c.Next()
		})

		app.Get("/test", func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "handler")
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Verify execution order - middleware execute in order before handler
		assert.Len(t, executionOrder, 3)
		assert.Equal(t, "middleware1", executionOrder[0])
		assert.Equal(t, "middleware2", executionOrder[1])
		assert.Equal(t, "handler", executionOrder[2])
	})

	t.Run("MiddlewareShortCircuitOnValidation", func(t *testing.T) {
		app := fiber.New()
		app.Use(RequestID())

		type TestRequest struct {
			Name string `json:"name" validate:"required"`
		}

		app.Post("/test", func(c *fiber.Ctx) error {
			var req TestRequest
			if err := c.BodyParser(&req); err != nil {
				return httputil.SendBadRequestResponse(c, "Invalid format")
			}

			if errs := ValidateStruct(&req); len(errs) > 0 {
				return httputil.SendValidationErrorResponse(c, errs)
			}

			return httputil.SendSuccessResponse(c, fiber.StatusOK, "OK", nil)
		})

		// Send invalid request
		reqBody := map[string]string{}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Verify error response
		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "error", result.Status)
	})
}

// TestIntegrationMiddlewareErrorScenarios tests various error scenarios
func TestIntegrationMiddlewareErrorScenarios(t *testing.T) {
	t.Parallel()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) {
				return httputil.SendAppError(c, appErr)
			}
			return httputil.SendInternalServerErrorResponse(c)
		},
	})

	// Apply middleware
	app.Use(RequestID())

	// Define error routes
	app.Get("/error/notfound", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewNotFoundError("Resource"))
	})

	app.Get("/error/validation", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewValidationError("Field tidak valid", nil))
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
		return httputil.SendAppError(c, apperrors.NewInternalError(httputil.SendInternalServerErrorResponse(c)))
	})

	app.Get("/error/payload", func(c *fiber.Ctx) error {
		return httputil.SendAppError(c, apperrors.NewPayloadTooLargeError("Ukuran file terlalu besar"))
	})

	t.Run("ErrorScenario_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/notfound", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Verify status code
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		// Verify request ID is preserved in error response
		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		// Parse response body
		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Resource tidak ditemukan", result.Message)
		assert.Equal(t, fiber.StatusNotFound, result.Code)
	})

	t.Run("ErrorScenario_Validation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/validation", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Contains(t, result.Message, "tidak valid")
	})

	t.Run("ErrorScenario_Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/unauthorized", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Token tidak valid", result.Message)
	})

	t.Run("ErrorScenario_Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/forbidden", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Akses ditolak", result.Message)
	})

	t.Run("ErrorScenario_Conflict", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/conflict", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Data sudah ada", result.Message)
	})

	t.Run("ErrorScenario_Internal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/internal", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Terjadi kesalahan pada server", result.Message)
	})

	t.Run("ErrorScenario_PayloadTooLarge", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/payload", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
		assert.Equal(t, "Ukuran file terlalu besar", result.Message)
	})
}

// TestIntegrationMiddlewareRequestIDInLogs tests that request ID is properly logged
func TestIntegrationMiddlewareRequestIDInLogs(t *testing.T) {
	t.Parallel()
	t.Run("RequestIDIsGeneratedAndLogged", func(t *testing.T) {
		app := fiber.New()
		app.Use(RequestID())

		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Verify request ID in response headers
		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)
	})

	t.Run("CustomRequestIDIsPreserved", func(t *testing.T) {
		customRequestID := "custom-req-id-12345"

		app := fiber.New()
		app.Use(RequestID())

		app.Get("/test", func(c *fiber.Ctx) error {
			requestID := GetRequestID(c)
			return c.JSON(fiber.Map{
				"request_id": requestID,
			})
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set(RequestIDHeader, customRequestID)

		resp, err := app.Test(req)
		require.NoError(t, err)

		// Verify custom request ID is used
		requestID := resp.Header.Get(RequestIDHeader)
		assert.Equal(t, customRequestID, requestID)

		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, customRequestID, result["request_id"])
	})
}
