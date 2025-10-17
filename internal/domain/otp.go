package domain

import "time"

type OTPType string

const (
	OTPTypeRegister      OTPType = "register"
	OTPTypeResetPassword OTPType = "reset_password"
)

type OTP struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Email        string     `json:"email" gorm:"not null;index"`
	UserName     string     `json:"user_name" gorm:"default:''"`
	PasswordHash string     `json:"password_hash" gorm:"default:''"`
	Code         string     `json:"code" gorm:"not null"`
	CodeHash     string     `json:"-" gorm:"not null"`
	Type         OTPType    `json:"type" gorm:"not null;index"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null;index"`
	Attempts     int        `json:"attempts" gorm:"default:0"`
	MaxAttempts  int        `json:"max_attempts" gorm:"default:5"`
	IsUsed       bool       `json:"is_used" gorm:"default:false"`
	ResendCount  int        `json:"resend_count" gorm:"default:0"`
	LastResendAt *time.Time `json:"last_resend_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
	Type  string `json:"type" validate:"required,oneof=register reset_password"`
}

type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"required,oneof=register reset_password"`
}

type ConfirmResetPasswordOTPRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Code        string `json:"code" validate:"required,len=6,numeric"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type OTPResponse struct {
	Message     string `json:"message"`
	ExpiresIn   int    `json:"expires_in"`
	UserID      *uint  `json:"user_id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
}
