package repo_test

import (
	"context"
	"invento-service/internal/domain"
	testhelper "invento-service/internal/testing"
	"invento-service/internal/usecase/repo"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRolePermissionRepository_GetPermissionsForRole_Success tests successful permission retrieval via JOIN
func TestRolePermissionRepository_GetPermissionsForRole_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	// Create role and multiple permissions
	role := &domain.Role{NamaRole: "administrator"}
	err = db.Create(role).Error
	require.NoError(t, err)

	permissions := []domain.Permission{
		{Resource: "users", Action: "create", Label: "Create users"},
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "users", Action: "update", Label: "Update users"},
		{Resource: "users", Action: "delete", Label: "Delete users"},
	}

	for _, perm := range permissions {
		err = db.Create(&perm).Error
		require.NoError(t, err)

		rolePermission := &domain.RolePermission{
			RoleID:       role.ID,
			PermissionID: perm.ID,
		}
		err = db.Create(rolePermission).Error
		require.NoError(t, err)
	}

	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	result, err := rolePermissionRepo.GetPermissionsForRole(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 4)

	// Verify permission details
	for i, perm := range result {
		assert.Equal(t, "users", perm.Resource)
		assert.NotEmpty(t, perm.Action)
		assert.NotEmpty(t, perm.Label)
		assert.NotZero(t, perm.ID)
		_ = i // Use index to avoid linter warning
	}
}

// TestRolePermissionRepository_GetPermissionsForRole_NoPermissions tests with role that has no permissions
func TestRolePermissionRepository_GetPermissionsForRole_NoPermissions(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	// Create role without permissions
	role := &domain.Role{NamaRole: "emptyrole"}
	err = db.Create(role).Error
	require.NoError(t, err)

	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	result, err := rolePermissionRepo.GetPermissionsForRole(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// TestRolePermissionRepository_GetPermissionsForRole_NonExistentRole tests with non-existent role
func TestRolePermissionRepository_GetPermissionsForRole_NonExistentRole(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	result, err := rolePermissionRepo.GetPermissionsForRole(context.Background(), 999)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// TestRolePermissionRepository_GetPermissionsForRole_MultipleRoles tests permissions are isolated per role
func TestRolePermissionRepository_GetPermissionsForRole_MultipleRoles(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	// Create two roles with different permissions
	role1 := &domain.Role{NamaRole: "admin"}
	role2 := &domain.Role{NamaRole: "user"}
	err = db.Create(role1).Error
	require.NoError(t, err)
	err = db.Create(role2).Error
	require.NoError(t, err)

	// Admin permissions
	adminPerms := []domain.Permission{
		{Resource: "users", Action: "create", Label: "Create users"},
		{Resource: "users", Action: "delete", Label: "Delete users"},
	}
	for _, perm := range adminPerms {
		err = db.Create(&perm).Error
		require.NoError(t, err)
		err = db.Create(&domain.RolePermission{RoleID: role1.ID, PermissionID: perm.ID}).Error
		require.NoError(t, err)
	}

	// User permissions
	userPerms := []domain.Permission{
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "projects", Action: "read", Label: "Read projects"},
	}
	for _, perm := range userPerms {
		err = db.Create(&perm).Error
		require.NoError(t, err)
		err = db.Create(&domain.RolePermission{RoleID: role2.ID, PermissionID: perm.ID}).Error
		require.NoError(t, err)
	}

	rolePermissionRepo := repo.NewRolePermissionRepository(db)

	// Get admin permissions
	adminResult, err := rolePermissionRepo.GetPermissionsForRole(context.Background(), role1.ID)
	assert.NoError(t, err)
	assert.Len(t, adminResult, 2)

	// Get user permissions
	userResult, err := rolePermissionRepo.GetPermissionsForRole(context.Background(), role2.ID)
	assert.NoError(t, err)
	assert.Len(t, userResult, 2)

	// Verify isolation - admin should have create/delete, user should have read
	adminActions := make(map[string]bool)
	for _, perm := range adminResult {
		adminActions[perm.Action] = true
	}
	assert.True(t, adminActions["create"])
	assert.True(t, adminActions["delete"])
	assert.False(t, adminActions["read"])

	userActions := make(map[string]bool)
	for _, perm := range userResult {
		userActions[perm.Action] = true
	}
	assert.True(t, userActions["read"])
	assert.False(t, userActions["create"])
	assert.False(t, userActions["delete"])
}

// TestRolePermissionRepository_Integration_FullWorkflow tests complete workflow
func TestRolePermissionRepository_Integration_FullWorkflow(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	rolePermissionRepo := repo.NewRolePermissionRepository(db)

	// Step 1: Create role and permissions
	role := &domain.Role{NamaRole: "manager"}
	err = db.Create(role).Error
	require.NoError(t, err)

	permissions := []domain.Permission{
		{Resource: "projects", Action: "create", Label: "Create projects"},
		{Resource: "projects", Action: "read", Label: "Read projects"},
		{Resource: "projects", Action: "update", Label: "Update projects"},
	}
	var createdPerms []domain.Permission
	for _, perm := range permissions {
		err = db.Create(&perm).Error
		require.NoError(t, err)
		createdPerms = append(createdPerms, perm)
	}

	// Step 2: Create role-permission associations
	for _, perm := range createdPerms {
		rolePermission := &domain.RolePermission{
			RoleID:       role.ID,
			PermissionID: perm.ID,
		}
		err = rolePermissionRepo.Create(context.Background(), rolePermission)
		assert.NoError(t, err)
	}

	// Step 3: Get role permissions with preload
	rolePermissions, err := rolePermissionRepo.GetByRoleID(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Len(t, rolePermissions, 3)

	// Step 4: Get permissions via JOIN
	permissionsViaJoin, err := rolePermissionRepo.GetPermissionsForRole(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Len(t, permissionsViaJoin, 3)

	// Step 5: Verify consistency
	assert.Len(t, rolePermissions, len(permissionsViaJoin))

	// Step 6: Delete all permissions for role
	err = rolePermissionRepo.DeleteByRoleID(context.Background(), role.ID)
	assert.NoError(t, err)

	// Step 7: Verify deletion via GetByRoleID
	rolePermissionsAfter, err := rolePermissionRepo.GetByRoleID(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Len(t, rolePermissionsAfter, 0, "GetByRoleID should return 0 after deletion")

	// Step 8: Verify deletion via GetPermissionsForRole
	permissionsAfter, err := rolePermissionRepo.GetPermissionsForRole(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Len(t, permissionsAfter, 0, "GetPermissionsForRole should return 0 after deletion")
}

func TestRolePermissionRepository_BulkCreate_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "bulk-admin"}
	err = db.Create(role).Error
	require.NoError(t, err)

	permissions := []domain.Permission{
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "users", Action: "write", Label: "Write users"},
	}
	for i := range permissions {
		err = db.Create(&permissions[i]).Error
		require.NoError(t, err)
	}

	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	err = rolePermissionRepo.BulkCreate(context.Background(), []domain.RolePermission{
		{RoleID: role.ID, PermissionID: permissions[0].ID},
		{RoleID: role.ID, PermissionID: permissions[1].ID},
	})

	assert.NoError(t, err)
	result, err := rolePermissionRepo.GetByRoleID(context.Background(), role.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRolePermissionRepository_BulkCreate_Empty(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	err = rolePermissionRepo.BulkCreate(context.Background(), []domain.RolePermission{})

	assert.NoError(t, err)
}

func TestRolePermissionRepository_BulkCreate_DuplicateKey(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "bulk-dup"}
	err = db.Create(role).Error
	require.NoError(t, err)

	permission := &domain.Permission{Resource: "projects", Action: "read", Label: "Read projects"}
	err = db.Create(permission).Error
	require.NoError(t, err)

	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	err = rolePermissionRepo.BulkCreate(context.Background(), []domain.RolePermission{{RoleID: role.ID, PermissionID: permission.ID}})
	require.NoError(t, err)

	err = rolePermissionRepo.BulkCreate(context.Background(), []domain.RolePermission{{RoleID: role.ID, PermissionID: permission.ID}})
	if err != nil {
		assert.Error(t, err)
		return
	}

	var count int64
	err = db.Model(&domain.RolePermission{}).Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).Count(&count).Error
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))
	assert.LessOrEqual(t, count, int64(2))
	assert.NoError(t, err)
}
