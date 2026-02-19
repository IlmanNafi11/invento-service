package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"invento-service/internal/dto"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpcontroller "invento-service/internal/controller/http"

	apperrors "invento-service/internal/errors"
	app_testing "invento-service/internal/testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserController_UpdateUserRole_Success(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := dto.UpdateUserRoleRequest{
		Role: "admin",
	}

	mockUserUC.On("UpdateUserRole", mock.Anything, "00000000-0000-0000-0000-000000000001", "admin").Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/00000000-0000-0000-0000-000000000001/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Role user berhasil diperbarui", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 4: UpdateUserRole_UserNotFound
func TestUserController_UpdateUserRole_UserNotFound(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := dto.UpdateUserRoleRequest{
		Role: "admin",
	}

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("UpdateUserRole", mock.Anything, "00000000-0000-0000-0000-000000000999", "admin").Return(appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/00000000-0000-0000-0000-000000000999/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 5: DeleteUser_Success
func TestUserController_DeleteUser_Success(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Delete("/api/v1/user/:id", controller.DeleteUser)

	mockUserUC.On("DeleteUser", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/user/00000000-0000-0000-0000-000000000001", http.NoBody)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "User berhasil dihapus", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 6: GetUserFiles_Success
func TestUserController_UpdateUserRole_InvalidRole(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := map[string]interface{}{
		"role": "", // Empty role should fail validation
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/00000000-0000-0000-0000-000000000001/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// Test 18: GetUserFiles_UserNotFound
func TestUserController_UpdateUserRole_Forbidden(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := dto.UpdateUserRoleRequest{
		Role: "admin",
	}

	appErr := apperrors.NewForbiddenError("Anda tidak memiliki akses untuk mengubah role ini")
	mockUserUC.On("UpdateUserRole", mock.Anything, "00000000-0000-0000-0000-000000000001", "admin").Return(appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/00000000-0000-0000-0000-000000000001/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 24: DeleteUser_Forbidden
func TestUserController_DeleteUser_Forbidden(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Delete("/api/v1/user/:id", controller.DeleteUser)

	appErr := apperrors.NewForbiddenError("Anda tidak memiliki akses untuk menghapus user ini")
	mockUserUC.On("DeleteUser", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(appErr)

	req := httptest.NewRequest("DELETE", "/api/v1/user/00000000-0000-0000-0000-000000000001", http.NoBody)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 25: DeleteUser_InternalError
func TestUserController_DeleteUser_InternalError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Delete("/api/v1/user/:id", controller.DeleteUser)

	mockUserUC.On("DeleteUser", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(errors.New("database error"))

	req := httptest.NewRequest("DELETE", "/api/v1/user/00000000-0000-0000-0000-000000000001", http.NoBody)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 26: GetUserFiles_InternalError
func TestUserController_UpdateUserRole_InternalError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Put("/api/v1/user/:id/role", controller.UpdateUserRole)

	reqBody := dto.UpdateUserRoleRequest{
		Role: "admin",
	}

	mockUserUC.On("UpdateUserRole", mock.Anything, "00000000-0000-0000-0000-000000000001", "admin").Return(errors.New("database error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/user/00000000-0000-0000-0000-000000000001/role", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 28: DownloadUserFiles_InternalError
func TestUserController_GetUsersForRole_Success(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Get("/api/v1/role/:id/users", controller.GetUsersForRole)

	expectedData := []dto.UserListItem{
		{
			ID:         "00000000-0000-0000-0000-000000000001",
			Email:      "admin@example.com",
			Role:       "admin",
			DibuatPada: time.Now(),
		},
	}

	mockUserUC.On("GetUsersForRole", mock.Anything, uint(1)).Return(expectedData, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/role/1/users", nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Daftar user untuk role berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

func TestUserController_GetUsersForRole_NotFound(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Get("/api/v1/role/:id/users", controller.GetUsersForRole)

	appErr := apperrors.NewNotFoundError("Role tidak ditemukan")
	mockUserUC.On("GetUsersForRole", mock.Anything, uint(999)).Return(nil, appErr)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/role/999/users", nil, "")

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

func TestUserController_GetUsersForRole_InternalError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Get("/api/v1/role/:id/users", controller.GetUsersForRole)

	mockUserUC.On("GetUsersForRole", mock.Anything, uint(1)).Return(nil, errors.New("db error"))

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/role/1/users", nil, "")

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

func TestUserController_BulkAssignRole_Success(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Post("/api/v1/role/:id/users/bulk", controller.BulkAssignRole)

	reqBody := dto.BulkAssignRoleRequest{
		UserIDs: []string{"user-1", "user-2"},
	}

	mockUserUC.On("BulkAssignRole", mock.Anything, []string{"user-1", "user-2"}, uint(1)).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/role/1/users/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Role berhasil ditetapkan ke user", response["message"])

	mockUserUC.AssertExpectations(t)
}

func TestUserController_BulkAssignRole_NotFound(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Post("/api/v1/role/:id/users/bulk", controller.BulkAssignRole)

	reqBody := dto.BulkAssignRoleRequest{
		UserIDs: []string{"user-1", "user-2"},
	}

	appErr := apperrors.NewNotFoundError("Role tidak ditemukan")
	mockUserUC.On("BulkAssignRole", mock.Anything, []string{"user-1", "user-2"}, uint(999)).Return(appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/role/999/users/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

func TestUserController_BulkAssignRole_ValidationError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Post("/api/v1/role/:id/users/bulk", controller.BulkAssignRole)

	reqBody := dto.BulkAssignRoleRequest{
		UserIDs: []string{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/role/1/users/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUserController_BulkAssignRole_InternalError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC, nil)

	app := fiber.New()
	app.Post("/api/v1/role/:id/users/bulk", controller.BulkAssignRole)

	reqBody := dto.BulkAssignRoleRequest{
		UserIDs: []string{"user-1", "user-2"},
	}

	mockUserUC.On("BulkAssignRole", mock.Anything, []string{"user-1", "user-2"}, uint(1)).Return(errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/role/1/users/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}
