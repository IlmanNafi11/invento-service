package storage_test

import (
	"invento-service/internal/storage"
	"os"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveFile_CopyFallbackPath(t *testing.T) {
	t.Parallel()
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
	err = storage.MoveFile(srcPath, dstPath)
	assert.Error(t, err)

	// Clean up the directory
	os.RemoveAll(dstPath)

	// Now move to a proper file path - this should work
	// and on most systems will use the direct rename path
	err = storage.MoveFile(srcPath, dstPath)
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
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "empty.txt")
	dstPath := filepath.Join(tempDir, "dest_empty.txt")

	// Create empty source file
	err := os.WriteFile(srcPath, []byte{}, 0644)
	require.NoError(t, err)

	// Move empty file
	err = storage.MoveFile(srcPath, dstPath)
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
	t.Parallel()
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
	err = storage.MoveFile(srcPath, dstPath)
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
	t.Parallel()
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

func TestMoveFile_SpecialCharactersInFilename(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "file with spaces & special-chars_123.txt")
	dstPath := filepath.Join(tempDir, "dest file (1).txt")

	// Create source file with special characters
	content := []byte("special chars test")
	err := os.WriteFile(srcPath, content, 0644)
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

func TestFormatFileSize(t *testing.T) {
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
			sizeStr := storage.FormatFileSize(tt.size)
			assert.Equal(t, tt.expectedStr, sizeStr)
		})
	}
}

func TestGetFileSizeFromPath(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Get file size
	sizeStr := storage.GetFileSizeFromPath(testFile)
	assert.NotEqual(t, "0B", sizeStr)
	assert.Contains(t, sizeStr, "B")
}

func TestGetFileSizeFromPath_NonExistent(t *testing.T) {
	t.Parallel()
	sizeStr := storage.GetFileSizeFromPath("/nonexistent/file.txt")
	assert.Equal(t, "0B", sizeStr)
}

func TestCreateZipArchive(t *testing.T) {
	t.Parallel()
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
	err = storage.CreateZipArchive([]string{file1, file2}, zipPath)
	assert.NoError(t, err)

	// Verify zip file exists
	_, err = os.Stat(zipPath)
	assert.NoError(t, err)

	// Verify zip file has content
	info, _ := os.Stat(zipPath)
	assert.Greater(t, info.Size(), int64(0))
}

func TestCreateZipArchive_NonExistentFile(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "archive.zip")

	err := storage.CreateZipArchive([]string{"/nonexistent/file.txt"}, zipPath)
	assert.Error(t, err)
}

func TestCreateUserDirectory(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			dir, err := storage.CreateUserDirectory(tt.email, tt.role)

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
	t.Parallel()
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
			t.Parallel()
			dir, err := storage.CreateModulDirectory(tt.email, tt.role, tt.fileType)

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
	t.Parallel()
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
			t.Parallel()
			dir, err := storage.CreateProfilDirectory(tt.email, tt.role)

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
	t.Parallel()
	tempDir := t.TempDir()

	// Test complete file workflow
	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("integration test content")

	// Create file
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(t, err)

	// Check file size
	sizeStr := storage.GetFileSizeFromPath(testFile)
	assert.NotEqual(t, "0B", sizeStr)

	// Move file
	newPath := filepath.Join(tempDir, "moved.txt")
	err = storage.MoveFile(testFile, newPath)
	assert.NoError(t, err)

	// Verify old path is gone
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))

	// Verify new path exists
	movedContent, err := os.ReadFile(newPath)
	assert.NoError(t, err)
	assert.Equal(t, content, movedContent)

	// Delete file
	err = storage.DeleteFile(newPath)
	assert.NoError(t, err)

	// Verify deletion
	_, err = os.Stat(newPath)
	assert.True(t, os.IsNotExist(err))
}
