package helper

import (
	"fiber-boiler-plate/internal/domain"
	"log"
	"time"
)

type TusProjectUploadCleanupRepository interface {
	GetExpiredUploads(before time.Time) ([]domain.TusUpload, error)
	GetAbandonedUploads(timeout time.Duration) ([]domain.TusUpload, error)
	UpdateStatus(id string, status string) error
	Delete(id string) error
}

type TusModulUploadCleanupRepository interface {
	GetExpiredUploads(before time.Time) ([]domain.TusModulUpload, error)
	GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error)
	UpdateStatus(id string, status string) error
	Delete(id string) error
}

type TusCleanup struct {
	projectRepo     TusProjectUploadCleanupRepository
	modulRepo       TusModulUploadCleanupRepository
	projectStore    *TusStore
	modulStore      *TusStore
	cleanupInterval time.Duration
	idleTimeout     time.Duration
	lockTTL         time.Duration
	stopChan        chan bool
	isRunning       bool
}

func NewTusCleanup(
	projectRepo TusProjectUploadCleanupRepository,
	modulRepo TusModulUploadCleanupRepository,
	projectStore *TusStore,
	modulStore *TusStore,
	cleanupInterval int,
	idleTimeout int,
) *TusCleanup {
	return &TusCleanup{
		projectRepo:     projectRepo,
		modulRepo:       modulRepo,
		projectStore:    projectStore,
		modulStore:      modulStore,
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

	tc.cleanupStaleLocks(tc.projectStore, "project")
	tc.cleanupStaleLocks(tc.modulStore, "modul")

	log.Println("TusCleanup: cleanup cycle selesai")
}

func (tc *TusCleanup) cleanupStaleLocks(store *TusStore, label string) {
	if store == nil {
		return
	}

	cleaned := store.CleanupStaleLocks(tc.lockTTL)
	if cleaned > 0 {
		log.Printf("TusCleanup: membersihkan %d stale lock %s", cleaned, label)
	}
}

func (tc *TusCleanup) cleanupUploads(
	uploadIDs []string,
	store *TusStore,
	newStatus string,
	label string,
	updateStatus func(id string, status string) error,
) int {
	cleaned := 0
	for _, uploadID := range uploadIDs {
		if store != nil {
			if err := store.Terminate(uploadID); err != nil {
				log.Printf("TusCleanup: gagal delete file %s upload %s: %v", label, uploadID, err)
			}
		}

		if err := updateStatus(uploadID, newStatus); err != nil {
			log.Printf("TusCleanup: gagal update status %s upload %s: %v", label, uploadID, err)
			continue
		}

		log.Printf("TusCleanup: %s upload %s berhasil di-cleanup", label, uploadID)
		cleaned++
	}

	return cleaned
}

func (tc *TusCleanup) CleanupExpiredProjects() error {
	if tc.projectRepo == nil {
		return nil
	}

	expiredUploads, err := tc.projectRepo.GetExpiredUploads(time.Now())
	if err != nil {
		return err
	}

	if len(expiredUploads) == 0 {
		return nil
	}

	uploadIDs := make([]string, 0, len(expiredUploads))
	for _, upload := range expiredUploads {
		uploadIDs = append(uploadIDs, upload.ID)
	}

	cleaned := tc.cleanupUploads(uploadIDs, tc.projectStore, domain.UploadStatusExpired, "project", tc.projectRepo.UpdateStatus)
	if cleaned > 0 {
		log.Printf("TusCleanup: menemukan dan cleanup %d project upload expired", cleaned)
	}

	return nil
}

func (tc *TusCleanup) CleanupAbandonedProjects() error {
	if tc.projectRepo == nil {
		return nil
	}

	abandonedUploads, err := tc.projectRepo.GetAbandonedUploads(tc.idleTimeout)
	if err != nil {
		return err
	}

	if len(abandonedUploads) == 0 {
		return nil
	}

	uploadIDs := make([]string, 0, len(abandonedUploads))
	for _, upload := range abandonedUploads {
		uploadIDs = append(uploadIDs, upload.ID)
	}

	cleaned := tc.cleanupUploads(uploadIDs, tc.projectStore, domain.UploadStatusFailed, "project abandoned", tc.projectRepo.UpdateStatus)
	if cleaned > 0 {
		log.Printf("TusCleanup: menemukan dan cleanup %d project upload abandoned", cleaned)
	}

	return nil
}

func (tc *TusCleanup) CleanupExpiredModuls() error {
	if tc.modulRepo == nil {
		return nil
	}

	expiredUploads, err := tc.modulRepo.GetExpiredUploads(time.Now())
	if err != nil {
		return err
	}

	if len(expiredUploads) == 0 {
		return nil
	}

	uploadIDs := make([]string, 0, len(expiredUploads))
	for _, upload := range expiredUploads {
		uploadIDs = append(uploadIDs, upload.ID)
	}

	cleaned := tc.cleanupUploads(uploadIDs, tc.modulStore, domain.UploadStatusExpired, "modul", tc.modulRepo.UpdateStatus)
	if cleaned > 0 {
		log.Printf("TusCleanup: menemukan dan cleanup %d modul upload expired", cleaned)
	}

	return nil
}

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

	uploadIDs := make([]string, 0, len(abandonedUploads))
	for _, upload := range abandonedUploads {
		uploadIDs = append(uploadIDs, upload.ID)
	}

	cleaned := tc.cleanupUploads(uploadIDs, tc.modulStore, domain.UploadStatusFailed, "modul abandoned", tc.modulRepo.UpdateStatus)
	if cleaned > 0 {
		log.Printf("TusCleanup: menemukan dan cleanup %d modul upload abandoned", cleaned)
	}

	return nil
}

func (tc *TusCleanup) cleanupSingleUpload(uploadID string, store *TusStore, deleteFn func(id string) error) error {
	if store != nil {
		if err := store.Terminate(uploadID); err != nil {
			return err
		}
	}

	return deleteFn(uploadID)
}

func (tc *TusCleanup) CleanupUpload(uploadID string) error {
	if tc.projectRepo == nil {
		return nil
	}

	return tc.cleanupSingleUpload(uploadID, tc.projectStore, tc.projectRepo.Delete)
}

func (tc *TusCleanup) CleanupModulUpload(uploadID string) error {
	if tc.modulRepo == nil {
		return nil
	}

	return tc.cleanupSingleUpload(uploadID, tc.modulStore, tc.modulRepo.Delete)
}
