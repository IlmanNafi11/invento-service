package helper_test

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRefreshTokenRepository for testing
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

func (m *MockRefreshTokenRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpiredTokens() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) CleanupExpired() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllUserTokens(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestHashPassword(t *testing.T) {
	password := "SecurePassword123!"

	hashedPassword, err := helper.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
	// Bcrypt hashes are typically 60 characters
	assert.Len(t, hashedPassword, 60)
}

func TestHashPassword_DifferentPasswords(t *testing.T) {
	password1 := "Password123"
	password2 := "DifferentPassword456"

	hash1, err1 := helper.HashPassword(password1)
	hash2, err2 := helper.HashPassword(password2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2)
}

func TestComparePassword_Valid(t *testing.T) {
	password := "CorrectPassword123"

	hashedPassword, err := helper.HashPassword(password)
	assert.NoError(t, err)

	err = helper.ComparePassword(hashedPassword, password)
	assert.NoError(t, err)
}

func TestComparePassword_Invalid(t *testing.T) {
	password := "CorrectPassword123"
	wrongPassword := "WrongPassword456"

	hashedPassword, err := helper.HashPassword(password)
	assert.NoError(t, err)

	err = helper.ComparePassword(hashedPassword, wrongPassword)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email atau password salah")
}

func TestComparePassword_EmptyPassword(t *testing.T) {
	hashedPassword := "$2a$10$abcdefghijklmnopqrstuvwxyz"
	emptyPassword := ""

	err := helper.ComparePassword(hashedPassword, emptyPassword)
	assert.Error(t, err)
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := helper.GenerateRefreshToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	// Refresh tokens are 64 hex characters (32 bytes)
	assert.Len(t, token, 64)
}

func TestGenerateRefreshToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)

	// Generate 100 tokens and verify uniqueness
	for i := 0; i < 100; i++ {
		token, err := helper.GenerateRefreshToken()
		assert.NoError(t, err)
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
		assert.Len(t, token, 64)
	}
}

func TestHashRefreshToken(t *testing.T) {
	token := "abc123def456"

	hashed := helper.HashRefreshToken(token)

	assert.NotEmpty(t, hashed)
	assert.NotEqual(t, token, hashed)
	// SHA256 base64 encoded hash is typically 44 characters
	assert.Len(t, hashed, 44)
}

func TestHashRefreshToken_Consistency(t *testing.T) {
	token := "test-token-123"

	hash1 := helper.HashRefreshToken(token)
	hash2 := helper.HashRefreshToken(token)

	assert.Equal(t, hash1, hash2, "Hashing same token should produce same result")
}

func TestGenerateResetToken(t *testing.T) {
	token, err := helper.GenerateResetToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	// Reset tokens are 32 hex characters (16 bytes)
	assert.Len(t, token, 32)
}

func TestGenerateResetToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)

	// Generate 100 tokens and verify uniqueness
	for i := 0; i < 100; i++ {
		token, err := helper.GenerateResetToken()
		assert.NoError(t, err)
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
		assert.Len(t, token, 32)
	}
}

func TestBuildAuthResponse(t *testing.T) {
	user := &domain.User{
		ID:    1,
		Email: "test@example.com",
	}
	accessToken := "valid-access-token"
	expiresIn := 3600

	response := helper.BuildAuthResponse(user, accessToken, expiresIn)

	assert.NotNil(t, response)
	assert.Equal(t, accessToken, response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, expiresIn, response.ExpiresIn)
	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, user.Email, response.User.Email)
}

func TestBuildRefreshTokenResponse(t *testing.T) {
	accessToken := "new-access-token"
	expiresIn := 7200

	response := helper.BuildRefreshTokenResponse(accessToken, expiresIn)

	assert.NotNil(t, response)
	assert.Equal(t, accessToken, response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, expiresIn, response.ExpiresIn)
}

func TestAuthHelper_EmptyPassword(t *testing.T) {
	password := ""

	hashedPassword, err := helper.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// Empty password should still match after hashing
	err = helper.ComparePassword(hashedPassword, password)
	assert.NoError(t, err)
}

func TestAuthHelper_VariousPasswords(t *testing.T) {
	passwords := []string{
		"simple",
		"Complex123",
		strings.Repeat("a", 50), // 50 character password (within bcrypt 72 byte limit)
	}

	for _, password := range passwords {
		t.Run("", func(t *testing.T) {
			hashedPassword, err := helper.HashPassword(password)
			assert.NoError(t, err)

			err = helper.ComparePassword(hashedPassword, password)
			assert.NoError(t, err)

			// Wrong password should fail
			err = helper.ComparePassword(hashedPassword, password+"x")
			assert.Error(t, err)
		})
	}
}

func TestGenerateResetToken_DifferentFromRefreshToken(t *testing.T) {
	refreshToken, err1 := helper.GenerateRefreshToken()
	resetToken, err2 := helper.GenerateResetToken()

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Refresh token is 64 chars, reset token is 32 chars
	assert.NotEqual(t, len(refreshToken), len(resetToken))
	assert.Len(t, refreshToken, 64)
	assert.Len(t, resetToken, 32)
}

func TestHashRefreshToken_DifferentTokens(t *testing.T) {
	token1 := "token1"
	token2 := "token2"

	hash1 := helper.HashRefreshToken(token1)
	hash2 := helper.HashRefreshToken(token2)

	assert.NotEqual(t, hash1, hash2)
}
