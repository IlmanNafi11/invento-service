package repo

import (
	"fiber-boiler-plate/internal/domain"
	"time"

	"gorm.io/gorm"
)

type TusModulUploadRepository interface {
	Create(upload *domain.TusModulUpload) error
	GetByID(id string) (*domain.TusModulUpload, error)
	GetByUserID(userID uint) ([]domain.TusModulUpload, error)
	UpdateOffset(id string, offset int64, progress float64) error
	UpdateStatus(id string, status string) error
	Complete(id string, modulID uint, filePath string) error
	Delete(id string) error
	GetExpiredUploads() ([]domain.TusModulUpload, error)
	GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error)
	CountActiveByUserID(userID uint) (int, error)
	GetActiveByUserID(userID uint) ([]domain.TusModulUpload, error)
}

type tusModulUploadRepository struct {
	db *gorm.DB
}

func NewTusModulUploadRepository(db *gorm.DB) TusModulUploadRepository {
	return &tusModulUploadRepository{
		db: db,
	}
}

func (r *tusModulUploadRepository) Create(upload *domain.TusModulUpload) error {
	result := r.db.Create(upload)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *tusModulUploadRepository) GetByID(id string) (*domain.TusModulUpload, error) {
	var upload domain.TusModulUpload
	err := r.db.Where("id = ?", id).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

func (r *tusModulUploadRepository) GetByUserID(userID uint) ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusModulUploadRepository) UpdateOffset(id string, offset int64, progress float64) error {
	return r.db.Model(&domain.TusModulUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_offset": offset,
			"progress":       progress,
			"updated_at":     time.Now(),
		}).Error
}

func (r *tusModulUploadRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&domain.TusModulUpload{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *tusModulUploadRepository) Complete(id string, modulID uint, filePath string) error {
	now := time.Now()
	return r.db.Model(&domain.TusModulUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"modul_id":      modulID,
			"file_path":     filePath,
			"status":        domain.ModulUploadStatusCompleted,
			"progress":      100.0,
			"completed_at":  &now,
			"updated_at":    now,
		}).Error
}

func (r *tusModulUploadRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.TusModulUpload{}).Error
}

func (r *tusModulUploadRepository) GetExpiredUploads() ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	err := r.db.Where("expires_at < ? AND status NOT IN (?)", 
		time.Now(), 
		[]string{domain.ModulUploadStatusCompleted, domain.ModulUploadStatusExpired, domain.ModulUploadStatusCancelled},
	).Find(&uploads).Error
	return uploads, err
}

func (r *tusModulUploadRepository) GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	cutoffTime := time.Now().Add(-timeout)
	err := r.db.Where("updated_at < ? AND status IN (?)", 
		cutoffTime,
		[]string{domain.ModulUploadStatusUploading, domain.ModulUploadStatusPending},
	).Find(&uploads).Error
	return uploads, err
}

func (r *tusModulUploadRepository) CountActiveByUserID(userID uint) (int, error) {
	var count int64
	err := r.db.Model(&domain.TusModulUpload{}).
		Where("user_id = ? AND status IN (?)", 
			userID, 
			[]string{domain.ModulUploadStatusQueued, domain.ModulUploadStatusPending, domain.ModulUploadStatusUploading},
		).Count(&count).Error
	return int(count), err
}

func (r *tusModulUploadRepository) GetActiveByUserID(userID uint) ([]domain.TusModulUpload, error) {
	var uploads []domain.TusModulUpload
	err := r.db.Where("user_id = ? AND status IN (?)", 
		userID, 
		[]string{domain.ModulUploadStatusQueued, domain.ModulUploadStatusPending, domain.ModulUploadStatusUploading},
	).Order("created_at ASC").Find(&uploads).Error
	return uploads, err
}
