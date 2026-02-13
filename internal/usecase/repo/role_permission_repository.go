package repo

import (
	"fiber-boiler-plate/internal/domain"

	"gorm.io/gorm"
)

type rolePermissionRepository struct {
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

func (r *rolePermissionRepository) Create(rolePermission *domain.RolePermission) error {
	return r.db.Create(rolePermission).Error
}

func (r *rolePermissionRepository) BulkCreate(rolePermissions []domain.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	return r.db.Create(&rolePermissions).Error
}

func (r *rolePermissionRepository) GetByRoleID(roleID uint) ([]domain.RolePermission, error) {
	var rolePermissions []domain.RolePermission
	err := r.db.Where("role_id = ?", roleID).Preload("Permission").Find(&rolePermissions).Error
	if err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *rolePermissionRepository) DeleteByRoleID(roleID uint) error {
	return r.db.Where("role_id = ?", roleID).Delete(&domain.RolePermission{}).Error
}

func (r *rolePermissionRepository) GetPermissionsForRole(roleID uint) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.Table("permissions").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
