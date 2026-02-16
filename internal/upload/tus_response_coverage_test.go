package upload

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestSendTusInitiateResponse_Success tests TUS initiate response
func TestSendTusInitiateResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		return SendTusInitiateResponse(c, "upload123", "http://example.com/upload/upload123", 1024000)
	})

	req, _ := http.NewRequest("POST", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Upload-Offset"))
	assert.NotEmpty(t, resp.Header.Get("Upload-Length"))
	assert.NotEmpty(t, resp.Header.Get("Location"))
}

// TestSendTusChunkResponse_Success tests TUS chunk response
func TestSendTusChunkResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Patch("/test", func(c *fiber.Ctx) error {
		return SendTusChunkResponse(c, 1024)
	})

	req, _ := http.NewRequest("PATCH", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
	assert.Equal(t, "1024", resp.Header.Get("Upload-Offset"))
}

// TestSendTusHeadResponse_Success tests TUS HEAD response
func TestSendTusHeadResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Head("/test", func(c *fiber.Ctx) error {
		return SendTusHeadResponse(c, 2048, 1024000)
	})

	req, _ := http.NewRequest("HEAD", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "2048", resp.Header.Get("Upload-Offset"))
	assert.Equal(t, "1024000", resp.Header.Get("Upload-Length"))
}

// TestSendTusDeleteResponse_Success tests TUS delete response
func TestSendTusDeleteResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Delete("/test", SendTusDeleteResponse)

	req, _ := http.NewRequest("DELETE", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

// TestSendTusSlotResponse_Success tests TUS slot response
func TestSendTusSlotResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusSlotResponse(c, true, "Slot tersedia", 0, 0, 3)
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestSendTusSlotResponse_Unavailable tests TUS slot unavailable
func TestSendTusSlotResponse_Unavailable(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusSlotResponse(c, false, "Slot penuh", 5, 1, 3)
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestSendTusSlotResponse_ModulShape(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusSlotResponse(c, true, "Slot tersedia", 0, 0, 10)
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestSendTusErrorResponse_Success tests TUS error response
func TestSendTusErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusErrorResponse(c, 400, "1.0.0")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
}

// TestSendTusErrorResponseWithOffset_Success tests TUS error response with offset
func TestSendTusErrorResponseWithOffset_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusErrorResponseWithOffset(c, 409, "1.0.0", 512)
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Equal(t, "512", resp.Header.Get("Upload-Offset"))
}

// TestSendTusErrorResponseWithLength_Success tests TUS error response with length
func TestSendTusErrorResponseWithLength_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusErrorResponseWithLength(c, 413, "1.0.0", 1024000)
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 413, resp.StatusCode)
	assert.Equal(t, "1024000", resp.Header.Get("Upload-Length"))
}

// TestSendTusValidationErrorResponse_Success tests TUS validation error
func TestSendTusValidationErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusValidationErrorResponse(c, "Metadata tidak valid")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestSendTusNotFoundErrorResponse_Success tests TUS not found error
func TestSendTusNotFoundErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusNotFoundErrorResponse(c, "Upload tidak ditemukan")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestSendTusNotFoundErrorResponse_DefaultMessage tests TUS not found with default
func TestSendTusNotFoundErrorResponse_DefaultMessage(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusNotFoundErrorResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestSendTusForbiddenErrorResponse_Success tests TUS forbidden error
func TestSendTusForbiddenErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", SendTusForbiddenErrorResponse)

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

// TestSendTusConflictErrorResponse_Success tests TUS conflict error
func TestSendTusConflictErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusConflictErrorResponse(c, "Upload sedang berlangsung")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
}

// TestSendTusConflictErrorResponse_DefaultMessage tests TUS conflict with default
func TestSendTusConflictErrorResponse_DefaultMessage(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusConflictErrorResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
}

// TestSendTusPayloadTooLargeErrorResponse_Success tests TUS payload too large
func TestSendTusPayloadTooLargeErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusPayloadTooLargeErrorResponse(c, "File terlalu besar")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 413, resp.StatusCode)
}

// TestSendTusPayloadTooLargeErrorResponse_DefaultMessage tests default payload error
func TestSendTusPayloadTooLargeErrorResponse_DefaultMessage(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusPayloadTooLargeErrorResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 413, resp.StatusCode)
}

// TestSendTusTooManyRequestsErrorResponse_Success tests TUS too many requests
func TestSendTusTooManyRequestsErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusTooManyRequestsErrorResponse(c, "Terlalu banyak request")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

// TestSendTusTooManyRequestsErrorResponse_DefaultMessage tests default too many requests
func TestSendTusTooManyRequestsErrorResponse_DefaultMessage(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SendTusTooManyRequestsErrorResponse(c, "")
	})

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
}

// TestSendTusInternalErrorResponse_Success tests TUS internal error
func TestSendTusInternalErrorResponse_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test", SendTusInternalErrorResponse)

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}
