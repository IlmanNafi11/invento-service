package repo

import (
	"fiber-boiler-plate/internal/domain"
	"time"

	"gorm.io/gorm"
)

type OTPRepository interface {
	Create(email string, userName string, passwordHash string, codeHash string, otpType domain.OTPType, expiresAt time.Time, maxAttempts int) (*domain.OTP, error)
	GetByEmail(email string, otpType domain.OTPType) (*domain.OTP, error)
	IncrementAttempts(id uint) error
	MarkAsUsed(id uint) error
	DeleteByEmail(email string, otpType domain.OTPType) error
	UpdateResendInfo(id uint, resendCount int, lastResendAt time.Time) error
	DeleteExpired() error
}

type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &otpRepository{
		db: db,
	}
}

func (r *otpRepository) Create(email string, userName string, passwordHash string, codeHash string, otpType domain.OTPType, expiresAt time.Time, maxAttempts int) (*domain.OTP, error) {
	otp := &domain.OTP{
		Email:        email,
		UserName:     userName,
		PasswordHash: passwordHash,
		CodeHash:     codeHash,
		Type:         otpType,
		ExpiresAt:    expiresAt,
		MaxAttempts:  maxAttempts,
		Attempts:     0,
		IsUsed:       false,
	}

	err := r.db.Create(otp).Error
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (r *otpRepository) GetByEmail(email string, otpType domain.OTPType) (*domain.OTP, error) {
	var otp domain.OTP
	err := r.db.Where("email = ? AND type = ? AND is_used = ? AND expires_at > ?", email, otpType, false, time.Now()).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (r *otpRepository) IncrementAttempts(id uint) error {
	return r.db.Model(&domain.OTP{}).Where("id = ?", id).Update("attempts", gorm.Expr("attempts + ?", 1)).Error
}

func (r *otpRepository) MarkAsUsed(id uint) error {
	return r.db.Model(&domain.OTP{}).Where("id = ?", id).Update("is_used", true).Error
}

func (r *otpRepository) DeleteByEmail(email string, otpType domain.OTPType) error {
	return r.db.Where("email = ? AND type = ?", email, otpType).Delete(&domain.OTP{}).Error
}

func (r *otpRepository) UpdateResendInfo(id uint, resendCount int, lastResendAt time.Time) error {
	return r.db.Model(&domain.OTP{}).Where("id = ?", id).Updates(map[string]interface{}{
		"resend_count":   resendCount,
		"last_resend_at": lastResendAt,
	}).Error
}

func (r *otpRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ? OR is_used = ?", time.Now(), true).Delete(&domain.OTP{}).Error
}
