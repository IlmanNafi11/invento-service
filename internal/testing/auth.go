package testing

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const testKeyID = "test"

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	RoleID    *int   `json:"role_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	KeyID     string `json:"kid"`
	jwt.RegisteredClaims
}

// GenerateTestRSAKeys generates a test RSA key pair for JWT signing
func GenerateTestRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return privateKey, &privateKey.PublicKey, nil
}

// GenerateTestToken generates a valid JWT test token
func GenerateTestToken(userID string, email, role string) string {
	privateKey, _, err := GenerateTestRSAKeys()
	if err != nil {
		panic(fmt.Sprintf("failed to generate test keys: %v", err))
	}

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: "access",
		KeyID:     testKeyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(fmt.Sprintf("failed to sign token: %v", err))
	}

	return tokenString
}

// GenerateTestTokenWithRoleID generates a test token with a specific role ID
func GenerateTestTokenWithRoleID(userID string, email, role string, roleID int) string {
	privateKey, _, err := GenerateTestRSAKeys()
	if err != nil {
		panic(fmt.Sprintf("failed to generate test keys: %v", err))
	}

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		RoleID:    &roleID,
		Role:      role,
		TokenType: "access",
		KeyID:     testKeyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(fmt.Sprintf("failed to sign token: %v", err))
	}

	return tokenString
}

// GenerateExpiredToken generates an expired JWT test token
func GenerateExpiredToken() string {
	privateKey, _, err := GenerateTestRSAKeys()
	if err != nil {
		panic(fmt.Sprintf("failed to generate test keys: %v", err))
	}

	claims := JWTClaims{
		UserID:    "00000000-0000-0000-0000-000000000001",
		Email:     "test@example.com",
		Role:      "user",
		TokenType: "access",
		KeyID:     testKeyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(fmt.Sprintf("failed to sign token: %v", err))
	}

	return tokenString
}

// GenerateInvalidToken generates an invalid JWT token (malformed)
func GenerateInvalidToken() string {
	return "invalid.token.here"
}

// GenerateTokenWithCustomExpiration generates a token with custom expiration
func GenerateTokenWithCustomExpiration(userID string, email, role string, expiration time.Duration) string {
	privateKey, _, err := GenerateTestRSAKeys()
	if err != nil {
		panic(fmt.Sprintf("failed to generate test keys: %v", err))
	}

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: "access",
		KeyID:     testKeyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(fmt.Sprintf("failed to sign token: %v", err))
	}

	return tokenString
}

// ParseTestToken parses a test token and returns the claims
func ParseTestToken(tokenString string) (*JWTClaims, error) {
	_, publicKey, err := GenerateTestRSAKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate test keys: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ExportPrivateKeyToPEM exports a private key to PEM format
func ExportPrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return privateKeyPEM
}

// ExportPublicKeyToPEM exports a public key to PEM format
func ExportPublicKeyToPEM(publicKey *rsa.PublicKey) []byte {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal public key: %v", err))
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return publicKeyPEM
}

// GenerateTestRefreshToken generates a test refresh token
func GenerateTestRefreshToken() string {
	privateKey, _, err := GenerateTestRSAKeys()
	if err != nil {
		panic(fmt.Sprintf("failed to generate test keys: %v", err))
	}

	claims := JWTClaims{
		UserID:    "00000000-0000-0000-0000-000000000001",
		Email:     "test@example.com",
		Role:      "user",
		TokenType: "refresh",
		KeyID:     testKeyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(fmt.Sprintf("failed to sign token: %v", err))
	}

	return tokenString
}
