package httputil

import (
	"invento-service/internal/domain"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestSendSuccessResponse_Success tests successful success response
func TestSendSuccessResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendSuccessResponse(c, 200, "Berhasil", map[string]string{"key": "value"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestSendErrorResponse_Success tests error response
func TestSendErrorResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendErrorResponse(c, 400, "Terjadi kesalahan", nil)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestSendListResponse_Success tests list response
func TestSendListResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		data := []string{"item1", "item2"}
		pagination := domain.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 2,
			TotalPages: 1,
		}
		return SendListResponse(c, 200, "Berhasil mengambil data", data, pagination)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestSendValidationErrorResponse_Success tests validation error response
func TestSendValidationErrorResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		// Note: ValidationError is in domain package
		return SendErrorResponse(c, 400, "Data validasi tidak valid", []fiber.Map{
			{"field": "email", "message": "Email tidak valid"},
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestSendBadRequestResponse_Success tests bad request response
func TestSendBadRequestResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendBadRequestResponse(c, "Request tidak valid")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestSendBadRequestResponse_DefaultMessage tests bad request with default message
func TestSendBadRequestResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendBadRequestResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestSendUnauthorizedResponse_Success tests unauthorized response
func TestSendUnauthorizedResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendUnauthorizedResponse(c)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestSendForbiddenResponse_Success tests forbidden response
func TestSendForbiddenResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendForbiddenResponse(c)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

// TestSendNotFoundResponse_Success tests not found response
func TestSendNotFoundResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendNotFoundResponse(c, "Data tidak ditemukan")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestSendNotFoundResponse_DefaultMessage tests not found with default message
func TestSendNotFoundResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendNotFoundResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestSendConflictResponse_Success tests conflict response
func TestSendConflictResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendConflictResponse(c, "Data sudah ada")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
}

// TestSendConflictResponse_DefaultMessage tests conflict with default message
func TestSendConflictResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendConflictResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
}

// TestSendPayloadTooLargeResponse_Success tests payload too large response
func TestSendPayloadTooLargeResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendPayloadTooLargeResponse(c, "Ukuran file terlalu besar")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 413, resp.StatusCode)
}

// TestSendPayloadTooLargeResponse_DefaultMessage tests payload too large with default
func TestSendPayloadTooLargeResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendPayloadTooLargeResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 413, resp.StatusCode)
}

// TestSendTooManyRequestsResponse_Success tests too many requests response
func TestSendTooManyRequestsResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTooManyRequestsResponse(c, "Terlalu banyak permintaan")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

// TestSendTooManyRequestsResponse_DefaultMessage tests too many requests with default
func TestSendTooManyRequestsResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTooManyRequestsResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

// TestSendInternalServerErrorResponse_Success tests internal server error response
func TestSendInternalServerErrorResponse_Success(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendInternalServerErrorResponse(c)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}
