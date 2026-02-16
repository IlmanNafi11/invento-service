package http_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	httpcontroller "invento-service/internal/controller/http"
	dto "invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	app_testing "invento-service/internal/testing"
	"invento-service/internal/upload"
)

func TestGetUploadInfo_Forbidden(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-456", "other@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/tus/upload/:id", controller.GetUploadInfo)

	mockUC.On("GetUploadInfo", mock.Anything, "test-upload-id", "user-456").Return(nil, apperrors.NewForbiddenError("tidak memiliki akses ke upload ini"))

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/upload/test-upload-id", nil, "")

	app_testing.AssertError(t, resp, 403)

	mockUC.AssertExpectations(t)
}

// TestCancelUpload_Success tests DELETE for cancellation
func TestCancelUpload_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/tus/upload/:id", controller.CancelUpload)

	mockUC.On("CancelUpload", mock.Anything, "test-upload-id", "user-123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/tus/upload/test-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelUpload_AlreadyCompleted tests cancellation of completed upload
func TestCancelUpload_AlreadyCompleted(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/tus/upload/:id", controller.CancelUpload)

	mockUC.On("CancelUpload", mock.Anything, "test-upload-id", "user-123").Return(apperrors.NewTusCompletedError())

	req := httptest.NewRequest("DELETE", "/api/v1/tus/upload/test-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// ============================================================================
// TUS Controller Tests - Project Update Upload
// ============================================================================

// TestInitiateProjectUpdateUpload_Success tests project update upload initiation
func TestInitiateProjectUpdateUpload_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/project/:id/upload", controller.InitiateProjectUpdateUpload)

	expectedResponse := &dto.TusUploadResponse{
		UploadID:  "test-update-upload-id",
		UploadURL: "/project/1/update/test-update-upload-id",
		Offset:    0,
		Length:    1048576,
	}

	metadata := dto.TusUploadInitRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	mockUC.On("InitiateProjectUpdateUpload", mock.Anything, uint(1), "user-123", int64(1048576), metadata).Return(expectedResponse, nil)

	metadataHeader := encodeTusMetadata(map[string]string{
		"nama_project": "Updated Project",
		"kategori":     "mobile",
		"semester":     "2",
	})

	req := httptest.NewRequest("POST", "/api/v1/tus/project/1/upload", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "1048576")
	req.Header.Set("Upload-Metadata", metadataHeader)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	// Verify TUS headers
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Contains(t, resp.Header.Get("Location"), "test-update-upload-id")

	mockUC.AssertExpectations(t)
}

// TestInitiateProjectUpdateUpload_ProjectNotFound tests non-existent project
func TestInitiateProjectUpdateUpload_ProjectNotFound(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/project/:id/upload", controller.InitiateProjectUpdateUpload)

	metadata := dto.TusUploadInitRequest{}
	mockUC.On("InitiateProjectUpdateUpload", mock.Anything, uint(999), "user-123", int64(1048576), metadata).Return(nil, apperrors.NewNotFoundError("project"))

	req := httptest.NewRequest("POST", "/api/v1/tus/project/999/upload", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "1048576")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestUploadProjectUpdateChunk_Success tests project chunk upload
func TestUploadProjectUpdateChunk_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/tus/project/:id/upload/:upload_id", controller.UploadProjectUpdateChunk)

	chunkData := []byte("test chunk data")
	mockUC.On("HandleProjectUpdateChunk", mock.Anything, uint(1), "test-update-upload-id", "user-123", int64(0), mock.Anything).Return(int64(len(chunkData)), nil)

	req := httptest.NewRequest("PATCH", "/api/v1/tus/project/1/upload/test-update-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", upload.TusContentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	// Verify TUS headers
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))

	mockUC.AssertExpectations(t)
}

// TestUploadProjectUpdateChunk_ProjectMismatch tests project ID mismatch
func TestUploadProjectUpdateChunk_ProjectMismatch(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/tus/project/:id/upload/:upload_id", controller.UploadProjectUpdateChunk)

	mockUC.On("HandleProjectUpdateChunk", mock.Anything, uint(1), "test-update-upload-id", "user-123", int64(0), mock.Anything).Return(int64(0), apperrors.NewValidationError("project ID tidak cocok", nil))

	chunkData := []byte("test chunk data")
	req := httptest.NewRequest("PATCH", "/api/v1/tus/project/1/upload/test-update-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", upload.TusContentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestGetProjectUpdateUploadStatus_Success tests project update upload status
func TestGetProjectUpdateUploadStatus_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Head("/api/v1/tus/project/:id/upload/:upload_id", controller.GetProjectUpdateUploadStatus)

	mockUC.On("GetProjectUpdateUploadStatus", mock.Anything, uint(1), "test-update-upload-id", "user-123").Return(int64(524288), int64(1048576), nil)

	req := httptest.NewRequest("HEAD", "/api/v1/tus/project/1/upload/test-update-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify TUS headers
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, "524288", resp.Header.Get("Upload-Offset"))
	assert.Equal(t, "1048576", resp.Header.Get("Upload-Length"))

	mockUC.AssertExpectations(t)
}

// ============================================================================
// TUS Controller Tests - Helper Endpoints
// ============================================================================

// TestCheckUploadSlot_Success tests upload slot availability check
func TestCheckUploadSlot_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/tus/slot", controller.CheckUploadSlot)

	expectedSlot := &dto.TusUploadSlotResponse{
		Available:     true,
		Message:       "Slot tersedia",
		QueueLength:   0,
		ActiveUpload:  false,
		MaxConcurrent: 3,
	}

	mockUC.On("CheckUploadSlot", mock.Anything, "user-123").Return(expectedSlot, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/slot", nil, "")

	app_testing.AssertSuccess(t, resp)
	app_testing.AssertJSONField(t, resp, "data.available", true)

	mockUC.AssertExpectations(t)
}

// TestResetUploadQueue_Success tests queue reset
func TestResetUploadQueue_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/reset", controller.ResetUploadQueue)

	mockUC.On("ResetUploadQueue", mock.Anything, "user-123").Return(nil)

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/tus/reset", nil, "")

	app_testing.AssertSuccess(t, resp)

	mockUC.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_Success tests GET for project update upload metadata
