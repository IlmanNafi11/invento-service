package helper

import (
	"fiber-boiler-plate/internal/domain"
	"log"
	"time"
)

type TusUploadRepository interface {
	GetExpired(before time.Time) ([]domain.TusUpload, error)
	UpdateStatus(id string, status string) error
	Delete(id string) error
	ListActive() ([]domain.TusUpload, error)
}

type TusCleanup struct {
	repo            TusUploadRepository
	store           *TusStore
	cleanupInterval time.Duration
	idleTimeout     time.Duration
	stopChan        chan bool
	isRunning       bool
}

func NewTusCleanup(repo TusUploadRepository, store *TusStore, cleanupInterval int, idleTimeout int) *TusCleanup {
	return &TusCleanup{
		repo:            repo,
		store:           store,
		cleanupInterval: time.Duration(cleanupInterval) * time.Second,
		idleTimeout:     time.Duration(idleTimeout) * time.Second,
		stopChan:        make(chan bool),
		isRunning:       false,
	}
}

func (tc *TusCleanup) Start() {
	if tc.isRunning {
		return
	}

	tc.isRunning = true
	go tc.run()
	log.Println("TusCleanup: background cleanup dimulai")
}

func (tc *TusCleanup) Stop() {
	if !tc.isRunning {
		return
	}

	tc.stopChan <- true
	tc.isRunning = false
	log.Println("TusCleanup: background cleanup dihentikan")
}

func (tc *TusCleanup) run() {
	ticker := time.NewTicker(tc.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tc.performCleanup()
		case <-tc.stopChan:
			return
		}
	}
}

func (tc *TusCleanup) performCleanup() {
	log.Println("TusCleanup: memulai cleanup cycle")

	if err := tc.CleanupExpired(); err != nil {
		log.Printf("TusCleanup: error cleanup expired: %v", err)
	}

	if err := tc.CleanupAbandoned(); err != nil {
		log.Printf("TusCleanup: error cleanup abandoned: %v", err)
	}

	log.Println("TusCleanup: cleanup cycle selesai")
}

func (tc *TusCleanup) CleanupExpired() error {
	now := time.Now()
	expiredUploads, err := tc.repo.GetExpired(now)
	if err != nil {
		return err
	}

	if len(expiredUploads) == 0 {
		return nil
	}

	log.Printf("TusCleanup: menemukan %d upload expired", len(expiredUploads))

	for _, upload := range expiredUploads {
		if err := tc.store.Terminate(upload.ID); err != nil {
			log.Printf("TusCleanup: gagal delete file upload %s: %v", upload.ID, err)
		}

		if err := tc.repo.UpdateStatus(upload.ID, domain.UploadStatusExpired); err != nil {
			log.Printf("TusCleanup: gagal update status upload %s: %v", upload.ID, err)
		}

		log.Printf("TusCleanup: upload %s berhasil di-expire", upload.ID)
	}

	return nil
}

func (tc *TusCleanup) CleanupAbandoned() error {
	activeUploads, err := tc.repo.ListActive()
	if err != nil {
		return err
	}

	if len(activeUploads) == 0 {
		return nil
	}

	now := time.Now()
	idleThreshold := now.Add(-tc.idleTimeout)

	abandonedCount := 0
	for _, upload := range activeUploads {
		if upload.UpdatedAt.Before(idleThreshold) {
			if err := tc.store.Terminate(upload.ID); err != nil {
				log.Printf("TusCleanup: gagal delete file upload abandoned %s: %v", upload.ID, err)
			}

			if err := tc.repo.UpdateStatus(upload.ID, domain.UploadStatusFailed); err != nil {
				log.Printf("TusCleanup: gagal update status upload abandoned %s: %v", upload.ID, err)
			}

			log.Printf("TusCleanup: upload abandoned %s berhasil di-cleanup", upload.ID)
			abandonedCount++
		}
	}

	if abandonedCount > 0 {
		log.Printf("TusCleanup: menemukan dan cleanup %d upload abandoned", abandonedCount)
	}

	return nil
}

func (tc *TusCleanup) CleanupUpload(uploadID string) error {
	if err := tc.store.Terminate(uploadID); err != nil {
		return err
	}

	if err := tc.repo.Delete(uploadID); err != nil {
		return err
	}

	return nil
}
