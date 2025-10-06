package helper

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

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
