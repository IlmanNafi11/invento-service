package repo

import (
	"fiber-boiler-plate/internal/domain"

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

func (r *modulRepository) GetByID(id uint) (*domain.Modul, error) {
	var modul domain.Modul
	err := r.db.First(&modul, id).Error
	if err != nil {
		return nil, err
	}
	return &modul, nil
}

func (r *modulRepository) GetByIDs(ids []uint, userID uint) ([]domain.Modul, error) {
	var moduls []domain.Modul
	err := r.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&moduls).Error
	if err != nil {
		return nil, err
	}
	return moduls, nil
}

func (r *modulRepository) GetByUserID(userID uint, search string, filterType string, filterSemester int, page, limit int) ([]domain.ModulListItem, int, error) {
	var modulListItems []domain.ModulListItem
	var total int64

	countQuery := r.db.Table("moduls").
		Where("user_id = ?", userID)

	if search != "" {
		searchPattern := "%" + search + "%"
		countQuery = countQuery.Where("nama_file LIKE ?", searchPattern)
	}

	if filterType != "" {
		countQuery = countQuery.Where("tipe = ?", filterType)
	}

	if filterSemester > 0 {
		countQuery = countQuery.Where("semester = ?", filterSemester)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := r.db.Table("moduls").
		Select("id, nama_file, tipe, ukuran, semester, path_file, updated_at as terakhir_diperbarui").
		Where("user_id = ?", userID)

	if search != "" {
		searchPattern := "%" + search + "%"
		dataQuery = dataQuery.Where("nama_file LIKE ?", searchPattern)
	}

	if filterType != "" {
		dataQuery = dataQuery.Where("tipe = ?", filterType)
	}

	if filterSemester > 0 {
		dataQuery = dataQuery.Where("semester = ?", filterSemester)
	}

	offset := (page - 1) * limit
	if err := dataQuery.Offset(offset).Limit(limit).Order("updated_at DESC").Scan(&modulListItems).Error; err != nil {
		return nil, 0, err
	}

	return modulListItems, int(total), nil
}

func (r *modulRepository) CountByUserID(userID uint) (int, error) {
	var count int64
	err := r.db.Model(&domain.Modul{}).Where("user_id = ?", userID).Count(&count).Error
	return int(count), err
}

func (r *modulRepository) Update(modul *domain.Modul) error {
	return r.db.Save(modul).Error
}

func (r *modulRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Modul{}, id).Error
}

func (r *modulRepository) UpdateMetadata(modul *domain.Modul) error {
	return r.db.Model(&domain.Modul{}).
		Where("id = ?", modul.ID).
		Updates(map[string]interface{}{
			"nama_file": modul.NamaFile,
			"semester":  modul.Semester,
		}).Error
}
