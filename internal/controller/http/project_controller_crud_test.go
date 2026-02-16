package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"invento-service/internal/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	httpcontroller "invento-service/internal/controller/http"

	apperrors "invento-service/internal/errors"
	app_testing "invento-service/internal/testing"
)

// MockProjectUsecase is a mock for ProjectUsecase interface
type MockProjectUsecase struct {
	mock.Mock
}

func (m *MockProjectUsecase) GetList(ctx context.Context, userID, search string, filterSemester int, filterKategori string, page, limit int) (*dto.ProjectListData, error) {
	args := m.Called(userID, search, filterSemester, filterKategori, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProjectListData), args.Error(1)
}

func (m *MockProjectUsecase) GetByID(ctx context.Context, projectID uint, userID string) (*dto.ProjectResponse, error) {
	args := m.Called(projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProjectResponse), args.Error(1)
}

func (m *MockProjectUsecase) UpdateMetadata(ctx context.Context, projectID uint, userID string, req dto.UpdateProjectRequest) error {
	args := m.Called(projectID, userID, req)
	return args.Error(0)
}

func (m *MockProjectUsecase) Delete(ctx context.Context, projectID uint, userID string) error {
	args := m.Called(projectID, userID)
	return args.Error(0)
}

func (m *MockProjectUsecase) Download(ctx context.Context, userID string, projectIDs []uint) (string, error) {
	args := m.Called(userID, projectIDs)
	return args.String(0), args.Error(1)
}

// Test 1: GetByID_Success

func TestProjectController_GetByID_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	expectedProject := &dto.ProjectResponse{
		ID:          1,
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path/file.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUC.On("GetByID", uint(1), "user-1").Return(expectedProject, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project/:id", controller.GetByID)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project/1", http.NoBody)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Detail project berhasil diambil", response["message"])
	assert.Equal(t, float64(200), response["code"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "Test Project", data["nama_project"])

	mockUC.AssertExpectations(t)
}

// Test 2: GetByID_ProjectNotFound
func TestProjectController_GetByID_ProjectNotFound(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	appErr := apperrors.NewNotFoundError("Project")
	mockUC.On("GetByID", uint(999), "user-1").Return(nil, appErr)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project/:id", controller.GetByID)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project/999", http.NoBody)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Project tidak ditemukan", response["message"])

	mockUC.AssertExpectations(t)
}

// Test 3: GetByID_InvalidID

// Test 5: GetList_WithFilters
func TestProjectController_GetList_WithFilters(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	expectedData := &dto.ProjectListData{
		Items: []dto.ProjectListItem{
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
		Pagination: dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	mockUC.On("GetList", "user-1", "search", 1, "website", 1, 10).Return(expectedData, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project", controller.GetList)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project?search=search&filter_semester=1&filter_kategori=website", http.NoBody)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	mockUC.AssertExpectations(t)
}

// Test 6: UpdateMetadata_Success
func TestProjectController_UpdateMetadata_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	updateReq := dto.UpdateProjectRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	mockUC.On("UpdateMetadata", uint(1), "user-1", updateReq).Return(nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Patch("/api/v1/project/:id", controller.UpdateMetadata)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
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
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Metadata project berhasil diperbarui", response["message"])

	mockUC.AssertExpectations(t)
}

// Test 7: Delete_Success
func TestProjectController_Delete_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	mockUC.On("Delete", uint(1), "user-1").Return(nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Delete("/api/v1/project/:id", controller.Delete)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	req := httptest.NewRequest("DELETE", "/api/v1/project/1", http.NoBody)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Project berhasil dihapus", response["message"])

	mockUC.AssertExpectations(t)
}

// Test 8: Download_SingleFile
func TestProjectController_Download_SingleFile(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	downloadReq := dto.ProjectDownloadRequest{
		IDs: []uint{1},
	}

	mockUC.On("Download", "user-1", []uint{1}).Return("/test/path/file.zip", nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Post("/api/v1/project/download", controller.Download)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
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
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	updateReq := dto.UpdateProjectRequest{
		NamaProject: "Updated Project",
	}

	appErr := apperrors.NewForbiddenError("tidak memiliki akses ke project ini")
	mockUC.On("UpdateMetadata", uint(2), "user-1", updateReq).Return(appErr)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "user1@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Patch("/api/v1/project/:id", controller.UpdateMetadata)

	token := app_testing.GenerateTestToken("user-1", "user1@example.com", "user")
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
	assert.Equal(t, "error", response["status"])

	mockUC.AssertExpectations(t)
}

// Test 10: GetList_PaginationEdgeCases
func TestProjectController_GetList_PaginationEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		queryParams   string
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "Default pagination",
			queryParams:   "",
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "Custom pagination",
			queryParams:   "?page=2&limit=20",
			expectedPage:  2,
			expectedLimit: 20,
		},
		{
			name:          "Invalid page defaults to 1",
			queryParams:   "?page=0&limit=10",
			expectedPage:  1,
			expectedLimit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(MockProjectUsecase)
			controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

			expectedData := &dto.ProjectListData{
				Items: []dto.ProjectListItem{},
				Pagination: dto.PaginationData{
					Page:       tt.expectedPage,
					Limit:      tt.expectedLimit,
					TotalItems: 0,
					TotalPages: 0,
				},
			}

			mockUC.On("GetList", "user-1", "", 0, "", tt.expectedPage, tt.expectedLimit).Return(expectedData, nil)

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_id", "user-1")
				c.Locals("user_email", "test@example.com")
				c.Locals("user_role", "user")
				return c.Next()
			})
			app.Get("/api/v1/project", controller.GetList)

			token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
			req := httptest.NewRequest("GET", "/api/v1/project"+tt.queryParams, http.NoBody)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, "success", response["status"])

			mockUC.AssertExpectations(t)
		})
	}
}

// Test 11: UpdateMetadata_InvalidRequestBody
func TestProjectController_UpdateMetadata_InvalidRequestBody(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Patch("/api/v1/project/:id", controller.UpdateMetadata)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	// Invalid JSON
	req := httptest.NewRequest("PATCH", "/api/v1/project/1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// Test 12: UpdateMetadata_ValidationFailure
func TestProjectController_UpdateMetadata_ValidationFailure(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	// Invalid update request (name too short, invalid semester)
	updateReq := dto.UpdateProjectRequest{
		NamaProject: "ab", // too short
		Kategori:    "invalid_category",
		Semester:    0, // invalid semester
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Patch("/api/v1/project/:id", controller.UpdateMetadata)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	bodyBytes, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PATCH", "/api/v1/project/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// Test 13: Delete_ProjectNotFound
