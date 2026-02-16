package upload

import (
	"context"
	"time"

	"invento-service/internal/domain"

	"github.com/rs/zerolog"
)

type TusProjectUploadCleanupRepository interface {
	GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusUpload, error)
	GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusUpload, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error
}

type TusModulUploadCleanupRepository interface {
	GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusModulUpload, error)
	GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusModulUpload, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error
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
	logger          zerolog.Logger
}

func NewTusCleanup(
	projectRepo TusProjectUploadCleanupRepository,
	modulRepo TusModulUploadCleanupRepository,
	projectStore *TusStore,
	modulStore *TusStore,
	cleanupInterval int,
	idleTimeout int,
	logger zerolog.Logger,
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
		logger:          logger.With().Str("component", "TusCleanup").Logger(),
	}
}

func (tc *TusCleanup) Start() {
	if tc.isRunning {
		return
	}

	tc.isRunning = true
	go tc.run()
	tc.logger.Info().Msg("background cleanup started")
}

func (tc *TusCleanup) Stop() {
	if !tc.isRunning {
		return
	}

	tc.stopChan <- true
	tc.isRunning = false
	tc.logger.Info().Msg("background cleanup stopped")
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
	tc.logger.Debug().Msg("starting cleanup cycle")

	if err := tc.CleanupExpiredProjects(); err != nil {
		tc.logger.Error().Err(err).Msg("error cleanup expired project")
	}

	if err := tc.CleanupAbandonedProjects(); err != nil {
		tc.logger.Error().Err(err).Msg("error cleanup abandoned project")
	}

	if err := tc.CleanupExpiredModuls(); err != nil {
		tc.logger.Error().Err(err).Msg("error cleanup expired modul")
	}

	if err := tc.CleanupAbandonedModuls(); err != nil {
		tc.logger.Error().Err(err).Msg("error cleanup abandoned modul")
	}

	tc.cleanupStaleLocks(tc.projectStore, "project")
	tc.cleanupStaleLocks(tc.modulStore, "modul")

	tc.logger.Debug().Msg("cleanup cycle completed")
}

func (tc *TusCleanup) cleanupStaleLocks(store *TusStore, label string) {
	if store == nil {
		return
	}

	cleaned := store.CleanupStaleLocks(tc.lockTTL)
	if cleaned > 0 {
		tc.logger.Info().Int("count", cleaned).Str("type", label).Msg("cleaned stale locks")
	}
}

func (tc *TusCleanup) cleanupUploads(
	ctx context.Context,
	uploadIDs []string,
	store *TusStore,
	newStatus string,
	label string,
	updateStatus func(ctx context.Context, id string, status string) error,
) int {
	cleaned := 0
	for _, uploadID := range uploadIDs {
		if store != nil {
			if err := store.Terminate(uploadID); err != nil {
				tc.logger.Error().Err(err).Str("type", label).Str("upload_id", uploadID).Msg("failed to delete upload file")
			}
		}

		if err := updateStatus(ctx, uploadID, newStatus); err != nil {
			tc.logger.Error().Err(err).Str("type", label).Str("upload_id", uploadID).Msg("failed to update upload status")
			continue
		}

		tc.logger.Info().Str("type", label).Str("upload_id", uploadID).Msg("upload cleaned up")
		cleaned++
	}

	return cleaned
}

func (tc *TusCleanup) CleanupExpiredProjects() error {
	if tc.projectRepo == nil {
		return nil
	}

	ctx := context.Background()
	expiredUploads, err := tc.projectRepo.GetExpiredUploads(ctx, time.Now())
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

	cleaned := tc.cleanupUploads(ctx, uploadIDs, tc.projectStore, domain.UploadStatusExpired, "project", tc.projectRepo.UpdateStatus)
	if cleaned > 0 {
		tc.logger.Info().Int("count", cleaned).Msg("cleaned expired project uploads")
	}

	return nil
}

func (tc *TusCleanup) CleanupAbandonedProjects() error {
	if tc.projectRepo == nil {
		return nil
	}

	ctx := context.Background()
	abandonedUploads, err := tc.projectRepo.GetAbandonedUploads(ctx, tc.idleTimeout)
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

	cleaned := tc.cleanupUploads(ctx, uploadIDs, tc.projectStore, domain.UploadStatusFailed, "project abandoned", tc.projectRepo.UpdateStatus)
	if cleaned > 0 {
		tc.logger.Info().Int("count", cleaned).Msg("cleaned abandoned project uploads")
	}

	return nil
}

func (tc *TusCleanup) CleanupExpiredModuls() error {
	if tc.modulRepo == nil {
		return nil
	}

	ctx := context.Background()
	expiredUploads, err := tc.modulRepo.GetExpiredUploads(ctx, time.Now())
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

	cleaned := tc.cleanupUploads(ctx, uploadIDs, tc.modulStore, domain.UploadStatusExpired, "modul", tc.modulRepo.UpdateStatus)
	if cleaned > 0 {
		tc.logger.Info().Int("count", cleaned).Msg("cleaned expired modul uploads")
	}

	return nil
}

func (tc *TusCleanup) CleanupAbandonedModuls() error {
	if tc.modulRepo == nil {
		return nil
	}

	ctx := context.Background()
	abandonedUploads, err := tc.modulRepo.GetAbandonedUploads(ctx, tc.idleTimeout)
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

	cleaned := tc.cleanupUploads(ctx, uploadIDs, tc.modulStore, domain.UploadStatusFailed, "modul abandoned", tc.modulRepo.UpdateStatus)
	if cleaned > 0 {
		tc.logger.Info().Int("count", cleaned).Msg("cleaned abandoned modul uploads")
	}

	return nil
}

func (tc *TusCleanup) cleanupSingleUpload(ctx context.Context, uploadID string, store *TusStore, deleteFn func(ctx context.Context, id string) error) error {
	if store != nil {
		if err := store.Terminate(uploadID); err != nil {
			return err
		}
	}

	return deleteFn(ctx, uploadID)
}

func (tc *TusCleanup) CleanupUpload(uploadID string) error {
	if tc.projectRepo == nil {
		return nil
	}

	return tc.cleanupSingleUpload(context.Background(), uploadID, tc.projectStore, tc.projectRepo.Delete)
}

func (tc *TusCleanup) CleanupModulUpload(uploadID string) error {
	if tc.modulRepo == nil {
		return nil
	}

	return tc.cleanupSingleUpload(context.Background(), uploadID, tc.modulStore, tc.modulRepo.Delete)
}
