package repo

import (
	"errors"
	"fmt"

	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type modulRepository struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func NewModulRepository(db *gorm.DB, logger zerolog.Logger) ModulRepository {
	return &modulRepository{
		db:     db,
		logger: logger.With().Str("component", "ModulRepository").Logger(),
	}
}

func (r *modulRepository) Create(modul *domain.Modul) error {
	if err := r.db.Create(modul).Error; err != nil {
		return fmt.Errorf("ModulRepository.Create: %w", err)
	}
	return nil
}

func (r *modulRepository) GetByID(id string) (*domain.Modul, error) {
	var modul domain.Modul
	err := r.db.Where("id = ?", id).First(&modul).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrRecordNotFound
		}
		return nil, fmt.Errorf("ModulRepository.GetByID: %w", err)
	}
	return &modul, nil
}

func (r *modulRepository) GetByIDs(ids []string, userID string) ([]domain.Modul, error) {
	var moduls []domain.Modul
	err := r.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&moduls).Error
	if err != nil {
		return nil, fmt.Errorf("ModulRepository.GetByIDs: %w", err)
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
		r.logger.Error().Err(err).Str("user_id", userID).Str("search", search).Str("filter_type", filterType).Str("filter_status", filterStatus).Msg("count query failed")
		return nil, 0, fmt.Errorf("ModulRepository.GetByUserID: count query: %w", err)
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
		r.logger.Error().Err(err).Str("user_id", userID).Str("search", search).Str("filter_type", filterType).Str("filter_status", filterStatus).Int("limit", limit).Int("offset", offset).Msg("data query failed")
		return nil, 0, fmt.Errorf("ModulRepository.GetByUserID: data query: %w", err)
	}

	return modulListItems, int(total), nil
}

func (r *modulRepository) CountByUserID(userID string) (int, error) {
	var count int64
	err := r.db.Model(&domain.Modul{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("ModulRepository.CountByUserID: %w", err)
	}
	return int(count), nil
}

func (r *modulRepository) Update(modul *domain.Modul) error {
	if err := r.db.Save(modul).Error; err != nil {
		return fmt.Errorf("ModulRepository.Update: %w", err)
	}
	return nil
}

func (r *modulRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&domain.Modul{}).Error; err != nil {
		return fmt.Errorf("ModulRepository.Delete: %w", err)
	}
	return nil
}

func (r *modulRepository) UpdateMetadata(modul *domain.Modul) error {
	if err := r.db.Model(&domain.Modul{}).
		Where("id = ?", modul.ID).
		Updates(map[string]interface{}{
			"judul":     modul.Judul,
			"deskripsi": modul.Deskripsi,
		}).Error; err != nil {
		return fmt.Errorf("ModulRepository.UpdateMetadata: %w", err)
	}
	return nil
}
