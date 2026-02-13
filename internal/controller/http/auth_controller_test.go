package http_test

import (
	"bytes"
	"encoding/json"
	"fiber-boiler-plate/config"
	httpcontroller "fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) Register(req domain.RegisterRequest) (string, *domain.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(1) == nil {
		return args.String(0), nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*domain.AuthResponse), args.Error(2)
}

func (m *MockAuthUsecase) Login(req domain.AuthRequest) (string, *domain.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(1) == nil {
		return args.String(0), nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*domain.AuthResponse), args.Error(2)
}

func (m *MockAuthUsecase) RefreshToken(refreshToken string) (string, *domain.RefreshTokenResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(1) == nil {
		return args.String(0), nil, args.Error(2)
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
	}
}

func decodeBodyMap(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func TestAuthController_Register_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	cookieHelper := helper.NewCookieHelper(cfg)
	controller := httpcontroller.NewAuthController(mockAuthUC, cookieHelper, cfg)

	app := fiber.New()
	app.Post("/register", controller.Register)

	reqBody := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &domain.AuthResponse{
		User: &domain.AuthUserResponse{
			ID:    "user-123",
			Name:  reqBody.Name,
			Email: reqBody.Email,
		},
		AccessToken: "access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   1234567890,
	}

	mockAuthUC.On("Register", reqBody).Return("refresh_token", expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body := decodeBodyMap(t, resp)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Registrasi berhasil", body["message"])
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "access_token", data["access_token"])
	assert.Equal(t, "Bearer", data["token_type"])
	assert.Equal(t, float64(3600), data["expires_in"])
	assert.Equal(t, float64(1234567890), data["expires_at"])
	assert.NotNil(t, data["user"])

	var hasAccessCookie, hasRefreshCookie bool
	for _, c := range resp.Cookies() {
		if c.Name == helper.AccessTokenCookieName {
			hasAccessCookie = true
		}
		if c.Name == helper.RefreshTokenCookieName {
			hasRefreshCookie = true
		}
	}
	assert.True(t, hasAccessCookie)
	assert.True(t, hasRefreshCookie)

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Register_EmailAlreadyExists(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

	app := fiber.New()
	app.Post("/register", controller.Register)

	reqBody := domain.RegisterRequest{Name: "Test User", Email: "test@example.com", Password: "password123"}
	mockAuthUC.On("Register", reqBody).Return("", (*domain.AuthResponse)(nil), apperrors.NewConflictError("Email sudah terdaftar"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Login_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

	app := fiber.New()
	app.Post("/login", controller.Login)

	reqBody := domain.AuthRequest{Email: "test@example.com", Password: "password123"}
	expectedResponse := &domain.AuthResponse{
		User:        &domain.AuthUserResponse{ID: "user-123", Email: reqBody.Email},
		AccessToken: "access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   1234567890,
	}
	mockAuthUC.On("Login", reqBody).Return("refresh_token", expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body := decodeBodyMap(t, resp)
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "access_token", data["access_token"])
	assert.Equal(t, float64(1234567890), data["expires_at"])
	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_Login_InvalidCredentials(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

	app := fiber.New()
	app.Post("/login", controller.Login)

	reqBody := domain.AuthRequest{Email: "test@example.com", Password: "wrongpassword"}
	mockAuthUC.On("Login", reqBody).Return("", (*domain.AuthResponse)(nil), apperrors.NewUnauthorizedError("Email atau password salah"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_RefreshToken_Success(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/refresh", controller.RefreshToken)

	expectedResponse := &domain.RefreshTokenResponse{
		AccessToken: "new_access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   1234567890,
	}
	mockAuthUC.On("RefreshToken", "old_refresh_token").Return("new_refresh_token", expectedResponse, nil)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: helper.RefreshTokenCookieName, Value: "old_refresh_token"})

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body := decodeBodyMap(t, resp)
	data := body["data"].(map[string]interface{})
	assert.Equal(t, "new_access_token", data["access_token"])
	assert.Equal(t, float64(1234567890), data["expires_at"])
	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_RefreshToken_MissingCookie(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/refresh", controller.RefreshToken)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	mockAuthUC.AssertNotCalled(t, "RefreshToken", mock.Anything)
}

func TestAuthController_Logout_ClearsCookies(t *testing.T) {
	mockAuthUC := new(MockAuthUsecase)
	cfg := getTestConfig()
	controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

	app := fiber.New()
	app.Post("/logout", func(c *fiber.Ctx) error {
		c.Locals("access_token", "access-to-logout")
		return controller.Logout(c)
	})

	mockAuthUC.On("Logout", "access-to-logout").Return(nil)

	req := httptest.NewRequest("POST", "/logout", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body := decodeBodyMap(t, resp)
	assert.Equal(t, true, body["success"])
	assert.Equal(t, "Logout berhasil", body["message"])

	cleared := map[string]bool{}
	for _, c := range resp.Cookies() {
		if c.Name == helper.AccessTokenCookieName || c.Name == helper.RefreshTokenCookieName {
			cleared[c.Name] = c.Value == ""
		}
	}
	assert.True(t, cleared[helper.AccessTokenCookieName])
	assert.True(t, cleared[helper.RefreshTokenCookieName])

	mockAuthUC.AssertExpectations(t)
}

func TestAuthController_RequestPasswordReset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockAuthUC := new(MockAuthUsecase)
		cfg := getTestConfig()
		controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

		app := fiber.New()
		app.Post("/api/v1/auth/reset-password", controller.RequestPasswordReset)

		reqBody := domain.ResetPasswordRequest{Email: "test@example.com"}
		mockAuthUC.On("RequestPasswordReset", reqBody).Return(nil)

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body := decodeBodyMap(t, resp)
		assert.Equal(t, true, body["success"])
		assert.Equal(t, "Link reset password telah dikirim ke email Anda", body["message"])

		mockAuthUC.AssertExpectations(t)
	})

	t.Run("validation error invalid email format", func(t *testing.T) {
		mockAuthUC := new(MockAuthUsecase)
		cfg := getTestConfig()
		controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

		app := fiber.New()
		app.Post("/api/v1/auth/reset-password", controller.RequestPasswordReset)

		reqBody := domain.ResetPasswordRequest{Email: "invalid-email"}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body := decodeBodyMap(t, resp)
		assert.Equal(t, false, body["success"])
		assert.Equal(t, "Data validasi tidak valid", body["message"])

		mockAuthUC.AssertNotCalled(t, "RequestPasswordReset", mock.Anything)
	})
}

func TestAuthController_Register_ValidationError(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		mockAuthUC := new(MockAuthUsecase)
		cfg := getTestConfig()
		controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

		app := fiber.New()
		app.Post("/register", controller.Register)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body := decodeBodyMap(t, resp)
		assert.Equal(t, false, body["success"])
		assert.Equal(t, "Data validasi tidak valid", body["message"])

		mockAuthUC.AssertNotCalled(t, "Register", mock.Anything)
	})
}

func TestAuthController_Login_EmptyBody(t *testing.T) {
	t.Run("empty body returns bad request", func(t *testing.T) {
		mockAuthUC := new(MockAuthUsecase)
		cfg := getTestConfig()
		controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

		app := fiber.New()
		app.Post("/login", controller.Login)

		req := httptest.NewRequest("POST", "/login", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body := decodeBodyMap(t, resp)
		assert.Equal(t, false, body["success"])
		assert.Equal(t, "Format request tidak valid", body["message"])

		mockAuthUC.AssertNotCalled(t, "Login", mock.Anything)
	})
}

func TestAuthController_RefreshToken_InvalidToken(t *testing.T) {
	t.Run("expired or invalid refresh token", func(t *testing.T) {
		mockAuthUC := new(MockAuthUsecase)
		cfg := getTestConfig()
		controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

		app := fiber.New()
		app.Post("/api/v1/auth/refresh", controller.RefreshToken)

		mockAuthUC.On("RefreshToken", "invalid_refresh_token").Return("", (*domain.RefreshTokenResponse)(nil), apperrors.NewUnauthorizedError("Refresh token tidak valid atau sudah expired"))

		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: helper.RefreshTokenCookieName, Value: "invalid_refresh_token"})

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body := decodeBodyMap(t, resp)
		assert.Equal(t, false, body["success"])
		assert.Equal(t, "Refresh token tidak valid atau sudah expired", body["message"])

		mockAuthUC.AssertExpectations(t)
	})
}

func TestAuthController_Logout_WithoutAccessToken(t *testing.T) {
	t.Run("logout succeeds and clears cookies without token", func(t *testing.T) {
		mockAuthUC := new(MockAuthUsecase)
		cfg := getTestConfig()
		controller := httpcontroller.NewAuthController(mockAuthUC, helper.NewCookieHelper(cfg), cfg)

		app := fiber.New()
		app.Post("/api/v1/auth/logout", controller.Logout)

		req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body := decodeBodyMap(t, resp)
		assert.Equal(t, true, body["success"])
		assert.Equal(t, "Logout berhasil", body["message"])

		cleared := map[string]bool{}
		for _, c := range resp.Cookies() {
			if c.Name == helper.AccessTokenCookieName || c.Name == helper.RefreshTokenCookieName {
				cleared[c.Name] = c.Value == ""
			}
		}
		assert.True(t, cleared[helper.AccessTokenCookieName])
		assert.True(t, cleared[helper.RefreshTokenCookieName])

		mockAuthUC.AssertNotCalled(t, "Logout", mock.Anything)
	})
}
