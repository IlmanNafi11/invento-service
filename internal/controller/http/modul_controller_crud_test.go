package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"invento-service/internal/controller/base"
	httpcontroller "invento-service/internal/controller/http"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/rbac"
	app_testing "invento-service/internal/testing"
)

// MockModulUsecase mocks the ModulUsecase interface
type MockModulUsecase struct {
	mock.Mock
}

func (m *MockModulUsecase) GetList(ctx context.Context, userID string, search string, filterType string, filterStatus string, page, limit int) (*dto.ModulListData, error) {
	args := m.Called(userID, search, filterType, filterStatus, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ModulListData), args.Error(1)
}

func (m *MockModulUsecase) GetByID(ctx context.Context, modulID string, userID string) (*dto.ModulResponse, error) {
	args := m.Called(modulID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ModulResponse), args.Error(1)
}

func (m *MockModulUsecase) UpdateMetadata(ctx context.Context, modulID string, userID string, req dto.UpdateModulRequest) error {
	args := m.Called(modulID, userID, req)
	return args.Error(0)
}

func (m *MockModulUsecase) Delete(ctx context.Context, modulID string, userID string) error {
	args := m.Called(modulID, userID)
	return args.Error(0)
}

func (m *MockModulUsecase) Download(ctx context.Context, userID string, modulIDs []string) (string, error) {
	args := m.Called(userID, modulIDs)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

// Helper function to create test base controller
func getTestBaseController() *base.BaseController {
	casbin := &rbac.CasbinEnforcer{}
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
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()

	// Setup middleware to inject authenticated user
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})

	app.Get("/api/v1/modul", controller.GetList)

	expectedData := &dto.ModulListData{
		Items: []dto.ModulListItem{
			{
				ID:                 "550e8400-e29b-41d4-a716-446655440001",
				Judul:              "Modul 1",
				Deskripsi:          "Deskripsi 1",
				FileName:           "modul1.pdf",
				MimeType:           "application/pdf",
				FileSize:           2621440,
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
				Status:             "completed",
				TerakhirDiperbarui: time.Now(),
			},
		},
		Pagination: dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 2,
			TotalPages: 1,
		},
	}

	mockModulUC.On("GetList", "user-1", "", "", "", 0, 0).Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/modul", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_GetList_WithSearchAndFilters tests with search and filters
func TestModulController_GetList_WithSearchAndFilters(t *testing.T) {
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/modul", controller.GetList)

	expectedData := &dto.ModulListData{
		Items: []dto.ModulListItem{
			{
				ID:                 "550e8400-e29b-41d4-a716-446655440001",
				Judul:              "Matematika Dasar",
				Deskripsi:          "Deskripsi Matematika",
				FileName:           "math.pdf",
				MimeType:           "application/pdf",
				FileSize:           3145728,
				Status:             "completed",
				TerakhirDiperbarui: time.Now(),
			},
		},
		Pagination: dto.PaginationData{
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
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Get("/api/v1/modul", controller.GetList)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/modul", nil, "")

	app_testing.AssertError(t, resp, fiber.StatusUnauthorized)
}

// TestModulController_UpdateMetadata_Success tests successful metadata update
func TestModulController_UpdateMetadata_Success(t *testing.T) {
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	reqBody := dto.UpdateModulRequest{
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
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	reqBody := dto.UpdateModulRequest{
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
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Patch("/api/v1/modul/:id", controller.UpdateMetadata)

	// Invalid: judul too short
	reqBody := dto.UpdateModulRequest{
		Judul: "ab",
	}

	resp := app_testing.MakeRequest(app, "PATCH", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440000", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusBadRequest)
}

// TestModulController_Delete_Success tests successful module deletion
func TestModulController_Delete_Success(t *testing.T) {
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/modul/:id", controller.Delete)

	mockModulUC.On("Delete", "550e8400-e29b-41d4-a716-446655440000", "user-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/modul/550e8400-e29b-41d4-a716-446655440000", http.NoBody)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Delete_NotFound tests deletion of non-existent module
func TestModulController_Delete_NotFound(t *testing.T) {
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

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
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

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
	t.Parallel()
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

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
