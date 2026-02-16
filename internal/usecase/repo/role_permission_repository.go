package repo

import (
	"context"

	"invento-service/internal/domain"

	"gorm.io/gorm"
)

type rolePermissionRepository struct {
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

func (r *rolePermissionRepository) Create(ctx context.Context, rolePermission *domain.RolePermission) error {
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

func (r *rolePermissionRepository) BulkCreate(ctx context.Context, rolePermissions []domain.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rolePermissions).Error
}

func (r *rolePermissionRepository) GetByRoleID(ctx context.Context, roleID uint) ([]domain.RolePermission, error) {
	var rolePermissions []domain.RolePermission
	err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Preload("Permission").Find(&rolePermissions).Error
	if err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *rolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&domain.RolePermission{}).Error
}

func (r *rolePermissionRepository) GetPermissionsForRole(ctx context.Context, roleID uint) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.WithContext(ctx).Table("permissions").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
