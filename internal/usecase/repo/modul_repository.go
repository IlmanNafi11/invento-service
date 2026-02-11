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

func (r *modulRepository) GetByIDs(ids []uint, userID string) ([]domain.Modul, error) {
	var moduls []domain.Modul
	err := r.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&moduls).Error
	if err != nil {
		return nil, err
	}
	return moduls, nil
}

func (r *modulRepository) GetByIDsForUser(ids []uint, ownerUserID string) ([]domain.Modul, error) {
	var moduls []domain.Modul
	if len(ids) == 0 {
		return moduls, nil
	}
	err := r.db.Where("id IN ? AND user_id = ?", ids, ownerUserID).Find(&moduls).Error
	if err != nil {
		return nil, err
	}
	return moduls, nil
}

func (r *modulRepository) GetByUserID(userID string, search string, filterType string, filterSemester int, page, limit int) ([]domain.ModulListItem, int, error) {
	var modulListItems []domain.ModulListItem
	var total int64

	offset := (page - 1) * limit

	countQuery := `
		SELECT COUNT(*) as total
		FROM moduls
		WHERE user_id = ?
			AND (? = '' OR nama_file LIKE CONCAT('%', ?, '%'))
			AND (? = '' OR tipe = ?)
			AND (? = 0 OR semester = ?)
	`

	if err := r.db.Raw(countQuery, userID, search, search, filterType, filterType, filterSemester, filterSemester).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT
			id,
			nama_file,
			tipe,
			ukuran,
			semester,
			path_file,
			updated_at as terakhir_diperbarui
		FROM moduls
		WHERE user_id = ?
			AND (? = '' OR nama_file LIKE CONCAT('%', ?, '%'))
			AND (? = '' OR tipe = ?)
			AND (? = 0 OR semester = ?)
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	if err := r.db.Raw(dataQuery, userID, search, search, filterType, filterType, filterSemester, filterSemester, limit, offset).Scan(&modulListItems).Error; err != nil {
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
