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

func (r *projectRepository) GetByIDs(ids []uint, userID uint) ([]domain.Project, error) {
	var projects []domain.Project
	err := r.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) GetByUserID(userID uint, search string, filterSemester int, filterKategori string, page, limit int) ([]domain.ProjectListItem, int, error) {
	var projectListItems []domain.ProjectListItem
	var total int64

	countQuery := r.db.Table("projects").
		Where("user_id = ?", userID)

	if search != "" {
		searchPattern := "%" + search + "%"
		countQuery = countQuery.Where("nama_project LIKE ?", searchPattern)
	}

	if filterSemester > 0 {
		countQuery = countQuery.Where("semester = ?", filterSemester)
	}

	if filterKategori != "" {
		countQuery = countQuery.Where("kategori = ?", filterKategori)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := r.db.Table("projects").
		Select("id, nama_project, kategori, semester, ukuran, path_file, updated_at as terakhir_diperbarui").
		Where("user_id = ?", userID)

	if search != "" {
		searchPattern := "%" + search + "%"
		dataQuery = dataQuery.Where("nama_project LIKE ?", searchPattern)
	}

	if filterSemester > 0 {
		dataQuery = dataQuery.Where("semester = ?", filterSemester)
	}

	if filterKategori != "" {
		dataQuery = dataQuery.Where("kategori = ?", filterKategori)
	}

	offset := (page - 1) * limit
	if err := dataQuery.Offset(offset).Limit(limit).Order("updated_at DESC").Scan(&projectListItems).Error; err != nil {
		return nil, 0, err
	}

	return projectListItems, int(total), nil
}

func (r *projectRepository) CountByUserID(userID uint) (int, error) {
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
