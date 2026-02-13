package repo

import (
	"fiber-boiler-plate/internal/domain"

	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *domain.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) GetByID(id uint) (*domain.Project, error) {
	var project domain.Project
	err := r.db.First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetByIDs(ids []uint, userID string) ([]domain.Project, error) {
	var projects []domain.Project
	err := r.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) GetByIDsForUser(ids []uint, ownerUserID string) ([]domain.Project, error) {
	var projects []domain.Project
	if len(ids) == 0 {
		return projects, nil
	}
	err := r.db.Where("id IN ? AND user_id = ?", ids, ownerUserID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) GetByUserID(userID string, search string, filterSemester int, filterKategori string, page, limit int) ([]domain.ProjectListItem, int, error) {
	var projectListItems []domain.ProjectListItem
	var total int64

	offset := (page - 1) * limit

	countQuery := `
		SELECT COUNT(*) as total
		FROM projects
		WHERE user_id = ?
			AND (? = '' OR LOWER(nama_project) LIKE '%' || LOWER(?) || '%')
			AND (? = 0 OR semester = ?)
			AND (? = '' OR kategori = ?)
	`

	if err := r.db.Raw(countQuery, userID, search, search, filterSemester, filterSemester, filterKategori, filterKategori).Scan(&total).Error; err != nil {
		return nil, 0, err
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
			AND (? = '' OR LOWER(nama_project) LIKE '%' || LOWER(?) || '%')
			AND (? = 0 OR semester = ?)
			AND (? = '' OR kategori = ?)
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	if err := r.db.Raw(dataQuery, userID, search, search, filterSemester, filterSemester, filterKategori, filterKategori, limit, offset).Scan(&projectListItems).Error; err != nil {
		return nil, 0, err
	}

	return projectListItems, int(total), nil
}

func (r *projectRepository) CountByUserID(userID string) (int, error) {
	var count int64
	err := r.db.Model(&domain.Project{}).Where("user_id = ?", userID).Count(&count).Error
	return int(count), err
}

func (r *projectRepository) Update(project *domain.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Project{}, id).Error
}
