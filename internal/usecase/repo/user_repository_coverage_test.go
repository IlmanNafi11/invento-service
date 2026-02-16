package repo_test

import (
	"context"
	"testing"

	"invento-service/internal/domain"
	testhelper "invento-service/internal/testing"
	"invento-service/internal/usecase/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_NewUserRepository(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	assert.NotNil(t, userRepo)
}

func TestUserRepository_Create_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:       "user-create-1",
		Email:    "create@example.com",
		Name:     "Create User",
		IsActive: true,
	}

	err = userRepo.Create(ctx, user)
	assert.NoError(t, err)

	var found domain.User
	err = db.First(&found, "id = ?", "user-create-1").Error
	assert.NoError(t, err)
	assert.Equal(t, "create@example.com", found.Email)
}

func TestUserRepository_Create_DuplicateID(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:       "user-dup-1",
		Email:    "dup1@example.com",
		Name:     "User One",
		IsActive: true,
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	user2 := &domain.User{
		ID:       "user-dup-1",
		Email:    "dup2@example.com",
		Name:     "User Two",
		IsActive: true,
	}
	err = userRepo.Create(ctx, user2)
	assert.Error(t, err)
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "Admin"}
	require.NoError(t, db.Create(role).Error)

	roleID := int(role.ID)
	user := &domain.User{
		ID:       "user-email-1",
		Email:    "find@example.com",
		Name:     "Find User",
		RoleID:   &roleID,
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByEmail(ctx, "find@example.com")
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "user-email-1", result.ID)
	assert.NotNil(t, result.Role)
	assert.Equal(t, "Admin", result.Role.NamaRole)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByEmail(ctx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserRepository_GetByEmail_InactiveUser(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user := &domain.User{
		ID:       "user-inactive-email",
		Email:    "inactive@example.com",
		Name:     "Inactive User",
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)
	// GORM skips zero-value bool on Create, so set inactive via raw update
	require.NoError(t, db.Model(&domain.User{}).Where("id = ?", "user-inactive-email").Update("is_active", false).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByEmail(ctx, "inactive@example.com")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "Dosen"}
	require.NoError(t, db.Create(role).Error)

	roleID := int(role.ID)
	user := &domain.User{
		ID:       "user-id-1",
		Email:    "byid@example.com",
		Name:     "ByID User",
		RoleID:   &roleID,
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByID(ctx, "user-id-1")
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "byid@example.com", result.Email)
	assert.NotNil(t, result.Role)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByID(ctx, "nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserRepository_GetByIDs_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	users := []domain.User{
		{ID: "user-ids-1", Email: "ids1@example.com", Name: "User 1", IsActive: true},
		{ID: "user-ids-2", Email: "ids2@example.com", Name: "User 2", IsActive: true},
		{ID: "user-ids-3", Email: "ids3@example.com", Name: "User 3", IsActive: true},
	}
	for i := range users {
		require.NoError(t, db.Create(&users[i]).Error)
	}
	require.NoError(t, db.Model(&domain.User{}).Where("id = ?", "user-ids-3").Update("is_active", false).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByIDs(ctx, []string{"user-ids-1", "user-ids-2", "user-ids-3"})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUserRepository_GetByIDs_EmptySlice(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByIDs(ctx, []string{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestUserRepository_UpdateRole_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "Admin"}
	require.NoError(t, db.Create(role).Error)

	user := &domain.User{
		ID:       "user-update-role",
		Email:    "updaterole@example.com",
		Name:     "Update Role User",
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	roleID := int(role.ID)
	err = userRepo.UpdateRole(ctx, "user-update-role", &roleID)
	assert.NoError(t, err)

	var updated domain.User
	require.NoError(t, db.First(&updated, "id = ?", "user-update-role").Error)
	require.NotNil(t, updated.RoleID)
	assert.Equal(t, roleID, *updated.RoleID)
}

func TestUserRepository_UpdateRole_SetToNil(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "Admin"}
	require.NoError(t, db.Create(role).Error)

	roleID := int(role.ID)
	user := &domain.User{
		ID:       "user-nil-role",
		Email:    "nilrole@example.com",
		Name:     "Nil Role User",
		RoleID:   &roleID,
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	err = userRepo.UpdateRole(ctx, "user-nil-role", nil)
	assert.NoError(t, err)
}

func TestUserRepository_UpdateProfile_AllFields(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user := &domain.User{
		ID:       "user-profile-all",
		Email:    "profile@example.com",
		Name:     "Old Name",
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	jk := "Laki-laki"
	foto := "https://example.com/photo.jpg"
	err = userRepo.UpdateProfile(ctx, "user-profile-all", "New Name", &jk, &foto)
	assert.NoError(t, err)

	var updated domain.User
	require.NoError(t, db.First(&updated, "id = ?", "user-profile-all").Error)
	assert.Equal(t, "New Name", updated.Name)
	require.NotNil(t, updated.JenisKelamin)
	assert.Equal(t, "Laki-laki", *updated.JenisKelamin)
	require.NotNil(t, updated.FotoProfil)
	assert.Equal(t, "https://example.com/photo.jpg", *updated.FotoProfil)
}

func TestUserRepository_UpdateProfile_NameOnly(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user := &domain.User{
		ID:       "user-profile-name",
		Email:    "nameonly@example.com",
		Name:     "Old Name",
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	err = userRepo.UpdateProfile(ctx, "user-profile-name", "New Name", nil, nil)
	assert.NoError(t, err)

	var updated domain.User
	require.NoError(t, db.First(&updated, "id = ?", "user-profile-name").Error)
	assert.Equal(t, "New Name", updated.Name)
}

func TestUserRepository_Delete_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user := &domain.User{
		ID:       "user-delete-1",
		Email:    "delete@example.com",
		Name:     "Delete User",
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	err = userRepo.Delete(ctx, "user-delete-1")
	assert.NoError(t, err)

	var deleted domain.User
	require.NoError(t, db.First(&deleted, "id = ?", "user-delete-1").Error)
	assert.False(t, deleted.IsActive)
}

func TestUserRepository_GetProfileWithCounts_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "Mahasiswa"}
	require.NoError(t, db.Create(role).Error)

	roleID := int(role.ID)
	user := &domain.User{
		ID:       "user-profile-counts",
		Email:    "counts@example.com",
		Name:     "Counts User",
		RoleID:   &roleID,
		IsActive: true,
	}
	require.NoError(t, db.Create(user).Error)

	for i := 0; i < 3; i++ {
		require.NoError(t, db.Create(&domain.Project{
			NamaProject: "Proj",
			UserID:      "user-profile-counts",
			Kategori:    "web",
			Semester:    1,
			Ukuran:      "s",
			PathFile:    "/p",
		}).Error)
	}
	for i := 0; i < 2; i++ {
		require.NoError(t, db.Create(&domain.Modul{
			Judul:     "Mod",
			Deskripsi: "d",
			UserID:    "user-profile-counts",
			FileName:  "f.pdf",
			FilePath:  "/m",
			FileSize:  100,
			MimeType:  "application/pdf",
			Status:    "completed",
		}).Error)
	}

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, projCount, modCount, err := userRepo.GetProfileWithCounts(ctx, "user-profile-counts")
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "counts@example.com", result.Email)
	assert.Equal(t, 3, projCount)
	assert.Equal(t, 2, modCount)
}

func TestUserRepository_GetProfileWithCounts_NotFound(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, _, _, err := userRepo.GetProfileWithCounts(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserRepository_BulkUpdateRole_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "Dosen"}
	require.NoError(t, db.Create(role).Error)

	users := []domain.User{
		{ID: "bulk-1", Email: "bulk1@example.com", Name: "Bulk 1", IsActive: true},
		{ID: "bulk-2", Email: "bulk2@example.com", Name: "Bulk 2", IsActive: true},
		{ID: "bulk-3", Email: "bulk3@example.com", Name: "Bulk 3", IsActive: true},
	}
	for i := range users {
		require.NoError(t, db.Create(&users[i]).Error)
	}
	require.NoError(t, db.Model(&domain.User{}).Where("id = ?", "bulk-3").Update("is_active", false).Error)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	err = userRepo.BulkUpdateRole(ctx, []string{"bulk-1", "bulk-2", "bulk-3"}, role.ID)
	assert.NoError(t, err)

	var u1, u3 domain.User
	require.NoError(t, db.First(&u1, "id = ?", "bulk-1").Error)
	require.NotNil(t, u1.RoleID)
	assert.Equal(t, int(role.ID), *u1.RoleID)

	require.NoError(t, db.First(&u3, "id = ?", "bulk-3").Error)
	assert.Nil(t, u3.RoleID)
}

func TestUserRepository_GetByRoleID_Success(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{NamaRole: "Admin"}
	require.NoError(t, db.Create(role).Error)

	roleID := int(role.ID)
	users := []domain.User{
		{ID: "role-1", Email: "role1@example.com", Name: "R1", RoleID: &roleID, IsActive: true},
		{ID: "role-2", Email: "role2@example.com", Name: "R2", RoleID: &roleID, IsActive: true},
		{ID: "role-3", Email: "role3@example.com", Name: "R3", IsActive: true},
	}
	for i := range users {
		require.NoError(t, db.Create(&users[i]).Error)
	}

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByRoleID(ctx, role.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUserRepository_GetByRoleID_NoResults(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userRepo := repo.NewUserRepository(db)
	ctx := context.Background()
	result, err := userRepo.GetByRoleID(ctx, 999)
	assert.NoError(t, err)
	assert.Empty(t, result)
}
