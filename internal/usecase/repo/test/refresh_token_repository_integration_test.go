package repo_test

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"fiber-boiler-plate/internal/usecase/repo"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// RefreshTokenRepositoryTestSuite handles all refresh token repository tests
type RefreshTokenRepositoryTestSuite struct {
	suite.Suite
	db    *gorm.DB
	repo  repo.RefreshTokenRepository
}

func (suite *RefreshTokenRepositoryTestSuite) SetupTest() {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(suite.T(), err)
	suite.db = db
	suite.repo = repo.NewRefreshTokenRepository(db)
}

func (suite *RefreshTokenRepositoryTestSuite) TearDownTest() {
	testhelper.TeardownTestDatabase(suite.db)
}

func (suite *RefreshTokenRepositoryTestSuite) TestCreate_Success() {
	userID := uint(1)
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	result, err := suite.repo.Create(userID, token, expiresAt)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), userID, result.UserID)
	assert.Equal(suite.T(), token, result.Token)
	assert.False(suite.T(), result.IsRevoked)
}

func (suite *RefreshTokenRepositoryTestSuite) TestGetByToken_Success() {
	userID := uint(1)
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	refreshToken := &domain.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}
	err := suite.db.Create(refreshToken).Error
	require.NoError(suite.T(), err)

	result, err := suite.repo.GetByToken(token)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), token, result.Token)
	assert.Equal(suite.T(), userID, result.UserID)
	assert.False(suite.T(), result.IsRevoked)
}

func (suite *RefreshTokenRepositoryTestSuite) TestGetByToken_NotFound() {
	result, err := suite.repo.GetByToken("nonexistent")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *RefreshTokenRepositoryTestSuite) TestGetByToken_Expired() {
	token := "expired_token"
	expiresAt := time.Now().Add(-1 * time.Hour)

	refreshToken := &domain.RefreshToken{
		UserID:    1,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}
	err := suite.db.Create(refreshToken).Error
	require.NoError(suite.T(), err)

	result, err := suite.repo.GetByToken(token)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *RefreshTokenRepositoryTestSuite) TestRevokeToken_Success() {
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	refreshToken := &domain.RefreshToken{
		UserID:    1,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}
	err := suite.db.Create(refreshToken).Error
	require.NoError(suite.T(), err)

	err = suite.repo.RevokeToken(token)
	assert.NoError(suite.T(), err)

	// Verify revoked
	var revokedToken domain.RefreshToken
	err = suite.db.Where("token = ?", token).First(&revokedToken).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), revokedToken.IsRevoked)
}

func (suite *RefreshTokenRepositoryTestSuite) TestRevokeAllUserTokens_Success() {
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
		err := suite.db.Create(&token).Error
		require.NoError(suite.T(), err)
	}

	err := suite.repo.RevokeAllUserTokens(userID)
	assert.NoError(suite.T(), err)

	// Verify user's tokens revoked
	var revokedCount int64
	suite.db.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, true).
		Count(&revokedCount)
	assert.Equal(suite.T(), int64(2), revokedCount)

	// Verify other user's token not revoked
	var otherToken domain.RefreshToken
	err = suite.db.Where("token = ?", "token3").First(&otherToken).Error
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), otherToken.IsRevoked)
}

func (suite *RefreshTokenRepositoryTestSuite) TestCleanupExpired_Success() {
	now := time.Now()

	expiredToken := &domain.RefreshToken{
		UserID:    1,
		Token:     "expired_token",
		ExpiresAt: now.Add(-1 * time.Hour),
		IsRevoked: false,
	}
	err := suite.db.Create(expiredToken).Error
	require.NoError(suite.T(), err)

	revokedToken := &domain.RefreshToken{
		UserID:    1,
		Token:     "revoked_token",
		ExpiresAt: now.Add(1 * time.Hour),
		IsRevoked: true,
	}
	err = suite.db.Create(revokedToken).Error
	require.NoError(suite.T(), err)

	validToken := &domain.RefreshToken{
		UserID:    1,
		Token:     "valid_token",
		ExpiresAt: now.Add(1 * time.Hour),
		IsRevoked: false,
	}
	err = suite.db.Create(validToken).Error
	require.NoError(suite.T(), err)

	err = suite.repo.CleanupExpired()
	assert.NoError(suite.T(), err)

	// Verify only valid token remains
	var count int64
	suite.db.Model(&domain.RefreshToken{}).Count(&count)
	assert.Equal(suite.T(), int64(1), count)

	var remainingToken domain.RefreshToken
	err = suite.db.First(&remainingToken).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "valid_token", remainingToken.Token)
}

func TestRefreshTokenRepositorySuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenRepositoryTestSuite))
}
