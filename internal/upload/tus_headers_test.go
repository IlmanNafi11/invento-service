package upload_test

import (
	"invento-service/internal/upload"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestGetTusHeaders_Basic(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	// Set some headers
	c.Request().Header.Set("Tus-Resumable", "1.0.0")
	c.Request().Header.Set("Upload-Offset", "1024")
	c.Request().Header.Set("Upload-Length", "1048576")
	c.Request().Header.Set("Upload-Metadata", "filename=test.txt")
	c.Request().Header.Set("Content-Type", "application/offset+octet-stream")
	c.Request().Header.Set("Content-Length", "4096")

	headers, err := upload.GetTusHeaders(c)
	assert.NoError(t, err)

	assert.Equal(t, "1.0.0", headers.TusResumable)
	assert.Equal(t, int64(1024), headers.UploadOffset)
	assert.Equal(t, int64(1048576), headers.UploadLength)
	assert.Equal(t, "filename=test.txt", headers.UploadMetadata)
	assert.Equal(t, "application/offset+octet-stream", headers.ContentType)
	assert.Equal(t, int64(4096), headers.ContentLength)
}

func TestGetTusHeaders_EmptyHeaders_Direct(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	headers, err := upload.GetTusHeaders(c)
	assert.NoError(t, err)

	assert.Empty(t, headers.TusResumable)
	assert.Equal(t, int64(0), headers.UploadOffset)
	assert.Equal(t, int64(0), headers.UploadLength)
	assert.Empty(t, headers.UploadMetadata)
	assert.Empty(t, headers.ContentType)
	assert.Equal(t, int64(0), headers.ContentLength)
}

func TestGetTusHeaders_InvalidNumbers(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	// Set invalid numeric headers
	c.Request().Header.Set("Upload-Offset", "invalid")
	c.Request().Header.Set("Upload-Length", "not-a-number")
	c.Request().Header.Set("Content-Length", "abc")

	_, err := upload.GetTusHeaders(c)
	assert.Error(t, err)
}

func TestTusHeaders_Constants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Tus-Resumable", upload.HeaderTusResumable)
	assert.Equal(t, "Upload-Offset", upload.HeaderUploadOffset)
	assert.Equal(t, "Upload-Length", upload.HeaderUploadLength)
	assert.Equal(t, "Upload-Metadata", upload.HeaderUploadMetadata)
	assert.Equal(t, "Content-Type", upload.HeaderContentType)
	assert.Equal(t, "Content-Length", upload.HeaderContentLength)
	assert.Equal(t, "Location", upload.HeaderLocation)

	assert.Equal(t, "1.0.0", upload.TusVersion)
	assert.Equal(t, "application/offset+octet-stream", upload.TusContentType)
	// Constants are int type, convert to int64 for comparison
	assert.Equal(t, int64(1048576), int64(upload.DefaultChunkSize))
	assert.Equal(t, int64(2097152), int64(upload.MaxChunkSize))
	assert.Equal(t, int64(524288000), int64(upload.MaxProjectFileSize))
	assert.Equal(t, int64(52428800), int64(upload.MaxModulFileSize))
}

func TestSetTusResponseHeaders(t *testing.T) {
	t.Parallel()
	// Just verify no panic - actual header testing requires full HTTP request/response
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	upload.SetTusResponseHeaders(c, 2048, 1048576)

	// Verify status code is set
	assert.Equal(t, fiber.StatusOK, c.Response().StatusCode())

	app.ReleaseCtx(c)
}

func TestSetTusResponseHeaders_ZeroLength_Direct(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	upload.SetTusResponseHeaders(c, 1024, 0)

	// Verify no panic occurs
	assert.Equal(t, fiber.StatusOK, c.Response().StatusCode())

	app.ReleaseCtx(c)
}

func TestSetTusLocationHeader(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	location := "/api/tus/upload-123"
	upload.SetTusLocationHeader(c, location)

	// Verify no panic occurs
	assert.Equal(t, fiber.StatusOK, c.Response().StatusCode())

	app.ReleaseCtx(c)
}

func TestSetTusOffsetHeader(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	upload.SetTusOffsetHeader(c, 4096)

	// Verify no panic occurs
	assert.Equal(t, fiber.StatusOK, c.Response().StatusCode())

	app.ReleaseCtx(c)
}

func TestValidateChunkSize_ValidSizes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		size int64
	}{
		{"Minimum valid size", 1},
		{"Default chunk size", upload.DefaultChunkSize},
		{"Max chunk size", upload.MaxChunkSize},
		{"Small chunk", 1024},
		{"Medium chunk", 512 * 1024},
		{"1MB chunk", 1024 * 1024},
		{"Just under max", upload.MaxChunkSize - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := upload.ValidateChunkSize(tt.size)
			assert.NoError(t, err)
		})
	}
}

func TestValidateChunkSize_InvalidSizes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		size        int64
		expectedErr string
	}{
		{"Zero size", 0, "ukuran chunk tidak valid"},
		{"Negative size", -1, "ukuran chunk tidak valid"},
		{"Over max size", upload.MaxChunkSize + 1, "2 MB"},
		{"Very large size", 10 * 1024 * 1024, "2 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := upload.ValidateChunkSize(tt.size)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestBuildTusErrorResponse_Conflict(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	offset := int64(1024)
	err := upload.BuildTusErrorResponse(c, fiber.StatusConflict, offset)

	// Verify no error and correct status code
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, c.Response().StatusCode())

	app.ReleaseCtx(c)
}

func TestBuildTusErrorResponse_OtherStatus_Direct(t *testing.T) {
	t.Parallel()
	// Note: BuildTusErrorResponse calls c.SendStatus() which returns early
	// The important thing is that no error is returned and status code is set
	t.Run("Bad Request", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})

		err := upload.BuildTusErrorResponse(c, fiber.StatusBadRequest, 0)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, c.Response().StatusCode())

		app.ReleaseCtx(c)
	})

	t.Run("Not Found", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})

		err := upload.BuildTusErrorResponse(c, fiber.StatusNotFound, 0)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, c.Response().StatusCode())

		app.ReleaseCtx(c)
	})

	t.Run("Payload Too Large", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})

		err := upload.BuildTusErrorResponse(c, fiber.StatusRequestEntityTooLarge, 0)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusRequestEntityTooLarge, c.Response().StatusCode())

		app.ReleaseCtx(c)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})

		err := upload.BuildTusErrorResponse(c, fiber.StatusInternalServerError, 0)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, c.Response().StatusCode())

		app.ReleaseCtx(c)
	})
}

func TestBuildTusErrorResponse_NegativeOffset(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	err := upload.BuildTusErrorResponse(c, fiber.StatusConflict, -1)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, c.Response().StatusCode())
	// Offset should not be set for negative offset
	assert.Empty(t, c.Get("Upload-Offset"))

	app.ReleaseCtx(c)
}

func TestGetTusHeaders_PartialHeaders(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	// Set only some headers
	c.Request().Header.Set("Upload-Offset", "512")
	c.Request().Header.Set("Content-Type", "application/octet-stream")

	headers, err := upload.GetTusHeaders(c)
	assert.NoError(t, err)

	assert.Empty(t, headers.TusResumable)
	assert.Equal(t, int64(512), headers.UploadOffset)
	assert.Equal(t, int64(0), headers.UploadLength)
	assert.Empty(t, headers.UploadMetadata)
	assert.Equal(t, "application/octet-stream", headers.ContentType)
	assert.Equal(t, int64(0), headers.ContentLength)
}

func TestValidateChunkSize_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		size        int64
		shouldError bool
	}{
		{"Exactly 1 byte", 1, false},
		{"Exactly max chunk size", upload.MaxChunkSize, false},
		{"One byte over max", upload.MaxChunkSize + 1, true},
		{"Max int64", 9223372036854775807, true},
		{"Min int64", -9223372036854775808, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := upload.ValidateChunkSize(tt.size)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
