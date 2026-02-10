package http_test

import (
	"bytes"
	"encoding/json"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/controller/http"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/domain"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) Register(req domain.RegisterRequest) (string, *domain.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return "", nil, args.Error(1)
	}
	return args.String(0), args.Get(1).(*domain.AuthResponse), args.Error(2)
}

func (m *MockAuthUsecase) Login(req domain.AuthRequest) (string, *domain.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return "", nil, args.Error(1)
	}
	return args.String(0), args.Get(1).(*domain.AuthResponse), args.Error(2)
}

func (m *MockAuthUsecase) RefreshToken(refreshToken string) (string, *domain.RefreshTokenResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return "", nil, args.Error(1)
	}
	return args.String(0), args.Get(1).(*domain.RefreshTokenResponse), args.Error(2)
}

func (m *MockAuthUsecase) ResetPassword(req domain.ResetPasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockAuthUsecase) ConfirmResetPassword(req domain.NewPasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockAuthUsecase) Logout(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func getTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env:            "development",
			CorsOriginDev:  "http://localhost:5173",
			CorsOriginProd: "https://yourdomain.com",
		},
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 168,
		},
		Upload: config.UploadConfig{
			TusVersion: "1.0.0",
			ChunkSize:  1048576, // 1MB
			MaxSize:    52428800,
		},
	}
}

func TestAuthController_Register_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/register", controller.Register)

	reqBody := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &domain.AuthResponse{
		User: domain.User{
			ID:    1,
			Name:  reqBody.Name,
			Email: reqBody.Email,
		},
		AccessToken: "access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	mockAuthUC.On("Register", reqBody).Return("refresh_token", expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Register_EmailAlreadyExists(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/register", controller.Register)

	reqBody := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockAuthUC.On("Register", reqBody).Return("", (*domain.AuthResponse)(nil), apperrors.NewConflictError("Email sudah terdaftar"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Login_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/login", controller.Login)

	reqBody := domain.AuthRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &domain.AuthResponse{
		User: domain.User{
			ID:    1,
			Email: reqBody.Email,
		},
		AccessToken: "access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	mockAuthUC.On("Login", reqBody).Return("refresh_token", expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Login_InvalidCredentials(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/login", controller.Login)

	reqBody := domain.AuthRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockAuthUC.On("Login", reqBody).Return("", (*domain.AuthResponse)(nil), apperrors.NewUnauthorizedError("Email atau password salah"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_RefreshToken_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/refresh", controller.RefreshToken)

	refreshToken := "valid_refresh_token"

	expectedResponse := &domain.RefreshTokenResponse{
		AccessToken: "new_access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	mockAuthUC.On("RefreshToken", refreshToken).Return("new_refresh_token", expectedResponse, nil)

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_ResetPassword_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/reset-password", controller.ResetPassword)

	reqBody := domain.ResetPasswordRequest{
		Email: "test@example.com",
	}

	mockAuthUC.On("ResetPassword", reqBody).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_ConfirmResetPassword_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/confirm-reset", controller.ConfirmResetPassword)

	reqBody := domain.NewPasswordRequest{
		Token:       "valid_token",
		NewPassword: "newpassword123",
	}

	mockAuthUC.On("ConfirmResetPassword", reqBody).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/confirm-reset", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Logout_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/logout", controller.Logout)

	refreshToken := "valid_refresh_token"

	mockAuthUC.On("Logout", refreshToken).Return(nil)

	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Logout_MissingToken(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := http.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/logout", controller.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
