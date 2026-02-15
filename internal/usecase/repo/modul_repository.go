package repo

import (
	"invento-service/internal/domain"
	"log"

	"gorm.io/gorm"
)

type modulRepository struct {
	db *gorm.DB
}

func NewModulRepository(db *gorm.DB) ModulRepository {
	return &modulRepository{db: db}
}

func (r *modulRepository) Create(modul *domain.Modul) error {
	return r.db.Create(modul).Error
}

func (r *modulRepository) GetByID(id string) (*domain.Modul, error) {
	var modul domain.Modul
	err := r.db.Where("id = ?", id).First(&modul).Error
	if err != nil {
		return nil, err
	}
	return &modul, nil
}

func (r *modulRepository) GetByIDs(ids []string, userID string) ([]domain.Modul, error) {
	var moduls []domain.Modul
	err := r.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&moduls).Error
	if err != nil {
		return nil, err
	}
	return moduls, nil
}

func (r *modulRepository) GetByUserID(userID string, search string, filterType string, filterStatus string, page, limit int) ([]domain.ModulListItem, int, error) {
	var modulListItems []domain.ModulListItem
	var total int64

	offset := (page - 1) * limit

	countQuery := `
		SELECT COUNT(*) as total
		FROM moduls
		WHERE user_id = ?
			AND (? = '' OR LOWER(judul) LIKE '%' || LOWER(?) || '%' OR LOWER(deskripsi) LIKE '%' || LOWER(?) || '%')
			AND (? = '' OR mime_type = ?)
			AND (? = '' OR status = ?)
	`

	if err := r.db.Raw(countQuery, userID, search, search, search, filterType, filterType, filterStatus, filterStatus).Scan(&total).Error; err != nil {
		log.Printf("[ERROR] ModulRepository.GetByUserID count query failed - query: %s, params: userID=%s, search=%s, filterType=%s, filterStatus=%s, error: %v",
			countQuery, userID, search, filterType, filterStatus, err)
		return nil, 0, err
	}

	dataQuery := `
		SELECT
			id,
			judul,
			deskripsi,
			file_name,
			mime_type,
			file_size,
			status,
			updated_at as terakhir_diperbarui
		FROM moduls
		WHERE user_id = ?
			AND (? = '' OR LOWER(judul) LIKE '%' || LOWER(?) || '%' OR LOWER(deskripsi) LIKE '%' || LOWER(?) || '%')
			AND (? = '' OR mime_type = ?)
			AND (? = '' OR status = ?)
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	if err := r.db.Raw(dataQuery, userID, search, search, search, filterType, filterType, filterStatus, filterStatus, limit, offset).Scan(&modulListItems).Error; err != nil {
		log.Printf("[ERROR] ModulRepository.GetByUserID data query failed - query: %s, params: userID=%s, search=%s, filterType=%s, filterStatus=%s, limit=%d, offset=%d, error: %v",
			dataQuery, userID, search, filterType, filterStatus, limit, offset, err)
		return nil, 0, err
	}

	return modulListItems, int(total), nil
}

func (r *modulRepository) CountByUserID(userID string) (int, error) {
	var count int64
	err := r.db.Model(&domain.Modul{}).Where("user_id = ?", userID).Count(&count).Error
	return int(count), err
}

func (r *modulRepository) Update(modul *domain.Modul) error {
	return r.db.Save(modul).Error
}

func (r *modulRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Modul{}).Error
}

func (r *modulRepository) UpdateMetadata(modul *domain.Modul) error {
	return r.db.Model(&domain.Modul{}).
		Where("id = ?", modul.ID).
		Updates(map[string]interface{}{
			"judul":     modul.Judul,
			"deskripsi": modul.Deskripsi,
		}).Error
}
