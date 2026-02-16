package httputil_test

import (
	dto "invento-service/internal/dto"
	"invento-service/internal/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStruct_ValidData(t *testing.T) {
	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	errors := httputil.ValidateStruct(req)

	assert.Empty(t, errors)
}

func TestValidateStruct_InvalidData(t *testing.T) {
	req := dto.RegisterRequest{
		Name:     "",
		Email:    "invalid-email",
		Password: "123",
	}

	errors := httputil.ValidateStruct(req)

	assert.NotEmpty(t, errors)
	assert.Greater(t, len(errors), 0)

	for _, validationError := range errors {
		assert.NotEmpty(t, validationError.Field)
		assert.NotEmpty(t, validationError.Message)
	}
}

func TestValidateStruct_EmptyEmail(t *testing.T) {
	req := dto.AuthRequest{
		Email:    "",
		Password: "password123",
	}

	errors := httputil.ValidateStruct(req)

	assert.NotEmpty(t, errors)
	assert.Greater(t, len(errors), 0)
}

func TestValidateStruct_ShortPassword(t *testing.T) {
	req := dto.AuthRequest{
		Email:    "test@example.com",
		Password: "123",
	}

	errors := httputil.ValidateStruct(req)

	assert.NotEmpty(t, errors)
	assert.Greater(t, len(errors), 0)
}

func TestValidateStruct_ResetPasswordRequest(t *testing.T) {
	validReq := dto.ResetPasswordRequest{
		Email: "test@example.com",
	}

	validErrors := httputil.ValidateStruct(validReq)
	assert.Empty(t, validErrors)

	invalidReq := dto.ResetPasswordRequest{
		Email: "invalid-email",
	}

	invalidErrors := httputil.ValidateStruct(invalidReq)
	assert.NotEmpty(t, invalidErrors)
}

func TestValidateStruct_RefreshTokenRequest(t *testing.T) {
	validReq := dto.RefreshTokenRequest{
		RefreshToken: "valid_token_123",
	}

	validErrors := httputil.ValidateStruct(validReq)
	assert.Empty(t, validErrors)

	invalidReq := dto.RefreshTokenRequest{
		RefreshToken: "",
	}

	invalidErrors := httputil.ValidateStruct(invalidReq)
	assert.NotEmpty(t, invalidErrors)
}

// Test structs for various validation tags

type TestRequired struct {
	Field string `validate:"required"`
}

type TestEmail struct {
	Email string `validate:"email"`
}

type TestMin struct {
	Field string `validate:"min=5"`
}

type TestMax struct {
	Field string `validate:"max=10"`
}

type TestLen struct {
	Field string `validate:"len=5"`
}

type TestEq struct {
	Field string `validate:"eq=exact"`
}

type TestNe struct {
	Field string `validate:"ne=forbidden"`
}

type TestLt struct {
	Field int `validate:"lt=10"`
}

type TestLte struct {
	Field int `validate:"lte=10"`
}

type TestGt struct {
	Field int `validate:"gt=5"`
}

type TestGte struct {
	Field int `validate:"gte=5"`
}

type TestOneof struct {
	Field string `validate:"oneof=red green blue"`
}

type TestURL struct {
	URL string `validate:"url"`
}

type TestURI struct {
	URI string `validate:"uri"`
}

type TestAlpha struct {
	Field string `validate:"alpha"`
}

type TestAlphanum struct {
	Field string `validate:"alphanum"`
}

type TestNumeric struct {
	Field string `validate:"numeric"`
}

type TestNumber struct {
	Field string `validate:"number"`
}

type TestHexadecimal struct {
	Field string `validate:"hexadecimal"`
}

type TestHexcolor struct {
	Field string `validate:"hexcolor"`
}

type TestRgb struct {
	Field string `validate:"rgb"`
}

type TestRgba struct {
	Field string `validate:"rgba"`
}

type TestHsl struct {
	Field string `validate:"hsl"`
}

type TestHsla struct {
	Field string `validate:"hsla"`
}

type TestUUID struct {
	Field string `validate:"uuid"`
}

type TestUUID3 struct {
	Field string `validate:"uuid3"`
}

type TestUUID4 struct {
	Field string `validate:"uuid4"`
}

type TestUUID5 struct {
	Field string `validate:"uuid5"`
}

type TestISBN struct {
	Field string `validate:"isbn"`
}

type TestISBN10 struct {
	Field string `validate:"isbn10"`
}

type TestISBN13 struct {
	Field string `validate:"isbn13"`
}

type TestContainsany struct {
	Field string `validate:"containsany=!@#"`
}

type TestContains struct {
	Field string `validate:"contains=keyword"`
}

type TestExcludes struct {
	Field string `validate:"excludes=forbidden"`
}

type TestExcludesall struct {
	Field string `validate:"excludesall=123"`
}

type TestExcludesrune struct {
	Field string `validate:"excludesrune=@"`
}

type TestStartswith struct {
	Field string `validate:"startswith=prefix"`
}

type TestEndswith struct {
	Field string `validate:"endswith=suffix"`
}

type TestDatetime struct {
	Field string `validate:"datetime=2006-01-02"`
}

// Tests for ValidateStruct with various validation tags

func TestValidateStruct_Required_Valid(t *testing.T) {
	req := TestRequired{Field: "value"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Required_Invalid(t *testing.T) {
	req := TestRequired{Field: ""}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Equal(t, "Field", errors[0].Field)
	assert.Contains(t, errors[0].Message, "wajib diisi")
}

func TestValidateStruct_Email_Valid(t *testing.T) {
	req := TestEmail{Email: "test@example.com"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Email_Invalid(t *testing.T) {
	req := TestEmail{Email: "invalid-email"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "Format email tidak valid")
}

func TestValidateStruct_Min_Valid(t *testing.T) {
	req := TestMin{Field: "abcdefgh"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Min_Invalid(t *testing.T) {
	req := TestMin{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "minimal")
	assert.Contains(t, errors[0].Message, "5")
}

func TestValidateStruct_Max_Valid(t *testing.T) {
	req := TestMax{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Max_Invalid(t *testing.T) {
	req := TestMax{Field: "abcdefghijk"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "maksimal")
	assert.Contains(t, errors[0].Message, "10")
}

func TestValidateStruct_Len_Valid(t *testing.T) {
	req := TestLen{Field: "abcde"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Len_Invalid(t *testing.T) {
	req := TestLen{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus")
	assert.Contains(t, errors[0].Message, "5 karakter")
}

func TestValidateStruct_Eq_Valid(t *testing.T) {
	req := TestEq{Field: "exact"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Eq_Invalid(t *testing.T) {
	req := TestEq{Field: "other"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus sama dengan")
	assert.Contains(t, errors[0].Message, "exact")
}

func TestValidateStruct_Ne_Valid(t *testing.T) {
	req := TestNe{Field: "allowed"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Ne_Invalid(t *testing.T) {
	req := TestNe{Field: "forbidden"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh sama dengan")
	assert.Contains(t, errors[0].Message, "forbidden")
}

func TestValidateStruct_Lt_Valid(t *testing.T) {
	req := TestLt{Field: 5}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Lt_Invalid(t *testing.T) {
	req := TestLt{Field: 15}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus kurang dari")
	assert.Contains(t, errors[0].Message, "10")
}

func TestValidateStruct_Lte_Valid(t *testing.T) {
	req := TestLte{Field: 10}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Lte_Invalid(t *testing.T) {
	req := TestLte{Field: 15}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus kurang dari atau sama dengan")
}

func TestValidateStruct_Gt_Valid(t *testing.T) {
	req := TestGt{Field: 10}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Gt_Invalid(t *testing.T) {
	req := TestGt{Field: 3}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus lebih dari")
	assert.Contains(t, errors[0].Message, "5")
}

func TestValidateStruct_Gte_Valid(t *testing.T) {
	req := TestGte{Field: 5}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Gte_Invalid(t *testing.T) {
	req := TestGte{Field: 3}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus lebih dari atau sama dengan")
}

func TestValidateStruct_Oneof_Valid(t *testing.T) {
	req := TestOneof{Field: "red"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Oneof_Invalid(t *testing.T) {
	req := TestOneof{Field: "yellow"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus salah satu dari")
	assert.Contains(t, errors[0].Message, "red green blue")
}

func TestValidateStruct_URL_Valid(t *testing.T) {
	req := TestURL{URL: "https://example.com"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_URL_Invalid(t *testing.T) {
	req := TestURL{URL: "not-a-url"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "Format URL tidak valid")
}

func TestValidateStruct_URI_Valid(t *testing.T) {
	req := TestURI{URI: "http://example.com/path"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_URI_Invalid(t *testing.T) {
	req := TestURI{URI: "not a uri"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "Format URI tidak valid")
}

func TestValidateStruct_Alpha_Valid(t *testing.T) {
	req := TestAlpha{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Alpha_Invalid(t *testing.T) {
	req := TestAlpha{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "hanya boleh berisi huruf")
}

func TestValidateStruct_Alphanum_Valid(t *testing.T) {
	req := TestAlphanum{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Alphanum_Invalid(t *testing.T) {
	req := TestAlphanum{Field: "abc-123"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "hanya boleh berisi huruf dan angka")
}

func TestValidateStruct_Numeric_Valid(t *testing.T) {
	req := TestNumeric{Field: "123"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Numeric_Invalid(t *testing.T) {
	req := TestNumeric{Field: "12a3"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "hanya boleh berisi angka")
}

func TestValidateStruct_Number_Valid(t *testing.T) {
	req := TestNumber{Field: "123"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Number_Invalid(t *testing.T) {
	req := TestNumber{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa angka")
}

func TestValidateStruct_Hexadecimal_Valid(t *testing.T) {
	req := TestHexadecimal{Field: "abc123DEF"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hexadecimal_Invalid(t *testing.T) {
	req := TestHexadecimal{Field: "xyz"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa hexadecimal")
}

func TestValidateStruct_Hexcolor_Valid(t *testing.T) {
	req := TestHexcolor{Field: "#fff"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hexcolor_Invalid(t *testing.T) {
	req := TestHexcolor{Field: "ggg"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna hex")
}

func TestValidateStruct_Rgb_Valid(t *testing.T) {
	req := TestRgb{Field: "rgb(255, 255, 255)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Rgb_Invalid(t *testing.T) {
	req := TestRgb{Field: "not rgb"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna RGB")
}

func TestValidateStruct_Rgba_Valid(t *testing.T) {
	req := TestRgba{Field: "rgba(255, 255, 255, 0.5)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Rgba_Invalid(t *testing.T) {
	req := TestRgba{Field: "not rgba"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna RGBA")
}

func TestValidateStruct_Hsl_Valid(t *testing.T) {
	req := TestHsl{Field: "hsl(120, 100%, 50%)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hsl_Invalid(t *testing.T) {
	req := TestHsl{Field: "not hsl"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna HSL")
}

func TestValidateStruct_Hsla_Valid(t *testing.T) {
	req := TestHsla{Field: "hsla(120, 100%, 50%, 0.5)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hsla_Invalid(t *testing.T) {
	req := TestHsla{Field: "not hsla"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna HSLA")
}

func TestValidateStruct_UUID_Valid(t *testing.T) {
	req := TestUUID{Field: "550e8400-e29b-41d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID_Invalid(t *testing.T) {
	req := TestUUID{Field: "not-uuid"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID")
}

func TestValidateStruct_UUID3_Valid(t *testing.T) {
	req := TestUUID3{Field: "550e8400-e29b-31d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID3_Invalid(t *testing.T) {
	req := TestUUID3{Field: "not-uuid3"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID versi 3")
}

func TestValidateStruct_UUID4_Valid(t *testing.T) {
	req := TestUUID4{Field: "550e8400-e29b-41d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID4_Invalid(t *testing.T) {
	req := TestUUID4{Field: "not-uuid4"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID versi 4")
}

func TestValidateStruct_UUID5_Valid(t *testing.T) {
	req := TestUUID5{Field: "550e8400-e29b-51d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID5_Invalid(t *testing.T) {
	req := TestUUID5{Field: "not-uuid5"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID versi 5")
}

func TestValidateStruct_ISBN_Valid(t *testing.T) {
	req := TestISBN{Field: "978-3-16-148410-0"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_ISBN_Invalid(t *testing.T) {
	req := TestISBN{Field: "not-isbn"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa ISBN")
}

func TestValidateStruct_ISBN10_Valid(t *testing.T) {
	req := TestISBN10{Field: "0-306-40615-2"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_ISBN10_Invalid(t *testing.T) {
	req := TestISBN10{Field: "not-isbn10"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa ISBN-10")
}

func TestValidateStruct_ISBN13_Valid(t *testing.T) {
	req := TestISBN13{Field: "978-3-16-148410-0"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_ISBN13_Invalid(t *testing.T) {
	req := TestISBN13{Field: "not-isbn13"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa ISBN-13")
}

func TestValidateStruct_Containsany_Valid(t *testing.T) {
	req := TestContainsany{Field: "test@example.com"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Containsany_Invalid(t *testing.T) {
	req := TestContainsany{Field: "testexamplecom"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus mengandung salah satu dari")
	assert.Contains(t, errors[0].Message, "!@#")
}

func TestValidateStruct_Contains_Valid(t *testing.T) {
	req := TestContains{Field: "this is a keyword test"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Contains_Invalid(t *testing.T) {
	req := TestContains{Field: "this is a test"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus mengandung")
	assert.Contains(t, errors[0].Message, "keyword")
}

func TestValidateStruct_Excludes_Valid(t *testing.T) {
	req := TestExcludes{Field: "allowed text"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Excludes_Invalid(t *testing.T) {
	req := TestExcludes{Field: "forbidden text"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh mengandung")
	assert.Contains(t, errors[0].Message, "forbidden")
}

func TestValidateStruct_Excludesall_Valid(t *testing.T) {
	req := TestExcludesall{Field: "abcdef"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Excludesall_Invalid(t *testing.T) {
	req := TestExcludesall{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh mengandung karakter")
	assert.Contains(t, errors[0].Message, "123")
}

func TestValidateStruct_Excludesrune_Valid(t *testing.T) {
	req := TestExcludesrune{Field: "without-at"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Excludesrune_Invalid(t *testing.T) {
	req := TestExcludesrune{Field: "with@sign"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh mengandung karakter")
	assert.Contains(t, errors[0].Message, "@")
}

func TestValidateStruct_Startswith_Valid(t *testing.T) {
	req := TestStartswith{Field: "prefix_value"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Startswith_Invalid(t *testing.T) {
	req := TestStartswith{Field: "value"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus diawali dengan")
	assert.Contains(t, errors[0].Message, "prefix")
}

func TestValidateStruct_Endswith_Valid(t *testing.T) {
	req := TestEndswith{Field: "value_suffix"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Endswith_Invalid(t *testing.T) {
	req := TestEndswith{Field: "value"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus diakhiri dengan")
	assert.Contains(t, errors[0].Message, "suffix")
}

func TestValidateStruct_Datetime_Valid(t *testing.T) {
	req := TestDatetime{Field: "2023-12-25"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Datetime_Invalid(t *testing.T) {
	req := TestDatetime{Field: "not-a-date"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa tanggal dengan format")
	assert.Contains(t, errors[0].Message, "2006-01-02")
}

// Test getValidationMessage directly via ValidateStruct with Indonesian messages

func TestGetValidationMessage_Required(t *testing.T) {
	req := TestRequired{Field: ""}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field wajib diisi", errors[0].Message)
}

func TestGetValidationMessage_Email(t *testing.T) {
	req := TestEmail{Email: "invalid"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Format email tidak valid", errors[0].Message)
}

func TestGetValidationMessage_Min(t *testing.T) {
	req := TestMin{Field: "ab"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field minimal 5 karakter", errors[0].Message)
}

func TestGetValidationMessage_Max(t *testing.T) {
	req := TestMax{Field: "abcdefghijk"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field maksimal 10 karakter", errors[0].Message)
}

func TestGetValidationMessage_Len(t *testing.T) {
	req := TestLen{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus 5 karakter", errors[0].Message)
}

func TestGetValidationMessage_Eq(t *testing.T) {
	req := TestEq{Field: "wrong"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus sama dengan exact", errors[0].Message)
}

func TestGetValidationMessage_Ne(t *testing.T) {
	req := TestNe{Field: "forbidden"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh sama dengan forbidden", errors[0].Message)
}

func TestGetValidationMessage_Lt(t *testing.T) {
	req := TestLt{Field: 15}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus kurang dari 10", errors[0].Message)
}

func TestGetValidationMessage_Lte(t *testing.T) {
	req := TestLte{Field: 11}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus kurang dari atau sama dengan 10", errors[0].Message)
}

func TestGetValidationMessage_Gt(t *testing.T) {
	req := TestGt{Field: 3}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus lebih dari 5", errors[0].Message)
}

func TestGetValidationMessage_Gte(t *testing.T) {
	req := TestGte{Field: 4}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus lebih dari atau sama dengan 5", errors[0].Message)
}

func TestGetValidationMessage_Oneof(t *testing.T) {
	req := TestOneof{Field: "invalid"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus salah satu dari: red green blue", errors[0].Message)
}

func TestGetValidationMessage_URL(t *testing.T) {
	req := TestURL{URL: "invalid-url"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Format URL tidak valid", errors[0].Message)
}

func TestGetValidationMessage_URI(t *testing.T) {
	req := TestURI{URI: "invalid uri"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Format URI tidak valid", errors[0].Message)
}

func TestGetValidationMessage_Alpha(t *testing.T) {
	req := TestAlpha{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field hanya boleh berisi huruf", errors[0].Message)
}

func TestGetValidationMessage_Alphanum(t *testing.T) {
	req := TestAlphanum{Field: "abc-123"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field hanya boleh berisi huruf dan angka", errors[0].Message)
}

func TestGetValidationMessage_Numeric(t *testing.T) {
	req := TestNumeric{Field: "12ab"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field hanya boleh berisi angka", errors[0].Message)
}

func TestGetValidationMessage_Number(t *testing.T) {
	req := TestNumber{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa angka", errors[0].Message)
}

func TestGetValidationMessage_Hexadecimal(t *testing.T) {
	req := TestHexadecimal{Field: "xyz"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa hexadecimal", errors[0].Message)
}

func TestGetValidationMessage_Hexcolor(t *testing.T) {
	req := TestHexcolor{Field: "ggg"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna hex", errors[0].Message)
}

func TestGetValidationMessage_Rgb(t *testing.T) {
	req := TestRgb{Field: "not-rgb"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna RGB", errors[0].Message)
}

func TestGetValidationMessage_Rgba(t *testing.T) {
	req := TestRgba{Field: "not-rgba"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna RGBA", errors[0].Message)
}

func TestGetValidationMessage_Hsl(t *testing.T) {
	req := TestHsl{Field: "not-hsl"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna HSL", errors[0].Message)
}

func TestGetValidationMessage_Hsla(t *testing.T) {
	req := TestHsla{Field: "not-hsla"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna HSLA", errors[0].Message)
}

func TestGetValidationMessage_UUID(t *testing.T) {
	req := TestUUID{Field: "not-uuid"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID", errors[0].Message)
}

func TestGetValidationMessage_UUID3(t *testing.T) {
	req := TestUUID3{Field: "not-uuid3"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID versi 3", errors[0].Message)
}

func TestGetValidationMessage_UUID4(t *testing.T) {
	req := TestUUID4{Field: "not-uuid4"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID versi 4", errors[0].Message)
}

func TestGetValidationMessage_UUID5(t *testing.T) {
	req := TestUUID5{Field: "not-uuid5"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID versi 5", errors[0].Message)
}

func TestGetValidationMessage_ISBN(t *testing.T) {
	req := TestISBN{Field: "not-isbn"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa ISBN", errors[0].Message)
}

func TestGetValidationMessage_ISBN10(t *testing.T) {
	req := TestISBN10{Field: "not-isbn10"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa ISBN-10", errors[0].Message)
}

func TestGetValidationMessage_ISBN13(t *testing.T) {
	req := TestISBN13{Field: "not-isbn13"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa ISBN-13", errors[0].Message)
}

func TestGetValidationMessage_Containsany(t *testing.T) {
	req := TestContainsany{Field: "abcdef"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus mengandung salah satu dari: !@#", errors[0].Message)
}

func TestGetValidationMessage_Contains(t *testing.T) {
	req := TestContains{Field: "test"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus mengandung: keyword", errors[0].Message)
}

func TestGetValidationMessage_Excludes(t *testing.T) {
	req := TestExcludes{Field: "forbidden"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh mengandung: forbidden", errors[0].Message)
}

func TestGetValidationMessage_Excludesall(t *testing.T) {
	req := TestExcludesall{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh mengandung karakter: 123", errors[0].Message)
}

func TestGetValidationMessage_Excludesrune(t *testing.T) {
	req := TestExcludesrune{Field: "test@"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh mengandung karakter: @", errors[0].Message)
}

func TestGetValidationMessage_Startswith(t *testing.T) {
	req := TestStartswith{Field: "test"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus diawali dengan: prefix", errors[0].Message)
}

func TestGetValidationMessage_Endswith(t *testing.T) {
	req := TestEndswith{Field: "test"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus diakhiri dengan: suffix", errors[0].Message)
}

func TestGetValidationMessage_Datetime(t *testing.T) {
	req := TestDatetime{Field: "invalid-date"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa tanggal dengan format: 2006-01-02", errors[0].Message)
}

// Test multiple errors at once

func TestValidateStruct_MultipleErrors(t *testing.T) {
	type TestMultiple struct {
		Name     string `validate:"required,min=3"`
		Email    string `validate:"required,email"`
		Age      int    `validate:"gte=18"`
		Password string `validate:"required,min=8"`
	}

	req := TestMultiple{
		Name:     "a",
		Email:    "not-email",
		Age:      15,
		Password: "short",
	}

	errors := httputil.ValidateStruct(req)

	assert.GreaterOrEqual(t, len(errors), 4)

	// Check that all errors have non-empty fields and messages
	for _, err := range errors {
		assert.NotEmpty(t, err.Field)
		assert.NotEmpty(t, err.Message)
	}
}
