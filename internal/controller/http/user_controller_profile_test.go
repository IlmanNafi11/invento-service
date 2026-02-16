package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"invento-service/internal/dto"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"
	app_testing "invento-service/internal/testing"
	apperrors "invento-service/internal/errors"
	httpcontroller "invento-service/internal/controller/http"
)

func TestUserController_GetProfile_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/profile", controller.GetProfile)

	expectedData := &dto.ProfileData{
		Name:          "Test User",
		Email:         "test@example.com",
		JenisKelamin:  stringPtr("Laki-laki"),
		FotoProfil:    stringPtr("/uploads/profiles/test.jpg"),
		Role:          "admin",
		CreatedAt:     time.Now(),
		JumlahProject: 5,
		JumlahModul:   10,
	}

	mockUserUC.On("GetProfile", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(expectedData, nil)

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/profile", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Profil user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 9: UpdateProfile_Success
func TestUserController_UpdateProfile_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	expectedData := &dto.ProfileData{
		Name:          "Updated User",
		Email:         "test@example.com",
		JenisKelamin:  stringPtr("Perempuan"),
		FotoProfil:    stringPtr("/uploads/profiles/updated.jpg"),
		Role:          "user",
		CreatedAt:     time.Now(),
		JumlahProject: 3,
		JumlahModul:   7,
	}

	reqBody := dto.UpdateProfileRequest{
		Name:         "Updated User",
		JenisKelamin: "Perempuan",
	}

	mockUserUC.On("UpdateProfile", mock.Anything, "00000000-0000-0000-0000-000000000001", reqBody, (*multipart.FileHeader)(nil)).Return(expectedData, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "user")
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
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Profil berhasil diperbarui", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 10: UpdateProfile_WithPhoto
func TestUserController_UpdateProfile_WithPhoto(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	expectedData := &dto.ProfileData{
		Name:          "Test User",
		Email:         "test@example.com",
		JenisKelamin:  stringPtr("Laki-laki"),
		FotoProfil:    stringPtr("/uploads/profiles/new_photo.jpg"),
		Role:          "admin",
		CreatedAt:     time.Now(),
		JumlahProject: 5,
		JumlahModul:   10,
	}

	reqBody := dto.UpdateProfileRequest{
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

	mockUserUC.On("UpdateProfile", mock.Anything, "00000000-0000-0000-0000-000000000001", reqBody, mock.Anything).Return(expectedData, nil)

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "admin")
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
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/user/permissions", controller.GetUserPermissions)

	expectedData := []dto.UserPermissionItem{
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

	mockUserUC.On("GetUserPermissions", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(expectedData, nil)

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/user/permissions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Permissions user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 12: GetUserPermissions_EmptyPermissions
func TestUserController_GetUserPermissions_EmptyPermissions(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/user/permissions", controller.GetUserPermissions)

	expectedData := []dto.UserPermissionItem{}

	mockUserUC.On("GetUserPermissions", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(expectedData, nil)

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/user/permissions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	mockUserUC.AssertExpectations(t)
}

// Test 13: DownloadUserFiles_Success
func TestUserController_UpdateProfile_UserNotFound(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	reqBody := dto.UpdateProfileRequest{
		Name:         "Updated User",
		JenisKelamin: "Perempuan",
	}

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("UpdateProfile", mock.Anything, "00000000-0000-0000-0000-000000000001", reqBody, (*multipart.FileHeader)(nil)).Return(nil, appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "user")
	req := httptest.NewRequest("PUT", "/api/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 20: UpdateProfile_ValidationError
func TestUserController_UpdateProfile_ValidationError(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	reqBody := map[string]string{
		"name": "", // Empty name should fail validation
	}

	bodyBytes, _ := json.Marshal(reqBody)
	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "user")
	req := httptest.NewRequest("PUT", "/api/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Test 21: UpdateProfile_InvalidJenisKelamin
func TestUserController_UpdateProfile_InvalidJenisKelamin(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	reqBody := dto.UpdateProfileRequest{
		Name:         "Test User",
		JenisKelamin: "Invalid", // Invalid value
	}

	bodyBytes, _ := json.Marshal(reqBody)
	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "user")
	req := httptest.NewRequest("PUT", "/api/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Test 22: UpdateUserRole_Forbidden
func TestUserController_GetProfile_InternalError(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/profile", controller.GetProfile)

	mockUserUC.On("GetProfile", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(nil, errors.New("database error"))

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/profile", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

func TestUserController_GetProfile_NotFound(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/profile", controller.GetProfile)

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("GetProfile", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(nil, appErr)

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/profile", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

func TestUserController_GetUserPermissions_InternalError(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/user/permissions", controller.GetUserPermissions)

	mockUserUC.On("GetUserPermissions", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(nil, errors.New("db error"))

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/user/permissions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

func TestUserController_GetUserPermissions_NotFound(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Get("/api/v1/user/permissions", controller.GetUserPermissions)

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("GetUserPermissions", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(nil, appErr)

	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "admin")
	req := httptest.NewRequest("GET", "/api/v1/user/permissions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

func TestUserController_UpdateProfile_InternalError(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := setupTestAppWithAuthForUser(controller)
	app.Put("/api/v1/profile", controller.UpdateProfile)

	reqBody := dto.UpdateProfileRequest{
		Name:         "Updated User",
		JenisKelamin: "Perempuan",
	}

	mockUserUC.On("UpdateProfile", mock.Anything, "00000000-0000-0000-0000-000000000001", reqBody, (*multipart.FileHeader)(nil)).Return(nil, errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	token := app_testing.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "user")
	req := httptest.NewRequest("PUT", "/api/v1/profile", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-Test-User-ID", "1")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}
