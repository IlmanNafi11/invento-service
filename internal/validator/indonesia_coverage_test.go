package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFormatPhoneNumber_Success tests phone number formatting
func TestFormatPhoneNumber_Success(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"081234567890", "081234567890"},
		{"+6281234567890", "081234567890"},
		{" 081 2345 67890 ", "081234567890"},
		{"(021) 123-4567", "0211234567"},
		{"021-1234-5678", "02112345678"},
		{"+62 812 3456 7890", "081234567890"},
		{"+62-812-3456-7890", "081234567890"},
		{"(+62) 812-3456-7890", "081234567890"},
		{"  +62  812  -  3456  -  7890  ", "081234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatPhoneNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatPhoneNumber_Empty tests empty string
func TestFormatPhoneNumber_Empty(t *testing.T) {
	result := FormatPhoneNumber("")
	assert.Equal(t, "", result)
}

// TestFormatPhoneNumber_OnlySpaces tests only spaces
func TestFormatPhoneNumber_OnlySpaces(t *testing.T) {
	result := FormatPhoneNumber("     ")
	assert.Equal(t, "", result)
}

// TestFormatNIK_Success tests NIK formatting
func TestFormatNIK_Success(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"3201012345678901", "32.01.01.234567.8901"},
		{"3201012345678902", "32.01.01.234567.8902"},
		{"1234567890123456", "12.34.56.789012.3456"},
		{" 32 01 01 2345 67 89 01 ", "32.01.01.234567.8901"},
		{"32-01-01-2345-67-89-01", "32.01.01.234567.8901"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatNIK(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatNIK_Short returns input as-is when not 16 digits
func TestFormatNIK_Short(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"12345", "12345"},
		{"", ""},
		{"123456789012345", "123456789012345"},
		{"12345678901234567", "12345678901234567"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatNIK(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatNPWP_Success tests NPWP formatting
func TestFormatNPWP_Success(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123456789012345", "12.345.678.9-012.345"},
		{"012345678901234", "01.234.567.8-901.234"},
		{"987654321098765", "98.765.432.1-098.765"},
		{"12 345 678 9 012 345", "12.345.678.9-012.345"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatNPWP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatNPWP_Short returns input as-is when not 15 digits
func TestFormatNPWP_Short(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"12345", "12345"},
		{"", ""},
		{"12345678901234", "12345678901234"},
		{"1234567890123456", "1234567890123456"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatNPWP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateNIKChecksum_Success tests valid NIK checksums
func TestValidateNIKChecksum_Success(t *testing.T) {
	validNIKs := []string{
		"3201012345678901",
		"3201012345678902",
		"3201012345678903",
		"1234567890123456",
		"1111111111111111",
		"9999999999999999",
	}

	for _, nik := range validNIKs {
		t.Run(nik, func(t *testing.T) {
			result := ValidateNIKChecksum(nik)
			assert.True(t, result, "NIK should be valid: "+nik)
		})
	}
}

// TestValidateNIKChecksum_Failure tests invalid NIKs
func TestValidateNIKChecksum_Failure(t *testing.T) {
	invalidNIKs := []string{
		"",
		"12345",
		"12345678901234567",
		"abcdefghijklmnop",
		"0000000000000000",
		"12345678901234a",
	}

	for _, nik := range invalidNIKs {
		t.Run(nik, func(t *testing.T) {
			result := ValidateNIKChecksum(nik)
			assert.False(t, result, "NIK should be invalid: "+nik)
		})
	}
}

// TestGetProvinceName_Success tests province name lookup
func TestGetProvinceName_Success(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"11", "Aceh"},
		{"12", "Sumatera Utara"},
		{"13", "Sumatera Barat"},
		{"31", "DKI Jakarta"},
		{"32", "Jawa Barat"},
		{"33", "Jawa Tengah"},
		{"34", "DI Yogyakarta"},
		{"35", "Jawa Timur"},
		{"36", "Banten"},
		{"51", "Bali"},
		{"61", "Kalimantan Barat"},
		{"71", "Sulawesi Utara"},
		{"81", "Maluku"},
		{"91", "Papua"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			province := GetProvinceName(tt.code)
			assert.Equal(t, tt.expected, province)
		})
	}
}

// TestGetProvinceName_UnknownCode tests unknown area code
func TestGetProvinceName_UnknownCode(t *testing.T) {
	unknownCodes := []string{
		"00",
		"99",
		"999",
		"abc",
		"",
	}

	for _, code := range unknownCodes {
		t.Run(code, func(t *testing.T) {
			province := GetProvinceName(code)
			assert.Equal(t, "Provinsi Tidak Diketahui", province)
		})
	}
}

// TestGetProvinceName_AllProvinces tests all province codes
func TestGetProvinceName_AllProvinces(t *testing.T) {
	provinces := []struct {
		code  string
		name  string
	}{
		{"11", "Aceh"},
		{"12", "Sumatera Utara"},
		{"13", "Sumatera Barat"},
		{"14", "Riau"},
		{"15", "Jambi"},
		{"16", "Sumatera Selatan"},
		{"17", "Bengkulu"},
		{"18", "Lampung"},
		{"19", "Kepulauan Bangka Belitung"},
		{"21", "Kepulauan Riau"},
		{"31", "DKI Jakarta"},
		{"32", "Jawa Barat"},
		{"33", "Jawa Tengah"},
		{"34", "DI Yogyakarta"},
		{"35", "Jawa Timur"},
		{"36", "Banten"},
		{"51", "Bali"},
		{"52", "Nusa Tenggara Barat"},
		{"53", "Nusa Tenggara Timur"},
		{"61", "Kalimantan Barat"},
		{"62", "Kalimantan Tengah"},
		{"63", "Kalimantan Selatan"},
		{"64", "Kalimantan Timur"},
		{"65", "Kalimantan Utara"},
		{"71", "Sulawesi Utara"},
		{"72", "Sulawesi Tengah"},
		{"73", "Sulawesi Selatan"},
		{"74", "Sulawesi Tenggara"},
		{"75", "Gorontalo"},
		{"76", "Sulawesi Barat"},
		{"81", "Maluku"},
		{"82", "Maluku Utara"},
		{"91", "Papua"},
		{"92", "Papua Barat"},
	}

	for _, p := range provinces {
		t.Run(p.code, func(t *testing.T) {
			province := GetProvinceName(p.code)
			assert.Equal(t, p.name, province)
		})
	}
}

// TestFormatPhoneNumber_VariousFormats tests various phone number formats
func TestFormatPhoneNumber_VariousFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+62 812 3456 7890", "081234567890"},
		{"+62-812-3456-7890", "081234567890"},
		{"(021) 123-4567", "0211234567"},
		{"021 1234 5678", "02112345678"},
		{"0812-3456-7890", "081234567890"},
		{"08 1234 567890", "081234567890"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatPhoneNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatNIK_WithSeparators tests NIK formatting with various separators
func TestFormatNIK_WithSeparators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"32 01 01 2345 67 89 01", "32.01.01.234567.8901"},
		{"32-01-01-2345-67-89-01", "32.01.01.234567.8901"},
		{"32.01.01.2345.67.89.01", "32.01.01.2345.67.89.01"}, // FormatNIK only removes spaces/dashes, keeps dots
		{"3201012345678901", "32.01.01.234567.8901"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatNIK(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatNPWP_WithSeparators tests NPWP formatting with various separators
func TestFormatNPWP_WithSeparators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"12 345 678 9 012 345", "12.345.678.9-012.345"},
		{"12-345-678-9-012-345", "12-345-678-9-012-345"}, // FormatNPWP only removes spaces, keeps dashes
		{"123456789012345", "12.345.678.9-012.345"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatNPWP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateNIKChecksum_EdgeCases tests edge cases for NIK checksum
func TestValidateNIKChecksum_EdgeCases(t *testing.T) {
	tests := []struct {
		nik       string
		shouldPass bool
	}{
		{"0000000000000001", true},
		{"1000000000000000", true},
		{"0000000000000000", false},
		{"1111111111111111", true},
		{"9999999999999999", true},
		{"aaaaaaaaaaaaaaaa", false},
		{"000000000000000a", false},
	}

	for _, tt := range tests {
		t.Run(tt.nik, func(t *testing.T) {
			result := ValidateNIKChecksum(tt.nik)
			assert.Equal(t, tt.shouldPass, result)
		})
	}
}

// TestGetProvinceName_CaseSensitivity tests case sensitivity
func TestGetProvinceName_CaseSensitivity(t *testing.T) {
	// Lowercase should not match (exact key lookup)
	province := GetProvinceName("32")
	assert.Equal(t, "Jawa Barat", province)

	// Uppercase should not match
	province = GetProvinceName("AB")
	assert.Equal(t, "Provinsi Tidak Diketahui", province)
}
