package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserRepository_GetByEmail_Success tests successful user retrieval by email
func TestUserRepository_GetByEmail_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	// Create test role
	role := &domain.Role{
		NamaRole: "user",
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	// Create test user
	user := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		RoleID:   &role.ID,
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Create repository
	userRepo := NewUserRepository(db)

	// Test GetByEmail
	result, err := userRepo.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "Test User", result.Name)
	assert.True(t, result.IsActive)
}

// TestUserRepository_GetByID_Success tests successful user retrieval by ID
func TestUserRepository_GetByID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{
		NamaRole: "user",
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	user := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		RoleID:   &role.ID,
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	userRepo := NewUserRepository(db)

	result, err := userRepo.GetByID(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.True(t, result.IsActive)
}

// TestUserRepository_Create_Success tests successful user creation
func TestUserRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role := &domain.Role{
		NamaRole: "user",
	}
	err = db.Create(role).Error
	require.NoError(t, err)

	user := &domain.User{
		Name:     "New User",
		Email:    "new@example.com",
		Password: "hashedpassword",
		RoleID:   &role.ID,
		IsActive: true,
	}

	userRepo := NewUserRepository(db)

	err = userRepo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	// Verify user was created
	var count int64
	db.Model(&domain.User{}).Where("id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestUserRepository_UpdatePassword_Success tests successful password update
func TestUserRepository_UpdatePassword_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "oldpassword",
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	userRepo := NewUserRepository(db)

	err = userRepo.UpdatePassword("test@example.com", "newpassword")
	assert.NoError(t, err)

	// Verify password was updated
	var updatedUser domain.User
	err = db.First(&updatedUser, user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "newpassword", updatedUser.Password)
}

// TestUserRepository_UpdateRole_Success tests successful role update
func TestUserRepository_UpdateRole_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role1 := &domain.Role{
		NamaRole: "user",
	}
	err = db.Create(role1).Error
	require.NoError(t, err)

	role2 := &domain.Role{
		NamaRole: "admin",
	}
	err = db.Create(role2).Error
	require.NoError(t, err)

	user := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		RoleID:   &role1.ID,
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	userRepo := NewUserRepository(db)

	err = userRepo.UpdateRole(user.ID, &role2.ID)
	assert.NoError(t, err)

	// Verify role was updated
	var updatedUser domain.User
	err = db.First(&updatedUser, user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, role2.ID, *updatedUser.RoleID)
}

// TestUserRepository_UpdateProfile_Success tests successful profile update
func TestUserRepository_UpdateProfile_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	userRepo := NewUserRepository(db)

	newName := "Updated Name"
	newGender := "Laki-laki"
	newPhoto := "/uploads/new.jpg"

	err = userRepo.UpdateProfile(user.ID, newName, &newGender, &newPhoto)
	assert.NoError(t, err)

	// Verify profile was updated
	var updatedUser domain.User
	err = db.First(&updatedUser, user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, newName, updatedUser.Name)
	assert.Equal(t, &newGender, updatedUser.JenisKelamin)
	assert.Equal(t, &newPhoto, updatedUser.FotoProfil)
}

// TestUserRepository_Delete_Success tests successful user soft delete
func TestUserRepository_Delete_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		IsActive: true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	userRepo := NewUserRepository(db)

	err = userRepo.Delete(user.ID)
	assert.NoError(t, err)

	// Verify user was soft deleted
	var deletedUser domain.User
	err = db.First(&deletedUser, user.ID).Error
	assert.NoError(t, err)
	assert.False(t, deletedUser.IsActive)
}

// TestUserRepository_GetAll_Success tests successful user list retrieval
func TestUserRepository_GetAll_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	role1 := &domain.Role{
		NamaRole: "user",
	}
	err = db.Create(role1).Error
	require.NoError(t, err)

	role2 := &domain.Role{
		NamaRole: "admin",
	}
	err = db.Create(role2).Error
	require.NoError(t, err)

	// Create test users
	users := []domain.User{
		{
			Name:     "User 1",
			Email:    "user1@example.com",
			Password: "hashed",
			RoleID:   &role1.ID,
			IsActive: true,
		},
		{
			Name:     "User 2",
			Email:    "user2@example.com",
			Password: "hashed",
			RoleID:   &role2.ID,
			IsActive: true,
		},
		{
			Name:     "User 3",
			Email:    "user3@example.com",
			Password: "hashed",
			RoleID:   &role1.ID,
			IsActive: true,
		},
	}

	for _, user := range users {
		err = db.Create(&user).Error
		require.NoError(t, err)
	}

	userRepo := NewUserRepository(db)

	// Test without filters
	result, total, err := userRepo.GetAll("", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 3, total)

	// Test with search
	result, total, err = userRepo.GetAll("user1", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with role filter
	result, total, err = userRepo.GetAll("", "admin", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test pagination
	result, total, err = userRepo.GetAll("", "", 1, 2)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 3, total)
}
