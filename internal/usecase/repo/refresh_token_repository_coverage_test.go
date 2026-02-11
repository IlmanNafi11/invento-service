package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRefreshTokenRepository_Create_Success tests successful refresh token creation
func TestRefreshTokenRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewRefreshTokenRepository(db)
	userID := uint(1)
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	result, err := repo.Create(userID, token, expiresAt)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, token, result.Token)
	assert.False(t, result.IsRevoked)
}

// TestRefreshTokenRepository_GetByToken_Success tests successful refresh token retrieval
func TestRefreshTokenRepository_GetByToken_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	refreshToken := &domain.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}
	err = db.Create(refreshToken).Error
	require.NoError(t, err)

	repo := NewRefreshTokenRepository(db)
	result, err := repo.GetByToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, token, result.Token)
	assert.Equal(t, userID, result.UserID)
	assert.False(t, result.IsRevoked)
}

// TestRefreshTokenRepository_RevokeToken_Success tests successful token revocation
func TestRefreshTokenRepository_RevokeToken_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	refreshToken := &domain.RefreshToken{
		UserID:    1,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}
	err = db.Create(refreshToken).Error
	require.NoError(t, err)

	repo := NewRefreshTokenRepository(db)
	err = repo.RevokeToken(token)
	assert.NoError(t, err)

	// Verify revoked
	var revokedToken domain.RefreshToken
	err = db.Where("token = ?", token).First(&revokedToken).Error
	assert.NoError(t, err)
	assert.True(t, revokedToken.IsRevoked)
}

// TestRefreshTokenRepository_RevokeAllUserTokens_Success tests successful revocation of all user tokens
func TestRefreshTokenRepository_RevokeAllUserTokens_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)
	expiresAt := time.Now().Add(24 * time.Hour)

	tokens := []domain.RefreshToken{
		{
			UserID:    userID,
			Token:     "token1",
			ExpiresAt: expiresAt,
			IsRevoked: false,
		},
		{
			UserID:    userID,
			Token:     "token2",
			ExpiresAt: expiresAt,
			IsRevoked: false,
		},
		{
			UserID:    2, // Different user
			Token:     "token3",
			ExpiresAt: expiresAt,
			IsRevoked: false,
		},
	}

	for _, token := range tokens {
		err := db.Create(&token).Error
		require.NoError(t, err)
	}

	repo := NewRefreshTokenRepository(db)
	err = repo.RevokeAllUserTokens(userID)
	assert.NoError(t, err)

	// Verify user's tokens revoked
	var revokedCount int64
	db.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, true).
		Count(&revokedCount)
	assert.Equal(t, int64(2), revokedCount)

	// Verify other user's token not revoked
	var otherToken domain.RefreshToken
	err = db.Where("token = ?", "token3").First(&otherToken).Error
	assert.NoError(t, err)
	assert.False(t, otherToken.IsRevoked)
}

// TestRefreshTokenRepository_CleanupExpired_Success tests successful cleanup of expired tokens
func TestRefreshTokenRepository_CleanupExpired_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	now := time.Now()

	expiredToken := &domain.RefreshToken{
		UserID:    1,
		Token:     "expired_token",
		ExpiresAt: now.Add(-1 * time.Hour),
		IsRevoked: false,
	}
	err = db.Create(expiredToken).Error
	require.NoError(t, err)

	revokedToken := &domain.RefreshToken{
		UserID:    1,
		Token:     "revoked_token",
		ExpiresAt: now.Add(1 * time.Hour),
		IsRevoked: true,
	}
	err = db.Create(revokedToken).Error
	require.NoError(t, err)

	validToken := &domain.RefreshToken{
		UserID:    1,
		Token:     "valid_token",
		ExpiresAt: now.Add(1 * time.Hour),
		IsRevoked: false,
	}
	err = db.Create(validToken).Error
	require.NoError(t, err)

	repo := NewRefreshTokenRepository(db)
	err = repo.CleanupExpired()
	assert.NoError(t, err)

	// Verify only valid token remains
	var count int64
	db.Model(&domain.RefreshToken{}).Count(&count)
	assert.Equal(t, int64(1), count)

	var remainingToken domain.RefreshToken
	err = db.First(&remainingToken).Error
	assert.NoError(t, err)
	assert.Equal(t, "valid_token", remainingToken.Token)
}
