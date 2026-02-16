package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"invento-service/config"
	"invento-service/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFileManagerTest(t *testing.T) (fm *storage.FileManager, tmpDir string) {
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
	return storage.NewFileManager(cfg), tempDir
}

func TestNewFileManager(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
		},
	}

	fm := storage.NewFileManager(cfg)

	assert.NotNil(t, fm)
}

func TestFileManager_GenerateRandomDirectory(t *testing.T) {
	t.Parallel()
	fm, _ := setupFileManagerTest(t)

	dir1, err := fm.GenerateRandomDirectory()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(dir1))

	dir2, err := fm.GenerateRandomDirectory()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(dir2))

	// Should be different (very high probability)
	assert.NotEqual(t, dir1, dir2)
}

func TestFileManager_GetUserUploadPath(t *testing.T) {
	t.Parallel()
	fm, tempDir := setupFileManagerTest(t)
	userID := "123"

	path, err := fm.GetUserUploadPath(userID)
	assert.NoError(t, err)

	expectedPath := filepath.Join(tempDir, "projects", "123")
	assert.Equal(t, expectedPath, path)

	// Verify directory was created
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFileManager_GetUserUploadPath_Production(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "production",
		},
		Upload: config.UploadConfig{
			PathProduction: tempDir,
		},
	}
	fm := storage.NewFileManager(cfg)
	userID := "456"

	path, err := fm.GetUserUploadPath(userID)
	assert.NoError(t, err)

	expectedPath := filepath.Join(tempDir, "projects", "456")
	assert.Equal(t, expectedPath, path)

	// Verify directory was created
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
	_ = info // use the variable
}

func TestFileManager_CreateProjectUploadDirectory(t *testing.T) {
	t.Parallel()
	fm, tempDir := setupFileManagerTest(t)
	userID := "789"

	projectPath, randomDir, err := fm.CreateProjectUploadDirectory(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, randomDir)
	assert.Equal(t, 10, len(randomDir))

	// Verify path format
	expectedPrefix := filepath.Join(tempDir, "projects", "789")
	assert.Contains(t, projectPath, expectedPrefix)

	// Verify directory was created
	info, err := os.Stat(projectPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFileManager_GetProjectFilePath(t *testing.T) {
	t.Parallel()
	fm, tempDir := setupFileManagerTest(t)
	userID := "100"
	randomDir := "abc123"
	filename := "testfile.zip"

	path := fm.GetProjectFilePath(userID, randomDir, filename)
	expectedPath := filepath.Join(tempDir, "projects", "100", "abc123", "testfile.zip")

	assert.Equal(t, expectedPath, path)
}

func TestFileManager_DeleteUserDirectory(t *testing.T) {
	t.Parallel()
	fm, _ := setupFileManagerTest(t)
	userID := "200"

	// Create directory first
	userPath, err := fm.GetUserUploadPath(userID)
	require.NoError(t, err)

	// Verify directory exists
	_, err = os.Stat(userPath)
	assert.NoError(t, err)

	// Delete directory
	err = fm.DeleteUserDirectory(userID)
	assert.NoError(t, err)

	// Verify directory is gone
	_, err = os.Stat(userPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFileManager_DeleteProjectDirectory(t *testing.T) {
	t.Parallel()
	fm, _ := setupFileManagerTest(t)
	userID := "300"

	// Create project directory first
	projectPath, randomDir, err := fm.CreateProjectUploadDirectory(userID)
	require.NoError(t, err)

	// Create a file in it
	testFile := filepath.Join(projectPath, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(testFile)
	assert.NoError(t, err)

	// Delete project directory
	err = fm.DeleteProjectDirectory(userID, randomDir)
	assert.NoError(t, err)

	// Verify directory is gone
	_, err = os.Stat(projectPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFileManager_GetUploadFilePath(t *testing.T) {
	t.Parallel()
	fm, _ := setupFileManagerTest(t)
	uploadID := "upload-123"

	path := fm.GetUploadFilePath(uploadID)

	// The path should contain the upload ID
	assert.Contains(t, path, "uploads")
	assert.Contains(t, path, uploadID)
}

func TestFileManager_GetUploadFilePath_Production(t *testing.T) {
	t.Parallel()
	_ = t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "production",
		},
		Upload: config.UploadConfig{
			TempPathProduction: "/tmp/prod/temp",
		},
	}
	fm := storage.NewFileManager(cfg)
	uploadID := "upload-prod-456"

	path := fm.GetUploadFilePath(uploadID)
	expectedPath := "/tmp/prod/temp/uploads/upload-prod-456"

	assert.Equal(t, expectedPath, path)
}

func TestFileManager_GetModulBasePath(t *testing.T) {
	t.Parallel()
	fm, tempDir := setupFileManagerTest(t)

	path := fm.GetModulBasePath()
	assert.Equal(t, tempDir, path)
}

func TestFileManager_GetModulBasePath_Production(t *testing.T) {
	t.Parallel()
	_ = t.TempDir() // needed for test isolation
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "production",
		},
		Upload: config.UploadConfig{
			PathProduction: "/tmp/prod",
		},
	}
	fm := storage.NewFileManager(cfg)

	path := fm.GetModulBasePath()
	assert.Equal(t, "/tmp/prod", path)
}

func TestFileManager_GetUserModulPath(t *testing.T) {
	t.Parallel()
	fm, tempDir := setupFileManagerTest(t)
	userID := "500"

	path, err := fm.GetUserModulPath(userID)
	assert.NoError(t, err)

	expectedPath := filepath.Join(tempDir, "moduls", "500")
	assert.Equal(t, expectedPath, path)

	// Verify directory was created
	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFileManager_CreateModulUploadDirectory(t *testing.T) {
	t.Parallel()
	fm, tempDir := setupFileManagerTest(t)
	userID := "600"

	modulPath, randomDir, err := fm.CreateModulUploadDirectory(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, randomDir)
	assert.Equal(t, 10, len(randomDir))

	// Verify path format
	expectedPrefix := filepath.Join(tempDir, "moduls", "600")
	assert.Contains(t, modulPath, expectedPrefix)

	// Verify directory was created
	info, err := os.Stat(modulPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFileManager_GetModulFilePath(t *testing.T) {
	t.Parallel()
	fm, tempDir := setupFileManagerTest(t)
	userID := "700"
	randomDir := "xyz789"
	filename := "modul.pdf"

	path := fm.GetModulFilePath(userID, randomDir, filename)
	expectedPath := filepath.Join(tempDir, "moduls", "700", "xyz789", "modul.pdf")

	assert.Equal(t, expectedPath, path)
}

func TestFileManager_DeleteModulDirectory(t *testing.T) {
	t.Parallel()
	fm, _ := setupFileManagerTest(t)
	userID := "800"

	// Create modul directory first
	modulPath, randomDir, err := fm.CreateModulUploadDirectory(userID)
	require.NoError(t, err)

	// Create a file in it
	testFile := filepath.Join(modulPath, "modul.pdf")
	err = os.WriteFile(testFile, []byte("modul content"), 0o644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(testFile)
	assert.NoError(t, err)

	// Delete modul directory
	err = fm.DeleteModulDirectory(userID, randomDir)
	assert.NoError(t, err)

	// Verify directory is gone
	_, err = os.Stat(modulPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFileManager_DeleteUserModulDirectory(t *testing.T) {
	t.Parallel()
	fm, _ := setupFileManagerTest(t)
	userID := "900"

	// Create user modul directory first
	userModulPath, err := fm.GetUserModulPath(userID)
	require.NoError(t, err)

	// Create some subdirectories
	_, randomDir1, err := fm.CreateModulUploadDirectory(userID)
	require.NoError(t, err)
	_, randomDir2, err := fm.CreateModulUploadDirectory(userID)
	require.NoError(t, err)

	// Verify directories exist
	_, err = os.Stat(filepath.Join(userModulPath, randomDir1))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(userModulPath, randomDir2))
	assert.NoError(t, err)

	// Delete entire user modul directory
	err = fm.DeleteUserModulDirectory(userID)
	assert.NoError(t, err)

	// Verify directory is gone
	_, err = os.Stat(userModulPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFileManager_MultipleUsers(t *testing.T) {
	t.Parallel()
	fm, _ := setupFileManagerTest(t)

	// Create directories for multiple users
	userIDs := []string{"1", "2", "3", "4", "5"}
	paths := make(map[string]string)

	for _, uid := range userIDs {
		path, err := fm.GetUserUploadPath(uid)
		assert.NoError(t, err)
		paths[uid] = path

		// Verify each user's directory exists
		info, err := os.Stat(path)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	}

	// Verify all paths are unique
	uniquePaths := make(map[string]bool)
	for _, path := range paths {
		uniquePaths[path] = true
	}
	assert.Equal(t, len(userIDs), len(uniquePaths))
}
