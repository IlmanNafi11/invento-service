package validator

import (
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAllowedFileTypes_Success tests getting allowed file types
func TestGetAllowedFileTypes_Success(t *testing.T) {
	t.Parallel()
	types := GetAllowedFileTypes()

	assert.NotEmpty(t, types)
	assert.Contains(t, types, ".pdf")
	assert.Contains(t, types, ".docx")
	assert.Contains(t, types, ".jpg")
	assert.Contains(t, types, ".png")
}

// TestGetMaxUploadSize_Success tests getting max upload size for different contexts
func TestGetMaxUploadSize_Success(t *testing.T) {
	t.Parallel()
	size := GetMaxUploadSize("avatar")
	assert.Equal(t, int64(2*1024*1024), size)

	size = GetMaxUploadSize("document")
	assert.Equal(t, int64(10*1024*1024), size)

	size = GetMaxUploadSize("video")
	assert.Equal(t, int64(100*1024*1024), size)

	size = GetMaxUploadSize("unknown")
	assert.Equal(t, int64(5*1024*1024), size)
}

// TestValidateFile_Success tests successful file validation
func TestValidateFile_Success(t *testing.T) {
	t.Parallel()
	tmpfile, err := os.CreateTemp("", "test")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := strings.Repeat("test", 100)
	tmpfile.WriteString(content)
	tmpfile.Close()

	fileHeader := &multipart.FileHeader{
		Filename: "test.pdf",
		Size:     int64(len(content)),
	}

	err = ValidateFile(fileHeader, []string{".pdf", ".docx"}, 10*1024*1024)
	assert.NoError(t, err)
}

// TestValidateFile_InvalidType tests file validation with invalid type
func TestValidateFile_InvalidType(t *testing.T) {
	t.Parallel()
	fileHeader := &multipart.FileHeader{
		Filename: "test.exe",
		Size:     1024,
	}

	err := ValidateFile(fileHeader, []string{".pdf", ".docx"}, 10*1024*1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

// TestValidateFile_TooLarge tests file validation with size exceeded
func TestValidateFile_TooLarge(t *testing.T) {
	t.Parallel()
	fileHeader := &multipart.FileHeader{
		Filename: "test.pdf",
		Size:     20 * 1024 * 1024,
	}

	err := ValidateFile(fileHeader, []string{".pdf"}, 10*1024*1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

// TestValidateFile_NilHeader tests file validation with nil header
func TestValidateFile_NilHeader(t *testing.T) {
	t.Parallel()
	err := ValidateFile(nil, []string{".pdf"}, 10*1024*1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestValidateFile_NotFileHeader tests validation with wrong type
func TestValidateFile_NotFileHeader(t *testing.T) {
	t.Parallel()
	err := ValidateFile("string", []string{".pdf"}, 10*1024*1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a FileHeader")
}

// TestParseFormFile_Success tests successful form file parsing
func TestParseFormFile_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		file, err := ParseFormFile(c, "document", []string{".pdf"}, 10*1024*1024)
		assert.Error(t, err)
		assert.Nil(t, file)
		return c.SendStatus(200)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestValidateMultipartForm_Success tests successful multipart form validation
func TestValidateMultipartForm_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		err := ValidateMultipartForm(c, 32<<20)
		assert.Error(t, err)
		return c.SendStatus(200)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestValidateMultipartForm_WithActualData tests with actual multipart data
func TestValidateMultipartForm_WithActualData(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		err := ValidateMultipartForm(c, 32<<20)
		assert.NoError(t, err)
		return c.SendStatus(200)
	})

	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "test")
	writer.Close()

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestGetFileExtension tests getting file extension
func TestGetFileExtension(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename string
		expected string
	}{
		{"test.pdf", ".pdf"},
		{"document.docx", ".docx"},
		{"archive.tar.gz", ".gz"},
		{"noextension", ""},
		{".hidden", ".hidden"},
		{"file.TXT", ".TXT"},
		{"image.jpeg", ".jpeg"},
		{"data.json", ".json"},
		{"script.js", ".js"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getFileExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateFile_AllowedTypesWildcard tests wildcard allowed types
func TestValidateFile_AllowedTypesWildcard(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		filename     string
		allowedTypes []string
		shouldPass   bool
	}{
		{"PDFInList", "test.pdf", []string{".pdf", ".docx"}, true},
		{"DOCXInList", "test.docx", []string{".pdf", ".docx"}, true},
		{"EXENotInList", "test.exe", []string{".pdf", ".docx"}, false},
		{"EmptyList", "test.pdf", []string{}, true},
		{"NoExtension", "testfile", []string{".pdf"}, false},
		{"CaseInsensitive", "test.PDF", []string{".pdf"}, true},
		{"JPGAllowed", "test.jpg", []string{".jpg", ".png"}, true},
		{"PNGAllowed", "test.png", []string{".jpg", ".png"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     1024,
			}

			err := ValidateFile(fileHeader, tt.allowedTypes, 10*1024*1024)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestValidateFile_SizeVariations tests various file sizes
func TestValidateFile_SizeVariations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		size       int64
		maxSize    int64
		shouldPass bool
	}{
		{"ZeroSize", 0, 1024, true},
		{"SmallSize", 100, 1024, true},
		{"ExactSize", 1024, 1024, true},
		{"OverSize", 1025, 1024, false},
		{"NoLimit", 999999, 0, true},
		{"VerySmall", 1, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := &multipart.FileHeader{
				Filename: "test.pdf",
				Size:     tt.size,
			}

			err := ValidateFile(fileHeader, []string{".pdf"}, tt.maxSize)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestValidateFile_MultipleExtensions tests handling multiple extensions
func TestValidateFile_MultipleExtensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		filename     string
		allowedTypes []string
		shouldPass   bool
	}{
		{"TarGz", "archive.tar.gz", []string{".gz"}, true},
		{"TarGzNotInList", "archive.tar.gz", []string{".tar"}, false},
		{"DoubleExtension", "file.min.js", []string{".js"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     1024,
			}

			err := ValidateFile(fileHeader, tt.allowedTypes, 10*1024*1024)
			if tt.shouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestValidateFile_ErrorMessages tests error message formats
func TestValidateFile_ErrorMessages(t *testing.T) {
	t.Parallel()
	t.Run("InvalidTypeError", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "test.exe",
			Size:     1024,
		}

		err := ValidateFile(fileHeader, []string{".pdf", ".docx"}, 10*1024*1024)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
		assert.Contains(t, err.Error(), ".exe")
	})

	t.Run("SizeError", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "test.pdf",
			Size:     20 * 1024 * 1024,
		}

		err := ValidateFile(fileHeader, []string{".pdf"}, 10*1024*1024)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum")
	})

	t.Run("NilError", func(t *testing.T) {
		err := ValidateFile(nil, []string{".pdf"}, 10*1024*1024)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("NotFileHeaderError", func(t *testing.T) {
		err := ValidateFile("not a file header", []string{".pdf"}, 10*1024*1024)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a FileHeader")
	})
}

// TestValidateFile_AllExtensions tests all common file extensions
func TestValidateFile_AllExtensions(t *testing.T) {
	t.Parallel()
	extensions := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".zip", ".rar", ".jpg", ".jpeg", ".png"}

	for _, ext := range extensions {
		t.Run(ext, func(t *testing.T) {
			fileHeader := &multipart.FileHeader{
				Filename: "test" + ext,
				Size:     1024,
			}

			err := ValidateFile(fileHeader, extensions, 10*1024*1024)
			assert.NoError(t, err)
		})
	}
}

// TestValidateFile_NoDotInExtension tests extension without dot
func TestValidateFile_NoDotInExtension(t *testing.T) {
	t.Parallel()
	fileHeader := &multipart.FileHeader{
		Filename: "test.pdf",
		Size:     1024,
	}

	// Allowed types without dot - should still work
	err := ValidateFile(fileHeader, []string{"pdf", "docx"}, 10*1024*1024)
	assert.NoError(t, err)
}
