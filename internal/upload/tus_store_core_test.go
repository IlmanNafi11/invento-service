package upload_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== TusStore Tests ====================

func TestNewTusStore_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)

	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)

	assert.NotNil(t, store)
}

func TestTusStore_InitiateUpload_Success(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})

	assert.NoError(t, err)

	// Verify info was saved
	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, uploadID, info.ID)
	assert.Equal(t, fileSize, info.Size)
	assert.Equal(t, int64(0), info.Offset)
}

func TestTusStore_InitiateUpload_FileTooLarge(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(200 * 1024 * 1024) // Larger than max

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi batas maksimal")
}

func TestTusStore_InitiateUpload_InvalidSize(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	tests := []struct {
		name     string
		fileSize int64
		wantErr  bool
	}{
		{"zero size", 0, true},
		{"negative size", -1, true},
		{"valid size", 1024, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.NewUpload(upload.TusFileInfo{ID: "test-" + tt.name, Size: tt.fileSize})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTusStore_WriteChunk_Success(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	chunk := []byte("test chunk content")
	newOffset, err := store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk)), newOffset)

	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, newOffset, info.Offset)
}

func TestTusStore_WriteChunk_UploadNotFound(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	chunk := []byte("test content")
	_, err := store.WriteChunk("nonexistent", 0, bytes.NewReader(chunk))

	assert.Error(t, err)
}

func TestTusStore_GetInfo_Success(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	info, err := store.GetInfo(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, uploadID, info.ID)
	assert.Equal(t, fileSize, info.Size)
}

func TestTusStore_GetInfo_NotFound(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	_, err := store.GetInfo("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

func TestTusStore_IsComplete_True(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(20)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	chunk := []byte("test chunk content test!")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	isComplete, err := store.IsComplete(uploadID)

	assert.NoError(t, err)
	assert.True(t, isComplete)
}

func TestTusStore_IsComplete_False(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(100)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	isComplete, err := store.IsComplete(uploadID)

	assert.NoError(t, err)
	assert.False(t, isComplete)
}

func TestTusStore_GetProgress_Success(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(100)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	progress, err := store.GetProgress(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, float64(0), progress)
}

func TestTusStore_GetProgress_HalfComplete(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(20)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	chunk := []byte("1234567890") // 10 bytes
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	progress, err := store.GetProgress(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, float64(50), progress)
}

func TestTusStore_GetOffset_Success(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	offset, err := store.GetOffset(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), offset)
}

func TestTusStore_Terminate_Success(t *testing.T) {
	t.Parallel()
	store, tempDir := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	err = store.Terminate(uploadID)

	assert.NoError(t, err)

	// Verify files are deleted
	uploadPath := filepath.Join(tempDir, "temp", "uploads", uploadID)
	_, err = os.Stat(uploadPath)
	assert.True(t, os.IsNotExist(err))
}

func TestTusStore_UpdateMetadata_Success(t *testing.T) {
	t.Parallel()
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	metadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	err = store.UpdateMetadata(uploadID, metadata)
	assert.NoError(t, err)

	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, "value1", info.Metadata["key1"])
	assert.Equal(t, "value2", info.Metadata["key2"])
}

func TestTusStore_FinalizeUpload_Success(t *testing.T) {
	t.Parallel()
	store, tempDir := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(20)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	chunk := []byte("test chunk content test!")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	finalPath := filepath.Join(tempDir, "final", "file.zip")
	err = store.FinalizeUpload(uploadID, finalPath)

	assert.NoError(t, err)

	// Verify file was moved
	_, err = os.Stat(finalPath)
	assert.NoError(t, err)
}
