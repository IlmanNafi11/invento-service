package httputil_test

import (
	"invento-service/internal/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidationMessage_UUID5(t *testing.T) {
	t.Parallel()
	req := TestUUID5{Field: "not-uuid5"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa UUID versi 5", errors[0].Message)
}

func TestGetValidationMessage_ISBN(t *testing.T) {
	t.Parallel()
	req := TestISBN{Field: "not-isbn"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa ISBN", errors[0].Message)
}

func TestGetValidationMessage_ISBN10(t *testing.T) {
	t.Parallel()
	req := TestISBN10{Field: "not-isbn10"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa ISBN-10", errors[0].Message)
}

func TestGetValidationMessage_ISBN13(t *testing.T) {
	t.Parallel()
	req := TestISBN13{Field: "not-isbn13"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa ISBN-13", errors[0].Message)
}

func TestGetValidationMessage_Containsany(t *testing.T) {
	t.Parallel()
	req := TestContainsany{Field: "abcdef"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus mengandung salah satu dari: !@#", errors[0].Message)
}

func TestGetValidationMessage_Contains(t *testing.T) {
	t.Parallel()
	req := TestContains{Field: "test"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus mengandung: keyword", errors[0].Message)
}

func TestGetValidationMessage_Excludes(t *testing.T) {
	t.Parallel()
	req := TestExcludes{Field: "forbidden"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh mengandung: forbidden", errors[0].Message)
}

func TestGetValidationMessage_Excludesall(t *testing.T) {
	t.Parallel()
	req := TestExcludesall{Field: "abc123"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh mengandung karakter: 123", errors[0].Message)
}

func TestGetValidationMessage_Excludesrune(t *testing.T) {
	t.Parallel()
	req := TestExcludesrune{Field: "test@"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field tidak boleh mengandung karakter: @", errors[0].Message)
}

func TestGetValidationMessage_Startswith(t *testing.T) {
	t.Parallel()
	req := TestStartswith{Field: "test"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus diawali dengan: prefix", errors[0].Message)
}

func TestGetValidationMessage_Endswith(t *testing.T) {
	t.Parallel()
	req := TestEndswith{Field: "test"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus diakhiri dengan: suffix", errors[0].Message)
}

func TestGetValidationMessage_Datetime(t *testing.T) {
	t.Parallel()
	req := TestDatetime{Field: "invalid-date"}
	errors := httputil.ValidateStruct(req)
	assert.Equal(t, "Field harus berupa tanggal dengan format: 2006-01-02", errors[0].Message)
}

// Test multiple errors at once

func TestValidateStruct_MultipleErrors(t *testing.T) {
	t.Parallel()
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
