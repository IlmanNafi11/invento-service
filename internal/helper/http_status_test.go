package helper

import (
	"testing"
)

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"StatusOK", StatusOK, 200},
		{"StatusCreated", StatusCreated, 201},
		{"StatusNoContent", StatusNoContent, 204},
		{"StatusBadRequest", StatusBadRequest, 400},
		{"StatusUnauthorized", StatusUnauthorized, 401},
		{"StatusForbidden", StatusForbidden, 403},
		{"StatusNotFound", StatusNotFound, 404},
		{"StatusConflict", StatusConflict, 409},
		{"StatusPayloadTooLarge", StatusPayloadTooLarge, 413},
		{"StatusInternalServerError", StatusInternalServerError, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestStatusText(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{StatusOK, "OK"},
		{StatusCreated, "Created"},
		{StatusNoContent, "No Content"},
		{StatusBadRequest, "Bad Request"},
		{StatusUnauthorized, "Unauthorized"},
		{StatusForbidden, "Forbidden"},
		{StatusNotFound, "Not Found"},
		{StatusConflict, "Conflict"},
		{StatusPayloadTooLarge, "Payload Too Large"},
		{StatusInternalServerError, "Internal Server Error"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if text, ok := StatusText[tt.code]; !ok {
				t.Errorf("StatusText missing entry for %d", tt.code)
			} else if text != tt.expected {
				t.Errorf("StatusText[%d] = %s, want %s", tt.code, text, tt.expected)
			}
		})
	}
}

func TestDefaultMessages(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{StatusBadRequest, "Request tidak valid"},
		{StatusUnauthorized, "Tidak memiliki akses"},
		{StatusForbidden, "Akses ditolak"},
		{StatusNotFound, "Data tidak ditemukan"},
		{StatusConflict, "Data sudah ada"},
		{StatusPayloadTooLarge, "Ukuran data melebihi batas maksimal"},
		{StatusInternalServerError, "Terjadi kesalahan pada server"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if msg, ok := DefaultMessages[tt.code]; !ok {
				t.Errorf("DefaultMessages missing entry for %d", tt.code)
			} else if msg != tt.expected {
				t.Errorf("DefaultMessages[%d] = %s, want %s", tt.code, msg, tt.expected)
			}
		})
	}
}

func TestGetDefaultMessage(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "Bad Request message",
			code:     StatusBadRequest,
			expected: "Request tidak valid",
		},
		{
			name:     "Unauthorized message",
			code:     StatusUnauthorized,
			expected: "Tidak memiliki akses",
		},
		{
			name:     "Forbidden message",
			code:     StatusForbidden,
			expected: "Akses ditolak",
		},
		{
			name:     "Not Found message",
			code:     StatusNotFound,
			expected: "Data tidak ditemukan",
		},
		{
			name:     "Conflict message",
			code:     StatusConflict,
			expected: "Data sudah ada",
		},
		{
			name:     "Internal Server Error message",
			code:     StatusInternalServerError,
			expected: "Terjadi kesalahan pada server",
		},
		{
			name:     "Unknown status code returns default",
			code:     418,
			expected: "Terjadi kesalahan",
		},
		{
			name:     "Zero status code returns default",
			code:     0,
			expected: "Terjadi kesalahan",
		},
		{
			name:     "Negative status code returns default",
			code:     -1,
			expected: "Terjadi kesalahan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultMessage(tt.code)
			if result != tt.expected {
				t.Errorf("GetDefaultMessage(%d) = %s, want %s", tt.code, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultMessage_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"StatusOK (no default message)", StatusOK},
		{"StatusCreated (no default message)", StatusCreated},
		{"StatusNoContent (no default message)", StatusNoContent},
		{"Very large status code", 999},
		{"Status 202", 202},
		{"Status 206", 206},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultMessage(tt.code)
			if result == "" {
				t.Errorf("GetDefaultMessage(%d) returned empty string", tt.code)
			}
			// All should return the default error message
			if result != "Terjadi kesalahan" {
				// Check if it's a valid default message
				found := false
				for _, msg := range DefaultMessages {
					if result == msg {
						found = true
						break
					}
				}
				if !found && result != "Terjadi kesalahan" {
					t.Errorf("GetDefaultMessage(%d) = %s, expected default message", tt.code, result)
				}
			}
		})
	}
}

func TestStatusTextCompleteness(t *testing.T) {
	// Ensure all status constants have corresponding text entries
	statusConstants := []int{
		StatusOK, StatusCreated, StatusNoContent, StatusBadRequest,
		StatusUnauthorized, StatusForbidden, StatusNotFound, StatusConflict,
		StatusPayloadTooLarge, StatusInternalServerError,
	}

	for _, code := range statusConstants {
		if _, ok := StatusText[code]; !ok {
			t.Errorf("StatusText missing entry for constant %d", code)
		}
	}
}

func TestDefaultMessagesCompleteness(t *testing.T) {
	// Ensure error status codes have default messages
	errorCodes := []int{
		StatusBadRequest, StatusUnauthorized, StatusForbidden,
		StatusNotFound, StatusConflict, StatusPayloadTooLarge,
		StatusInternalServerError,
	}

	for _, code := range errorCodes {
		if _, ok := DefaultMessages[code]; !ok {
			t.Errorf("DefaultMessages missing entry for error code %d", code)
		}
	}
}
