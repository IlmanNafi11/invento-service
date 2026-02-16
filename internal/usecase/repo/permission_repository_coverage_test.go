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

// TestPermissionRepository_Create_Success tests successful permission creation
func TestPermissionRepository_Create_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	permission := &domain.Permission{
		Resource: "users",
		Action:   "read",
		Label:    "Read users",
	}

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	err = permissionRepo.Create(ctx, permission)
	assert.NoError(t, err)
	assert.NotZero(t, permission.ID)
}

// TestPermissionRepository_GetByID_Success tests successful permission retrieval by ID
func TestPermissionRepository_GetByID_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	permission := &domain.Permission{
		Resource: "users",
		Action:   "read",
		Label:    "Read users",
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	result, err := permissionRepo.GetByID(ctx, permission.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, permission.ID, result.ID)
	assert.Equal(t, "users", result.Resource)
}

// TestPermissionRepository_GetByResourceAndAction_Success tests successful permission retrieval
func TestPermissionRepository_GetByResourceAndAction_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	permission := &domain.Permission{
		Resource: "users",
		Action:   "write",
		Label:    "Write users",
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	result, err := permissionRepo.GetByResourceAndAction(ctx, "users", "write")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "users", result.Resource)
	assert.Equal(t, "write", result.Action)
}

// TestPermissionRepository_GetAll_Success tests successful retrieval of all permissions
func TestPermissionRepository_GetAll_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	permissions := []domain.Permission{
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "users", Action: "write", Label: "Write users"},
		{Resource: "projects", Action: "read", Label: "Read projects"},
	}

	for _, perm := range permissions {
		err = db.Create(&perm).Error
		require.NoError(t, err)
	}

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	result, err := permissionRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

// TestPermissionRepository_GetAvailablePermissions_Success tests successful retrieval of available permissions
func TestPermissionRepository_GetAvailablePermissions_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	permissions := []domain.Permission{
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "users", Action: "write", Label: "Write users"},
		{Resource: "projects", Action: "read", Label: "Read projects"},
		{Resource: "projects", Action: "write", Label: "Write projects"},
	}

	for _, perm := range permissions {
		err = db.Create(&perm).Error
		require.NoError(t, err)
	}

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	result, err := permissionRepo.GetAvailablePermissions(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2) // 2 resources: users, projects

	// Verify structure
	for _, resourcePerm := range result {
		if resourcePerm.Name == "users" {
			assert.Len(t, resourcePerm.Permissions, 2)
		} else if resourcePerm.Name == "projects" {
			assert.Len(t, resourcePerm.Permissions, 2)
		}
	}
}

// TestPermissionRepository_BulkCreate_Success tests successful bulk creation
func TestPermissionRepository_BulkCreate_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	permissions := []domain.Permission{
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "users", Action: "write", Label: "Write users"},
		{Resource: "users", Action: "delete", Label: "Delete users"},
	}

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	err = permissionRepo.BulkCreate(ctx, permissions)
	assert.NoError(t, err)

	// Verify all created
	var count int64
	db.Model(&domain.Permission{}).Count(&count)
	assert.Equal(t, int64(3), count)
}

func TestPermissionRepository_GetAllByResourceActions_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	seed := []domain.Permission{
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "users", Action: "write", Label: "Write users"},
		{Resource: "projects", Action: "create", Label: "Create projects"},
	}
	for i := range seed {
		err = db.Create(&seed[i]).Error
		require.NoError(t, err)
	}

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	result, err := permissionRepo.GetAllByResourceActions(ctx, map[string][]string{
		"users":    {"read"},
		"projects": {"create"},
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	actions := map[string]bool{}
	for _, perm := range result {
		actions[perm.Resource+":"+perm.Action] = true
	}
	assert.True(t, actions["users:read"])
	assert.True(t, actions["projects:create"])
}

func TestPermissionRepository_GetAllByResourceActions_Empty(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	result, err := permissionRepo.GetAllByResourceActions(ctx, map[string][]string{})

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestPermissionRepository_GetAllByResourceActions_NotFound(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	err = db.Create(&domain.Permission{Resource: "users", Action: "read", Label: "Read users"}).Error
	require.NoError(t, err)

	permissionRepo := repo.NewPermissionRepository(db)
	ctx := context.Background()
	result, err := permissionRepo.GetAllByResourceActions(ctx, map[string][]string{
		"projects": {"delete"},
	})

	assert.NoError(t, err)
	assert.Empty(t, result)
}
