package supabase

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"invento-service/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apperrors "invento-service/internal/errors"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authTestKeyID = "auth-test-key-id"

type authTestJWKS struct {
	privateKey *ecdsa.PrivateKey
	keyID      string
	jwksURL    string
	server     *httptest.Server
}

func newAuthTestJWKS(t *testing.T) *authTestJWKS {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	jwksPayload := map[string]interface{}{
		"keys": []map[string]string{
			{
				"kty": "EC",
				"crv": "P-256",
				"x":   base64.RawURLEncoding.EncodeToString(privateKey.X.Bytes()),
				"y":   base64.RawURLEncoding.EncodeToString(privateKey.Y.Bytes()),
				"kid": authTestKeyID,
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

	t.Cleanup(server.Close)

	return &authTestJWKS{
		privateKey: privateKey,
		keyID:      authTestKeyID,
		jwksURL:    server.URL,
		server:     server,
	}
}

func newTestAuthService(t *testing.T, serverURL string) *AuthService {
	t.Helper()

	jwks := newAuthTestJWKS(t)
	verifier, err := NewJWTVerifier(jwks.jwksURL)
	require.NoError(t, err)
	t.Cleanup(verifier.Shutdown)

	return &AuthService{
		authURL:     serverURL,
		serviceKey:  "test-service-key", // test fixture - not a real credential
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		jwtVerifier: verifier,
	}
}

func newTestToken(t *testing.T, privateKey *ecdsa.PrivateKey, keyID string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, &SupabaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email: "tester@example.com",
	})
	token.Header["kid"] = keyID

	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)
	return tokenString
}

func parseAppError(t *testing.T, err error) *apperrors.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr
}

func TestNewAuthService(t *testing.T) {
	t.Parallel()
	jwks := newAuthTestJWKS(t)
	authURL := jwks.server.URL + "/auth/v1"

	svc, err := NewAuthService(authURL, "service-key")

	require.NoError(t, err)
	require.NotNil(t, svc)
	t.Cleanup(svc.jwtVerifier.Shutdown)
	assert.Equal(t, authURL, svc.authURL)
	assert.Equal(t, "service-key", svc.serviceKey)
	require.NotNil(t, svc.httpClient)
	assert.Equal(t, 30*time.Second, svc.httpClient.Timeout)
	require.NotNil(t, svc.jwtVerifier)
}

func TestAuthService_VerifyJWT(t *testing.T) {
	t.Parallel()
	jwks := newAuthTestJWKS(t)
	verifier, err := NewJWTVerifier(jwks.jwksURL)
	require.NoError(t, err)
	t.Cleanup(verifier.Shutdown)

	svc := &AuthService{jwtVerifier: verifier}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		token := newTestToken(t, jwks.privateKey, jwks.keyID)

		claims, err := svc.VerifyJWT(token)
		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "user-123", claims.GetUserID())
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		_, err := svc.VerifyJWT("invalid-jwt")
		require.Error(t, err)
	})
}

func TestAuthService_Register(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		var gotAPIKey, gotContentType string
		var gotBody map[string]interface{}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/signup", r.URL.Path)

			gotAPIKey = r.Header.Get("apikey")
			gotContentType = r.Header.Get("Content-Type")

			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"acc","refresh_token":"ref","token_type":"bearer","expires_in":3600,"user":{"id":"uid-1","email":"user@example.com","user_metadata":{"name":"Budi"}}}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		resp, err := svc.Register(context.Background(), domain.AuthServiceRegisterRequest{
			Email:    "user@example.com",
			Password: "password123",
			Name:     "Budi",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "acc", resp.AccessToken)
		assert.Equal(t, "ref", resp.RefreshToken)
		assert.Equal(t, "bearer", resp.TokenType)
		assert.Equal(t, 3600, resp.ExpiresIn)
		require.NotNil(t, resp.User)
		assert.Equal(t, "uid-1", resp.User.ID)
		assert.Equal(t, "user@example.com", resp.User.Email)
		assert.Equal(t, "Budi", resp.User.Name)
		assert.Equal(t, "test-service-key", gotAPIKey)
		assert.Equal(t, "application/json", gotContentType)

		assert.Equal(t, "user@example.com", gotBody["email"])
		assert.Equal(t, "password123", gotBody["password"])
		assert.Equal(t, true, gotBody["email_confirm"])
		dataMap, ok := gotBody["data"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Budi", dataMap["name"])
	})

	t.Run("error response returns app error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"msg":"User already registered","error_code":"user_already_exists"}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		resp, err := svc.Register(context.Background(), domain.AuthServiceRegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Existing",
		})

		require.Nil(t, resp)
		appErr := parseAppError(t, err)
		assert.Equal(t, apperrors.ErrConflict, appErr.Code)
	})

	t.Run("request creation error", func(t *testing.T) {
		t.Parallel()
		svc := newTestAuthService(t, "://bad-url")

		resp, err := svc.Register(context.Background(), domain.AuthServiceRegisterRequest{
			Email:    "user@example.com",
			Password: "password123",
			Name:     "Budi",
		})

		require.Nil(t, resp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create request")
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Parallel()
	t.Run("success with user metadata name", func(t *testing.T) {
		t.Parallel()
		var gotPath, gotQuery, gotAPIKey, gotContentType string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotQuery = r.URL.RawQuery
			gotAPIKey = r.Header.Get("apikey")
			gotContentType = r.Header.Get("Content-Type")

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"acc","refresh_token":"ref","token_type":"bearer","expires_in":3600,"user":{"id":"uid-2","email":"user@example.com","user_metadata":{"name":"Siti"}}}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		resp, err := svc.Login(context.Background(), "user@example.com", "password123")

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.User)
		assert.Equal(t, "Siti", resp.User.Name)
		assert.Equal(t, "/token", gotPath)
		assert.Equal(t, "grant_type=password", gotQuery)
		assert.Equal(t, "test-service-key", gotAPIKey)
		assert.Equal(t, "application/json", gotContentType)
	})

	t.Run("success without user metadata name", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"acc","refresh_token":"ref","token_type":"bearer","expires_in":3600,"user":{"id":"uid-3","email":"user@example.com"}}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		resp, err := svc.Login(context.Background(), "user@example.com", "password123")

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.User)
		assert.Equal(t, "", resp.User.Name)
	})

	t.Run("error response returns app error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Invalid login credentials"}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		resp, err := svc.Login(context.Background(), "wrong@example.com", "wrong")

		require.Nil(t, resp)
		appErr := parseAppError(t, err)
		assert.Equal(t, apperrors.ErrUnauthorized, appErr.Code)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		var gotAPIKey, gotContentType string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/token", r.URL.Path)
			assert.Equal(t, "grant_type=refresh_token", r.URL.RawQuery)
			gotAPIKey = r.Header.Get("apikey")
			gotContentType = r.Header.Get("Content-Type")

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"new-acc","refresh_token":"new-ref","token_type":"bearer","expires_in":1800,"user":{"id":"uid-4","email":"user@example.com"}}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		resp, err := svc.RefreshToken(context.Background(), "refresh-token")

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "new-acc", resp.AccessToken)
		assert.Equal(t, "new-ref", resp.RefreshToken)
		assert.Equal(t, "uid-4", resp.User.ID)
		assert.Equal(t, "test-service-key", gotAPIKey)
		assert.Equal(t, "application/json", gotContentType)
	})

	t.Run("error response returns app error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"msg":"refresh token expired","error_code":"refresh_token_expired"}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		resp, err := svc.RefreshToken(context.Background(), "expired-token")

		require.Nil(t, resp)
		appErr := parseAppError(t, err)
		assert.Equal(t, apperrors.ErrUnauthorized, appErr.Code)
	})
}

func TestAuthService_RequestPasswordReset(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		var gotAPIKey, gotContentType string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/recover", r.URL.Path)
			gotAPIKey = r.Header.Get("apikey")
			gotContentType = r.Header.Get("Content-Type")
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		err := svc.RequestPasswordReset(context.Background(), "user@example.com", "https://app/reset")

		require.NoError(t, err)
		assert.Equal(t, "test-service-key", gotAPIKey)
		assert.Equal(t, "application/json", gotContentType)
	})

	t.Run("error response returns formatted error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"msg":"invalid email"}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		err := svc.RequestPasswordReset(context.Background(), "bad-email", "https://app/reset")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "password reset failed (status 400)")
		assert.True(t, strings.Contains(err.Error(), "invalid email"))
		var appErr *apperrors.AppError
		assert.False(t, errors.As(err, &appErr))
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		var gotAuthorization, gotAPIKey string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/logout", r.URL.Path)
			gotAuthorization = r.Header.Get("Authorization")
			gotAPIKey = r.Header.Get("apikey")
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		err := svc.Logout(context.Background(), "access-token-123")

		require.NoError(t, err)
		assert.Equal(t, "Bearer access-token-123", gotAuthorization)
		assert.Equal(t, "test-service-key", gotAPIKey)
	})

	t.Run("error response returns app error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"msg":"invalid token","error_code":"invalid_token"}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		err := svc.Logout(context.Background(), "expired-token")

		appErr := parseAppError(t, err)
		assert.Equal(t, apperrors.ErrUnauthorized, appErr.Code)
	})
}

func TestAuthService_DeleteUser(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		var gotAuthorization, gotAPIKey, gotPath string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			gotPath = r.URL.Path
			gotAuthorization = r.Header.Get("Authorization")
			gotAPIKey = r.Header.Get("apikey")
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		err := svc.DeleteUser(context.Background(), "uid-999")

		require.NoError(t, err)
		assert.Equal(t, "/admin/users/uid-999", gotPath)
		assert.Equal(t, "Bearer test-service-key", gotAuthorization)
		assert.Equal(t, "test-service-key", gotAPIKey)
	})

	t.Run("error response returns app error", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"msg":"user not found","error_code":"user_not_found"}`))
		}))
		defer ts.Close()

		svc := newTestAuthService(t, ts.URL)
		err := svc.DeleteUser(context.Background(), "missing-user")

		appErr := parseAppError(t, err)
		assert.Equal(t, apperrors.ErrNotFound, appErr.Code)
	})
}
