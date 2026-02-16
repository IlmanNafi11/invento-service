package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"invento-service/internal/domain"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"

	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		return fmt.Errorf("ProjectRepository.Create: %w", err)
	}
	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, id uint) (*domain.Project, error) {
	var project domain.Project
	err := r.db.WithContext(ctx).First(&project, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrRecordNotFound
		}
		return nil, fmt.Errorf("ProjectRepository.GetByID: %w", err)
	}
	return &project, nil
}

func (r *projectRepository) GetByIDs(ctx context.Context, ids []uint, userID string) ([]domain.Project, error) {
	var projects []domain.Project
	if len(ids) == 0 {
		return projects, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ? AND user_id = ?", ids, userID).Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("ProjectRepository.GetByIDs: %w", err)
	}
	return projects, nil
}

func (r *projectRepository) GetByUserID(ctx context.Context, userID string, search string, filterSemester int, filterKategori string, page, limit int) ([]dto.ProjectListItem, int, error) {
	var projectListItems []dto.ProjectListItem
	var total int64

	offset := (page - 1) * limit

	escapedSearch := search
	if search != "" {
		replacer := strings.NewReplacer("%", "\\%", "_", "\\_", "\\", "\\\\")
		escapedSearch = replacer.Replace(search)
	}

	countQuery := `
		SELECT COUNT(*) as total
		FROM projects
		WHERE user_id = ?
			AND (? = '' OR LOWER(nama_project) LIKE '%' || LOWER(?) || '%' ESCAPE '\')
			AND (? = 0 OR semester = ?)
			AND (? = '' OR kategori = ?)
	`

	if err := r.db.WithContext(ctx).Raw(countQuery, userID, search, escapedSearch, filterSemester, filterSemester, filterKategori, filterKategori).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("ProjectRepository.GetByUserID: count query: %w", err)
	}

	dataQuery := `
		SELECT
			id,
			nama_project,
			kategori,
			semester,
			ukuran,
			path_file,
		updated_at as terakhir_diperbarui
		FROM projects
		WHERE user_id = ?
			AND (? = '' OR LOWER(nama_project) LIKE '%' || LOWER(?) || '%' ESCAPE '\')
			AND (? = 0 OR semester = ?)
			AND (? = '' OR kategori = ?)
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	if err := r.db.WithContext(ctx).Raw(dataQuery, userID, search, escapedSearch, filterSemester, filterSemester, filterKategori, filterKategori, limit, offset).Scan(&projectListItems).Error; err != nil {
		return nil, 0, fmt.Errorf("ProjectRepository.GetByUserID: data query: %w", err)
	}

	return projectListItems, int(total), nil
}

func (r *projectRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Project{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("ProjectRepository.CountByUserID: %w", err)
	}
	return int(count), nil
}

func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	if err := r.db.WithContext(ctx).Model(project).Updates(map[string]interface{}{
		"nama_project": project.NamaProject,
		"kategori":     project.Kategori,
		"semester":     project.Semester,
		"ukuran":       project.Ukuran,
		"path_file":    project.PathFile,
	}).Error; err != nil {
		return fmt.Errorf("ProjectRepository.Update: %w", err)
	}
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Project{}, id).Error; err != nil {
		return fmt.Errorf("ProjectRepository.Delete: %w", err)
	}
	return nil
}
