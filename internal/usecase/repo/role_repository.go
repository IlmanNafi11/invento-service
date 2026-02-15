package repo

import (
	"invento-service/internal/domain"
	"fmt"

	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id uint) (*domain.Role, error) {
	var role domain.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Where("nama_role = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(role *domain.Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Role{}, id).Error
}

func (r *roleRepository) GetAll(search string, page, limit int) ([]domain.RoleListItem, int, error) {
	var total int64

	query := r.db.Model(&domain.Role{})

	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		query = query.Where("nama_role ILIKE ?", searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var roleListItems []domain.RoleListItem

	listQuery := r.db.Model(&domain.Role{}).
		Select("roles.id, roles.nama_role, roles.updated_at as tanggal_diperbarui, COUNT(role_permissions.id) as jumlah_permission").
		Joins("LEFT JOIN role_permissions ON role_permissions.role_id = roles.id").
		Group("roles.id").
		Order("roles.updated_at DESC").
		Offset(offset).
		Limit(limit)

	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		listQuery = listQuery.Where("roles.nama_role ILIKE ?", searchPattern)
	}

	if err := listQuery.Scan(&roleListItems).Error; err != nil {
		return nil, 0, err
	}

	if roleListItems == nil {
		roleListItems = []domain.RoleListItem{}
	}

	return roleListItems, int(total), nil
}
