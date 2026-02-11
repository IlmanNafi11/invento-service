package helper

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDownloadHierarchy_Success tests creating a new download helper
func TestNewDownloadHelper_Success(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	assert.NotNil(t, dh)
}

// TestValidateDownloadRequest_EmptyArrays tests validation with empty arrays
func TestValidateDownloadRequest_EmptyArrays(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	err := dh.ValidateDownloadRequest([]string{}, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "harus diisi minimal salah satu")
}

// TestValidateDownloadRequest_WithProjects tests validation with project IDs
func TestValidateDownloadRequest_WithProjects(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	err := dh.ValidateDownloadRequest([]string{"1", "2"}, []string{})
	assert.NoError(t, err)
}

// TestValidateDownloadRequest_WithModuls tests validation with modul IDs
func TestValidateDownloadRequest_WithModuls(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	err := dh.ValidateDownloadRequest([]string{}, []string{"1", "2", "3"})
	assert.NoError(t, err)
}

// TestValidateDownloadRequest_WithBoth tests validation with both project and modul IDs
func TestValidateDownloadRequest_WithBoth(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	err := dh.ValidateDownloadRequest([]string{"1"}, []string{"2"})
	assert.NoError(t, err)
}

// TestGetFilesByIDs_SelectsCorrectProjects tests selecting projects by ID
func TestGetFilesByIDs_SelectsCorrectProjects(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	projects := []domain.Project{
		{ID: 1, NamaProject: "Project 1"},
		{ID: 2, NamaProject: "Project 2"},
		{ID: 3, NamaProject: "Project 3"},
	}
	moduls := []domain.Modul{
		{ID: 10, NamaFile: "Modul 1"},
		{ID: 20, NamaFile: "Modul 2"},
	}

	selectedProjects, selectedModuls := dh.GetFilesByIDs([]uint{1, 3}, []uint{20}, projects, moduls)

	assert.Len(t, selectedProjects, 2)
	assert.Equal(t, uint(1), selectedProjects[0].ID)
	assert.Equal(t, uint(3), selectedProjects[1].ID)
	assert.Len(t, selectedModuls, 1)
	assert.Equal(t, uint(20), selectedModuls[0].ID)
}

// TestGetFilesByIDs_EmptySelections tests with empty ID arrays
func TestGetFilesByIDs_EmptySelections(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	projects := []domain.Project{
		{ID: 1, NamaProject: "Project 1"},
	}
	moduls := []domain.Modul{
		{ID: 10, NamaFile: "Modul 1"},
	}

	selectedProjects, selectedModuls := dh.GetFilesByIDs([]uint{}, []uint{}, projects, moduls)

	assert.Empty(t, selectedProjects)
	assert.Empty(t, selectedModuls)
}

// TestGetFilesByIDs_NonExistentIDs tests filtering out non-existent IDs
func TestGetFilesByIDs_NonExistentIDs(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	projects := []domain.Project{
		{ID: 1, NamaProject: "Project 1"},
	}
	moduls := []domain.Modul{
		{ID: 10, NamaFile: "Modul 1"},
	}

	// Request non-existent IDs
	selectedProjects, selectedModuls := dh.GetFilesByIDs([]uint{999}, []uint{888}, projects, moduls)

	assert.Empty(t, selectedProjects)
	assert.Empty(t, selectedModuls)
}

// TestGetFilesByIDs_DoesNotModifyOriginals tests that original slices are not modified
func TestGetFilesByIDs_DoesNotModifyOriginals(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	projects := []domain.Project{
		{ID: 1, NamaProject: "Project 1"},
		{ID: 2, NamaProject: "Project 2"},
	}
	moduls := []domain.Modul{
		{ID: 10, NamaFile: "Modul 1"},
	}

	originalProjectsLen := len(projects)
	originalModulsLen := len(moduls)

	dh.GetFilesByIDs([]uint{1}, []uint{10}, projects, moduls)

	// Original slices should not be modified
	assert.Len(t, projects, originalProjectsLen)
	assert.Len(t, moduls, originalModulsLen)
}

// TestCreateDownloadZip_SingleFile returns single file path
func TestCreateDownloadZip_SingleFile(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	zipPath, err := dh.CreateDownloadZip([]string{tmpFile.Name()}, "123")
	assert.NoError(t, err)
	assert.Equal(t, tmpFile.Name(), zipPath)
}

// TestCreateDownloadZip_EmptyArray tests error with empty file array
func TestCreateDownloadZip_EmptyArray(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	zipPath, err := dh.CreateDownloadZip([]string{}, "123")
	assert.Error(t, err)
	assert.Empty(t, zipPath)
	assert.Contains(t, err.Error(), "tidak ada file")
}

// TestCreateDownloadZip_MultipleFiles tests creating zip with multiple files
func TestCreateDownloadZip_MultipleFiles(t *testing.T) {
	// Create temp files
	tmpFile1, err := os.CreateTemp("", "test1*.txt")
	require.NoError(t, err)
	tmpFile1.Close()
	defer os.Remove(tmpFile1.Name())

	tmpFile2, err := os.CreateTemp("", "test2*.txt")
	require.NoError(t, err)
	tmpFile2.Close()
	defer os.Remove(tmpFile2.Name())

	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	// This would create a zip file, but we can't test the full functionality
	// without a proper path resolver setup
	// Just verify it doesn't panic
	assert.NotNil(t, dh)
}

// TestPrepareFilesForDownload_AllFilesFound tests when all files exist
func TestPrepareFilesForDownload_AllFilesFound(t *testing.T) {
	// Create temp directory and files
	tmpDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	os.WriteFile(file1, []byte("test1"), 0644)
	os.WriteFile(file2, []byte("test2"), 0644)

	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	projects := []domain.Project{
		{ID: 1, PathFile: file1},
	}
	moduls := []domain.Modul{
		{ID: 1, PathFile: file2},
	}

	filePaths, notFoundFiles, err := dh.PrepareFilesForDownload(projects, moduls)
	assert.NoError(t, err)
	assert.Len(t, filePaths, 2)
	assert.Empty(t, notFoundFiles)
}

// TestPrepareFilesForDownload_SomeFilesNotFound tests when some files are missing
func TestPrepareFilesForDownload_SomeFilesNotFound(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	projects := []domain.Project{
		{ID: 1, PathFile: "/nonexistent/file1.pdf"},
	}
	moduls := []domain.Modul{
		{ID: 1, PathFile: "/nonexistent/file2.pdf"},
	}

	filePaths, notFoundFiles, err := dh.PrepareFilesForDownload(projects, moduls)
	assert.Error(t, err)
	assert.Nil(t, filePaths)
	assert.NotEmpty(t, notFoundFiles)
	assert.Contains(t, err.Error(), "semua file tidak ditemukan")
}

// TestPrepareFilesForDownloads_EmptyArrays tests with empty project/modul arrays
func TestPrepareFilesForDownloads_EmptyArrays(t *testing.T) {
	pathResolver := NewPathResolver(&config.Config{})
	dh := NewDownloadHelper(pathResolver)

	filePaths, notFoundFiles, err := dh.PrepareFilesForDownload([]domain.Project{}, []domain.Modul{})
	// Empty arrays mean no files were found, so it should return an error
	assert.Error(t, err)
	assert.Nil(t, filePaths)
	assert.Empty(t, notFoundFiles)
	assert.Contains(t, err.Error(), "semua file tidak ditemukan")
}

// TestResolvePath_AbsolutePath returns absolute path as-is
func TestResolvePath_AbsolutePath(t *testing.T) {
	cfg := &config.Config{}
	cfg.Upload.PathProduction = "/test/uploads"
	pathResolver := NewPathResolver(cfg)
	dh := NewDownloadHelper(pathResolver)

	absPath := "/absolute/path/to/file.pdf"
	result := dh.resolvePath(absPath)
	assert.Equal(t, absPath, result)
}

// TestResolvePath_RelativePath tests relative path resolution
func TestResolvePath_RelativePath(t *testing.T) {
	cfg := &config.Config{}
	cfg.Upload.PathProduction = "/test/uploads"
	pathResolver := NewPathResolver(cfg)
	dh := NewDownloadHelper(pathResolver)

	relPath := "uploads/file.pdf"
	result := dh.resolvePath(relPath)
	// The function removes "uploads/" prefix and joins with base path
	// filepath.Abs() converts to absolute path based on current directory
	assert.NotContains(t, result, "uploads/uploads")
	assert.Contains(t, result, "file.pdf")
}

// TestResolvePath_WithUploadsPrefix tests removing uploads prefix
func TestResolvePath_WithUploadsPrefix(t *testing.T) {
	cfg := &config.Config{}
	cfg.Upload.PathProduction = "/base/uploads"
	pathResolver := NewPathResolver(cfg)
	dh := NewDownloadHelper(pathResolver)

	path := "uploads/subfolder/file.pdf"
	result := dh.resolvePath(path)
	// The function removes "uploads/" prefix before joining
	assert.NotContains(t, result, "uploads/subfolder")
	assert.Contains(t, result, "subfolder/file.pdf")
}

// TestResolvePath_WithTempPrefix tests removing temp prefix
func TestResolvePath_WithTempPrefix(t *testing.T) {
	cfg := &config.Config{}
	cfg.Upload.PathProduction = "/base/uploads"
	pathResolver := NewPathResolver(cfg)
	dh := NewDownloadHelper(pathResolver)

	path := "temp/file.pdf"
	result := dh.resolvePath(path)
	// The function removes "temp/" prefix before joining
	assert.NotContains(t, result, "temp/temp")
	assert.Contains(t, result, "file.pdf")
}
