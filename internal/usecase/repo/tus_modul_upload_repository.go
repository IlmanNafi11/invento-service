package repo

import (
	"context"
	"time"

	"invento-service/internal/domain"

	"gorm.io/gorm"
)

type tusModulUploadRepository struct {
	db *gorm.DB
}

func NewTusModulUploadRepository(db *gorm.DB) TusModulUploadRepository {
	return &tusModulUploadRepository{
		db: db,
	}
}

func (r *tusModulUploadRepository) Create(ctx context.Context, upload *domain.TusModulUpload) error {
	result := r.db.WithContext(ctx).Create(upload)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *tusModulUploadRepository) GetByID(ctx context.Context, id string) (*domain.TusModulUpload, error) {
	var upload domain.TusModulUpload
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

func (r *tusModulUploadRepository) GetByUserID(ctx context.Context, userID string) ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusModulUploadRepository) UpdateOffset(ctx context.Context, id string, offset int64, progress float64) error {
	return r.db.WithContext(ctx).Model(&domain.TusModulUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_offset": offset,
			"progress":       progress,
			"updated_at":     time.Now(),
		}).Error
}

func (r *tusModulUploadRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&domain.TusModulUpload{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *tusModulUploadRepository) Complete(ctx context.Context, id string, modulID string, filePath string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.TusModulUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"modul_id":     modulID,
			"file_path":    filePath,
			"status":       domain.UploadStatusCompleted,
			"progress":     100.0,
			"completed_at": &now,
			"updated_at":   now,
		}).Error
}

func (r *tusModulUploadRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.TusModulUpload{}).Error
}

func (r *tusModulUploadRepository) GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	err := r.db.WithContext(ctx).Where("expires_at < ? AND status NOT IN (?)",
		before,
		[]string{domain.UploadStatusCompleted, domain.UploadStatusExpired, domain.UploadStatusCancelled},
	).Find(&uploads).Error
	return uploads, err
}

func (r *tusModulUploadRepository) GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	cutoffTime := time.Now().Add(-timeout)
	err := r.db.WithContext(ctx).Where("updated_at < ? AND status IN (?)",
		cutoffTime,
		[]string{domain.UploadStatusUploading, domain.UploadStatusPending},
	).Find(&uploads).Error
	return uploads, err
}

func (r *tusModulUploadRepository) CountActiveByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.TusModulUpload{}).
		Where("user_id = ? AND status IN (?)",
			userID,
			[]string{domain.UploadStatusPending, domain.UploadStatusUploading},
		).Count(&count).Error
	return count, err
}

func (r *tusModulUploadRepository) GetActiveByUserID(ctx context.Context, userID string) ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	err := r.db.WithContext(ctx).Where("user_id = ? AND status IN (?)",
		userID,
		[]string{domain.UploadStatusPending, domain.UploadStatusUploading},
	).Order("created_at ASC").Find(&uploads).Error
	return uploads, err
}

func (r *tusModulUploadRepository) GetActiveUploadIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&domain.TusModulUpload{}).
		Where("status IN (?)", []string{domain.UploadStatusPending, domain.UploadStatusUploading}).
		Pluck("id", &ids).Error
	return ids, err
}
