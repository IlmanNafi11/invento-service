package supabase

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "invento-service/internal/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockAuthResponse(statusCode int, body string) *http.Response {
	rr := httptest.NewRecorder()
	rr.WriteHeader(statusCode)
	_, _ = rr.WriteString(body)
	return rr.Result()
}

func requireAppError(t *testing.T, err *apperrors.AppError) *apperrors.AppError {
	t.Helper()
	require.NotNil(t, err)
	return err
}

func TestParseAuthError_StatusBasedMappings(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		statusCode     int
		body           string
		expectedCode   string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "400 weak password via error_code",
			statusCode:     http.StatusBadRequest,
			body:           `{"msg":"Password should be at least 8 characters","error_code":"weak_password"}`,
			expectedCode:   apperrors.ErrValidation,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Password terlalu lemah",
		},
		{
			name:           "401 invalid credentials via error_code",
			statusCode:     http.StatusUnauthorized,
			body:           `{"msg":"Invalid login credentials","error_code":"invalid_credentials"}`,
			expectedCode:   apperrors.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Email atau password salah",
		},
		{
			name:           "403 email not confirmed via error_code",
			statusCode:     http.StatusForbidden,
			body:           `{"msg":"Email not confirmed","error_code":"email_not_confirmed"}`,
			expectedCode:   apperrors.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "Email belum dikonfirmasi",
		},
		{
			name:           "404 user not found via error_code",
			statusCode:     http.StatusNotFound,
			body:           `{"msg":"User not found","error_code":"user_not_found"}`,
			expectedCode:   apperrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "User tidak ditemukan",
		},
		{
			name:           "422 email exists via error_code",
			statusCode:     http.StatusUnprocessableEntity,
			body:           `{"msg":"User already registered","error_code":"email_exists"}`,
			expectedCode:   apperrors.ErrConflict,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "Email sudah terdaftar",
		},
		{
			name:           "429 rate limited via status code",
			statusCode:     http.StatusTooManyRequests,
			body:           `{"msg":"Too many requests"}`,
			expectedCode:   apperrors.ErrValidation,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Terlalu banyak permintaan, coba lagi nanti",
		},
		{
			name:           "500 unknown error falls back to internal",
			statusCode:     http.StatusInternalServerError,
			body:           `{"msg":"Unexpected upstream failure","error_code":"unknown_failure"}`,
			expectedCode:   apperrors.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Terjadi kesalahan pada server",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp := newMockAuthResponse(tc.statusCode, tc.body)

			appErr := requireAppError(t, ParseAuthError(resp))
			assert.Equal(t, tc.expectedCode, appErr.Code)
			assert.Equal(t, tc.expectedStatus, appErr.HTTPStatus)
			assert.Equal(t, tc.expectedMsg, appErr.Message)
		})
	}
}

func TestParseAuthError_OAuthStyleErrorField(t *testing.T) {
	t.Parallel()
	t.Run("maps invalid_grant from oauth style payload", func(t *testing.T) {
		t.Parallel()
		resp := newMockAuthResponse(http.StatusUnauthorized, `{"error":"invalid_grant","error_description":"Invalid login credentials"}`)

		appErr := requireAppError(t, ParseAuthError(resp))
		assert.Equal(t, apperrors.ErrUnauthorized, appErr.Code)
		assert.Equal(t, http.StatusUnauthorized, appErr.HTTPStatus)
		assert.Equal(t, "Email atau password salah", appErr.Message)
	})
}

func TestParseAuthError_InvalidAndEmptyBody(t *testing.T) {
	t.Parallel()
	t.Run("invalid json body returns internal app error", func(t *testing.T) {
		t.Parallel()
		resp := newMockAuthResponse(http.StatusBadRequest, `{"msg":`)

		appErr := requireAppError(t, ParseAuthError(resp))
		assert.Equal(t, apperrors.ErrInternal, appErr.Code)
		assert.Equal(t, http.StatusInternalServerError, appErr.HTTPStatus)
		assert.NotNil(t, appErr.Internal)
	})

	t.Run("empty body returns internal app error", func(t *testing.T) {
		t.Parallel()
		resp := newMockAuthResponse(http.StatusBadRequest, "")

		appErr := requireAppError(t, ParseAuthError(resp))
		assert.Equal(t, apperrors.ErrInternal, appErr.Code)
		assert.Equal(t, http.StatusInternalServerError, appErr.HTTPStatus)
		assert.NotNil(t, appErr.Internal)
	})
}

func TestIsEmailTaken(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsEmailTaken(&AuthError{ErrorCode: "email_exists"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{Message: "User already registered"}
		assert.True(t, IsEmailTaken(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsEmailTaken(&AuthError{ErrorCode: "invalid_credentials", Message: "something else"}))
	})
}

func TestIsInvalidCredentials(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsInvalidCredentials(&AuthError{ErrorCode: "invalid_credentials"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{ErrorDesc: "Invalid login credentials"}
		assert.True(t, IsInvalidCredentials(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsInvalidCredentials(&AuthError{ErrorCode: "user_not_found", Message: "user not found"}))
	})
}

func TestIsUserNotFound(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsUserNotFound(&AuthError{ErrorCode: "user_not_found"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{Message: "User not found"}
		assert.True(t, IsUserNotFound(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsUserNotFound(&AuthError{ErrorCode: "invalid_token", Message: "invalid token"}))
	})
}

func TestIsInvalidToken(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsInvalidToken(&AuthError{ErrorCode: "invalid_token"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{Message: "Invalid token claim"}
		assert.True(t, IsInvalidToken(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsInvalidToken(&AuthError{ErrorCode: "weak_password", Message: "password should be at least 8"}))
	})
}

func TestIsRefreshTokenExpired(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsRefreshTokenExpired(&AuthError{ErrorCode: "refresh_token_not_found"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{Message: "Already used refresh token"}
		assert.True(t, IsRefreshTokenExpired(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsRefreshTokenExpired(&AuthError{ErrorCode: "user_not_found", Message: "user not found"}))
	})
}

func TestIsWeakPassword(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsWeakPassword(&AuthError{ErrorCode: "weak_password"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{ErrorDesc: "Password should be at least 8 characters"}
		assert.True(t, IsWeakPassword(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsWeakPassword(&AuthError{ErrorCode: "email_not_confirmed", Message: "email not confirmed"}))
	})
}

func TestIsRateLimited(t *testing.T) {
	t.Parallel()
	t.Run("matches by status code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsRateLimited(&AuthError{StatusCode: http.StatusTooManyRequests}))
	})

	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsRateLimited(&AuthError{ErrorCode: "over_request_limit"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{Message: "Rate limit exceeded"}
		assert.True(t, IsRateLimited(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsRateLimited(&AuthError{StatusCode: http.StatusBadRequest, Message: "bad request"}))
	})
}

func TestIsUserAlreadyConfirmed(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsUserAlreadyConfirmed(&AuthError{ErrorCode: "user_already_confirmed"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{Message: "User already confirmed"}
		assert.True(t, IsUserAlreadyConfirmed(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsUserAlreadyConfirmed(&AuthError{ErrorCode: "email_not_confirmed", Message: "email not confirmed"}))
	})
}

func TestIsEmailNotConfirmed(t *testing.T) {
	t.Parallel()
	t.Run("matches by error_code", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsEmailNotConfirmed(&AuthError{ErrorCode: "email_not_confirmed"}))
	})

	t.Run("matches by message", func(t *testing.T) {
		t.Parallel()
		authErr := &AuthError{Message: "Email not confirmed"}
		assert.True(t, IsEmailNotConfirmed(authErr))
	})

	t.Run("non matching input", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsEmailNotConfirmed(&AuthError{ErrorCode: "user_not_found", Message: "user not found"}))
	})
}

func TestIsCheckers_NilInput(t *testing.T) {
	t.Parallel()
	t.Run("all checker functions return false for nil", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsEmailTaken(nil))
		assert.False(t, IsInvalidCredentials(nil))
		assert.False(t, IsUserNotFound(nil))
		assert.False(t, IsInvalidToken(nil))
		assert.False(t, IsRefreshTokenExpired(nil))
		assert.False(t, IsWeakPassword(nil))
		assert.False(t, IsRateLimited(nil))
		assert.False(t, IsUserAlreadyConfirmed(nil))
		assert.False(t, IsEmailNotConfirmed(nil))
	})
}
