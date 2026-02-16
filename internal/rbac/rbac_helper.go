package rbac

import (
	"context"
	"errors"

	"invento-service/internal/domain"
	"invento-service/internal/dto"
)

// RBACPermissionRepository defines the permission repo methods needed by RBACHelper.
// This avoids importing repo package (which would create import cycles with testing).
type RBACPermissionRepository interface {
	GetAllByResourceActions(ctx context.Context, permissions map[string][]string) ([]domain.Permission, error)
}

// RBACRolePermissionRepository defines the role-permission repo methods needed by RBACHelper.
type RBACRolePermissionRepository interface {
	BulkCreate(ctx context.Context, rolePermissions []domain.RolePermission) error
	DeleteByRoleID(ctx context.Context, roleID uint) error
}

type RBACHelper struct {
	casbinEnforcer CasbinEnforcerInterface
}

func NewRBACHelper(casbinEnforcer CasbinEnforcerInterface) *RBACHelper {
	return &RBACHelper{
		casbinEnforcer: casbinEnforcer,
	}
}

func (rh *RBACHelper) SetRolePermissions(
	ctx context.Context,
	roleID uint,
	roleName string,
	permissions map[string][]string,
	permissionRepo RBACPermissionRepository,
	rolePermissionRepo RBACRolePermissionRepository,
) ([]dto.RolePermissionDetail, int, error) {
	dbPermissions, err := permissionRepo.GetAllByResourceActions(ctx, permissions)
	if err != nil {
		return nil, 0, errors.New("gagal mengambil data permission")
	}

	permLookup := make(map[string]*domain.Permission)
	for i := range dbPermissions {
		key := dbPermissions[i].Resource + ":" + dbPermissions[i].Action
		permLookup[key] = &dbPermissions[i]
	}

	var rolePermissions []domain.RolePermission
	var permissionDetails []dto.RolePermissionDetail
	permissionCount := 0

	for resource, actions := range permissions {
		var resourceActions []string
		for _, action := range actions {
			key := resource + ":" + action
			perm, exists := permLookup[key]
			if !exists {
				continue
			}

			rolePermissions = append(rolePermissions, domain.RolePermission{
				RoleID:       roleID,
				PermissionID: perm.ID,
			})

			if err := rh.casbinEnforcer.AddPermissionForRole(roleName, resource, action); err != nil {
				continue
			}

			resourceActions = append(resourceActions, action)
			permissionCount++
		}

		if len(resourceActions) > 0 {
			permissionDetails = append(permissionDetails, dto.RolePermissionDetail{
				Resource: resource,
				Actions:  resourceActions,
			})
		}
	}

	if len(rolePermissions) > 0 {
		if err := rolePermissionRepo.BulkCreate(ctx, rolePermissions); err != nil {
			return nil, 0, errors.New("gagal membuat role permission")
		}
	}

	return permissionDetails, permissionCount, nil
}

func (rh *RBACHelper) RemoveAllRolePermissions(ctx context.Context, roleID uint, roleName string, rolePermissionRepo RBACRolePermissionRepository) error {
	if err := rolePermissionRepo.DeleteByRoleID(ctx, roleID); err != nil {
		return errors.New("gagal menghapus permission lama")
	}

	if err := rh.casbinEnforcer.RemoveAllPermissionsForRole(roleName); err != nil {
		return errors.New("gagal menghapus policy casbin lama")
	}

	return nil
}

func (rh *RBACHelper) BuildRoleDetailResponse(role *domain.Role, permissions []domain.Permission) *dto.RoleDetailResponse {
	resourceMap := make(map[string][]string)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], perm.Action)
	}

	var permissionDetails []dto.RolePermissionDetail
	for resource, actions := range resourceMap {
		permissionDetails = append(permissionDetails, dto.RolePermissionDetail{
			Resource: resource,
			Actions:  actions,
		})
	}

	return &dto.RoleDetailResponse{
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
