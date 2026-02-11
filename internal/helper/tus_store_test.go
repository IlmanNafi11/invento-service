package helper

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTusStore(t *testing.T, maxSize int64) *TusStore {
	t.Helper()
	base := t.TempDir()
	resolver := &PathResolver{
		env:                 "development",
		pathDevelopment:     filepath.Join(base, "uploads"),
		tempPathDevelopment: filepath.Join(base, "temp"),
	}

	return NewTusStore(resolver, maxSize)
}

func TestTusStore_NewTusStore_InitializesConfigAndLocks(t *testing.T) {
	store := newTestTusStore(t, 1024)

	require.NotNil(t, store)
	assert.Equal(t, int64(1024), store.maxFileSize)
	assert.NotNil(t, store.pathResolver)
	assert.NotNil(t, store.locks)
	assert.Equal(t, 0, store.GetLockCount())
}

func TestTusStore_NewUpload_CreatesDirectoryFileAndInfo(t *testing.T) {
	store := newTestTusStore(t, 1024)
	info := TusFileInfo{ID: "up1", Size: 128, Offset: 0, Metadata: map[string]string{"name": "a.zip"}}

	err := store.NewUpload(info)
	require.NoError(t, err)

	_, err = os.Stat(store.pathResolver.GetUploadPath("up1"))
	assert.NoError(t, err)

	stat, err := os.Stat(store.pathResolver.GetUploadFilePath("up1"))
	require.NoError(t, err)
	assert.Equal(t, int64(128), stat.Size())

	readInfo, err := store.GetInfo("up1")
	require.NoError(t, err)
	assert.Equal(t, info.ID, readInfo.ID)
	assert.Equal(t, info.Size, readInfo.Size)
	assert.Equal(t, info.Metadata["name"], readInfo.Metadata["name"])
}

func TestTusStore_NewUpload_RejectsOversizedFile(t *testing.T) {
	store := newTestTusStore(t, 100)

	err := store.NewUpload(TusFileInfo{ID: "big", Size: 101, Metadata: map[string]string{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
}

func TestTusStore_WriteChunk_WritesAtOffsetAndUpdatesInfo(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "up2", Size: 32, Metadata: map[string]string{}}))

	offset, err := store.WriteChunk("up2", 5, bytes.NewBufferString("hello"))
	require.NoError(t, err)
	assert.Equal(t, int64(10), offset)

	info, err := store.GetInfo("up2")
	require.NoError(t, err)
	assert.Equal(t, int64(10), info.Offset)

	data, err := os.ReadFile(store.pathResolver.GetUploadFilePath("up2"))
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data[5:10]))
}

func TestTusStore_WriteChunk_ConcurrentWritesLockingWorks(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "up3", Size: 64, Metadata: map[string]string{}}))

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	offCh := make(chan int64, 2)

	write := func() {
		defer wg.Done()
		off, err := store.WriteChunk("up3", 0, bytes.NewBufferString("abcde"))
		errCh <- err
		offCh <- off
	}

	wg.Add(2)
	go write()
	go write()
	wg.Wait()
	close(errCh)
	close(offCh)

	for err := range errCh {
		assert.NoError(t, err)
	}
	for off := range offCh {
		assert.Equal(t, int64(5), off)
	}

	info, err := store.GetInfo("up3")
	require.NoError(t, err)
	assert.Equal(t, int64(5), info.Offset)
	assert.Equal(t, 1, store.GetLockCount())
}

func TestTusStore_WriteChunk_InvalidUploadIDReturnsError(t *testing.T) {
	store := newTestTusStore(t, 1024)

	_, err := store.WriteChunk("missing", 0, bytes.NewBufferString("x"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gagal membuka file")
}

func TestTusStore_GetInfo_ReturnsInfoAndErrors(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "up4", Size: 16, Metadata: map[string]string{}}))

	info, err := store.GetInfo("up4")
	require.NoError(t, err)
	assert.Equal(t, "up4", info.ID)

	_, err = store.GetInfo("nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")
}

func TestTusStore_GetInfo_CorruptInfoFileReturnsError(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "badinfo", Size: 16, Metadata: map[string]string{}}))

	err := os.WriteFile(store.pathResolver.GetUploadInfoPath("badinfo"), []byte("{"), 0o644)
	require.NoError(t, err)

	_, err = store.GetInfo("badinfo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gagal parse info")
}

func TestTusStore_FinalizeUpload_MovesToDestinationAndCleansTemp(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "up5", Size: 8, Metadata: map[string]string{}}))
	_, err := store.WriteChunk("up5", 0, bytes.NewBufferString("12345678"))
	require.NoError(t, err)

	finalPath := filepath.Join(t.TempDir(), "final", "result.zip")
	err = store.FinalizeUpload("up5", finalPath)
	require.NoError(t, err)

	_, err = os.Stat(finalPath)
	assert.NoError(t, err)
	_, err = os.Stat(store.pathResolver.GetUploadPath("up5"))
	assert.True(t, os.IsNotExist(err))
	assert.Equal(t, 0, store.GetLockCount())
}

func TestTusStore_FinalizeUpload_MissingTemporaryFileReturnsError(t *testing.T) {
	store := newTestTusStore(t, 1024)

	err := store.FinalizeUpload("missing", filepath.Join(t.TempDir(), "a.zip"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file temporary tidak ditemukan")
}

func TestTusStore_Terminate_RemovesUploadDirectory(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "up6", Size: 8, Metadata: map[string]string{}}))

	err := store.Terminate("up6")
	require.NoError(t, err)

	_, err = os.Stat(store.pathResolver.GetUploadPath("up6"))
	assert.True(t, os.IsNotExist(err))
	assert.Equal(t, 0, store.GetLockCount())
}

func TestTusStore_GetProgressGetOffsetAndIsComplete_ReturnExpectedValues(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "up7", Size: 10, Metadata: map[string]string{}}))

	progress, err := store.GetProgress("up7")
	require.NoError(t, err)
	assert.Equal(t, 0.0, progress)

	offset, err := store.GetOffset("up7")
	require.NoError(t, err)
	assert.Equal(t, int64(0), offset)

	complete, err := store.IsComplete("up7")
	require.NoError(t, err)
	assert.False(t, complete)

	_, err = store.WriteChunk("up7", 0, bytes.NewBufferString("1234567890"))
	require.NoError(t, err)

	progress, err = store.GetProgress("up7")
	require.NoError(t, err)
	assert.Equal(t, 100.0, progress)

	complete, err = store.IsComplete("up7")
	require.NoError(t, err)
	assert.True(t, complete)
}

func TestTusStore_UpdateMetadata_UpdatesAndPersists(t *testing.T) {
	store := newTestTusStore(t, 1024)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "up8", Size: 10, Metadata: map[string]string{"a": "1"}}))

	err := store.UpdateMetadata("up8", map[string]string{"b": "2", "a": "3"})
	require.NoError(t, err)

	info, err := store.GetInfo("up8")
	require.NoError(t, err)
	assert.Equal(t, "3", info.Metadata["a"])
	assert.Equal(t, "2", info.Metadata["b"])
}

func TestTusStore_UpdateMetadata_InvalidUploadIDReturnsError(t *testing.T) {
	store := newTestTusStore(t, 1024)

	err := store.UpdateMetadata("missing", map[string]string{"a": "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")
}

func TestTusStore_CleanupStaleLocks_RemovesOnlyOldLocks(t *testing.T) {
	store := newTestTusStore(t, 1024)

	_ = store.getLock("old")
	_ = store.getLock("new")

	store.locksMutex.Lock()
	store.locks["old"].lastAccess = time.Now().Add(-2 * time.Hour)
	store.locks["new"].lastAccess = time.Now()
	store.locksMutex.Unlock()

	removed := store.CleanupStaleLocks(30 * time.Minute)
	assert.Equal(t, 1, removed)
	assert.Equal(t, 1, store.GetLockCount())
}

func TestTusStore_GetLockCount_ReturnsCurrentCount(t *testing.T) {
	store := newTestTusStore(t, 1024)
	assert.Equal(t, 0, store.GetLockCount())

	for i := 0; i < 3; i++ {
		_ = store.getLock(fmt.Sprintf("id-%d", i))
	}

	assert.Equal(t, 3, store.GetLockCount())
}
