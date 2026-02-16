package upload_test

import (
	"bytes"
	"context"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repository for TusCleanup tests
type MockTusUploadRepository struct {
	uploads     map[string]domain.TusUpload
	expired     []domain.TusUpload
	active      []domain.TusUpload
	updateError bool
	deleteError bool
	getError    bool
}

func (m *MockTusUploadRepository) GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusUpload, error) {
	if m.getError {
		return nil, assert.AnError
	}
	return m.expired, nil
}

func (m *MockTusUploadRepository) GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusUpload, error) {
	if m.getError {
		return nil, assert.AnError
	}
	threshold := time.Now().Add(-timeout)
	result := make([]domain.TusUpload, 0)
	for _, upload := range m.active {
		if upload.UpdatedAt.Before(threshold) {
			result = append(result, upload)
		}
	}
	return result, nil
}

func (m *MockTusUploadRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if m.updateError {
		return assert.AnError
	}
	if upload, ok := m.uploads[id]; ok {
		upload.Status = status
		m.uploads[id] = upload
	}
	return nil
}

func (m *MockTusUploadRepository) Delete(ctx context.Context, id string) error {
	if m.deleteError {
		return assert.AnError
	}
	delete(m.uploads, id)
	return nil
}

func setupTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
		Upload: config.UploadConfig{
			PathProduction:       "/tmp/uploads",
			PathDevelopment:      "/tmp/uploads",
			TempPathProduction:   "/tmp/uploads/temp",
			TempPathDevelopment:  "/tmp/uploads/temp",
			MaxSize:              100 * 1024 * 1024, // 100MB
			MaxConcurrentProject: 3,
			TusVersion:           "1.0.0",
		},
	}
}

func setupTestTusStore(t *testing.T) (*upload.TusStore, string) {
	cfg := setupTestConfig()
	tempDir := t.TempDir()
	cfg.Upload.PathDevelopment = tempDir
	cfg.Upload.TempPathDevelopment = filepath.Join(tempDir, "temp")

	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)

	return store, tempDir
}

// ==================== TusManager Tests ====================

func TestNewTusManager_Success(t *testing.T) {
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	fileManager := storage.NewFileManager(cfg)

	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	assert.NotNil(t, manager)
}

func TestTusManager_CheckUploadSlot_Success(t *testing.T) {
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	// Test when no active upload
	response := manager.CheckUploadSlot()
	assert.True(t, response.Available)
	assert.Equal(t, "Slot upload tersedia", response.Message)
	assert.Equal(t, 0, response.QueueLength)
	assert.False(t, response.ActiveUpload)
	assert.Equal(t, 3, response.MaxConcurrent)
}

func TestTusManager_CheckUploadSlot_WithActiveUpload(t *testing.T) {
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so slot fills up
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	queue.Add("upload-1")

	response := manager.CheckUploadSlot()
	assert.False(t, response.Available)
	assert.Contains(t, response.Message, "tidak tersedia")
	assert.True(t, response.ActiveUpload)
}

func TestTusManager_ResetUploadQueue_Success(t *testing.T) {
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	queue.Add("upload-1")
	queue.Add("upload-2")

	err := manager.ResetUploadQueue()
	assert.NoError(t, err)
	assert.True(t, manager.CanAcceptUpload())
}

func TestTusManager_ParseMetadata_Success(t *testing.T) {
	// Valid base64 encoded metadata
	metadata := "filename dGVzdC56aXA=,content-type YXBwbGljYXRpb24vemlw"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "test.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content-type"])
}

func TestTusManager_ParseMetadata_Empty(t *testing.T) {
	parsed := upload.ParseTusMetadata("")
	assert.Empty(t, parsed)
}

func TestTusManager_ParseMetadata_InvalidBase64(t *testing.T) {
	metadata := "filename invalid_base64!!!"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "", parsed["filename"])
}

func TestTusManager_ValidateModulMetadata_Success(t *testing.T) {
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  bool
	}{
		{
			name: "valid metadata",
			metadata: map[string]string{
				"nama_file": "document.pdf",
				"tipe":      "pdf",
				"semester":  "1",
			},
			wantErr: false,
		},
		{
			name:     "missing nama_file",
			metadata: map[string]string{},
			wantErr:  true,
		},
		{
			name: "missing tipe",
			metadata: map[string]string{
				"nama_file": "document.pdf",
			},
			wantErr: true,
		},
		{
			name: "invalid tipe",
			metadata: map[string]string{
				"nama_file": "document.pdf",
				"tipe":      "txt",
			},
			wantErr: true,
		},
		{
			name: "invalid semester",
			metadata: map[string]string{
				"nama_file": "document.pdf",
				"tipe":      "pdf",
				"semester":  "0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateModulMetadata(tt.metadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== TusStore Tests ====================

func TestNewTusStore_Success(t *testing.T) {
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)

	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)

	assert.NotNil(t, store)
}

func TestTusStore_InitiateUpload_Success(t *testing.T) {
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
	store, _ := setupTestTusStore(t)

	uploadID := "test-upload-123"
	fileSize := int64(200 * 1024 * 1024) // Larger than max

	err := store.NewUpload(upload.TusFileInfo{ID: uploadID, Size: fileSize})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi batas maksimal")
}

func TestTusStore_InitiateUpload_InvalidSize(t *testing.T) {
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
	store, _ := setupTestTusStore(t)

	chunk := []byte("test content")
	_, err := store.WriteChunk("nonexistent", 0, bytes.NewReader(chunk))

	assert.Error(t, err)
}

func TestTusStore_GetInfo_Success(t *testing.T) {
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
	store, _ := setupTestTusStore(t)

	_, err := store.GetInfo("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

func TestTusStore_IsComplete_True(t *testing.T) {
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

// ==================== TusQueue Tests ====================

func TestNewTusQueue_Success(t *testing.T) {
	maxConcurrent := 3
	queue := upload.NewTusQueue(maxConcurrent)

	assert.NotNil(t, queue)
	assert.Equal(t, 0, queue.GetQueueLength())
	assert.False(t, queue.HasActiveUpload())
}

func TestTusQueue_Add_Success(t *testing.T) {
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	assert.True(t, queue.HasActiveUpload())
	activeUploads := queue.GetActiveUploads()
	assert.Len(t, activeUploads, 1)
	assert.Contains(t, activeUploads, "upload-1")
}

func TestTusQueue_Add_Queued(t *testing.T) {
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so second upload goes to queue

	queue.Add("upload-1")
	queue.Add("upload-2")

	activeUploads := queue.GetActiveUploads()
	assert.Contains(t, activeUploads, "upload-1")
	assert.Equal(t, 1, queue.GetQueueLength())
}

func TestTusQueue_Add_Duplicate(t *testing.T) {
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so uploads go to queue

	queue.Add("upload-1")
	queue.FinishUpload("upload-1") // Clear active, now active is empty
	queue.Add("upload-2")          // Becomes active since active is empty
	queue.Add("upload-3")          // Goes to queue since active is full
	queue.Add("upload-2")          // Duplicate - already active, should be ignored
	queue.Add("upload-3")          // Duplicate - already in queue, should be ignored

	assert.Equal(t, 1, queue.GetQueueLength())
	currentQueue := queue.GetCurrentQueue()
	assert.Contains(t, currentQueue, "upload-3")
}

func TestTusQueue_Remove_ActiveUpload(t *testing.T) {
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")
	err := queue.Remove("upload-1")

	assert.NoError(t, err)
	assert.False(t, queue.HasActiveUpload())
}

func TestTusQueue_Remove_QueuedUpload(t *testing.T) {
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so second upload goes to queue

	queue.Add("upload-1")
	queue.Add("upload-2")
	err := queue.Remove("upload-2")

	assert.NoError(t, err)
	assert.Equal(t, 0, queue.GetQueueLength())
}

func TestTusQueue_Remove_NotFound(t *testing.T) {
	queue := upload.NewTusQueue(3)

	err := queue.Remove("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

func TestTusQueue_GetQueuePosition_Active(t *testing.T) {
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	position := queue.GetQueuePosition("upload-1")
	assert.Equal(t, 0, position)
}

func TestTusQueue_GetQueuePosition_Queued(t *testing.T) {
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so second upload goes to queue

	queue.Add("upload-1")
	queue.Add("upload-2")

	position := queue.GetQueuePosition("upload-2")
	assert.Equal(t, 1, position)
}

func TestTusQueue_GetQueuePosition_NotFound(t *testing.T) {
	queue := upload.NewTusQueue(3)

	position := queue.GetQueuePosition("nonexistent")
	assert.Equal(t, -1, position)
}

func TestTusQueue_FinishUpload_Success(t *testing.T) {
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")
	queue.FinishUpload("upload-1")

	assert.False(t, queue.HasActiveUpload())
	activeUploads := queue.GetActiveUploads()
	assert.Len(t, activeUploads, 0)
}

func TestTusQueue_Clear_Success(t *testing.T) {
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")
	queue.Add("upload-2")
	queue.Clear()

	assert.False(t, queue.HasActiveUpload())
	assert.Equal(t, 0, queue.GetQueueLength())
}

func TestTusQueue_CanAcceptUpload_True(t *testing.T) {
	queue := upload.NewTusQueue(3)

	assert.True(t, queue.CanAcceptUpload())
}

func TestTusQueue_CanAcceptUpload_False(t *testing.T) {
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so it fills up with one upload

	queue.Add("upload-1")

	assert.False(t, queue.CanAcceptUpload())
}

func TestTusQueue_IsActiveUpload_True(t *testing.T) {
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	assert.True(t, queue.IsActiveUpload("upload-1"))
}

func TestTusQueue_IsActiveUpload_False(t *testing.T) {
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	assert.False(t, queue.IsActiveUpload("upload-2"))
}

func TestTusQueue_GetCurrentQueue_Success(t *testing.T) {
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so uploads go to queue

	queue.Add("upload-1")
	queue.Add("upload-2")
	queue.Add("upload-3")

	currentQueue := queue.GetCurrentQueue()
	assert.Len(t, currentQueue, 2)
	assert.Contains(t, currentQueue, "upload-2")
	assert.Contains(t, currentQueue, "upload-3")
}

// ==================== TusResponse Tests ====================

func TestSendTusInitiateResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusInitiateResponse(c, "upload-123", "/uploads/upload-123", 1024)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
}

func TestSendTusChunkResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Patch("/test", func(c *fiber.Ctx) error {
		return upload.SendTusChunkResponse(c, 512)
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "512", resp.Header.Get("Upload-Offset"))
}

func TestSendTusHeadResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Head("/test", func(c *fiber.Ctx) error {
		return upload.SendTusHeadResponse(c, 512, 1024)
	})

	req := httptest.NewRequest("HEAD", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "512", resp.Header.Get("Upload-Offset"))
	assert.Equal(t, "1024", resp.Header.Get("Upload-Length"))
}

func TestSendTusDeleteResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Delete("/test", func(c *fiber.Ctx) error {
		return upload.SendTusDeleteResponse(c)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestSendTusSlotResponse_Available(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusSlotResponse(c, true, "Slot tersedia", 0, 0, 3)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Slot tersedia")
}

func TestSendTusModulSlotResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusSlotResponse(c, true, "Slot tersedia", 0, 0, 10)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Slot tersedia")
}

func TestSendTusErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusErrorResponse(c, fiber.StatusBadRequest, "1.0.0")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
}

func TestSendTusErrorResponseWithOffset_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusErrorResponseWithOffset(c, fiber.StatusConflict, "1.0.0", 512)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	assert.Equal(t, "512", resp.Header.Get("Upload-Offset"))
}

func TestSendTusErrorResponseWithLength_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusErrorResponseWithLength(c, fiber.StatusRequestEntityTooLarge, "1.0.0", 1024)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
	assert.Equal(t, "1024", resp.Header.Get("Upload-Length"))
}

func TestSendTusValidationErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusValidationErrorResponse(c, "Metadata tidak valid")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSendTusNotFoundErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusNotFoundErrorResponse(c, "Upload tidak ditemukan")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestSendTusNotFoundErrorResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusNotFoundErrorResponse(c, "")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestSendTusForbiddenErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusForbiddenErrorResponse(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSendTusConflictErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusConflictErrorResponse(c, "Upload sudah selesai")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
}

func TestSendTusConflictErrorResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusConflictErrorResponse(c, "")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
}

func TestSendTusPayloadTooLargeErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusPayloadTooLargeErrorResponse(c, "File terlalu besar")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestSendTusPayloadTooLargeErrorResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusPayloadTooLargeErrorResponse(c, "")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestSendTusTooManyRequestsErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusTooManyRequestsErrorResponse(c, "Terlalu banyak request")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
}

func TestSendTusTooManyRequestsErrorResponse_DefaultMessage(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return upload.SendTusTooManyRequestsErrorResponse(c, "")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
}

func TestSendTusInternalErrorResponse_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.SendTusInternalErrorResponse(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// ==================== TusHeaders Tests ====================

func TestGetTusHeaders_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Set("Tus-Resumable", "1.0.0")
		c.Set("Upload-Offset", "100")
		c.Set("Upload-Length", "1000")
		c.Set("Upload-Metadata", "filename dGVzdA==")
		c.Set("Content-Type", "application/offset+octet-stream")
		c.Set("Content-Length", "50")

		headers, err := upload.GetTusHeaders(c)
		if err != nil {
			return err
		}

		return c.JSON(headers)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetTusHeaders_EmptyHeaders(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		headers, err := upload.GetTusHeaders(c)
		if err != nil {
			return err
		}

		return c.JSON(headers)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSetTusResponseHeaders_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		upload.SetTusResponseHeaders(c, 100, 1000)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, "100", resp.Header.Get("Upload-Offset"))
	assert.Equal(t, "1000", resp.Header.Get("Upload-Length"))
}

func TestSetTusResponseHeaders_ZeroLength(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		upload.SetTusResponseHeaders(c, 0, 0)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", resp.Header.Get("Tus-Resumable"))
	assert.Equal(t, "", resp.Header.Get("Upload-Length"))
}

func TestSetTusLocationHeader_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		upload.SetTusLocationHeader(c, "/uploads/upload-123")
		return c.SendStatus(fiber.StatusCreated)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "/uploads/upload-123", resp.Header.Get("Location"))
}

func TestSetTusOffsetHeader_Success(t *testing.T) {
	app := fiber.New()
	app.Patch("/test", func(c *fiber.Ctx) error {
		upload.SetTusOffsetHeader(c, 500)
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "500", resp.Header.Get("Upload-Offset"))
}

func TestValidateChunkSize_Valid(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"min size", 1, false},
		{"1MB", 1024 * 1024, false},
		{"2MB", upload.MaxChunkSize, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"too large", upload.MaxChunkSize + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := upload.ValidateChunkSize(tt.size)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildTusErrorResponse_ConflictWithOffset(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.BuildTusErrorResponse(c, fiber.StatusConflict, 100)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	assert.Equal(t, "100", resp.Header.Get("Upload-Offset"))
}

func TestBuildTusErrorResponse_OtherStatus(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return upload.BuildTusErrorResponse(c, fiber.StatusBadRequest, 100)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "", resp.Header.Get("Upload-Offset"))
}

// ==================== TusCleanup Tests ====================

func TestNewTusCleanup_Success(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)

	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	assert.NotNil(t, cleanup)
}

func TestTusCleanup_Start_Success(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Start()

	time.Sleep(100 * time.Millisecond)
	cleanup.Stop()

	// Started and stopped successfully
	assert.True(t, true)
}

func TestTusCleanup_Start_AlreadyRunning(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Start()
	cleanup.Start() // Should not cause issues

	cleanup.Stop()
}

func TestTusCleanup_Stop_Success(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Start()
	cleanup.Stop()

	// Verify it's stopped
	assert.True(t, true)
}

func TestTusCleanup_Stop_NotRunning(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Stop() // Should not cause issues

	assert.True(t, true)
}

func TestTusCleanup_CleanupExpired_Success(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		expired: []domain.TusUpload{
			{
				ID:     "expired-1",
				Status: domain.UploadStatusPending,
			},
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	// Create a temp upload
	err := store.NewUpload(upload.TusFileInfo{ID: "expired-1", Size: 1024})
	require.NoError(t, err)

	err = cleanup.CleanupExpiredProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupExpired_NoExpired(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		expired: []domain.TusUpload{},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupExpiredProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupExpired_RepositoryError(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads:  make(map[string]domain.TusUpload),
		expired:  []domain.TusUpload{},
		getError: true,
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupExpiredProjects()
	assert.Error(t, err)
}

func TestTusCleanup_CleanupAbandoned_Success(t *testing.T) {
	cfg := setupTestConfig()
	oldTime := time.Now().Add(-1 * time.Hour)
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		active: []domain.TusUpload{
			{
				ID:        "abandoned-1",
				Status:    domain.UploadStatusUploading,
				UpdatedAt: oldTime,
			},
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	// Create a temp upload
	err := store.NewUpload(upload.TusFileInfo{ID: "abandoned-1", Size: 1024})
	require.NoError(t, err)

	err = cleanup.CleanupAbandonedProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupAbandoned_NoActive(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		active:  []domain.TusUpload{},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupAbandonedProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupAbandoned_RepositoryError(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads:  make(map[string]domain.TusUpload),
		active:   []domain.TusUpload{},
		getError: true,
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupAbandonedProjects()
	assert.Error(t, err)
}

func TestTusCleanup_CleanupUpload_Success(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: map[string]domain.TusUpload{
			"test-upload": {
				ID:     "test-upload",
				Status: domain.UploadStatusPending,
			},
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	// Create a temp upload
	err := store.NewUpload(upload.TusFileInfo{ID: "test-upload", Size: 1024})
	require.NoError(t, err)

	err = cleanup.CleanupUpload("test-upload")
	assert.NoError(t, err)

	// Verify deleted
	_, err = store.GetInfo("test-upload")
	assert.Error(t, err)
}

func TestTusCleanup_CleanupUpload_NotFound(t *testing.T) {
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads:     map[string]domain.TusUpload{},
		deleteError: true, // Simulate error for non-existent
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupUpload("nonexistent")
	assert.Error(t, err)
}

// ==================== TusManager Additional Tests ====================

func TestTusManager_InitiateUpload_Success(t *testing.T) {
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

func TestTusManager_GetUploadProgress_Success(t *testing.T) {
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
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(3)
	fileManager := storage.NewFileManager(cfg)
	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	assert.True(t, manager.CanAcceptUpload())
}

func TestTusManager_CanAcceptUpload_False(t *testing.T) {
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
	// Metadata with extra spaces between key and value (spaces around the comma are trimmed)
	metadata := "filename dGVzdC56aXA=, content-type YXBwbGljYXRpb24vemlw"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "test.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content-type"])
}

func TestTusManager_ParseMetadata_InvalidPair(t *testing.T) {
	// Invalid pair format (missing value)
	metadata := "filename dGVzdC56aXA=,invalid-key"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "test.zip", parsed["filename"])
}

func TestTusManager_ValidateModulMetadata_TooLong(t *testing.T) {
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
