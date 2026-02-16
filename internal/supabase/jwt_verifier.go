package supabase

import (
	"errors"
	"fmt"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

type SupabaseClaims struct {
	jwt.RegisteredClaims
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone,omitempty"`
	Role         string                 `json:"role"`
	AAL          string                 `json:"aal"`
	SessionID    string                 `json:"session_id"`
	IsAnonymous  bool                   `json:"is_anonymous"`
	AppMetadata  map[string]interface{} `json:"app_metadata,omitempty"`
	UserMetadata map[string]interface{} `json:"user_metadata,omitempty"`
}

func (c *SupabaseClaims) GetAppRole() string {
	if c.AppMetadata == nil {
		return ""
	}

	if role, ok := c.AppMetadata["role"].(string); ok {
		return role
	}

	return ""
}

func (c *SupabaseClaims) GetUserID() string {
	return c.Subject
}

var (
	ErrTokenMalformed        = errors.New("token is malformed")
	ErrTokenExpired          = errors.New("token is expired")
	ErrTokenSignatureInvalid = errors.New("invalid token signature")
	ErrTokenInvalidClaims    = errors.New("invalid token claims")
	ErrTokenWrongAlgorithm   = errors.New("unexpected signing algorithm")
)

// JWTVerifier memverifikasi JWT Supabase menggunakan JWKS (ES256).
type JWTVerifier struct {
	jwks *keyfunc.JWKS
}

// NewJWTVerifier membuat verifier yang mengambil public key dari JWKS endpoint Supabase.
// jwksURL: {SUPABASE_URL}/auth/v1/.well-known/jwks.json
func NewJWTVerifier(jwksURL string) (*JWTVerifier, error) {
	options := keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  5 * time.Minute,
		RefreshTimeout:    30 * time.Second,
		RefreshUnknownKID: true,
	}

	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil JWKS dari %s: %w", jwksURL, err)
	}

	return &JWTVerifier{jwks: jwks}, nil
}

func (v *JWTVerifier) Verify(tokenString string) (*SupabaseClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&SupabaseClaims{},
		v.jwks.Keyfunc,
	)
	if err != nil {
		return nil, v.categorizeError(err)
	}

	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalidClaims
	}

	return claims, nil
}

// Shutdown menghentikan goroutine background refresh JWKS.
func (v *JWTVerifier) Shutdown() {
	if v.jwks != nil {
		v.jwks.EndBackground()
	}
}

func (v *JWTVerifier) categorizeError(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenMalformed):
		return ErrTokenMalformed
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return ErrTokenSignatureInvalid
	case errors.Is(err, jwt.ErrTokenExpired), errors.Is(err, jwt.ErrTokenNotValidYet):
		return ErrTokenExpired
	default:
		return fmt.Errorf("token validation error: %w", err)
	}
}
