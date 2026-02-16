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

// TestRoleRepository_Create_Success tests successful role creation
func TestRoleRepository_Create_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{
		NamaRole: "editor",
	}

	roleRepo := repo.NewRoleRepository(db)
	ctx := context.Background()
	err = roleRepo.Create(ctx, role)
	assert.NoError(t, err)
	assert.NotZero(t, role.ID)

	// Verify role was created
	var foundRole domain.Role
	err = db.First(&foundRole, role.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "editor", foundRole.NamaRole)
}

// TestRoleRepository_GetByID_Success tests successful role retrieval by ID
func TestRoleRepository_GetByID_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{
		NamaRole: "admin",
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	roleRepo := repo.NewRoleRepository(db)
	ctx := context.Background()
	result, err := roleRepo.GetByID(ctx, role.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, role.ID, result.ID)
	assert.Equal(t, "admin", result.NamaRole)
}

// TestRoleRepository_GetByName_Success tests successful role retrieval by name
func TestRoleRepository_GetByName_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{
		NamaRole: "moderator",
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	roleRepo := repo.NewRoleRepository(db)
	ctx := context.Background()
	result, err := roleRepo.GetByName(ctx, "moderator")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "moderator", result.NamaRole)
}

// TestRoleRepository_Update_Success tests successful role update
func TestRoleRepository_Update_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{
		NamaRole: "oldname",
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	roleRepo := repo.NewRoleRepository(db)
	ctx := context.Background()
	role.NamaRole = "newname"
	err = roleRepo.Update(ctx, role)
	assert.NoError(t, err)

	// Verify update
	var updatedRole domain.Role
	err = db.First(&updatedRole, role.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "newname", updatedRole.NamaRole)
}

// TestRoleRepository_Delete_Success tests successful role deletion
func TestRoleRepository_Delete_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{
		NamaRole: "todelete",
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	roleRepo := repo.NewRoleRepository(db)
	ctx := context.Background()
	err = roleRepo.Delete(ctx, role.ID)
	assert.NoError(t, err)

	// Verify deletion
	var deletedRole domain.Role
	err = db.First(&deletedRole, role.ID).Error
	assert.Error(t, err)
}

// TestRoleRepository_GetAll_Success tests successful role list retrieval
func TestRoleRepository_GetAll_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	roles := []domain.Role{
		{NamaRole: "admin"},
		{NamaRole: "editor"},
		{NamaRole: "viewer"},
	}

	for _, role := range roles {
		err := db.Create(&role).Error
		require.NoError(t, err)
	}

	roleRepo := repo.NewRoleRepository(db)
	ctx := context.Background()

	// Test without filters
	result, total, err := roleRepo.GetAll(ctx, "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 3, total)

	// Test pagination
	result, total, err = roleRepo.GetAll(ctx, "", 1, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 3, total)
}
