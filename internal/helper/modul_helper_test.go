package helper_test

import (
	"invento-service/config"
	"invento-service/internal/helper"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModulHelper(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
		},
	}

	modulHelper := helper.NewModulHelper(cfg)

	assert.NotNil(t, modulHelper)
}

func TestModulHelper_GenerateModulIdentifier(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
	}
	modulHelper := helper.NewModulHelper(cfg)

	identifier, err := modulHelper.GenerateModulIdentifier()

	assert.NoError(t, err)
	assert.Len(t, identifier, 10)
	// Should only contain lowercase letters and numbers
	for _, c := range identifier {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	}
}

func TestModulHelper_GenerateModulIdentifier_Uniqueness(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
	}
	modulHelper := helper.NewModulHelper(cfg)

	// Generate multiple identifiers and check they're different
	identifiers := make(map[string]bool)
	for i := 0; i < 100; i++ {
		identifier, err := modulHelper.GenerateModulIdentifier()
		assert.NoError(t, err)
		assert.False(t, identifiers[identifier], "Identifier should be unique")
		identifiers[identifier] = true
	}
}

func TestModulHelper_BuildModulPath(t *testing.T) {
	tests := []struct {
		name       string
		env        string
		userID     string
		identifier string
		filename   string
	}{
		{
			name:       "Development environment",
			env:        "development",
			userID:     "123",
			identifier: "abc123",
			filename:   "file.pdf",
		},
		{
			name:       "Production environment",
			env:        "production",
			userID:     "456",
			identifier: "xyz789",
			filename:   "doc.docx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{Env: tt.env},
				Upload: config.UploadConfig{
					PathDevelopment: "/dev/uploads",
					PathProduction:  "/prod/uploads",
				},
			}
			modulHelper := helper.NewModulHelper(cfg)

			path := modulHelper.BuildModulPath(tt.userID, tt.identifier, tt.filename)

			assert.Contains(t, path, "moduls")
			assert.Contains(t, path, tt.identifier)
			assert.Contains(t, path, tt.filename)
		})
	}
}

func TestModulHelper_BuildModulDirectory(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
		Upload: config.UploadConfig{
			PathDevelopment: "/uploads",
		},
	}
	modulHelper := helper.NewModulHelper(cfg)

	userID := "123"
	identifier := "test-id"

	dir := modulHelper.BuildModulDirectory(userID, identifier)

	assert.Contains(t, dir, "moduls")
	assert.Contains(t, dir, "123")
	assert.Contains(t, dir, identifier)
}

func TestModulHelper_ValidateModulFile(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
	}
	modulHelper := helper.NewModulHelper(cfg)

	tests := []struct {
		name        string
		filename    string
		size        int64
		shouldError bool
	}{
		{
			name:        "Valid PDF file",
			filename:    "document.pdf",
			size:        1024 * 1024,
			shouldError: false,
		},
		{
			name:        "Valid DOCX file",
			filename:    "report.docx",
			size:        5 * 1024 * 1024,
			shouldError: false,
		},
		{
			name:        "Valid XLSX file",
			filename:    "data.xlsx",
			size:        10 * 1024 * 1024,
			shouldError: false,
		},
		{
			name:        "Valid PPTX file",
			filename:    "presentation.pptx",
			size:        20 * 1024 * 1024,
			shouldError: false,
		},
		{
			name:        "Invalid extension - txt",
			filename:    "document.txt",
			size:        1024,
			shouldError: true,
		},
		{
			name:        "Invalid extension - jpg",
			filename:    "image.jpg",
			size:        1024,
			shouldError: true,
		},
		{
			name:        "File too large - over 50MB",
			filename:    "large.pdf",
			size:        51 * 1024 * 1024,
			shouldError: true,
		},
		{
			name:        "Exactly 50MB",
			filename:    "max.pdf",
			size:        50 * 1024 * 1024,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock file header
			fileHeader := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     tt.size,
			}

			err := modulHelper.ValidateModulFile(fileHeader)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModulHelper_ValidateModulFileSize(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
	}
	modulHelper := helper.NewModulHelper(cfg)

	tests := []struct {
		name        string
		size        int64
		shouldError bool
	}{
		{"Valid size - 1MB", 1024 * 1024, false},
		{"Valid size - 50MB", 50 * 1024 * 1024, false},
		{"Invalid size - 0", 0, true},
		{"Invalid size - negative", -1, true},
		{"Invalid size - over 50MB", 51 * 1024 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modulHelper.ValidateModulFileSize(tt.size)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateModulFileExtension(t *testing.T) {
	tests := []struct {
		name        string
		tipe        string
		shouldError bool
	}{
		{"Valid - docx", "docx", false},
		{"Valid - xlsx", "xlsx", false},
		{"Valid - pdf", "pdf", false},
		{"Valid - pptx", "pptx", false},
		{"Invalid - txt", "txt", true},
		{"Invalid - pdf with extra", "pdf2", true},
		{"Invalid - empty", "", true},
		{"Invalid - PDF (uppercase)", "PDF", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := helper.ValidateModulFileExtension(tt.tipe)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetModulFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		tipe     string
		expected string
	}{
		{"docx type", "docx", ".docx"},
		{"xlsx type", "xlsx", ".xlsx"},
		{"pdf type", "pdf", ".pdf"},
		{"pptx type", "pptx", ".pptx"},
		{"unknown type", "unknown", ""},
		{"empty type", "", ""},
		{"invalid type", "txt", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.GetModulFileExtension(tt.tipe)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModulHelper_GetBasePath(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected string
	}{
		{
			name:     "Development environment",
			env:      "development",
			expected: "/dev/path",
		},
		{
			name:     "Production environment",
			env:      "production",
			expected: "/prod/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{Env: tt.env},
				Upload: config.UploadConfig{
					PathDevelopment: "/dev/path",
					PathProduction:  "/prod/path",
				},
			}
			modulHelper := helper.NewModulHelper(cfg)

			// We can't directly test getBasePath as it's private,
			// but we can verify the behavior through BuildModulPath
			path := modulHelper.BuildModulPath("1", "id", "file.pdf")
			assert.Contains(t, path, tt.expected)
		})
	}
}

func TestModulHelper_EdgeCases(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
		Upload: config.UploadConfig{
			PathDevelopment: "/uploads",
		},
	}
	modulHelper := helper.NewModulHelper(cfg)

	t.Run("BuildModulPath with empty filename", func(t *testing.T) {
		path := modulHelper.BuildModulPath("123", "id", "")
		assert.NotEmpty(t, path)
	})

	t.Run("BuildModulDirectory with zero userID", func(t *testing.T) {
		dir := modulHelper.BuildModulDirectory("0", "test-id")
		assert.Contains(t, dir, "0")
	})

	t.Run("BuildModulDirectory with empty identifier", func(t *testing.T) {
		dir := modulHelper.BuildModulDirectory("123", "")
		assert.Contains(t, dir, "123")
	})
}

func TestModulHelper_FileSizeBoundary(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
	}
	modulHelper := helper.NewModulHelper(cfg)

	// Test exact boundary: 50MB
	exactMaxSize := int64(50 * 1024 * 1024)
	fileHeader := &multipart.FileHeader{
		Filename: "test.pdf",
		Size:     exactMaxSize,
	}
	err := modulHelper.ValidateModulFile(fileHeader)
	assert.NoError(t, err, "File at exactly 50MB should be valid")

	// Test one byte over
	overMaxSize := int64(50*1024*1024 + 1)
	fileHeader.Size = overMaxSize
	err = modulHelper.ValidateModulFile(fileHeader)
	assert.Error(t, err, "File over 50MB should be rejected")
}

func TestModulHelper_CaseInsensitiveExtension(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
	}
	modulHelper := helper.NewModulHelper(cfg)

	tests := []string{
		"file.PDF",
		"file.Pdf",
		"file.DOCX",
		"file.DocX",
		"file.XLSX",
		"file.Xlsx",
		"file.PPTX",
		"file.Pptx",
	}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			fileHeader := &multipart.FileHeader{
				Filename: filename,
				Size:     1024,
			}

			// GetFileExtension converts to lowercase, so these should pass
			err := modulHelper.ValidateModulFile(fileHeader)
			assert.NoError(t, err, "Extensions should be case-insensitive (GetFileExtension converts to lowercase)")
		})
	}
}
