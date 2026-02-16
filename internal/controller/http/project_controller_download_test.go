package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"invento-service/internal/dto"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	httpcontroller "invento-service/internal/controller/http"

	apperrors "invento-service/internal/errors"
	app_testing "invento-service/internal/testing"
)

func TestProjectController_Delete_ProjectNotFound(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	appErr := apperrors.NewNotFoundError("Project")
	mockUC.On("Delete", uint(999), "user-1").Return(appErr)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Delete("/api/v1/project/:id", controller.Delete)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	req := httptest.NewRequest("DELETE", "/api/v1/project/999", http.NoBody)
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

// Test 14: Download_EmptyIDs
func TestProjectController_Download_EmptyIDs(t *testing.T) {
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
	app.Post("/api/v1/project/download", controller.Download)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	// Empty IDs array
	bodyBytes, _ := json.Marshal(dto.ProjectDownloadRequest{IDs: []uint{}})
	req := httptest.NewRequest("POST", "/api/v1/project/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// Test 15: Download_InvalidRequestBody
func TestProjectController_Download_InvalidRequestBody(t *testing.T) {
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
	app.Post("/api/v1/project/download", controller.Download)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/project/download", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

// Test 16: Download_ProjectNotFound
func TestProjectController_Download_ProjectNotFound(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	downloadReq := dto.ProjectDownloadRequest{
		IDs: []uint{1, 2},
	}

	appErr := apperrors.NewNotFoundError("Project")
	mockUC.On("Download", "user-1", []uint{1, 2}).Return("", appErr)

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

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "Project tidak ditemukan", response["message"])

	mockUC.AssertExpectations(t)
}

// Test 17: GetList_UseCaseError
func TestProjectController_GetList_UseCaseError(t *testing.T) {
	t.Parallel()
	mockUC := new(MockProjectUsecase)
	controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

	appErr := apperrors.NewInternalError(fmt.Errorf("Database error"))
	mockUC.On("GetList", "user-1", "search", 0, "", 1, 10).Return(nil, appErr)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		c.Locals("user_email", "test@example.com")
		c.Locals("user_role", "user")
		return c.Next()
	})
	app.Get("/api/v1/project", controller.GetList)

	token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
	req := httptest.NewRequest("GET", "/api/v1/project?search=search", http.NoBody)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response["status"])

	mockUC.AssertExpectations(t)
}

// Test 18: GetList_FilterSemesterBoundaryCases
func TestProjectController_GetList_FilterSemesterBoundaryCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		filterSemester int
		expectedResult bool
	}{
		{
			name:           "Valid semester (8)",
			filterSemester: 8,
			expectedResult: true,
		},
		{
			name:           "Invalid semester (0)",
			filterSemester: 0,
			expectedResult: true, // Controller handles this, not validation
		},
		{
			name:           "Invalid semester (9)",
			filterSemester: 9,
			expectedResult: true, // Controller handles this, not validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			mockUC.On("GetList", "user-1", "", tt.filterSemester, "", 1, 10).Return(expectedData, nil)

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_id", "user-1")
				c.Locals("user_email", "test@example.com")
				c.Locals("user_role", "user")
				return c.Next()
			})
			app.Get("/api/v1/project", controller.GetList)

			token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
			url := "/api/v1/project"
			if tt.filterSemester > 0 {
				url += fmt.Sprintf("?filter_semester=%d", tt.filterSemester)
			}
			req := httptest.NewRequest("GET", url, http.NoBody)
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

// Test 19: GetList_FilterKategoriBoundaryCases
func TestProjectController_GetList_FilterKategoriBoundaryCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		filterKategori string
		expectedResult bool
	}{
		{
			name:           "Valid category (website)",
			filterKategori: "website",
			expectedResult: true,
		},
		{
			name:           "Valid category (mobile)",
			filterKategori: "mobile",
			expectedResult: true,
		},
		{
			name:           "Valid category (iot)",
			filterKategori: "iot",
			expectedResult: true,
		},
		{
			name:           "Valid category (machine_learning)",
			filterKategori: "machine_learning",
			expectedResult: true,
		},
		{
			name:           "Valid category (deep_learning)",
			filterKategori: "deep_learning",
			expectedResult: true,
		},
		{
			name:           "Invalid category",
			filterKategori: "invalid_category",
			expectedResult: true, // Controller handles this, not validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(MockProjectUsecase)
			controller := httpcontroller.NewProjectController(mockUC, "https://test.supabase.co", nil)

			expectedData := &dto.ProjectListData{
				Items: []dto.ProjectListItem{
					{
						ID:                 1,
						NamaProject:        "Test Project",
						Kategori:           tt.filterKategori,
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

			mockUC.On("GetList", "user-1", "", 0, tt.filterKategori, 1, 10).Return(expectedData, nil)

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_id", "user-1")
				c.Locals("user_email", "test@example.com")
				c.Locals("user_role", "user")
				return c.Next()
			})
			app.Get("/api/v1/project", controller.GetList)

			token := app_testing.GenerateTestToken("user-1", "test@example.com", "user")
			url := "/api/v1/project"
			if tt.filterKategori != "" {
				url += fmt.Sprintf("?filter_kategori=%s", tt.filterKategori)
			}
			req := httptest.NewRequest("GET", url, http.NoBody)
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

// Test 20: GetList_PaginationBoundaries
