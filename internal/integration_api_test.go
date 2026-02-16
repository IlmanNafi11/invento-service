package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"invento-service/internal/dto"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"

	apperrors "invento-service/internal/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationDTOMiddleware tests DTO validation integration with middleware
func TestIntegrationDTOMiddleware(t *testing.T) {
	t.Parallel()
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
			return httputil.SendBadRequestResponse(c, "Format request tidak valid")
		}

		// Validate using middleware helper
		if errs := middleware.ValidateStruct(&req); len(errs) > 0 {
			return httputil.SendValidationErrorResponse(c, errs)
		}

		return httputil.SendSuccessResponse(c, fiber.StatusOK, "Data valid", req)
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

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
	})
}

// TestIntegrationPaginationWithMiddleware tests pagination DTO integration
func TestIntegrationPaginationWithMiddleware(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/items", func(c *fiber.Ctx) error {
		var pagReq dto.PaginationRequest
		if err := c.QueryParser(&pagReq); err != nil {
			return httputil.SendBadRequestResponse(c, "Parameter tidak valid")
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

		paginationData := dto.PaginationData{
			Page:       pagReq.Page,
			Limit:      pagReq.Limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		}

		return httputil.SendListResponse(c, fiber.StatusOK, "Items retrieved", items, paginationData)
	})

	t.Run("ValidPagination_Default", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result.Status)

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
		req := httptest.NewRequest("GET", "/items?page=2&limit=20", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result.Status)

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
		offset := pagReq.GetOffset()
		assert.Equal(t, 20, offset, "Offset should be (3-1)*10 = 20")

		pagReq.Page = 1
		offset = pagReq.GetOffset()
		assert.Equal(t, 0, offset, "Offset for page 1 should be 0")
	})
}

// TestIntegrationCompleteRequestCycle tests a complete request/response cycle
func TestIntegrationCompleteRequestCycle(t *testing.T) {
	t.Parallel()
	// Setup complete application stack
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
	app.Use(middleware.RequestID())

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
			return httputil.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := middleware.ValidateStruct(&req); len(errs) > 0 {
			return httputil.SendValidationErrorResponse(c, errs)
		}

		// Mock user creation
		user := map[string]interface{}{
			"id":    1,
			"name":  req.Name,
			"email": req.Email,
		}

		return httputil.SendSuccessResponse(c, fiber.StatusCreated, "User berhasil dibuat", user)
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "999" {
			return httputil.SendNotFoundResponse(c, "User")
		}

		user := map[string]interface{}{
			"id":    id,
			"name":  "John Doe",
			"email": "john@example.com",
		}

		return httputil.SendSuccessResponse(c, fiber.StatusOK, "User ditemukan", user)
	})

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		var req UpdateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return httputil.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := middleware.ValidateStruct(&req); len(errs) > 0 {
			return httputil.SendValidationErrorResponse(c, errs)
		}

		user := map[string]interface{}{
			"id":   c.Params("id"),
			"name": req.Name,
		}

		return httputil.SendSuccessResponse(c, fiber.StatusOK, "User berhasil diupdate", user)
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "999" {
			return httputil.SendNotFoundResponse(c, "User")
		}

		return httputil.SendSuccessResponse(c, fiber.StatusOK, "User berhasil dihapus", nil)
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

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
	})

	t.Run("GetUser_Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/1", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "User ditemukan", result.Message)
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/999", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
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

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "User berhasil diupdate", result.Message)
	})

	t.Run("DeleteUser_Success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/1", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "User berhasil dihapus", result.Message)
	})

	t.Run("DeleteUser_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/999", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "error", result.Status)
	})
}
