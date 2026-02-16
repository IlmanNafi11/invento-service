package httputil_test

import (
	"testing"

	"invento-service/internal/httputil"

	"github.com/stretchr/testify/assert"
)

func TestGetValidationMessage_Alpha(t *testing.T) {
	t.Parallel()
	req := TestAlpha{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field hanya boleh berisi huruf", errors[0].Message)
}

func TestGetValidationMessage_Alphanum(t *testing.T) {
	t.Parallel()
	req := TestAlphanum{Field: "abc-123"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field hanya boleh berisi huruf dan angka", errors[0].Message)
}

func TestGetValidationMessage_Numeric(t *testing.T) {
	t.Parallel()
	req := TestNumeric{Field: "12ab"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field hanya boleh berisi angka", errors[0].Message)
}

func TestGetValidationMessage_Number(t *testing.T) {
	t.Parallel()
	req := TestNumber{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa angka", errors[0].Message)
}

func TestGetValidationMessage_Hexadecimal(t *testing.T) {
	t.Parallel()
	req := TestHexadecimal{Field: "xyz"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa hexadecimal", errors[0].Message)
}

func TestGetValidationMessage_Hexcolor(t *testing.T) {
	t.Parallel()
	req := TestHexcolor{Field: "ggg"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna hex", errors[0].Message)
}

func TestGetValidationMessage_Rgb(t *testing.T) {
	t.Parallel()
	req := TestRgb{Field: "not-rgb"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna RGB", errors[0].Message)
}

func TestGetValidationMessage_Rgba(t *testing.T) {
	t.Parallel()
	req := TestRgba{Field: "not-rgba"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna RGBA", errors[0].Message)
}

func TestGetValidationMessage_Hsl(t *testing.T) {
	t.Parallel()
	req := TestHsl{Field: "not-hsl"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna HSL", errors[0].Message)
}

func TestGetValidationMessage_Hsla(t *testing.T) {
	t.Parallel()
	req := TestHsla{Field: "not-hsla"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa warna HSLA", errors[0].Message)
}

func TestGetValidationMessage_UUID(t *testing.T) {
	t.Parallel()
	req := TestUUID{Field: "not-uuid"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID", errors[0].Message)
}

func TestGetValidationMessage_UUID3(t *testing.T) {
	t.Parallel()
	req := TestUUID3{Field: "not-uuid3"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID versi 3", errors[0].Message)
}

func TestGetValidationMessage_UUID4(t *testing.T) {
	t.Parallel()
	req := TestUUID4{Field: "not-uuid4"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID versi 4", errors[0].Message)
}
