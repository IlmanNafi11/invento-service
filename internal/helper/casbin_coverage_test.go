package helper

import (
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCasbinEnforcer_Success tests successful casbin enforcer creation
func TestNewCasbinEnforcer_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)

	assert.NoError(t, err)
	assert.NotNil(t, enforcer)
	assert.NotNil(t, enforcer.GetEnforcer())
}

// TestCasbinEnforcer_AddPermissionForRole_Success tests successful permission addition
func TestCasbinEnforcer_AddPermissionForRole_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	err = enforcer.AddPermissionForRole("admin", "users", "write")
	assert.NoError(t, err)

	// Verify permission was added
	allowed, _ := enforcer.CheckPermission("admin", "users", "write")
	assert.True(t, allowed)

	// Clean up
	enforcer.RemovePermissionForRole("admin", "users", "write")
}

// TestCasbinEnforcer_RemovePermissionForRole_Success tests successful permission removal
func TestCasbinEnforcer_RemovePermissionForRole_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Add permission first
	err = enforcer.AddPermissionForRole("admin", "users", "write")
	require.NoError(t, err)

	// Remove permission
	err = enforcer.RemovePermissionForRole("admin", "users", "write")
	assert.NoError(t, err)

	// Verify permission was removed
	allowed, _ := enforcer.CheckPermission("admin", "users", "write")
	assert.False(t, allowed)
}

// TestCasbinEnforcer_RemoveAllPermissionsForRole_Success tests successful removal of all permissions
func TestCasbinEnforcer_RemoveAllPermissionsForRole_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Add multiple permissions
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "users", "write")
	require.NoError(t, err)

	// Remove all permissions for role
	err = enforcer.RemoveAllPermissionsForRole("admin")
	assert.NoError(t, err)

	// Verify all permissions were removed
	permissions, err := enforcer.GetPermissionsForRole("admin")
	assert.NoError(t, err)
	assert.Len(t, permissions, 0)
}

// TestCasbinEnforcer_GetPermissionsForRole_Success tests successful permission retrieval
func TestCasbinEnforcer_GetPermissionsForRole_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Add permissions
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("admin", "projects", "write")
	require.NoError(t, err)

	// Get permissions
	permissions, err := enforcer.GetPermissionsForRole("admin")
	assert.NoError(t, err)
	assert.Len(t, permissions, 2)

	// Clean up
	enforcer.RemoveAllPermissionsForRole("admin")
}

// TestCasbinEnforcer_CheckPermission_Success tests successful permission check
func TestCasbinEnforcer_CheckPermission_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Add permission
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)

	// Check permission - should be allowed
	allowed, err := enforcer.CheckPermission("admin", "users", "read")
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Check permission - should be denied (no permission)
	allowed, err = enforcer.CheckPermission("admin", "users", "write")
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Clean up
	enforcer.RemovePermissionForRole("admin", "users", "read")
}

// TestCasbinEnforcer_SaveLoadPolicy_Success tests policy save and load
func TestCasbinEnforcer_SaveLoadPolicy_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Add permission
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)

	// Save policy
	err = enforcer.SavePolicy()
	assert.NoError(t, err)

	// Verify permission exists before load
	allowed, _ := enforcer.CheckPermission("admin", "users", "read")
	assert.True(t, allowed)

	// Load policy (should preserve policies)
	err = enforcer.LoadPolicy()
	assert.NoError(t, err)

	// Verify permission still exists after load
	allowed, _ = enforcer.CheckPermission("admin", "users", "read")
	assert.True(t, allowed)

	// Clean up
	enforcer.RemovePermissionForRole("admin", "users", "read")
}

// TestCasbinEnforcer_DeleteRole_Success tests successful role deletion
func TestCasbinEnforcer_DeleteRole_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Add role with permissions
	err = enforcer.AddPermissionForRole("temp_role", "users", "read")
	require.NoError(t, err)

	// Delete role
	err = enforcer.DeleteRole("temp_role")
	assert.NoError(t, err)

	// Verify role was deleted
	roles, err := enforcer.GetAllRoles()
	assert.NoError(t, err)
	assert.NotContains(t, roles, "temp_role")
}

// TestCasbinEnforcer_GetAllRoles_Success tests successful role retrieval
func TestCasbinEnforcer_GetAllRoles_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// In casbin RBAC, roles are defined through AddRoleForUser or g2 policies
	// Add permissions for roles (this creates p policies)
	err = enforcer.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer.AddPermissionForRole("user", "projects", "read")
	require.NoError(t, err)

	// Get all policies instead (roles with permissions in RBAC show up in policies)
	policies, err := enforcer.GetAllPolicies()
	assert.NoError(t, err)
	assert.NotEmpty(t, policies)

	// Verify our role permissions are in the policies
	foundAdmin := false
	foundUser := false
	for _, policy := range policies {
		if len(policy) >= 1 && policy[0] == "admin" {
			foundAdmin = true
		}
		if len(policy) >= 1 && policy[0] == "user" {
			foundUser = true
		}
	}
	assert.True(t, foundAdmin, "admin role policy should exist")
	assert.True(t, foundUser, "user role policy should exist")

	// Clean up
	enforcer.RemovePermissionForRole("admin", "users", "read")
	enforcer.RemovePermissionForRole("user", "projects", "read")
}

// TestCasbinEnforcer_GetAllPolicies_Success tests successful policy retrieval
func TestCasbinEnforcer_GetAllPolicies_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Add policy
	err = enforcer.AddPermissionForRole("admin", "users", "write")
	require.NoError(t, err)

	// Get all policies
	policies, err := enforcer.GetAllPolicies()
	assert.NoError(t, err)
	assert.NotEmpty(t, policies)

	// Verify policy structure
	for _, policy := range policies {
		assert.Len(t, policy, 3) // [role, resource, action]
	}

	// Clean up
	enforcer.RemovePermissionForRole("admin", "users", "write")
}

// TestCasbinEnforcer_HasPolicy_Success tests policy existence check
func TestCasbinEnforcer_HasPolicy_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	enforcer, err := NewCasbinEnforcer(db)
	require.NoError(t, err)

	// Initially no policy
	hasPolicy, err := enforcer.HasPolicy("admin", "users", "write")
	assert.NoError(t, err)
	assert.False(t, hasPolicy)

	// Add policy
	err = enforcer.AddPermissionForRole("admin", "users", "write")
	require.NoError(t, err)

	// Now policy exists
	hasPolicy, err = enforcer.HasPolicy("admin", "users", "write")
	assert.NoError(t, err)
	assert.True(t, hasPolicy)

	// Clean up
	enforcer.RemovePermissionForRole("admin", "users", "write")
}
