package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"
)

func GenerateOTP(length int) (string, error) {
	otp := ""
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += fmt.Sprintf("%d", num.Int64())
	}
	return otp, nil
}

func HashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
}

func VerifyOTPHash(otp string, hash string) bool {
	return HashOTP(otp) == hash
}

type OTPValidator struct {
	maxAttempts   int
	expiryMinutes int
}

func NewOTPValidator(maxAttempts int, expiryMinutes int) *OTPValidator {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if expiryMinutes <= 0 {
		expiryMinutes = 10
	}
	return &OTPValidator{
		maxAttempts:   maxAttempts,
		expiryMinutes: expiryMinutes,
	}
}

func (ov *OTPValidator) ValidateOTPRecord(attempts int, maxAttempts int) error {
	if attempts >= maxAttempts {
		return errors.New("kode otp telah melampaui batas percobaan maksimal")
	}
	return nil
}

func (ov *OTPValidator) GetMaxAttempts() int {
	return ov.maxAttempts
}

func (ov *OTPValidator) GetExpiryMinutes() int {
	return ov.expiryMinutes
}

// ResendRateLimiter handles rate limiting for OTP resends
type ResendRateLimiter struct {
	maxResendTimes int
	cooldownSecs   int
}

// NewResendRateLimiter creates a new resend rate limiter
func NewResendRateLimiter(maxResendTimes int, cooldownSecs int) *ResendRateLimiter {
	if maxResendTimes <= 0 {
		maxResendTimes = 5
	}
	if cooldownSecs <= 0 {
		cooldownSecs = 60
	}
	return &ResendRateLimiter{
		maxResendTimes: maxResendTimes,
		cooldownSecs:   cooldownSecs,
	}
}

// CanResend checks if OTP can be resent based on count and last resend time
func (rl *ResendRateLimiter) CanResend(resendCount int, lastResendAt *time.Time) (bool, time.Duration) {
	if resendCount >= rl.maxResendTimes {
		return false, 0
	}

	if lastResendAt != nil {
		elapsed := time.Since(*lastResendAt)
		cooldown := time.Duration(rl.cooldownSecs) * time.Second
		if elapsed < cooldown {
			remaining := cooldown - elapsed
			return false, remaining
		}
	}

	return true, 0
}

// GetMaxResendTimes returns the maximum number of resends allowed
func (rl *ResendRateLimiter) GetMaxResendTimes() int {
	return rl.maxResendTimes
}

// GetCooldownSeconds returns the cooldown period in seconds
func (rl *ResendRateLimiter) GetCooldownSeconds() int {
	return rl.cooldownSecs
}
