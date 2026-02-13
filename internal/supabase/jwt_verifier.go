package supabase

import (
	"errors"
	"fmt"

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

type JWTVerifier struct {
	secret []byte
}

func NewJWTVerifier(secret string) *JWTVerifier {
	return &JWTVerifier{secret: []byte(secret)}
}

func (v *JWTVerifier) Verify(tokenString string) (*SupabaseClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&SupabaseClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrTokenWrongAlgorithm, token.Header["alg"])
			}
			return v.secret, nil
		},
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
