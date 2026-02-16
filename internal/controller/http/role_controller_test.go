package http_test

import (
	"bytes"
	"encoding/json"

	"invento-service/internal/domain"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	// Import alias for the http package to test it
	httpcontroller "invento-service/internal/controller/http"
)

// MockRoleUsecase is a mock for usecase.RoleUsecase
type MockRoleUsecase struct {
	mock.Mock
}

func (m *MockRoleUsecase) GetAvailablePermissions() ([]domain.ResourcePermissions, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ResourcePermissions), args.Error(1)
}

func (m *MockRoleUsecase) GetRoleList(params domain.RoleListQueryParams) (*domain.RoleListData, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RoleListData), args.Error(1)
}

func (m *MockRoleUsecase) CreateRole(req domain.RoleCreateRequest) (*domain.RoleDetailResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RoleDetailResponse), args.Error(1)
}

func (m *MockRoleUsecase) GetRoleDetail(id uint) (*domain.RoleDetailResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RoleDetailResponse), args.Error(1)
}

func (m *MockRoleUsecase) UpdateRole(id uint, req domain.RoleUpdateRequest) (*domain.RoleDetailResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RoleDetailResponse), args.Error(1)
}

func (m *MockRoleUsecase) DeleteRole(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// TestRoleController_GetAvailablePermissions_Success tests successful retrieval
func TestRoleController_GetAvailablePermissions_Success(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Get("/role/permissions", controller.GetAvailablePermissions)

	expectedPermissions := []domain.ResourcePermissions{
		{
			Name: "user",
			Permissions: []domain.PermissionItem{
				{Action: "create", Label: "Buat User"},
				{Action: "read", Label: "Lihat User"},
			},
		},
	}
	mockRoleUC.On("GetAvailablePermissions").Return(expectedPermissions, nil)

	req := httptest.NewRequest("GET", "/role/permissions", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_GetRoleList_Success tests successful role list retrieval
func TestRoleController_GetRoleList_Success(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Get("/role", controller.GetRoleList)

	expectedResult := &domain.RoleListData{
		Items: []domain.RoleListItem{
			{ID: 1, NamaRole: "admin", JumlahPermission: 5, TanggalDiperbarui: time.Now()},
		},
		Pagination: dto.PaginationData{Page: 1, Limit: 10, TotalPages: 1, TotalItems: 1},
	}
	mockRoleUC.On("GetRoleList", mock.Anything).Return(expectedResult, nil)

	req := httptest.NewRequest("GET", "/role?page=1&limit=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_CreateRole_Success tests successful role creation
func TestRoleController_CreateRole_Success(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Post("/role", controller.CreateRole)

	reqBody := domain.RoleCreateRequest{
		NamaRole:    "editor",
		Permissions: map[string][]string{"user": {"read"}},
	}

	expectedResponse := &domain.RoleDetailResponse{
		ID: 1, NamaRole: "editor",
		Permissions:      []domain.RolePermissionDetail{{Resource: "user", Actions: []string{"read"}}},
		JumlahPermission: 1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	mockRoleUC.On("CreateRole", reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_GetRoleDetail_Success tests successful role detail retrieval
func TestRoleController_GetRoleDetail_Success(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Get("/role/:id", controller.GetRoleDetail)

	expectedResponse := &domain.RoleDetailResponse{
		ID: 1, NamaRole: "admin",
		Permissions:      []domain.RolePermissionDetail{{Resource: "user", Actions: []string{"read"}}},
		JumlahPermission: 1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	mockRoleUC.On("GetRoleDetail", uint(1)).Return(expectedResponse, nil)

	req := httptest.NewRequest("GET", "/role/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_UpdateRole_Success tests successful role update
func TestRoleController_UpdateRole_Success(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Put("/role/:id", controller.UpdateRole)

	reqBody := domain.RoleUpdateRequest{
		NamaRole:    "superadmin",
		Permissions: map[string][]string{"user": {"read", "create"}},
	}

	expectedResponse := &domain.RoleDetailResponse{
		ID: 1, NamaRole: "superadmin",
		Permissions:      []domain.RolePermissionDetail{{Resource: "user", Actions: []string{"read", "create"}}},
		JumlahPermission: 1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	mockRoleUC.On("UpdateRole", uint(1), reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/role/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_DeleteRole_Success tests successful role deletion
func TestRoleController_DeleteRole_Success(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Delete("/role/:id", controller.DeleteRole)

	mockRoleUC.On("DeleteRole", uint(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/role/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_ErrorCases tests error handling
func TestRoleController_ErrorCases(t *testing.T) {
	tests := []struct {
		name           string
		testFunc       func(*MockRoleUsecase, *fiber.App) (*http.Response, error)
		expectedStatus int
	}{
		{
			name: "GetAvailablePermissions error",
			testFunc: func(mockUC *MockRoleUsecase, app *fiber.App) (*http.Response, error) {
				mockUC.On("GetAvailablePermissions").Return([]domain.ResourcePermissions{}, assert.AnError)
				req := httptest.NewRequest("GET", "/role/permissions", nil)
				return app.Test(req)
			},
			expectedStatus: fiber.StatusInternalServerError,
		},
		{
			name: "GetRoleDetail not found",
			testFunc: func(mockUC *MockRoleUsecase, app *fiber.App) (*http.Response, error) {
				mockUC.On("GetRoleDetail", uint(999)).Return(nil, apperrors.NewNotFoundError("role"))
				req := httptest.NewRequest("GET", "/role/999", nil)
				return app.Test(req)
			},
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name: "CreateRole duplicate name",
			testFunc: func(mockUC *MockRoleUsecase, app *fiber.App) (*http.Response, error) {
				reqBody := domain.RoleCreateRequest{NamaRole: "admin", Permissions: map[string][]string{"user": {"read"}}}
				mockUC.On("CreateRole", reqBody).Return(nil, apperrors.NewConflictError("nama role sudah ada"))
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/role", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				return app.Test(req)
			},
			expectedStatus: fiber.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleUC := new(MockRoleUsecase)
			controller := httpcontroller.NewRoleController(mockRoleUC, nil)
			app := fiber.New()
			app.Get("/role/permissions", controller.GetAvailablePermissions)
			app.Get("/role/:id", controller.GetRoleDetail)
			app.Post("/role", controller.CreateRole)

			resp, err := tt.testFunc(mockRoleUC, app)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockRoleUC.AssertExpectations(t)
		})
	}
}

// TestRoleController_UpdateRole_NotFound tests update of non-existent role
func TestRoleController_UpdateRole_NotFound(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Put("/role/:id", controller.UpdateRole)

	reqBody := domain.RoleUpdateRequest{
		NamaRole:    "updated_role",
		Permissions: map[string][]string{"user": {"read"}},
	}

	appErr := apperrors.NewNotFoundError("Role tidak ditemukan")
	mockRoleUC.On("UpdateRole", uint(999), reqBody).Return(nil, appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/role/999", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_UpdateRole_ValidationError tests update with invalid data
func TestRoleController_UpdateRole_ValidationError(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Put("/role/:id", controller.UpdateRole)

	reqBody := map[string]interface{}{
		"nama_role":   "", // Empty name should fail validation
		"permissions": map[string][]string{"user": {"read"}},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/role/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestRoleController_UpdateRole_DuplicateName tests update with existing role name
func TestRoleController_UpdateRole_DuplicateName(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Put("/role/:id", controller.UpdateRole)

	reqBody := domain.RoleUpdateRequest{
		NamaRole:    "admin",
		Permissions: map[string][]string{"user": {"read"}},
	}

	appErr := apperrors.NewConflictError("nama role sudah ada")
	mockRoleUC.On("UpdateRole", uint(1), reqBody).Return(nil, appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/role/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_DeleteRole_NotFound tests deletion of non-existent role
func TestRoleController_DeleteRole_NotFound(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Delete("/role/:id", controller.DeleteRole)

	appErr := apperrors.NewNotFoundError("Role tidak ditemukan")
	mockRoleUC.On("DeleteRole", uint(999)).Return(appErr)

	req := httptest.NewRequest("DELETE", "/role/999", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_DeleteRole_Forbidden tests deletion of role that's in use
func TestRoleController_DeleteRole_Forbidden(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Delete("/role/:id", controller.DeleteRole)

	appErr := apperrors.NewForbiddenError("Role sedang digunakan oleh user lain")
	mockRoleUC.On("DeleteRole", uint(1)).Return(appErr)

	req := httptest.NewRequest("DELETE", "/role/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_GetRoleList_InternalError tests internal server error
func TestRoleController_GetRoleList_InternalError(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Get("/role", controller.GetRoleList)

	mockRoleUC.On("GetRoleList", mock.Anything).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/role?page=1&limit=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockRoleUC.AssertExpectations(t)
}

// TestRoleController_CreateRole_ValidationError tests creation with invalid data
func TestRoleController_CreateRole_ValidationError(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Post("/role", controller.CreateRole)

	reqBody := map[string]interface{}{
		"nama_role":   "", // Empty name
		"permissions": map[string][]string{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestRoleController_GetRoleDetail_InvalidID tests detail with invalid ID format
func TestRoleController_GetRoleDetail_InvalidID(t *testing.T) {
	mockRoleUC := new(MockRoleUsecase)
	controller := httpcontroller.NewRoleController(mockRoleUC, nil)
	app := fiber.New()
	app.Get("/role/:id", controller.GetRoleDetail)

	req := httptest.NewRequest("GET", "/role/invalid", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Fiber returns 500 for invalid path parameter types that can't be parsed
	// This is expected behavior
	assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
}
