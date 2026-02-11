package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fiber-boiler-plate/config"
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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

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
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/logout", controller.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthController_RefreshToken_MissingToken(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/refresh", controller.RefreshToken)

	req := httptest.NewRequest("POST", "/refresh", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthController_RefreshToken_InvalidToken(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/refresh", controller.RefreshToken)

	refreshToken := "invalid_refresh_token"

	mockAuthUC.On("RefreshToken", refreshToken).Return("", (*domain.RefreshTokenResponse)(nil), apperrors.NewUnauthorizedError("Token tidak valid atau expired"))

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_RefreshToken_ExpiredToken(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/refresh", controller.RefreshToken)

	refreshToken := "expired_refresh_token"

	mockAuthUC.On("RefreshToken", refreshToken).Return("", (*domain.RefreshTokenResponse)(nil), apperrors.NewUnauthorizedError("Token expired"))

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Logout_InvalidToken(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/logout", controller.Logout)

	refreshToken := "invalid_token"

	mockAuthUC.On("Logout", refreshToken).Return(apperrors.NewNotFoundError("Token tidak ditemukan"))

	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_ResetPassword_EmailNotFound(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/reset-password", controller.ResetPassword)

	reqBody := domain.ResetPasswordRequest{
		Email: "nonexistent@example.com",
	}

	mockAuthUC.On("ResetPassword", reqBody).Return(apperrors.NewNotFoundError("Email tidak ditemukan"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_ResetPassword_InvalidEmail(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/reset-password", controller.ResetPassword)

	reqBody := map[string]string{
		"email": "invalid-email-format",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthController_ConfirmResetPassword_InvalidToken(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/confirm-reset", controller.ConfirmResetPassword)

	reqBody := domain.NewPasswordRequest{
		Token:       "invalid_token",
		NewPassword: "newpassword123",
	}

	mockAuthUC.On("ConfirmResetPassword", reqBody).Return(apperrors.NewUnauthorizedError("Token tidak valid atau expired"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/confirm-reset", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_ConfirmResetPassword_WeakPassword(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/confirm-reset", controller.ConfirmResetPassword)

	reqBody := map[string]string{
		"token":        "valid_token",
		"new_password": "123", // Too short
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/confirm-reset", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthController_Register_ValidationError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/register", controller.Register)

	reqBody := map[string]string{
		"name":     "Test",
		"email":    "invalid-email",
		"password": "123",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthController_Login_ValidationError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/login", controller.Login)

	reqBody := map[string]string{
		"email":    "not-an-email",
		"password": "",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestAuthController_Register_InternalError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/register", controller.Register)

	reqBody := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockAuthUC.On("Register", reqBody).Return("", (*domain.AuthResponse)(nil), errors.New("database connection failed"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Login_InternalError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/login", controller.Login)

	reqBody := domain.AuthRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockAuthUC.On("Login", reqBody).Return("", (*domain.AuthResponse)(nil), errors.New("database error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Logout_InternalError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/logout", controller.Logout)

	refreshToken := "valid_token"

	mockAuthUC.On("Logout", refreshToken).Return(errors.New("database error"))

	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_ResetPassword_InternalError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/reset-password", controller.ResetPassword)

	reqBody := domain.ResetPasswordRequest{
		Email: "test@example.com",
	}

	mockAuthUC.On("ResetPassword", reqBody).Return(errors.New("email service error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_ConfirmResetPassword_InternalError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/confirm-reset", controller.ConfirmResetPassword)

	reqBody := domain.NewPasswordRequest{
		Token:       "valid_token",
		NewPassword: "newpassword123",
	}

	mockAuthUC.On("ConfirmResetPassword", reqBody).Return(errors.New("database error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/confirm-reset", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_RefreshToken_InternalError(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, cfg)

	app := fiber.New()
	app.Post("/refresh", controller.RefreshToken)

	refreshToken := "valid_token"

	mockAuthUC.On("RefreshToken", refreshToken).Return("", (*domain.RefreshTokenResponse)(nil), errors.New("token generation failed"))

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockAuthUC.AssertExpectations(t)
}
