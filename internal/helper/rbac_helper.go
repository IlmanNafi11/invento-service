package helper

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
)

type RBACHelper struct {
	casbinEnforcer *CasbinEnforcer
}

func NewRBACHelper(casbinEnforcer *CasbinEnforcer) *RBACHelper {
	return &RBACHelper{
		casbinEnforcer: casbinEnforcer,
	}
}

func (rh *RBACHelper) SyncPermissionsToRole(roleName string, permissions map[string][]string, permissionRepo interface{}) ([]domain.RolePermissionDetail, int, error) {
	type PermissionRepo interface {
		GetByResourceAndAction(resource, action string) (*domain.Permission, error)
	}

	repo, ok := permissionRepo.(PermissionRepo)
	if !ok {
		return nil, 0, errors.New("invalid permission repository")
	}

	var permissionDetails []domain.RolePermissionDetail
	permissionCount := 0

	for resource, actions := range permissions {
		var resourceActions []string
		for _, action := range actions {
			_, err := repo.GetByResourceAndAction(resource, action)
			if err != nil {
				continue
			}

			if err := rh.casbinEnforcer.AddPermissionForRole(roleName, resource, action); err != nil {
				continue
			}

			resourceActions = append(resourceActions, action)
			permissionCount++
		}

		if len(resourceActions) > 0 {
			permissionDetails = append(permissionDetails, domain.RolePermissionDetail{
				Resource: resource,
				Actions:  resourceActions,
			})
		}
	}

	return permissionDetails, permissionCount, nil
}

func (rh *RBACHelper) CreateRolePermissions(roleID uint, permissions map[string][]string, permissionRepo interface{}, rolePermissionRepo interface{}) ([]domain.RolePermissionDetail, int, error) {
	type PermissionRepo interface {
		GetByResourceAndAction(resource, action string) (*domain.Permission, error)
	}

	type RolePermissionRepo interface {
		Create(rolePermission *domain.RolePermission) error
	}

	pRepo, ok := permissionRepo.(PermissionRepo)
	if !ok {
		return nil, 0, errors.New("invalid permission repository")
	}

	rpRepo, ok := rolePermissionRepo.(RolePermissionRepo)
	if !ok {
		return nil, 0, errors.New("invalid role permission repository")
	}

	var permissionDetails []domain.RolePermissionDetail
	permissionCount := 0

	for resource, actions := range permissions {
		var resourceActions []string
		for _, action := range actions {
			permission, err := pRepo.GetByResourceAndAction(resource, action)
			if err != nil {
				continue
			}

			rolePermission := &domain.RolePermission{
				RoleID:       roleID,
				PermissionID: permission.ID,
			}

			if err := rpRepo.Create(rolePermission); err != nil {
				continue
			}

			resourceActions = append(resourceActions, action)
			permissionCount++
		}

		if len(resourceActions) > 0 {
			permissionDetails = append(permissionDetails, domain.RolePermissionDetail{
				Resource: resource,
				Actions:  resourceActions,
			})
		}
	}

	return permissionDetails, permissionCount, nil
}

func (rh *RBACHelper) RemoveAllRolePermissions(roleID uint, roleName string, rolePermissionRepo interface{}) error {
	type RolePermissionRepo interface {
		DeleteByRoleID(roleID uint) error
	}

	repo, ok := rolePermissionRepo.(RolePermissionRepo)
	if !ok {
		return errors.New("invalid role permission repository")
	}

	if err := repo.DeleteByRoleID(roleID); err != nil {
		return errors.New("gagal menghapus permission lama")
	}

	if err := rh.casbinEnforcer.RemoveAllPermissionsForRole(roleName); err != nil {
		return errors.New("gagal menghapus policy casbin lama")
	}

	return nil
}

func (rh *RBACHelper) BuildRoleDetailResponse(role *domain.Role, permissions []domain.Permission) *domain.RoleDetailResponse {
	resourceMap := make(map[string][]string)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], perm.Action)
	}

	var permissionDetails []domain.RolePermissionDetail
	for resource, actions := range resourceMap {
		permissionDetails = append(permissionDetails, domain.RolePermissionDetail{
			Resource: resource,
			Actions:  actions,
		})
	}

	return &domain.RoleDetailResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		Permissions:      permissionDetails,
		JumlahPermission: len(permissions),
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}
}

func (rh *RBACHelper) ValidatePermissionFormat(permissions map[string][]string) error {
	if len(permissions) == 0 {
		return errors.New("permission tidak boleh kosong")
	}

	for resource, actions := range permissions {
		if resource == "" {
			return errors.New("nama resource tidak boleh kosong")
		}

		if len(actions) == 0 {
			return errors.New("action untuk resource " + resource + " tidak boleh kosong")
		}

		for _, action := range actions {
			if action == "" {
				return errors.New("action tidak boleh kosong untuk resource " + resource)
			}
		}
	}

	return nil
}

func (rh *RBACHelper) CheckUserPermission(roleName, resource, action string) (bool, error) {
	return rh.casbinEnforcer.CheckPermission(roleName, resource, action)
}

func (rh *RBACHelper) SavePolicy() error {
	return rh.casbinEnforcer.SavePolicy()
}
