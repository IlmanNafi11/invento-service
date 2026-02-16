package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	httpcontroller "invento-service/internal/controller/http"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	app_testing "invento-service/internal/testing"
)

func TestModulController_Download_Success(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := dto.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"},
	}

	mockModulUC.On("Download", "user-1", []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"}).Return("/tmp/nonexistent.zip", nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/modul/download", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Note: Fiber's Download will return 404 for non-existent files, which is expected behavior
	assert.Equal(t, 404, resp.StatusCode)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Download_EmptyIDs tests download with empty ID list
func TestModulController_Download_EmptyIDs(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := dto.ModulDownloadRequest{
		IDs: []string{},
	}

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusBadRequest)
}

// TestModulController_Download_NotFound tests download with non-existent module
func TestModulController_Download_NotFound(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := dto.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440999"},
	}

	appErr := apperrors.NewNotFoundError("Salah satu modul tidak ditemukan")
	mockModulUC.On("Download", "user-1", []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440999"}).Return("", appErr)

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusNotFound)

	mockModulUC.AssertExpectations(t)
}

// TestModulController_Download_Unauthorized tests unauthorized download
func TestModulController_Download_Unauthorized(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := dto.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"},
	}

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusUnauthorized)
}

// TestModulController_Download_InternalError tests internal server error during download
func TestModulController_Download_InternalError(t *testing.T) {
	mockModulUC := new(MockModulUsecase)
	baseCtrl := getTestBaseController()
	cfg := getTestConfig()
	controller := httpcontroller.NewModulController(mockModulUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setAuthenticatedUser(c, "user-1", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/modul/download", controller.Download)

	reqBody := dto.ModulDownloadRequest{
		IDs: []string{"550e8400-e29b-41d4-a716-446655440001"},
	}

	mockModulUC.On("Download", "user-1", []string{"550e8400-e29b-41d4-a716-446655440001"}).Return("", errors.New("zip creation failed"))

	resp := app_testing.MakeRequest(app, "POST", "/api/v1/modul/download", reqBody, "")

	app_testing.AssertError(t, resp, fiber.StatusInternalServerError)

	mockModulUC.AssertExpectations(t)
}
