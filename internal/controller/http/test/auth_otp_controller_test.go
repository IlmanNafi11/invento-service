package http_test

import (
	"bytes"
	"encoding/json"
	apphttp "fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthOTPUsecase is a mock for usecase.AuthOTPUsecase
type MockAuthOTPUsecase struct {
	mock.Mock
}

func (m *MockAuthOTPUsecase) RegisterWithOTP(req domain.RegisterRequest) (*domain.OTPResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTPResponse), args.Error(1)
}

func (m *MockAuthOTPUsecase) VerifyRegisterOTP(req domain.VerifyOTPRequest) (string, *domain.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil || args.Get(0).(string) == "" {
		return "", nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*domain.AuthResponse), args.Error(2)
}

func (m *MockAuthOTPUsecase) ResendRegisterOTP(req domain.ResendOTPRequest) (*domain.OTPResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTPResponse), args.Error(1)
}

func (m *MockAuthOTPUsecase) InitiateResetPassword(req domain.ResetPasswordRequest) (*domain.OTPResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTPResponse), args.Error(1)
}

func (m *MockAuthOTPUsecase) VerifyResetPasswordOTP(req domain.VerifyOTPRequest) (*domain.OTPResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTPResponse), args.Error(1)
}

func (m *MockAuthOTPUsecase) ConfirmResetPasswordWithOTP(email string, newPassword string) (string, *domain.AuthResponse, error) {
	args := m.Called(email, newPassword)
	if args.Get(0) == nil || args.Get(0).(string) == "" {
		return "", nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*domain.AuthResponse), args.Error(2)
}

func (m *MockAuthOTPUsecase) ResendResetPasswordOTP(req domain.ResendOTPRequest) (*domain.OTPResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTPResponse), args.Error(1)
}

func getTestConfigForOTP() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env:            "development",
			CorsOriginDev:  "http://localhost:5173",
			CorsOriginProd: "https://yourdomain.com",
		},
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 168,
		},
	}
}

// TestGenerateOTP_Success tests successful OTP generation
func TestAuthOTPController_GenerateOTP_Success(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/register", controller.RegisterWithOTP)

	reqBody := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	expectedResponse := &domain.OTPResponse{
		Message:   "Kode OTP telah dikirim ke email Anda",
		ExpiresIn: 300,
	}

	mockAuthOTPUC.On("RegisterWithOTP", reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Kode OTP telah dikirim ke email Anda", response["message"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestValidateOTP_Success tests successful OTP validation
func TestAuthOTPController_ValidateOTP_Success(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/register/verify", controller.VerifyRegisterOTP)

	reqBody := domain.VerifyOTPRequest{
		Email: "test@student.polije.ac.id",
		Code:  "123456",
		Type:  "register",
	}

	expectedAuthResponse := &domain.AuthResponse{
		User: domain.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@student.polije.ac.id",
		},
		AccessToken: "access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	mockAuthOTPUC.On("VerifyRegisterOTP", reqBody).Return("refresh_token", expectedAuthResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/register/verify", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Verifikasi OTP berhasil, akun dibuat", response["message"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestValidateOTP_Invalid tests OTP validation with invalid code
func TestAuthOTPController_ValidateOTP_Invalid(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/register/verify", controller.VerifyRegisterOTP)

	reqBody := domain.VerifyOTPRequest{
		Email: "test@student.polije.ac.id",
		Code:  "000000",
		Type:  "register",
	}

	mockAuthOTPUC.On("VerifyRegisterOTP", reqBody).Return("", (*domain.AuthResponse)(nil), apperrors.NewUnauthorizedError("kode otp salah"))

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/register/verify", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestInitiateResetPassword_Success tests successful password reset initiation
func TestAuthOTPController_InitiateResetPassword_Success(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/reset-password", controller.InitiateResetPassword)

	reqBody := domain.ResetPasswordRequest{
		Email: "test@student.polije.ac.id",
	}

	expectedResponse := &domain.OTPResponse{
		Message:   "Kode OTP untuk reset password telah dikirim ke email Anda",
		ExpiresIn: 300,
	}

	mockAuthOTPUC.On("InitiateResetPassword", reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestVerifyResetPasswordOTP_Success tests successful reset password OTP verification
func TestAuthOTPController_VerifyResetPasswordOTP_Success(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/reset-password/verify", controller.VerifyResetPasswordOTP)

	reqBody := domain.VerifyOTPRequest{
		Email: "test@student.polije.ac.id",
		Code:  "123456",
		Type:  "reset_password",
	}

	expectedResponse := &domain.OTPResponse{
		Message:   "Kode OTP valid, silakan masukkan password baru Anda",
		ExpiresIn: 300,
	}

	mockAuthOTPUC.On("VerifyResetPasswordOTP", reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/reset-password/verify", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestConfirmResetPasswordWithOTP_Success tests successful password reset confirmation
func TestAuthOTPController_ConfirmResetPasswordWithOTP_Success(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/reset-password/confirm", controller.ConfirmResetPasswordWithOTP)

	reqBody := domain.ConfirmResetPasswordOTPRequest{
		Email:       "test@student.polije.ac.id",
		Code:        "123456",
		NewPassword: "newpassword123",
	}

	expectedAuthResponse := &domain.AuthResponse{
		User: domain.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@student.polije.ac.id",
		},
		AccessToken: "new_access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}

	mockAuthOTPUC.On("ConfirmResetPasswordWithOTP", reqBody.Email, reqBody.NewPassword).Return("new_refresh_token", expectedAuthResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/reset-password/confirm", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Password berhasil direset", response["message"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestResendRegisterOTP_Success tests successful OTP resend for registration
func TestAuthOTPController_ResendRegisterOTP_Success(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/register/resend", controller.ResendRegisterOTP)

	reqBody := domain.ResendOTPRequest{
		Email: "test@student.polije.ac.id",
		Type:  "register",
	}

	expectedResponse := &domain.OTPResponse{
		Message:   "Kode OTP baru telah dikirim ke email Anda",
		ExpiresIn: 300,
	}

	mockAuthOTPUC.On("ResendRegisterOTP", reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/register/resend", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestResendResetPasswordOTP_Success tests successful OTP resend for password reset
func TestAuthOTPController_ResendResetPasswordOTP_Success(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/reset-password/resend", controller.ResendResetPasswordOTP)

	reqBody := domain.ResendOTPRequest{
		Email: "test@student.polije.ac.id",
		Type:  "reset_password",
	}

	expectedResponse := &domain.OTPResponse{
		Message:   "Kode OTP baru telah dikirim ke email Anda",
		ExpiresIn: 300,
	}

	mockAuthOTPUC.On("ResendResetPasswordOTP", reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/reset-password/resend", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockAuthOTPUC.AssertExpectations(t)
}

// TestAuthOTPController_ErrorCases tests various error scenarios
func TestAuthOTPController_ErrorCases(t *testing.T) {
	tests := []struct {
		name           string
		testFunc       func(*MockAuthOTPUsecase, *fiber.App) (*http.Response, error)
		expectedStatus int
	}{
		{
			name: "Register with invalid email",
			testFunc: func(mockUC *MockAuthOTPUsecase, app *fiber.App) (*http.Response, error) {
				reqBody := domain.RegisterRequest{
					Name:     "Test User",
					Email:    "invalid-email",
					Password: "password123",
				}
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/api/v1/auth/otp/register", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				return app.Test(req)
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "Verify OTP with invalid request",
			testFunc: func(mockUC *MockAuthOTPUsecase, app *fiber.App) (*http.Response, error) {
				reqBody := map[string]string{
					"email": "test@student.polije.ac.id",
					// Missing code field
				}
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest("POST", "/api/v1/auth/otp/register/verify", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				return app.Test(req)
			},
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthOTPUC := new(MockAuthOTPUsecase)
			cfg := getTestConfigForOTP()
			controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

			app := fiber.New()
			app.Post("/api/v1/auth/otp/register", controller.RegisterWithOTP)
			app.Post("/api/v1/auth/otp/register/verify", controller.VerifyRegisterOTP)

			resp, err := tt.testFunc(mockAuthOTPUC, app)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if len(mockAuthOTPUC.Calls) > 0 {
				mockAuthOTPUC.AssertExpectations(t)
			}
		})
	}
}

// TestAuthOTPController_ResponseStructure tests response structure consistency
func TestAuthOTPController_ResponseStructure(t *testing.T) {
	mockAuthOTPUC := new(MockAuthOTPUsecase)
	cfg := getTestConfigForOTP()
	controller := apphttp.NewAuthOTPController(mockAuthOTPUC, cfg)

	app := fiber.New()
	app.Post("/api/v1/auth/otp/register", controller.RegisterWithOTP)

	reqBody := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	expectedResponse := &domain.OTPResponse{
		Message:   "Kode OTP telah dikirim ke email Anda",
		ExpiresIn: 300,
	}

	mockAuthOTPUC.On("RegisterWithOTP", reqBody).Return(expectedResponse, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/otp/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

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

	mockAuthOTPUC.AssertExpectations(t)
}
