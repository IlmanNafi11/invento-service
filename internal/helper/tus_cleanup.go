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

type TusModulUploadCleanupRepository interface {
	GetExpiredUploads() ([]domain.TusModulUpload, error)
	GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error)
	UpdateStatus(id string, status string) error
	Delete(id string) error
}

type TusCleanup struct {
	projectRepo     TusUploadRepository
	modulRepo       TusModulUploadCleanupRepository
	store           *TusStore
	cleanupInterval time.Duration
	idleTimeout     time.Duration
	lockTTL         time.Duration
	stopChan        chan bool
	isRunning       bool
}

func NewTusCleanup(
	projectRepo TusUploadRepository,
	modulRepo TusModulUploadCleanupRepository,
	store *TusStore,
	cleanupInterval int,
	idleTimeout int,
) *TusCleanup {
	return &TusCleanup{
		projectRepo:     projectRepo,
		modulRepo:       modulRepo,
		store:           store,
		cleanupInterval: time.Duration(cleanupInterval) * time.Second,
		idleTimeout:     time.Duration(idleTimeout) * time.Second,
		lockTTL:         30 * time.Minute,
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

	if err := tc.CleanupExpiredProjects(); err != nil {
		log.Printf("TusCleanup: error cleanup expired project: %v", err)
	}

	if err := tc.CleanupAbandonedProjects(); err != nil {
		log.Printf("TusCleanup: error cleanup abandoned project: %v", err)
	}

	if err := tc.CleanupExpiredModuls(); err != nil {
		log.Printf("TusCleanup: error cleanup expired modul: %v", err)
	}

	if err := tc.CleanupAbandonedModuls(); err != nil {
		log.Printf("TusCleanup: error cleanup abandoned modul: %v", err)
	}

	if tc.store != nil {
		cleaned := tc.store.CleanupStaleLocks(tc.lockTTL)
		if cleaned > 0 {
			log.Printf("TusCleanup: membersihkan %d stale lock", cleaned)
		}
	}

	log.Println("TusCleanup: cleanup cycle selesai")
}

// CleanupExpiredProjects membersihkan project upload yang sudah expired
func (tc *TusCleanup) CleanupExpiredProjects() error {
	now := time.Now()
	expiredUploads, err := tc.projectRepo.GetExpired(now)
	if err != nil {
		return err
	}

	if len(expiredUploads) == 0 {
		return nil
	}

	log.Printf("TusCleanup: menemukan %d project upload expired", len(expiredUploads))

	for _, upload := range expiredUploads {
		if err := tc.store.Terminate(upload.ID); err != nil {
			log.Printf("TusCleanup: gagal delete file project upload %s: %v", upload.ID, err)
		}

		if err := tc.projectRepo.UpdateStatus(upload.ID, domain.UploadStatusExpired); err != nil {
			log.Printf("TusCleanup: gagal update status project upload %s: %v", upload.ID, err)
		}

		log.Printf("TusCleanup: project upload %s berhasil di-expire", upload.ID)
	}

	return nil
}

// CleanupAbandonedProjects membersihkan project upload yang sudah tidak aktif
func (tc *TusCleanup) CleanupAbandonedProjects() error {
	activeUploads, err := tc.projectRepo.ListActive()
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
				log.Printf("TusCleanup: gagal delete file project upload abandoned %s: %v", upload.ID, err)
			}

			if err := tc.projectRepo.UpdateStatus(upload.ID, domain.UploadStatusFailed); err != nil {
				log.Printf("TusCleanup: gagal update status project upload abandoned %s: %v", upload.ID, err)
			}

			log.Printf("TusCleanup: project upload abandoned %s berhasil di-cleanup", upload.ID)
			abandonedCount++
		}
	}

	if abandonedCount > 0 {
		log.Printf("TusCleanup: menemukan dan cleanup %d project upload abandoned", abandonedCount)
	}

	return nil
}

// CleanupExpiredModuls membersihkan modul upload yang sudah expired
func (tc *TusCleanup) CleanupExpiredModuls() error {
	if tc.modulRepo == nil {
		return nil
	}

	expiredUploads, err := tc.modulRepo.GetExpiredUploads()
	if err != nil {
		return err
	}

	if len(expiredUploads) == 0 {
		return nil
	}

	log.Printf("TusCleanup: menemukan %d modul upload expired", len(expiredUploads))

	for _, upload := range expiredUploads {
		if err := tc.store.Terminate(upload.ID); err != nil {
			log.Printf("TusCleanup: gagal delete file modul upload %s: %v", upload.ID, err)
		}

		if err := tc.modulRepo.UpdateStatus(upload.ID, domain.ModulUploadStatusExpired); err != nil {
			log.Printf("TusCleanup: gagal update status modul upload %s: %v", upload.ID, err)
		}

		log.Printf("TusCleanup: modul upload %s berhasil di-expire", upload.ID)
	}

	return nil
}

// CleanupAbandonedModuls membersihkan modul upload yang sudah tidak aktif
func (tc *TusCleanup) CleanupAbandonedModuls() error {
	if tc.modulRepo == nil {
		return nil
	}

	abandonedUploads, err := tc.modulRepo.GetAbandonedUploads(tc.idleTimeout)
	if err != nil {
		return err
	}

	if len(abandonedUploads) == 0 {
		return nil
	}

	log.Printf("TusCleanup: menemukan %d modul upload abandoned", len(abandonedUploads))

	for _, upload := range abandonedUploads {
		if err := tc.store.Terminate(upload.ID); err != nil {
			log.Printf("TusCleanup: gagal delete file modul upload abandoned %s: %v", upload.ID, err)
		}

		if err := tc.modulRepo.UpdateStatus(upload.ID, domain.ModulUploadStatusFailed); err != nil {
			log.Printf("TusCleanup: gagal update status modul upload abandoned %s: %v", upload.ID, err)
		}

		log.Printf("TusCleanup: modul upload abandoned %s berhasil di-cleanup", upload.ID)
	}

	return nil
}

// CleanupUpload membersihkan satu project upload tertentu
func (tc *TusCleanup) CleanupUpload(uploadID string) error {
	if err := tc.store.Terminate(uploadID); err != nil {
		return err
	}

	if err := tc.projectRepo.Delete(uploadID); err != nil {
		return err
	}

	return nil
}

// CleanupModulUpload membersihkan satu modul upload tertentu
func (tc *TusCleanup) CleanupModulUpload(uploadID string) error {
	if tc.modulRepo == nil {
		return nil
	}

	if err := tc.store.Terminate(uploadID); err != nil {
		return err
	}

	if err := tc.modulRepo.Delete(uploadID); err != nil {
		return err
	}

	return nil
}
