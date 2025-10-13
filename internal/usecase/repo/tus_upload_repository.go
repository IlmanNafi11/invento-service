package repo

import (
	"log"
	"reflect"
	"fiber-boiler-plate/internal/domain"
	"time"

	"gorm.io/gorm"
)

type tusUploadRepository struct {
	db *gorm.DB
}

func NewTusUploadRepository(db *gorm.DB) TusUploadRepository {
	return &tusUploadRepository{
		db: db,
	}
}

func (r *tusUploadRepository) Create(upload *domain.TusUpload) error {
	log.Printf("DEBUG: Repository Create() called")
	log.Printf("DEBUG: Upload struct: %+v", upload)
	
	// Debug field values
	v := reflect.ValueOf(*upload)
	t := reflect.TypeOf(*upload)
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		log.Printf("DEBUG: Field %s (%s) = %v", field.Name, field.Type, value.Interface())
	}
	
	// Debug SQL query
	result := r.db.Create(upload)
	if result.Error != nil {
		log.Printf("ERROR: GORM Create() failed - Error: %v", result.Error)
		return result.Error
	}
	
	log.Printf("DEBUG: Repository Create() success - Rows affected: %d", result.RowsAffected)
	log.Printf("DEBUG: Created upload ID: %s", upload.ID)
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

func (r *tusUploadRepository) GetByUserID(userID uint) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
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

func (r *tusUploadRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&domain.TusUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (r *tusUploadRepository) GetExpired(before time.Time) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("expires_at < ? AND status NOT IN (?)", before, []string{
		domain.UploadStatusCompleted,
		domain.UploadStatusExpired,
		domain.UploadStatusCancelled,
	}).Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) GetByUserIDAndStatus(userID uint, status string) ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("user_id = ? AND status = ?", userID, status).
		Order("created_at DESC").
		Find(&uploads).Error
	return uploads, err
}

func (r *tusUploadRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.TusUpload{}).Error
}

func (r *tusUploadRepository) ListActive() ([]domain.TusUpload, error) {
	var uploads []domain.TusUpload
	err := r.db.Where("status IN (?)", []string{
		domain.UploadStatusQueued,
		domain.UploadStatusUploading,
	}).Find(&uploads).Error
	return uploads, err
}
