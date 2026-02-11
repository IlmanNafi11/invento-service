package helper_test

import (
	"fiber-boiler-plate/internal/helper"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestCasbinEnforcer creates an in-memory Casbin enforcer for testing
func setupTestCasbinEnforcer(t *testing.T) *helper.CasbinEnforcer {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "failed to open test database")

	enforcer, err := helper.NewCasbinEnforcer(db)
	require.NoError(t, err, "failed to create casbin enforcer")

	return enforcer
}

// =============================================================================
// NewCasbinEnforcer Tests
// =============================================================================

func TestNewCasbinEnforcer_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	enforcer, err := helper.NewCasbinEnforcer(db)

	assert.NoError(t, err)
	assert.NotNil(t, enforcer)
	assert.NotNil(t, enforcer.GetEnforcer())
}

func TestNewCasbinEnforcer_WithExistingDB(t *testing.T) {
	// Create first enforcer and add policy
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	enforcer1, err := helper.NewCasbinEnforcer(db)
	require.NoError(t, err)

	err = enforcer1.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer1.SavePolicy()
	require.NoError(t, err)

	// Create second enforcer with same DB - should load existing policies
	enforcer2, err := helper.NewCasbinEnforcer(db)
	require.NoError(t, err)

	hasPolicy, err := enforcer2.HasPolicy("admin", "users", "read")
	assert.NoError(t, err)
	assert.True(t, hasPolicy)
}

// =============================================================================
// AddPermissionForRole Tests
// =============================================================================

func TestAddPermissionForRole_Success(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "projects", "create")

	assert.NoError(t, err)

	// Verify permission was added
	hasPolicy, err := enforcer.HasPolicy("admin", "projects", "create")
	assert.NoError(t, err)
	assert.True(t, hasPolicy)
}

func TestAddPermissionForRole_MultiplePolicies(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	policies := []struct {
		role     string
		resource string
		action   string
	}{
		{"admin", "users", "create"},
		{"admin", "users", "read"},
		{"admin", "users", "update"},
		{"admin", "users", "delete"},
		{"manager", "projects", "read"},
		{"manager", "projects", "update"},
	}

	for _, p := range policies {
		err := enforcer.AddPermissionForRole(p.role, p.resource, p.action)
		assert.NoError(t, err)
	}

	// Verify all policies were added
	for _, p := range policies {
		hasPolicy, err := enforcer.HasPolicy(p.role, p.resource, p.action)
		assert.NoError(t, err)
		assert.True(t, hasPolicy, "expected policy %s %s %s to exist", p.role, p.resource, p.action)
	}
}

func TestAddPermissionForRole_DuplicatePolicy(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// Add same policy twice
	err1 := enforcer.AddPermissionForRole("admin", "users", "read")
	assert.NoError(t, err1)

	err2 := enforcer.AddPermissionForRole("admin", "users", "read")
	assert.NoError(t, err2) // Casbin ignores duplicates without error
}

// =============================================================================
// RemovePermissionForRole Tests
// =============================================================================

func TestRemovePermissionForRole_Success(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// Add then remove
	err := enforcer.AddPermissionForRole("admin", "projects", "delete")
	require.NoError(t, err)

	err = enforcer.RemovePermissionForRole("admin", "projects", "delete")
	assert.NoError(t, err)

	// Verify removal
	hasPolicy, err := enforcer.HasPolicy("admin", "projects", "delete")
	assert.NoError(t, err)
	assert.False(t, hasPolicy)
}

func TestRemovePermissionForRole_NonExistent(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// Removing non-existent policy should not error
	err := enforcer.RemovePermissionForRole("nonexistent", "resource", "action")
	assert.NoError(t, err)
}

func TestRemovePermissionForRole_PartialMatch(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// Add multiple permissions for admin
	err := enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "users", "write")
	require.NoError(t, err)

	// Remove only one
	err = enforcer.RemovePermissionForRole("admin", "users", "read")
	assert.NoError(t, err)

	// Write should still exist
	hasPolicy, err := enforcer.HasPolicy("admin", "users", "write")
	assert.NoError(t, err)
	assert.True(t, hasPolicy)

	// Read should be gone
	hasPolicy, err = enforcer.HasPolicy("admin", "users", "read")
	assert.NoError(t, err)
	assert.False(t, hasPolicy)
}

// =============================================================================
// RemoveAllPermissionsForRole Tests
// =============================================================================

func TestRemoveAllPermissionsForRole_Success(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// Add multiple permissions
	err := enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "projects", "delete")
	require.NoError(t, err)

	// Remove all permissions for admin
	err = enforcer.RemoveAllPermissionsForRole("admin")
	assert.NoError(t, err)

	// Verify all removed
	permissions, err := enforcer.GetPermissionsForRole("admin")
	assert.NoError(t, err)
	assert.Empty(t, permissions)
}

func TestRemoveAllPermissionsForRole_DoesNotAffectOtherRoles(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// Add permissions for multiple roles
	err := enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("manager", "projects", "read")
	require.NoError(t, err)

	// Remove all for admin
	err = enforcer.RemoveAllPermissionsForRole("admin")
	assert.NoError(t, err)

	// Manager permissions should still exist
	hasPolicy, err := enforcer.HasPolicy("manager", "projects", "read")
	assert.NoError(t, err)
	assert.True(t, hasPolicy)
}

// =============================================================================
// CheckPermission Tests
// =============================================================================

func TestCheckPermission_Allowed(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)

	allowed, err := enforcer.CheckPermission("admin", "users", "read")

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermission_Denied_NoPolicy(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// No policies added
	allowed, err := enforcer.CheckPermission("guest", "users", "read")

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermission_Denied_WrongAction(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("manager", "projects", "read")
	require.NoError(t, err)

	// Check for different action
	allowed, err := enforcer.CheckPermission("manager", "projects", "delete")

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermission_Denied_WrongResource(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)

	// Check for different resource
	allowed, err := enforcer.CheckPermission("admin", "projects", "read")

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckPermission_MultipleRolesAndResources(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	// Setup policies
	err := enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("manager", "projects", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("viewer", "projects", "read")
	require.NoError(t, err)

	tests := []struct {
		name     string
		role     string
		resource string
		action   string
		expected bool
	}{
		{"admin can create users", "admin", "users", "create", true},
		{"admin can read users", "admin", "users", "read", true},
		{"admin cannot read projects", "admin", "projects", "read", false},
		{"manager can read projects", "manager", "projects", "read", true},
		{"manager cannot create users", "manager", "users", "create", false},
		{"viewer can read projects", "viewer", "projects", "read", true},
		{"viewer cannot delete projects", "viewer", "projects", "delete", false},
		{"nonexistent role denied", "nonexistent", "users", "read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.CheckPermission(tt.role, tt.resource, tt.action)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

// =============================================================================
// GetPermissionsForRole Tests
// =============================================================================

func TestGetPermissionsForRole_Success(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "projects", "delete")
	require.NoError(t, err)

	permissions, err := enforcer.GetPermissionsForRole("admin")

	assert.NoError(t, err)
	assert.Len(t, permissions, 3)
}

func TestGetPermissionsForRole_Empty(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	permissions, err := enforcer.GetPermissionsForRole("nonexistent")

	assert.NoError(t, err)
	assert.Empty(t, permissions)
}

func TestGetPermissionsForRole_DoesNotIncludeOtherRoles(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("manager", "projects", "read")
	require.NoError(t, err)

	permissions, err := enforcer.GetPermissionsForRole("admin")

	assert.NoError(t, err)
	assert.Len(t, permissions, 1)
	assert.Equal(t, "admin", permissions[0][0])
	assert.Equal(t, "users", permissions[0][1])
	assert.Equal(t, "create", permissions[0][2])
}

// =============================================================================
// GetRolesForUser Tests
// =============================================================================

func TestGetRolesForUser_Success(t *testing.T) {
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
	enforcer := setupTestCasbinEnforcer(t)

	roles, err := enforcer.GetRolesForUser("user-without-roles")

	assert.NoError(t, err)
	assert.Empty(t, roles)
}

func TestGetRolesForUser_MultipleUsers(t *testing.T) {
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
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-456", "developer")

	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user-456")
	assert.Contains(t, roles, "developer")
}

func TestAddRoleForUser_DuplicateRole(t *testing.T) {
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
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddRoleForUser("user-abc", "admin")
	require.NoError(t, err)

	err = enforcer.RemoveRoleForUser("user-abc", "admin")
	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user-abc")
	assert.Empty(t, roles)
}

func TestRemoveRoleForUser_KeepsOtherRoles(t *testing.T) {
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
	enforcer := setupTestCasbinEnforcer(t)

	users, err := enforcer.GetUsersForRole("empty-role")

	assert.NoError(t, err)
	assert.Empty(t, users)
}

// =============================================================================
// DeleteAllRolesForUser Tests
// =============================================================================

func TestDeleteAllRolesForUser_Success(t *testing.T) {
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
	enforcer := setupTestCasbinEnforcer(t)

	roles, err := enforcer.GetAllRoles()

	assert.NoError(t, err)
	assert.Empty(t, roles)
}

// =============================================================================
// GetAllPolicies Tests
// =============================================================================

func TestGetAllPolicies_Success(t *testing.T) {
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
	enforcer := setupTestCasbinEnforcer(t)

	policies, err := enforcer.GetAllPolicies()

	assert.NoError(t, err)
	assert.Empty(t, policies)
}

// =============================================================================
// HasPolicy Tests
// =============================================================================

func TestHasPolicy_Exists(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	err := enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)

	hasPolicy, err := enforcer.HasPolicy("admin", "users", "read")

	assert.NoError(t, err)
	assert.True(t, hasPolicy)
}

func TestHasPolicy_NotExists(t *testing.T) {
	enforcer := setupTestCasbinEnforcer(t)

	hasPolicy, err := enforcer.HasPolicy("nonexistent", "resource", "action")

	assert.NoError(t, err)
	assert.False(t, hasPolicy)
}

// =============================================================================
// SavePolicy and LoadPolicy Tests
// =============================================================================

func TestSavePolicy_Success(t *testing.T) {
	// Use shared cache for persistent in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	enforcer, err := helper.NewCasbinEnforcer(db)
	require.NoError(t, err)

	err = enforcer.AddPermissionForRole("admin", "users", "create")
	require.NoError(t, err)

	err = enforcer.SavePolicy()
	assert.NoError(t, err)
}

func TestLoadPolicy_Success(t *testing.T) {
	// Use shared cache for persistent in-memory database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	enforcer, err := helper.NewCasbinEnforcer(db)
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
	casbinEnforcer := setupTestCasbinEnforcer(t)

	enforcer := casbinEnforcer.GetEnforcer()

	assert.NotNil(t, enforcer)
}

// =============================================================================
// Integration Tests (User inherits role permissions)
// =============================================================================

func TestCheckPermission_UserInheritsRolePermissions(t *testing.T) {
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
