package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
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
