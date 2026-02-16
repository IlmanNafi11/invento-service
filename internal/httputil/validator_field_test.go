package httputil_test

import (
	"testing"

	"invento-service/internal/httputil"

	"github.com/stretchr/testify/assert"
)

func TestValidateStruct_Numeric_Invalid(t *testing.T) {
	t.Parallel()
	req := TestNumeric{Field: "12a3"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "hanya boleh berisi angka")
}

func TestValidateStruct_Number_Valid(t *testing.T) {
	t.Parallel()
	req := TestNumber{Field: "123"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Number_Invalid(t *testing.T) {
	t.Parallel()
	req := TestNumber{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa angka")
}

func TestValidateStruct_Hexadecimal_Valid(t *testing.T) {
	t.Parallel()
	req := TestHexadecimal{Field: "abc123DEF"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hexadecimal_Invalid(t *testing.T) {
	t.Parallel()
	req := TestHexadecimal{Field: "xyz"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa hexadecimal")
}

func TestValidateStruct_Hexcolor_Valid(t *testing.T) {
	t.Parallel()
	req := TestHexcolor{Field: "#fff"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hexcolor_Invalid(t *testing.T) {
	t.Parallel()
	req := TestHexcolor{Field: "ggg"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna hex")
}

func TestValidateStruct_Rgb_Valid(t *testing.T) {
	t.Parallel()
	req := TestRgb{Field: "rgb(255, 255, 255)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Rgb_Invalid(t *testing.T) {
	t.Parallel()
	req := TestRgb{Field: "not rgb"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna RGB")
}

func TestValidateStruct_Rgba_Valid(t *testing.T) {
	t.Parallel()
	req := TestRgba{Field: "rgba(255, 255, 255, 0.5)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Rgba_Invalid(t *testing.T) {
	t.Parallel()
	req := TestRgba{Field: "not rgba"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna RGBA")
}

func TestValidateStruct_Hsl_Valid(t *testing.T) {
	t.Parallel()
	req := TestHsl{Field: "hsl(120, 100%, 50%)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hsl_Invalid(t *testing.T) {
	t.Parallel()
	req := TestHsl{Field: "not hsl"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna HSL")
}

func TestValidateStruct_Hsla_Valid(t *testing.T) {
	t.Parallel()
	req := TestHsla{Field: "hsla(120, 100%, 50%, 0.5)"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Hsla_Invalid(t *testing.T) {
	t.Parallel()
	req := TestHsla{Field: "not hsla"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa warna HSLA")
}

func TestValidateStruct_UUID_Valid(t *testing.T) {
	t.Parallel()
	req := TestUUID{Field: "550e8400-e29b-41d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID_Invalid(t *testing.T) {
	t.Parallel()
	req := TestUUID{Field: "not-uuid"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID")
}

func TestValidateStruct_UUID3_Valid(t *testing.T) {
	t.Parallel()
	req := TestUUID3{Field: "550e8400-e29b-31d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID3_Invalid(t *testing.T) {
	t.Parallel()
	req := TestUUID3{Field: "not-uuid3"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID versi 3")
}

func TestValidateStruct_UUID4_Valid(t *testing.T) {
	t.Parallel()
	req := TestUUID4{Field: "550e8400-e29b-41d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID4_Invalid(t *testing.T) {
	t.Parallel()
	req := TestUUID4{Field: "not-uuid4"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID versi 4")
}

func TestValidateStruct_UUID5_Valid(t *testing.T) {
	t.Parallel()
	req := TestUUID5{Field: "550e8400-e29b-51d4-a716-446655440000"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_UUID5_Invalid(t *testing.T) {
	t.Parallel()
	req := TestUUID5{Field: "not-uuid5"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa UUID versi 5")
}

func TestValidateStruct_ISBN_Valid(t *testing.T) {
	t.Parallel()
	req := TestISBN{Field: "978-3-16-148410-0"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_ISBN_Invalid(t *testing.T) {
	t.Parallel()
	req := TestISBN{Field: "not-isbn"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa ISBN")
}

func TestValidateStruct_ISBN10_Valid(t *testing.T) {
	t.Parallel()
	req := TestISBN10{Field: "0-306-40615-2"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_ISBN10_Invalid(t *testing.T) {
	t.Parallel()
	req := TestISBN10{Field: "not-isbn10"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa ISBN-10")
}

func TestValidateStruct_ISBN13_Valid(t *testing.T) {
	t.Parallel()
	req := TestISBN13{Field: "978-3-16-148410-0"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_ISBN13_Invalid(t *testing.T) {
	t.Parallel()
	req := TestISBN13{Field: "not-isbn13"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa ISBN-13")
}

func TestValidateStruct_Containsany_Valid(t *testing.T) {
	t.Parallel()
	req := TestContainsany{Field: "test@example.com"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Containsany_Invalid(t *testing.T) {
	t.Parallel()
	req := TestContainsany{Field: "testexamplecom"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus mengandung salah satu dari")
	assert.Contains(t, errors[0].Message, "!@#")
}

func TestValidateStruct_Contains_Valid(t *testing.T) {
	t.Parallel()
	req := TestContains{Field: "this is a keyword test"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Contains_Invalid(t *testing.T) {
	t.Parallel()
	req := TestContains{Field: "this is a test"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus mengandung")
	assert.Contains(t, errors[0].Message, "keyword")
}

func TestValidateStruct_Excludes_Valid(t *testing.T) {
	t.Parallel()
	req := TestExcludes{Field: "allowed text"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Excludes_Invalid(t *testing.T) {
	t.Parallel()
	req := TestExcludes{Field: "forbidden text"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh mengandung")
	assert.Contains(t, errors[0].Message, "forbidden")
}

func TestValidateStruct_Excludesall_Valid(t *testing.T) {
	t.Parallel()
	req := TestExcludesall{Field: "abcdef"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Excludesall_Invalid(t *testing.T) {
	t.Parallel()
	req := TestExcludesall{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh mengandung karakter")
	assert.Contains(t, errors[0].Message, "123")
}

func TestValidateStruct_Excludesrune_Valid(t *testing.T) {
	t.Parallel()
	req := TestExcludesrune{Field: "without-at"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Excludesrune_Invalid(t *testing.T) {
	t.Parallel()
	req := TestExcludesrune{Field: "with@sign"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "tidak boleh mengandung karakter")
	assert.Contains(t, errors[0].Message, "@")
}

func TestValidateStruct_Startswith_Valid(t *testing.T) {
	t.Parallel()
	req := TestStartswith{Field: "prefix_value"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Startswith_Invalid(t *testing.T) {
	t.Parallel()
	req := TestStartswith{Field: "value"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus diawali dengan")
	assert.Contains(t, errors[0].Message, "prefix")
}

func TestValidateStruct_Endswith_Valid(t *testing.T) {
	t.Parallel()
	req := TestEndswith{Field: "value_suffix"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Endswith_Invalid(t *testing.T) {
	t.Parallel()
	req := TestEndswith{Field: "value"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus diakhiri dengan")
	assert.Contains(t, errors[0].Message, "suffix")
}

func TestValidateStruct_Datetime_Valid(t *testing.T) {
	t.Parallel()
	req := TestDatetime{Field: "2023-12-25"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Datetime_Invalid(t *testing.T) {
	t.Parallel()
	req := TestDatetime{Field: "not-a-date"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "harus berupa tanggal dengan format")
	assert.Contains(t, errors[0].Message, "2006-01-02")
}

// Test getValidationMessage directly via ValidateStruct with Indonesian messages

func TestGetValidationMessage_Required(t *testing.T) {
	t.Parallel()
	req := TestRequired{Field: ""}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field wajib diisi", errors[0].Message)
}

func TestGetValidationMessage_Email(t *testing.T) {
	t.Parallel()
	req := TestEmail{Email: "invalid"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Format email tidak valid", errors[0].Message)
}

func TestGetValidationMessage_Min(t *testing.T) {
	t.Parallel()
	req := TestMin{Field: "ab"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field minimal 5 karakter", errors[0].Message)
}

func TestGetValidationMessage_Max(t *testing.T) {
	t.Parallel()
	req := TestMax{Field: "abcdefghijk"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field maksimal 10 karakter", errors[0].Message)
}

func TestGetValidationMessage_Len(t *testing.T) {
	t.Parallel()
	req := TestLen{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus 5 karakter", errors[0].Message)
}

func TestGetValidationMessage_Eq(t *testing.T) {
	t.Parallel()
	req := TestEq{Field: "wrong"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus sama dengan exact", errors[0].Message)
}

func TestGetValidationMessage_Ne(t *testing.T) {
	t.Parallel()
	req := TestNe{Field: "forbidden"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh sama dengan forbidden", errors[0].Message)
}

func TestGetValidationMessage_Lt(t *testing.T) {
	t.Parallel()
	req := TestLt{Field: 15}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus kurang dari 10", errors[0].Message)
}

func TestGetValidationMessage_Lte(t *testing.T) {
	t.Parallel()
	req := TestLte{Field: 11}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus kurang dari atau sama dengan 10", errors[0].Message)
}

func TestGetValidationMessage_Gt(t *testing.T) {
	t.Parallel()
	req := TestGt{Field: 3}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus lebih dari 5", errors[0].Message)
}

func TestGetValidationMessage_Gte(t *testing.T) {
	t.Parallel()
	req := TestGte{Field: 4}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus lebih dari atau sama dengan 5", errors[0].Message)
}

func TestGetValidationMessage_Oneof(t *testing.T) {
	t.Parallel()
	req := TestOneof{Field: "invalid"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus salah satu dari: red green blue", errors[0].Message)
}

func TestGetValidationMessage_URL(t *testing.T) {
	t.Parallel()
	req := TestURL{URL: "invalid-url"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Format URL tidak valid", errors[0].Message)
}

func TestGetValidationMessage_URI(t *testing.T) {
	t.Parallel()
	req := TestURI{URI: "invalid uri"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Format URI tidak valid", errors[0].Message)
}
