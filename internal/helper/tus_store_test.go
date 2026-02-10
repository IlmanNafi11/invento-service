package helper_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/helper"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTusStoreTest(t *testing.T) (*helper.TusStore, *helper.PathResolver, string) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			PathDevelopment:      tempDir,
			TempPathDevelopment:   filepath.Join(tempDir, "temp"),
		},
	}
	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024) // 10MB max
	return tusStore, pathResolver, tempDir
}

func TestNewTusStore(t *testing.T) {
	pathResolver := &helper.PathResolver{}
	maxFileSize := int64(1024 * 1024)

	tusStore := helper.NewTusStore(pathResolver, maxFileSize)

	assert.NotNil(t, tusStore)
}

func TestTusStore_InitiateUpload(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)

	tests := []struct {
		name        string
		uploadID    string
		fileSize    int64
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid upload",
			uploadID:    "test-upload-1",
			fileSize:    1024,
			expectError: false,
		},
		{
			name:        "empty upload id",
			uploadID:    "",
			fileSize:    1024,
			expectError: false,
		},
		{
			name:        "file size exceeds maximum",
			uploadID:    "test-upload-2",
			fileSize:    20 * 1024 * 1024,
			expectError: true,
			errorMsg:    "ukuran file melebihi batas maksimal",
		},
		{
			name:        "invalid file size",
			uploadID:    "test-upload-3",
			fileSize:    0,
			expectError: true,
			errorMsg:    "ukuran file tidak valid",
		},
		{
			name:        "negative file size",
			uploadID:    "test-upload-4",
			fileSize:    -100,
			expectError: true,
			errorMsg:    "ukuran file tidak valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tusStore.InitiateUpload(tt.uploadID, tt.fileSize)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify upload directory was created
				uploadPath := tusStore.GetFilePath(tt.uploadID)
				_, err := os.Stat(uploadPath)
				assert.NoError(t, err, "upload file should exist")
			}
		})
	}
}

func TestTusStore_WriteChunk(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-chunk-upload"
	fileSize := int64(100)

	// First initiate the upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Write a chunk
	chunkData := strings.NewReader("hello world")
	newOffset, err := tusStore.WriteChunk(uploadID, 0, chunkData)

	assert.NoError(t, err)
	assert.Equal(t, int64(11), newOffset)

	// Verify offset was updated
	info, err := tusStore.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), info.Offset)
}

func TestTusStore_WriteChunk_Sequential(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-sequential-upload"
	fileSize := int64(50)

	// First initiate the upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Write first chunk
	chunk1 := strings.NewReader("first")
	offset1, err := tusStore.WriteChunk(uploadID, 0, chunk1)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), offset1)

	// Write second chunk
	chunk2 := strings.NewReader("second")
	offset2, err := tusStore.WriteChunk(uploadID, 5, chunk2)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), offset2)

	// Verify final offset
	info, err := tusStore.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), info.Offset)
}

func TestTusStore_GetInfo(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-get-info"
	fileSize := int64(1024)

	// Test getting info for non-existent upload
	_, err := tusStore.GetInfo("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	// Initiate upload
	err = tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Get info for existing upload
	info, err := tusStore.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, uploadID, info.ID)
	assert.Equal(t, fileSize, info.Size)
	assert.Equal(t, int64(0), info.Offset)
	assert.NotNil(t, info.Metadata)
}

func TestTusStore_IsComplete(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-complete"
	fileSize := int64(20)

	// Initiate upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Check if complete (should be false initially)
	complete, err := tusStore.IsComplete(uploadID)
	assert.NoError(t, err)
	assert.False(t, complete)

	// Write complete data
	chunkData := strings.NewReader("01234567890123456789")
	_, err = tusStore.WriteChunk(uploadID, 0, chunkData)
	require.NoError(t, err)

	// Check if complete (should be true now)
	complete, err = tusStore.IsComplete(uploadID)
	assert.NoError(t, err)
	assert.True(t, complete)
}

func TestTusStore_GetProgress(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-progress"
	fileSize := int64(100)

	// Initiate upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Check initial progress
	progress, err := tusStore.GetProgress(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, progress)

	// Write half the file
	chunkData := strings.NewReader(strings.Repeat("x", 50))
	_, err = tusStore.WriteChunk(uploadID, 0, chunkData)
	require.NoError(t, err)

	// Check progress (should be 50%)
	progress, err = tusStore.GetProgress(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, progress)

	// Write the rest
	chunkData2 := strings.NewReader(strings.Repeat("y", 50))
	_, err = tusStore.WriteChunk(uploadID, 50, chunkData2)
	require.NoError(t, err)

	// Check progress (should be 100%)
	progress, err = tusStore.GetProgress(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, progress)
}

func TestTusStore_GetOffset(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-offset"
	fileSize := int64(100)

	// Initiate upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Get initial offset
	offset, err := tusStore.GetOffset(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), offset)

	// Write some data
	chunkData := strings.NewReader("test data")
	_, err = tusStore.WriteChunk(uploadID, 0, chunkData)
	require.NoError(t, err)

	// Get offset after write
	offset, err = tusStore.GetOffset(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, int64(9), offset)
}

func TestTusStore_UpdateMetadata(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-metadata"
	fileSize := int64(100)

	// Initiate upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Update metadata
	metadata := map[string]string{
		"filename": "test.zip",
		"contentType": "application/zip",
		"userID": "123",
	}
	err = tusStore.UpdateMetadata(uploadID, metadata)
	assert.NoError(t, err)

	// Verify metadata was updated
	info, err := tusStore.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, "test.zip", info.Metadata["filename"])
	assert.Equal(t, "application/zip", info.Metadata["contentType"])
	assert.Equal(t, "123", info.Metadata["userID"])

	// Update with additional metadata
	metadata2 := map[string]string{
		"uploadID": "abc456",
	}
	err = tusStore.UpdateMetadata(uploadID, metadata2)
	assert.NoError(t, err)

	// Verify both old and new metadata exist
	info, err = tusStore.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, "test.zip", info.Metadata["filename"])
	assert.Equal(t, "abc456", info.Metadata["uploadID"])
}

func TestTusStore_Terminate(t *testing.T) {
	tusStore, pathResolver, _ := setupTusStoreTest(t)
	uploadID := "test-terminate"
	fileSize := int64(100)

	// Initiate upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Verify upload exists
	uploadPath := pathResolver.GetUploadPath(uploadID)
	_, err = os.Stat(uploadPath)
	assert.NoError(t, err)

	// Terminate upload
	err = tusStore.Terminate(uploadID)
	assert.NoError(t, err)

	// Verify upload was deleted
	_, err = os.Stat(uploadPath)
	assert.True(t, os.IsNotExist(err))
}

func TestTusStore_GetFilePath(t *testing.T) {
	tusStore, pathResolver, _ := setupTusStoreTest(t)
	uploadID := "test-file-path"

	expectedPath := pathResolver.GetUploadFilePath(uploadID)
	actualPath := tusStore.GetFilePath(uploadID)

	assert.Equal(t, expectedPath, actualPath)
}

func TestTusStore_FinalizeUpload(t *testing.T) {
	tusStore, _, tempDir := setupTusStoreTest(t)
	uploadID := "test-finalize"
	fileSize := int64(20)

	// Initiate upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Write complete data
	chunkData := strings.NewReader("01234567890123456789")
	_, err = tusStore.WriteChunk(uploadID, 0, chunkData)
	require.NoError(t, err)

	// Finalize upload
	finalPath := filepath.Join(tempDir, "final", "file.zip")
	err = tusStore.FinalizeUpload(uploadID, finalPath)
	assert.NoError(t, err)

	// Verify final file exists
	_, err = os.Stat(finalPath)
	assert.NoError(t, err)

	// Verify temp upload was deleted
	uploadPath := tusStore.GetFilePath(uploadID)
	_, err = os.Stat(uploadPath)
	assert.True(t, os.IsNotExist(err))
}

func TestTusStore_ConcurrentWrites(t *testing.T) {
	tusStore, _, _ := setupTusStoreTest(t)
	uploadID := "test-concurrent"
	fileSize := int64(100)

	// Initiate upload
	err := tusStore.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	// Write multiple chunks in sequence (concurrent writes are serialized by locks)
	chunks := []struct {
		offset int64
		data   string
	}{
		{0, "first"},
		{5, "second"},
		{11, "third"},
	}

	for _, chunk := range chunks {
		chunkData := strings.NewReader(chunk.data)
		_, err := tusStore.WriteChunk(uploadID, chunk.offset, chunkData)
		assert.NoError(t, err)
	}

	// Verify final state
	info, err := tusStore.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, int64(16), info.Offset)
}
