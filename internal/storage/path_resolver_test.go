package storage_test

import (
	"invento-service/config"
	"invento-service/internal/storage"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPathResolverTest(t *testing.T) (*storage.PathResolver, string) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			PathDevelopment:     tempDir,
			PathProduction:      filepath.Join(tempDir, "prod"),
			TempPathDevelopment: filepath.Join(tempDir, "temp-dev"),
			TempPathProduction:  filepath.Join(tempDir, "temp-prod"),
		},
	}
	return storage.NewPathResolver(cfg), tempDir
}

func TestNewPathResolver(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			PathDevelopment:     tempDir,
			PathProduction:      filepath.Join(tempDir, "prod"),
			TempPathDevelopment: filepath.Join(tempDir, "temp"),
		},
	}

	pr := storage.NewPathResolver(cfg)

	assert.NotNil(t, pr)
}

func TestPathResolver_GetBasePath_Development(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)

	basePath := pr.GetBasePath()
	assert.Equal(t, tempDir, basePath)
}

func TestPathResolver_GetBasePath_Production(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "production",
		},
		Upload: config.UploadConfig{
			PathProduction: tempDir,
		},
	}
	pr := storage.NewPathResolver(cfg)

	basePath := pr.GetBasePath()
	assert.Equal(t, tempDir, basePath)
}

func TestPathResolver_GetTempPath_Development(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)

	tempPath := pr.GetTempPath()
	assert.Equal(t, filepath.Join(tempDir, "temp-dev"), tempPath)
}

func TestPathResolver_GetTempPath_Production(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "production",
		},
		Upload: config.UploadConfig{
			TempPathProduction: filepath.Join(tempDir, "temp-prod"),
		},
	}
	pr := storage.NewPathResolver(cfg)

	tempPath := pr.GetTempPath()
	assert.Equal(t, filepath.Join(tempDir, "temp-prod"), tempPath)
}

func TestPathResolver_GetProjectPath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := "123"

	projectPath := pr.GetProjectPath(userID)
	expectedPath := filepath.Join(tempDir, "projects", "123")

	assert.Equal(t, expectedPath, projectPath)
}

func TestPathResolver_GetProjectFilePath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := "456"
	identifier := "abc123"
	filename := "project.zip"

	filePath := pr.GetProjectFilePath(userID, identifier, filename)
	expectedPath := filepath.Join(tempDir, "projects", "456", "abc123", "project.zip")

	assert.Equal(t, expectedPath, filePath)
}

func TestPathResolver_GetProjectDirectory(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := "789"
	identifier := "xyz789"

	dirPath := pr.GetProjectDirectory(userID, identifier)
	expectedPath := filepath.Join(tempDir, "projects", "789", "xyz789")

	assert.Equal(t, expectedPath, dirPath)
}

func TestPathResolver_GetUploadPath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	uploadID := "upload-123"

	uploadPath := pr.GetUploadPath(uploadID)
	expectedPath := filepath.Join(tempDir, "temp-dev", "uploads", "upload-123")

	assert.Equal(t, expectedPath, uploadPath)
}

func TestPathResolver_GetUploadFilePath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	uploadID := "upload-456"

	filePath := pr.GetUploadFilePath(uploadID)
	expectedPath := filepath.Join(tempDir, "temp-dev", "uploads", "upload-456", "file.zip")

	assert.Equal(t, expectedPath, filePath)
}

func TestPathResolver_GetUploadInfoPath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	uploadID := "upload-789"

	infoPath := pr.GetUploadInfoPath(uploadID)
	expectedPath := filepath.Join(tempDir, "temp-dev", "uploads", "upload-789", "info.json")

	assert.Equal(t, expectedPath, infoPath)
}

func TestPathResolver_EnsureDirectoryExists(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)

	// Create a new directory
	testDir := filepath.Join(tempDir, "test", "nested", "dir")
	err := pr.EnsureDirectoryExists(testDir)
	assert.NoError(t, err)

	// Verify directory exists
	info, err := os.Stat(testDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestPathResolver_DirectoryExists(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)

	// Test non-existent directory
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	assert.False(t, pr.DirectoryExists(nonExistentDir))

	// Create directory
	existingDir := filepath.Join(tempDir, "existing")
	err := os.MkdirAll(existingDir, 0755)
	require.NoError(t, err)

	// Test existing directory
	assert.True(t, pr.DirectoryExists(existingDir))
}

func TestPathResolver_FileExists(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)

	// Test non-existent file
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	assert.False(t, pr.FileExists(nonExistentFile))

	// Create file
	existingFile := filepath.Join(tempDir, "existing.txt")
	err := os.WriteFile(existingFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Test existing file
	assert.True(t, pr.FileExists(existingFile))

	// Directory should return false for FileExists
	assert.False(t, pr.FileExists(tempDir))
}

func TestPathResolver_GetProfilPath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := "100"

	profilPath := pr.GetProfilPath(userID)
	expectedPath := filepath.Join(tempDir, "profil", "100")

	assert.Equal(t, expectedPath, profilPath)
}

func TestPathResolver_GetProfilFilePath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := "200"
	filename := "avatar.jpg"

	filePath := pr.GetProfilFilePath(userID, filename)
	expectedPath := filepath.Join(tempDir, "profil", "200", "avatar.jpg")

	assert.Equal(t, expectedPath, filePath)
}

func TestPathResolver_GetProfilDirectory(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := "300"

	dirPath := pr.GetProfilDirectory(userID)
	expectedPath := filepath.Join(tempDir, "profil", "300")

	assert.Equal(t, expectedPath, dirPath)
}

func TestPathResolver_GetModulPath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := uint(400)
	identifier := "mod123"

	modulPath := pr.GetModulPath(userID, identifier)
	expectedPath := filepath.Join(tempDir, "moduls", "400", "mod123")

	assert.Equal(t, expectedPath, modulPath)
}

func TestPathResolver_GetModulFilePath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)
	userID := uint(500)
	identifier := "mod456"
	filename := "lesson.pdf"

	filePath := pr.GetModulFilePath(userID, identifier, filename)
	expectedPath := filepath.Join(tempDir, "moduls", "500", "mod456", "lesson.pdf")

	assert.Equal(t, expectedPath, filePath)
}

func TestPathResolver_ConvertToAPIPath(t *testing.T) {
	pr, tempDir := setupPathResolverTest(t)

	tests := []struct {
		name         string
		absolutePath string
		expectedPath string
	}{
		{
			name:         "valid file in base path",
			absolutePath: filepath.Join(tempDir, "projects", "123", "file.zip"),
			expectedPath: "/uploads/projects/123/file.zip",
		},
		{
			name:         "nested path",
			absolutePath: filepath.Join(tempDir, "projects", "123", "dir", "file.zip"),
			expectedPath: "/uploads/projects/123/dir/file.zip",
		},
		{
			name:         "profil file",
			absolutePath: filepath.Join(tempDir, "profil", "456", "avatar.jpg"),
			expectedPath: "/uploads/profil/456/avatar.jpg",
		},
		{
			name:         "empty path",
			absolutePath: "",
			expectedPath: "",
		},
		{
			name:         "nil pointer",
			absolutePath: "",
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil pointer" {
				result := pr.ConvertToAPIPath(nil)
				assert.Nil(t, result)
			} else {
				result := pr.ConvertToAPIPath(&tt.absolutePath)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedPath, *result)
			}
		})
	}
}

func TestPathResolver_ConvertToAPIPath_OutsideBase(t *testing.T) {
	pr, _ := setupPathResolverTest(t)

	// Path outside base path should be returned unchanged
	outsidePath := "/etc/passwd"
	result := pr.ConvertToAPIPath(&outsidePath)

	assert.NotNil(t, result)
	assert.Equal(t, outsidePath, *result)
}

func TestPathResolver_ConvertToAPIPath_RelativePath(t *testing.T) {
	pr, _ := setupPathResolverTest(t)

	// Relative path should be handled correctly
	relativePath := filepath.Join(pr.GetBasePath(), "test.txt")
	apiPath := pr.ConvertToAPIPath(&relativePath)

	assert.NotNil(t, apiPath)
	assert.Contains(t, *apiPath, "/uploads/")
	assert.Contains(t, *apiPath, "test.txt")
}
