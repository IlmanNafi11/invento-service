package supabase

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKeyID = "test-key-id"

type testJWKSVerifier struct {
	verifier   *JWTVerifier
	privateKey *ecdsa.PrivateKey
	keyID      string
}

func newTestJWKSVerifier(t *testing.T) *testJWKSVerifier {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	jwksPayload := map[string]interface{}{
		"keys": []map[string]string{
			{
				"kty": "EC",
				"crv": "P-256",
				"x":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.X.Bytes()),
				"y":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.Y.Bytes()),
				"kid": testKeyID,
				"use": "sig",
				"alg": "ES256",
			},
		},
	}

	jwksBytes, err := json.Marshal(jwksPayload)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksBytes)
	}))

	verifier, err := NewJWTVerifier(server.URL)
	require.NoError(t, err)

	t.Cleanup(func() {
		verifier.Shutdown()
		server.Close()
	})

	return &testJWKSVerifier{
		verifier:   verifier,
		privateKey: privateKey,
		keyID:      testKeyID,
	}
}

func (tq *testJWKSVerifier) generateES256Token(t *testing.T, claims *SupabaseClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = tq.keyID
	tokenString, err := token.SignedString(tq.privateKey)
	require.NoError(t, err)
	return tokenString
}

func generateHS256Token(t *testing.T, claims *SupabaseClaims, secret string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = testKeyID
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
	setup := newTestJWKSVerifier(t)
	token := setup.generateES256Token(t, validClaims())

	claims, err := setup.verifier.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "test-user-id", claims.GetUserID())
	assert.Equal(t, "test@student.polije.ac.id", claims.Email)
	assert.Equal(t, "Mahasiswa", claims.GetAppRole())
}

func TestVerify_ExpiredToken(t *testing.T) {
	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))

	setup := newTestJWKSVerifier(t)
	token := setup.generateES256Token(t, claims)

	_, err := setup.verifier.Verify(token)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestVerify_InvalidSignature(t *testing.T) {
	setup := newTestJWKSVerifier(t)

	wrongPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, validClaims())
	token.Header["kid"] = setup.keyID
	tokenString, err := token.SignedString(wrongPrivateKey)
	require.NoError(t, err)

	_, err = setup.verifier.Verify(tokenString)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenSignatureInvalid)
}

func TestVerify_WrongAlgorithm(t *testing.T) {
	setup := newTestJWKSVerifier(t)
	hs256Token := generateHS256Token(t, validClaims(), "hs256-secret-for-wrong-alg-test")

	_, err := setup.verifier.Verify(hs256Token)
	require.Error(t, err)
	assert.ErrorContains(t, err, "token \"alg\" parameter value \"HS256\"")
}

func TestVerify_MalformedToken(t *testing.T) {
	setup := newTestJWKSVerifier(t)
	_, err := setup.verifier.Verify("this-is-not-a-jwt")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenMalformed)
}

func TestVerify_ClockSkewTolerance(t *testing.T) {
	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-3 * time.Second))
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

	setup := newTestJWKSVerifier(t)
	token := setup.generateES256Token(t, claims)

	oldTimeFunc := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time {
		return time.Now().Add(-4 * time.Second)
	}
	t.Cleanup(func() { jwt.TimeFunc = oldTimeFunc })

	verifiedClaims, err := setup.verifier.Verify(token)
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
