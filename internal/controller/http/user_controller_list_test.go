package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"invento-service/internal/dto"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	httpcontroller "invento-service/internal/controller/http"

	apperrors "invento-service/internal/errors"
	app_testing "invento-service/internal/testing"
)

func TestUserController_GetUserList_Success(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user", controller.GetUserList)

	expectedData := &dto.UserListData{
		Items: []dto.UserListItem{
			{
				ID:         "00000000-0000-0000-0000-000000000001",
				Email:      "user1@example.com",
				Role:       "admin",
				DibuatPada: time.Now(),
			},
			{
				ID:         "00000000-0000-0000-0000-000000000002",
				Email:      "user2@example.com",
				Role:       "user",
				DibuatPada: time.Now(),
			},
		},
		Pagination: dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 2,
		},
	}

	params := dto.UserListQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserUC.On("GetUserList", mock.Anything, params).Return(expectedData, nil)

	req := app_testing.GetRequestURL("/api/v1/user", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Daftar user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 2: GetUserList_WithSearchAndFilter
func TestUserController_GetUserList_WithSearchAndFilter(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user", controller.GetUserList)

	expectedData := &dto.UserListData{
		Items: []dto.UserListItem{
			{
				ID:         "00000000-0000-0000-0000-000000000001",
				Email:      "admin@example.com",
				Role:       "admin",
				DibuatPada: time.Now(),
			},
		},
		Pagination: dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 1,
		},
	}

	params := dto.UserListQueryParams{
		Search:     "admin",
		FilterRole: "admin",
		Page:       1,
		Limit:      10,
	}

	mockUserUC.On("GetUserList", mock.Anything, params).Return(expectedData, nil)

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
	assert.Equal(t, "success", response["status"])

	mockUserUC.AssertExpectations(t)
}

// Test 3: UpdateUserRole_Success
func TestUserController_GetUserFiles_Success(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user/:id/files", controller.GetUserFiles)

	expectedData := &dto.UserFilesData{
		Items: []dto.UserFileItem{
			{
				ID:          "00000000-0001-0000-0000-000000000001",
				NamaFile:    "project1.zip",
				Kategori:    "Project",
				DownloadURL: "/uploads/projects/project1.zip",
			},
			{
				ID:          "00000000-0001-0000-0000-000000000002",
				NamaFile:    "modul1.pdf",
				Kategori:    "Modul",
				DownloadURL: "/uploads/moduls/modul1.pdf",
			},
		},
		Pagination: dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 2,
		},
	}

	params := dto.UserFilesQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserUC.On("GetUserFiles", mock.Anything, "00000000-0000-0000-0000-000000000001", params).Return(expectedData, nil)

	req := app_testing.GetRequestURL("/api/v1/user/00000000-0000-0000-0000-000000000001/files", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Daftar file user berhasil diambil", response["message"])

	mockUserUC.AssertExpectations(t)
}

// Test 7: GetUserFiles_WithSearch
func TestUserController_GetUserFiles_WithSearch(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user/:id/files", controller.GetUserFiles)

	expectedData := &dto.UserFilesData{
		Items: []dto.UserFileItem{
			{
				ID:          "00000000-0001-0000-0000-000000000001",
				NamaFile:    "project1.zip",
				Kategori:    "Project",
				DownloadURL: "/uploads/projects/project1.zip",
			},
		},
		Pagination: dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalPages: 1,
			TotalItems: 1,
		},
	}

	params := dto.UserFilesQueryParams{
		Search: "project",
		Page:   1,
		Limit:  10,
	}

	mockUserUC.On("GetUserFiles", mock.Anything, "00000000-0000-0000-0000-000000000001", params).Return(expectedData, nil)

	req := app_testing.GetRequestURL("/api/v1/user/00000000-0000-0000-0000-000000000001/files", map[string]string{
		"search": "project",
		"page":   "1",
		"limit":  "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	mockUserUC.AssertExpectations(t)
}

// Test 8: GetProfile_Success
func TestUserController_DownloadUserFiles_Success(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Post("/api/v1/user/:id/download", controller.DownloadUserFiles)

	reqBody := dto.DownloadUserFilesRequest{
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

	mockUserUC.On("DownloadUserFiles", mock.Anything, "00000000-0000-0000-0000-000000000001", []string{"1", "2"}, []string{"3", "4"}).Return(tmpFile.Name(), nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/user/00000000-0000-0000-0000-000000000001/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Download returns 200 with file
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 14: DownloadUserFiles_EmptyIDs
func TestUserController_DownloadUserFiles_EmptyIDs(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Post("/api/v1/user/:id/download", controller.DownloadUserFiles)

	reqBody := dto.DownloadUserFilesRequest{
		ProjectIDs: []uint{},
		ModulIDs:   []uint{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/user/00000000-0000-0000-0000-000000000001/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "Project IDs atau Modul IDs")
}

// Test 15: DownloadUserFiles_UserNotFound
func TestUserController_DownloadUserFiles_UserNotFound(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Post("/api/v1/user/:id/download", controller.DownloadUserFiles)

	reqBody := dto.DownloadUserFilesRequest{
		ProjectIDs: []uint{1, 2},
		ModulIDs:   []uint{},
	}

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("DownloadUserFiles", mock.Anything, "00000000-0000-0000-0000-000000000999", []string{"1", "2"}, []string{}).Return("", appErr)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/user/00000000-0000-0000-0000-000000000999/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 16: GetUserList_InternalError
func TestUserController_GetUserList_InternalError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user", controller.GetUserList)

	params := dto.UserListQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserUC.On("GetUserList", mock.Anything, params).Return(nil, errors.New("database error"))

	req := app_testing.GetRequestURL("/api/v1/user", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 17: UpdateUserRole_InvalidRole
func TestUserController_GetUserFiles_UserNotFound(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user/:id/files", controller.GetUserFiles)

	params := dto.UserFilesQueryParams{
		Page:  1,
		Limit: 10,
	}

	appErr := apperrors.NewNotFoundError("User tidak ditemukan")
	mockUserUC.On("GetUserFiles", mock.Anything, "00000000-0000-0000-0000-000000000999", params).Return(nil, appErr)

	req := app_testing.GetRequestURL("/api/v1/user/00000000-0000-0000-0000-000000000999/files", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 19: UpdateProfile_UserNotFound
func TestUserController_GetUserFiles_InternalError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Get("/api/v1/user/:id/files", controller.GetUserFiles)

	params := dto.UserFilesQueryParams{
		Page:  1,
		Limit: 10,
	}

	mockUserUC.On("GetUserFiles", mock.Anything, "00000000-0000-0000-0000-000000000001", params).Return(nil, errors.New("database error"))

	req := app_testing.GetRequestURL("/api/v1/user/00000000-0000-0000-0000-000000000001/files", map[string]string{
		"page":  "1",
		"limit": "10",
	})

	resp := app_testing.MakeRequest(app, "GET", req, nil, "")

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}

// Test 27: UpdateUserRole_InternalError
func TestUserController_DownloadUserFiles_InternalError(t *testing.T) {
	t.Parallel()
	mockUserUC := new(MockUserUsecase)
	controller := httpcontroller.NewUserController(mockUserUC)

	app := fiber.New()
	app.Post("/api/v1/user/:id/download", controller.DownloadUserFiles)

	reqBody := dto.DownloadUserFilesRequest{
		ProjectIDs: []uint{1},
		ModulIDs:   []uint{},
	}

	mockUserUC.On("DownloadUserFiles", mock.Anything, "00000000-0000-0000-0000-000000000001", []string{"1"}, []string{}).Return("", errors.New("zip creation failed"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/user/00000000-0000-0000-0000-000000000001/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockUserUC.AssertExpectations(t)
}
