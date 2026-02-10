package helper

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"fiber-boiler-plate/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig creates a test configuration
func setupTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
		Upload: config.UploadConfig{
			PathProduction:      "/tmp/uploads",
			PathDevelopment:     "/tmp/uploads",
			TempPathProduction:  "/tmp/uploads/temp",
			TempPathDevelopment: "/tmp/uploads/temp",
			MaxSize:             100 * 1024 * 1024, // 100MB
			MaxConcurrentProject: 3,
			TusVersion:          "1.0.0",
		},
	}
}

// setupTestTusManager creates a test TusManager with temp directory
func setupTestTusManager(t *testing.T) (*TusManager, *TusStore) {
	cfg := setupTestConfig()
	tempDir := t.TempDir()
	cfg.Upload.PathDevelopment = tempDir
	cfg.Upload.TempPathDevelopment = filepath.Join(tempDir, "temp")

	pathResolver := NewPathResolver(cfg)
	store := NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := NewTusQueue(cfg.Upload.MaxConcurrentProject)
	fileManager := NewFileManager(cfg)
	manager := NewTusManager(store, queue, fileManager, cfg)

	_ = tempDir
	return manager, store
}

// ==================== NewTusManager Tests ====================

func TestNewTusManager_Success(t *testing.T) {
	cfg := setupTestConfig()
	pathResolver := NewPathResolver(cfg)
	store := NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := NewTusQueue(cfg.Upload.MaxConcurrentProject)
	fileManager := NewFileManager(cfg)

	manager := NewTusManager(store, queue, fileManager, cfg)

	assert.NotNil(t, manager)
	assert.Equal(t, store, manager.store)
	assert.Equal(t, queue, manager.queue)
	assert.Equal(t, fileManager, manager.fileManager)
	assert.Equal(t, cfg, manager.config)
}

// ==================== CheckUploadSlot Tests ====================

func TestTusManager_CheckUploadSlot_NoActiveUpload(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	response := manager.CheckUploadSlot()

	assert.True(t, response.Available)
	assert.Equal(t, "Slot upload tersedia", response.Message)
	assert.Equal(t, 0, response.QueueLength)
	assert.False(t, response.ActiveUpload)
	assert.Equal(t, 3, response.MaxConcurrent)
}

func TestTusManager_CheckUploadSlot_WithActiveUpload(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.queue.Add("upload-1")

	response := manager.CheckUploadSlot()

	assert.False(t, response.Available)
	assert.Contains(t, response.Message, "tidak tersedia")
	assert.Contains(t, response.Message, "0")
	assert.True(t, response.ActiveUpload)
	assert.Equal(t, 3, response.MaxConcurrent)
}

func TestTusManager_CheckUploadSlot_QueueFull(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.queue.Add("upload-1")
	manager.queue.Add("upload-2")
	manager.queue.Add("upload-3")

	response := manager.CheckUploadSlot()

	assert.False(t, response.Available)
	assert.Contains(t, response.Message, "tidak tersedia")
	assert.Contains(t, response.Message, "2")
	assert.Equal(t, 2, response.QueueLength)
	assert.True(t, response.ActiveUpload)
}

// ==================== ResetUploadQueue Tests ====================

func TestTusManager_ResetUploadQueue_EmptyQueue(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	err := manager.ResetUploadQueue()

	assert.NoError(t, err)
	assert.True(t, manager.CanAcceptUpload())
}

func TestTusManager_ResetUploadQueue_WithActiveUpload(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-reset"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	manager.queue.Add(uploadID)

	err = manager.ResetUploadQueue()

	assert.NoError(t, err)
	assert.False(t, manager.queue.HasActiveUpload())
	assert.Equal(t, 0, manager.queue.GetQueueLength())

	// Verify upload was terminated
	_, err = store.GetInfo(uploadID)
	assert.Error(t, err)
}

func TestTusManager_ResetUploadQueue_WithQueuedUploads(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.queue.Add("upload-1")
	manager.queue.Add("upload-2")
	manager.queue.Add("upload-3")

	err := manager.ResetUploadQueue()

	assert.NoError(t, err)
	assert.False(t, manager.queue.HasActiveUpload())
	assert.Equal(t, 0, manager.queue.GetQueueLength())
}

// ==================== ParseMetadata Tests ====================

func TestTusManager_ParseMetadata_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := "filename dGVzdC56aXA=,content-type YXBwbGljYXRpb24vemlw"

	parsed, err := manager.ParseMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, "test.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content-type"])
}

func TestTusManager_ParseMetadata_Empty(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	parsed, err := manager.ParseMetadata("")

	assert.NoError(t, err)
	assert.Empty(t, parsed)
}

func TestTusManager_ParseMetadata_SingleValue(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := "filename dGVzdC56aXA="

	parsed, err := manager.ParseMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, "test.zip", parsed["filename"])
}

func TestTusManager_ParseMetadata_MultipleValues(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := "filename dGVzdC56aXA=,content-type YXBwbGljYXRpb24vemlw,user-id MTIz"

	parsed, err := manager.ParseMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, "test.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content-type"])
	assert.Equal(t, "123", parsed["user-id"])
}

func TestTusManager_ParseMetadata_InvalidBase64(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := "filename invalid_base64!!!"

	_, err := manager.ParseMetadata(metadata)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding")
}

func TestTusManager_ParseMetadata_InvalidPair(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := "filename dGVzdC56aXA=,invalid-key"

	parsed, err := manager.ParseMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, "test.zip", parsed["filename"])
}

func TestTusManager_ParseMetadata_WithSpaces(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := "filename dGVzdC56aXA=, content-type YXBwbGljYXRpb24vemlw"

	parsed, err := manager.ParseMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, "test.zip", parsed["filename"])
	assert.Equal(t, "application/zip", parsed["content-type"])
}

// ==================== ValidateProjectMetadata Tests ====================

func TestTusManager_ValidateProjectMetadata_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  bool
	}{
		{
			name: "valid with all fields",
			metadata: map[string]string{
				"nama_project": "Test Project",
				"kategori":     "website",
				"semester":     "1",
			},
			wantErr: false,
		},
		{
			name: "valid with default kategori",
			metadata: map[string]string{
				"nama_project": "Test Project",
			},
			wantErr: false,
		},
		{
			name: "valid mobile category",
			metadata: map[string]string{
				"nama_project": "Mobile App",
				"kategori":     "mobile",
			},
			wantErr: false,
		},
		{
			name: "valid iot category",
			metadata: map[string]string{
				"nama_project": "IoT Device",
				"kategori":     "iot",
			},
			wantErr: false,
		},
		{
			name: "valid machine_learning category",
			metadata: map[string]string{
				"nama_project": "ML Model",
				"kategori":     "machine_learning",
			},
			wantErr: false,
		},
		{
			name: "valid deep_learning category",
			metadata: map[string]string{
				"nama_project": "DL Model",
				"kategori":     "deep_learning",
			},
			wantErr: false,
		},
		{
			name: "valid max semester",
			metadata: map[string]string{
				"nama_project": "Test Project",
				"semester":     "8",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataCopy := make(map[string]string)
			for k, v := range tt.metadata {
				metadataCopy[k] = v
			}
			err := manager.ValidateProjectMetadata(metadataCopy)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTusManager_ValidateProjectMetadata_Errors(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  string
	}{
		{
			name:     "missing nama_project",
			metadata: map[string]string{},
			wantErr:  "nama_project wajib diisi",
		},
		{
			name: "empty nama_project",
			metadata: map[string]string{
				"nama_project": "",
			},
			wantErr: "nama_project wajib diisi",
		},
		{
			name: "nama_project too short",
			metadata: map[string]string{
				"nama_project": "ab",
			},
			wantErr: "nama_project harus antara 3-255 karakter",
		},
		{
			name: "nama_project too long",
			metadata: map[string]string{
				"nama_project": string(make([]byte, 256)),
			},
			wantErr: "nama_project harus antara 3-255 karakter",
		},
		{
			name: "invalid kategori",
			metadata: map[string]string{
				"nama_project": "Test Project",
				"kategori":     "invalid",
			},
			wantErr: "kategori tidak valid",
		},
		{
			name: "invalid semester - zero",
			metadata: map[string]string{
				"nama_project": "Test Project",
				"semester":     "0",
			},
			wantErr: "semester harus antara 1-8",
		},
		{
			name: "invalid semester - negative",
			metadata: map[string]string{
				"nama_project": "Test Project",
				"semester":     "-1",
			},
			wantErr: "semester harus antara 1-8",
		},
		{
			name: "invalid semester - too high",
			metadata: map[string]string{
				"nama_project": "Test Project",
				"semester":     "9",
			},
			wantErr: "semester harus antara 1-8",
		},
		{
			name: "invalid semester - not a number",
			metadata: map[string]string{
				"nama_project": "Test Project",
				"semester":     "abc",
			},
			wantErr: "semester harus antara 1-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataCopy := make(map[string]string)
			for k, v := range tt.metadata {
				metadataCopy[k] = v
			}
			err := manager.ValidateProjectMetadata(metadataCopy)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestTusManager_ValidateProjectMetadata_SetsDefaultKategori(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := map[string]string{
		"nama_project": "Test Project",
	}

	err := manager.ValidateProjectMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, "website", metadata["kategori"])
}

// ==================== ValidateModulMetadata Tests ====================

func TestTusManager_ValidateModulMetadata_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	validTypes := []string{"docx", "xlsx", "pdf", "pptx"}

	for _, fileTipe := range validTypes {
		metadata := map[string]string{
			"nama_file": "document." + fileTipe,
			"tipe":      fileTipe,
			"semester":  "1",
		}

		err := manager.ValidateModulMetadata(metadata)
		assert.NoError(t, err, "Should accept type: "+fileTipe)
	}
}

func TestTusManager_ValidateModulMetadata_Errors(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  string
	}{
		{
			name:     "missing nama_file",
			metadata: map[string]string{},
			wantErr:  "nama_file wajib diisi",
		},
		{
			name: "empty nama_file",
			metadata: map[string]string{
				"nama_file": "",
			},
			wantErr: "nama_file wajib diisi",
		},
		{
			name: "nama_file too short",
			metadata: map[string]string{
				"nama_file": "ab",
			},
			wantErr: "nama_file harus antara 3-255 karakter",
		},
		{
			name: "nama_file too long",
			metadata: map[string]string{
				"nama_file": string(make([]byte, 256)),
			},
			wantErr: "nama_file harus antara 3-255 karakter",
		},
		{
			name: "missing tipe",
			metadata: map[string]string{
				"nama_file": "document.pdf",
			},
			wantErr: "tipe wajib diisi",
		},
		{
			name: "empty tipe",
			metadata: map[string]string{
				"nama_file": "document.pdf",
				"tipe":      "",
			},
			wantErr: "tipe wajib diisi",
		},
		{
			name: "invalid tipe",
			metadata: map[string]string{
				"nama_file": "document.txt",
				"tipe":      "txt",
			},
			wantErr: "tipe file tidak valid",
		},
		{
			name: "invalid semester",
			metadata: map[string]string{
				"nama_file": "document.pdf",
				"tipe":      "pdf",
				"semester":  "0",
			},
			wantErr: "semester harus antara 1-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataCopy := make(map[string]string)
			for k, v := range tt.metadata {
				metadataCopy[k] = v
			}
			err := manager.ValidateModulMetadata(metadataCopy)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// ==================== InitiateUpload Tests ====================

func TestTusManager_InitiateUpload_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)

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
	manager, _ := setupTestTusManager(t)

	uploadID := "test-upload-large"
	fileSize := manager.config.Upload.MaxSize + 1
	metadata := map[string]string{"filename": "large.zip"}

	err := manager.InitiateUpload(uploadID, fileSize, metadata)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi batas maksimal")
}

func TestTusManager_InitiateUpload_InvalidSize(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	tests := []struct {
		name     string
		fileSize int64
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "zero size",
			fileSize: 0,
			wantErr:  true,
			errMsg:   "ukuran file tidak valid",
		},
		{
			name:     "negative size",
			fileSize: -1,
			wantErr:  true,
			errMsg:   "ukuran file tidak valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.InitiateUpload("test-"+tt.name, tt.fileSize, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// ==================== HandleChunk Tests ====================

func TestTusManager_HandleChunk_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-chunk"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("test chunk data")
	newOffset, err := manager.HandleChunk(uploadID, 0, bytes.NewReader(chunk))

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk)), newOffset)

	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, newOffset, info.Offset)
}

func TestTusManager_HandleChunk_UploadNotFound(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	chunk := []byte("test data")
	_, err := manager.HandleChunk("nonexistent", 0, bytes.NewReader(chunk))

	assert.Error(t, err)
}

func TestTusManager_HandleChunk_MultipleChunks(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-multiple"
	fileSize := int64(30)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk1 := []byte("first chunk data!")
	offset1, err := manager.HandleChunk(uploadID, 0, bytes.NewReader(chunk1))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk1)), offset1)

	chunk2 := []byte("second chunk data")
	offset2, err := manager.HandleChunk(uploadID, offset1, bytes.NewReader(chunk2))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk1)+len(chunk2)), offset2)

	info, err := store.GetInfo(uploadID)
	assert.NoError(t, err)
	assert.Equal(t, offset2, info.Offset)
}

// ==================== GetUploadStatus Tests ====================

func TestTusManager_GetUploadStatus_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-status"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	offset, size, err := manager.GetUploadStatus(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, fileSize, size)
}

func TestTusManager_GetUploadStatus_AfterChunk(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-status-chunk"
	fileSize := int64(100)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("test data")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	offset, size, err := manager.GetUploadStatus(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunk)), offset)
	assert.Equal(t, fileSize, size)
}

func TestTusManager_GetUploadStatus_NotFound(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	_, _, err := manager.GetUploadStatus("nonexistent")

	assert.Error(t, err)
}

// ==================== GetUploadInfo Tests ====================

func TestTusManager_GetUploadInfo_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-info"
	fileSize := int64(2048)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	info, err := manager.GetUploadInfo(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, uploadID, info.ID)
	assert.Equal(t, fileSize, info.Size)
	assert.Equal(t, int64(0), info.Offset)
}

func TestTusManager_GetUploadInfo_NotFound(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	_, err := manager.GetUploadInfo("nonexistent")

	assert.Error(t, err)
}

// ==================== CancelUpload Tests ====================

func TestTusManager_CancelUpload_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-cancel"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	err = manager.CancelUpload(uploadID)

	assert.NoError(t, err)

	_, err = store.GetInfo(uploadID)
	assert.Error(t, err)
}

func TestTusManager_CancelUpload_NotFound(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	err := manager.CancelUpload("nonexistent")

	assert.NoError(t, err)
}

// ==================== FinalizeUpload Tests ====================

func TestTusManager_FinalizeUpload_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)
	tempDir := t.TempDir()

	uploadID := "test-upload-finalize"
	fileSize := int64(20)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("test chunk content test!")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	finalPath := filepath.Join(tempDir, "final", "file.zip")
	err = manager.FinalizeUpload(uploadID, finalPath)

	assert.NoError(t, err)
}

func TestTusManager_FinalizeUpload_NotComplete(t *testing.T) {
	manager, store := setupTestTusManager(t)
	tempDir := t.TempDir()

	uploadID := "test-upload-incomplete"
	fileSize := int64(100)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("partial data")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	finalPath := filepath.Join(tempDir, "final", "file.zip")
	err = manager.FinalizeUpload(uploadID, finalPath)

	assert.NoError(t, err)
}

// ==================== IsUploadComplete Tests ====================

func TestTusManager_IsUploadComplete_True(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-complete"
	fileSize := int64(20)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("test chunk content test!")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	isComplete, err := manager.IsUploadComplete(uploadID)

	assert.NoError(t, err)
	assert.True(t, isComplete)
}

func TestTusManager_IsUploadComplete_False(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-pending"
	fileSize := int64(100)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	isComplete, err := manager.IsUploadComplete(uploadID)

	assert.NoError(t, err)
	assert.False(t, isComplete)
}

// ==================== GetUploadProgress Tests ====================

func TestTusManager_GetUploadProgress_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-progress"
	fileSize := int64(100)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	progress, err := manager.GetUploadProgress(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, float64(0), progress)
}

func TestTusManager_GetUploadProgress_HalfComplete(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-half-progress"
	fileSize := int64(20)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("1234567890")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	progress, err := manager.GetUploadProgress(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, float64(50), progress)
}

// ==================== AddToQueue Tests ====================

func TestTusManager_AddToQueue_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.AddToQueue("upload-1")

	assert.True(t, manager.queue.HasActiveUpload())
	assert.Equal(t, "upload-1", manager.queue.GetActiveUpload())
}

func TestTusManager_AddToQueue_Multiple(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.AddToQueue("upload-1")
	manager.AddToQueue("upload-2")
	manager.AddToQueue("upload-3")

	assert.Equal(t, "upload-1", manager.queue.GetActiveUpload())
	assert.Equal(t, 2, manager.queue.GetQueueLength())
}

// ==================== RemoveFromQueue Tests ====================

func TestTusManager_RemoveFromQueue_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.AddToQueue("upload-1")
	err := manager.RemoveFromQueue("upload-1")

	assert.NoError(t, err)
	assert.False(t, manager.queue.HasActiveUpload())
}

func TestTusManager_RemoveFromQueue_NotFound(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	err := manager.RemoveFromQueue("nonexistent")

	assert.Error(t, err)
}

// ==================== CanAcceptUpload Tests ====================

func TestTusManager_CanAcceptUpload_True(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	assert.True(t, manager.CanAcceptUpload())
}

func TestTusManager_CanAcceptUpload_False(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.AddToQueue("upload-1")

	assert.False(t, manager.CanAcceptUpload())
}

// ==================== IsActiveUpload Tests ====================

func TestTusManager_IsActiveUpload_True(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.AddToQueue("upload-1")

	assert.True(t, manager.IsActiveUpload("upload-1"))
}

func TestTusManager_IsActiveUpload_False(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.AddToQueue("upload-1")

	assert.False(t, manager.IsActiveUpload("upload-2"))
}

// ==================== ReadChunkFromBody Tests ====================

func TestTusManager_ReadChunkFromBody_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	body := []byte("test content")
	expectedSize := int64(len(body))

	reader, err := manager.ReadChunkFromBody(body, expectedSize)

	assert.NoError(t, err)
	assert.NotNil(t, reader)

	data, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, body, data)
}

func TestTusManager_ReadChunkFromBody_SizeMismatch(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	body := []byte("test")
	expectedSize := int64(100)

	_, err := manager.ReadChunkFromBody(body, expectedSize)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak cocok")
}

func TestTusManager_ReadChunkFromBody_Empty(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	body := []byte{}

	_, err := manager.ReadChunkFromBody(body, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kosong")
}

// ==================== ExtractUserIDFromMetadata Tests ====================

func TestTusManager_ExtractUserIDFromMetadata_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := map[string]string{
		"user_id": "123",
	}

	userID, err := manager.ExtractUserIDFromMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, uint(123), userID)
}

func TestTusManager_ExtractUserIDFromMetadata_Missing(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := map[string]string{}

	_, err := manager.ExtractUserIDFromMetadata(metadata)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id tidak ditemukan")
}

func TestTusManager_ExtractUserIDFromMetadata_Invalid(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := map[string]string{
		"user_id": "invalid",
	}

	_, err := manager.ExtractUserIDFromMetadata(metadata)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id tidak valid")
}

func TestTusManager_ExtractUserIDFromMetadata_Zero(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := map[string]string{
		"user_id": "0",
	}

	userID, err := manager.ExtractUserIDFromMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, uint(0), userID)
}

// ==================== GetDefaultTusHeaders Tests ====================

func TestTusManager_GetDefaultTusHeaders_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	headers := manager.GetDefaultTusHeaders()

	assert.NotNil(t, headers)
	assert.Equal(t, manager.config.Upload.TusVersion, headers["Tus-Resumable"])
	assert.Len(t, headers, 1)
}

// ==================== ValidateFileSize Tests ====================

func TestTusManager_ValidateFileSize_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	tests := []struct {
		name     string
		fileSize int64
		maxSize  int64
		wantErr  bool
	}{
		{
			name:     "valid small size",
			fileSize: 1024,
			maxSize:  10 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "valid max size",
			fileSize: 10 * 1024 * 1024,
			maxSize:  10 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "too large",
			fileSize: 20 * 1024 * 1024,
			maxSize:  10 * 1024 * 1024,
			wantErr:  true,
		},
		{
			name:     "zero size",
			fileSize: 0,
			maxSize:  10 * 1024 * 1024,
			wantErr:  true,
		},
		{
			name:     "negative size",
			fileSize: -1,
			maxSize:  10 * 1024 * 1024,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateFileSize(tt.fileSize, tt.maxSize)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== ValidateTusVersion Tests ====================

func TestTusManager_ValidateTusVersion_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	err := manager.ValidateTusVersion("1.0.0")

	assert.NoError(t, err)
}

func TestTusManager_ValidateTusVersion_Invalid(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	versions := []string{"0.0.1", "1.0.1", "2.0.0", ""}

	for _, version := range versions {
		t.Run("version_"+version, func(t *testing.T) {
			err := manager.ValidateTusVersion(version)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "versi TUS protocol tidak didukung")
		})
	}
}

// ==================== ValidateOffset Tests ====================

func TestTusManager_ValidateOffset_Success(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-offset"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	serverOffset, err := manager.ValidateOffset(uploadID, 0)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), serverOffset)
}

func TestTusManager_ValidateOffset_Mismatch(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-offset"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	_, err = manager.ValidateOffset(uploadID, 100)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "offset tidak cocok")
}

func TestTusManager_ValidateOffset_NotFound(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	_, err := manager.ValidateOffset("nonexistent", 0)

	assert.Error(t, err)
}

func TestTusManager_ValidateOffset_AfterChunk(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-offset-chunk"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("test data")
	newOffset, err := store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	serverOffset, err := manager.ValidateOffset(uploadID, newOffset)

	assert.NoError(t, err)
	assert.Equal(t, newOffset, serverOffset)
}

// ==================== ValidateContentType Tests ====================

func TestTusManager_ValidateContentType_Success(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	err := manager.ValidateContentType("application/offset+octet-stream")

	assert.NoError(t, err)
}

func TestTusManager_ValidateContentType_Invalid(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	invalidTypes := []string{
		"application/json",
		"text/plain",
		"application/octet-stream",
		"",
		"multipart/form-data",
	}

	for _, contentType := range invalidTypes {
		t.Run("type_"+contentType, func(t *testing.T) {
			err := manager.ValidateContentType(contentType)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Content-Type harus application/offset+octet-stream")
		})
	}
}

// ==================== Edge Cases and Additional Coverage ====================

func TestTusManager_CheckUploadSlot_MaxConcurrentFromConfig(t *testing.T) {
	cfg := setupTestConfig()
	cfg.Upload.MaxConcurrentProject = 5
	pathResolver := NewPathResolver(cfg)
	store := NewTusStore(pathResolver, cfg.Upload.MaxSize)
	queue := NewTusQueue(cfg.Upload.MaxConcurrentProject)
	fileManager := NewFileManager(cfg)
	manager := NewTusManager(store, queue, fileManager, cfg)

	response := manager.CheckUploadSlot()

	assert.Equal(t, 5, response.MaxConcurrent)
}

func TestTusManager_ParseMetadata_UnicodeCharacters(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := "filename 5Lit5paHLnppcg=="

	parsed, err := manager.ParseMetadata(metadata)

	assert.NoError(t, err)
	// The base64 decodes to "中文.zip" but the actual output depends on encoding
	// Just verify it decoded successfully
	assert.NotEmpty(t, parsed["filename"])
	assert.Contains(t, parsed["filename"], "中文")
}

func TestTusManager_ValidateProjectMetadata_LongestValid(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	longName := string(make([]byte, 255))
	for i := range longName {
		longName = longName[:i] + "a"
	}

	metadata := map[string]string{
		"nama_project": longName,
	}

	err := manager.ValidateProjectMetadata(metadata)

	assert.NoError(t, err)
}

func TestTusManager_ValidateModulMetadata_LongestValid(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	longName := string(make([]byte, 255))
	for i := range longName {
		longName = longName[:i] + "a"
	}

	metadata := map[string]string{
		"nama_file": longName,
		"tipe":      "pdf",
	}

	err := manager.ValidateModulMetadata(metadata)

	assert.NoError(t, err)
}

func TestTusManager_InitiateUpload_MaxSize(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	uploadID := "test-upload-max"
	fileSize := manager.config.Upload.MaxSize

	err := manager.InitiateUpload(uploadID, fileSize, nil)

	assert.NoError(t, err)
}

func TestTusManager_ReadChunkFromBody_LargeChunk(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	body := make([]byte, 1024*1024)
	for i := range body {
		body[i] = byte(i % 256)
	}

	reader, err := manager.ReadChunkFromBody(body, int64(len(body)))

	assert.NoError(t, err)
	assert.NotNil(t, reader)
}

func TestTusManager_ExtractUserIDFromMetadata_LargeNumber(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	metadata := map[string]string{
		"user_id": "4294967295",
	}

	userID, err := manager.ExtractUserIDFromMetadata(metadata)

	assert.NoError(t, err)
	assert.Equal(t, uint(4294967295), userID)
}

func TestTusManager_HandleChunk_EmptyChunk(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-empty"
	fileSize := int64(1024)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte{}
	newOffset, err := manager.HandleChunk(uploadID, 0, bytes.NewReader(chunk))

	assert.NoError(t, err)
	assert.Equal(t, int64(0), newOffset)
}

func TestTusManager_GetUploadProgress_Complete(t *testing.T) {
	manager, store := setupTestTusManager(t)

	uploadID := "test-upload-progress-complete"
	fileSize := int64(24)

	err := store.InitiateUpload(uploadID, fileSize)
	require.NoError(t, err)

	chunk := []byte("test chunk content test!")
	_, err = store.WriteChunk(uploadID, 0, bytes.NewReader(chunk))
	require.NoError(t, err)

	progress, err := manager.GetUploadProgress(uploadID)

	assert.NoError(t, err)
	assert.Equal(t, float64(100), progress)
}

func TestTusManager_AddToQueue_Duplicate(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	// Add first upload as active
	manager.AddToQueue("upload-1")

	// Add second upload - goes to queue
	manager.AddToQueue("upload-2")

	// Try to add duplicate to queue - should be ignored
	manager.AddToQueue("upload-2")

	assert.True(t, manager.queue.HasActiveUpload())
	assert.Equal(t, 1, manager.queue.GetQueueLength())
}

func TestTusManager_RemoveFromQueue_QueuedUpload(t *testing.T) {
	manager, _ := setupTestTusManager(t)

	manager.AddToQueue("upload-1")
	manager.AddToQueue("upload-2")
	err := manager.RemoveFromQueue("upload-2")

	assert.NoError(t, err)
	assert.Equal(t, 0, manager.queue.GetQueueLength())
}
