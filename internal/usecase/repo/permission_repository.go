package repo

import (
	"invento-service/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(permission *domain.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) GetByID(id uint) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByResourceAndAction(resource, action string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.Where("resource = ? AND action = ?", resource, action).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetAllByResourceActions(permissions map[string][]string) ([]domain.Permission, error) {
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
	query := r.db.Where(strings.Join(conditions, " OR "), args...)
	if err := query.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *permissionRepository) GetAll() ([]domain.Permission, error) {
	var permissions []domain.Permission
	err := r.db.Order("resource ASC, action ASC").Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) GetAvailablePermissions() ([]domain.ResourcePermissions, error) {
	var permissions []domain.Permission
	if err := r.db.Order("resource ASC, action ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}

	resourceMap := make(map[string][]domain.PermissionItem)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], domain.PermissionItem{
			Action: perm.Action,
			Label:  perm.Label,
		})
	}

	var result []domain.ResourcePermissions
	for resource, perms := range resourceMap {
		result = append(result, domain.ResourcePermissions{
			Name:        resource,
			Permissions: perms,
		})
	}

	return result, nil
}

func (r *permissionRepository) BulkCreate(permissions []domain.Permission) error {
	return r.db.Create(&permissions).Error
}
