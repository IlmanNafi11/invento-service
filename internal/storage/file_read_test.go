package storage_test

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"invento-service/internal/storage"
)

func TestGenerateUniqueIdentifier(t *testing.T) {
	t.Parallel()
	id1, err := storage.GenerateUniqueIdentifier(10)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(id1)) // hex encoding doubles the length

	id2, err := storage.GenerateUniqueIdentifier(10)
	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)
}

func TestGenerateRandomString(t *testing.T) {
	t.Parallel()
	str1, err := storage.GenerateRandomString(10)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(str1))

	str2, err := storage.GenerateRandomString(10)
	assert.NoError(t, err)
	assert.NotEqual(t, str1, str2)

	// Check that only valid characters are used
	for _, c := range str1 {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	}
}

func TestGetFileExtension(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename    string
		expectedExt string
	}{
		{"test.txt", ".txt"},
		{"document.pdf", ".pdf"},
		{"image.JPG", ".jpg"},
		{"archive.ZIP", ".zip"},
		{"file", ""},
		{"path/to/file.txt", ".txt"},
		{"file.with.multiple.dots.txt", ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			ext := storage.GetFileExtension(tt.filename)
			assert.Equal(t, tt.expectedExt, ext)
		})
	}
}

func TestValidateZipFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "valid zip file",
			filename:    "archive.zip",
			expectError: false,
		},
		{
			name:        "uppercase zip",
			filename:    "archive.ZIP",
			expectError: false,
		},
		{
			name:        "not a zip file",
			filename:    "document.pdf",
			expectError: true,
		},
		{
			name:        "no extension",
			filename:    "file",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a mock file header
			header := &multipart.FileHeader{
				Filename: tt.filename,
			}

			err := storage.ValidateZipFile(header)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "zip")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateModulFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{"valid pdf", "document.pdf", false},
		{"valid docx", "file.docx", false},
		{"valid xlsx", "spreadsheet.xlsx", false},
		{"valid pptx", "presentation.pptx", false},
		{"valid jpg", "image.jpg", false},
		{"valid jpeg", "image.jpeg", false},
		{"valid png", "photo.png", false},
		{"valid gif", "animation.gif", false},
		{"invalid zip", "archive.zip", true},
		{"invalid txt", "file.txt", true},
		{"no extension", "file", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			header := &multipart.FileHeader{
				Filename: tt.filename,
			}

			err := storage.ValidateModulFile(header)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename     string
		expectedType string
	}{
		{"document.pdf", "pdf"},
		{"file.docx", "docx"},
		{"spreadsheet.xlsx", "xlsx"},
		{"presentation.pptx", "pptx"},
		{"image.jpg", "jpeg"},
		{"image.jpeg", "jpeg"},
		{"photo.png", "png"},
		{"anim.gif", "gif"},
		{"unknown.txt", "unknown"},
		{"noext", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			fileType := storage.GetFileType(tt.filename)
			assert.Equal(t, tt.expectedType, fileType)
		})
	}
}

func TestGetFileSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		size        int64
		expectedStr string
	}{
		{"bytes", 500, "500B"},
		{"kilobytes", 1536, "1.50KB"},
		{"megabytes", 2097152, "2.00MB"},
		{"gigabytes", 2147483648, "2.00GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			header := &multipart.FileHeader{
				Size: tt.size,
			}

			sizeStr := storage.GetFileSize(header)
			assert.Equal(t, tt.expectedStr, sizeStr)
		})
	}
}

func TestDetectProjectCategory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename         string
		expectedCategory string
	}{
		{"website project", "website"},
		{"web app", "website"},
		{"mobile app", "mobile"},
		{"android project", "mobile"},
		{"ios development", "mobile"},
		{"iot sensor", "iot"},
		{"arduino controller", "iot"},
		{"machine learning", "machine_learning"},
		{"ml model", "machine_learning"},
		{"deep learning", "deep_learning"},
		{"neural network", "deep_learning"},
		{"unknown project", "website"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			category := storage.DetectProjectCategory(tt.filename)
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}

func TestValidateImageFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		filename    string
		size        int64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid png under limit",
			filename:    "photo.png",
			size:        1024 * 1024,
			expectError: false,
		},
		{
			name:        "valid jpg under limit",
			filename:    "image.jpg",
			size:        1024 * 1024,
			expectError: false,
		},
		{
			name:        "valid jpeg under limit",
			filename:    "pic.jpeg",
			size:        1024 * 1024,
			expectError: false,
		},
		{
			name:        "png exceeds 2MB limit",
			filename:    "large.png",
			size:        3 * 1024 * 1024,
			expectError: true,
			errorMsg:    "ukuran foto profil tidak boleh lebih dari 2MB",
		},
		{
			name:        "invalid format",
			filename:    "document.pdf",
			size:        1024,
			expectError: true,
			errorMsg:    "format foto profil harus png, jpg, atau jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			header := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     tt.size,
			}

			err := storage.ValidateImageFile(header)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveUploadedFile(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test.txt")
	// Create a multipart form with a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("test content"))
	require.NoError(t, err)
	writer.Close()
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	err = req.ParseMultipartForm(32 << 20)
	require.NoError(t, err)

	file, header, err := req.FormFile("file")
	require.NoError(t, err)
	defer file.Close()

	// Save the uploaded file
	err = storage.SaveUploadedFile(header, destPath)
	assert.NoError(t, err)

	// Verify file was saved
	savedContent, err := os.ReadFile(destPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test content"), savedContent)
}

func TestDeleteFile(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create a test file
	err := os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(testFile)
	assert.NoError(t, err)

	// Delete the file
	err = storage.DeleteFile(testFile)
	assert.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteFile_NonExistent(t *testing.T) {
	t.Parallel()
	// Deleting a non-existent file should not return an error
	err := storage.DeleteFile("/nonexistent/path/to/file.txt")
	assert.NoError(t, err)
}

func TestMoveFile(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Create source file
	content := []byte("test content for move")
	err := os.WriteFile(srcPath, content, 0o644)
	require.NoError(t, err)

	// Move file
	err = storage.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, dstContent)
}

func TestMoveFile_NonExistentSource(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "nonexistent.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Try to move non-existent file
	err := storage.MoveFile(srcPath, dstPath)
	assert.Error(t, err)

	// Verify destination was not created
	_, err = os.Stat(dstPath)
	assert.True(t, os.IsNotExist(err))
}

func TestMoveFile_SameSourceAndDestination(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "file.txt")
	dstPath := srcPath

	// Create source file
	content := []byte("test content")
	err := os.WriteFile(srcPath, content, 0o644)
	require.NoError(t, err)

	// Move file to itself should succeed (no-op)
	err = storage.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify file still exists and content is unchanged
	savedContent, err := os.ReadFile(srcPath)
	assert.NoError(t, err)
	assert.Equal(t, content, savedContent)
}

func TestMoveFile_DestinationDirectoryNotExists(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "subdir", "dest.txt")

	// Create source file
	content := []byte("test content for move")
	err := os.WriteFile(srcPath, content, 0o644)
	require.NoError(t, err)

	// Move file to non-existent directory - should fail or create parent dir
	err = storage.MoveFile(srcPath, dstPath)
	assert.Error(t, err)

	// Verify source still exists (move failed)
	_, err = os.Stat(srcPath)
	assert.NoError(t, err)
}

func TestMoveFile_OverwriteExistingDestination(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Create source file with content
	srcContent := []byte("source content")
	err := os.WriteFile(srcPath, srcContent, 0o644)
	require.NoError(t, err)

	// Create destination file with different content
	dstContent := []byte("destination content")
	err = os.WriteFile(dstPath, dstContent, 0o644)
	require.NoError(t, err)

	// Move file - should overwrite destination
	err = storage.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination now has source content
	resultContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, srcContent, resultContent)
}

func TestMoveFile_MoveToSubdirectory(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	subDir := filepath.Join(tempDir, "subdir")
	dstPath := filepath.Join(subDir, "dest.txt")

	// Create subdirectory
	err := os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	// Create source file
	content := []byte("test content")
	err = os.WriteFile(srcPath, content, 0o644)
	require.NoError(t, err)

	// Move file to subdirectory
	err = storage.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, dstContent)
}
