package httputil_test

import (
	"invento-service/internal/dto"
	"invento-service/internal/httputil"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSendSuccessResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		data := map[string]interface{}{
			"id":   1,
			"name": "Test",
		}
		return httputil.SendSuccessResponse(c, fiber.StatusOK, "Data berhasil diambil", data)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSendErrorResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return httputil.SendErrorResponse(c, fiber.StatusBadRequest, "Request tidak valid", nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSendInternalServerErrorResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return httputil.SendInternalServerErrorResponse(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestSendUnauthorizedResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return httputil.SendUnauthorizedResponse(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSendForbiddenResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return httputil.SendForbiddenResponse(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSendNotFoundResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return httputil.SendNotFoundResponse(c, "Data tidak ditemukan")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestSendListResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		items := []map[string]interface{}{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		}
		pagination := dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 50,
			TotalPages: 5,
		}
		return httputil.SendListResponse(c, fiber.StatusOK, "Data berhasil diambil", items, pagination)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSendValidationErrorResponse(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		errors := []dto.ValidationError{
			{Field: "email", Message: "Email wajib diisi"},
			{Field: "password", Message: "Password minimal 8 karakter"},
		}
		return httputil.SendValidationErrorResponse(c, errors)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
