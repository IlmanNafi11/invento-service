package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	app_testing "fiber-boiler-plate/internal/testing"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserUsecase is a mock implementation of usecase.UserUsecase
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) GetUserList(params domain.UserListQueryParams) (*domain.UserListData, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserListData), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserRole(userID uint, roleName string) error {
	args := m.Called(userID, roleName)
	return args.Error(0)
}

func (m *MockUserUsecase) DeleteUser(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserUsecase) GetUserFiles(userID uint, params domain.UserFilesQueryParams) (*domain.UserFilesData, error) {
	args := m.Called(userID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserFilesData), args.Error(1)
}

func (m *MockUserUsecase) GetProfile(userID uint) (*domain.ProfileData, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProfileData), args.Error(1)
}

func (m *MockUserUsecase) UpdateProfile(userID uint, req domain.UpdateProfileRequest, fotoProfil interface{}) (*domain.ProfileData, error) {
	args := m.Called(userID, req, fotoProfil)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProfileData), args.Error(1)
}

func (m *MockUserUsecase) GetUserPermissions(userID uint) ([]domain.UserPermissionItem, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserPermissionItem), args.Error(1)
}

func (m *MockUserUsecase) DownloadUserFiles(ownerUserID uint, projectIDs, modulIDs []uint) (string, error) {
	args := m.Called(ownerUserID, projectIDs, modulIDs)
	if args.Get(0) == nil || args.Get(0).(string) == "" {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

// Helper function to create a test app with authenticated middleware for UserController
func setupTestAppWithAuthForUser(controller *http.UserController) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
	})

	// Mock authentication middleware that sets user ID from header
	app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			// Extract user ID from a test header for simplicity
			if userID := c.Get("X-Test-User-ID"); userID != "" {
				c.Locals("user_id", uint(1)) // Mock authenticated user
				c.Locals("user_email", "test@example.com")
				c.Locals("user_role", "admin")
			}
		}
		return c.Next()
	})

	return app
}

// Test 1: GetUserList_Success
func TestUserController_GetUserList_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user", controller.GetUserList)

	expectedData := &domain.UserListData{
		Items: []domain.UserListItem{
			{
				ID:         1,
				Email:      "user1@example.com",
				Role:       "admin",
				DibuatPada: time.Now(),
			},
			{
				ID:         2,
				Email:      "user2@example.com",
				Role:       "user",
				DibuatPada: time.Now(),
			},
		},
		Pagination: domain.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 2,
		},
	}

	params := domain.UserListQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserUC.On("GetUserList", params).Return(expectedData, nil)

	req := app_testing.GetRequestURL("/api/v1/user", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Daftar user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 2: GetUserList_WithSearchAndFilter
func TestUserController_GetUserList_WithSearchAndFilter(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user", controller.GetUserList)

	expectedData := &domain.UserListData{
		Items: []domain.UserListItem{
			{
				ID:         1,
				Email:      "admin@example.com",
				Role:       "admin",
				DibuatPada: time.Now(),
			},
		},
		Pagination: domain.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 1,
		},
	}

	params := domain.UserListQueryParams{
		Search:     "admin",
		FilterRole: "admin",
		Page:       1,
		Limit:      10,
	}

	mockUserUC.On("GetUserList", params).Return(expectedData, nil)

	req := app_testing.GetRequestURL("/api/v1/user", map[string]string{
		"search":      "admin",
		"filter_role": "admin",
		"page":        "1",
		"limit":       "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockUserUC.AssertExpectations(t)
}

// Test 3: UpdateUserRole_Success
func TestUserController_UpdateUserRole_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := domain.UpdateUserRoleRequest{
		Role: "admin",
	}

	mockUserUC.On("UpdateUserRole", uint(1), "admin").Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/1/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Role user berhasil diperbarui", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 4: UpdateUserRole_UserNotFound
func TestUserController_UpdateUserRole_UserNotFound(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := domain.UpdateUserRoleRequest{
		Role: "admin",
	}

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("UpdateUserRole", uint(999), "admin").Return(appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/999/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 5: DeleteUser_Success
func TestUserController_DeleteUser_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Delete("/api/v1/user/:id", controller.DeleteUser)

	mockUserUC.On("DeleteUser", uint(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/user/1", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "User berhasil dihapus", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 6: GetUserFiles_Success
func TestUserController_GetUserFiles_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user/:id/files", controller.GetUserFiles)

	expectedData := &domain.UserFilesData{
		Items: []domain.UserFileItem{
			{
				ID:          1,
				NamaFile:    "project1.zip",
				Kategori:    "Project",
				DownloadURL: "/uploads/projects/project1.zip",
			},
			{
				ID:          2,
				NamaFile:    "modul1.pdf",
				Kategori:    "Modul",
				DownloadURL: "/uploads/moduls/modul1.pdf",
			},
		},
		Pagination: domain.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 2,
		},
	}

	params := domain.UserFilesQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserUC.On("GetUserFiles", uint(1), params).Return(expectedData, nil)

	req := app_testing.GetRequestURL("/api/v1/user/1/files", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Daftar file user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 7: GetUserFiles_WithSearch
func TestUserController_GetUserFiles_WithSearch(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user/:id/files", controller.GetUserFiles)

	expectedData := &domain.UserFilesData{
		Items: []domain.UserFileItem{
			{
				ID:          1,
				NamaFile:    "project1.zip",
				Kategori:    "Project",
				DownloadURL: "/uploads/projects/project1.zip",
			},
		},
		Pagination: domain.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 1,
		},
	}

	params := domain.UserFilesQueryParams{
		Search: "project",
		Page:   1,
		Limit:  10,
	}

	mockUserUC.On("GetUserFiles", uint(1), params).Return(expectedData, nil)

	req := app_testing.GetRequestURL("/api/v1/user/1/files", map[string]string{
		"search": "project",
		"page":   "1",
		"limit":  "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockUserUC.AssertExpectations(t)
}

// Test 8: GetProfile_Success
func TestUserController_GetProfile_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/profile", controller.GetProfile)

	expectedData := &domain.ProfileData{
		Name:          "Test User",
		Email:         "test@example.com",
		JenisKelamin:  stringPtr("Laki-laki"),
		FotoProfil:    stringPtr("/uploads/profiles/test.jpg"),
		Role:          "admin",
		CreatedAt:     time.Now(),
		JumlahProject: 5,
		JumlahModul:   10,
	}

	mockUserUC.On("GetProfile", uint(1)).Return(expectedData, nil)

	token := app_testing.GenerateTestToken(1, "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/profile", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Profil user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 9: UpdateProfile_Success
func TestUserController_UpdateProfile_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	expectedData := &domain.ProfileData{
		Name:          "Updated User",
		Email:         "test@example.com",
		JenisKelamin:  stringPtr("Perempuan"),
		FotoProfil:    stringPtr("/uploads/profiles/updated.jpg"),
		Role:          "user",
		CreatedAt:     time.Now(),
		JumlahProject: 3,
		JumlahModul:   7,
	}

	reqBody := domain.UpdateProfileRequest{
		Name:         "Updated User",
		JenisKelamin: "Perempuan",
	}

	mockUserUC.On("UpdateProfile", uint(1), reqBody, nil).Return(expectedData, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	req := httptest.NewRequest("PUT", "/api/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Profil berhasil diperbarui", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 10: UpdateProfile_WithPhoto
func TestUserController_UpdateProfile_WithPhoto(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	expectedData := &domain.ProfileData{
		Name:          "Test User",
		Email:         "test@example.com",
		JenisKelamin:  stringPtr("Laki-laki"),
		FotoProfil:    stringPtr("/uploads/profiles/new_photo.jpg"),
		Role:          "admin",
		CreatedAt:     time.Now(),
		JumlahProject: 5,
		JumlahModul:   10,
	}

	reqBody := domain.UpdateProfileRequest{
		Name:         "Test User",
		JenisKelamin: "Laki-laki",
	}

	// Create a mock file header
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("jenis_kelamin", "Laki-laki")
	part, _ := writer.CreateFormFile("foto_profil", "photo.jpg")
	part.Write([]byte("mock photo content"))
	writer.Close()

	mockUserUC.On("UpdateProfile", uint(1), reqBody, mock.Anything).Return(expectedData, nil)

	token := app_testing.GenerateTestToken(1, "test@example.com", "admin")
	req := httptest.NewRequest("PUT", "/api/v1/profile", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 11: GetUserPermissions_Success
func TestUserController_GetUserPermissions_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/user/permissions", controller.GetUserPermissions)

	expectedData := []domain.UserPermissionItem{
		{
			Resource: "users",
			Actions:  []string{"read", "create", "update", "delete"},
		},
		{
			Resource: "projects",
			Actions:  []string{"read", "create", "update"},
		},
		{
			Resource: "moduls",
			Actions:  []string{"read", "create"},
		},
	}

	mockUserUC.On("GetUserPermissions", uint(1)).Return(expectedData, nil)

	token := app_testing.GenerateTestToken(1, "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/user/permissions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Permissions user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 12: GetUserPermissions_EmptyPermissions
func TestUserController_GetUserPermissions_EmptyPermissions(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/user/permissions", controller.GetUserPermissions)

	expectedData := []domain.UserPermissionItem{}

	mockUserUC.On("GetUserPermissions", uint(1)).Return(expectedData, nil)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/user/permissions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockUserUC.AssertExpectations(t)
}

// Test 13: DownloadUserFiles_Success
func TestUserController_DownloadUserFiles_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Post("/api/v1/user/:id/download", controller.DownloadUserFiles)

	reqBody := domain.DownloadUserFilesRequest{
		ProjectIDs: []uint{1, 2},
		ModulIDs:   []uint{3, 4},
	}

	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test_download_*.zip")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("test zip content")
	assert.NoError(t, err)
	tmpFile.Close()

	mockUserUC.On("DownloadUserFiles", uint(1), reqBody.ProjectIDs, reqBody.ModulIDs).Return(tmpFile.Name(), nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/user/1/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Download returns 200 with file
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 14: DownloadUserFiles_EmptyIDs
func TestUserController_DownloadUserFiles_EmptyIDs(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Post("/api/v1/user/:id/download", controller.DownloadUserFiles)

	reqBody := domain.DownloadUserFilesRequest{
		ProjectIDs: []uint{},
		ModulIDs:   []uint{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/user/1/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["message"], "Project IDs atau Modul IDs")
}

// Test 15: DownloadUserFiles_UserNotFound
func TestUserController_DownloadUserFiles_UserNotFound(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Post("/api/v1/user/:id/download", controller.DownloadUserFiles)

	reqBody := domain.DownloadUserFilesRequest{
		ProjectIDs: []uint{1, 2},
		ModulIDs:   []uint{},
	}

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("DownloadUserFiles", uint(999), reqBody.ProjectIDs, reqBody.ModulIDs).Return("", appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/user/999/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 16: GetUserList_InternalError
func TestUserController_GetUserList_InternalError(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user", controller.GetUserList)

	params := domain.UserListQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserUC.On("GetUserList", params).Return(nil, errors.New("database error"))

	req := app_testing.GetRequestURL("/api/v1/user", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 17: UpdateUserRole_InvalidRole
func TestUserController_UpdateUserRole_InvalidRole(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := http.NewUserController(mockUserUC)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := map[string]interface{}{
		"role": "", // Empty role should fail validation
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/1/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
