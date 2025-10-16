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

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uint) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(email, hashedPassword string) error {
	args := m.Called(email, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateProfile(userID uint, name string, jenisKelamin *string, fotoProfil *string) error {
	args := m.Called(userID, name, jenisKelamin, fotoProfil)
	return args.Error(0)
}

func (m *MockUserRepository) GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error) {
	args := m.Called(search, filterRole, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.UserListItem), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) UpdateRole(userID uint, roleID *uint) error {
	args := m.Called(userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(userID uint, token string, expiresAt time.Time) (*domain.RefreshToken, error) {
	args := m.Called(userID, token, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) GetByToken(token string) (*domain.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllUserTokens(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) CleanupExpired() error {
	args := m.Called()
	return args.Error(0)
}

type MockPasswordResetTokenRepository struct {
	mock.Mock
}

func (m *MockPasswordResetTokenRepository) Create(email, token string, expiresAt time.Time) (*domain.PasswordResetToken, error) {
	args := m.Called(email, token, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetTokenRepository) GetByToken(token string) (*domain.PasswordResetToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetTokenRepository) MarkAsUsed(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockPasswordResetTokenRepository) CleanupExpired() error {
	args := m.Called()
	return args.Error(0)
}

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(id uint) (*domain.Role, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(name string) (*domain.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(role *domain.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetAll(search string, page, limit int) ([]domain.RoleListItem, int, error) {
	args := m.Called(search, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.RoleListItem), args.Int(1), args.Error(2)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}

	if _, err := helper.NewJWTManager(cfg); err != nil {
		t.Skip("Skipping test due to missing JWT keys")
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	role := &domain.Role{
		ID:       1,
		NamaRole: "mahasiswa",
	}

	mockUserRepo.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)
	mockRoleRepo.On("GetByName", "mahasiswa").Return(role, nil)
	mockUserRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

	hashedToken := helper.HashRefreshToken("plain_refresh_token")
	refreshToken := &domain.RefreshToken{
		ID:        1,
		UserID:    1,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockRefreshTokenRepo.On("Create", mock.AnythingOfType("uint"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(refreshToken, nil)

	refreshTokenResult, result, err := authUC.Register(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, refreshTokenResult)
	assert.NotEqual(t, hashedToken, refreshTokenResult)
	assert.Equal(t, req.Email, result.User.Email)
	assert.NotEmpty(t, result.AccessToken)
	assert.Equal(t, "Bearer", result.TokenType)

	mockUserRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_EmailAlreadyExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}

	if _, err := helper.NewJWTManager(cfg); err != nil {
		t.Skip("Skipping test due to missing JWT keys")
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	req := domain.RegisterRequest{
		Name:     "Test User",
		Email:    "test@student.polije.ac.id",
		Password: "password123",
	}

	existingUser := &domain.User{
		ID:    1,
		Email: req.Email,
		Name:  "Existing User",
	}

	mockUserRepo.On("GetByEmail", req.Email).Return(existingUser, nil)

	refreshTokenResult, result, err := authUC.Register(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, refreshTokenResult)
	assert.Equal(t, "email sudah terdaftar", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Name:     "Test User",
		IsActive: true,
	}

	req := domain.AuthRequest{
		Email:    user.Email,
		Password: password,
	}

	mockUserRepo.On("GetByEmail", req.Email).Return(user, nil)

	hashedToken := helper.HashRefreshToken("plain_refresh_token")
	refreshToken := &domain.RefreshToken{
		ID:        1,
		UserID:    user.ID,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockRefreshTokenRepo.On("Create", user.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(refreshToken, nil)

	refreshTokenResult, result, err := authUC.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, refreshTokenResult)
	assert.NotEqual(t, hashedToken, refreshTokenResult)
	assert.Equal(t, user.ID, result.User.ID)
	assert.Equal(t, user.Email, result.User.Email)
	assert.NotEmpty(t, result.AccessToken)

	mockUserRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_InvalidCredentials(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}

	if _, err := helper.NewJWTManager(cfg); err != nil {
		t.Skip("Skipping test due to missing JWT keys")
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	req := domain.AuthRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockUserRepo.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)

	refreshTokenResult, result, err := authUC.Login(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, refreshTokenResult)
	assert.Equal(t, "email atau password salah", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_RefreshToken_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}

	if _, err := helper.NewJWTManager(cfg); err != nil {
		t.Skip("Skipping test due to missing JWT keys")
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	refreshTokenString := "valid_refresh_token"
	hashedTokenString := helper.HashRefreshToken(refreshTokenString)
	userID := uint(1)

	refreshToken := &domain.RefreshToken{
		ID:        1,
		UserID:    userID,
		Token:     hashedTokenString,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IsRevoked: false,
	}

	user := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Name:     "Test User",
		IsActive: true,
	}

	mockRefreshTokenRepo.On("GetByToken", hashedTokenString).Return(refreshToken, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockRefreshTokenRepo.On("RevokeToken", hashedTokenString).Return(nil)
	newHashedToken := helper.HashRefreshToken("new_refresh_token")
	newRefreshToken := &domain.RefreshToken{
		ID:        2,
		UserID:    userID,
		Token:     newHashedToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockRefreshTokenRepo.On("Create", userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(newRefreshToken, nil)

	newRefreshTokenResult, result, err := authUC.RefreshToken(refreshTokenString)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, newRefreshTokenResult)
	assert.NotEqual(t, newHashedToken, newRefreshTokenResult)
	assert.NotEmpty(t, result.AccessToken)
	assert.Equal(t, "Bearer", result.TokenType)

	mockRefreshTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_ResetPassword_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}

	if _, err := helper.NewJWTManager(cfg); err != nil {
		t.Skip("Skipping test due to missing JWT keys")
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	email := "test@example.com"
	user := &domain.User{
		ID:    1,
		Email: email,
		Name:  "Test User",
	}

	req := domain.ResetPasswordRequest{
		Email: email,
	}

	mockUserRepo.On("GetByEmail", email).Return(user, nil)

	resetToken := &domain.PasswordResetToken{
		ID:        1,
		Email:     email,
		Token:     "reset_token",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	mockResetTokenRepo.On("Create", email, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(resetToken, nil)

	err := authUC.ResetPassword(req)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockResetTokenRepo.AssertExpectations(t)
}

func TestAuthUsecase_ConfirmResetPassword_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}

	// Skip test if JWT keys are not available
	if _, err := helper.NewJWTManager(cfg); err != nil {
		t.Skip("Skipping test due to missing JWT keys")
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	token := "valid_reset_token"
	email := "test@example.com"
	newPassword := "newpassword123"

	resetToken := &domain.PasswordResetToken{
		ID:        1,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
		IsUsed:    false,
	}

	req := domain.NewPasswordRequest{
		Token:       token,
		NewPassword: newPassword,
	}

	mockResetTokenRepo.On("GetByToken", token).Return(resetToken, nil)
	mockUserRepo.On("UpdatePassword", email, mock.AnythingOfType("string")).Return(nil)
	mockResetTokenRepo.On("MarkAsUsed", token).Return(nil)

	err := authUC.ConfirmResetPassword(req)

	assert.NoError(t, err)

	mockResetTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_Logout_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockResetTokenRepo := new(MockPasswordResetTokenRepository)
	mockRoleRepo := new(MockRoleRepository)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/home/ilmannafi/Documents/invento-be/keys/private.pem",
			PublicKeyPath:           "/home/ilmannafi/Documents/invento-be/keys/public.pem",
			PrivateKeyRotationPath:  "/home/ilmannafi/Documents/invento-be/keys/private_rotation.pem",
			PublicKeyRotationPath:   "/home/ilmannafi/Documents/invento-be/keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
	}
	
	if _, err := helper.NewJWTManager(cfg); err != nil {
		t.Skip("Skipping test due to missing JWT keys")
	}

	authUC := usecase.NewAuthUsecase(mockUserRepo, mockRefreshTokenRepo, mockResetTokenRepo, mockRoleRepo, cfg)

	token := "refresh_token_to_revoke"
	hashedToken := helper.HashRefreshToken(token)

	mockRefreshTokenRepo.On("RevokeToken", hashedToken).Return(nil)

	err := authUC.Logout(token)

	assert.NoError(t, err)

	mockRefreshTokenRepo.AssertExpectations(t)
}
