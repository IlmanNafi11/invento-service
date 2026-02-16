package httputil_test

import (
	"invento-service/internal/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStruct_Alpha_Valid(t *testing.T) {
	t.Parallel()
	req := TestAlpha{Field: "abc"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Alpha_Invalid(t *testing.T) {
	t.Parallel()
	req := TestAlpha{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "hanya boleh berisi huruf")
}

func TestValidateStruct_Alphanum_Valid(t *testing.T) {
	t.Parallel()
	req := TestAlphanum{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_Alphanum_Invalid(t *testing.T) {
	t.Parallel()
	req := TestAlphanum{Field: "abc-123"}
	errors := httputil.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "hanya boleh berisi huruf dan angka")
}

func TestValidateStruct_Numeric_Valid(t *testing.T) {
	t.Parallel()
	req := TestNumeric{Field: "123"}
	errors := httputil.ValidateStruct(req)
	assert.Empty(t, errors)
}
