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

// OTPRepositoryTestSuite handles all OTP repository tests
type OTPRepositoryTestSuite struct {
	suite.Suite
	db       *gorm.DB
	otpRepo  repo.OTPRepository
}

func (suite *OTPRepositoryTestSuite) SetupTest() {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(suite.T(), err)
	suite.db = db
	suite.otpRepo = repo.NewOTPRepository(db)
}

func (suite *OTPRepositoryTestSuite) TearDownTest() {
	testhelper.TeardownTestDatabase(suite.db)
}

func (suite *OTPRepositoryTestSuite) TestCreate_Success() {
	email := "test@example.com"
	userName := "testuser"
	passwordHash := "hashedpass"
	codeHash := "code123"
	otpType := domain.OTPTypeRegister
	expiresAt := time.Now().Add(10 * time.Minute)
	maxAttempts := 5

	result, err := suite.otpRepo.Create(email, userName, passwordHash, codeHash, otpType, expiresAt, maxAttempts)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), email, result.Email)
	assert.Equal(suite.T(), userName, result.UserName)
	assert.Equal(suite.T(), passwordHash, result.PasswordHash)
	assert.Equal(suite.T(), codeHash, result.CodeHash)
	assert.Equal(suite.T(), otpType, result.Type)
	assert.False(suite.T(), result.IsUsed)
	assert.Equal(suite.T(), 0, result.Attempts)
}

func (suite *OTPRepositoryTestSuite) TestGetByEmail_Success() {
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
	err := suite.db.Create(otp).Error
	require.NoError(suite.T(), err)

	result, err := suite.otpRepo.GetByEmail(email, otpType)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), email, result.Email)
	assert.Equal(suite.T(), otpType, result.Type)
}

func (suite *OTPRepositoryTestSuite) TestGetByEmail_NotFound() {
	result, err := suite.otpRepo.GetByEmail("nonexistent@example.com", domain.OTPTypeRegister)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *OTPRepositoryTestSuite) TestGetByEmail_Expired() {
	email := "test@example.com"
	otpType := domain.OTPTypeRegister
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired

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
	err := suite.db.Create(otp).Error
	require.NoError(suite.T(), err)

	result, err := suite.otpRepo.GetByEmail(email, otpType)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *OTPRepositoryTestSuite) TestGetByEmail_Used() {
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
		IsUsed:       true, // Already used
	}
	err := suite.db.Create(otp).Error
	require.NoError(suite.T(), err)

	result, err := suite.otpRepo.GetByEmail(email, otpType)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *OTPRepositoryTestSuite) TestIncrementAttempts_Success() {
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
	err := suite.db.Create(otp).Error
	require.NoError(suite.T(), err)

	err = suite.otpRepo.IncrementAttempts(otp.ID)
	assert.NoError(suite.T(), err)

	// Verify attempts incremented
	var updatedOTP domain.OTP
	err = suite.db.First(&updatedOTP, otp.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, updatedOTP.Attempts)
}

func (suite *OTPRepositoryTestSuite) TestMarkAsUsed_Success() {
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
	err := suite.db.Create(otp).Error
	require.NoError(suite.T(), err)

	err = suite.otpRepo.MarkAsUsed(otp.ID)
	assert.NoError(suite.T(), err)

	// Verify marked as used
	var updatedOTP domain.OTP
	err = suite.db.First(&updatedOTP, otp.ID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedOTP.IsUsed)
}

func (suite *OTPRepositoryTestSuite) TestDeleteByEmail_Success() {
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
	err := suite.db.Create(otp).Error
	require.NoError(suite.T(), err)

	err = suite.otpRepo.DeleteByEmail("test@example.com", otpType)
	assert.NoError(suite.T(), err)

	// Verify deleted
	var count int64
	suite.db.Model(&domain.OTP{}).Where("email = ?", "test@example.com").Count(&count)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *OTPRepositoryTestSuite) TestUpdateResendInfo_Success() {
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
	err := suite.db.Create(otp).Error
	require.NoError(suite.T(), err)

	resendCount := 2
	lastResendAt := time.Now()

	err = suite.otpRepo.UpdateResendInfo(otp.ID, resendCount, lastResendAt)
	assert.NoError(suite.T(), err)

	// Verify updated
	var updatedOTP domain.OTP
	err = suite.db.First(&updatedOTP, otp.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), resendCount, updatedOTP.ResendCount)
	assert.WithinDuration(suite.T(), lastResendAt, *updatedOTP.LastResendAt, time.Second)
}

func (suite *OTPRepositoryTestSuite) TestDeleteExpired_Success() {
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
	err := suite.db.Create(expiredOTP).Error
	require.NoError(suite.T(), err)

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
	err = suite.db.Create(usedOTP).Error
	require.NoError(suite.T(), err)

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
	err = suite.db.Create(validOTP).Error
	require.NoError(suite.T(), err)

	err = suite.otpRepo.DeleteExpired()
	assert.NoError(suite.T(), err)

	// Verify expired and used OTPs deleted, valid OTP remains
	var count int64
	suite.db.Model(&domain.OTP{}).Count(&count)
	assert.Equal(suite.T(), int64(1), count)

	var remainingOTP domain.OTP
	err = suite.db.First(&remainingOTP).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "valid@example.com", remainingOTP.Email)
}

func TestOTPRepositorySuite(t *testing.T) {
	suite.Run(t, new(OTPRepositoryTestSuite))
}
