package rbac_test

import (
	"invento-service/internal/rbac"
	"testing"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// createTestDB opens an in-memory SQLite database and pre-creates
// the casbin_rule table required by NewCasbinEnforcer (which disables
// auto-migrate in production).
func createTestDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "failed to open test database")

	// Pre-create the casbin_rule table because production code calls
	// TurnOffAutoMigrate before NewAdapterByDB.
	require.NoError(t, db.AutoMigrate(&gormadapter.CasbinRule{}),
		"failed to create casbin_rule table")

	return db
}

// setupTestCasbinEnforcer creates an in-memory Casbin enforcer for testing
func setupTestCasbinEnforcer(t *testing.T) *rbac.CasbinEnforcer {
	t.Helper()

	db := createTestDB(t, ":memory:")

	enforcer, err := rbac.NewCasbinEnforcer(db)
	require.NoError(t, err, "failed to create casbin enforcer")

	return enforcer
}

// =============================================================================
// NewCasbinEnforcer Tests
// =============================================================================

func TestNewCasbinEnforcer_Success(t *testing.T) {
	db := createTestDB(t, ":memory:")

	enforcer, err := rbac.NewCasbinEnforcer(db)

	assert.NoError(t, err)
	assert.NotNil(t, enforcer)
	assert.NotNil(t, enforcer.GetEnforcer())
}

func TestNewCasbinEnforcer_WithExistingDB(t *testing.T) {
	// Create first enforcer and add policy
	db := createTestDB(t, "file:existing_db?mode=memory&cache=shared")

	enforcer1, err := rbac.NewCasbinEnforcer(db)
	require.NoError(t, err)

	err = enforcer1.AddPermissionForRole("admin", "users", "read")
	require.NoError(t, err)
	err = enforcer1.SavePolicy()
	require.NoError(t, err)

	// Create second enforcer with same DB - should load existing policies
	enforcer2, err := rbac.NewCasbinEnforcer(db)
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
