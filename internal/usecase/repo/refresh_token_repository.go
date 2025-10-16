package repo

import (
	"fiber-boiler-plate/internal/domain"
	"time"

	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(userID uint, token string, expiresAt time.Time) (*domain.RefreshToken, error) {
	refreshToken := &domain.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	err := r.db.Create(refreshToken).Error
	if err != nil {
		return nil, err
	}
	return refreshToken, nil
}

func (r *refreshTokenRepository) GetByToken(token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) RevokeToken(token string) error {
	return r.db.Model(&domain.RefreshToken{}).Where("token = ?", token).Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) RevokeAllUserTokens(userID uint) error {
	return r.db.Model(&domain.RefreshToken{}).Where("user_id = ?", userID).Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) CleanupExpired() error {
	return r.db.Where("expires_at < ? OR is_revoked = ?", time.Now(), true).Delete(&domain.RefreshToken{}).Error
}
