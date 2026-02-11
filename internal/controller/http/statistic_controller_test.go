package http_test

import (
	"encoding/json"
	"errors"
	
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	// Import alias for the http package to test it
	httpcontroller "fiber-boiler-plate/internal/controller/http"
)

// MockStatisticUsecase is a mock for usecase.StatisticUsecase
type MockStatisticUsecase struct {
	mock.Mock
}

func (m *MockStatisticUsecase) GetStatistics(userID string, userRole string) (*domain.StatisticData, error) {
	args := m.Called(userID, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.StatisticData), args.Error(1)
}

// Helper function to create a test app with authenticated context for StatisticController
func setupTestAppWithAuthForStatistic(controller *httpcontroller.StatisticController) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
	})

	// Mock authentication middleware that sets user ID and role from headers
	app.Use(func(c *fiber.Ctx) error {
		if userID := c.Get("X-Test-User-ID"); userID != "" {
			c.Locals("user_id", "00000000-0000-0000-0000-000000000001")
			c.Locals("user_email", "test@example.com")
			c.Locals("user_role", c.Get("X-Test-User-Role", "admin"))
		}
		return c.Next()
	})

	return app
}

// TestStatisticController_GetStatistics_AdminUser_Success tests successful statistics retrieval for admin user
func TestStatisticController_GetStatistics_AdminUser_Success(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := setupTestAppWithAuthForStatistic(controller)
	app.Get("/api/v1/statistic", controller.GetStatistics)

	totalProject := 10
	totalModul := 25
	totalUser := 5
	totalRole := 3

	expectedData := &domain.StatisticData{
		TotalProject: &totalProject,
		TotalModul:   &totalModul,
		TotalUser:    &totalUser,
		TotalRole:    &totalRole,
	}

	mockStatisticUC.On("GetStatistics", "00000000-0000-0000-0000-000000000001", "admin").Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "admin")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Data statistik berhasil diambil", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(totalProject), data["total_project"])
	assert.Equal(t, float64(totalModul), data["total_modul"])
	assert.Equal(t, float64(totalUser), data["total_user"])
	assert.Equal(t, float64(totalRole), data["total_role"])

	mockStatisticUC.AssertExpectations(t)
}

// TestStatisticController_GetStatistics_RegularUser_PartialData tests statistics retrieval for regular user with limited permissions
func TestStatisticController_GetStatistics_RegularUser_PartialData(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := setupTestAppWithAuthForStatistic(controller)
	app.Get("/api/v1/statistic", controller.GetStatistics)

	totalProject := 5
	totalModul := 10

	// Regular user only gets project and modul statistics
	expectedData := &domain.StatisticData{
		TotalProject: &totalProject,
		TotalModul:   &totalModul,
		TotalUser:    nil,
		TotalRole:    nil,
	}

	mockStatisticUC.On("GetStatistics", "00000000-0000-0000-0000-000000000001", "user").Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "user")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(totalProject), data["total_project"])
	assert.Equal(t, float64(totalModul), data["total_modul"])
	assert.Nil(t, data["total_user"])
	assert.Nil(t, data["total_role"])

	mockStatisticUC.AssertExpectations(t)
}

// TestStatisticController_GetStatistics_EmptyData tests statistics retrieval when user has no data
func TestStatisticController_GetStatistics_EmptyData(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := setupTestAppWithAuthForStatistic(controller)
	app.Get("/api/v1/statistic", controller.GetStatistics)

	// User with no data gets nil for all fields
	expectedData := &domain.StatisticData{
		TotalProject: nil,
		TotalModul:   nil,
		TotalUser:    nil,
		TotalRole:    nil,
	}

	mockStatisticUC.On("GetStatistics", "00000000-0000-0000-0000-000000000001", "user").Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "user")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Data statistik berhasil diambil", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Nil(t, data["total_project"])
	assert.Nil(t, data["total_modul"])
	assert.Nil(t, data["total_user"])
	assert.Nil(t, data["total_role"])

	mockStatisticUC.AssertExpectations(t)
}

// TestStatisticController_GetStatistics_Unauthorized tests statistics retrieval without authentication
func TestStatisticController_GetStatistics_Unauthorized(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
	})
	app.Get("/api/v1/statistic", controller.GetStatistics)

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	// No authentication headers set

	resp, err := app.Test(req)

	assert.NoError(t, err)
	// BaseController returns 401 when user_id is not in locals
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestStatisticController_GetStatistics_InternalError tests statistics retrieval with internal server error
func TestStatisticController_GetStatistics_InternalError(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := setupTestAppWithAuthForStatistic(controller)
	app.Get("/api/v1/statistic", controller.GetStatistics)

	mockStatisticUC.On("GetStatistics", "00000000-0000-0000-0000-000000000001", "admin").Return(nil, errors.New("database connection failed"))

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "admin")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockStatisticUC.AssertExpectations(t)
}

// TestStatisticController_GetStatistics_AppError tests statistics retrieval with AppError
func TestStatisticController_GetStatistics_AppError(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := setupTestAppWithAuthForStatistic(controller)
	app.Get("/api/v1/statistic", controller.GetStatistics)

	appErr := apperrors.NewForbiddenError("Anda tidak memiliki akses ke statistik")
	mockStatisticUC.On("GetStatistics", "00000000-0000-0000-0000-000000000001", "user").Return(nil, appErr)

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "user")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])

	mockStatisticUC.AssertExpectations(t)
}

// TestStatisticController_GetStatistics_ResponseHeaders tests that response includes required headers
func TestStatisticController_GetStatistics_ResponseHeaders(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := setupTestAppWithAuthForStatistic(controller)
	app.Get("/api/v1/statistic", controller.GetStatistics)

	totalProject := 10
	expectedData := &domain.StatisticData{
		TotalProject: &totalProject,
	}

	mockStatisticUC.On("GetStatistics", "00000000-0000-0000-0000-000000000001", "admin").Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "admin")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Check for required API headers
	apiVersion := resp.Header.Get("X-API-Version")
	assert.NotEmpty(t, apiVersion, "X-API-Version header should be present")

	apiDeprecated := resp.Header.Get("X-API-Deprecated")
	assert.NotEmpty(t, apiDeprecated, "X-API-Deprecated header should be present")

	mockStatisticUC.AssertExpectations(t)
}

// TestStatisticController_GetStatistics_ResponseStructure tests that response has correct structure
func TestStatisticController_GetStatistics_ResponseStructure(t *testing.T) {
	mockStatisticUC := new(MockStatisticUsecase)
	controller := httpcontroller.NewStatisticController(mockStatisticUC)

	app := setupTestAppWithAuthForStatistic(controller)
	app.Get("/api/v1/statistic", controller.GetStatistics)

	totalProject := 10
	totalModul := 25

	expectedData := &domain.StatisticData{
		TotalProject: &totalProject,
		TotalModul:   &totalModul,
	}

	mockStatisticUC.On("GetStatistics", "00000000-0000-0000-0000-000000000001", "admin").Return(expectedData, nil)

	req := httptest.NewRequest("GET", "/api/v1/statistic", nil)
	req.Header.Set("X-Test-User-ID", "1")
	req.Header.Set("X-Test-User-Role", "admin")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "success", "Response should have 'success' field")
	assert.Contains(t, response, "message", "Response should have 'message' field")
	assert.Contains(t, response, "code", "Response should have 'code' field")
	assert.Contains(t, response, "data", "Response should have 'data' field")
	assert.Contains(t, response, "timestamp", "Response should have 'timestamp' field")

	// Verify field types
	assert.IsType(t, true, response["success"])
	assert.IsType(t, "", response["message"])
	assert.IsType(t, float64(0), response["code"])
	assert.IsType(t, map[string]interface{}{}, response["data"])

	mockStatisticUC.AssertExpectations(t)
}
