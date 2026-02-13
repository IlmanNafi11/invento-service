package helper

import (
	"errors"
	"sync"
)

type TusQueue struct {
	queue         []string
	activeUploads map[string]bool
	maxConcurrent int
	mutex         sync.RWMutex
}

func NewTusQueue(maxConcurrent int) *TusQueue {
	return &TusQueue{
		queue:         make([]string, 0),
		activeUploads: make(map[string]bool),
		maxConcurrent: maxConcurrent,
	}
}

func (tq *TusQueue) Add(uploadID string) {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if tq.activeUploads[uploadID] {
		return
	}

	for _, id := range tq.queue {
		if id == uploadID {
			return
		}
	}

	if len(tq.activeUploads) < tq.maxConcurrent {
		tq.activeUploads[uploadID] = true
		return
	}

	tq.queue = append(tq.queue, uploadID)
}

func (tq *TusQueue) GetActiveUploads() []string {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	activeUploads := make([]string, 0, len(tq.activeUploads))
	for id := range tq.activeUploads {
		activeUploads = append(activeUploads, id)
	}

	return activeUploads
}

func (tq *TusQueue) HasActiveUpload() bool {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return len(tq.activeUploads) > 0
}

func (tq *TusQueue) GetQueuePosition(uploadID string) int {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	if tq.activeUploads[uploadID] {
		return 0
	}

	for i, id := range tq.queue {
		if id == uploadID {
			return i + 1
		}
	}

	return -1
}

func (tq *TusQueue) GetQueueLength() int {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return len(tq.queue)
}

func (tq *TusQueue) Remove(uploadID string) error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if tq.activeUploads[uploadID] {
		delete(tq.activeUploads, uploadID)
		return nil
	}

	for i, id := range tq.queue {
		if id == uploadID {
			tq.queue = append(tq.queue[:i], tq.queue[i+1:]...)
			return nil
		}
	}

	return errors.New("upload tidak ditemukan dalam antrian")
}

func (tq *TusQueue) FinishUpload(uploadID string) string {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if tq.activeUploads[uploadID] {
		delete(tq.activeUploads, uploadID)

		if len(tq.queue) == 0 {
			return ""
		}

		nextUploadID := tq.queue[0]
		tq.queue = tq.queue[1:]
		tq.activeUploads[nextUploadID] = true

		return nextUploadID
	}

	for i, id := range tq.queue {
		if id != uploadID {
			continue
		}

		tq.queue = append(tq.queue[:i], tq.queue[i+1:]...)
		return ""
	}

	return ""
}

func (tq *TusQueue) Clear() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	tq.queue = make([]string, 0)
	tq.activeUploads = make(map[string]bool)
}

func (tq *TusQueue) CanAcceptUpload() bool {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return len(tq.activeUploads) < tq.maxConcurrent
}

func (tq *TusQueue) IsActiveUpload(uploadID string) bool {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return tq.activeUploads[uploadID]
}

func (tq *TusQueue) GetCurrentQueue() []string {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	queueCopy := make([]string, len(tq.queue))
	copy(queueCopy, tq.queue)

	return queueCopy
}

func (tq *TusQueue) LoadFromDB(activeIDs []string) {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	tq.activeUploads = make(map[string]bool)
	tq.queue = make([]string, 0)

	seen := make(map[string]bool)
	for _, uploadID := range activeIDs {
		if seen[uploadID] {
			continue
		}
		seen[uploadID] = true

		if len(tq.activeUploads) < tq.maxConcurrent {
			tq.activeUploads[uploadID] = true
			continue
		}

		tq.queue = append(tq.queue, uploadID)
	}
}

func (tq *TusQueue) GetActiveCount() int {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return len(tq.activeUploads)
}
