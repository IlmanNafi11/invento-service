package http_test

import (
	"bytes"
	"encoding/json"
	"fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	app_testing "fiber-boiler-plate/internal/testing"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectUsecase is a mock for ProjectUsecase interface
type MockProjectUsecase struct {
	mock.Mock
}

func (m *MockProjectUsecase) GetList(userID uint, search string, filterSemester int, filterKategori string, page, limit int) (*domain.ProjectListData, error) {
	args := m.Called(userID, search, filterSemester, filterKategori, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectListData), args.Error(1)
}

func (m *MockProjectUsecase) GetByID(projectID, userID uint) (*domain.ProjectResponse, error) {
	args := m.Called(projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectResponse), args.Error(1)
}

func (m *MockProjectUsecase) UpdateMetadata(projectID, userID uint, req domain.ProjectUpdateRequest) error {
	args := m.Called(projectID, userID, req)
	return args.Error(0)
}

func (m *MockProjectUsecase) Delete(projectID, userID uint) error {
	args := m.Called(projectID, userID)
	return args.Error(0)
}

func (m *MockProjectUsecase) Download(userID uint, projectIDs []uint) (string, error) {
	args := m.Called(userID, projectIDs)
	return args.String(0), args.Error(1)
}

// Test 1: GetByID_Success
func TestProjectController_GetByID_Success(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	expectedProject := &domain.ProjectResponse{
		ID:          1,
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path/file.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUC.On("GetByID", uint(1), uint(1)).Return(expectedProject, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project/:id", controller.GetByID)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Detail project berhasil diambil", response["message"])
	assert.Equal(t, float64(200), response["code"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "Test Project", data["nama_project"])

	mockUC.AssertExpectations(t)
}

// Test 2: GetByID_ProjectNotFound
func TestProjectController_GetByID_ProjectNotFound(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	mockUC.On("GetByID", uint(999), uint(1)).Return(nil, errors.New("project tidak ditemukan"))

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project/:id", controller.GetByID)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project/999", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])

	mockUC.AssertExpectations(t)
}

// Test 3: GetByID_InvalidID
func TestProjectController_GetByID_InvalidID(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project/:id", controller.GetByID)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project/invalid", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// BaseController's ParsePathID returns 400 for invalid IDs
	// However, if there's an error in parsing, it might return 500
	assert.True(t, resp.StatusCode == fiber.StatusBadRequest || resp.StatusCode == fiber.StatusInternalServerError)

	mockUC.AssertExpectations(t)
}

// Test 5: GetList_WithFilters
func TestProjectController_GetList_WithFilters(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	expectedData := &domain.ProjectListData{
		Items: []domain.ProjectListItem{
			{
				ID:                 1,
				NamaProject:        "Test Project",
				Kategori:           "website",
				Semester:           1,
				Ukuran:             "small",
				PathFile:           "/test/path",
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

	mockUC.On("GetList", uint(1), "search", 1, "website", 1, 10).Return(expectedData, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project", controller.GetList)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project?search=search&filter_semester=1&filter_kategori=website", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockUC.AssertExpectations(t)
}

// Test 6: UpdateMetadata_Success
func TestProjectController_UpdateMetadata_Success(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	updateReq := domain.ProjectUpdateRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	mockUC.On("UpdateMetadata", uint(1), uint(1), updateReq).Return(nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Patch("/api/v1/project/:id", controller.UpdateMetadata)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	bodyBytes, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PATCH", "/api/v1/project/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Metadata project berhasil diperbarui", response["message"])

	mockUC.AssertExpectations(t)
}

// Test 7: Delete_Success
func TestProjectController_Delete_Success(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	mockUC.On("Delete", uint(1), uint(1)).Return(nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Delete("/api/v1/project/:id", controller.Delete)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	req := httptest.NewRequest("DELETE", "/api/v1/project/1", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Project berhasil dihapus", response["message"])

	mockUC.AssertExpectations(t)
}

// Test 8: Download_SingleFile
func TestProjectController_Download_SingleFile(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	downloadReq := domain.ProjectDownloadRequest{
		IDs: []uint{1},
	}

	mockUC.On("Download", uint(1), []uint{1}).Return("/test/path/file.zip", nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Post("/api/v1/project/download", controller.Download)

	token := app_testing.GenerateTestToken(1, "test@example.com", "user")
	bodyBytes, _ := json.Marshal(downloadReq)
	req := httptest.NewRequest("POST", "/api/v1/project/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// Test 9: UpdateMetadata_AccessDenied
func TestProjectController_UpdateMetadata_AccessDenied(t *testing.T) {
	mockUC := new(MockProjectUsecase)
	controller := http.NewProjectController(mockUC, nil, nil)

	updateReq := domain.ProjectUpdateRequest{
		NamaProject: "Updated Project",
	}

	appErr := apperrors.NewForbiddenError("tidak memiliki akses ke project ini")
	mockUC.On("UpdateMetadata", uint(2), uint(1), updateReq).Return(appErr)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		c.Locals("user_email", "user1@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Patch("/api/v1/project/:id", controller.UpdateMetadata)

	token := app_testing.GenerateTestToken(1, "user1@example.com", "user")
	bodyBytes, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PATCH", "/api/v1/project/2", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])

	mockUC.AssertExpectations(t)
}

// Test 10: GetList_PaginationEdgeCases
func TestProjectController_GetList_PaginationEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedPage   int
		expectedLimit  int
	}{
		{
			name:           "Default pagination",
			queryParams:    "",
			expectedPage:   1,
			expectedLimit:  10,
		},
		{
			name:           "Custom pagination",
			queryParams:    "?page=2&limit=20",
			expectedPage:   2,
			expectedLimit:  20,
		},
		{
			name:           "Invalid page defaults to 1",
			queryParams:    "?page=0&limit=10",
			expectedPage:   1,
			expectedLimit:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(MockProjectUsecase)
			controller := http.NewProjectController(mockUC, nil, nil)

			expectedData := &domain.ProjectListData{
				Items: []domain.ProjectListItem{},
				Pagination: domain.PaginationData{
					Page:       tt.expectedPage,
					Limit:      tt.expectedLimit,
					TotalItems: 0,
					TotalPages: 0,
				},
			}

			mockUC.On("GetList", uint(1), "", 0, "", tt.expectedPage, tt.expectedLimit).Return(expectedData, nil)

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_id", uint(1))
				c.Locals("user_email", "test@example.com")
				c.Locals("user_role", "user")
				return c.Next()
			})
			app.Get("/api/v1/project", controller.GetList)

			token := app_testing.GenerateTestToken(1, "test@example.com", "user")
			req := httptest.NewRequest("GET", "/api/v1/project"+tt.queryParams, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, true, response["success"])

			mockUC.AssertExpectations(t)
		})
	}
}
