package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	goPlaygroundValidator "github.com/go-playground/validator/v10"
)

// validateIndonesiaPhoneNumber validates Indonesian phone number formats.
// Supports:
// - Landline: (0xx) xxxxxxx or 0xx-xxxxxxx or 0xxxxxxxxx
// - Mobile: 08xxxxxxxxxx or +628xxxxxxxxxx
// - With area codes for major cities (Jakarta, Surabaya, etc.)
//
// Usage as struct tag: `validate:"id_phone"`
func validateIndonesiaPhoneNumber(fl goPlaygroundValidator.FieldLevel) bool {
	phone := strings.TrimSpace(fl.Field().String())

	// Remove common separators
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Convert +62 to 0
	if strings.HasPrefix(phone, "+62") {
		phone = "0" + phone[3:]
	}

	// Must start with 0
	if !strings.HasPrefix(phone, "0") {
		return false
	}

	// Length check: 9-13 digits after country code
	if len(phone) < 9 || len(phone) > 14 {
		return false
	}

	// Must be all digits
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(phone) {
		return false
	}

	// Valid area codes/prefixes
	validPrefixes := []string{
		"021", "022", "024", "031", "0341", "061", "062", "064", "065", "071",
		"074", "075", "076", "077", "078", "081", "082", "083", "084", "085",
		"086", "087", "088", "089", "09",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(phone, prefix) {
			return true
		}
	}

	return false
}

// validateIndonesiaMobileNumber validates Indonesian mobile phone numbers.
// Supports formats:
// - 08xxxxxxxxxx (domestic format)
// - +628xxxxxxxxxx (international format)
// - Must be 11-13 digits
// - Valid prefixes: 081, 082, 083, 085, 087, 088, 089
//
// Usage as struct tag: `validate:"id_mobile"`
func validateIndonesiaMobileNumber(fl goPlaygroundValidator.FieldLevel) bool {
	phone := strings.TrimSpace(fl.Field().String())

	// Remove common separators
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Convert +62 to 0
	if strings.HasPrefix(phone, "+62") {
		phone = "0" + phone[3:]
	}

	// Must start with 08 (Indonesian mobile prefix)
	if !strings.HasPrefix(phone, "08") {
		return false
	}

	// Must be 11-13 digits
	if len(phone) < 11 || len(phone) > 13 {
		return false
	}

	// Must be all digits
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(phone) {
		return false
	}

	// Valid mobile operator prefixes
	validPrefixes := []string{
		"0811", "0812", "0813", "0821", "0822", "0823", // Telkomsel
		"0852", "0853", "0851", // Telkomsel (As)
		"0814", "0815", "0816", "0855", "0856", "0857", "0858", // Indosat Ooredoo
		"0817", "0818", "0819", "0859", "0877", "0878", // XL Axiata
		"0895", "0896", "0897", "0898", "0899", // Tri (3)
		"0881", "0882", "0883", "0884", "0885", "0886", "0887", "0888", "0889", // Smartfren
		"0831", "0832", "0833", "0838", // Axis
	}

	// Check first 4 digits
	prefix := phone[:4]
	for _, validPrefix := range validPrefixes {
		if strings.HasPrefix(prefix, validPrefix) || strings.HasPrefix(validPrefix, prefix) {
			return true
		}
	}

	return false
}

// validateNIK validates Indonesian National Identity Card (Nomor Induk Kependudukan) number.
// Format: 16 digits
// Structure: PPRRSSDDMMYYXXXX
// - PP: Province code (01-94, excluding some numbers)
// - RR: Regency/City code
// - SS: District code
// - DD: Birth date (for women, add 40 to day)
// - MM: Birth month (01-12)
// - YY: Birth year
// - XXXX: Serial number
//
// Usage as struct tag: `validate:"nik"`
func validateNIK(fl goPlaygroundValidator.FieldLevel) bool {
	nik := strings.TrimSpace(fl.Field().String())

	// Remove any separators
	nik = strings.ReplaceAll(nik, " ", "")
	nik = strings.ReplaceAll(nik, "-", "")
	nik = strings.ReplaceAll(nik, ".", "")

	// Must be 16 digits
	if len(nik) != 16 {
		return false
	}

	// Must be all digits
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(nik) {
		return false
	}

	// Validate province code (first 2 digits: 01-94)
	provinceCode, err := strconv.Atoi(nik[0:2])
	if err != nil || provinceCode < 1 || provinceCode > 94 {
		return false
	}

	// Exclude invalid province codes
	invalidProvinces := map[int]bool{
		20: true, // Historically invalid
	}
	if invalidProvinces[provinceCode] {
		return false
	}

	// Validate birth date (digits 7-8: day, 5-6: month)
	dayStr := nik[6:8]
	monthStr := nik[4:6]

	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return false
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return false
	}

	// For women, day is added by 40
	actualDay := day
	if day > 40 {
		actualDay = day - 40
	}

	// Validate day based on month
	if actualDay < 1 || actualDay > 31 {
		return false
	}

	// Check specific month limits
	if month == 2 && actualDay > 29 {
		return false
	}
	if month == 4 || month == 6 || month == 9 || month == 11 {
		if actualDay > 30 {
			return false
		}
	}

	return true
}

// validateNPWP validates Indonesian Tax ID (Nomor Pokok Wajib Pajak) number.
// Format: XX.XXX.XXX.X-XXX.XXX
// - XX: Tax office code (2 digits)
// - XXX.XXX.XXX: Taxpayer identification (9 digits)
// - XXX: Security code (3 digits)
// - XXX: Branch code (000 for main office)
//
// Usage as struct tag: `validate:"npwp"`
func validateNPWP(fl goPlaygroundValidator.FieldLevel) bool {
	npwp := strings.TrimSpace(fl.Field().String())

	// Remove all separators (dots and dashes)
	npwp = strings.ReplaceAll(npwp, ".", "")
	npwp = strings.ReplaceAll(npwp, "-", "")

	// Must be 15 digits
	if len(npwp) != 15 {
		return false
	}

	// Must be all digits
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(npwp) {
		return false
	}

	// Validate tax office code (first 2 digits: 01-99)
	taxOfficeCode, err := strconv.Atoi(npwp[0:2])
	if err != nil || taxOfficeCode < 1 || taxOfficeCode > 99 {
		return false
	}

	// Security code should not be 000 (digits 12-14)
	securityCode := npwp[11:14]
	if securityCode == "000" {
		return false
	}

	return true
}

// validateIndonesiaPostalCode validates Indonesian postal codes.
// Format: 5 digits
// First digit: 1-9 (island/group code)
//
// Usage as struct tag: `validate:"id_postal_code"`
func validateIndonesiaPostalCode(fl goPlaygroundValidator.FieldLevel) bool {
	postalCode := strings.TrimSpace(fl.Field().String())

	// Remove separators
	postalCode = strings.ReplaceAll(postalCode, " ", "")
	postalCode = strings.ReplaceAll(postalCode, "-", "")

	// Must be 5 digits
	if len(postalCode) != 5 {
		return false
	}

	// Must be all digits
	if !regexp.MustCompile(`^[0-9]+$`).MatchString(postalCode) {
		return false
	}

	// First digit should be 1-9
	firstDigit := postalCode[0]
	if firstDigit < '1' || firstDigit > '9' {
		return false
	}

	return true
}

// ValidateIndonesiaPhoneNumber exports the validator for use in middleware
func ValidateIndonesiaPhoneNumber(fl goPlaygroundValidator.FieldLevel) bool {
	return validateIndonesiaPhoneNumber(fl)
}

// ValidateIndonesiaMobileNumber exports the validator for use in middleware
func ValidateIndonesiaMobileNumber(fl goPlaygroundValidator.FieldLevel) bool {
	return validateIndonesiaMobileNumber(fl)
}

// ValidateNIK exports the validator for use in middleware
func ValidateNIK(fl goPlaygroundValidator.FieldLevel) bool {
	return validateNIK(fl)
}

// ValidateNPWP exports the validator for use in middleware
func ValidateNPWP(fl goPlaygroundValidator.FieldLevel) bool {
	return validateNPWP(fl)
}

// ValidateIndonesiaPostalCode exports the validator for use in middleware
func ValidateIndonesiaPostalCode(fl goPlaygroundValidator.FieldLevel) bool {
	return validateIndonesiaPostalCode(fl)
}

// FormatPhoneNumber formats Indonesian phone number to standard format.
// Converts +62 to 0 and removes unnecessary separators.
//
// Parameters:
//   - phone: Phone number string
//
// Returns:
//   - string: Formatted phone number
//
// Usage:
//
//	phone := validator.FormatPhoneNumber("+6281234567890")
//	// Result: "081234567890"
func FormatPhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "-", "")

	if strings.HasPrefix(phone, "+62") {
		phone = "0" + phone[3:]
	}

	return phone
}

// FormatNIK formats NIK to standard format (with dots for readability).
//
// Parameters:
//   - nik: NIK string
//
// Returns:
//   - string: Formatted NIK (PP.RR.SS.DDMMYY.XXXX)
//
// Usage:
//
//	nik := "3201012345678901"
//	formatted := validator.FormatNIK(nik)
//	// Result: "32.01.01.234567.8901"
func FormatNIK(nik string) string {
	nik = strings.ReplaceAll(nik, " ", "")
	nik = strings.ReplaceAll(nik, "-", "")

	if len(nik) != 16 {
		return nik
	}

	return fmt.Sprintf("%s.%s.%s.%s.%s",
		nik[0:2], nik[2:4], nik[4:6], nik[6:12], nik[12:16])
}

// FormatNPWP formats NPWP to standard format (XX.XXX.XXX.X-XXX.XXX).
//
// Parameters:
//   - npwp: NPWP string
//
// Returns:
//   - string: Formatted NPWP
//
// Usage:
//
//	npwp := "123456789012345"
//	formatted := validator.FormatNPWP(npwp)
//	// Result: "12.345.678.9-123.456"
func FormatNPWP(npwp string) string {
	npwp = strings.ReplaceAll(npwp, " ", "")

	if len(npwp) != 15 {
		return npwp
	}

	return fmt.Sprintf("%s.%s.%s.%s-%s.%s",
		npwp[0:2], npwp[2:5], npwp[5:8], npwp[8:9], npwp[9:12], npwp[12:15])
}

// ValidateNIKChecksum performs additional validation on NIK using basic checksum.
// Note: NIK doesn't have an official checksum, but this provides basic validation.
//
// Parameters:
//   - nik: NIK string
//
// Returns:
//   - bool: True if checksum passes
//
// Usage:
//
//	if validator.ValidateNIKChecksum(nik) {
//	    fmt.Println("NIK format is valid")
//	}
func ValidateNIKChecksum(nik string) bool {
	nik = strings.ReplaceAll(nik, " ", "")
	nik = strings.ReplaceAll(nik, "-", "")
	nik = strings.ReplaceAll(nik, ".", "")

	if len(nik) != 16 {
		return false
	}

	// Basic check: sum of digits should be > 0 and < some reasonable value
	sum := 0
	for _, char := range nik {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}
		sum += digit
	}

	// Check sum is reasonable (minimum 1, maximum 9*16=144)
	return sum > 0 && sum <= 144
}

// GetProvinceName returns the province name from NIK province code.
//
// Parameters:
//   - provinceCode: Two-digit province code
//
// Returns:
//   - string: Province name in Indonesian, or "Provinsi Tidak Diketahui" if not found
//
// Usage:
//
//	province := validator.GetProvinceName("32")
//	// Result: "Jawa Barat"
func GetProvinceName(provinceCode string) string {
	provinces := map[string]string{
		"11": "Aceh",
		"12": "Sumatera Utara",
		"13": "Sumatera Barat",
		"14": "Riau",
		"15": "Jambi",
		"16": "Sumatera Selatan",
		"17": "Bengkulu",
		"18": "Lampung",
		"19": "Kepulauan Bangka Belitung",
		"21": "Kepulauan Riau",
		"31": "DKI Jakarta",
		"32": "Jawa Barat",
		"33": "Jawa Tengah",
		"34": "DI Yogyakarta",
		"35": "Jawa Timur",
		"36": "Banten",
		"51": "Bali",
		"52": "Nusa Tenggara Barat",
		"53": "Nusa Tenggara Timur",
		"61": "Kalimantan Barat",
		"62": "Kalimantan Tengah",
		"63": "Kalimantan Selatan",
		"64": "Kalimantan Timur",
		"65": "Kalimantan Utara",
		"71": "Sulawesi Utara",
		"72": "Sulawesi Tengah",
		"73": "Sulawesi Selatan",
		"74": "Sulawesi Tenggara",
		"75": "Gorontalo",
		"76": "Sulawesi Barat",
		"81": "Maluku",
		"82": "Maluku Utara",
		"91": "Papua",
		"92": "Papua Barat",
	}

	if name, exists := provinces[provinceCode]; exists {
		return name
	}
	return "Provinsi Tidak Diketahui"
}
