package testing_test

import (
	testutil "fiber-boiler-plate/internal/testing"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ExampleSetupTestApp demonstrates how to use the testing utilities
func ExampleSetupTestApp() {
	// Setup a test app
	app := testutil.SetupTestApp(nil)

	// Don't forget to cleanup
	_ = testutil.TeardownTestApp(app)
}

// TestExample_TokenGeneration demonstrates generating test tokens
func TestExample_TokenGeneration(t *testing.T) {
	// Generate test tokens
	token := testutil.GenerateTestToken("00000000-0000-0000-0000-000000000001", "test@example.com", "user")
	assert.NotEmpty(t, token)

	// Generate token with role ID
	tokenWithRoleID := testutil.GenerateTestTokenWithRoleID("00000000-0000-0000-0000-000000000001", "test@example.com", "admin", 2)
	assert.NotEmpty(t, tokenWithRoleID)

	// Generate expired token
	expiredToken := testutil.GenerateExpiredToken()
	assert.NotEmpty(t, expiredToken)

	// Generate token with custom expiration
	customToken := testutil.GenerateTokenWithCustomExpiration("00000000-0000-0000-0000-000000000001", "test@example.com", "user", 0)
	assert.NotEmpty(t, customToken)
}

// TestExample_WithFixtures demonstrates using test fixtures
func TestExample_WithFixtures(t *testing.T) {
	// Get test fixtures
	user := testutil.GetTestUser()
	project := testutil.GetTestProject()
	modul := testutil.GetTestModul()

	// Assert fixture data - User ID is now a string (UUID)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test Project", project.NamaProject)
	assert.Equal(t, "Test Modul", modul.NamaFile)
}

// TestExample_ResponseAssertions demonstrates response assertion helpers
func TestExample_ResponseAssertions(t *testing.T) {
	t.Skip("Skipping response assertion test - requires valid config")

	// This test demonstrates how to use response assertion helpers
	// but requires a proper config.Config to be passed to SetupTestApp
	//
	// To use this test in your own code:
	// cfg := config.LoadConfig() // or setup test config
	// app := testutil.SetupTestApp(cfg)
	// resp := testutil.MakeRequest(app, "GET", "/api/v1/health", nil, "")
	// testutil.AssertStatusCode(t, resp, 200)
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

// TestExample_URLBuilder demonstrates URL building helpers
func TestExample_URLBuilder(t *testing.T) {
	// Build URL with query parameters
	url := testutil.GetRequestURL("/api/v1/users", map[string]string{
		"page":  "1",
		"limit": "10",
		"sort":  "name",
		"order": "asc",
	})

	// Check that all expected parameters are present (order doesn't matter in URLs)
	assert.Contains(t, url, "/api/v1/users?")
	assert.Contains(t, url, "page=1")
	assert.Contains(t, url, "limit=10")
	assert.Contains(t, url, "sort=name")
	assert.Contains(t, url, "order=asc")
}

// Note: ParseTestToken requires using the same key pair that was used to sign the token
// Since GenerateTestToken creates new keys each time, you'll need to generate the keys
// separately and use them for both signing and parsing in your tests.
