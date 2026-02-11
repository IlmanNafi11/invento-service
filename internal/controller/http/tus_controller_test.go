package http_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fiber-boiler-plate/config"
	httpcontroller "fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	app_testing "fiber-boiler-plate/internal/testing"
)

// MockTusUploadUsecase mocks the TusUploadUsecase interface
type MockTusUploadUsecase struct {
	mock.Mock
}

func (m *MockTusUploadUsecase) CheckUploadSlot(userID string) (*domain.TusUploadSlotResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUploadSlotResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) ResetUploadQueue(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockTusUploadUsecase) InitiateUpload(userID string, userEmail string, userRole string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	args := m.Called(userID, userEmail, userRole, fileSize, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUploadResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) HandleChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(uploadID, userID, offset, mock.Anything)
	if args.Get(1) != nil {
		return args.Get(0).(int64), args.Error(1)
	}
	return args.Get(0).(int64), nil
}

func (m *MockTusUploadUsecase) GetUploadInfo(uploadID string, userID string) (*domain.TusUploadInfoResponse, error) {
	args := m.Called(uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUploadInfoResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) CancelUpload(uploadID string, userID string) error {
	args := m.Called(uploadID, userID)
	return args.Error(0)
}

func (m *MockTusUploadUsecase) GetUploadStatus(uploadID string, userID string) (int64, int64, error) {
	args := m.Called(uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusUploadUsecase) InitiateProjectUpdateUpload(projectID uint, userID string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	args := m.Called(projectID, userID, fileSize, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUploadResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) HandleProjectUpdateChunk(projectID uint, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(projectID, uploadID, userID, offset, mock.Anything)
	if args.Get(1) != nil {
		return args.Get(0).(int64), args.Error(1)
	}
	return args.Get(0).(int64), nil
}

func (m *MockTusUploadUsecase) GetProjectUpdateUploadStatus(projectID uint, uploadID string, userID string) (int64, int64, error) {
	args := m.Called(projectID, uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusUploadUsecase) GetProjectUpdateUploadInfo(projectID uint, uploadID string, userID string) (*domain.TusUploadInfoResponse, error) {
	args := m.Called(projectID, uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUploadInfoResponse), args.Error(1)
}

func (m *MockTusUploadUsecase) CancelProjectUpdateUpload(projectID uint, uploadID string, userID string) error {
	args := m.Called(projectID, uploadID, userID)
	return args.Error(0)
}

// Helper function to create test config
func getTusTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			TusVersion: "1.0.0",
			ChunkSize:  1048576, // 1MB
			MaxSize:    52428800,
			MaxSizeProject: 524288000,
		},
	}
}

// Helper function to set authenticated user in context
func setTusAuthenticatedUser(c *fiber.Ctx, userID string, email, role string) {
	c.Locals("user_id", userID)
	c.Locals("user_email", email)
	c.Locals("user_role", role)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	expectedResponse := &domain.TusUploadResponse{
		UploadID:  "test-upload-id",
		UploadURL: "/project/upload/test-upload-id",
		Offset:    0,
		Length:    1048576,
	}

	mockUC.On("InitiateUpload", "user-123", "test@example.com", "user", int64(1048576), metadata).Return(expectedResponse, nil)

	metadataHeader := encodeTusMetadata(map[string]string{
		"nama_project": "Test Project",
		"kategori":     "website",
		"semester":     "1",
	})

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", nil)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", nil)
	// Missing Tus-Resumable header
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestInitiateUpload_InvalidTusVersion tests wrong TUS version
func TestInitiateUpload_InvalidTusVersion(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", nil)
	req.Header.Set("Tus-Resumable", "0.0.1") // Wrong version
	req.Header.Set("Upload-Length", "1048576")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 412, resp.StatusCode) // Precondition Failed
}

// TestInitiateUpload_Unauthorized tests unauthorized access
func TestInitiateUpload_Unauthorized(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/tus/upload", controller.InitiateUpload)

	req := httptest.NewRequest("POST", "/api/v1/tus/upload", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "1048576")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestUploadChunk_Success tests successful chunk upload
func TestUploadChunk_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/tus/upload/:id", controller.UploadChunk)

	chunkData := []byte("test chunk data")
	mockUC.On("HandleChunk", "test-upload-id", "user-123", int64(0), mock.Anything).Return(int64(len(chunkData)), nil)

	req := httptest.NewRequest("PATCH", "/api/v1/tus/upload/test-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", helper.TusContentType)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/tus/upload/:id", controller.UploadChunk)

	appErr := apperrors.NewTusOffsetError(500, 0)
	mockUC.On("HandleChunk", "test-upload-id", "user-123", int64(0), mock.Anything).Return(int64(500), appErr)

	chunkData := []byte("test chunk data")
	req := httptest.NewRequest("PATCH", "/api/v1/tus/upload/test-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", helper.TusContentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Equal(t, "500", resp.Header.Get("Upload-Offset"))

	mockUC.AssertExpectations(t)
}

// TestUploadChunk_Unauthorized tests unauthorized chunk upload
func TestUploadChunk_Unauthorized(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Patch("/api/v1/tus/upload/:id", controller.UploadChunk)

	chunkData := []byte("test chunk data")
	req := httptest.NewRequest("PATCH", "/api/v1/tus/upload/test-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", helper.TusContentType)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestUploadChunk_InvalidContentType tests invalid content type
func TestUploadChunk_InvalidContentType(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
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
	assert.Equal(t, 412, resp.StatusCode)
}

// TestGetUploadStatus_Success tests HEAD for upload progress
func TestGetUploadStatus_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Head("/api/v1/tus/upload/:id", controller.GetUploadStatus)

	mockUC.On("GetUploadStatus", "test-upload-id", "user-123").Return(int64(524288), int64(1048576), nil)

	req := httptest.NewRequest("HEAD", "/api/v1/tus/upload/test-upload-id", nil)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Head("/api/v1/tus/upload/:id", controller.GetUploadStatus)

	mockUC.On("GetUploadStatus", "nonexistent-id", "user-123").Return(int64(0), int64(0), apperrors.NewNotFoundError("upload"))

	req := httptest.NewRequest("HEAD", "/api/v1/tus/upload/nonexistent-id", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestGetUploadInfo_Success tests GET for upload metadata
func TestGetUploadInfo_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/tus/upload/:id", controller.GetUploadInfo)

	expectedInfo := &domain.TusUploadInfoResponse{
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

	mockUC.On("GetUploadInfo", "test-upload-id", "user-123").Return(expectedInfo, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/upload/test-upload-id", nil, "")

	app_testing.AssertSuccess(t, resp)
	app_testing.AssertDataFieldExists(t, resp)

	mockUC.AssertExpectations(t)
}

// TestGetUploadInfo_Forbidden tests access to another user's upload
func TestGetUploadInfo_Forbidden(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-456", "other@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/tus/upload/:id", controller.GetUploadInfo)

	mockUC.On("GetUploadInfo", "test-upload-id", "user-456").Return(nil, apperrors.NewForbiddenError("tidak memiliki akses ke upload ini"))

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/upload/test-upload-id", nil, "")

	app_testing.AssertError(t, resp, 403)

	mockUC.AssertExpectations(t)
}

// TestCancelUpload_Success tests DELETE for cancellation
func TestCancelUpload_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/tus/upload/:id", controller.CancelUpload)

	mockUC.On("CancelUpload", "test-upload-id", "user-123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/tus/upload/test-upload-id", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelUpload_AlreadyCompleted tests cancellation of completed upload
func TestCancelUpload_AlreadyCompleted(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/tus/upload/:id", controller.CancelUpload)

	mockUC.On("CancelUpload", "test-upload-id", "user-123").Return(apperrors.NewTusCompletedError())

	req := httptest.NewRequest("DELETE", "/api/v1/tus/upload/test-upload-id", nil)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/project/:id/upload", controller.InitiateProjectUpdateUpload)

	expectedResponse := &domain.TusUploadResponse{
		UploadID:  "test-update-upload-id",
		UploadURL: "/project/1/update/test-update-upload-id",
		Offset:    0,
		Length:    1048576,
	}

	metadata := domain.TusUploadInitRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	mockUC.On("InitiateProjectUpdateUpload", uint(1), "user-123", int64(1048576), metadata).Return(expectedResponse, nil)

	metadataHeader := encodeTusMetadata(map[string]string{
		"nama_project": "Updated Project",
		"kategori":     "mobile",
		"semester":     "2",
	})

	req := httptest.NewRequest("POST", "/api/v1/tus/project/1/upload", nil)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/project/:id/upload", controller.InitiateProjectUpdateUpload)

	metadata := domain.TusUploadInitRequest{}
	mockUC.On("InitiateProjectUpdateUpload", uint(999), "user-123", int64(1048576), metadata).Return(nil, apperrors.NewNotFoundError("project"))

	req := httptest.NewRequest("POST", "/api/v1/tus/project/999/upload", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", "1048576")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestUploadProjectUpdateChunk_Success tests project chunk upload
func TestUploadProjectUpdateChunk_Success(t *testing.T) {
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
	mockUC.On("HandleProjectUpdateChunk", uint(1), "test-update-upload-id", "user-123", int64(0), mock.Anything).Return(int64(len(chunkData)), nil)

	req := httptest.NewRequest("PATCH", "/api/v1/tus/project/1/upload/test-update-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", helper.TusContentType)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/tus/project/:id/upload/:upload_id", controller.UploadProjectUpdateChunk)

	mockUC.On("HandleProjectUpdateChunk", uint(1), "test-update-upload-id", "user-123", int64(0), mock.Anything).Return(int64(0), apperrors.NewValidationError("project ID tidak cocok", nil))

	chunkData := []byte("test chunk data")
	req := httptest.NewRequest("PATCH", "/api/v1/tus/project/1/upload/test-update-upload-id", bytes.NewReader(chunkData))
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set("Content-Type", helper.TusContentType)
	req.Header.Set("Content-Length", strconv.Itoa(len(chunkData)))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestGetProjectUpdateUploadStatus_Success tests project update upload status
func TestGetProjectUpdateUploadStatus_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Head("/api/v1/tus/project/:id/upload/:upload_id", controller.GetProjectUpdateUploadStatus)

	mockUC.On("GetProjectUpdateUploadStatus", uint(1), "test-update-upload-id", "user-123").Return(int64(524288), int64(1048576), nil)

	req := httptest.NewRequest("HEAD", "/api/v1/tus/project/1/upload/test-update-upload-id", nil)
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/tus/slot", controller.CheckUploadSlot)

	expectedSlot := &domain.TusUploadSlotResponse{
		Available:     true,
		Message:       "Slot tersedia",
		QueueLength:   0,
		ActiveUpload:  false,
		MaxConcurrent: 3,
	}

	mockUC.On("CheckUploadSlot", "user-123").Return(expectedSlot, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/slot", nil, "")

	app_testing.AssertSuccess(t, resp)
	app_testing.AssertJSONField(t, resp, "data.available", true)

	mockUC.AssertExpectations(t)
}

// TestResetUploadQueue_Success tests queue reset
func TestResetUploadQueue_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/reset", controller.ResetUploadQueue)

	mockUC.On("ResetUploadQueue", "user-123").Return(nil)

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/tus/reset", nil, "")

	app_testing.AssertSuccess(t, resp)

	mockUC.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_Success tests GET for project update upload metadata
func TestGetProjectUpdateUploadInfo_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/project/:id/upload/:upload_id", controller.GetProjectUpdateUploadInfo)

	expectedInfo := &domain.TusUploadInfoResponse{
		UploadID:  "test-project-upload-id",
		Status:    domain.UploadStatusUploading,
		Progress:  50.0,
		Offset:    524288,
		Length:    1048576,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUC.On("GetProjectUpdateUploadInfo", uint(1), "test-project-upload-id", "user-123").Return(expectedInfo, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/project/1/upload/test-project-upload-id", nil, "")

	app_testing.AssertSuccess(t, resp)

	mockUC.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_Unauthorized tests GET without authentication
func TestGetProjectUpdateUploadInfo_Unauthorized(t *testing.T) {
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
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/project/:id/upload/:upload_id", controller.GetProjectUpdateUploadInfo)

	mockUC.On("GetProjectUpdateUploadInfo", uint(1), "nonexistent-upload-id", "user-123").Return(nil, apperrors.NewNotFoundError("Upload not found"))

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/project/1/upload/nonexistent-upload-id", nil, "")

	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_Success tests DELETE for project update upload cancellation
func TestCancelProjectUpdateUpload_Success(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/project/:id/upload/:upload_id", controller.CancelProjectUpdateUpload)

	mockUC.On("CancelProjectUpdateUpload", uint(1), "test-project-upload-id", "user-123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/project/1/upload/test-project-upload-id", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_Unauthorized tests DELETE without authentication
func TestCancelProjectUpdateUpload_Unauthorized(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Delete("/api/v1/project/:id/upload/:upload_id", controller.CancelProjectUpdateUpload)

	req := httptest.NewRequest("DELETE", "/api/v1/project/1/upload/test-project-upload-id", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_NotFound tests DELETE with non-existent upload
func TestCancelProjectUpdateUpload_NotFound(t *testing.T) {
	mockUC := new(MockTusUploadUsecase)
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusController(mockUC, cfg)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/project/:id/upload/:upload_id", controller.CancelProjectUpdateUpload)

	mockUC.On("CancelProjectUpdateUpload", uint(1), "nonexistent-upload-id", "user-123").Return(apperrors.NewNotFoundError("Upload not found"))

	req := httptest.NewRequest("DELETE", "/api/v1/project/1/upload/nonexistent-upload-id", nil)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	mockUC.AssertExpectations(t)
}
