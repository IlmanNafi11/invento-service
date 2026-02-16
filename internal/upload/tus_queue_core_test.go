package upload_test

import (
	"invento-service/internal/upload"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== TusQueue Tests ====================

func TestNewTusQueue_Success(t *testing.T) {
	t.Parallel()
	maxConcurrent := 3
	queue := upload.NewTusQueue(maxConcurrent)

	assert.NotNil(t, queue)
	assert.Equal(t, 0, queue.GetQueueLength())
	assert.False(t, queue.HasActiveUpload())
}

func TestTusQueue_Add_Success(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	assert.True(t, queue.HasActiveUpload())
	activeUploads := queue.GetActiveUploads()
	assert.Len(t, activeUploads, 1)
	assert.Contains(t, activeUploads, "upload-1")
}

func TestTusQueue_Add_Queued(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so second upload goes to queue

	queue.Add("upload-1")
	queue.Add("upload-2")

	activeUploads := queue.GetActiveUploads()
	assert.Contains(t, activeUploads, "upload-1")
	assert.Equal(t, 1, queue.GetQueueLength())
}

func TestTusQueue_Add_Duplicate(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")
	err := queue.Remove("upload-1")

	assert.NoError(t, err)
	assert.False(t, queue.HasActiveUpload())
}

func TestTusQueue_Remove_QueuedUpload(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so second upload goes to queue

	queue.Add("upload-1")
	queue.Add("upload-2")
	err := queue.Remove("upload-2")

	assert.NoError(t, err)
	assert.Equal(t, 0, queue.GetQueueLength())
}

func TestTusQueue_Remove_NotFound(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	err := queue.Remove("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

func TestTusQueue_GetQueuePosition_Active(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	position := queue.GetQueuePosition("upload-1")
	assert.Equal(t, 0, position)
}

func TestTusQueue_GetQueuePosition_Queued(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so second upload goes to queue

	queue.Add("upload-1")
	queue.Add("upload-2")

	position := queue.GetQueuePosition("upload-2")
	assert.Equal(t, 1, position)
}

func TestTusQueue_GetQueuePosition_NotFound(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	position := queue.GetQueuePosition("nonexistent")
	assert.Equal(t, -1, position)
}

func TestTusQueue_FinishUpload_Success(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")
	queue.FinishUpload("upload-1")

	assert.False(t, queue.HasActiveUpload())
	activeUploads := queue.GetActiveUploads()
	assert.Len(t, activeUploads, 0)
}

func TestTusQueue_Clear_Success(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")
	queue.Add("upload-2")
	queue.Clear()

	assert.False(t, queue.HasActiveUpload())
	assert.Equal(t, 0, queue.GetQueueLength())
}

func TestTusQueue_CanAcceptUpload_True(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	assert.True(t, queue.CanAcceptUpload())
}

func TestTusQueue_CanAcceptUpload_False(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so it fills up with one upload

	queue.Add("upload-1")

	assert.False(t, queue.CanAcceptUpload())
}

func TestTusQueue_IsActiveUpload_True(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	assert.True(t, queue.IsActiveUpload("upload-1"))
}

func TestTusQueue_IsActiveUpload_False(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(3)

	queue.Add("upload-1")

	assert.False(t, queue.IsActiveUpload("upload-2"))
}

func TestTusQueue_GetCurrentQueue_Success(t *testing.T) {
	t.Parallel()
	queue := upload.NewTusQueue(1) // maxConcurrent=1 so uploads go to queue

	queue.Add("upload-1")
	queue.Add("upload-2")
	queue.Add("upload-3")

	currentQueue := queue.GetCurrentQueue()
	assert.Len(t, currentQueue, 2)
	assert.Contains(t, currentQueue, "upload-2")
	assert.Contains(t, currentQueue, "upload-3")
}
