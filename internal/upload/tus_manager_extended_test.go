package upload_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"invento-service/internal/storage"
	"invento-service/internal/upload"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== TusManager Additional Tests ====================

func TestTusManager_InitiateUpload_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-123"
	fileSize := int64(1024)
	metadata := map[string]string{"filename": "test.zip"}

	err := manager.InitiateUpload(uploadID, fileSize, metadata)

	assert.NoError(t, err)

	// Verify upload was initiated
	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, uploadID, info.ID)
	assert.Equal(t, fileSize, info.Size)
}

func TestTusManager_InitiateUpload_FileTooLarge(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-large"
	fileSize := cfg.Upload.MaxSize + 1
	metadata := map[string]string{"filename": "large.zip"}

	err := manager.InitiateUpload(uploadID, fileSize, metadata)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi batas maksimal")
}

func TestTusManager_InitiateUpload_InvalidSize(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	tests := []struct {
		name     string
		fileSize int64
		wantErr  bool
	}{
		{"zero size", 0, true},
		{"negative size", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.InitiateUpload("test-"+tt.name, tt.fileSize, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTusManager_HandleChunk_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-chunk"
	fileSize := int64(1024)

	// First initiate upload
	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	// Handle chunk
	chunk := []byte("test chunk data")
	newOffset, err := manager.HandleChunk(uploadID, 0, bytes.NewReader(chunk))

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk)), newOffset)

	// Verify offset updated
	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, newOffset, info.Offset)
}

func TestTusManager_HandleChunk_UploadNotFound(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	chunk := []byte("test data")
	_, err := manager.HandleChunk("nonexistent", 0, bytes.NewReader(chunk))

	assert.Error(t, err)
}

func TestTusManager_GetUploadStatus_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-status"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	offset, size, err := manager.GetUploadStatus(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, fileSize, size)
}

func TestTusManager_GetUploadStatus_NotFound(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	_, _, err := manager.GetUploadStatus("nonexistent")

	assert.Error(t, err)
}

func TestTusManager_GetUploadInfo_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-info"
	fileSize := int64(2048)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	info, err := manager.GetUploadInfo(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, uploadID, info.ID)
	assert.Equal(t, fileSize, info.Size)
	assert.Equal(t, int64(0), info.Offset)
}

func TestTusManager_GetUploadInfo_NotFound(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	_, err := manager.GetUploadInfo("nonexistent")

	assert.Error(t, err)
}

func TestTusManager_CancelUpload_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-cancel"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	err = manager.CancelUpload(uploadID)

	assert.NoError(t, err)

	// Verify upload was terminated
	_, err = store.GetInfo(uploadID)
	assert.Error(t, err)
}

func TestTusManager_CancelUpload_NotFound(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	// Terminate on non-existent upload doesn't return error (os.RemoveAll succeeds)
	err := manager.CancelUpload("nonexistent")

	assert.NoError(t, err)
}

func TestTusManager_FinalizeUpload_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	tempDir := t.TempDir()
	pathResolver := storage.NewPathResolver(cfg)
	cfg.Upload.PathDevelopment = tempDir
	cfg.Upload.TempPathDevelopment = filepath.Join(tempDir, "temp")
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-finalize"
	fileSize := int64(20)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	chunk := []byte("test chunk content test!")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	finalPath := filepath.Join(tempDir, "final", "file.zip")
	err = manager.FinalizeUpload(uploadID, finalPath)

	assert.NoError(t, err)

	// Verify file was moved
	_, err = os.Stat(finalPath)
	assert.NoError(t, err)
}

func TestTusManager_FinalizeUpload_NotComplete(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	tempDir := t.TempDir()
	pathResolver := storage.NewPathResolver(cfg)
	cfg.Upload.PathDevelopment = tempDir
	cfg.Upload.TempPathDevelopment = filepath.Join(tempDir, "temp")
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-incomplete"
	fileSize := int64(100)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	// Write partial data (not complete)
	chunk := []byte("partial data")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	finalPath := filepath.Join(tempDir, "final", "file.zip")
	err = manager.FinalizeUpload(uploadID, finalPath)

	// FinalizeUpload will succeed even if incomplete (it just moves the file)
	// The completion check happens elsewhere
	assert.NoError(t, err)
}

func TestTusManager_IsUploadComplete_True(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-complete"
	fileSize := int64(20)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	chunk := []byte("test chunk content test!")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	isComplete, err := manager.IsUploadComplete(uploadID)

	assert.NoError(t, err)
	assert.True(t, isComplete)
}

func TestTusManager_IsUploadComplete_False(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-pending"
	fileSize := int64(100)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	isComplete, err := manager.IsUploadComplete(uploadID)

	assert.NoError(t, err)
	assert.False(t, isComplete)
}
