package supabase

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "super-secret-jwt-token-with-at-least-32-characters!!"

func generateTestToken(t *testing.T, claims *SupabaseClaims, secret string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenString
}

func validClaims() *SupabaseClaims {
	return &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "test-user-id",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "https://test.supabase.co/auth/v1",
		},
		Email: "test@student.polije.ac.id",
		Role:  "authenticated",
		AAL:   "aal1",
		AppMetadata: map[string]interface{}{
			"role": "Mahasiswa",
		},
	}
}

func TestVerify_ValidToken(t *testing.T) {
	verifier := NewJWTVerifier(testSecret)
	token := generateTestToken(t, validClaims(), testSecret)

	claims, err := verifier.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "test-user-id", claims.GetUserID())
	assert.Equal(t, "test@student.polije.ac.id", claims.Email)
	assert.Equal(t, "Mahasiswa", claims.GetAppRole())
}

func TestVerify_ExpiredToken(t *testing.T) {
	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))

	verifier := NewJWTVerifier(testSecret)
	token := generateTestToken(t, claims, testSecret)

	_, err := verifier.Verify(token)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestVerify_InvalidSignature(t *testing.T) {
	verifier := NewJWTVerifier(testSecret)
	token := generateTestToken(t, validClaims(), "wrong-secret-with-at-least-32-characters!!")

	_, err := verifier.Verify(token)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenSignatureInvalid)
}

func TestVerify_WrongAlgorithm(t *testing.T) {
	claims := validClaims()
	originalToken := generateTestToken(t, claims, testSecret)

	parts := strings.Split(originalToken, ".")
	require.Len(t, parts, 3)

	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	headerBytes, err := json.Marshal(header)
	require.NoError(t, err)
	parts[0] = base64.RawURLEncoding.EncodeToString(headerBytes)
	tamperedToken := strings.Join(parts, ".")

	verifier := NewJWTVerifier(testSecret)
	_, err = verifier.Verify(tamperedToken)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenWrongAlgorithm))
}

func TestVerify_MalformedToken(t *testing.T) {
	verifier := NewJWTVerifier(testSecret)
	_, err := verifier.Verify("this-is-not-a-jwt")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenMalformed)
}

func TestVerify_ClockSkewTolerance(t *testing.T) {
	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-3 * time.Second))
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	token := generateTestToken(t, claims, testSecret)

	oldTimeFunc := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time {
		return time.Now().Add(-4 * time.Second)
	}
	t.Cleanup(func() { jwt.TimeFunc = oldTimeFunc })

	verifier := NewJWTVerifier(testSecret)
	verifiedClaims, err := verifier.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "test-user-id", verifiedClaims.GetUserID())
}

func TestGetAppRole(t *testing.T) {
	claims := validClaims()
	assert.Equal(t, "Mahasiswa", claims.GetAppRole())
}

func TestGetAppRole_NilMetadata(t *testing.T) {
	claims := validClaims()
	claims.AppMetadata = nil
	assert.Equal(t, "", claims.GetAppRole())
}

func TestGetUserID(t *testing.T) {
	claims := validClaims()
	assert.Equal(t, "test-user-id", claims.GetUserID())
}
