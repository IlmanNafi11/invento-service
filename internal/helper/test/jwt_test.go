package helper_test

import (
	"fiber-boiler-plate/internal/helper"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessToken_Success(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	var roleID *uint
	role := "mahasiswa"
	secret := "test-secret"
	expireHours := 1

	token, err := helper.GenerateAccessToken(userID, email, roleID, role, secret, expireHours)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".")
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	token, err := helper.GenerateRefreshToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 64)
}

func TestGenerateResetToken_Success(t *testing.T) {
	token, err := helper.GenerateResetToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 32)
}

func TestValidateAccessToken_Success(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	var roleID *uint
	role := "mahasiswa"
	secret := "test-secret"
	expireHours := 1

	token, err := helper.GenerateAccessToken(userID, email, roleID, role, secret, expireHours)
	assert.NoError(t, err)

	claims, err := helper.ValidateAccessToken(token, secret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, "access", claims.TokenType)
}

func TestValidateAccessToken_InvalidSecret(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	var roleID *uint
	role := "mahasiswa"
	secret := "test-secret"
	wrongSecret := "wrong-secret"
	expireHours := 1

	token, err := helper.GenerateAccessToken(userID, email, roleID, role, secret, expireHours)
	assert.NoError(t, err)

	claims, err := helper.ValidateAccessToken(token, wrongSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	userID := uint(1)
	email := "test@example.com"
	var roleID *uint
	role := "mahasiswa"
	secret := "test-secret"
	expireHours := -1

	token, err := helper.GenerateAccessToken(userID, email, roleID, role, secret, expireHours)
	assert.NoError(t, err)

	time.Sleep(time.Second * 1)

	claims, err := helper.ValidateAccessToken(token, secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTClaims_Structure(t *testing.T) {
	claims := helper.JWTClaims{
		UserID:    1,
		Email:     "test@example.com",
		TokenType: "access",
	}

	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "access", claims.TokenType)
}

func TestGenerateAccessToken_DifferentUsers(t *testing.T) {
	secret := "test-secret"
	expireHours := 1
	var roleID1 *uint
	var roleID2 *uint
	role1 := "mahasiswa"
	role2 := "dosen"

	token1, err1 := helper.GenerateAccessToken(1, "user1@example.com", roleID1, role1, secret, expireHours)
	token2, err2 := helper.GenerateAccessToken(2, "user2@example.com", roleID2, role2, secret, expireHours)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2)

	claims1, err1 := helper.ValidateAccessToken(token1, secret)
	claims2, err2 := helper.ValidateAccessToken(token2, secret)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, claims1.UserID, claims2.UserID)
	assert.NotEqual(t, claims1.Email, claims2.Email)
}

func TestGenerateRefreshToken_Uniqueness(t *testing.T) {
	token1, err1 := helper.GenerateRefreshToken()
	token2, err2 := helper.GenerateRefreshToken()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2)
}

func TestGenerateResetToken_Uniqueness(t *testing.T) {
	token1, err1 := helper.GenerateResetToken()
	token2, err2 := helper.GenerateResetToken()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2)
}
