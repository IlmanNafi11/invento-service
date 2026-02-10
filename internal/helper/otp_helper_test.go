package helper_test

import (
	"fiber-boiler-plate/internal/helper"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{"4 digit OTP", 4, false},
		{"6 digit OTP", 6, false},
		{"8 digit OTP", 8, false},
		{"1 digit OTP", 1, false},
		{"12 digit OTP", 12, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := helper.GenerateOTP(tt.length)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, otp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.length, len(otp))
				assert.True(t, strings.HasPrefix(otp, "0") || strings.HasPrefix(otp, "1") ||
					strings.HasPrefix(otp, "2") || strings.HasPrefix(otp, "3") ||
					strings.HasPrefix(otp, "4") || strings.HasPrefix(otp, "5") ||
					strings.HasPrefix(otp, "6") || strings.HasPrefix(otp, "7") ||
					strings.HasPrefix(otp, "8") || strings.HasPrefix(otp, "9"))
			}
		})
	}
}

func TestGenerateOTP_Uniqueness(t *testing.T) {
	otp1, err1 := helper.GenerateOTP(6)
	otp2, err2 := helper.GenerateOTP(6)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, otp1, otp2, "OTPs should be unique")
}

func TestHashOTP(t *testing.T) {
	otp := "123456"

	hash := helper.HashOTP(otp)

	assert.NotEmpty(t, hash)
	assert.NotEqual(t, otp, hash)
	assert.Equal(t, 64, len(hash)) // SHA256 produces 64 character hex string
}

func TestHashOTP_Consistency(t *testing.T) {
	otp := "654321"

	hash1 := helper.HashOTP(otp)
	hash2 := helper.HashOTP(otp)

	assert.Equal(t, hash1, hash2, "Same OTP should produce same hash")
}

func TestVerifyOTPHash_Valid(t *testing.T) {
	otp := "987654"

	hash := helper.HashOTP(otp)
	isValid := helper.VerifyOTPHash(otp, hash)

	assert.True(t, isValid)
}

func TestVerifyOTPHash_Invalid(t *testing.T) {
	otp := "111111"
	wrongOTP := "222222"

	hash := helper.HashOTP(otp)
	isValid := helper.VerifyOTPHash(wrongOTP, hash)

	assert.False(t, isValid)
}

func TestNewOTPValidator(t *testing.T) {
	tests := []struct {
		name          string
		maxAttempts   int
		expiryMinutes int
		expectedMax   int
		expectedExpiry int
	}{
		{
			name:          "valid parameters",
			maxAttempts:   3,
			expiryMinutes: 15,
			expectedMax:   3,
			expectedExpiry: 15,
		},
		{
			name:          "zero max attempts defaults to 5",
			maxAttempts:   0,
			expiryMinutes: 10,
			expectedMax:   5,
			expectedExpiry: 10,
		},
		{
			name:          "negative max attempts defaults to 5",
			maxAttempts:   -1,
			expiryMinutes: 10,
			expectedMax:   5,
			expectedExpiry: 10,
		},
		{
			name:          "zero expiry defaults to 10",
			maxAttempts:   3,
			expiryMinutes: 0,
			expectedMax:   3,
			expectedExpiry: 10,
		},
		{
			name:          "negative expiry defaults to 10",
			maxAttempts:   3,
			expiryMinutes: -5,
			expectedMax:   3,
			expectedExpiry: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := helper.NewOTPValidator(tt.maxAttempts, tt.expiryMinutes)

			assert.Equal(t, tt.expectedMax, validator.GetMaxAttempts())
			assert.Equal(t, tt.expectedExpiry, validator.GetExpiryMinutes())
		})
	}
}

func TestOTPValidator_ValidateOTPRecord(t *testing.T) {
	validator := helper.NewOTPValidator(5, 10)

	tests := []struct {
		name        string
		attempts    int
		maxAttempts int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "within limit",
			attempts:    0,
			maxAttempts: 5,
			expectError: false,
		},
		{
			name:        "at limit",
			attempts:    4,
			maxAttempts: 5,
			expectError: false,
		},
		{
			name:        "exceeds limit",
			attempts:    5,
			maxAttempts: 5,
			expectError: true,
			errorMsg:    "kode otp telah melampaui batas percobaan maksimal",
		},
		{
			name:        "far exceeds limit",
			attempts:    10,
			maxAttempts: 5,
			expectError: true,
			errorMsg:    "kode otp telah melampaui batas percobaan maksimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOTPRecord(tt.attempts, tt.maxAttempts)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewResendRateLimiter(t *testing.T) {
	tests := []struct {
		name             string
		maxResendTimes   int
		cooldownSecs     int
		expectedMax      int
		expectedCooldown int
	}{
		{
			name:             "valid parameters",
			maxResendTimes:   3,
			cooldownSecs:     60,
			expectedMax:      3,
			expectedCooldown: 60,
		},
		{
			name:             "zero resend times defaults to 5",
			maxResendTimes:   0,
			cooldownSecs:     30,
			expectedMax:      5,
			expectedCooldown: 30,
		},
		{
			name:             "negative resend times defaults to 5",
			maxResendTimes:   -1,
			cooldownSecs:     30,
			expectedMax:      5,
			expectedCooldown: 30,
		},
		{
			name:             "zero cooldown defaults to 60",
			maxResendTimes:   3,
			cooldownSecs:     0,
			expectedMax:      3,
			expectedCooldown: 60,
		},
		{
			name:             "negative cooldown defaults to 60",
			maxResendTimes:   3,
			cooldownSecs:     -10,
			expectedMax:      3,
			expectedCooldown: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := helper.NewResendRateLimiter(tt.maxResendTimes, tt.cooldownSecs)

			assert.Equal(t, tt.expectedMax, limiter.GetMaxResendTimes())
			assert.Equal(t, tt.expectedCooldown, limiter.GetCooldownSeconds())
		})
	}
}

func TestResendRateLimiter_CanResend(t *testing.T) {
	limiter := helper.NewResendRateLimiter(3, 60)

	tests := []struct {
		name         string
		resendCount  int
		lastResendAt *time.Time
		expectCan    bool
		expectWait   time.Duration
	}{
		{
			name:         "first resend, no previous",
			resendCount:  0,
			lastResendAt: nil,
			expectCan:    true,
			expectWait:   0,
		},
		{
			name:         "within limit, cooldown passed",
			resendCount:  1,
			lastResendAt: func() *time.Time { t := time.Now().Add(-61 * time.Second); return &t }(),
			expectCan:    true,
			expectWait:   0,
		},
		{
			name:         "within limit, cooldown not passed",
			resendCount:  1,
			lastResendAt: func() *time.Time { t := time.Now().Add(-30 * time.Second); return &t }(),
			expectCan:    false,
			expectWait:   func() time.Duration { d, _ := time.ParseDuration("30s"); return d }(),
		},
		{
			name:         "at limit",
			resendCount:  2,
			lastResendAt: nil,
			expectCan:    true,
			expectWait:   0,
		},
		{
			name:         "exceeds limit",
			resendCount:  3,
			lastResendAt: nil,
			expectCan:    false,
			expectWait:   0,
		},
		{
			name:         "far exceeds limit",
			resendCount:  10,
			lastResendAt: nil,
			expectCan:    false,
			expectWait:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canResend, waitTime := limiter.CanResend(tt.resendCount, tt.lastResendAt)

			assert.Equal(t, tt.expectCan, canResend)

			if tt.expectWait > 0 {
				// Allow small tolerance for timing differences
				assert.InDelta(t, tt.expectWait.Seconds(), waitTime.Seconds(), 2)
			} else {
				assert.Equal(t, tt.expectWait, waitTime)
			}
		})
	}
}

func TestResendRateLimiter_CanResend_NoLastResend(t *testing.T) {
	limiter := helper.NewResendRateLimiter(5, 60)

	// Test with resend count but no last resend time
	canResend, waitTime := limiter.CanResend(2, nil)

	assert.True(t, canResend)
	assert.Equal(t, time.Duration(0), waitTime)
}

func TestResendRateLimiter_Getters(t *testing.T) {
	limiter := helper.NewResendRateLimiter(3, 120)

	assert.Equal(t, 3, limiter.GetMaxResendTimes())
	assert.Equal(t, 120, limiter.GetCooldownSeconds())
}
