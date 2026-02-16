package http_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/rbac"
	"invento-service/internal/upload"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	base "invento-service/internal/controller/base"
	httpcontroller "invento-service/internal/controller/http"

	dto "invento-service/internal/dto"
	apperrors "invento-service/internal/errors"

	app_testing "invento-service/internal/testing"
)

// MockTusUploadUsecase mocks the TusUploadUsecase interface
type MockTusUploadUsecase struct {
	mock.Mock
}

func (m *MockTusUploadUsecase) CheckUploadSlot(ctx context.Context, userID string) (*dto.TusUploadSlotResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusUploadSlotResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) ResetUploadQueue(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTusUploadUsecase) InitiateUpload(ctx context.Context, userID, userEmail, userRole string, fileSize int64, metadata dto.TusUploadInitRequest) (*dto.TusUploadResponse, error) {
	args := m.Called(ctx, userID, userEmail, userRole, fileSize, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusUploadResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) HandleChunk(ctx context.Context, uploadID, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(ctx, uploadID, userID, offset, mock.Anything)
	if args.Get(1) != nil {
		return args.Get(0).(int64), args.Error(1)
	}
	return args.Get(0).(int64), nil
}

func (m *MockTusUploadUsecase) GetUploadInfo(ctx context.Context, uploadID, userID string) (*dto.TusUploadInfoResponse, error) {
	args := m.Called(ctx, uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusUploadInfoResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) CancelUpload(ctx context.Context, uploadID, userID string) error {
	args := m.Called(ctx, uploadID, userID)
	return args.Error(0)
}

func (m *MockTusUploadUsecase) GetUploadStatus(ctx context.Context, uploadID, userID string) (offset, length int64, err error) {
	args := m.Called(ctx, uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusUploadUsecase) InitiateProjectUpdateUpload(ctx context.Context, projectID uint, userID string, fileSize int64, metadata dto.TusUploadInitRequest) (*dto.TusUploadResponse, error) {
	args := m.Called(ctx, projectID, userID, fileSize, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusUploadResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) HandleProjectUpdateChunk(ctx context.Context, projectID uint, uploadID, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(ctx, projectID, uploadID, userID, offset, mock.Anything)
	if args.Get(1) != nil {
		return args.Get(0).(int64), args.Error(1)
	}
	return args.Get(0).(int64), nil
}

func (m *MockTusUploadUsecase) GetProjectUpdateUploadStatus(ctx context.Context, projectID uint, uploadID, userID string) (offset, length int64, err error) {
	args := m.Called(ctx, projectID, uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusUploadUsecase) GetProjectUpdateUploadInfo(ctx context.Context, projectID uint, uploadID, userID string) (*dto.TusUploadInfoResponse, error) {
	args := m.Called(ctx, projectID, uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TusUploadInfoResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) CancelProjectUpdateUpload(ctx context.Context, projectID uint, uploadID, userID string) error {
	args := m.Called(ctx, projectID, uploadID, userID)
	return args.Error(0)
}

// Helper function to create test config
func getTusTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			TusVersion:     "1.0.0",
			ChunkSize:      1048576, // 1MB
			MaxSize:        52428800,
			MaxSizeProject: 524288000,
		},
	}
}

func getTusBaseController() *base.BaseController {
	casbin := &rbac.CasbinEnforcer{}
	return base.NewBaseController("https://test.supabase.co", casbin)
}

// Helper function to set authenticated user in context
func setTusAuthenticatedUser(c *fiber.Ctx, userID, email string) {
	c.Locals("user_id", userID)
	c.Locals("user_email", email)
	c.Locals("user_role", "user")
}

// Helper to encode TUS metadata
func encodeTusMetadata(metadata map[string]string) string {
	var pairs []string
	for key, value := range metadata {
		encoded := base64.StdEncoding.EncodeToString([]byte(value))
		pairs = append(pairs, key+" "+encoded)
	}
	result := ""
	for i, pair := range pairs {
		if i > 0 {
			result += ","
		}
		result += pair
	}
	return result
}

// ============================================================================
// TUS Controller Tests - Project Upload
// ============================================================================

// TestInitiateUpload_Success tests successful TUS upload initiation

func TestInitiateUpload_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	metadata := dto.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	expectedResponse := &dto.TusUploadResponse{
		UploadID:  "test-upload-id",
		UploadURL: "/project/upload/test-upload-id",
		Offset:    0,
		Length:    1048576,
	}

	mockUC.On("InitiateUpload", mock.Anything, "user-123", "test@example.com", "user", int64(1048576), metadata).Return(expectedResponse, nil)

	metadataHeader := encodeTusMetadata(map[string]string{
		"nama_project": "Test Project",
		"kategori":     "website",
		"semester":     "1",
	})

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "1048576")
	req.Header.Set("Upload-Metadata", metadataHeader)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	// Verify TUS headers
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, "0", resp.Header.Get("Upload-Offset"))
	assert.Equal(t, "1048576", resp.Header.Get("Upload-Length"))
	assert.Contains(t, resp.Header.Get("Location"), "test-upload-id")

	mockUC.AssertExpectations(t)
}

// TestInitiateUpload_InvalidHeaders tests missing TUS-Resumable header
func TestInitiateUpload_InvalidHeaders(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", http.NoBody)
	// Missing Tus-Resumable header
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 412, resp.StatusCode) // TUS spec: 412 Precondition Failed for missing/invalid Tus-Resumable
}

// TestInitiateUpload_InvalidTusVersion tests wrong TUS version
func TestInitiateUpload_InvalidTusVersion(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", http.NoBody)
	req.Header.Set("Tus-Resumable", "0.0.1") // Wrong version
	req.Header.Set("Upload-Length", "1048576")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 412, resp.StatusCode) // Precondition Failed
}

// TestInitiateUpload_Unauthorized tests unauthorized access
func TestInitiateUpload_Unauthorized(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "1048576")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestUploadChunk_Success tests successful chunk upload
func TestUploadChunk_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Patch("/api/v1/tus/upload/:id", controller.UploadChunk)

	chunkData := []byte("test chunk data")
	mockUC.On("HandleChunk", mock.Anything, "test-upload-id", "user-123", int64(0), mock.Anything).Return(int64(len(chunkData)), nil)

	req := httptest.NewRequest("PATCH", "/api/v1/tus/upload/test-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", upload.TusContentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	// Verify TUS headers
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, strconv.Itoa(len(chunkData)), resp.Header.Get("Upload-Offset"))

	mockUC.AssertExpectations(t)
}

// TestUploadChunk_InvalidOffset tests offset mismatch
func TestUploadChunk_InvalidOffset(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Patch("/api/v1/tus/upload/:id", controller.UploadChunk)

	appErr := apperrors.NewTusOffsetError(500, 0)
	mockUC.On("HandleChunk", mock.Anything, "test-upload-id", "user-123", int64(0), mock.Anything).Return(int64(500), appErr)

	chunkData := []byte("test chunk data")
	req := httptest.NewRequest("PATCH", "/api/v1/tus/upload/test-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", upload.TusContentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestUploadChunk_Unauthorized tests unauthorized chunk upload
func TestUploadChunk_Unauthorized(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Patch("/api/v1/tus/upload/:id", controller.UploadChunk)

	chunkData := []byte("test chunk data")
	req := httptest.NewRequest("PATCH", "/api/v1/tus/upload/test-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", upload.TusContentType)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestUploadChunk_InvalidContentType tests invalid content type
func TestUploadChunk_InvalidContentType(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Patch("/api/v1/tus/upload/:id", controller.UploadChunk)

	chunkData := []byte("test chunk data")
	req := httptest.NewRequest("PATCH", "/api/v1/tus/upload/test-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", "application/json") // Wrong content type

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestGetUploadStatus_Success tests HEAD for upload progress
func TestGetUploadStatus_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Head("/api/v1/tus/upload/:id", controller.GetUploadStatus)

	mockUC.On("GetUploadStatus", mock.Anything, "test-upload-id", "user-123").Return(int64(524288), int64(1048576), nil)

	req := httptest.NewRequest("HEAD", "/api/v1/tus/upload/test-upload-id", http.NoBody)
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

// TestGetUploadStatus_NotFound tests non-existent upload
func TestGetUploadStatus_NotFound(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Head("/api/v1/tus/upload/:id", controller.GetUploadStatus)

	mockUC.On("GetUploadStatus", mock.Anything, "nonexistent-id", "user-123").Return(int64(0), int64(0), apperrors.NewNotFoundError("upload"))

	req := httptest.NewRequest("HEAD", "/api/v1/tus/upload/nonexistent-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestGetUploadInfo_Success tests GET for upload metadata
func TestGetUploadInfo_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com")
		return c.Next()
	})
	app.Get("/api/v1/tus/upload/:id", controller.GetUploadInfo)

	expectedInfo := &dto.TusUploadInfoResponse{
		UploadID:    "test-upload-id",
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
		Status:      domain.UploadStatusUploading,
		Progress:    50.0,
		Offset:      524288,
		Length:      1048576,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUC.On("GetUploadInfo", mock.Anything, "test-upload-id", "user-123").Return(expectedInfo, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/upload/test-upload-id", nil, "")

	app_testing.AssertSuccess(t, resp)
	app_testing.AssertDataFieldExists(t, resp)

	mockUC.AssertExpectations(t)
}

// TestGetUploadInfo_Forbidden tests access to another user's upload
