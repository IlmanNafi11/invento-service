package internal

import (
	"bytes"
	"encoding/json"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/helper"
	"invento-service/internal/logger"
	"invento-service/internal/middleware"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationMiddlewareChain tests that all middleware work together correctly
func TestIntegrationMiddlewareChain(t *testing.T) {
	// Create logger for testing
	log := logger.NewLogger(logger.INFO, logger.TextFormat)

	// Create Fiber app with middleware chain
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"code":    code,
			})
		},
	})

	// Apply middleware chain: RequestID -> Logging
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(log))

	// Add test handler
	app.Get("/test", func(c *fiber.Ctx) error {
		requestID := middleware.GetRequestID(c)
		return c.JSON(fiber.Map{
			"success":    true,
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
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "Test successful", result["message"])
	assert.Equal(t, requestID, result["request_id"], "Request ID should match header")
}

// TestIntegrationErrorHandling tests error handling integration with AppError
func TestIntegrationErrorHandling(t *testing.T) {
	app := fiber.New()

	// Error handler that returns AppError
	app.Get("/error/notfound", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewNotFoundError("Test Resource"))
	})

	app.Get("/error/validation", func(c *fiber.Ctx) error {
		return helper.SendAppError(c, apperrors.NewValidationError("Field wajib diisi", nil))
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
		return helper.SendAppError(c, apperrors.NewInternalError(fmt.Errorf("database error")))
	})

	t.Run("NotFoundError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/notfound", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Test Resource tidak ditemukan", result.Message)
		assert.Equal(t, fiber.StatusNotFound, result.Code)
	})

	t.Run("ValidationError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/validation", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Field wajib diisi", result.Message)
		assert.Equal(t, fiber.StatusBadRequest, result.Code)
	})

	t.Run("UnauthorizedError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/unauthorized", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Token tidak valid", result.Message)
		assert.Equal(t, fiber.StatusUnauthorized, result.Code)
	})

	t.Run("ForbiddenError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/forbidden", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Akses ditolak", result.Message)
		assert.Equal(t, fiber.StatusForbidden, result.Code)
	})

	t.Run("ConflictError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/conflict", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Data sudah ada", result.Message)
		assert.Equal(t, fiber.StatusConflict, result.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/error/internal", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Equal(t, "Terjadi kesalahan pada server", result.Message)
		assert.Equal(t, fiber.StatusInternalServerError, result.Code)
	})
}

// TestIntegrationDTOMiddleware tests DTO validation integration with middleware
func TestIntegrationDTOMiddleware(t *testing.T) {
	// Create test request type
	type TestRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	app := fiber.New()

	// Apply validation middleware
	app.Post("/test", func(c *fiber.Ctx) error {
		var req TestRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.SendBadRequestResponse(c, "Format request tidak valid")
		}

		// Validate using middleware helper
		if errs := middleware.ValidateStruct(&req); len(errs) > 0 {
			return helper.SendValidationErrorResponse(c, errs)
		}

		return helper.SendSuccessResponse(c, fiber.StatusOK, "Data valid", req)
	})

	t.Run("ValidRequest", func(t *testing.T) {
		reqBody := TestRequest{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result.Success)
		assert.Equal(t, "Data valid", result.Message)
	})

	t.Run("InvalidRequest_MissingRequired", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "john@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "valid")
	})

	t.Run("InvalidRequest_InvalidEmail", func(t *testing.T) {
		reqBody := TestRequest{
			Name:  "John Doe",
			Email: "invalid-email",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
	})
}

// TestIntegrationPaginationWithMiddleware tests pagination DTO integration
func TestIntegrationPaginationWithMiddleware(t *testing.T) {
	app := fiber.New()

	app.Get("/items", func(c *fiber.Ctx) error {
		var pagReq dto.PaginationRequest
		if err := c.QueryParser(&pagReq); err != nil {
			return helper.SendBadRequestResponse(c, "Parameter tidak valid")
		}

		// Set defaults BEFORE validation
		if pagReq.Page == 0 {
			pagReq.Page = dto.DefaultPage
		}
		if pagReq.Limit == 0 {
			pagReq.Limit = dto.DefaultLimit
		}

		// Validate only Order if provided, skip Page/Limit validation since we set defaults
		// The DTO validation will pass now that Page and Limit have valid values

		// Calculate offset (used to skip items in real queries)
		offset := pagReq.GetOffset()
		_ = offset // Offset is calculated for potential use in database queries

		// Mock items
		items := []map[string]interface{}{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		}
		totalItems := 100

		totalPages := (totalItems + pagReq.Limit - 1) / pagReq.Limit

		paginationData := domain.PaginationData{
			Page:       pagReq.Page,
			Limit:      pagReq.Limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		}

		return helper.SendListResponse(c, fiber.StatusOK, "Items retrieved", items, paginationData)
	})

	t.Run("ValidPagination_Default", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result.Success)

		listData, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		pagination, ok := listData["pagination"].(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, float64(dto.DefaultPage), pagination["page"])
		assert.Equal(t, float64(dto.DefaultLimit), pagination["limit"])
		assert.Equal(t, float64(100), pagination["total_items"])
		assert.Equal(t, float64(10), pagination["total_pages"])
	})

	t.Run("ValidPagination_Custom", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items?page=2&limit=20", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result.Success)

		listData, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		pagination, ok := listData["pagination"].(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, float64(2), pagination["page"])
		assert.Equal(t, float64(20), pagination["limit"])
		assert.Equal(t, float64(5), pagination["total_pages"])
	})

	t.Run("Pagination_GetOffset", func(t *testing.T) {
		pagReq := dto.PaginationRequest{
			Page:  3,
			Limit: 10,
		}
		var offset = pagReq.GetOffset()
		assert.Equal(t, 20, offset, "Offset should be (3-1)*10 = 20")

		pagReq.Page = 1
		offset = pagReq.GetOffset()
		assert.Equal(t, 0, offset, "Offset for page 1 should be 0")
	})
}

// TestIntegrationCompleteRequestCycle tests a complete request/response cycle
func TestIntegrationCompleteRequestCycle(t *testing.T) {
	// Setup complete application stack
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
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(log))

	// Define request types
	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
	}

	type UpdateUserRequest struct {
		Name string `json:"name" validate:"required,min=3"`
	}

	// Routes
	app.Post("/users", func(c *fiber.Ctx) error {
		var req CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := middleware.ValidateStruct(&req); len(errs) > 0 {
			return helper.SendValidationErrorResponse(c, errs)
		}

		// Mock user creation
		user := map[string]interface{}{
			"id":    1,
			"name":  req.Name,
			"email": req.Email,
		}

		return helper.SendSuccessResponse(c, fiber.StatusCreated, "User berhasil dibuat", user)
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "999" {
			return helper.SendNotFoundResponse(c, "User")
		}

		user := map[string]interface{}{
			"id":    id,
			"name":  "John Doe",
			"email": "john@example.com",
		}

		return helper.SendSuccessResponse(c, fiber.StatusOK, "User ditemukan", user)
	})

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		var req UpdateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := middleware.ValidateStruct(&req); len(errs) > 0 {
			return helper.SendValidationErrorResponse(c, errs)
		}

		user := map[string]interface{}{
			"id":   c.Params("id"),
			"name": req.Name,
		}

		return helper.SendSuccessResponse(c, fiber.StatusOK, "User berhasil diupdate", user)
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "999" {
			return helper.SendNotFoundResponse(c, "User")
		}

		return helper.SendSuccessResponse(c, fiber.StatusOK, "User berhasil dihapus", nil)
	})

	t.Run("CreateUser_Success", func(t *testing.T) {
		reqBody := CreateUserRequest{
			Name:  "Jane Doe",
			Email: "jane@example.com",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		// Verify request ID in response
		requestID := resp.Header.Get(middleware.RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result.Success)
		assert.Equal(t, "User berhasil dibuat", result.Message)

		user, ok := result.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Jane Doe", user["name"])
		assert.Equal(t, "jane@example.com", user["email"])
	})

	t.Run("CreateUser_ValidationError", func(t *testing.T) {
		reqBody := map[string]string{
			"name": "Jo", // Too short
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

	t.Run("GetUser_Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result.Success)
		assert.Equal(t, "User ditemukan", result.Message)
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/999", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
		// SendNotFoundResponse uses the message parameter directly
		assert.Equal(t, "User", result.Message)
	})

	t.Run("UpdateUser_Success", func(t *testing.T) {
		reqBody := UpdateUserRequest{
			Name: "Updated Name",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/users/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result.Success)
		assert.Equal(t, "User berhasil diupdate", result.Message)
	})

	t.Run("DeleteUser_Success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result domain.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.True(t, result.Success)
		assert.Equal(t, "User berhasil dihapus", result.Message)
	})

	t.Run("DeleteUser_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/999", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result domain.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.False(t, result.Success)
	})
}

// TestIntegrationRequestIDPropagation tests that request ID is propagated through the entire request
func TestIntegrationRequestIDPropagation(t *testing.T) {
	log := logger.NewLogger(logger.INFO, logger.TextFormat)

	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(log))

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
