package helper_test

import (
	"fiber-boiler-plate/internal/helper"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTusQueue(t *testing.T) {
	queue := helper.NewTusQueue(3)

	assert.NotNil(t, queue)
}

func TestTusQueue_Add(t *testing.T) {
	queue := helper.NewTusQueue(2)

	// Add first upload - should become active
	queue.Add("upload1")
	assert.Equal(t, "upload1", queue.GetActiveUpload())
	assert.Equal(t, 0, queue.GetQueueLength())

	// Add second upload - should be queued
	queue.Add("upload2")
	assert.Equal(t, "upload1", queue.GetActiveUpload())
	assert.Equal(t, 1, queue.GetQueueLength())

	// Add third upload - should be queued
	queue.Add("upload3")
	assert.Equal(t, "upload1", queue.GetActiveUpload())
	assert.Equal(t, 2, queue.GetQueueLength())
}

func TestTusQueue_Add_Duplicate(t *testing.T) {
	queue := helper.NewTusQueue(2)

	// Add first upload - becomes active
	queue.Add("upload1")

	// Add different upload - goes to queue
	queue.Add("upload2")

	// Try to add upload2 again - should be ignored (already in queue)
	queue.Add("upload2")

	// Queue length should still be 1 (only upload2 is queued)
	assert.Equal(t, 1, queue.GetQueueLength())
}

func TestTusQueue_GetActiveUpload(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Initially no active upload
	assert.Empty(t, queue.GetActiveUpload())

	// Add upload
	queue.Add("upload1")
	assert.Equal(t, "upload1", queue.GetActiveUpload())

	// Finish active upload
	queue.FinishActiveUpload()
	assert.Empty(t, queue.GetActiveUpload())
}

func TestTusQueue_HasActiveUpload(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Initially no active upload
	assert.False(t, queue.HasActiveUpload())

	// Add upload
	queue.Add("upload1")
	assert.True(t, queue.HasActiveUpload())

	// Finish active upload
	queue.FinishActiveUpload()
	assert.False(t, queue.HasActiveUpload())
}

func TestTusQueue_GetQueuePosition(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Add uploads
	queue.Add("upload1") // active
	queue.Add("upload2") // position 1
	queue.Add("upload3") // position 2

	// Check positions
	assert.Equal(t, 0, queue.GetQueuePosition("upload1"))
	assert.Equal(t, 1, queue.GetQueuePosition("upload2"))
	assert.Equal(t, 2, queue.GetQueuePosition("upload3"))

	// Non-existent upload
	assert.Equal(t, -1, queue.GetQueuePosition("upload999"))
}

func TestTusQueue_GetQueueLength(t *testing.T) {
	queue := helper.NewTusQueue(2)

	// Initially empty
	assert.Equal(t, 0, queue.GetQueueLength())

	// Add first upload (active, not queued)
	queue.Add("upload1")
	assert.Equal(t, 0, queue.GetQueueLength())

	// Add more uploads (queued)
	queue.Add("upload2")
	queue.Add("upload3")
	assert.Equal(t, 2, queue.GetQueueLength())
}

func TestTusQueue_Remove(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Add uploads
	queue.Add("upload1") // active
	queue.Add("upload2") // queued

	// Remove queued upload
	err := queue.Remove("upload2")
	assert.NoError(t, err)
	assert.Equal(t, 0, queue.GetQueueLength())
	assert.Equal(t, -1, queue.GetQueuePosition("upload2"))

	// Remove active upload
	err = queue.Remove("upload1")
	assert.NoError(t, err)
	assert.Empty(t, queue.GetActiveUpload())

	// Try to remove non-existent upload
	err = queue.Remove("upload999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload tidak ditemukan dalam antrian")
}

func TestTusQueue_FinishActiveUpload(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Add upload
	queue.Add("upload1")
	assert.NotEmpty(t, queue.GetActiveUpload())

	// Finish active upload
	queue.FinishActiveUpload()
	assert.Empty(t, queue.GetActiveUpload())

	// Next upload should become active (if any)
	queue.Add("upload2")
	queue.Add("upload3")
	assert.Equal(t, "upload2", queue.GetActiveUpload())
}

func TestTusQueue_Clear(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Add uploads
	queue.Add("upload1") // active
	queue.Add("upload2") // queued
	queue.Add("upload3") // queued

	// Clear queue
	queue.Clear()

	assert.Empty(t, queue.GetActiveUpload())
	assert.Equal(t, 0, queue.GetQueueLength())
	assert.Equal(t, -1, queue.GetQueuePosition("upload1"))
	assert.Equal(t, -1, queue.GetQueuePosition("upload2"))
}

func TestTusQueue_CanAcceptUpload(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Initially can accept
	assert.True(t, queue.CanAcceptUpload())

	// Add upload
	queue.Add("upload1")
	assert.False(t, queue.CanAcceptUpload())

	// Finish active upload
	queue.FinishActiveUpload()
	assert.True(t, queue.CanAcceptUpload())
}

func TestTusQueue_IsActiveUpload(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// No active upload
	assert.False(t, queue.IsActiveUpload("upload1"))

	// Add upload
	queue.Add("upload1")
	assert.True(t, queue.IsActiveUpload("upload1"))
	assert.False(t, queue.IsActiveUpload("upload2"))

	// Finish active upload
	queue.FinishActiveUpload()
	assert.False(t, queue.IsActiveUpload("upload1"))
}

func TestTusQueue_GetCurrentQueue(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Initially empty
	currentQueue := queue.GetCurrentQueue()
	assert.Equal(t, 0, len(currentQueue))

	// Add uploads
	queue.Add("upload1") // active
	queue.Add("upload2") // queued
	queue.Add("upload3") // queued

	currentQueue = queue.GetCurrentQueue()
	assert.Equal(t, 2, len(currentQueue))
	assert.Equal(t, "upload2", currentQueue[0])
	assert.Equal(t, "upload3", currentQueue[1])

	// Verify it's a copy (modifying shouldn't affect original)
	currentQueue[0] = "modified"
	assert.Equal(t, "upload2", queue.GetCurrentQueue()[0])
}

func TestTusQueue_Workflow(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Add first upload
	queue.Add("upload1")
	assert.True(t, queue.HasActiveUpload())
	assert.Equal(t, 0, queue.GetQueuePosition("upload1"))

	// Queue second upload
	queue.Add("upload2")
	assert.Equal(t, 1, queue.GetQueuePosition("upload2"))

	// Finish first upload
	queue.FinishActiveUpload()
	assert.False(t, queue.HasActiveUpload())

	// Check if can accept new upload
	assert.True(t, queue.CanAcceptUpload())

	// Add third upload - should become active
	queue.Add("upload3")
	assert.True(t, queue.HasActiveUpload())
	assert.Equal(t, "upload3", queue.GetActiveUpload())

	// Second upload should still be in queue
	assert.Equal(t, 1, queue.GetQueuePosition("upload2"))
}

func TestTusQueue_ConcurrentAccess(t *testing.T) {
	queue := helper.NewTusQueue(10)
	var wg sync.WaitGroup

	// Simulate concurrent additions
	numGoroutines := 100
	uploadsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < uploadsPerGoroutine; j++ {
				uploadID, _ := helper.GenerateRandomString(10)
				queue.Add(uploadID)
			}
		}(i)
	}

	wg.Wait()

	// Verify queue state is consistent
	// We should have one active upload and the rest queued
	totalUploads := numGoroutines * uploadsPerGoroutine
	queueLength := queue.GetQueueLength()
	activeUpload := queue.GetActiveUpload()

	// Total should be active + queued
	activeCount := 0
	if activeUpload != "" {
		activeCount = 1
	}
	assert.Equal(t, totalUploads, queueLength + activeCount)
}

func TestTusQueue_RemoveAndNext(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Build a queue
	queue.Add("upload1") // active
	queue.Add("upload2") // queued
	queue.Add("upload3") // queued

	// Remove active upload
	err := queue.Remove("upload1")
	assert.NoError(t, err)

	// No automatic promotion to active in this implementation
	// Need to check if next upload can be added
	assert.True(t, queue.CanAcceptUpload())

	// Add new upload - should become active
	queue.Add("upload4")
	assert.Equal(t, "upload4", queue.GetActiveUpload())
}

func TestTusQueue_EmptyQueueOperations(t *testing.T) {
	queue := helper.NewTusQueue(1)

	// Operations on empty queue
	assert.False(t, queue.HasActiveUpload())
	assert.Empty(t, queue.GetActiveUpload())
	assert.Equal(t, 0, queue.GetQueueLength())
	assert.True(t, queue.CanAcceptUpload())
	assert.Equal(t, -1, queue.GetQueuePosition("nonexistent"))

	// Finish when no active upload - should not panic
	queue.FinishActiveUpload()

	// Clear empty queue - should not panic
	queue.Clear()
}
