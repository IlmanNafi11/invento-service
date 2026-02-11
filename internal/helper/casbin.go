package helper

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// CasbinEnforcerInterface defines the interface for Casbin enforcer operations
type CasbinEnforcerInterface interface {
	AddPermissionForRole(roleName, resource, action string) error
	RemovePermissionForRole(roleName, resource, action string) error
	RemoveAllPermissionsForRole(roleName string) error
	GetPermissionsForRole(roleName string) ([][]string, error)
	CheckPermission(roleName, resource, action string) (bool, error)
	SavePolicy() error
	LoadPolicy() error
	DeleteRole(roleName string) error
	GetEnforcer() *casbin.Enforcer
	GetAllRoles() ([]string, error)
	GetAllPolicies() ([][]string, error)
	HasPolicy(roleName, resource, action string) (bool, error)
	AddRoleForUser(userID, roleName string) error
	RemoveRoleForUser(userID, roleName string) error
	GetRolesForUser(userID string) ([]string, error)
	GetUsersForRole(roleName string) ([]string, error)
	DeleteAllRolesForUser(userID string) error
}

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

// Ensure CasbinEnforcer implements CasbinEnforcerInterface
var _ CasbinEnforcerInterface = (*CasbinEnforcer)(nil)

func NewCasbinEnforcer(db *gorm.DB) (*CasbinEnforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat casbin adapter: %w", err)
	}

	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := casbinmodel.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat casbin model: %w", err)
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat casbin enforcer: %w", err)
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("gagal load policy: %w", err)
	}

	return &CasbinEnforcer{enforcer: e}, nil
}

func (ce *CasbinEnforcer) AddPermissionForRole(roleName, resource, action string) error {
	_, err := ce.enforcer.AddPolicy(roleName, resource, action)
	if err != nil {
		return fmt.Errorf("gagal menambahkan permission: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) RemovePermissionForRole(roleName, resource, action string) error {
	_, err := ce.enforcer.RemovePolicy(roleName, resource, action)
	if err != nil {
		return fmt.Errorf("gagal menghapus permission: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) RemoveAllPermissionsForRole(roleName string) error {
	_, err := ce.enforcer.RemoveFilteredPolicy(0, roleName)
	if err != nil {
		return fmt.Errorf("gagal menghapus semua permission untuk role: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) GetPermissionsForRole(roleName string) ([][]string, error) {
	permissions, err := ce.enforcer.GetFilteredPolicy(0, roleName)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil permissions: %w", err)
	}
	return permissions, nil
}

func (ce *CasbinEnforcer) CheckPermission(roleName, resource, action string) (bool, error) {
	allowed, err := ce.enforcer.Enforce(roleName, resource, action)
	if err != nil {
		return false, fmt.Errorf("gagal memeriksa permission: %w", err)
	}
	return allowed, nil
}

func (ce *CasbinEnforcer) SavePolicy() error {
	if err := ce.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("gagal menyimpan policy: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) LoadPolicy() error {
	if err := ce.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("gagal load policy: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) DeleteRole(roleName string) error {
	_, err := ce.enforcer.DeleteRole(roleName)
	if err != nil {
		return fmt.Errorf("gagal menghapus role: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) GetEnforcer() *casbin.Enforcer {
	return ce.enforcer
}

func (ce *CasbinEnforcer) GetAllRoles() ([]string, error) {
	roles, err := ce.enforcer.GetAllRoles()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil semua role: %w", err)
	}
	return roles, nil
}

func (ce *CasbinEnforcer) GetAllPolicies() ([][]string, error) {
	policies, err := ce.enforcer.GetPolicy()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil semua policy: %w", err)
	}
	return policies, nil
}

func (ce *CasbinEnforcer) HasPolicy(roleName, resource, action string) (bool, error) {
	hasPolicy, err := ce.enforcer.HasPolicy(roleName, resource, action)
	if err != nil {
		return false, fmt.Errorf("gagal memeriksa policy: %w", err)
	}
	return hasPolicy, nil
}

func (ce *CasbinEnforcer) AddRoleForUser(userID, roleName string) error {
	_, err := ce.enforcer.AddGroupingPolicy(userID, roleName)
	if err != nil {
		return fmt.Errorf("gagal menambahkan role untuk user: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) RemoveRoleForUser(userID, roleName string) error {
	_, err := ce.enforcer.RemoveGroupingPolicy(userID, roleName)
	if err != nil {
		return fmt.Errorf("gagal menghapus role dari user: %w", err)
	}
	return nil
}

func (ce *CasbinEnforcer) GetRolesForUser(userID string) ([]string, error) {
	roles, err := ce.enforcer.GetRolesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil roles untuk user: %w", err)
	}
	return roles, nil
}

func (ce *CasbinEnforcer) GetUsersForRole(roleName string) ([]string, error) {
	users, err := ce.enforcer.GetUsersForRole(roleName)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil users untuk role: %w", err)
	}
	return users, nil
}

func (ce *CasbinEnforcer) DeleteAllRolesForUser(userID string) error {
	_, err := ce.enforcer.DeleteRolesForUser(userID)
	if err != nil {
		return fmt.Errorf("gagal menghapus semua role dari user: %w", err)
	}
	return nil
}
