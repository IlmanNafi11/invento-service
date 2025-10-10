package repo

import (
	"fiber-boiler-plate/internal/domain"
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
	var roles []domain.Role
	var total int64

	query := r.db.Model(&domain.Role{})

	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		query = query.Where("nama_role LIKE ?", searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	var roleListItems []domain.RoleListItem
	for _, role := range roles {
		var permissionCount int64
		r.db.Model(&domain.RolePermission{}).Where("role_id = ?", role.ID).Count(&permissionCount)

		roleListItems = append(roleListItems, domain.RoleListItem{
			ID:                role.ID,
			NamaRole:          role.NamaRole,
			JumlahPermission:  int(permissionCount),
			TanggalDiperbarui: role.UpdatedAt,
		})
	}

	return roleListItems, int(total), nil
}
