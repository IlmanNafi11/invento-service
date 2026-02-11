package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPasswordResetTokenRepository_Create_Success tests successful password reset token creation
func TestPasswordResetTokenRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewPasswordResetTokenRepository(db)
	email := "test@example.com"
	token := "reset_token_123"
	expiresAt := time.Now().Add(time.Hour)

	result, err := repo.Create(email, token, expiresAt)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, token, result.Token)
	assert.False(t, result.IsUsed)
}

// TestPasswordResetTokenRepository_GetByToken_Success tests successful token retrieval
func TestPasswordResetTokenRepository_GetByToken_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	token := "reset_token_123"
	expiresAt := time.Now().Add(time.Hour)

	resetToken := &domain.PasswordResetToken{
		Email:     "test@example.com",
		Token:     token,
		ExpiresAt: expiresAt,
		IsUsed:    false,
	}
	err = db.Create(resetToken).Error
	require.NoError(t, err)

	repo := NewPasswordResetTokenRepository(db)
	result, err := repo.GetByToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, token, result.Token)
	assert.False(t, result.IsUsed)
}

// TestPasswordResetTokenRepository_MarkAsUsed_Success tests successful marking as used
func TestPasswordResetTokenRepository_MarkAsUsed_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	token := "reset_token_123"
	expiresAt := time.Now().Add(time.Hour)

	resetToken := &domain.PasswordResetToken{
		Email:     "test@example.com",
		Token:     token,
		ExpiresAt: expiresAt,
		IsUsed:    false,
	}
	err = db.Create(resetToken).Error
	require.NoError(t, err)

	repo := NewPasswordResetTokenRepository(db)
	err = repo.MarkAsUsed(token)
	assert.NoError(t, err)

	// Verify marked as used
	var updatedToken domain.PasswordResetToken
	err = db.Where("token = ?", token).First(&updatedToken).Error
	assert.NoError(t, err)
	assert.True(t, updatedToken.IsUsed)
}

// TestPasswordResetTokenRepository_CleanupExpired_Success tests successful cleanup
func TestPasswordResetTokenRepository_CleanupExpired_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	now := time.Now()

	expiredToken := &domain.PasswordResetToken{
		Email:     "expired@example.com",
		Token:     "expired_token",
		ExpiresAt: now.Add(-1 * time.Hour),
		IsUsed:    false,
	}
	err = db.Create(expiredToken).Error
	require.NoError(t, err)

	usedToken := &domain.PasswordResetToken{
		Email:     "used@example.com",
		Token:     "used_token",
		ExpiresAt: now.Add(1 * time.Hour),
		IsUsed:    true,
	}
	err = db.Create(usedToken).Error
	require.NoError(t, err)

	validToken := &domain.PasswordResetToken{
		Email:     "valid@example.com",
		Token:     "valid_token",
		ExpiresAt: now.Add(1 * time.Hour),
		IsUsed:    false,
	}
	err = db.Create(validToken).Error
	require.NoError(t, err)

	repo := NewPasswordResetTokenRepository(db)
	err = repo.CleanupExpired()
	assert.NoError(t, err)

	// Verify only valid token remains
	var count int64
	db.Model(&domain.PasswordResetToken{}).Count(&count)
	assert.Equal(t, int64(1), count)

	var remainingToken domain.PasswordResetToken
	err = db.First(&remainingToken).Error
	assert.NoError(t, err)
	assert.Equal(t, "valid_token", remainingToken.Token)
}
