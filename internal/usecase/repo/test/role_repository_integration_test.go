package repo_test

import (
	"context"
	"testing"

	"invento-service/internal/domain"
	"invento-service/internal/usecase/repo"

	testhelper "invento-service/internal/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// RoleRepositoryTestSuite handles all role repository tests
type RoleRepositoryTestSuite struct {
	suite.Suite
	db       *gorm.DB
	roleRepo repo.RoleRepository
}

func (suite *RoleRepositoryTestSuite) SetupTest() {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(suite.T(), err)
	suite.db = db
	suite.roleRepo = repo.NewRoleRepository(db)
}

func (suite *RoleRepositoryTestSuite) TearDownTest() {
	testhelper.TeardownTestDatabase(suite.db)
}

func (suite *RoleRepositoryTestSuite) TestCreate_Success() {
	role := &domain.Role{
		NamaRole: "editor",
	}

	err := suite.roleRepo.Create(context.Background(), role)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), role.ID)

	// Verify role was created
	var foundRole domain.Role
	err = suite.db.First(&foundRole, role.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "editor", foundRole.NamaRole)
}

func (suite *RoleRepositoryTestSuite) TestGetByID_Success() {
	role := &domain.Role{
		NamaRole: "admin",
	}
	err := suite.db.Create(role).Error
	require.NoError(suite.T(), err)

	result, err := suite.roleRepo.GetByID(context.Background(), role.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), role.ID, result.ID)
	assert.Equal(suite.T(), "admin", result.NamaRole)
}

func (suite *RoleRepositoryTestSuite) TestGetByID_NotFound() {
	result, err := suite.roleRepo.GetByID(context.Background(), 999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *RoleRepositoryTestSuite) TestGetByName_Success() {
	role := &domain.Role{
		NamaRole: "moderator",
	}
	err := suite.db.Create(role).Error
	require.NoError(suite.T(), err)

	result, err := suite.roleRepo.GetByName(context.Background(), "moderator")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "moderator", result.NamaRole)
}

func (suite *RoleRepositoryTestSuite) TestGetByName_NotFound() {
	result, err := suite.roleRepo.GetByName(context.Background(), "nonexistent")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *RoleRepositoryTestSuite) TestUpdate_Success() {
	role := &domain.Role{
		NamaRole: "oldname",
	}
	err := suite.db.Create(role).Error
	require.NoError(suite.T(), err)

	role.NamaRole = "newname"
	err = suite.roleRepo.Update(context.Background(), role)
	assert.NoError(suite.T(), err)

	// Verify update
	var updatedRole domain.Role
	err = suite.db.First(&updatedRole, role.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "newname", updatedRole.NamaRole)
}

func (suite *RoleRepositoryTestSuite) TestDelete_Success() {
	role := &domain.Role{
		NamaRole: "todelete",
	}
	err := suite.db.Create(role).Error
	require.NoError(suite.T(), err)

	err = suite.roleRepo.Delete(context.Background(), role.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	var deletedRole domain.Role
	err = suite.db.First(&deletedRole, role.ID).Error
	assert.Error(suite.T(), err)
}

func (suite *RoleRepositoryTestSuite) TestGetAll_Success() {
	roles := []domain.Role{
		{NamaRole: "admin"},
		{NamaRole: "editor"},
		{NamaRole: "viewer"},
	}

	for _, role := range roles {
		err := suite.db.Create(&role).Error
		require.NoError(suite.T(), err)
	}

	// Test without filters
	result, total, err := suite.roleRepo.GetAll(context.Background(), "", 1, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 3)
	assert.Equal(suite.T(), 3, total)
	// Search tidak diuji di SQLite karena query repository memakai ILIKE (khusus PostgreSQL).

	// Test pagination
	result, total, err = suite.roleRepo.GetAll(context.Background(), "", 1, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), 3, total)
}

func TestRoleRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RoleRepositoryTestSuite))
}
