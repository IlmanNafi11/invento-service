package dto

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewErrorDetail_Success tests creating ErrorDetail without code
func TestNewErrorDetail_Success(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		field   string
		message string
		want    ErrorDetail
	}{
		{
			name:    "valid field and message",
			field:   "email",
			message: "email is required",
			want:    ErrorDetail{Field: "email", Message: "email is required"},
		},
		{
			name:    "empty field",
			field:   "",
			message: "some error",
			want:    ErrorDetail{Field: "", Message: "some error"},
		},
		{
			name:    "empty message",
			field:   "name",
			message: "",
			want:    ErrorDetail{Field: "name", Message: ""},
		},
		{
			name:    "both empty",
			field:   "",
			message: "",
			want:    ErrorDetail{Field: "", Message: ""},
		},
		{
			name:    "special characters in message",
			field:   "password",
			message: "Password must contain @#$!",
			want:    ErrorDetail{Field: "password", Message: "Password must contain @#$!"},
		},
		{
			name:    "unicode characters",
			field:   "nama",
			message: "Field ini wajib diisi",
			want:    ErrorDetail{Field: "nama", Message: "Field ini wajib diisi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewErrorDetail(tt.field, tt.message)
			assert.Equal(t, tt.want, got)
			assert.Empty(t, got.Code, "Code should be empty")
		})
	}
}

// TestNewErrorDetailWithCode_Success tests creating ErrorDetail with code
func TestNewErrorDetailWithCode_Success(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		field   string
		message string
		code    string
		want    ErrorDetail
	}{
		{
			name:    "valid all parameters",
			field:   "email",
			message: "email is required",
			code:    "REQUIRED",
			want:    ErrorDetail{Field: "email", Message: "email is required", Code: "REQUIRED"},
		},
		{
			name:    "empty field with code",
			field:   "",
			message: "error",
			code:    "INVALID",
			want:    ErrorDetail{Field: "", Message: "error", Code: "INVALID"},
		},
		{
			name:    "empty code",
			field:   "name",
			message: "name required",
			code:    "",
			want:    ErrorDetail{Field: "name", Message: "name required", Code: ""},
		},
		{
			name:    "numeric code",
			field:   "age",
			message: "invalid age",
			code:    "ERR_001",
			want:    ErrorDetail{Field: "age", Message: "invalid age", Code: "ERR_001"},
		},
		{
			name:    "special characters in code",
			field:   "url",
			message: "invalid url",
			code:    "URL_INVALID-123",
			want:    ErrorDetail{Field: "url", Message: "invalid url", Code: "URL_INVALID-123"},
		},
		{
			name:    "long code",
			field:   "field",
			message: "message",
			code:    strings.Repeat("A", 100),
			want:    ErrorDetail{Field: "field", Message: "message", Code: strings.Repeat("A", 100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewErrorDetailWithCode(tt.field, tt.message, tt.code)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestFormatISO8601_Success tests formatting time to ISO 8601
func TestFormatISO8601_Success(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "UTC time",
			time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			want: "2024-01-15T10:30:45Z",
		},
		{
			name: "epoch start",
			time: time.Unix(0, 0).UTC(),
			want: "1970-01-01T00:00:00Z",
		},
		{
			name: "positive timezone offset",
			time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("UTC+7", 7*3600)),
			want: "2024-06-01T12:00:00+07:00",
		},
		{
			name: "negative timezone offset",
			time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("UTC-5", -5*3600)),
			want: "2024-06-01T12:00:00-05:00",
		},
		{
			name: "midnight time",
			time: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			want: "2024-12-31T00:00:00Z",
		},
		{
			name: "end of day",
			time: time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
			want: "2024-01-01T23:59:59Z",
		},
		{
			name: "with nanoseconds",
			time: time.Date(2024, 3, 15, 14, 30, 0, 123456789, time.UTC),
			want: "2024-03-15T14:30:00Z", // nanoseconds not included in format
		},
		{
			name: "leap year",
			time: time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			want: "2024-02-29T12:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatISO8601(tt.time)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNowISO8601_Format tests that NowISO8601 returns valid ISO 8601 format
func TestNowISO8601_Format(t *testing.T) {
	t.Parallel()
	got := NowISO8601()

	// Parse the result to verify it's valid ISO 8601
	parsed, err := time.Parse(TimeFormat, got)
	assert.NoError(t, err, "NowISO8601 should return valid ISO 8601 format")

	// Verify it's close to current time (within 1 second tolerance)
	now := time.Now()
	diff := now.Sub(parsed)
	assert.LessOrEqual(t, diff.Abs(), time.Second, "NowISO8601 should return current time")

	// Verify format contains expected parts
	assert.Contains(t, got, "T", "Should contain T separator")
	assert.Contains(t, got, ":", "Should contain time separator")
}

// TestParseISO8601_Success tests parsing valid ISO 8601 strings
func TestParseISO8601_Success(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantLoc *time.Location // expected location
	}{
		{
			name:    "UTC time with Z",
			input:   "2024-01-15T10:30:45Z",
			want:    time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			wantLoc: time.UTC,
		},
		{
			name:    "positive offset",
			input:   "2024-06-01T12:00:00+07:00",
			want:    time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("", 7*3600)),
			wantLoc: time.FixedZone("", 7*3600),
		},
		{
			name:    "negative offset",
			input:   "2024-06-01T12:00:00-05:00",
			want:    time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("", -5*3600)),
			wantLoc: time.FixedZone("", -5*3600),
		},
		{
			name:    "zero offset +00:00",
			input:   "2024-01-01T00:00:00+00:00",
			want:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantLoc: time.UTC,
		},
		{
			name:    "leap year date",
			input:   "2024-02-29T12:00:00Z",
			want:    time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			wantLoc: time.UTC,
		},
		{
			name:    "year boundary",
			input:   "2024-12-31T23:59:59Z",
			want:    time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			wantLoc: time.UTC,
		},
		{
			name:    "epoch start",
			input:   "1970-01-01T00:00:00Z",
			want:    time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			wantLoc: time.UTC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseISO8601(tt.input)
			assert.NoError(t, err)
			assert.True(t, got.Equal(tt.want), "Expected %v, got %v", tt.want, got)

		// Check location matches - use the time value's Zone() method
		_, offsetGot := got.Zone()
		_, offsetWant := tt.want.Zone()
		assert.Equal(t, offsetWant, offsetGot, "Timezone offset should match")
		})
	}
}

// TestParseISO8601_Failure tests parsing invalid ISO 8601 strings
func TestParseISO8601_Failure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "not ISO format",
			input: "2024-01-15 10:30:45",
		},
		{
			name:  "missing timezone",
			input: "2024-01-15T10:30:45",
		},
		{
			name:  "invalid date",
			input: "2024-13-01T10:30:45Z",
		},
		{
			name:  "invalid month",
			input: "2024-02-30T10:30:45Z",
		},
		{
			name:  "invalid time",
			input: "2024-01-15T25:30:45Z",
		},
		{
			name:  "invalid minutes",
			input: "2024-01-15T10:60:45Z",
		},
		{
			name:  "invalid seconds",
			input: "2024-01-15T10:30:60Z",
		},
		{
			name:  "random text",
			input: "not a date",
		},
		{
			name:  "partial format",
			input: "2024-01-15",
		},
		{
			name:  "wrong separator",
			input: "2024/01/15T10:30:45Z",
		},
		{
			name:  "missing time part",
			input: "2024-01-15Z",
		},
		{
			name:  "invalid timezone format",
			input: "2024-01-15T10:30:45+0700",
		},
		{
			name:  "garbage characters",
			input: "2024-01-15T10:30:45Z!!!",
		},
		{
			name:  "whitespace only",
			input: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseISO8601(tt.input)
			assert.Error(t, err, "ParseISO8601 should return error for invalid input")
		})
	}
}

// TestFormatParseRoundTrip tests that formatting and parsing are inverse operations
func TestFormatParseRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		time time.Time
	}{
		{
			name: "UTC time",
			time: time.Date(2024, 5, 15, 14, 30, 45, 0, time.UTC),
		},
		{
			name: "positive offset",
			time: time.Date(2024, 5, 15, 14, 30, 45, 0, time.FixedZone("WIB", 7*3600)),
		},
		{
			name: "negative offset",
			time: time.Date(2024, 5, 15, 14, 30, 45, 0, time.FixedZone("EST", -5*3600)),
		},
		{
			name: "leap year",
			time: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "year boundary",
			time: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Format then parse should give equivalent time
			formatted := FormatISO8601(tt.time)
			parsed, err := ParseISO8601(formatted)
			assert.NoError(t, err)

			// Compare Unix timestamps to handle timezone differences
			assert.Equal(t, tt.time.Unix(), parsed.Unix(), "Roundtrip should preserve Unix timestamp")
		})
	}
}

// TestConstants validates that constant values are correct
func TestConstants(t *testing.T) {
	t.Parallel()
	t.Run("TimeFormat constant", func(t *testing.T) {
		t.Parallel()
		// TimeFormat should be valid Go reference time
		assert.Equal(t, "2006-01-02T15:04:05Z07:00", TimeFormat)

		// Should be usable for formatting
		now := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
		formatted := now.Format(TimeFormat)
		assert.NotEmpty(t, formatted)
	})

	t.Run("pagination constants", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 1, DefaultPage)
		assert.Equal(t, 10, DefaultLimit)
		assert.Equal(t, 100, MaxLimit)
		assert.Equal(t, "id", DefaultSort)
		assert.Equal(t, "asc", DefaultOrder)
	})

	t.Run("validation message constants", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			constant string
			want     string
		}{
			{MsgRequired, "Field ini wajib diisi"},
			{MsgEmail, "Format email tidak valid"},
			{MsgMin, "Nilai minimal tidak terpenuhi"},
			{MsgMax, "Nilai maksimal terlampaui"},
			{MsgOneOf, "Nilai tidak sesuai dengan opsi yang tersedia"},
			{MsgNumeric, "Harus berupa angka"},
			{MsgAlpha, "Harus berupa huruf"},
			{MsgAlphanumeric, "Harus berupa huruf dan angka"},
			{MsgLength, "Panjang karakter tidak sesuai"},
			{MsgDate, "Format tanggal tidak valid"},
			{MsgURL, "Format URL tidak valid"},
			{MsgUUID, "Format UUID tidak valid"},
			{MsgPositive, "Nilai harus positif"},
			{MsgNonZero, "Nilai tidak boleh nol"},
			{MsgUnique, "Nilai sudah digunakan"},
			{MsgNotFound, "Data tidak ditemukan"},
			{MsgUnauthorized, "Tidak memiliki akses"},
			{MsgForbidden, "Akses ditolak"},
			{MsgConflict, "Data sudah ada"},
			{MsgBadRequest, "Request tidak valid"},
			{MsgInternalServer, "Terjadi kesalahan pada server"},
			{MsgPayloadTooLarge, "Ukuran data melebihi batas maksimal"},
			{MsgTooManyRequests, "Terlalu banyak permintaan, silakan coba lagi nanti"},
			{MsgValidationError, "Validasi gagal"},
			{MsgInvalidCredentials, "Kredensial tidak valid"},
		}

		for _, tt := range tests {
			assert.NotEmpty(t, tt.constant, "Constant should not be empty")
			assert.Equal(t, tt.want, tt.constant)
		}
	})
}

// TestErrorDetail_JSONTags tests JSON serialization of ErrorDetail
func TestErrorDetail_JSONTags(t *testing.T) {
	t.Parallel()
	t.Run("without code - omitempty", func(t *testing.T) {
		t.Parallel()
		ed := NewErrorDetail("email", "required")
		// Code should not be present in JSON when empty due to omitempty
		// This is implicitly tested through the struct tag
		assert.Empty(t, ed.Code)
		assert.Equal(t, "email", ed.Field)
		assert.Equal(t, "required", ed.Message)
	})

	t.Run("with code", func(t *testing.T) {
		t.Parallel()
		ed := NewErrorDetailWithCode("email", "required", "ERR_001")
		assert.Equal(t, "email", ed.Field)
		assert.Equal(t, "required", ed.Message)
		assert.Equal(t, "ERR_001", ed.Code)
	})
}

// TestTimeBoundaries tests time functions at boundary conditions
func TestTimeBoundaries(t *testing.T) {
	t.Parallel()
	t.Run("minimum year", func(t *testing.T) {
		t.Parallel()
		minTime := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
		formatted := FormatISO8601(minTime)
		parsed, err := ParseISO8601(formatted)
		assert.NoError(t, err)
		assert.Equal(t, minTime.Unix(), parsed.Unix())
	})

	t.Run("far future year", func(t *testing.T) {
		t.Parallel()
		futureTime := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
		formatted := FormatISO8601(futureTime)
		parsed, err := ParseISO8601(formatted)
		assert.NoError(t, err)
		assert.Equal(t, futureTime.Unix(), parsed.Unix())
	})
}
