package helper

import (
	"bytes"
	"encoding/base64"
	"path/filepath"
	"testing"

	"fiber-boiler-plate/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTusManager(t *testing.T) (*TusManager, *TusStore, *TusQueue) {
	t.Helper()

	base := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
		Upload: config.UploadConfig{
			MaxSize:              1024 * 1024,
			MaxConcurrentProject: 2,
			TusVersion:           "1.0.0",
			PathDevelopment:      filepath.Join(base, "uploads"),
			TempPathDevelopment:  filepath.Join(base, "temp"),
		},
	}

	resolver := NewPathResolver(cfg)
	store := NewTusStore(resolver, cfg.Upload.MaxSize)
	queue := NewTusQueue(cfg.Upload.MaxConcurrentProject)
	manager := NewTusManager(store, queue, nil, cfg)

	return manager, store, queue
}

func TestTusManager_CheckUploadSlot_ReturnsCorrectAvailability(t *testing.T) {
	manager, _, queue := newTestTusManager(t)

	resp := manager.CheckUploadSlot()
	assert.True(t, resp.Available)
	assert.False(t, resp.ActiveUpload)
	assert.Equal(t, 0, resp.QueueLength)
	assert.Equal(t, 2, resp.MaxConcurrent)

	queue.Add("u1")
	queue.Add("u2")
	queue.Add("u3")

	resp = manager.CheckUploadSlot()
	assert.False(t, resp.Available)
	assert.True(t, resp.ActiveUpload)
	assert.Equal(t, 1, resp.QueueLength)
	assert.Contains(t, resp.Message, "tidak tersedia")
}

func TestTusManager_ResetUploadQueue_TerminatesActiveAndClearsQueue(t *testing.T) {
	manager, store, queue := newTestTusManager(t)

	require.NoError(t, store.NewUpload(TusFileInfo{ID: "u1", Size: 10, Metadata: map[string]string{}}))
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "u2", Size: 10, Metadata: map[string]string{}}))
	queue.Add("u1")
	queue.Add("u2")

	err := manager.ResetUploadQueue()
	require.NoError(t, err)
	assert.Equal(t, 0, queue.GetActiveCount())
	assert.Equal(t, 0, queue.GetQueueLength())

	_, err = store.GetInfo("u1")
	assert.Error(t, err)
	_, err = store.GetInfo("u2")
	assert.Error(t, err)
}

func TestTusManager_ParseMetadata_DecodesBase64Pairs(t *testing.T) {
	name := base64.StdEncoding.EncodeToString([]byte("doc.zip"))
	typev := base64.StdEncoding.EncodeToString([]byte("application/zip"))

	data := ParseTusMetadata("filename " + name + ",content_type " + typev)
	assert.Equal(t, "doc.zip", data["filename"])
	assert.Equal(t, "application/zip", data["content_type"])
}

func TestTusManager_ParseMetadata_InvalidBase64DefaultsEmptyValue(t *testing.T) {
	data := ParseTusMetadata("filename !!!")
	assert.Contains(t, data, "filename")
	assert.Equal(t, "", data["filename"])
}

func TestTusManager_ValidateModulMetadata_ValidatesFields(t *testing.T) {
	manager, _, _ := newTestTusManager(t)

	err := manager.ValidateModulMetadata(map[string]string{"tipe": "pdf"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nama_file wajib diisi")

	err = manager.ValidateModulMetadata(map[string]string{"nama_file": "doc.pdf"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tipe wajib diisi")

	err = manager.ValidateModulMetadata(map[string]string{"nama_file": "doc.pdf", "tipe": "txt"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tipe file tidak valid")

	err = manager.ValidateModulMetadata(map[string]string{"nama_file": "doc.pdf", "tipe": "pdf", "semester": "0"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "semester harus antara 1-8")

	err = manager.ValidateModulMetadata(map[string]string{"nama_file": "doc.pdf", "tipe": "pdf", "semester": "7"})
	assert.NoError(t, err)
}

func TestTusManager_InitiateUpload_ValidatesSizeAndCreatesUpload(t *testing.T) {
	manager, store, _ := newTestTusManager(t)

	err := manager.InitiateUpload("too-big", manager.config.Upload.MaxSize+1, map[string]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "melebihi batas maksimal")

	err = manager.InitiateUpload("invalid", 0, map[string]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ukuran file tidak valid")

	err = manager.InitiateUpload("ok", 50, map[string]string{"name": "x"})
	require.NoError(t, err)

	info, err := store.GetInfo("ok")
	require.NoError(t, err)
	assert.Equal(t, int64(50), info.Size)
	assert.Equal(t, "x", info.Metadata["name"])
}

func TestTusManager_HandleChunk_DelegatesToStoreWriteChunk(t *testing.T) {
	manager, store, _ := newTestTusManager(t)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "u3", Size: 20, Metadata: map[string]string{}}))

	offset, err := manager.HandleChunk("u3", 0, bytes.NewBufferString("hello"))
	require.NoError(t, err)
	assert.Equal(t, int64(5), offset)

	_, err = manager.HandleChunk("missing", 0, bytes.NewBufferString("a"))
	assert.Error(t, err)
}
