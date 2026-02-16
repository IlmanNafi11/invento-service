package upload_test

import (
	"testing"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// ==================== TusManager Tests ====================

func TestNewTusManager_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := upload.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	fileManager := storage.NewFileManager(cfg)

	manager := upload.NewTusManager(store, queue, fileManager, cfg, zerolog.Nop())

	assert.NotNil(t, manager)
}

func TestTusManager_CheckUploadSlot_Success(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	// Valid base64 encoded metadata
	metadata := "filename dGVzdC56aXA=,content-type YXBwbGljYXRpb24vemlw"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "test.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content-type"])
}

func TestTusManager_ParseMetadata_Empty(t *testing.T) {
	t.Parallel()
	parsed := upload.ParseTusMetadata("")
	assert.Empty(t, parsed)
}

func TestTusManager_ParseMetadata_InvalidBase64(t *testing.T) {
	t.Parallel()
	metadata := "filename invalid_base64!!!"

	parsed := upload.ParseTusMetadata(metadata)
	assert.Equal(t, "", parsed["filename"])
}

func TestTusManager_ValidateModulMetadata_Success(t *testing.T) {
	t.Parallel()
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
