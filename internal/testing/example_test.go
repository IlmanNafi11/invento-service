package testing_test

import (
	testutil "fiber-boiler-plate/internal/testing"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ExampleSetupTestApp demonstrates how to use the testing utilities
func ExampleSetupTestApp() {
	// Create a test configuration
	cfg := testutil.SetupTestConfig()

	// Setup a test app
	app := testutil.SetupTestApp(cfg)

	// Don't forget to cleanup
	defer testutil.TeardownTestApp(app)
}

// TestExample_TokenGeneration demonstrates generating test tokens
func TestExample_TokenGeneration(t *testing.T) {
	// Generate test tokens
	token := testutil.GenerateTestToken(1, "test@example.com", "user")
	assert.NotEmpty(t, token)

	// Generate token with role ID
	tokenWithRoleID := testutil.GenerateTestTokenWithRoleID(1, "test@example.com", "admin", 2)
	assert.NotEmpty(t, tokenWithRoleID)

	// Generate expired token
	expiredToken := testutil.GenerateExpiredToken()
	assert.NotEmpty(t, expiredToken)

	// Generate token with custom expiration
	customToken := testutil.GenerateTokenWithCustomExpiration(1, "test@example.com", "user", 0)
	assert.NotEmpty(t, customToken)
}

// TestExample_WithFixtures demonstrates using test fixtures
func TestExample_WithFixtures(t *testing.T) {
	// Get test fixtures
	user := testutil.GetTestUser()
	project := testutil.GetTestProject()
	modul := testutil.GetTestModul()

	// Assert fixture data
	assert.Equal(t, uint(1), user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test Project", project.NamaProject)
	assert.Equal(t, "Test Modul", modul.NamaFile)
}

// TestExample_DatabaseSeeding demonstrates database setup with test data
func TestExample_DatabaseSeeding(t *testing.T) {
	// Setup database with test data
	db, err := testutil.SetupTestDatabaseWithData()
	assert.NoError(t, err)
	defer testutil.TeardownTestDatabase(db)

	// Assert records were created
	var userCount, roleCount, projectCount, modulCount int64
	db.Table("users").Count(&userCount)
	db.Table("roles").Count(&roleCount)
	db.Table("projects").Count(&projectCount)
	db.Table("moduls").Count(&modulCount)

	assert.Equal(t, int64(2), userCount)    // 2 users
	assert.Equal(t, int64(2), roleCount)    // 2 roles
	assert.Equal(t, int64(2), projectCount) // 2 projects
	assert.Equal(t, int64(2), modulCount)   // 2 moduls
}

// TestExample_ResponseAssertions demonstrates response assertion helpers
func TestExample_ResponseAssertions(t *testing.T) {
	cfg := testutil.SetupTestConfig()
	app := testutil.SetupTestApp(cfg)

	// Make a simple request
	resp := testutil.MakeRequest(app, "GET", "/nonexistent", nil, "")
	defer resp.Body.Close()

	// Use assertion helpers
	testutil.AssertStatusCode(t, resp, 404)
}

// TestExample_FixtureRequests demonstrates request fixtures
func TestExample_FixtureRequests(t *testing.T) {
	// Get request fixtures
	registerReq := testutil.GetTestRegisterRequest()
	authReq := testutil.GetTestAuthRequest()
	createProjectReq := testutil.GetTestCreateProjectRequest()

	// Assert fixture data
	assert.Equal(t, "New User", registerReq.Name)
	assert.Equal(t, "newuser@example.com", registerReq.Email)
	assert.Equal(t, "test@example.com", authReq.Email)
	assert.Equal(t, "password123", authReq.Password)
	assert.Equal(t, "New Project", createProjectReq.NamaProject)
}

// TestExample_FixtureResponses demonstrates response fixtures
func TestExample_FixtureResponses(t *testing.T) {
	// Get response fixtures
	authResp := testutil.GetTestAuthResponse()
	projectResp := testutil.GetTestProjectResponse()
	modulResp := testutil.GetTestModulResponse()

	// Assert fixture data
	assert.Equal(t, "test_access_token", authResp.AccessToken)
	assert.Equal(t, "Bearer", authResp.TokenType)
	assert.Equal(t, "Test Project", projectResp.NamaProject)
	assert.Equal(t, "Test Modul", modulResp.NamaFile)
}

// TestExample_HealthCheckFixtures demonstrates health check fixtures
func TestExample_HealthCheckFixtures(t *testing.T) {
	// Get health check fixtures
	basicHealth := testutil.GetTestBasicHealthCheck()
	comprehensiveHealth := testutil.GetTestComprehensiveHealthCheck()
	systemMetrics := testutil.GetTestSystemMetrics()

	// Assert fixture data
	assert.Equal(t, "test-app", basicHealth.App)
	assert.Equal(t, "test-app", comprehensiveHealth.App.Name)
	assert.Equal(t, "1.0.0", comprehensiveHealth.App.Version)
	assert.NotNil(t, systemMetrics.System)
	assert.NotNil(t, systemMetrics.Database)
}

// TestExample_StatisticsFixtures demonstrates statistics fixtures
func TestExample_StatisticsFixtures(t *testing.T) {
	// Get statistics fixture
	stats := testutil.GetTestStatistics()

	// Assert fixture data
	assert.NotNil(t, stats.Data.TotalProject)
	assert.NotNil(t, stats.Data.TotalModul)
	assert.NotNil(t, stats.Data.TotalUser)
	assert.NotNil(t, stats.Data.TotalRole)
	assert.Equal(t, 45, *stats.Data.TotalProject)
	assert.Equal(t, 120, *stats.Data.TotalModul)
	assert.Equal(t, 150, *stats.Data.TotalUser)
	assert.Equal(t, 3, *stats.Data.TotalRole)
}

// TestExample_DatabaseHelpers demonstrates database helper functions
func TestExample_DatabaseHelpers(t *testing.T) {
	// Setup database
	db, err := testutil.SetupTestDatabase()
	assert.NoError(t, err)
	defer testutil.TeardownTestDatabase(db)

	// Create test user
	user, err := testutil.CreateTestUser(db, "testuser@example.com", "Test User", 1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser@example.com", user.Email)

	// Create test project
	project, err := testutil.CreateTestProject(db, "Test Project", user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "Test Project", project.NamaProject)

	// Create test modul
	modul, err := testutil.CreateTestModul(db, "Test Modul", user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, modul)
	assert.Equal(t, "Test Modul", modul.NamaFile)
}

// TestExample_URLBuilder demonstrates URL building helpers
func TestExample_URLBuilder(t *testing.T) {
	// Build URL with query parameters
	url := testutil.GetRequestURL("/api/v1/users", map[string]string{
		"page":  "1",
		"limit": "10",
		"sort":  "name",
		"order": "asc",
	})

	assert.Equal(t, "/api/v1/users?page=1&limit=10&sort=name&order=asc", url)
}
