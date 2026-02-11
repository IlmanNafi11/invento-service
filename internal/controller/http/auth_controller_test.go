package http_test

import (
	"bytes"
	"encoding/json"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

func (m *MockAuthUsecase) Logout(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAuthUsecase) RequestPasswordReset(req domain.ResetPasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func getTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env:            "development",
			CorsOriginDev:  "http://localhost:5173",
			CorsOriginProd: "https://yourdomain.com",
		},
		Supabase: config.SupabaseConfig{
			URL:        "https://test.supabase.co",
			AnonKey:    "test_anon_key",
			ServiceKey: "test_service_role_key",
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
			ID:    "user-123",
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
			ID:    "user-123",
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

	req := httptest.NewRequest("POST", "/refresh", nil)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthController_RefreshToken_InternalError(t *testing.T) {
	t.Skip("RefreshToken delegates to Supabase Auth")
}
