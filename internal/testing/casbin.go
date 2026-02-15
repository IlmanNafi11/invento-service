package testing

import (
	"invento-service/internal/helper"
	"fmt"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

const CasbinModelText = `[request_definition]
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

type TestCasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func NewTestCasbinEnforcer() (helper.CasbinEnforcerInterface, error) {
	m, err := casbinmodel.NewModelFromString(CasbinModelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	return &TestCasbinEnforcer{enforcer: e}, nil
}

func NewTestCasbinEnforcerWithPolicies(policies [][]string) (helper.CasbinEnforcerInterface, error) {
	te, err := NewTestCasbinEnforcer()
	if err != nil {
		return nil, err
	}

	testEnforcer, ok := te.(*TestCasbinEnforcer)
	if !ok {
		return nil, fmt.Errorf("invalid test casbin enforcer type")
	}

	for _, policy := range policies {
		if len(policy) >= 3 {
			if _, err := testEnforcer.enforcer.AddPolicy(policy[0], policy[1], policy[2]); err != nil {
				return nil, fmt.Errorf("failed to add policy %v: %w", policy, err)
			}
		}
	}

	return testEnforcer, nil
}

func NewTestCasbinEnforcerFromFile(policyPath string) (helper.CasbinEnforcerInterface, error) {
	m, err := casbinmodel.NewModelFromString(CasbinModelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	adapter := fileadapter.NewAdapter(policyPath)

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	return &TestCasbinEnforcer{enforcer: e}, nil
}

func (te *TestCasbinEnforcer) AddPermissionForRole(roleName, resource, action string) error {
	_, err := te.enforcer.AddPolicy(roleName, resource, action)
	if err != nil {
		return fmt.Errorf("failed to add permission: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) RemovePermissionForRole(roleName, resource, action string) error {
	_, err := te.enforcer.RemovePolicy(roleName, resource, action)
	if err != nil {
		return fmt.Errorf("failed to remove permission: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) RemoveAllPermissionsForRole(roleName string) error {
	_, err := te.enforcer.RemoveFilteredPolicy(0, roleName)
	if err != nil {
		return fmt.Errorf("failed to remove all permissions for role: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) GetPermissionsForRole(roleName string) ([][]string, error) {
	permissions, err := te.enforcer.GetFilteredPolicy(0, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	return permissions, nil
}

func (te *TestCasbinEnforcer) CheckPermission(roleName, resource, action string) (bool, error) {
	allowed, err := te.enforcer.Enforce(roleName, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return allowed, nil
}

func (te *TestCasbinEnforcer) SavePolicy() error {
	if err := te.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save policy: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) LoadPolicy() error {
	if err := te.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load policy: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) DeleteRole(roleName string) error {
	_, err := te.enforcer.DeleteRole(roleName)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) GetEnforcer() *casbin.Enforcer {
	return te.enforcer
}

func (te *TestCasbinEnforcer) GetAllRoles() ([]string, error) {
	roles, err := te.enforcer.GetAllRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	return roles, nil
}

func (te *TestCasbinEnforcer) GetAllPolicies() ([][]string, error) {
	policies, err := te.enforcer.GetPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to get all policies: %w", err)
	}
	return policies, nil
}

func (te *TestCasbinEnforcer) HasPolicy(roleName, resource, action string) (bool, error) {
	hasPolicy, err := te.enforcer.HasPolicy(roleName, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to check policy: %w", err)
	}
	return hasPolicy, nil
}

func (te *TestCasbinEnforcer) AddRoleForUser(userID, roleName string) error {
	_, err := te.enforcer.AddGroupingPolicy(userID, roleName)
	if err != nil {
		return fmt.Errorf("failed to add role for user: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) RemoveRoleForUser(userID, roleName string) error {
	_, err := te.enforcer.RemoveGroupingPolicy(userID, roleName)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) GetRolesForUser(userID string) ([]string, error) {
	roles, err := te.enforcer.GetRolesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}
	return roles, nil
}

func (te *TestCasbinEnforcer) GetUsersForRole(roleName string) ([]string, error) {
	users, err := te.enforcer.GetUsersForRole(roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}
	return users, nil
}

func (te *TestCasbinEnforcer) DeleteAllRolesForUser(userID string) error {
	_, err := te.enforcer.DeleteRolesForUser(userID)
	if err != nil {
		return fmt.Errorf("failed to delete all roles from user: %w", err)
	}
	return nil
}

func (te *TestCasbinEnforcer) Reset() error {
	te.enforcer.ClearPolicy()
	return nil
}

var _ helper.CasbinEnforcerInterface = (*TestCasbinEnforcer)(nil)
