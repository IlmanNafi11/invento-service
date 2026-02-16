package upload_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
)

func TestTusManager_GetUploadProgress_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-progress"
	fileSize := int64(100)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	progress, err := manager.GetUploadProgress(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, float64(0), progress)
}

func TestTusManager_GetUploadProgress_HalfComplete(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-half-progress"
	fileSize := int64(20)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	chunk := []byte("1234567890") // 10 bytes
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	progress, err := manager.GetUploadProgress(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, float64(50), progress)
}

func TestTusManager_AddToQueue_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	manager.AddToQueue("upload-1")

	assert.True(t, queue.HasActiveUpload())
	activeUploads := queue.GetActiveUploads()
	assert.Contains(t, activeUploads, "upload-1")
}

func TestTusManager_AddToQueue_Multiple(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so queue fills up
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	manager.AddToQueue("upload-1") // Active
	manager.AddToQueue("upload-2") // Queued
	manager.AddToQueue("upload-3") // Queued

	activeUploads := queue.GetActiveUploads()
	assert.Contains(t, activeUploads, "upload-1")
	assert.Equal(t, 2, queue.GetQueueLength())
}

func TestTusManager_RemoveFromQueue_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	manager.AddToQueue("upload-1")
	err := manager.RemoveFromQueue("upload-1")

	assert.NoError(t, err)
	assert.False(t, queue.HasActiveUpload())
}

func TestTusManager_RemoveFromQueue_NotFound(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	err := manager.RemoveFromQueue("nonexistent")

	assert.Error(t, err)
}

func TestTusManager_CanAcceptUpload_True(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	assert.True(t, manager.CanAcceptUpload())
}

func TestTusManager_CanAcceptUpload_False(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so it fills up with one upload
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	manager.AddToQueue("upload-1")

	assert.False(t, manager.CanAcceptUpload())
}

func TestTusManager_IsActiveUpload_True(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	manager.AddToQueue("upload-1")

	assert.True(t, manager.IsActiveUpload("upload-1"))
}

func TestTusManager_IsActiveUpload_False(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	manager.AddToQueue("upload-1")

	assert.False(t, manager.IsActiveUpload("upload-2"))
}

func TestTusManager_GetDefaultTusHeaders_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	headers := manager.GetDefaultTusHeaders()

	assert.NotNil(t, headers)
	assert.Equal(t, cfg.Upload.TusVersion, headers["Tus-Resumable"])
}

func TestTusManager_ResetUploadQueue_WithActiveUpload(t *testing.T) {
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

	uploadID := "test-upload-reset"
	fileSize := int64(1024)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	queue.Add(uploadID)

	err = manager.ResetUploadQueue()

	assert.NoError(t, err)
	assert.False(t, queue.HasActiveUpload())
	assert.Equal(t, 0, queue.GetQueueLength())

	// Verify upload was terminated
	_, err = store.GetInfo(uploadID)
	assert.Error(t, err)
}

func TestTusManager_ParseMetadata_WithSpaces(t *testing.T) {
	t.Parallel()
	// Metadata with extra spaces between key and value (spaces around the comma are trimmed)
	metadata := "filename dGVzdC56aXA=, content-type YXBwbGljYXRpb24vemlw"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "test.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content-type"])
}

func TestTusManager_ParseMetadata_InvalidPair(t *testing.T) {
	t.Parallel()
	// Invalid pair format (missing value)
	metadata := "filename dGVzdC56aXA=,invalid-key"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "test.zip", parsed["filename"])
}

func TestTusManager_ValidateModulMetadata_TooLong(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	// Create nama_file > 255 characters
	longName := string(make([]byte, 256))
	for i := range longName {
		longName = longName[:i] + "a"
	}

	metadata := map[string]string{
		"nama_file": longName,
		"tipe":      "pdf",
	}

	err := manager.ValidateModulMetadata(metadata)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "3-255 karakter")
}

func TestTusManager_ValidateModulMetadata_AllTypes(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	validTypes := []string{"docx", "xlsx", "pdf", "pptx"}

	for _, fileTipe := range validTypes {
		metadata := map[string]string{
			"nama_file": "document." + fileTipe,
			"tipe":      fileTipe,
		}

		err := manager.ValidateModulMetadata(metadata)
		assert.NoError(t, err, "Should accept type: "+fileTipe)
	}
}

func TestTusManager_HandleChunk_MultipleChunks(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-multiple"
	fileSize := int64(30)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	// First chunk
	chunk1 := []byte("first chunk data!")
	offset1, err := manager.HandleChunk(uploadID, 0, bytes.NewReader(chunk1))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk1)), offset1)

	// Second chunk
	chunk2 := []byte("second chunk data")
	offset2, err := manager.HandleChunk(uploadID, offset1, bytes.NewReader(chunk2))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk1)+len(chunk2)), offset2)

	// Verify final offset
	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, offset2, info.Offset)
}

func TestTusManager_CheckUploadSlot_QueueFull(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so queue fills up
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	// Fill the queue - first goes active, rest go to queue
	queue.Add("upload-1") // Active
	queue.Add("upload-2") // Queued
	queue.Add("upload-3") // Queued

	response := manager.CheckUploadSlot()

	assert.False(t, response.Available)
	assert.Contains(t, response.Message, "tidak tersedia")
	assert.Equal(t, 2, response.QueueLength) // Two queued besides active
	assert.True(t, response.ActiveUpload)
	assert.Equal(t, 3, response.MaxConcurrent) // From config
}

func TestTusManager_GetUploadStatus_AfterChunk(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	uploadID := "test-upload-status-chunk"
	fileSize := int64(100)

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})
	require.NoError(t, err)

	// Write a chunk
	chunk := []byte("test data")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	offset, size, err := manager.GetUploadStatus(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk)), offset)
	assert.Equal(t, fileSize, size)
}
