package helper

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/usecase/repo"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthHelper struct {
	refreshTokenRepo repo.RefreshTokenRepository
	config           *config.Config
}

func NewAuthHelper(refreshTokenRepo repo.RefreshTokenRepository, config *config.Config) *AuthHelper {
	return &AuthHelper{
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
	}
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("gagal mengenkripsi password")
	}

	return string(hashedPassword), nil
}

func ComparePassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return errors.New("email atau password salah")
	}

	return nil
}

func GenerateTokenPair(user *domain.User, secret string, expireHours int, refreshExpireHours int, refreshTokenRepo repo.RefreshTokenRepository) (string, string, error) {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	accessToken, err := GenerateAccessToken(user.ID, user.Email, user.RoleID, roleName, secret, expireHours)
	if err != nil {
		return "", "", errors.New("gagal generate access token")
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return "", "", errors.New("gagal generate refresh token")
	}

	refreshTokenExpiry := time.Now().Add(time.Hour * time.Duration(refreshExpireHours))
	if _, err := refreshTokenRepo.Create(user.ID, refreshToken, refreshTokenExpiry); err != nil {
		return "", "", errors.New("gagal simpan refresh token")
	}

	return accessToken, refreshToken, nil
}

func BuildAuthResponse(user *domain.User, accessToken string, refreshToken string, expiresIn int) *domain.AuthResponse {
	return &domain.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}
}

func BuildRefreshTokenResponse(accessToken string, refreshToken string, expiresIn int) *domain.RefreshTokenResponse {
	return &domain.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}
}

func (ah *AuthHelper) GenerateAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	accessToken, refreshToken, err := GenerateTokenPair(
		user,
		ah.config.JWT.Secret,
		ah.config.JWT.ExpireHours,
		ah.config.JWT.RefreshTokenExpireHours,
		ah.refreshTokenRepo,
	)
	if err != nil {
		return nil, err
	}

	expiresIn := ah.config.JWT.ExpireHours * 3600
	return BuildAuthResponse(user, accessToken, refreshToken, expiresIn), nil
}

func (ah *AuthHelper) RevokeAndGenerateNewTokens(oldRefreshToken string, user *domain.User) (*domain.RefreshTokenResponse, error) {
	if err := ah.refreshTokenRepo.RevokeToken(oldRefreshToken); err != nil {
		return nil, errors.New("gagal revoke refresh token lama")
	}

	accessToken, refreshToken, err := GenerateTokenPair(
		user,
		ah.config.JWT.Secret,
		ah.config.JWT.ExpireHours,
		ah.config.JWT.RefreshTokenExpireHours,
		ah.refreshTokenRepo,
	)
	if err != nil {
		return nil, err
	}

	expiresIn := ah.config.JWT.ExpireHours * 3600
	return BuildRefreshTokenResponse(accessToken, refreshToken, expiresIn), nil
}
