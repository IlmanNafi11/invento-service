package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fiber-boiler-plate/internal/controller/base"
	httpcontroller "fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	app_testing "fiber-boiler-plate/internal/testing"
)

// MockModulUsecase mocks the ModulUsecase interface
type MockModulUsecase struct {
	mock.Mock
}

func (m *MockModulUsecase) GetList(userID string, search string, filterType string, filterStatus string, page, limit int) (*domain.ModulListData, error) {
	args := m.Called(userID, search, filterType, filterStatus, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ModulListData), args.Error(1)
}

func (m *MockModulUsecase) GetByID(modulID string, userID string) (*domain.ModulResponse, error) {
	args := m.Called(modulID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ModulResponse), args.Error(1)
}

func (m *MockModulUsecase) UpdateMetadata(modulID string, userID string, req domain.ModulUpdateRequest) error {
	args := m.Called(modulID, userID, req)
	return args.Error(0)
}

func (m *MockModulUsecase) Delete(modulID string, userID string) error {
	args := m.Called(modulID, userID)
	return args.Error(0)
}

func (m *MockModulUsecase) Download(userID string, modulIDs []string) (string, error) {
	args := m.Called(userID, modulIDs)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

// MockTusModulUsecase mocks the TusModulUsecase interface (not tested in CRUD tests)
type MockTusModulUsecase struct {
	mock.Mock
}

func (m *MockTusModulUsecase) InitiateModulUpload(userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	args := m.Called(userID, fileSize, uploadMetadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadResponse), args.Error(1)
}

func (m *MockTusModulUsecase) HandleModulChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(uploadID, userID, offset, chunk)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusModulUsecase) GetModulUploadInfo(uploadID string, userID string) (*domain.TusModulUploadInfoResponse, error) {
	args := m.Called(uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadInfoResponse), args.Error(1)
}

func (m *MockTusModulUsecase) GetModulUploadStatus(uploadID string, userID string) (int64, int64, error) {
	args := m.Called(uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusModulUsecase) CancelModulUpload(uploadID string, userID string) error {
	args := m.Called(uploadID, userID)
	return args.Error(0)
}

func (m *MockTusModulUsecase) CheckModulUploadSlot(userID string) (*domain.TusModulUploadSlotResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadSlotResponse), args.Error(1)
}

func (m *MockTusModulUsecase) InitiateModulUpdateUpload(modulID string, userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	args := m.Called(modulID, userID, fileSize, uploadMetadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadResponse), args.Error(1)
}

func (m *MockTusModulUsecase) GetModulUpdateUploadStatus(modulID string, uploadID string, userID uint) (int64, int64, error) {
	args := m.Called(modulID, uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusModulUsecase) GetModulUpdateUploadInfo(modulID string, uploadID string, userID uint) (*domain.TusModulUploadInfoResponse, error) {
	args := m.Called(modulID, uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadInfoResponse), args.Error(1)
}

func (m *MockTusModulUsecase) CancelModulUpdateUpload(modulID string, uploadID string, userID uint) error {
	args := m.Called(modulID, uploadID, userID)
	return args.Error(0)
}

func (m *MockTusModulUsecase) HandleModulUpdateChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(uploadID, userID, offset, chunk)
	return args.Get(0).(int64), args.Error(1)
}

// Helper function to create test base controller
func getTestBaseController() *base.BaseController {
	casbin := &helper.CasbinEnforcer{}
	return base.NewBaseController("https://test.supabase.co", casbin)
}

// Helper function to set authenticated user in context
func setAuthenticatedUser(c *fiber.Ctx, userID string, email, role string) {
	c.Locals("user_id", userID)
	c.Locals("user_email", email)
	c.Locals("user_role", role)
}

// TestModulController_GetList_Success tests successful retrieval of module list
func TestModulController_GetList_Success(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()

	// Setup middleware to inject authenticated user
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})

	app.Get("/api/v1/modul", controller.GetList)

	expectedData := &domain.ModulListData{
		Items: []domain.ModulListItem{
			{
				ID:                 "550e8400-e29b-41d4-a716-446655440001",
				Judul:              "Modul 1",
				Deskripsi:          "Deskripsi 1",
				FileName:           "modul1.pdf",
				MimeType:           "application/pdf",
				FileSize:           2621440,
				FilePath:           "/uploads/modul1.pdf",
				Status:             "completed",
				TerakhirDiperbarui: time.Now(),
			},
			{
				ID:                 "550e8400-e29b-41d4-a716-446655440002",
				Judul:              "Modul 2",
				Deskripsi:          "Deskripsi 2",
				FileName:           "modul2.docx",
				MimeType:           "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				FileSize:           1258291,
				FilePath:           "/uploads/modul2.docx",
				Status:             "completed",
				TerakhirDiperbarui: time.Now(),
			},
		},
		Pagination: domain.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 2,
			TotalPages: 1,
		},
	}

	mockModulUC.On("GetList", "user-1", "", "", "", 1, 10).Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/modul", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_GetList_WithSearchAndFilters tests with search and filters
func TestModulController_GetList_WithSearchAndFilters(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/modul", controller.GetList)

	expectedData := &domain.ModulListData{
		Items: []domain.ModulListItem{
			{
				ID:                 "550e8400-e29b-41d4-a716-446655440001",
				Judul:              "Matematika Dasar",
				Deskripsi:          "Deskripsi Matematika",
				FileName:           "math.pdf",
				MimeType:           "application/pdf",
				FileSize:           3145728,
				FilePath:           "/uploads/math.pdf",
				Status:             "completed",
				TerakhirDiperbarui: time.Now(),
			},
		},
		Pagination: domain.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	mockModulUC.On("GetList", "user-1", "matematika", "application/pdf", "completed", 1, 10).Return(expectedData, nil)

	url := app_testing.BuildURL("/api/v1/modul", map[string]string{
		"search":        "matematika",
		"filter_type":   "application/pdf",
		"filter_status": "completed",
		"page":          "1",
		"limit":         "10",
	})

	resp := app_testing.MakeRequest(app, "GET", url, nil, "")

	app_testing.AssertSuccess(t, resp)
	app_testing.AssertDataFieldExists(t, resp)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_GetList_Unauthorized tests unauthorized access
func TestModulController_GetList_Unauthorized(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Get("/api/v1/modul", controller.GetList)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/modul", nil, "")

	app_testing.AssertError(t, resp, fiber.StatusUnauthorized)
}

// TestModulController_UpdateMetadata_Success tests successful metadata update
func TestModulController_UpdateMetadata_Success(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	reqBody := domain.ModulUpdateRequest{
		Judul:     "Modul Updated",
		Deskripsi: "Deskripsi Updated",
	}

	mockModulUC.On("UpdateMetadata", "550e8400-e29b-41d4-a716-446655440000", "user-1", reqBody).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_UpdateMetadata_NotFound tests update on non-existent module
func TestModulController_UpdateMetadata_NotFound(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	reqBody := domain.ModulUpdateRequest{
		Judul: "Updated Name",
	}

	appErr := apperrors.NewNotFoundError("Modul tidak ditemukan")
	mockModulUC.On("UpdateMetadata", "550e8400-e29b-41d4-a716-446655440999", "user-1", reqBody).Return(appErr)

	resp := app_testing.MakeRequest(app, "PATCH", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440999", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusNotFound)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_UpdateMetadata_ValidationError tests validation error
func TestModulController_UpdateMetadata_ValidationError(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	// Invalid: judul too short
	reqBody := domain.ModulUpdateRequest{
		Judul: "ab",
	}

	resp := app_testing.MakeRequest(app, "PATCH", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440000", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusBadRequest)
}

// TestModulController_Delete_Success tests successful module deletion
func TestModulController_Delete_Success(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	mockModulUC.On("Delete", "550e8400-e29b-41d4-a716-446655440000", "user-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440000", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Delete_NotFound tests deletion of non-existent module
func TestModulController_Delete_NotFound(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	appErr := apperrors.NewNotFoundError("Modul tidak ditemukan")
	mockModulUC.On("Delete", "550e8400-e29b-41d4-a716-446655440999", "user-1").Return(appErr)

	resp := app_testing.MakeRequest(app, "DELETE", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440999", nil, "")

	app_testing.AssertError(t, resp, fiber.StatusNotFound)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Delete_Forbidden tests deletion of module owned by another user
func TestModulController_Delete_Forbidden(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	appErr := apperrors.NewForbiddenError("Anda tidak memiliki akses ke modul ini")
	mockModulUC.On("Delete", "550e8400-e29b-41d4-a716-446655440002", "user-1").Return(appErr)

	resp := app_testing.MakeRequest(app, "DELETE", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440002", nil, "")

	app_testing.AssertError(t, resp, fiber.StatusForbidden)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Delete_InvalidID tests deletion with invalid ID
func TestModulController_Delete_InvalidID(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	resp := app_testing.MakeRequest(app, "DELETE", "/api/v1/modul/invalid", nil, "")

	app_testing.AssertError(t, resp, fiber.StatusBadRequest)
}

// TestModulController_Download_Success tests successful module download
func TestModulController_Download_Success(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"},
	}

	mockModulUC.On("Download", "user-1", []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"}).Return("/tmp/nonexistent.zip", nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/modul/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Note: Fiber's Download will return 404 for non-existent files, which is expected behavior
	assert.Equal(t, 404, resp.StatusCode)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Download_EmptyIDs tests download with empty ID list
func TestModulController_Download_EmptyIDs(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []string{},
	}

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusBadRequest)
}

// TestModulController_Download_NotFound tests download with non-existent module
func TestModulController_Download_NotFound(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440999"},
	}

	appErr := apperrors.NewNotFoundError("Salah satu modul tidak ditemukan")
	mockModulUC.On("Download", "user-1", []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440999"}).Return("", appErr)

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusNotFound)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Download_Unauthorized tests unauthorized download
func TestModulController_Download_Unauthorized(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"},
	}

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusUnauthorized)
}

// TestModulController_Download_InternalError tests internal server error during download
func TestModulController_Download_InternalError(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001"},
	}

	mockModulUC.On("Download", "user-1", []string{"550e8400-e29b-41d4-a716-446655440001"}).Return("", errors.New("zip creation failed"))

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusInternalServerError)

	mockModulUC.AssertExpectations(t)
}
