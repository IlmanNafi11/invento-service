package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPasswordStrengthValidation tests password strength validation through validator
func TestPasswordStrengthValidation(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("password_strength", ValidatePasswordStrength)

	type TestStruct struct {
		Password string `validate:"password_strength"`
	}

	tests := []struct {
		name      string
		password  string
		shouldErr bool
	}{
		{"ValidComplex", "Pass123!", false},
		{"ValidLonger", "MySecure@Pass2024", false},
		{"TooShort", "Pass1!", true},
		{"NoLowercase", "PASSWORD123!", true},
		{"NoUppercase", "password123!", true},
		{"NoDigit", "Password!", true},
		{"NoSpecial", "Password123", true},
		{"Empty", "", true},
		{"Exactly8Chars", "Pass123!", false},
		{"SevenChars", "Pass1!a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Password: tt.password}
			err := validate.Struct(s)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIndonesiaPhoneNumberValidation tests Indonesian phone validation
func TestIndonesiaPhoneNumberValidation(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("id_phone", ValidateIndonesiaPhoneNumber)

	type TestStruct struct {
		Phone string `validate:"id_phone"`
	}

	tests := []struct {
		name      string
		phone     string
		shouldErr bool
	}{
		{"ValidMobile", "081234567890", false},
		{"ValidMobileInternational", "+6281234567890", false},
		{"ValidLandline", "0211234567", false},
		{"ValidMinLength9", "091234567", false}, // Minimum 9 digits is valid
		{"InvalidNoZero", "81234567890", true},
		{"InvalidTooShort8", "08123456", true}, // 8 digits is too short
		{"InvalidLetters", "abcdefghijk", true},
		{"ValidWithSpaces", " (021) 123-4567 ", false},
		{"ValidMobileLong", "08123456789012", false},
		{"InvalidTooLong", "081234567890123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Phone: tt.phone}
			err := validate.Struct(s)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIndonesiaMobileNumberValidation tests Indonesian mobile validation
func TestIndonesiaMobileNumberValidation(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("id_mobile", ValidateIndonesiaMobileNumber)

	type TestStruct struct {
		Mobile string `validate:"id_mobile"`
	}

	tests := []struct {
		name      string
		mobile    string
		shouldErr bool
	}{
		{"ValidTelkomsel", "081234567890", false},
		{"ValidIndosat", "085612345678", false},
		{"ValidXL", "081712345678", false},
		{"ValidTri", "089512345678", false},
		{"ValidSmartfren", "088112345678", false},
		{"ValidAxis", "083112345678", false},
		{"ValidInternational", "+6281234567890", false},
		{"InvalidLandline", "0211234567", true},
		{"InvalidPrefix", "08012345678", true},
		{"InvalidTooShort", "0812345678", true},
		{"InvalidTooLong", "081234567890123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Mobile: tt.mobile}
			err := validate.Struct(s)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNIKValidation tests NIK validation
func TestNIKValidation(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("nik", ValidateNIK)

	type TestStruct struct {
		NIK string `validate:"nik"`
	}

	tests := []struct {
		name      string
		nik       string
		shouldErr bool
	}{
		{"ValidJakarta", "3201010100000001", false},
		{"ValidFebruary", "3201012900000001", false},
		{"ValidFemaleDay", "3201010100000041", false},
		{"InvalidTooShort", "32010123456789", true},
		{"InvalidTooLong", "32010123456789012", true},
		{"InvalidMonth00", "3201001000000001", true},
		{"InvalidMonth13", "3201131000000001", true},
		{"InvalidDay32", "3201013200000001", true},
		{"InvalidProvince00", "0001010100000001", true},
		{"InvalidProvince20", "2001010100000001", true},
		{"InvalidProvince95", "9501010100000001", true},
		{"InvalidLetters", "32010100abcdefgh", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{NIK: tt.nik}
			err := validate.Struct(s)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNPWPValidation tests NPWP validation
func TestNPWPValidation(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("npwp", ValidateNPWP)

	type TestStruct struct {
		NPWP string `validate:"npwp"`
	}

	tests := []struct {
		name      string
		npwp      string
		shouldErr bool
	}{
		{"ValidStandard", "012345678901234", false},
		{"ValidWithSeparators", "01.234.567.8-901.234", false},
		{"InvalidTooShort", "01234567890123", true},
		{"InvalidTooLong", "0123456789012345", true},
		{"InvalidTaxOffice00", "001234567890123", true},
		{"InvalidSecurityCode000", "012345678900001", true},
		{"InvalidLetters", "01234567890abcd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{NPWP: tt.npwp}
			err := validate.Struct(s)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIndonesiaPostalCodeValidation tests postal code validation
func TestIndonesiaPostalCodeValidation(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("id_postal_code", ValidateIndonesiaPostalCode)

	type TestStruct struct {
		PostalCode string `validate:"id_postal_code"`
	}

	tests := []struct {
		name      string
		code      string
		shouldErr bool
	}{
		{"ValidJakarta", "10110", false},
		{"ValidBandung", "40123", false},
		{"ValidSurabaya", "60123", false},
		{"ValidBali", "80123", false},
		{"ValidPapua", "99999", false},
		{"InvalidStartsWith0", "01234", true},
		{"InvalidTooShort", "1234", true},
		{"InvalidTooLong", "123456", true},
		{"InvalidLetters", "1234a", true},
		{"InvalidAllLetters", "abcde", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{PostalCode: tt.code}
			err := validate.Struct(s)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMultipleValidations tests multiple validations together
func TestMultipleValidations(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("password_strength", ValidatePasswordStrength)
	validate.RegisterValidation("id_phone", ValidateIndonesiaPhoneNumber)
	validate.RegisterValidation("id_mobile", ValidateIndonesiaMobileNumber)
	validate.RegisterValidation("nik", ValidateNIK)
	validate.RegisterValidation("npwp", ValidateNPWP)
	validate.RegisterValidation("id_postal_code", ValidateIndonesiaPostalCode)

	type UserForm struct {
		Password    string `validate:"password_strength"`
		Phone       string `validate:"id_phone"`
		Mobile      string `validate:"id_mobile"`
		NIK         string `validate:"nik"`
		NPWP        string `validate:"npwp"`
		PostalCode  string `validate:"id_postal_code"`
	}

	validForm := UserForm{
		Password:   "SecurePass123!",
		Phone:      "0211234567",
		Mobile:     "081234567890",
		NIK:        "3201010100000001",
		NPWP:       "012345678901234",
		PostalCode: "12345",
	}

	err := validate.Struct(validForm)
	require.NoError(t, err)

	invalidForm := UserForm{
		Password:   "weak",
		Phone:      "invalid",
		Mobile:     "0211234567",
		NIK:        "3201010100000001",
		NPWP:       "012345678901234",
		PostalCode: "12345",
	}

	err = validate.Struct(invalidForm)
	assert.Error(t, err)
}
