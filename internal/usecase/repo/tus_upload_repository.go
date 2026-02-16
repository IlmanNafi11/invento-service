package repo

import (
	"context"
	"time"

	"invento-service/internal/domain"

	"gorm.io/gorm"
)

type tusUploadRepository struct {
	db *gorm.DB
}

func NewTusUploadRepository(db *gorm.DB) *tusUploadRepository {
	return &tusUploadRepository{
		db: db,
	}
}

func (r *tusUploadRepository) Create(ctx context.Context, upload *domain.TusUpload) error {
	result := r.db.WithContext(ctx).Create(upload)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *tusUploadRepository) GetByID(ctx context.Context, id string) (*domain.TusUpload, error) {
	var upload domain.TusUpload
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

func (r *tusUploadRepository) GetByUserID(ctx context.Context, userID string) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) GetActiveByUserID(ctx context.Context, userID string) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.WithContext(ctx).Where("user_id = ? AND status IN (?)", userID, []string{domain.UploadStatusQueued, domain.UploadStatusPending, domain.UploadStatusUploading}).
		Order("created_at ASC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) CountActiveByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.TusUpload{}).
		Where("user_id = ? AND status IN (?)", userID, []string{domain.UploadStatusPending, domain.UploadStatusUploading}).
		Count(&count).Error
	return count, err
}

func (r *tusUploadRepository) UpdateOffset(ctx context.Context, id string, offset int64, progress float64) error {
	return r.db.WithContext(ctx).Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_offset": offset,
			"progress":       progress,
		}).Error
}

func (r *tusUploadRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
		}).Error
}

func (r *tusUploadRepository) Complete(ctx context.Context, id string, projectID uint, filePath string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"project_id":   projectID,
			"file_path":    filePath,
			"status":       domain.UploadStatusCompleted,
			"progress":     100.0,
			"completed_at": &now,
		}).Error
}

func (r *tusUploadRepository) GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.WithContext(ctx).Where("expires_at < ? AND status NOT IN (?)", before, []string{
		domain.UploadStatusCompleted,
		domain.UploadStatusExpired,
		domain.UploadStatusCancelled,
	}).Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	cutoffTime := time.Now().Add(-timeout)
	err := r.db.WithContext(ctx).Where("updated_at < ? AND status IN (?)", cutoffTime, []string{
		domain.UploadStatusQueued,
		domain.UploadStatusUploading,
		domain.UploadStatusPending,
	}).Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.TusUpload{}).Error
}

func (r *tusUploadRepository) ListActive(ctx context.Context) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.WithContext(ctx).Where("status IN (?)", []string{
		domain.UploadStatusPending,
		domain.UploadStatusUploading,
	}).Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) GetActiveUploadIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&domain.TusUpload{}).
		Where("status IN (?)", []string{domain.UploadStatusQueued, domain.UploadStatusPending, domain.UploadStatusUploading}).
		Pluck("id", &ids).Error
	return ids, err
}
