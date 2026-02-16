package upload_test

import (
	"net/http/httptest"
	"testing"
	"invento-service/internal/upload"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ==================== TusHeaders Tests ====================

func TestGetTusHeaders_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Set("Tus-Resumable", "1.0.0")
		c.Set("Upload-Offset", "100")
		c.Set("Upload-Length", "1000")
		c.Set("Upload-Metadata", "filename dGVzdA==")
		c.Set("Content-Type", "application/offset+octet-stream")
		c.Set("Content-Length", "50")

		headers, err := upload.GetTusHeaders(c)
		if err != nil {
			return err
		}

		return c.JSON(headers)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetTusHeaders_EmptyHeaders(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		headers, err := upload.GetTusHeaders(c)
		if err != nil {
			return err
		}

		return c.JSON(headers)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSetTusResponseHeaders_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		upload.SetTusResponseHeaders(c, 100, 1000)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, "100", resp.Header.Get("Upload-Offset"))
	assert.Equal(t, "1000", resp.Header.Get("Upload-Length"))
}

func TestSetTusResponseHeaders_ZeroLength(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		upload.SetTusResponseHeaders(c, 0, 0)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, "", resp.Header.Get("Upload-Length"))
}

func TestSetTusLocationHeader_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		upload.SetTusLocationHeader(c, "/uploads/upload-123")
		return c.SendStatus(fiber.StatusCreated)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "/uploads/upload-123", resp.Header.Get("Location"))
}

func TestSetTusOffsetHeader_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Patch("/test", func(c *fiber.Ctx) error {
		upload.SetTusOffsetHeader(c, 500)
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "500", resp.Header.Get("Upload-Offset"))
}

func TestValidateChunkSize_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"min size", 1, false},
		{"1MB", 1024 * 1024, false},
		{"2MB", upload.MaxChunkSize, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"too large", upload.MaxChunkSize + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := upload.ValidateChunkSize(tt.size)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildTusErrorResponse_ConflictWithOffset(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.BuildTusErrorResponse(c, fiber.StatusConflict, 100)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	assert.Equal(t, "100", resp.Header.Get("Upload-Offset"))
}

func TestBuildTusErrorResponse_OtherStatus(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.BuildTusErrorResponse(c, fiber.StatusBadRequest, 100)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get("Upload-Offset"))
}
