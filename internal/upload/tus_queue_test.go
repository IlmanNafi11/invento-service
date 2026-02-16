package upload

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTusQueue_NewTusQueue_InitializesState(t *testing.T) {
	queue := NewTusQueue(3)

	require.NotNil(t, queue)
	assert.Equal(t, 3, queue.maxConcurrent)
	assert.Empty(t, queue.queue)
	assert.Empty(t, queue.activeUploads)
}

func TestTusQueue_Add_ActiveQueueAndDedupBehavior(t *testing.T) {
	queue := NewTusQueue(2)

	queue.Add("u1")
	queue.Add("u2")
	queue.Add("u3")
	queue.Add("u4")
	queue.Add("u3")
	queue.Add("u1")

	active := queue.GetActiveUploads()
	assert.Len(t, active, 2)
	assert.ElementsMatch(t, []string{"u1", "u2"}, active)
	assert.Equal(t, []string{"u3", "u4"}, queue.GetCurrentQueue())
}

func TestTusQueue_GetActiveUploads_ReturnsAllActiveUploads(t *testing.T) {
	queue := NewTusQueue(3)
	queue.Add("u1")
	queue.Add("u2")

	assert.ElementsMatch(t, []string{"u1", "u2"}, queue.GetActiveUploads())
}

func TestTusQueue_HasActiveUpload_TrueAndFalseCases(t *testing.T) {
	queue := NewTusQueue(1)
	assert.False(t, queue.HasActiveUpload())

	queue.Add("u1")
	assert.True(t, queue.HasActiveUpload())

	require.NoError(t, queue.Remove("u1"))
	assert.False(t, queue.HasActiveUpload())
}

func TestTusQueue_GetQueueLength_ReturnsCorrectCount(t *testing.T) {
	queue := NewTusQueue(1)
	queue.Add("u1")
	queue.Add("u2")
	queue.Add("u3")

	assert.Equal(t, 2, queue.GetQueueLength())
}

func TestTusQueue_GetQueuePosition_ReturnsActiveQueuedAndMissingPositions(t *testing.T) {
	queue := NewTusQueue(1)
	queue.Add("u1")
	queue.Add("u2")
	queue.Add("u3")

	assert.Equal(t, 0, queue.GetQueuePosition("u1"))
	assert.Equal(t, 1, queue.GetQueuePosition("u2"))
	assert.Equal(t, 2, queue.GetQueuePosition("u3"))
	assert.Equal(t, -1, queue.GetQueuePosition("missing"))
}

func TestTusQueue_Remove_RemovesFromActiveAndQueue(t *testing.T) {
	queue := NewTusQueue(1)
	queue.Add("u1")
	queue.Add("u2")

	require.NoError(t, queue.Remove("u1"))
	assert.False(t, queue.IsActiveUpload("u1"))

	require.NoError(t, queue.Remove("u2"))
	assert.Equal(t, 0, queue.GetQueueLength())

	err := queue.Remove("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")
}

func TestTusQueue_FinishUpload_PromotesNextUpload(t *testing.T) {
	queue := NewTusQueue(1)
	queue.Add("u1")
	queue.Add("u2")
	queue.Add("u3")

	promoted := queue.FinishUpload("u1")
	assert.Equal(t, "u2", promoted)
	assert.True(t, queue.IsActiveUpload("u2"))
	assert.Equal(t, 1, queue.GetQueueLength())

	assert.Equal(t, "", queue.FinishUpload("not-active"))

	promoted = queue.FinishUpload("u2")
	assert.Equal(t, "u3", promoted)

	assert.Equal(t, "", queue.FinishUpload("u3"))
	assert.False(t, queue.HasActiveUpload())
}

func TestTusQueue_CanAcceptUpload_RespectsMaxConcurrent(t *testing.T) {
	queue := NewTusQueue(2)
	assert.True(t, queue.CanAcceptUpload())

	queue.Add("u1")
	assert.True(t, queue.CanAcceptUpload())

	queue.Add("u2")
	assert.False(t, queue.CanAcceptUpload())
}

func TestTusQueue_IsActiveUpload_ChecksActiveMap(t *testing.T) {
	queue := NewTusQueue(1)
	queue.Add("u1")

	assert.True(t, queue.IsActiveUpload("u1"))
	assert.False(t, queue.IsActiveUpload("u2"))
}

func TestTusQueue_GetCurrentQueue_ReturnsCopy(t *testing.T) {
	queue := NewTusQueue(1)
	queue.Add("u1")
	queue.Add("u2")

	current := queue.GetCurrentQueue()
	require.Equal(t, []string{"u2"}, current)

	current[0] = "changed"
	assert.Equal(t, []string{"u2"}, queue.GetCurrentQueue())
}

func TestTusQueue_Clear_ResetsQueueAndActiveUploads(t *testing.T) {
	queue := NewTusQueue(1)
	queue.Add("u1")
	queue.Add("u2")

	queue.Clear()

	assert.Empty(t, queue.GetActiveUploads())
	assert.Empty(t, queue.GetCurrentQueue())
	assert.Equal(t, 0, queue.GetQueueLength())
	assert.Equal(t, 0, queue.GetActiveCount())
}

func TestTusQueue_LoadFromDB_LoadsActiveThenQueueDeduplicated(t *testing.T) {
	queue := NewTusQueue(2)
	queue.LoadFromDB([]string{"u1", "u1", "u2", "u3", "u4"})

	assert.ElementsMatch(t, []string{"u1", "u2"}, queue.GetActiveUploads())
	assert.Equal(t, []string{"u3", "u4"}, queue.GetCurrentQueue())
	assert.Equal(t, 2, queue.GetActiveCount())
}

func TestTusQueue_GetActiveCount_ReturnsCorrectCount(t *testing.T) {
	queue := NewTusQueue(3)
	assert.Equal(t, 0, queue.GetActiveCount())

	queue.Add("u1")
	queue.Add("u2")
	assert.Equal(t, 2, queue.GetActiveCount())
}

func TestTusQueue_ConcurrentAccess_AddRemoveFinishUploadSafely(t *testing.T) {
	queue := NewTusQueue(10)

	ids := make([]string, 120)
	for i := range ids {
		ids[i] = fmt.Sprintf("upload-%03d", i)
	}

	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)
		go func(uploadID string) {
			defer wg.Done()
			queue.Add(uploadID)
		}(id)
	}

	for i := 0; i < 80; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = queue.Remove(ids[idx])
		}(i)
	}

	for i := 80; i < 120; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			queue.FinishUpload(ids[idx])
		}(i)
	}

	wg.Wait()

	active := queue.GetActiveUploads()
	queued := queue.GetCurrentQueue()

	seen := map[string]bool{}
	for _, id := range active {
		assert.False(t, seen[id])
		seen[id] = true
	}
	for _, id := range queued {
		assert.False(t, seen[id])
		seen[id] = true
	}

	assert.LessOrEqual(t, len(active), 10)
	assert.Equal(t, len(queued), queue.GetQueueLength())
	assert.Equal(t, len(active), queue.GetActiveCount())
}
