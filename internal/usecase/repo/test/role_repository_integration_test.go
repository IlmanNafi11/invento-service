package repo_test

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"fiber-boiler-plate/internal/usecase/repo"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// RoleRepositoryTestSuite handles all role repository tests
type RoleRepositoryTestSuite struct {
	suite.Suite
	db        *gorm.DB
	roleRepo  repo.RoleRepository
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

	err := suite.roleRepo.Create(role)
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

	result, err := suite.roleRepo.GetByID(role.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), role.ID, result.ID)
	assert.Equal(suite.T(), "admin", result.NamaRole)
}

func (suite *RoleRepositoryTestSuite) TestGetByID_NotFound() {
	result, err := suite.roleRepo.GetByID(999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

func (suite *RoleRepositoryTestSuite) TestGetByName_Success() {
	role := &domain.Role{
		NamaRole: "moderator",
	}
	err := suite.db.Create(role).Error
	require.NoError(suite.T(), err)

	result, err := suite.roleRepo.GetByName("moderator")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "moderator", result.NamaRole)
}

func (suite *RoleRepositoryTestSuite) TestGetByName_NotFound() {
	result, err := suite.roleRepo.GetByName("nonexistent")
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
	err = suite.roleRepo.Update(role)
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

	err = suite.roleRepo.Delete(role.ID)
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
	result, total, err := suite.roleRepo.GetAll("", 1, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 3)
	assert.Equal(suite.T(), 3, total)

	// Test with search
	result, total, err = suite.roleRepo.GetAll("edit", 1, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), 1, total)

	// Test pagination
	result, total, err = suite.roleRepo.GetAll("", 1, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), 3, total)
}

func TestRoleRepositorySuite(t *testing.T) {
	suite.Run(t, new(RoleRepositoryTestSuite))
}
