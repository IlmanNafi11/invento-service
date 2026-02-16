package httputil_test

import (
	dto "invento-service/internal/dto"
	"invento-service/internal/httputil"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidateStruct_ValidData(t *testing.T) {
	t.Parallel()
	req := dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	errors := httputil.ValidateStruct(req)

	assert.Empty(t, errors)
}

func TestValidateStruct_InvalidData(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	req := dto.AuthRequest{
		Email:    "",
		Password: "password123",
	}

	errors := httputil.ValidateStruct(req)

	assert.NotEmpty(t, errors)
	assert.Greater(t, len(errors), 0)
}

func TestValidateStruct_ShortPassword(t *testing.T) {
	t.Parallel()
	req := dto.AuthRequest{
		Email:    "test@example.com",
		Password: "123",
	}

	errors := httputil.ValidateStruct(req)

	assert.NotEmpty(t, errors)
	assert.Greater(t, len(errors), 0)
}

func TestValidateStruct_ResetPasswordRequest(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	req := TestRequired{Field: "value"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Required_Invalid(t *testing.T) {
	t.Parallel()
	req := TestRequired{Field: ""}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Equal(t, "Field", errors[0].Field)
	assert.Contains(t, errors[0].Message, "wajib diisi")
}

func TestValidateStruct_Email_Valid(t *testing.T) {
	t.Parallel()
	req := TestEmail{Email: "test@example.com"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Email_Invalid(t *testing.T) {
	t.Parallel()
	req := TestEmail{Email: "invalid-email"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "Format email tidak valid")
}

func TestValidateStruct_Min_Valid(t *testing.T) {
	t.Parallel()
	req := TestMin{Field: "abcdefgh"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Min_Invalid(t *testing.T) {
	t.Parallel()
	req := TestMin{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "minimal")
	assert.Contains(t, errors[0].Message, "5")
}

func TestValidateStruct_Max_Valid(t *testing.T) {
	t.Parallel()
	req := TestMax{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Max_Invalid(t *testing.T) {
	t.Parallel()
	req := TestMax{Field: "abcdefghijk"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "maksimal")
	assert.Contains(t, errors[0].Message, "10")
}

func TestValidateStruct_Len_Valid(t *testing.T) {
	t.Parallel()
	req := TestLen{Field: "abcde"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Len_Invalid(t *testing.T) {
	t.Parallel()
	req := TestLen{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus")
	assert.Contains(t, errors[0].Message, "5 karakter")
}

func TestValidateStruct_Eq_Valid(t *testing.T) {
	t.Parallel()
	req := TestEq{Field: "exact"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Eq_Invalid(t *testing.T) {
	t.Parallel()
	req := TestEq{Field: "other"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus sama dengan")
	assert.Contains(t, errors[0].Message, "exact")
}

func TestValidateStruct_Ne_Valid(t *testing.T) {
	t.Parallel()
	req := TestNe{Field: "allowed"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Ne_Invalid(t *testing.T) {
	t.Parallel()
	req := TestNe{Field: "forbidden"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh sama dengan")
	assert.Contains(t, errors[0].Message, "forbidden")
}

func TestValidateStruct_Lt_Valid(t *testing.T) {
	t.Parallel()
	req := TestLt{Field: 5}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Lt_Invalid(t *testing.T) {
	t.Parallel()
	req := TestLt{Field: 15}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus kurang dari")
	assert.Contains(t, errors[0].Message, "10")
}

func TestValidateStruct_Lte_Valid(t *testing.T) {
	t.Parallel()
	req := TestLte{Field: 10}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Lte_Invalid(t *testing.T) {
	t.Parallel()
	req := TestLte{Field: 15}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus kurang dari atau sama dengan")
}

func TestValidateStruct_Gt_Valid(t *testing.T) {
	t.Parallel()
	req := TestGt{Field: 10}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Gt_Invalid(t *testing.T) {
	t.Parallel()
	req := TestGt{Field: 3}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus lebih dari")
	assert.Contains(t, errors[0].Message, "5")
}

func TestValidateStruct_Gte_Valid(t *testing.T) {
	t.Parallel()
	req := TestGte{Field: 5}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Gte_Invalid(t *testing.T) {
	t.Parallel()
	req := TestGte{Field: 3}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus lebih dari atau sama dengan")
}

func TestValidateStruct_Oneof_Valid(t *testing.T) {
	t.Parallel()
	req := TestOneof{Field: "red"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Oneof_Invalid(t *testing.T) {
	t.Parallel()
	req := TestOneof{Field: "yellow"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus salah satu dari")
	assert.Contains(t, errors[0].Message, "red green blue")
}

func TestValidateStruct_URL_Valid(t *testing.T) {
	t.Parallel()
	req := TestURL{URL: "https://example.com"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_URL_Invalid(t *testing.T) {
	t.Parallel()
	req := TestURL{URL: "not-a-url"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "Format URL tidak valid")
}

func TestValidateStruct_URI_Valid(t *testing.T) {
	t.Parallel()
	req := TestURI{URI: "http://example.com/path"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_URI_Invalid(t *testing.T) {
	t.Parallel()
	req := TestURI{URI: "not a uri"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "Format URI tidak valid")
}
