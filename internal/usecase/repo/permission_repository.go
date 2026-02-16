package repo

import (
	"context"
	"strings"

	"invento-service/internal/domain"
	"invento-service/internal/dto"

	"gorm.io/gorm"
)

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByResourceAndAction(ctx context.Context, resource, action string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).Where("resource = ? AND action = ?", resource, action).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetAllByResourceActions(ctx context.Context, permissions map[string][]string) ([]domain.Permission, error) {
	if len(permissions) == 0 {
		return nil, nil
	}

	var conditions []string
	var args []interface{}
	for resource, actions := range permissions {
		for _, action := range actions {
			conditions = append(conditions, "(resource = ? AND action = ?)")
			args = append(args, resource, action)
		}
	}

	var result []domain.Permission
	query := r.db.WithContext(ctx).Where(strings.Join(conditions, " OR "), args...)
	if err := query.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *permissionRepository) GetAll(ctx context.Context) ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.WithContext(ctx).Order("resource ASC, action ASC").Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) GetAvailablePermissions(ctx context.Context) ([]dto.ResourcePermissions, error) {
	var permissions []domain.Permission
	if err := r.db.WithContext(ctx).Order("resource ASC, action ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	resourceMap := make(map[string][]dto.PermissionItem)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], dto.PermissionItem{
			Action: perm.Action,
			Label:  perm.Label,
		})
	}

	var result []dto.ResourcePermissions
	for resource, perms := range resourceMap {
		result = append(result, dto.ResourcePermissions{
			Name:        resource,
			Permissions: perms,
		})
	}

	return result, nil
}

func (r *permissionRepository) BulkCreate(ctx context.Context, permissions []domain.Permission) error {
	return r.db.WithContext(ctx).Create(&permissions).Error
}
