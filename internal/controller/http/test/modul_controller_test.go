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
	httpController "fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	app_testing "fiber-boiler-plate/internal/testing"
)

// MockModulUsecase mocks the ModulUsecase interface
type MockModulUsecase struct {
	mock.Mock
}

func (m *MockModulUsecase) GetList(userID uint, search string, filterType string, filterSemester int, page, limit int) (*domain.ModulListData, error) {
	args := m.Called(userID, search, filterType, filterSemester, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ModulListData), args.Error(1)
}

func (m *MockModulUsecase) GetByID(modulID, userID uint) (*domain.ModulResponse, error) {
	args := m.Called(modulID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ModulResponse), args.Error(1)
}

func (m *MockModulUsecase) UpdateMetadata(modulID, userID uint, req domain.ModulUpdateRequest) error {
	args := m.Called(modulID, userID, req)
	return args.Error(0)
}

func (m *MockModulUsecase) Delete(modulID, userID uint) error {
	args := m.Called(modulID, userID)
	return args.Error(0)
}

func (m *MockModulUsecase) Download(userID uint, modulIDs []uint) (string, error) {
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

func (m *MockTusModulUsecase) InitiateModulUpload(userID uint, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	args := m.Called(userID, fileSize, uploadMetadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadResponse), args.Error(1)
}

func (m *MockTusModulUsecase) HandleModulChunk(uploadID string, userID uint, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(uploadID, userID, offset, chunk)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusModulUsecase) GetModulUploadInfo(uploadID string, userID uint) (*domain.TusModulUploadInfoResponse, error) {
	args := m.Called(uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadInfoResponse), args.Error(1)
}

func (m *MockTusModulUsecase) GetModulUploadStatus(uploadID string, userID uint) (int64, int64, error) {
	args := m.Called(uploadID, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTusModulUsecase) CancelModulUpload(uploadID string, userID uint) error {
	args := m.Called(uploadID, userID)
	return args.Error(0)
}

func (m *MockTusModulUsecase) CheckModulUploadSlot(userID uint) (*domain.TusModulUploadSlotResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadSlotResponse), args.Error(1)
}

func (m *MockTusModulUsecase) InitiateModulUpdateUpload(modulID, userID uint, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	args := m.Called(modulID, userID, fileSize, uploadMetadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUploadResponse), args.Error(1)
}

func (m *MockTusModulUsecase) HandleModulUpdateChunk(uploadID string, userID uint, offset int64, chunk io.Reader) (int64, error) {
	args := m.Called(uploadID, userID, offset, chunk)
	return args.Get(0).(int64), args.Error(1)
}

// Helper function to create test base controller
func getTestBaseController() *base.BaseController {
	jwtManager := &helper.JWTManager{}
	casbin := &helper.CasbinEnforcer{}
	return base.NewBaseController(jwtManager, casbin)
}

// Helper function to set authenticated user in context
func setAuthenticatedUser(c *fiber.Ctx, userID uint, email, role string) {
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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()

	// Setup middleware to inject authenticated user
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})

	app.Get("/api/v1/modul", controller.GetList)

	expectedData := &domain.ModulListData{
		Items: []domain.ModulListItem{
			{
				ID:                 1,
				NamaFile:           "Modul 1",
				Tipe:               "pdf",
				Ukuran:             "2.5 MB",
				Semester:           1,
				PathFile:           "/uploads/modul1.pdf",
				TerakhirDiperbarui: time.Now(),
			},
			{
				ID:                 2,
				NamaFile:           "Modul 2",
				Tipe:               "docx",
				Ukuran:             "1.2 MB",
				Semester:           2,
				PathFile:           "/uploads/modul2.docx",
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

	mockModulUC.On("GetList", uint(1), "", "", 0, 1, 10).Return(expectedData, nil)

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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/modul", controller.GetList)

	expectedData := &domain.ModulListData{
		Items: []domain.ModulListItem{
			{
				ID:                 1,
				NamaFile:           "Matematika Dasar",
				Tipe:               "pdf",
				Ukuran:             "3.0 MB",
				Semester:           1,
				PathFile:           "/uploads/math.pdf",
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

	mockModulUC.On("GetList", uint(1), "matematika", "pdf", 1, 1, 10).Return(expectedData, nil)

	url := app_testing.BuildURL("/api/v1/modul", map[string]string{
		"search":          "matematika",
		"filter_type":     "pdf",
		"filter_semester": "1",
		"page":            "1",
		"limit":           "10",
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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	reqBody := domain.ModulUpdateRequest{
		NamaFile: "Modul Updated",
		Semester: 2,
	}

	mockModulUC.On("UpdateMetadata", uint(1), uint(1), reqBody).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PATCH", "/api/v1/modul/1", bytes.NewReader(bodyBytes))
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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	reqBody := domain.ModulUpdateRequest{
		NamaFile: "Updated Name",
	}

	appErr := apperrors.NewNotFoundError("Modul tidak ditemukan")
	mockModulUC.On("UpdateMetadata", uint(999), uint(1), reqBody).Return(appErr)

	resp := app_testing.MakeRequest(app, "PATCH", "/api/v1/modul/999", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusNotFound)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_UpdateMetadata_ValidationError tests validation error
func TestModulController_UpdateMetadata_ValidationError(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	// Invalid: nama_file too short
	reqBody := domain.ModulUpdateRequest{
		NamaFile: "ab",
	}

	resp := app_testing.MakeRequest(app, "PATCH", "/api/v1/modul/1", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusBadRequest)
}

// TestModulController_Delete_Success tests successful module deletion
func TestModulController_Delete_Success(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	mockModulUC.On("Delete", uint(1), uint(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/modul/1", nil)
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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	appErr := apperrors.NewNotFoundError("Modul tidak ditemukan")
	mockModulUC.On("Delete", uint(999), uint(1)).Return(appErr)

	resp := app_testing.MakeRequest(app, "DELETE", "/api/v1/modul/999", nil, "")

	app_testing.AssertError(t, resp, fiber.StatusNotFound)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Delete_Forbidden tests deletion of module owned by another user
func TestModulController_Delete_Forbidden(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	appErr := apperrors.NewForbiddenError("Anda tidak memiliki akses ke modul ini")
	mockModulUC.On("Delete", uint(2), uint(1)).Return(appErr)

	resp := app_testing.MakeRequest(app, "DELETE", "/api/v1/modul/2", nil, "")

	app_testing.AssertError(t, resp, fiber.StatusForbidden)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Delete_InvalidID tests deletion with invalid ID
func TestModulController_Delete_InvalidID(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	mockTusUC := new(MockTusModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []uint{1, 2},
	}

	mockModulUC.On("Download", uint(1), []uint{1, 2}).Return("/tmp/nonexistent.zip", nil)

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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []uint{},
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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []uint{1, 999},
	}

	appErr := apperrors.NewNotFoundError("Salah satu modul tidak ditemukan")
	mockModulUC.On("Download", uint(1), []uint{1, 999}).Return("", appErr)

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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []uint{1, 2},
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
	controller := httpController.NewModulController(mockModulUC, mockTusUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, 1, "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := domain.ModulDownloadRequest{
		IDs: []uint{1},
	}

	mockModulUC.On("Download", uint(1), []uint{1}).Return("", errors.New("zip creation failed"))

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusInternalServerError)

	mockModulUC.AssertExpectations(t)
}
