package helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"fiber-boiler-plate/internal/usecase/repo"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestRSAKeys generates temporary RSA key pairs for testing
func generateTestRSAKeys() error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Generate rotation private key
	privateKeyRotation, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Save private key
	privateKeyFile, err := os.Create("/tmp/test_private.pem")
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	// Save rotation private key
	privateKeyRotationFile, err := os.Create("/tmp/test_private_rotation.pem")
	if err != nil {
		return err
	}
	defer privateKeyRotationFile.Close()

	privateKeyRotationPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKeyRotation),
	}
	if err := pem.Encode(privateKeyRotationFile, privateKeyRotationPEM); err != nil {
		return err
	}

	// Save public key
	publicKeyFile, err := os.Create("/tmp/test_public.pem")
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		return err
	}

	// Save rotation public key
	publicKeyRotationFile, err := os.Create("/tmp/test_public_rotation.pem")
	if err != nil {
		return err
	}
	defer publicKeyRotationFile.Close()

	publicKeyRotationBytes, err := x509.MarshalPKIXPublicKey(&privateKeyRotation.PublicKey)
	if err != nil {
		return err
	}

	publicKeyRotationPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyRotationBytes,
	}
	if err := pem.Encode(publicKeyRotationFile, publicKeyRotationPEM); err != nil {
		return err
	}

	return nil
}

// cleanupTestRSAKeys removes the temporary test keys
func cleanupTestRSAKeys() {
	os.Remove("/tmp/test_private.pem")
	os.Remove("/tmp/test_private_rotation.pem")
	os.Remove("/tmp/test_public.pem")
	os.Remove("/tmp/test_public_rotation.pem")
}

// TestNewAuthHelper_Success tests successful auth helper creation
func TestNewAuthHelper_Success(t *testing.T) {
	err := generateTestRSAKeys()
	require.NoError(t, err)
	defer cleanupTestRSAKeys()

	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	jwtManager, err := NewJWTManager(&config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/tmp/test_private.pem",
			PublicKeyPath:           "/tmp/test_public.pem",
			PrivateKeyRotationPath:  "/tmp/test_private_rotation.pem",
			PublicKeyRotationPath:   "/tmp/test_public_rotation.pem",
			ExpireHours:             24,
			RefreshTokenExpireHours: 168,
		},
	})
	require.NoError(t, err)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 168,
			ExpireHours:             24,
		},
	}

	refreshTokenRepo := repo.NewRefreshTokenRepository(db)

	authHelper := NewAuthHelper(refreshTokenRepo, jwtManager, cfg)

	assert.NotNil(t, authHelper)
	assert.NotNil(t, authHelper.jwtManager)
	assert.NotNil(t, authHelper.config)
}

// TestGenerateTokenPair_Success tests successful token pair generation
func TestGenerateTokenPair_Success(t *testing.T) {
	err := generateTestRSAKeys()
	require.NoError(t, err)
	defer cleanupTestRSAKeys()

	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	jwtManager, err := NewJWTManager(&config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/tmp/test_private.pem",
			PublicKeyPath:           "/tmp/test_public.pem",
			PrivateKeyRotationPath:  "/tmp/test_private_rotation.pem",
			PublicKeyRotationPath:   "/tmp/test_public_rotation.pem",
			ExpireHours:             24,
			RefreshTokenExpireHours: 168,
		},
	})
	require.NoError(t, err)

	refreshTokenRepo := repo.NewRefreshTokenRepository(db)

	role := &domain.Role{NamaRole: "user"}
	err = db.Create(role).Error
	require.NoError(t, err)

	user := &domain.User{
		ID:     1,
		Email:  "test@example.com",
		Name:   "Test User",
		RoleID: &role.ID,
	}

	accessToken, refreshToken, err := GenerateTokenPair(user, jwtManager, 168, refreshTokenRepo)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

// TestAuthHelper_GenerateAuthResponse_Success tests successful auth response generation
func TestAuthHelper_GenerateAuthResponse_Success(t *testing.T) {
	err := generateTestRSAKeys()
	require.NoError(t, err)
	defer cleanupTestRSAKeys()

	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	jwtManager, err := NewJWTManager(&config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/tmp/test_private.pem",
			PublicKeyPath:           "/tmp/test_public.pem",
			PrivateKeyRotationPath:  "/tmp/test_private_rotation.pem",
			PublicKeyRotationPath:   "/tmp/test_public_rotation.pem",
			ExpireHours:             24,
			RefreshTokenExpireHours: 168,
		},
	})
	require.NoError(t, err)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 168,
			ExpireHours:             24,
		},
	}

	refreshTokenRepo := repo.NewRefreshTokenRepository(db)

	authHelper := NewAuthHelper(refreshTokenRepo, jwtManager, cfg)

	role := &domain.Role{NamaRole: "user"}
	err = db.Create(role).Error
	require.NoError(t, err)

	user := &domain.User{
		ID:     1,
		Email:  "test@example.com",
		Name:   "Test User",
		RoleID: &role.ID,
	}

	refreshToken, authResponse, err := authHelper.GenerateAuthResponse(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
	assert.NotNil(t, authResponse)
	assert.NotEmpty(t, authResponse.AccessToken)
	assert.Equal(t, "Bearer", authResponse.TokenType)
}

// TestAuthHelper_RevokeAndGenerateNewTokens_Success tests successful token refresh
func TestAuthHelper_RevokeAndGenerateNewTokens_Success(t *testing.T) {
	err := generateTestRSAKeys()
	require.NoError(t, err)
	defer cleanupTestRSAKeys()

	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	jwtManager, err := NewJWTManager(&config.Config{
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/tmp/test_private.pem",
			PublicKeyPath:           "/tmp/test_public.pem",
			PrivateKeyRotationPath:  "/tmp/test_private_rotation.pem",
			PublicKeyRotationPath:   "/tmp/test_public_rotation.pem",
			ExpireHours:             24,
			RefreshTokenExpireHours: 168,
		},
	})
	require.NoError(t, err)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			RefreshTokenExpireHours: 168,
			ExpireHours:             24,
		},
	}

	refreshTokenRepo := repo.NewRefreshTokenRepository(db)

	authHelper := NewAuthHelper(refreshTokenRepo, jwtManager, cfg)

	role := &domain.Role{NamaRole: "user"}
	err = db.Create(role).Error
	require.NoError(t, err)

	user := &domain.User{
		ID:     1,
		Email:  "test@example.com",
		Name:   "Test User",
		RoleID: &role.ID,
	}

	// First generate initial tokens
	refreshToken, _, err := authHelper.GenerateAuthResponse(user)
	require.NoError(t, err)

	// Now refresh the tokens
	newRefreshToken, refreshResponse, err := authHelper.RevokeAndGenerateNewTokens(refreshToken, user)

	assert.NoError(t, err)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotNil(t, refreshResponse)
	assert.NotEmpty(t, refreshResponse.AccessToken)
	assert.Equal(t, "Bearer", refreshResponse.TokenType)
}
