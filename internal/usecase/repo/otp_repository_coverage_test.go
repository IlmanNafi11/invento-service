package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOTPRepository_Create_Success tests successful OTP creation
func TestOTPRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	otpRepo := NewOTPRepository(db)
	email := "test@example.com"
	userName := "testuser"
	passwordHash := "hashedpass"
	codeHash := "code123"
	otpType := domain.OTPTypeRegister
	expiresAt := time.Now().Add(10 * time.Minute)
	maxAttempts := 5

	result, err := otpRepo.Create(email, userName, passwordHash, codeHash, otpType, expiresAt, maxAttempts)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, userName, result.UserName)
	assert.Equal(t, passwordHash, result.PasswordHash)
	assert.Equal(t, codeHash, result.CodeHash)
	assert.Equal(t, otpType, result.Type)
	assert.False(t, result.IsUsed)
	assert.Equal(t, 0, result.Attempts)
}

// TestOTPRepository_GetByEmail_Success tests successful OTP retrieval by email
func TestOTPRepository_GetByEmail_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	email := "test@example.com"
	otpType := domain.OTPTypeRegister
	expiresAt := time.Now().Add(10 * time.Minute)

	otp := &domain.OTP{
		Email:        email,
		UserName:     "testuser",
		PasswordHash: "hashed",
		CodeHash:     "code123",
		Type:         otpType,
		ExpiresAt:    expiresAt,
		MaxAttempts:  5,
		IsUsed:       false,
	}
	err = db.Create(otp).Error
	require.NoError(t, err)

	otpRepo := NewOTPRepository(db)
	result, err := otpRepo.GetByEmail(email, otpType)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, otpType, result.Type)
}

// TestOTPRepository_IncrementAttempts_Success tests successful increment of OTP attempts
func TestOTPRepository_IncrementAttempts_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	otp := &domain.OTP{
		Email:        "test@example.com",
		UserName:     "testuser",
		PasswordHash: "hashed",
		CodeHash:     "code123",
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		MaxAttempts:  5,
		IsUsed:       false,
		Attempts:     1,
	}
	err = db.Create(otp).Error
	require.NoError(t, err)

	otpRepo := NewOTPRepository(db)
	err = otpRepo.IncrementAttempts(otp.ID)
	assert.NoError(t, err)

	// Verify attempts incremented
	var updatedOTP domain.OTP
	err = db.First(&updatedOTP, otp.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, updatedOTP.Attempts)
}

// TestOTPRepository_MarkAsUsed_Success tests successful marking OTP as used
func TestOTPRepository_MarkAsUsed_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	otp := &domain.OTP{
		Email:        "test@example.com",
		UserName:     "testuser",
		PasswordHash: "hashed",
		CodeHash:     "code123",
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		MaxAttempts:  5,
		IsUsed:       false,
	}
	err = db.Create(otp).Error
	require.NoError(t, err)

	otpRepo := NewOTPRepository(db)
	err = otpRepo.MarkAsUsed(otp.ID)
	assert.NoError(t, err)

	// Verify marked as used
	var updatedOTP domain.OTP
	err = db.First(&updatedOTP, otp.ID).Error
	assert.NoError(t, err)
	assert.True(t, updatedOTP.IsUsed)
}

// TestOTPRepository_DeleteByEmail_Success tests successful deletion of OTP by email
func TestOTPRepository_DeleteByEmail_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	otpType := domain.OTPTypeRegister

	otp := &domain.OTP{
		Email:        "test@example.com",
		UserName:     "testuser",
		PasswordHash: "hashed",
		CodeHash:     "code123",
		Type:         otpType,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		MaxAttempts:  5,
		IsUsed:       false,
	}
	err = db.Create(otp).Error
	require.NoError(t, err)

	otpRepo := NewOTPRepository(db)
	err = otpRepo.DeleteByEmail("test@example.com", otpType)
	assert.NoError(t, err)

	// Verify deleted
	var count int64
	db.Model(&domain.OTP{}).Where("email = ?", "test@example.com").Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestOTPRepository_UpdateResendInfo_Success tests successful update of resend info
func TestOTPRepository_UpdateResendInfo_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	otp := &domain.OTP{
		Email:        "test@example.com",
		UserName:     "testuser",
		PasswordHash: "hashed",
		CodeHash:     "code123",
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		MaxAttempts:  5,
		IsUsed:       false,
	}
	err = db.Create(otp).Error
	require.NoError(t, err)

	otpRepo := NewOTPRepository(db)
	resendCount := 2
	lastResendAt := time.Now()

	err = otpRepo.UpdateResendInfo(otp.ID, resendCount, lastResendAt)
	assert.NoError(t, err)

	// Verify updated
	var updatedOTP domain.OTP
	err = db.First(&updatedOTP, otp.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, resendCount, updatedOTP.ResendCount)
	assert.WithinDuration(t, lastResendAt, *updatedOTP.LastResendAt, time.Second)
}

// TestOTPRepository_DeleteExpired_Success tests successful deletion of expired OTPs
func TestOTPRepository_DeleteExpired_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	// Create expired OTP
	expiredOTP := &domain.OTP{
		Email:        "expired@example.com",
		UserName:     "expireduser",
		PasswordHash: "hashed",
		CodeHash:     "code123",
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		MaxAttempts:  5,
		IsUsed:       false,
	}
	err = db.Create(expiredOTP).Error
	require.NoError(t, err)

	// Create used OTP
	usedOTP := &domain.OTP{
		Email:        "used@example.com",
		UserName:     "useduser",
		PasswordHash: "hashed",
		CodeHash:     "code456",
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		MaxAttempts:  5,
		IsUsed:       true,
	}
	err = db.Create(usedOTP).Error
	require.NoError(t, err)

	// Create valid OTP
	validOTP := &domain.OTP{
		Email:        "valid@example.com",
		UserName:     "validuser",
		PasswordHash: "hashed",
		CodeHash:     "code789",
		Type:         domain.OTPTypeRegister,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		MaxAttempts:  5,
		IsUsed:       false,
	}
	err = db.Create(validOTP).Error
	require.NoError(t, err)

	otpRepo := NewOTPRepository(db)
	err = otpRepo.DeleteExpired()
	assert.NoError(t, err)

	// Verify expired and used OTPs deleted, valid OTP remains
	var count int64
	db.Model(&domain.OTP{}).Count(&count)
	assert.Equal(t, int64(1), count)

	var remainingOTP domain.OTP
	err = db.First(&remainingOTP).Error
	assert.NoError(t, err)
	assert.Equal(t, "valid@example.com", remainingOTP.Email)
}
