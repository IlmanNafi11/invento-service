package rbac_test

import (
	"invento-service/internal/rbac"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GetRolesForUser Tests
// =============================================================================

func TestGetRolesForUser_Success(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-123", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-123", "manager")
	require.NoError(t, err)

	roles, err := enforcer.GetRolesForUser("user-123")

	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.Contains(t, roles, "admin")
	assert.Contains(t, roles, "manager")
}

func TestGetRolesForUser_NoRoles(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	roles, err := enforcer.GetRolesForUser("user-without-roles")

	assert.NoError(t, err)
	assert.Empty(t, roles)
}

func TestGetRolesForUser_MultipleUsers(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-1", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-2", "manager")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-3", "admin")
	require.NoError(t, err)

	roles1, _ := enforcer.GetRolesForUser("user-1")
	roles2, _ := enforcer.GetRolesForUser("user-2")
	roles3, _ := enforcer.GetRolesForUser("user-3")

	assert.Contains(t, roles1, "admin")
	assert.NotContains(t, roles1, "manager")
	assert.Contains(t, roles2, "manager")
	assert.NotContains(t, roles2, "admin")
	assert.Contains(t, roles3, "admin")
}

// =============================================================================
// AddRoleForUser Tests
// =============================================================================

func TestAddRoleForUser_Success(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-456", "developer")

	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user-456")
	assert.Contains(t, roles, "developer")
}

func TestAddRoleForUser_DuplicateRole(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err1 := enforcer.AddRoleForUser("user-789", "admin")
	assert.NoError(t, err1)

	err2 := enforcer.AddRoleForUser("user-789", "admin")
	assert.NoError(t, err2) // Should not error on duplicate

	roles, _ := enforcer.GetRolesForUser("user-789")
	assert.Len(t, roles, 1) // Should still have only one admin role
}

// =============================================================================
// RemoveRoleForUser Tests
// =============================================================================

func TestRemoveRoleForUser_Success(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-abc", "admin")
	require.NoError(t, err)

	err = enforcer.RemoveRoleForUser("user-abc", "admin")
	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user-abc")
	assert.Empty(t, roles)
}

func TestRemoveRoleForUser_KeepsOtherRoles(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-def", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-def", "manager")
	require.NoError(t, err)

	err = enforcer.RemoveRoleForUser("user-def", "admin")
	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user-def")
	assert.Len(t, roles, 1)
	assert.Contains(t, roles, "manager")
}

// =============================================================================
// GetUsersForRole Tests
// =============================================================================

func TestGetUsersForRole_Success(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-1", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-2", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-3", "manager")
	require.NoError(t, err)

	users, err := enforcer.GetUsersForRole("admin")

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Contains(t, users, "user-1")
	assert.Contains(t, users, "user-2")
}

func TestGetUsersForRole_NoUsers(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	users, err := enforcer.GetUsersForRole("empty-role")

	assert.NoError(t, err)
	assert.Empty(t, users)
}

// =============================================================================
// DeleteAllRolesForUser Tests
// =============================================================================

func TestDeleteAllRolesForUser_Success(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-xyz", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-xyz", "manager")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-xyz", "developer")
	require.NoError(t, err)

	err = enforcer.DeleteAllRolesForUser("user-xyz")
	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user-xyz")
	assert.Empty(t, roles)
}

func TestDeleteAllRolesForUser_DoesNotAffectOtherUsers(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-1", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-2", "admin")
	require.NoError(t, err)

	err = enforcer.DeleteAllRolesForUser("user-1")
	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user-2")
	assert.Contains(t, roles, "admin")
}

// =============================================================================
// DeleteRole Tests
// =============================================================================

func TestDeleteRole_Success(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	// Add role with permissions and users
	err := enforcer.AddPermissionForRole("test-role", "resource", "action")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-1", "test-role")
	require.NoError(t, err)

	err = enforcer.DeleteRole("test-role")
	assert.NoError(t, err)

	// Role should be removed from users
	roles, _ := enforcer.GetRolesForUser("user-1")
	assert.NotContains(t, roles, "test-role")
}

// =============================================================================
// GetAllRoles Tests
// =============================================================================

func TestGetAllRoles_WithGroupingPolicies(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-1", "admin")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-2", "manager")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-3", "developer")
	require.NoError(t, err)

	roles, err := enforcer.GetAllRoles()

	assert.NoError(t, err)
	assert.Len(t, roles, 3)
	assert.Contains(t, roles, "admin")
	assert.Contains(t, roles, "manager")
	assert.Contains(t, roles, "developer")
}

func TestGetAllRoles_Empty(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	roles, err := enforcer.GetAllRoles()

	assert.NoError(t, err)
	assert.Empty(t, roles)
}

// =============================================================================
// GetAllPolicies Tests
// =============================================================================

func TestGetAllPolicies_Success(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("manager", "projects", "read")
	require.NoError(t, err)

	policies, err := enforcer.GetAllPolicies()

	assert.NoError(t, err)
	assert.Len(t, policies, 3)
}

func TestGetAllPolicies_Empty(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	policies, err := enforcer.GetAllPolicies()

	assert.NoError(t, err)
	assert.Empty(t, policies)
}

// =============================================================================
// HasPolicy Tests
// =============================================================================

func TestHasPolicy_Exists(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)

	hasPolicy, err := enforcer.HasPolicy("admin", "users", "read")

	assert.NoError(t, err)
	assert.True(t, hasPolicy)
}

func TestHasPolicy_NotExists(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	hasPolicy, err := enforcer.HasPolicy("nonexistent", "resource", "action")

	assert.NoError(t, err)
	assert.False(t, hasPolicy)
}

// =============================================================================
// SavePolicy and LoadPolicy Tests
// =============================================================================

func TestSavePolicy_Success(t *testing.T) {
	t.Parallel()
	// Use shared cache with unique name to avoid conflicts with parallel tests
	db := createTestDB(t, "file:save_policy?mode=memory&cache=shared")

	enforcer, err := rbac.NewCasbinEnforcer(db)
	require.NoError(t, err)

	err = enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)

	err = enforcer.SavePolicy()
	assert.NoError(t, err)
}

func TestLoadPolicy_Success(t *testing.T) {
	t.Parallel()
	// Use shared cache with unique name to avoid conflicts with parallel tests
	db := createTestDB(t, "file:load_policy?mode=memory&cache=shared")

	enforcer, err := rbac.NewCasbinEnforcer(db)
	require.NoError(t, err)

	err = enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)
	err = enforcer.SavePolicy()
	require.NoError(t, err)

	err = enforcer.LoadPolicy()
	assert.NoError(t, err)

	hasPolicy, _ := enforcer.HasPolicy("admin", "users", "create")
	assert.True(t, hasPolicy)
}

// =============================================================================
// GetEnforcer Tests
// =============================================================================

func TestGetEnforcer_ReturnsValidEnforcer(t *testing.T) {
	t.Parallel()
	casbinEnforcer := setupTestCasbinEnforcer(t)

	enforcer := casbinEnforcer.GetEnforcer()

	assert.NotNil(t, enforcer)
}

// =============================================================================
// Integration Tests (User inherits role permissions)
// =============================================================================

func TestCheckPermission_UserInheritsRolePermissions(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	// Add permission to role
	err := enforcer.AddPermissionForRole("admin", "users", "delete")
	require.NoError(t, err)

	// Assign role to user
	err = enforcer.AddRoleForUser("user-admin-1", "admin")
	require.NoError(t, err)

	// User should have permission through role
	allowed, err := enforcer.CheckPermission("user-admin-1", "users", "delete")

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermission_UserWithoutRoleDenied(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	// Add permission to role
	err := enforcer.AddPermissionForRole("admin", "users", "delete")
	require.NoError(t, err)

	// User without role should be denied
	allowed, err := enforcer.CheckPermission("user-without-role", "users", "delete")

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermission_UserMultipleRoles(t *testing.T) {
	t.Parallel()
	enforcer := setupTestCasbinEnforcer(t)

	// Add permissions to different roles
	err := enforcer.AddPermissionForRole("reader", "docs", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("writer", "docs", "write")
	require.NoError(t, err)

	// Assign both roles to user
	err = enforcer.AddRoleForUser("user-multi", "reader")
	require.NoError(t, err)
	err = enforcer.AddRoleForUser("user-multi", "writer")
	require.NoError(t, err)

	// User should have both permissions
	canRead, _ := enforcer.CheckPermission("user-multi", "docs", "read")
	canWrite, _ := enforcer.CheckPermission("user-multi", "docs", "write")

	assert.True(t, canRead)
	assert.True(t, canWrite)
}
