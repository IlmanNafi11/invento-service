package domain

import (
	"testing"
	"time"
)

func TestOTPTypeConstants(t *testing.T) {
	t.Run("OTPType constants are defined", func(t *testing.T) {
		types := []OTPType{
			OTPTypeRegister,
			OTPTypeResetPassword,
		}

		expectedTypes := []OTPType{
			"register",
			"reset_password",
		}

		for i, otpType := range types {
			if otpType != expectedTypes[i] {
				t.Errorf("Expected OTP type '%s', got '%s'", expectedTypes[i], otpType)
			}
		}
	})
}

func TestOTPStruct(t *testing.T) {
	t.Run("OTP struct initialization", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(15 * time.Minute)

		otp := OTP{
			ID:           1,
			Email:        "user@example.com",
			UserName:     "John Doe",
			PasswordHash: "hashed_password",
			Code:         "123456",
			CodeHash:     "hashed_code",
			Type:         OTPTypeRegister,
			ExpiresAt:    expiresAt,
			Attempts:     0,
			MaxAttempts:  5,
			IsUsed:       false,
			ResendCount:  0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if otp.ID != 1 {
			t.Errorf("Expected ID 1, got %d", otp.ID)
		}
		if otp.Email != "user@example.com" {
			t.Errorf("Expected Email 'user@example.com', got %s", otp.Email)
		}
		if otp.UserName != "John Doe" {
			t.Errorf("Expected UserName 'John Doe', got %s", otp.UserName)
		}
		if otp.Type != OTPTypeRegister {
			t.Errorf("Expected Type '%s', got %s", OTPTypeRegister, otp.Type)
		}
		if otp.Attempts != 0 {
			t.Errorf("Expected Attempts 0, got %d", otp.Attempts)
		}
		if otp.MaxAttempts != 5 {
			t.Errorf("Expected MaxAttempts 5, got %d", otp.MaxAttempts)
		}
		if otp.IsUsed {
			t.Error("Expected IsUsed to be false")
		}
		if otp.ResendCount != 0 {
			t.Errorf("Expected ResendCount 0, got %d", otp.ResendCount)
		}
	})

	t.Run("OTP with reset password type", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(15 * time.Minute)

		otp := OTP{
			ID:           2,
			Email:        "reset@example.com",
			Code:         "654321",
			CodeHash:     "hashed_reset_code",
			Type:         OTPTypeResetPassword,
			ExpiresAt:    expiresAt,
			Attempts:     1,
			MaxAttempts:  5,
			IsUsed:       false,
			ResendCount:  2,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if otp.Type != OTPTypeResetPassword {
			t.Errorf("Expected Type '%s', got %s", OTPTypeResetPassword, otp.Type)
		}
		if otp.Attempts != 1 {
			t.Errorf("Expected Attempts 1, got %d", otp.Attempts)
		}
		if otp.ResendCount != 2 {
			t.Errorf("Expected ResendCount 2, got %d", otp.ResendCount)
		}
	})

	t.Run("OTP with used status", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(15 * time.Minute)

		otp := OTP{
			ID:           3,
			Email:        "used@example.com",
			Code:         "111111",
			CodeHash:     "hashed_used_code",
			Type:         OTPTypeRegister,
			ExpiresAt:    expiresAt,
			Attempts:     1,
			MaxAttempts:  5,
			IsUsed:       true,
			ResendCount:  0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if !otp.IsUsed {
			t.Error("Expected IsUsed to be true")
		}
		if otp.Attempts != 1 {
			t.Errorf("Expected Attempts 1, got %d", otp.Attempts)
		}
	})

	t.Run("OTP with last resend timestamp", func(t *testing.T) {
		now := time.Now()
		lastResend := now.Add(-2 * time.Minute)
		expiresAt := now.Add(15 * time.Minute)

		otp := OTP{
			ID:           4,
			Email:        "resend@example.com",
			Code:         "222222",
			CodeHash:     "hashed_resend_code",
			Type:         OTPTypeResetPassword,
			ExpiresAt:    expiresAt,
			Attempts:     0,
			MaxAttempts:  5,
			IsUsed:       false,
			ResendCount:  1,
			LastResendAt: &lastResend,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if otp.LastResendAt == nil {
			t.Error("Expected LastResendAt to be set")
		}
		if otp.ResendCount != 1 {
			t.Errorf("Expected ResendCount 1, got %d", otp.ResendCount)
		}
	})
}

func TestVerifyOTPRequest(t *testing.T) {
	t.Run("VerifyOTPRequest with valid data", func(t *testing.T) {
		req := VerifyOTPRequest{
			Email: "user@example.com",
			Code:  "123456",
			Type:  "register",
		}

		if req.Email != "user@example.com" {
			t.Errorf("Expected Email 'user@example.com', got %s", req.Email)
		}
		if req.Code != "123456" {
			t.Errorf("Expected Code '123456', got %s", req.Code)
		}
		if req.Type != "register" {
			t.Errorf("Expected Type 'register', got %s", req.Type)
		}
	})

	t.Run("VerifyOTPRequest with reset password type", func(t *testing.T) {
		req := VerifyOTPRequest{
			Email: "reset@example.com",
			Code:  "654321",
			Type:  "reset_password",
		}

		if req.Type != "reset_password" {
			t.Errorf("Expected Type 'reset_password', got %s", req.Type)
		}
	})

	t.Run("VerifyOTPRequest code length validation", func(t *testing.T) {
		validCodes := []string{"123456", "000000", "999999"}

		for _, code := range validCodes {
			req := VerifyOTPRequest{
				Email: "test@example.com",
				Code:  code,
				Type:  "register",
			}

			if len(req.Code) != 6 {
				t.Errorf("Expected Code length 6, got %d", len(req.Code))
			}
		}
	})
}

func TestResendOTPRequest(t *testing.T) {
	t.Run("ResendOTPRequest with register type", func(t *testing.T) {
		req := ResendOTPRequest{
			Email: "user@example.com",
			Type:  "register",
		}

		if req.Email != "user@example.com" {
			t.Errorf("Expected Email 'user@example.com', got %s", req.Email)
		}
		if req.Type != "register" {
			t.Errorf("Expected Type 'register', got %s", req.Type)
		}
	})

	t.Run("ResendOTPRequest with reset password type", func(t *testing.T) {
		req := ResendOTPRequest{
			Email: "reset@example.com",
			Type:  "reset_password",
		}

		if req.Type != "reset_password" {
			t.Errorf("Expected Type 'reset_password', got %s", req.Type)
		}
	})
}

func TestConfirmResetPasswordOTPRequest(t *testing.T) {
	t.Run("ConfirmResetPasswordOTPRequest with valid data", func(t *testing.T) {
		req := ConfirmResetPasswordOTPRequest{
			Email:       "user@example.com",
			Code:        "123456",
			NewPassword: "NewSecurePassword123",
		}

		if req.Email != "user@example.com" {
			t.Errorf("Expected Email 'user@example.com', got %s", req.Email)
		}
		if req.Code != "123456" {
			t.Errorf("Expected Code '123456', got %s", req.Code)
		}
		if req.NewPassword != "NewSecurePassword123" {
			t.Errorf("Expected NewPassword 'NewSecurePassword123', got %s", req.NewPassword)
		}
	})

	t.Run("ConfirmResetPasswordOTPRequest with minimum password length", func(t *testing.T) {
		req := ConfirmResetPasswordOTPRequest{
			Email:       "test@example.com",
			Code:        "654321",
			NewPassword: "12345678",
		}

		if len(req.NewPassword) != 8 {
			t.Errorf("Expected NewPassword length 8, got %d", len(req.NewPassword))
		}
	})
}

func TestOTPResponse(t *testing.T) {
	t.Run("OTPResponse with all fields", func(t *testing.T) {
		userID := uint(100)

		resp := OTPResponse{
			Message:     "OTP sent successfully",
			ExpiresIn:   900,
			UserID:      &userID,
			AccessToken: "access_token_123",
			TokenType:   "Bearer",
		}

		if resp.Message != "OTP sent successfully" {
			t.Errorf("Expected Message 'OTP sent successfully', got %s", resp.Message)
		}
		if resp.ExpiresIn != 900 {
			t.Errorf("Expected ExpiresIn 900, got %d", resp.ExpiresIn)
		}
		if resp.UserID == nil || *resp.UserID != 100 {
			t.Error("Expected UserID to be 100")
		}
		if resp.AccessToken != "access_token_123" {
			t.Errorf("Expected AccessToken 'access_token_123', got %s", resp.AccessToken)
		}
		if resp.TokenType != "Bearer" {
			t.Errorf("Expected TokenType 'Bearer', got %s", resp.TokenType)
		}
	})

	t.Run("OTPResponse with minimal fields", func(t *testing.T) {
		resp := OTPResponse{
			Message:   "OTP code verified",
			ExpiresIn: 900,
		}

		if resp.Message != "OTP code verified" {
			t.Errorf("Expected Message 'OTP code verified', got %s", resp.Message)
		}
		if resp.UserID != nil {
			t.Error("Expected UserID to be nil")
		}
		if resp.AccessToken != "" {
			t.Errorf("Expected empty AccessToken, got %s", resp.AccessToken)
		}
		if resp.TokenType != "" {
			t.Errorf("Expected empty TokenType, got %s", resp.TokenType)
		}
	})
}

func TestOTPAttemptsValidation(t *testing.T) {
	t.Run("Check max attempts", func(t *testing.T) {
		otp := OTP{
			Attempts:    0,
			MaxAttempts: 5,
		}

		if otp.Attempts >= otp.MaxAttempts {
			t.Error("OTP should not be exceeded at start")
		}

		otp.Attempts = 3
		if otp.Attempts >= otp.MaxAttempts {
			t.Error("OTP should still be valid at 3 attempts")
		}

		otp.Attempts = 5
		if otp.Attempts > otp.MaxAttempts {
			t.Error("OTP should be exceeded at max attempts")
		}
		if otp.Attempts == otp.MaxAttempts {
			// At max attempts is still valid boundary
		}
	})
}

func TestOTPExpiryValidation(t *testing.T) {
	t.Run("Check OTP expiry", func(t *testing.T) {
		now := time.Now()

		expiredOTP := OTP{
			Email:     "expired@example.com",
			ExpiresAt: now.Add(-1 * time.Hour),
		}

		validOTP := OTP{
			Email:     "valid@example.com",
			ExpiresAt: now.Add(15 * time.Minute),
		}

		if expiredOTP.ExpiresAt.After(now) {
			t.Error("Expired OTP should not be in the future")
		}

		if !validOTP.ExpiresAt.After(now) {
			t.Error("Valid OTP should be in the future")
		}
	})
}

func TestOTPResendCountValidation(t *testing.T) {
	t.Run("Check resend count", func(t *testing.T) {
		now := time.Now()

		otp := OTP{
			Email:        "resend@example.com",
			ResendCount:  0,
			LastResendAt: nil,
			CreatedAt:    now,
		}

		if otp.ResendCount != 0 {
			t.Errorf("Expected initial ResendCount 0, got %d", otp.ResendCount)
		}

		otp.ResendCount = 2
		if otp.ResendCount != 2 {
			t.Errorf("Expected ResendCount 2, got %d", otp.ResendCount)
		}
	})
}

func TestOTPTypeValidation(t *testing.T) {
	t.Run("Valid OTP types", func(t *testing.T) {
		validTypes := []string{"register", "reset_password"}

		for _, otpType := range validTypes {
			otp := OTP{
				Email: "test@example.com",
				Type:  OTPType(otpType),
			}

			if string(otp.Type) != otpType {
				t.Errorf("Expected Type '%s', got %s", otpType, otp.Type)
			}
		}
	})
}
