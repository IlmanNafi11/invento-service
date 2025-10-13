package helper

import (
	"errors"
	"sync"
)

type TusQueue struct {
	queue         []string
	activeUpload  string
	maxConcurrent int
	mutex         sync.RWMutex
}

func NewTusQueue(maxConcurrent int) *TusQueue {
	return &TusQueue{
		queue:         make([]string, 0),
		activeUpload:  "",
		maxConcurrent: maxConcurrent,
	}
}

func (tq *TusQueue) Enqueue(uploadID string) error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	for _, id := range tq.queue {
		if id == uploadID {
			return errors.New("upload sudah ada dalam antrian")
		}
	}

	if tq.activeUpload == uploadID {
		return errors.New("upload sedang diproses")
	}

	tq.queue = append(tq.queue, uploadID)
	return nil
}

func (tq *TusQueue) Dequeue() (string, error) {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if len(tq.queue) == 0 {
		return "", errors.New("antrian kosong")
	}

	uploadID := tq.queue[0]
	tq.queue = tq.queue[1:]
	tq.activeUpload = uploadID

	return uploadID, nil
}

func (tq *TusQueue) IsActive(uploadID string) bool {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return tq.activeUpload == uploadID
}

func (tq *TusQueue) GetActiveUpload() string {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return tq.activeUpload
}

func (tq *TusQueue) HasActiveUpload() bool {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return tq.activeUpload != ""
}

func (tq *TusQueue) GetQueuePosition(uploadID string) int {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	if tq.activeUpload == uploadID {
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

	if tq.activeUpload == uploadID {
		tq.activeUpload = ""
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

func (tq *TusQueue) FinishActiveUpload() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	tq.activeUpload = ""
}

func (tq *TusQueue) Clear() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	tq.queue = make([]string, 0)
	tq.activeUpload = ""
}

func (tq *TusQueue) CanAcceptUpload() bool {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return tq.activeUpload == ""
}

func (tq *TusQueue) StartUpload(uploadID string) error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if tq.activeUpload != "" && tq.activeUpload != uploadID {
		return errors.New("sudah ada upload yang sedang diproses")
	}

	tq.activeUpload = uploadID

	for i, id := range tq.queue {
		if id == uploadID {
			tq.queue = append(tq.queue[:i], tq.queue[i+1:]...)
			break
		}
	}

	return nil
}

func (tq *TusQueue) Add(uploadID string) {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if tq.activeUpload == "" {
		tq.activeUpload = uploadID
		return
	}

	tq.queue = append(tq.queue, uploadID)
}

func (tq *TusQueue) IsActiveUpload(uploadID string) bool {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	return tq.activeUpload == uploadID
}
