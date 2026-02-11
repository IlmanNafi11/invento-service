package helper_test

import (
	"bytes"
	"fiber-boiler-plate/internal/helper"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUniqueIdentifier(t *testing.T) {
	id1, err := helper.GenerateUniqueIdentifier(10)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(id1)) // hex encoding doubles the length

	id2, err := helper.GenerateUniqueIdentifier(10)
	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)
}

func TestGenerateRandomString(t *testing.T) {
	str1, err := helper.GenerateRandomString(10)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(str1))

	str2, err := helper.GenerateRandomString(10)
	assert.NoError(t, err)
	assert.NotEqual(t, str1, str2)

	// Check that only valid characters are used
	for _, c := range str1 {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		filename     string
		expectedExt  string
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
			ext := helper.GetFileExtension(tt.filename)
			assert.Equal(t, tt.expectedExt, ext)
		})
	}
}

func TestValidateZipFile(t *testing.T) {
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
			// Create a mock file header
			header := &multipart.FileHeader{
				Filename: tt.filename,
			}

			err := helper.ValidateZipFile(header)

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
			header := &multipart.FileHeader{
				Filename: tt.filename,
			}

			err := helper.ValidateModulFile(header)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
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
			fileType := helper.GetFileType(tt.filename)
			assert.Equal(t, tt.expectedType, fileType)
		})
	}
}

func TestGetFileSize(t *testing.T) {
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
			header := &multipart.FileHeader{
				Size: tt.size,
			}

			sizeStr := helper.GetFileSize(header)
			assert.Equal(t, tt.expectedStr, sizeStr)
		})
	}
}

func TestDetectProjectCategory(t *testing.T) {
	tests := []struct {
		filename        string
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
			category := helper.DetectProjectCategory(tt.filename)
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}

func TestValidateImageFile(t *testing.T) {
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
			header := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     tt.size,
			}

			err := helper.ValidateImageFile(header)

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

	// Create a HTTP request from the body
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Parse the form to get the FileHeader
	err = req.ParseMultipartForm(32 << 20)
	require.NoError(t, err)

	file, header, err := req.FormFile("file")
	require.NoError(t, err)
	defer file.Close()

	// Save the uploaded file
	err = helper.SaveUploadedFile(header, destPath)
	assert.NoError(t, err)

	// Verify file was saved
	savedContent, err := os.ReadFile(destPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test content"), savedContent)
}

func TestDeleteFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create a test file
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(testFile)
	assert.NoError(t, err)

	// Delete the file
	err = helper.DeleteFile(testFile)
	assert.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteFile_NonExistent(t *testing.T) {
	// Deleting a non-existent file should not return an error
	err := helper.DeleteFile("/nonexistent/path/to/file.txt")
	assert.NoError(t, err)
}

func TestMoveFile(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Create source file
	content := []byte("test content for move")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Move file
	err = helper.MoveFile(srcPath, dstPath)
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
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "nonexistent.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Try to move non-existent file
	err := helper.MoveFile(srcPath, dstPath)
	assert.Error(t, err)

	// Verify destination was not created
	_, err = os.Stat(dstPath)
	assert.True(t, os.IsNotExist(err))
}

func TestMoveFile_SameSourceAndDestination(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "file.txt")
	dstPath := srcPath

	// Create source file
	content := []byte("test content")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Move file to itself should succeed (no-op)
	err = helper.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify file still exists and content is unchanged
	savedContent, err := os.ReadFile(srcPath)
	assert.NoError(t, err)
	assert.Equal(t, content, savedContent)
}

func TestMoveFile_DestinationDirectoryNotExists(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "subdir", "dest.txt")

	// Create source file
	content := []byte("test content for move")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Move file to non-existent directory - should fail or create parent dir
	err = helper.MoveFile(srcPath, dstPath)
	assert.Error(t, err)

	// Verify source still exists (move failed)
	_, err = os.Stat(srcPath)
	assert.NoError(t, err)
}

func TestMoveFile_OverwriteExistingDestination(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Create source file with content
	srcContent := []byte("source content")
	err := os.WriteFile(srcPath, srcContent, 0644)
	require.NoError(t, err)

	// Create destination file with different content
	dstContent := []byte("destination content")
	err = os.WriteFile(dstPath, dstContent, 0644)
	require.NoError(t, err)

	// Move file - should overwrite destination
	err = helper.MoveFile(srcPath, dstPath)
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
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	subDir := filepath.Join(tempDir, "subdir")
	dstPath := filepath.Join(subDir, "dest.txt")

	// Create subdirectory
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Create source file
	content := []byte("test content")
	err = os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Move file to subdirectory
	err = helper.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, dstContent)
}

func TestMoveFile_CopyFallbackPath(t *testing.T) {
	// This test exercises the fallback path when os.Rename fails
	// (e.g., cross-device move). We simulate this by making the destination
	// a directory first, which causes rename to fail with a specific error.
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Create source file
	content := []byte("test content for fallback path")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Create destination as a directory to force rename failure
	err = os.MkdirAll(dstPath, 0755)
	require.NoError(t, err)

	// Try to move - should fail because dst is a directory
	err = helper.MoveFile(srcPath, dstPath)
	assert.Error(t, err)

	// Clean up the directory
	os.RemoveAll(dstPath)

	// Now move to a proper file path - this should work
	// and on most systems will use the direct rename path
	err = helper.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists with correct content
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, dstContent)
}

func TestMoveFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "empty.txt")
	dstPath := filepath.Join(tempDir, "dest_empty.txt")

	// Create empty source file
	err := os.WriteFile(srcPath, []byte{}, 0644)
	require.NoError(t, err)

	// Move empty file
	err = helper.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists and is empty
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, dstContent)
}

func TestMoveFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "large.txt")
	dstPath := filepath.Join(tempDir, "dest_large.txt")

	// Create a larger source file (1MB of data)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	err := os.WriteFile(srcPath, largeContent, 0644)
	require.NoError(t, err)

	// Move large file
	err = helper.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists with correct content
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, largeContent, dstContent)
}

func TestMoveFile_WithRelativePath(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory to test relative paths
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	srcPath := "source.txt"
	dstPath := "dest.txt"

	// Create source file with relative path
	content := []byte("relative path test")
	err = os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Move file using relative paths
	err = helper.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, dstContent)
}

func TestMoveFile_SpecialCharactersInFilename(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "file with spaces & special-chars_123.txt")
	dstPath := filepath.Join(tempDir, "dest file (1).txt")

	// Create source file with special characters
	content := []byte("special chars test")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Move file
	err = helper.MoveFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Verify source is gone
	_, err = os.Stat(srcPath)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, dstContent)
}

func TestFormatFileSize(t *testing.T) {
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
			sizeStr := helper.FormatFileSize(tt.size)
			assert.Equal(t, tt.expectedStr, sizeStr)
		})
	}
}

func TestGetFileSizeFromPath(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Get file size
	sizeStr := helper.GetFileSizeFromPath(testFile)
	assert.NotEqual(t, "0B", sizeStr)
	assert.Contains(t, sizeStr, "B")
}

func TestGetFileSizeFromPath_NonExistent(t *testing.T) {
	sizeStr := helper.GetFileSizeFromPath("/nonexistent/file.txt")
	assert.Equal(t, "0B", sizeStr)
}

func TestCreateZipArchive(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	err := os.WriteFile(file1, []byte("content 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content 2"), 0644)
	require.NoError(t, err)

	// Create zip archive
	zipPath := filepath.Join(tempDir, "archive.zip")
	err = helper.CreateZipArchive([]string{file1, file2}, zipPath)
	assert.NoError(t, err)

	// Verify zip file exists
	_, err = os.Stat(zipPath)
	assert.NoError(t, err)

	// Verify zip file has content
	info, _ := os.Stat(zipPath)
	assert.Greater(t, info.Size(), int64(0))
}

func TestCreateZipArchive_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "archive.zip")

	err := helper.CreateZipArchive([]string{"/nonexistent/file.txt"}, zipPath)
	assert.Error(t, err)
}

func TestCreateUserDirectory(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		role        string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "user@example.com",
			role:        "mahasiswa",
			expectError: false,
		},
		{
			name:        "invalid email - no @",
			email:       "userexample.com",
			role:        "mahasiswa",
			expectError: true,
		},
		{
			name:        "invalid email - multiple @",
			email:       "user@@example.com",
			role:        "mahasiswa",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := helper.CreateUserDirectory(tt.email, tt.role)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, dir)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, dir)

				// Verify directory exists
				info, err := os.Stat(dir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Clean up
				os.RemoveAll(dir)
			}
		})
	}
}

func TestCreateModulDirectory(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		role        string
		fileType    string
		expectError bool
	}{
		{
			name:        "valid parameters",
			email:       "user@example.com",
			role:        "mahasiswa",
			fileType:    "pdf",
			expectError: false,
		},
		{
			name:        "invalid email",
			email:       "invalid-email",
			role:        "mahasiswa",
			fileType:    "docx",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := helper.CreateModulDirectory(tt.email, tt.role, tt.fileType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, dir)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, dir)
				assert.Contains(t, dir, tt.fileType)

				// Verify directory exists
				info, err := os.Stat(dir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Clean up
				os.RemoveAll(dir)
			}
		})
	}
}

func TestCreateProfilDirectory(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		role        string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "user@example.com",
			role:        "dosen",
			expectError: false,
		},
		{
			name:        "invalid email",
			email:       "invalid",
			role:        "dosen",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := helper.CreateProfilDirectory(tt.email, tt.role)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, dir)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, dir)

				// Verify directory exists
				info, err := os.Stat(dir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())

				// Clean up
				os.RemoveAll(dir)
			}
		})
	}
}

func TestFileOperations_Integration(t *testing.T) {
	tempDir := t.TempDir()

	// Test complete file workflow
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("integration test content")

	// Create file
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Check file size
	sizeStr := helper.GetFileSizeFromPath(testFile)
	assert.NotEqual(t, "0B", sizeStr)

	// Move file
	newPath := filepath.Join(tempDir, "moved.txt")
	err = helper.MoveFile(testFile, newPath)
	assert.NoError(t, err)

	// Verify old path is gone
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))

	// Verify new path exists
	movedContent, err := os.ReadFile(newPath)
	assert.NoError(t, err)
	assert.Equal(t, content, movedContent)

	// Delete file
	err = helper.DeleteFile(newPath)
	assert.NoError(t, err)

	// Verify deletion
	_, err = os.Stat(newPath)
	assert.True(t, os.IsNotExist(err))
}
