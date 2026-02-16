package http_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"invento-service/internal/domain"
	"invento-service/internal/upload"

	httpcontroller "invento-service/internal/controller/http"

	dto "invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	app_testing "invento-service/internal/testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProjectUpdateUploadInfo_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Get("/api/v1/project/:id/upload/:upload_id", controller.GetProjectUpdateUploadInfo)

	expectedInfo := &dto.TusUploadInfoResponse{
		UploadID:  "test-project-upload-id",
		Status:    domain.UploadStatusUploading,
		Progress:  50.0,
		Offset:    524288,
		Length:    1048576,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUC.On("GetProjectUpdateUploadInfo", mock.Anything, uint(1), "test-project-upload-id", "user-123").Return(expectedInfo, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/project/1/upload/test-project-upload-id", nil, "")

	app_testing.AssertSuccess(t, resp)

	mockUC.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_Unauthorized tests GET without authentication
func TestGetProjectUpdateUploadInfo_Unauthorized(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Get("/api/v1/project/:id/upload/:upload_id", controller.GetProjectUpdateUploadInfo)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/project/1/upload/test-project-upload-id", nil, "")

	// Should return error due to missing auth
	assert.Equal(t, 401, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_NotFound tests GET with non-existent upload
func TestGetProjectUpdateUploadInfo_NotFound(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Get("/api/v1/project/:id/upload/:upload_id", controller.GetProjectUpdateUploadInfo)

	mockUC.On("GetProjectUpdateUploadInfo", mock.Anything, uint(1), "nonexistent-upload-id", "user-123").Return(nil, apperrors.NewNotFoundError("Upload not found"))

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/project/1/upload/nonexistent-upload-id", nil, "")

	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_Success tests DELETE for project update upload cancellation
func TestCancelProjectUpdateUpload_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Delete("/api/v1/project/:id/upload/:upload_id", controller.CancelProjectUpdateUpload)

	mockUC.On("CancelProjectUpdateUpload", mock.Anything, uint(1), "test-project-upload-id", "user-123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/project/1/upload/test-project-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_Unauthorized tests DELETE without authentication
func TestCancelProjectUpdateUpload_Unauthorized(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Delete("/api/v1/project/:id/upload/:upload_id", controller.CancelProjectUpdateUpload)

	req := httptest.NewRequest("DELETE", "/api/v1/project/1/upload/test-project-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_NotFound tests DELETE with non-existent upload
func TestCancelProjectUpdateUpload_NotFound(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Delete("/api/v1/project/:id/upload/:upload_id", controller.CancelProjectUpdateUpload)

	mockUC.On("CancelProjectUpdateUpload", mock.Anything, uint(1), "nonexistent-upload-id", "user-123").Return(apperrors.NewNotFoundError("Upload not found"))

	req := httptest.NewRequest("DELETE", "/api/v1/project/1/upload/nonexistent-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

type MockTusModulControllerUsecase struct {
	mock.Mock
}

func (m *MockTusModulControllerUsecase) InitiateModulUpload(ctx context.Context, userID string, fileSize int64, uploadMetadata string) (*dto.TusModulUploadResponse, error) {
	args := m.Called(ctx, userID, fileSize, uploadMetadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusModulUploadResponse), args.Error(1)
}

func (m *MockTusModulControllerUsecase) HandleModulChunk(ctx context.Context, uploadID, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(ctx, uploadID, userID, offset, mock.Anything)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusModulControllerUsecase) GetModulUploadInfo(ctx context.Context, uploadID, userID string) (*dto.TusModulUploadInfoResponse, error) {
	args := m.Called(ctx, uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusModulUploadInfoResponse), args.Error(1)
}

func (m *MockTusModulControllerUsecase) GetModulUploadStatus(ctx context.Context, uploadID, userID string) (offset, length int64, err error) {
	args := m.Called(ctx, uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusModulControllerUsecase) CancelModulUpload(ctx context.Context, uploadID, userID string) error {
	args := m.Called(ctx, uploadID, userID)
	return args.Error(0)
}

func (m *MockTusModulControllerUsecase) CheckModulUploadSlot(ctx context.Context, userID string) (*dto.TusModulUploadSlotResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusModulUploadSlotResponse), args.Error(1)
}

func (m *MockTusModulControllerUsecase) InitiateModulUpdateUpload(ctx context.Context, modulID, userID string, fileSize int64, uploadMetadata string) (*dto.TusModulUploadResponse, error) {
	args := m.Called(ctx, modulID, userID, fileSize, uploadMetadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusModulUploadResponse), args.Error(1)
}

func (m *MockTusModulControllerUsecase) HandleModulUpdateChunk(ctx context.Context, modulID, uploadID, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(ctx, modulID, uploadID, userID, offset, mock.Anything)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusModulControllerUsecase) GetModulUpdateUploadStatus(ctx context.Context, modulID, uploadID, userID string) (offset, length int64, err error) {
	args := m.Called(ctx, modulID, uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusModulControllerUsecase) GetModulUpdateUploadInfo(ctx context.Context, modulID, uploadID, userID string) (*dto.TusModulUploadInfoResponse, error) {
	args := m.Called(ctx, modulID, uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusModulUploadInfoResponse), args.Error(1)
}

func (m *MockTusModulControllerUsecase) CancelModulUpdateUpload(ctx context.Context, modulID, uploadID, userID string) error {
	args := m.Called(ctx, modulID, uploadID, userID)
	return args.Error(0)
}

func TestTusModulController_InitiateUpload_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusModulControllerUsecase)
	baseCtrl := getTusBaseController()
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusModulController(mockUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Post("/api/v1/tus/modul/upload", controller.InitiateUpload)

	metadataHeader := encodeTusMetadata(map[string]string{
		"judul":     "Modul Dasar",
		"deskripsi": "Deskripsi modul",
	})

	mockUC.On("InitiateModulUpload", mock.Anything, "user-123", int64(2048), metadataHeader).Return(&dto.TusModulUploadResponse{
		UploadID:  "modul-upload-id",
		UploadURL: "/modul/upload/modul-upload-id",
		Offset:    0,
		Length:    2048,
	}, nil)

	req := httptest.NewRequest("POST", "/api/v1/tus/modul/upload", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "2048")
	req.Header.Set("Upload-Metadata", metadataHeader)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))

	mockUC.AssertExpectations(t)
}

func TestTusModulController_UploadChunk_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusModulControllerUsecase)
	baseCtrl := getTusBaseController()
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusModulController(mockUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Patch("/api/v1/tus/modul/:upload_id", controller.UploadChunk)

	chunkData := []byte("abcde")
	mockUC.On("HandleModulChunk", mock.Anything, "modul-upload-id", "user-123", int64(0), mock.Anything).Return(int64(len(chunkData)), nil)

	req := httptest.NewRequest("PATCH", "/api/v1/tus/modul/modul-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", upload.TusContentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
	assert.Equal(t, strconv.Itoa(len(chunkData)), resp.Header.Get("Upload-Offset"))

	mockUC.AssertExpectations(t)
}

func TestTusModulController_GetUploadStatus_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusModulControllerUsecase)
	baseCtrl := getTusBaseController()
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusModulController(mockUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Head("/api/v1/tus/modul/:upload_id", controller.GetUploadStatus)

	mockUC.On("GetModulUploadStatus", mock.Anything, "modul-upload-id", "user-123").Return(int64(512), int64(1024), nil)

	req := httptest.NewRequest("HEAD", "/api/v1/tus/modul/modul-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "512", resp.Header.Get("Upload-Offset"))
	assert.Equal(t, "1024", resp.Header.Get("Upload-Length"))

	mockUC.AssertExpectations(t)
}
