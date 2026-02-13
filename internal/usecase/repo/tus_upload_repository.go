package repo

import (
	"fiber-boiler-plate/internal/domain"
	"time"

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

func (r *tusUploadRepository) Create(upload *domain.TusUpload) error {
	result := r.db.Create(upload)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *tusUploadRepository) GetByID(id string) (*domain.TusUpload, error) {
	var upload domain.TusUpload
	err := r.db.Where("id = ?", id).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

func (r *tusUploadRepository) GetByUserID(userID string) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) GetActiveByUserID(userID string) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("user_id = ? AND status IN (?)", userID, []string{domain.UploadStatusPending, domain.UploadStatusUploading}).
		Order("created_at ASC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) CountActiveByUserID(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&domain.TusUpload{}).
		Where("user_id = ? AND status IN (?)", userID, []string{domain.UploadStatusPending, domain.UploadStatusUploading}).
		Count(&count).Error
	return count, err
}

func (r *tusUploadRepository) UpdateOffset(id string, offset int64, progress float64) error {
	return r.db.Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_offset": offset,
			"progress":       progress,
			"updated_at":     time.Now(),
		}).Error
}

func (r *tusUploadRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (r *tusUploadRepository) Complete(id string, projectID uint, filePath string) error {
	now := time.Now()
	return r.db.Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"project_id":   projectID,
			"file_path":    filePath,
			"status":       domain.UploadStatusCompleted,
			"progress":     100.0,
			"completed_at": &now,
			"updated_at":   now,
		}).Error
}

func (r *tusUploadRepository) GetExpiredUploads(before time.Time) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("expires_at < ? AND status NOT IN (?)", before, []string{
		domain.UploadStatusCompleted,
		domain.UploadStatusExpired,
		domain.UploadStatusCancelled,
	}).Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) GetAbandonedUploads(timeout time.Duration) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	cutoffTime := time.Now().Add(-timeout)
	err := r.db.Where("updated_at < ? AND status IN (?)", cutoffTime, []string{
		domain.UploadStatusQueued,
		domain.UploadStatusUploading,
		domain.UploadStatusPending,
	}).Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.TusUpload{}).Error
}

func (r *tusUploadRepository) ListActive() ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("status IN (?)", []string{
		domain.UploadStatusPending,
		domain.UploadStatusUploading,
	}).Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) GetActiveUploadIDs() ([]string, error) {
	var ids []string
	err := r.db.Model(&domain.TusUpload{}).
		Where("status IN (?)", []string{domain.UploadStatusPending, domain.UploadStatusUploading}).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *tusUploadRepository) GetByUserIDAndStatus(userID string, status string) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) UpdateOffsetOnly(id string, offset int64) error {
	return r.db.Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_offset": offset,
			"updated_at":     time.Now(),
		}).Error
}

func (r *tusUploadRepository) UpdateUpload(upload *domain.TusUpload) error {
	return r.db.Where("id = ?", upload.ID).Updates(upload).Error
}
