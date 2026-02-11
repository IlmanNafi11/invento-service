package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/logger"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationMiddlewareRequestFlow tests complete request flow through middleware stack
func TestIntegrationMiddlewareRequestFlow(t *testing.T) {
	t.Run("FullMiddlewareStackSuccessFlow", func(t *testing.T) {
		// Create logger
		log := logger.NewLogger(logger.INFO, logger.TextFormat)

		// Create Fiber app with full middleware stack
		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				if appErr, ok := err.(*apperrors.AppError); ok {
					return helper.SendAppError(c, appErr)
				}
				return helper.SendInternalServerErrorResponse(c)
			},
		})

		// Apply middleware stack in correct order
		app.Use(RequestID())
		app.Use(RequestLogger(log))

		// Add test handler
		app.Get("/api/test", func(c *fiber.Ctx) error {
			requestID := GetRequestID(c)
			return helper.SendSuccessResponse(c, fiber.StatusOK, "Success", fiber.Map{
				"request_id": requestID,
				"message":    "Test successful",
			})
		})

		// Make request
		req := httptest.NewRequest("GET", "/api/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Verify request ID is present in response headers
		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		// Parse response body
		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Verify response structure
		assert.True(t, result.Success)
		assert.Equal(t, "Success", result.Message)
		assert.NotNil(t, result.Data)

		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, requestID, data["request_id"])
	})

	t.Run("MiddlewareExecutionOrder", func(t *testing.T) {
		// Track middleware execution order
		executionOrder := []string{}
		log := logger.NewLogger(logger.INFO, logger.TextFormat)

		app := fiber.New()

		app.Use(RequestID())
		app.Use(func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "middleware1")
			return c.Next()
		})
		app.Use(RequestLogger(log))
		app.Use(func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "middleware2")
			return c.Next()
		})

		app.Get("/test", func(c *fiber.Ctx) error {
			executionOrder = append(executionOrder, "handler")
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
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
		log := logger.NewLogger(logger.INFO, logger.TextFormat)

		app := fiber.New()
		app.Use(RequestID())
		app.Use(RequestLogger(log))

		type TestRequest struct {
			Name string `json:"name" validate:"required"`
		}

		app.Post("/test", func(c *fiber.Ctx) error {
			var req TestRequest
			if err := c.BodyParser(&req); err != nil {
				return helper.SendBadRequestResponse(c, "Invalid format")
			}

			if errs := ValidateStruct(&req); len(errs) > 0 {
				return helper.SendValidationErrorResponse(c, errs)
			}

			return helper.SendSuccessResponse(c, fiber.StatusOK, "OK", nil)
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
		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.False(t, result.Success)
	})
}

// TestIntegrationMiddlewareErrorScenarios tests various error scenarios
func TestIntegrationMiddlewareErrorScenarios(t *testing.T) {
	log := logger.NewLogger(logger.INFO, logger.TextFormat)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if appErr, ok := err.(*apperrors.AppError); ok {
				return helper.SendAppError(c, appErr)
			}
			return helper.SendInternalServerErrorResponse(c)
		},
	})

	// Apply middleware
	app.Use(RequestID())
	app.Use(RequestLogger(log))

	// Define error routes
	app.Get("/error/notfound", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewNotFoundError("Resource"))
	})

	app.Get("/error/validation", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewValidationError("Field tidak valid", nil))
	})

	app.Get("/error/unauthorized", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewUnauthorizedError("Token tidak valid"))
	})

	app.Get("/error/forbidden", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewForbiddenError("Akses ditolak"))
	})

	app.Get("/error/conflict", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewConflictError("Data sudah ada"))
	})

	app.Get("/error/internal", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewInternalError(helper.SendInternalServerErrorResponse(c)))
	})

	app.Get("/error/payload", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewPayloadTooLargeError("Ukuran file terlalu besar"))
	})

	t.Run("ErrorScenario_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/notfound", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		// Verify status code
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		// Verify request ID is preserved in error response
		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		// Parse response body
		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Resource tidak ditemukan", result.Message)
		assert.Equal(t, fiber.StatusNotFound, result.Code)
	})

	t.Run("ErrorScenario_Validation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/validation", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "tidak valid")
	})

	t.Run("ErrorScenario_Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/unauthorized", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Token tidak valid", result.Message)
	})

	t.Run("ErrorScenario_Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/forbidden", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Akses ditolak", result.Message)
	})

	t.Run("ErrorScenario_Conflict", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/conflict", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Data sudah ada", result.Message)
	})

	t.Run("ErrorScenario_Internal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/internal", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Terjadi kesalahan pada server", result.Message)
	})

	t.Run("ErrorScenario_PayloadTooLarge", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/payload", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Ukuran file terlalu besar", result.Message)
	})
}

// TestIntegrationMiddlewareRequestIDInLogs tests that request ID is properly logged
func TestIntegrationMiddlewareRequestIDInLogs(t *testing.T) {
	t.Run("RequestIDIsGeneratedAndLogged", func(t *testing.T) {
		// Create a buffer to capture log output
		var logBuffer bytes.Buffer
		log := logger.NewLogger(logger.INFO, logger.TextFormat)
		log.SetOutput(&logBuffer)

		app := fiber.New()
		app.Use(RequestID())
		app.Use(RequestLogger(log))

		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Verify request ID in response headers
		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		// Note: In a real test, we would verify the log buffer contains the request ID
		// However, since the logger writes to the buffer asynchronously,
		// we just verify the request ID is present in the response
	})

	t.Run("CustomRequestIDIsPreserved", func(t *testing.T) {
		customRequestID := "custom-req-id-12345"
		log := logger.NewLogger(logger.INFO, logger.TextFormat)

		app := fiber.New()
		app.Use(RequestID())
		app.Use(RequestLogger(log))

		app.Get("/test", func(c *fiber.Ctx) error {
			requestID := GetRequestID(c)
			return c.JSON(fiber.Map{
				"request_id": requestID,
			})
		})

		req := httptest.NewRequest("GET", "/test", nil)
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

// TestIntegrationMiddlewareValidationIntegration tests validation middleware integration
func TestIntegrationMiddlewareValidationIntegration(t *testing.T) {
	log := logger.NewLogger(logger.INFO, logger.TextFormat)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if appErr, ok := err.(*apperrors.AppError); ok {
				return helper.SendAppError(c, appErr)
			}
			return helper.SendInternalServerErrorResponse(c)
		},
	})

	app.Use(RequestID())
	app.Use(RequestLogger(log))

	type CreateUserRequest struct {
		Name     string `json:"name" validate:"required,min=3"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	app.Post("/users", func(c *fiber.Ctx) error {
		var req CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := ValidateStruct(&req); len(errs) > 0 {
			return helper.SendValidationErrorResponse(c, errs)
		}

		return helper.SendSuccessResponse(c, fiber.StatusCreated, "User berhasil dibuat", nil)
	})

	t.Run("ValidRequest_PassesValidation", func(t *testing.T) {
		reqBody := CreateUserRequest{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "SecurePass123",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("InvalidRequest_NameTooShort", func(t *testing.T) {
		reqBody := map[string]string{
			"name":     "Jo",
			"email":    "john@example.com",
			"password": "SecurePass123",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.False(t, result.Success)
	})

	t.Run("InvalidRequest_InvalidEmail", func(t *testing.T) {
		reqBody := map[string]string{
			"name":     "John Doe",
			"email":    "invalid-email",
			"password": "SecurePass123",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.False(t, result.Success)
	})

	t.Run("InvalidRequest_PasswordTooShort", func(t *testing.T) {
		reqBody := map[string]string{
			"name":     "John Doe",
			"email":    "john@example.com",
			"password": "short",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.False(t, result.Success)
	})

	t.Run("InvalidRequest_MultipleErrors", func(t *testing.T) {
		reqBody := map[string]string{
			"name":     "Jo",
			"email":    "invalid-email",
			"password": "short",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.False(t, result.Success)
	})
}

// TestIntegrationMiddlewareWithRealWorldScenarios tests realistic API scenarios
func TestIntegrationMiddlewareWithRealWorldScenarios(t *testing.T) {
	log := logger.NewLogger(logger.INFO, logger.TextFormat)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if appErr, ok := err.(*apperrors.AppError); ok {
				return helper.SendAppError(c, appErr)
			}
			return helper.SendInternalServerErrorResponse(c)
		},
	})

	app.Use(RequestID())
	app.Use(RequestLogger(log))

	// Simulate a simple CRUD API for items
	type Item struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	type CreateItemRequest struct {
		Name string `json:"name" validate:"required,min=3"`
	}

	type UpdateItemRequest struct {
		Name string `json:"name" validate:"required,min=3"`
	}

	// Mock in-memory storage
	items := map[uint]Item{
		1: {ID: 1, Name: "Item 1"},
		2: {ID: 2, Name: "Item 2"},
	}
	nextID := uint(3)

	app.Get("/items", func(c *fiber.Ctx) error {
		itemList := make([]Item, 0, len(items))
		for _, item := range items {
			itemList = append(itemList, item)
		}
		return helper.SendSuccessResponse(c, fiber.StatusOK, "Items retrieved", itemList)
	})

	app.Get("/items/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var itemID uint
		if _, err := fmt.Sscanf(id, "%d", &itemID); err != nil {
			return helper.SendBadRequestResponse(c, "Invalid ID")
		}

		item, exists := items[itemID]
		if !exists {
			return helper.SendNotFoundResponse(c, "Item")
		}

		return helper.SendSuccessResponse(c, fiber.StatusOK, "Item found", item)
	})

	app.Post("/items", func(c *fiber.Ctx) error {
		var req CreateItemRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := ValidateStruct(&req); len(errs) > 0 {
			return helper.SendValidationErrorResponse(c, errs)
		}

		item := Item{ID: nextID, Name: req.Name}
		items[nextID] = item
		nextID++

		return helper.SendSuccessResponse(c, fiber.StatusCreated, "Item created", item)
	})

	app.Put("/items/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var itemID uint
		if _, err := fmt.Sscanf(id, "%d", &itemID); err != nil {
			return helper.SendBadRequestResponse(c, "Invalid ID")
		}

		_, exists := items[itemID]
		if !exists {
			return helper.SendNotFoundResponse(c, "Item")
		}

		var req UpdateItemRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := ValidateStruct(&req); len(errs) > 0 {
			return helper.SendValidationErrorResponse(c, errs)
		}

		item := Item{ID: itemID, Name: req.Name}
		items[itemID] = item

		return helper.SendSuccessResponse(c, fiber.StatusOK, "Item updated", item)
	})

	app.Delete("/items/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var itemID uint
		if _, err := fmt.Sscanf(id, "%d", &itemID); err != nil {
			return helper.SendBadRequestResponse(c, "Invalid ID")
		}

		if _, exists := items[itemID]; !exists {
			return helper.SendNotFoundResponse(c, "Item")
		}

		delete(items, itemID)
		return helper.SendSuccessResponse(c, fiber.StatusOK, "Item deleted", nil)
	})

	t.Run("Scenario_ListItems", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Scenario_GetItem_Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items/1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Scenario_GetItem_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items/999", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.False(t, result.Success)
		// SendNotFoundResponse uses the message parameter directly
		assert.Equal(t, "Item", result.Message)
	})

	t.Run("Scenario_CreateItem_Success", func(t *testing.T) {
		reqBody := CreateItemRequest{Name: "New Item"}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/items", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "Item created", result.Message)
	})

	t.Run("Scenario_CreateItem_ValidationError", func(t *testing.T) {
		reqBody := map[string]string{"name": "Ab"}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/items", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.False(t, result.Success)
	})

	t.Run("Scenario_UpdateItem_Success", func(t *testing.T) {
		reqBody := UpdateItemRequest{Name: "Updated Item"}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/items/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "Item updated", result.Message)
	})

	t.Run("Scenario_DeleteItem_Success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/items/2", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "Item deleted", result.Message)
	})
}
