package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

// TestIntegrationMiddlewareValidationIntegration tests validation middleware integration
func TestIntegrationMiddlewareValidationIntegration(t *testing.T) {
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

	app.Use(RequestID())

	type CreateUserRequest struct {
		Name     string `json:"name" validate:"required,min=3"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	app.Post("/users", func(c *fiber.Ctx) error {
		var req CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return httputil.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := ValidateStruct(&req); len(errs) > 0 {
			return httputil.SendValidationErrorResponse(c, errs)
		}

		return httputil.SendSuccessResponse(c, fiber.StatusCreated, "User berhasil dibuat", nil)
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

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "error", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "error", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "error", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "error", result.Status)
	})
}

// TestIntegrationMiddlewareWithRealWorldScenarios tests realistic API scenarios
func TestIntegrationMiddlewareWithRealWorldScenarios(t *testing.T) {
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

	app.Use(RequestID())

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
		return httputil.SendSuccessResponse(c, fiber.StatusOK, "Items retrieved", itemList)
	})

	app.Get("/items/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var itemID uint
		if _, err := fmt.Sscanf(id, "%d", &itemID); err != nil {
			return httputil.SendBadRequestResponse(c, "Invalid ID")
		}

		item, exists := items[itemID]
		if !exists {
			return httputil.SendNotFoundResponse(c, "Item")
		}

		return httputil.SendSuccessResponse(c, fiber.StatusOK, "Item found", item)
	})

	app.Post("/items", func(c *fiber.Ctx) error {
		var req CreateItemRequest
		if err := c.BodyParser(&req); err != nil {
			return httputil.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := ValidateStruct(&req); len(errs) > 0 {
			return httputil.SendValidationErrorResponse(c, errs)
		}

		item := Item{ID: nextID, Name: req.Name}
		items[nextID] = item
		nextID++

		return httputil.SendSuccessResponse(c, fiber.StatusCreated, "Item created", item)
	})

	app.Put("/items/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var itemID uint
		if _, err := fmt.Sscanf(id, "%d", &itemID); err != nil {
			return httputil.SendBadRequestResponse(c, "Invalid ID")
		}

		_, exists := items[itemID]
		if !exists {
			return httputil.SendNotFoundResponse(c, "Item")
		}

		var req UpdateItemRequest
		if err := c.BodyParser(&req); err != nil {
			return httputil.SendBadRequestResponse(c, "Format request tidak valid")
		}

		if errs := ValidateStruct(&req); len(errs) > 0 {
			return httputil.SendValidationErrorResponse(c, errs)
		}

		item := Item{ID: itemID, Name: req.Name}
		items[itemID] = item

		return httputil.SendSuccessResponse(c, fiber.StatusOK, "Item updated", item)
	})

	app.Delete("/items/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var itemID uint
		if _, err := fmt.Sscanf(id, "%d", &itemID); err != nil {
			return httputil.SendBadRequestResponse(c, "Invalid ID")
		}

		if _, exists := items[itemID]; !exists {
			return httputil.SendNotFoundResponse(c, "Item")
		}

		delete(items, itemID)
		return httputil.SendSuccessResponse(c, fiber.StatusOK, "Item deleted", nil)
	})

	t.Run("Scenario_ListItems", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		requestID := resp.Header.Get(RequestIDHeader)
		assert.NotEmpty(t, requestID)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
	})

	t.Run("Scenario_GetItem_Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items/1", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
	})

	t.Run("Scenario_GetItem_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/items/999", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "error", result.Status)
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

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
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

		var result dto.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "error", result.Status)
	})

	t.Run("Scenario_UpdateItem_Success", func(t *testing.T) {
		reqBody := UpdateItemRequest{Name: "Updated Item"}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/items/1", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "Item updated", result.Message)
	})

	t.Run("Scenario_DeleteItem_Success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/items/2", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result dto.SuccessResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "Item deleted", result.Message)
	})
}
