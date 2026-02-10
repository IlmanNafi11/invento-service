package usecase_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MockOTPRepository struct {
	mock.Mock
}

func (m *MockOTPRepository) Create(email, userName, passwordHash, codeHash string, otpType domain.OTPType, expiresAt time.Time, maxAttempts int) (*domain.OTP, error) {
	args := m.Called(email, userName, passwordHash, codeHash, otpType, expiresAt, maxAttempts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTP), args.Error(1)
}

func (m *MockOTPRepository) GetByEmail(email string, otpType domain.OTPType) (*domain.OTP, error) {
	args := m.Called(email, otpType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTP), args.Error(1)
}

func (m *MockOTPRepository) MarkAsUsed(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOTPRepository) IncrementAttempts(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOTPRepository) DeleteByEmail(email string, otpType domain.OTPType) error {
	args := m.Called(email, otpType)
	return args.Error(0)
}

func (m *MockOTPRepository) UpdateResendInfo(id uint, resendCount int, lastResendAt time.Time) error {
	args := m.Called(id, resendCount, lastResendAt)
	return args.Error(0)
}

func (m *MockOTPRepository) DeleteExpired() error {
	args := m.Called()
	return args.Error(0)
}

// TestGenerateOTP_Success tests successful OTP generation during registration
func TestAuthOTPUsecase_GenerateOTP_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockOTPRepo := new(MockOTPRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "../../../keys/private.pem",
			PublicKeyPath:           "../../../keys/public.pem",
			PrivateKeyRotationPath:  "../../../keys/private_rotation.pem",
			PublicKeyRotationPath:   "../../../keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
		OTP: config.OTPConfig{
			Length:         6,
			ExpiryMinutes:  5,
			MaxAttempts:    5,
			ResendMaxTimes: 3,
			ResendCooldownSeconds: 60,
		},
		Resend: config.ResendConfig{
			APIKey: "test_key",
		},
	}

	authOTPUC := usecase.NewAuthOTPUsecase(mockUserRepo, mockRefreshTokenRepo, mockOTPRepo, mockRoleRepo, nil, cfg)

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	mockUserRepo.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)

	expiresAt := time.Now().Add(5 * time.Minute)
	otpRecord := &domain.OTP{
		ID:           1,
		Email:        req.Email,
		UserName:     req.Name,
		CodeHash:     helper.HashOTP("123456"),
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    expiresAt,
		MaxAttempts:  5,
		Attempts:     0,
		IsUsed:       false,
		ResendCount:  0,
	}

	mockOTPRepo.On("Create", req.Email, req.Name, mock.AnythingOfType("string"), mock.AnythingOfType("string"), domain.OTPTypeRegister, mock.AnythingOfType("time.Time"), 5).Return(otpRecord, nil)

	result, err := authOTPUC.RegisterWithOTP(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Kode OTP telah dikirim ke email Anda", result.Message)
	assert.Equal(t, 300, result.ExpiresIn)

	mockUserRepo.AssertExpectations(t)
	mockOTPRepo.AssertExpectations(t)
}

// TestValidateOTP_Success tests successful OTP validation
func TestAuthOTPUsecase_ValidateOTP_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockOTPRepo := new(MockOTPRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "../../../keys/private.pem",
			PublicKeyPath:           "../../../keys/public.pem",
			PrivateKeyRotationPath:  "../../../keys/private_rotation.pem",
			PublicKeyRotationPath:   "../../../keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
		OTP: config.OTPConfig{
			Length:         6,
			ExpiryMinutes:  5,
			MaxAttempts:    5,
			ResendMaxTimes: 3,
			ResendCooldownSeconds: 60,
		},
		Resend: config.ResendConfig{
			APIKey: "test_key",
		},
	}

	authOTPUC := usecase.NewAuthOTPUsecase(mockUserRepo, mockRefreshTokenRepo, mockOTPRepo, mockRoleRepo, nil, cfg)

	otpCode := "123456"
	otpHash := helper.HashOTP(otpCode)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	expiresAt := time.Now().Add(5 * time.Minute)
	otpRecord := &domain.OTP{
		ID:           1,
		Email:        "test@student.polije.ac.id",
		UserName:     "Test User",
		PasswordHash: string(hashedPassword),
		CodeHash:     otpHash,
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    expiresAt,
		MaxAttempts:  5,
		Attempts:     0,
		IsUsed:       false,
	}

	role := &domain.Role{
		ID:       1,
		NamaRole: "mahasiswa",
	}

	req := domain.VerifyOTPRequest{
		Email: "test@student.polije.ac.id",
		Code:  otpCode,
		Type:  "register",
	}

	mockOTPRepo.On("GetByEmail", req.Email, domain.OTPTypeRegister).Return(otpRecord, nil)
	mockRoleRepo.On("GetByName", "mahasiswa").Return(role, nil)
	mockUserRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
	mockOTPRepo.On("MarkAsUsed", uint(1)).Return(nil)
	mockOTPRepo.On("DeleteByEmail", req.Email, domain.OTPTypeRegister).Return(nil)

	refreshToken, authResp, err := authOTPUC.VerifyRegisterOTP(req)

	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, req.Email, authResp.User.Email)
	assert.NotEmpty(t, authResp.AccessToken)

	mockOTPRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestValidateOTP_Expired tests OTP validation with expired OTP
func TestAuthOTPUsecase_ValidateOTP_Expired(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockOTPRepo := new(MockOTPRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "../../../keys/private.pem",
			PublicKeyPath:           "../../../keys/public.pem",
			PrivateKeyRotationPath:  "../../../keys/private_rotation.pem",
			PublicKeyRotationPath:   "../../../keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
		OTP: config.OTPConfig{
			Length:         6,
			ExpiryMinutes:  5,
			MaxAttempts:    5,
			ResendMaxTimes: 3,
			ResendCooldownSeconds: 60,
		},
		Resend: config.ResendConfig{
			APIKey: "test_key",
		},
	}

	authOTPUC := usecase.NewAuthOTPUsecase(mockUserRepo, mockRefreshTokenRepo, mockOTPRepo, mockRoleRepo, nil, cfg)

	otpHash := helper.HashOTP("123456")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	expiresAt := time.Now().Add(-1 * time.Minute) // Expired
	otpRecord := &domain.OTP{
		ID:           1,
		Email:        "test@student.polije.ac.id",
		UserName:     "Test User",
		PasswordHash: string(hashedPassword),
		CodeHash:     otpHash,
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    expiresAt,
		MaxAttempts:  5,
		Attempts:     0,
		IsUsed:       false,
	}

	req := domain.VerifyOTPRequest{
		Email: "test@student.polije.ac.id",
		Code:  "123456",
		Type:  "register",
	}

	mockOTPRepo.On("GetByEmail", req.Email, domain.OTPTypeRegister).Return(otpRecord, nil)

	refreshToken, authResp, err := authOTPUC.VerifyRegisterOTP(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "kode otp salah")

	mockOTPRepo.AssertExpectations(t)
}

// TestValidateOTP_Invalid tests OTP validation with invalid code
func TestAuthOTPUsecase_ValidateOTP_Invalid(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockOTPRepo := new(MockOTPRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "../../../keys/private.pem",
			PublicKeyPath:           "../../../keys/public.pem",
			PrivateKeyRotationPath:  "../../../keys/private_rotation.pem",
			PublicKeyRotationPath:   "../../../keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
		OTP: config.OTPConfig{
			Length:         6,
			ExpiryMinutes:  5,
			MaxAttempts:    5,
			ResendMaxTimes: 3,
			ResendCooldownSeconds: 60,
		},
		Resend: config.ResendConfig{
			APIKey: "test_key",
		},
	}

	authOTPUC := usecase.NewAuthOTPUsecase(mockUserRepo, mockRefreshTokenRepo, mockOTPRepo, mockRoleRepo, nil, cfg)

	otpHash := helper.HashOTP("123456")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	expiresAt := time.Now().Add(5 * time.Minute)
	otpRecord := &domain.OTP{
		ID:           1,
		Email:        "test@student.polije.ac.id",
		UserName:     "Test User",
		PasswordHash: string(hashedPassword),
		CodeHash:     otpHash,
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    expiresAt,
		MaxAttempts:  5,
		Attempts:     0,
		IsUsed:       false,
	}

	req := domain.VerifyOTPRequest{
		Email: "test@student.polije.ac.id",
		Code:  "654321", // Wrong OTP
		Type:  "register",
	}

	mockOTPRepo.On("GetByEmail", req.Email, domain.OTPTypeRegister).Return(otpRecord, nil)
	mockOTPRepo.On("IncrementAttempts", uint(1)).Return(nil)

	refreshToken, authResp, err := authOTPUC.VerifyRegisterOTP(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "kode otp salah")

	mockOTPRepo.AssertExpectations(t)
}

// TestValidateOTP_TooManyAttempts tests OTP validation with too many attempts
func TestAuthOTPUsecase_ValidateOTP_TooManyAttempts(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockOTPRepo := new(MockOTPRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "../../../keys/private.pem",
			PublicKeyPath:           "../../../keys/public.pem",
			PrivateKeyRotationPath:  "../../../keys/private_rotation.pem",
			PublicKeyRotationPath:   "../../../keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
		OTP: config.OTPConfig{
			Length:         6,
			ExpiryMinutes:  5,
			MaxAttempts:    5,
			ResendMaxTimes: 3,
			ResendCooldownSeconds: 60,
		},
		Resend: config.ResendConfig{
			APIKey: "test_key",
		},
	}

	authOTPUC := usecase.NewAuthOTPUsecase(mockUserRepo, mockRefreshTokenRepo, mockOTPRepo, mockRoleRepo, nil, cfg)

	otpHash := helper.HashOTP("123456")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	expiresAt := time.Now().Add(5 * time.Minute)
	otpRecord := &domain.OTP{
		ID:           1,
		Email:        "test@student.polije.ac.id",
		UserName:     "Test User",
		PasswordHash: string(hashedPassword),
		CodeHash:     otpHash,
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    expiresAt,
		MaxAttempts:  5,
		Attempts:     5, // Max attempts reached
		IsUsed:       false,
	}

	req := domain.VerifyOTPRequest{
		Email: "test@student.polije.ac.id",
		Code:  "123456",
		Type:  "register",
	}

	mockOTPRepo.On("GetByEmail", req.Email, domain.OTPTypeRegister).Return(otpRecord, nil)

	refreshToken, authResp, err := authOTPUC.VerifyRegisterOTP(req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "terlalu banyak percobaan")

	mockOTPRepo.AssertExpectations(t)
}
