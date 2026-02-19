package supabase

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	apperrors "invento-service/internal/errors"
)

// AuthError represents the error response from Supabase Auth API.
type AuthError struct {
	Code       string `json:"code"`
	Message    string `json:"msg"`
	ErrorCode  string `json:"error_code"`
	Error      string `json:"error"`
	ErrorDesc  string `json:"error_description"`
	StatusCode int    `json:"-"`
}

// ParseAuthError parses a Supabase error response and converts it to an AppError.
func ParseAuthError(resp *http.Response) *apperrors.AppError {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return apperrors.NewInternalError(err)
	}

	var authErr AuthError
	if err := json.Unmarshal(body, &authErr); err != nil {
		return apperrors.NewInternalError(err)
	}
	authErr.StatusCode = resp.StatusCode

	return mapSupabaseError(&authErr)
}

func mapSupabaseError(authErr *AuthError) *apperrors.AppError {
	if IsEmailTaken(authErr) {
		return apperrors.NewConflictError("Email sudah terdaftar")
	}

	if IsInvalidCredentials(authErr) {
		return apperrors.NewUnauthorizedError("Email atau password salah")
	}

	if IsUserNotFound(authErr) {
		return apperrors.NewNotFoundError("User")
	}

	if IsInvalidToken(authErr) {
		return apperrors.NewUnauthorizedError("Token tidak valid atau kadaluarsa")
	}

	if IsRefreshTokenExpired(authErr) {
		return apperrors.NewUnauthorizedError("Refresh token sudah kadaluarsa")
	}

	if IsWeakPassword(authErr) {
		return apperrors.NewValidationError("Password terlalu lemah", nil)
	}

	if IsRateLimited(authErr) {
		return apperrors.NewValidationError("Terlalu banyak permintaan, coba lagi nanti", nil)
	}

	if IsUserAlreadyConfirmed(authErr) {
		return apperrors.NewConflictError("User sudah dikonfirmasi")
	}

	if IsEmailNotConfirmed(authErr) {
		return apperrors.NewEmailNotConfirmedError("Email belum dikonfirmasi. Silakan cek email Anda untuk konfirmasi akun.")
	}

	return apperrors.NewInternalError(nil)
}

// IsEmailTaken checks if the error indicates the email is already registered.
func IsEmailTaken(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "user_already_exists" || authErr.ErrorCode == "email_exists" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "user already registered") ||
		strings.Contains(lowerMsg, "email already exists") ||
		strings.Contains(lowerDesc, "user already registered") ||
		strings.Contains(lowerDesc, "email already exists")
}

// IsInvalidCredentials checks if the error indicates invalid login credentials.
func IsInvalidCredentials(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "invalid_credentials" || authErr.ErrorCode == "invalid_grant" {
		return true
	}

	if authErr.Error == "invalid_grant" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "invalid login credentials") ||
		strings.Contains(lowerMsg, "invalid password") ||
		strings.Contains(lowerDesc, "invalid login credentials") ||
		strings.Contains(lowerDesc, "invalid password")
}

// IsUserNotFound checks if the error indicates the user was not found.
func IsUserNotFound(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "user_not_found" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "user not found") ||
		strings.Contains(lowerDesc, "user not found")
}

// IsInvalidToken checks if the error indicates an invalid or malformed token.
func IsInvalidToken(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "invalid_token" ||
		authErr.ErrorCode == "bad_jwt" ||
		authErr.ErrorCode == "token_expired" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "invalid token") ||
		strings.Contains(lowerMsg, "jwt expired") ||
		strings.Contains(lowerMsg, "token has expired") ||
		strings.Contains(lowerDesc, "invalid token") ||
		strings.Contains(lowerDesc, "jwt expired") ||
		strings.Contains(lowerDesc, "token has expired")
}

// IsRefreshTokenExpired checks if the error indicates an expired refresh token.
func IsRefreshTokenExpired(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "refresh_token_expired" || authErr.ErrorCode == "refresh_token_not_found" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "refresh token") ||
		strings.Contains(lowerDesc, "refresh token")
}

// IsWeakPassword checks if the error indicates the password is too weak.
func IsWeakPassword(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "weak_password" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "weak password") ||
		strings.Contains(lowerMsg, "password should be") ||
		strings.Contains(lowerDesc, "weak password") ||
		strings.Contains(lowerDesc, "password should be")
}

// IsRateLimited checks if the error indicates rate limiting.
func IsRateLimited(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.StatusCode == http.StatusTooManyRequests {
		return true
	}

	if authErr.ErrorCode == "over_request_limit" || authErr.ErrorCode == "over_email_send_rate_limit" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "rate limit") ||
		strings.Contains(lowerMsg, "too many requests") ||
		strings.Contains(lowerDesc, "rate limit") ||
		strings.Contains(lowerDesc, "too many requests")
}

// IsUserAlreadyConfirmed checks if the error indicates the user is already confirmed.
func IsUserAlreadyConfirmed(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "user_already_confirmed" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "already confirmed") ||
		strings.Contains(lowerDesc, "already confirmed")
}

// IsEmailNotConfirmed checks if the error indicates the email is not confirmed.
func IsEmailNotConfirmed(authErr *AuthError) bool {
	if authErr == nil {
		return false
	}

	if authErr.ErrorCode == "email_not_confirmed" {
		return true
	}

	lowerMsg := strings.ToLower(authErr.Message)
	lowerDesc := strings.ToLower(authErr.ErrorDesc)

	return strings.Contains(lowerMsg, "email not confirmed") ||
		strings.Contains(lowerMsg, "confirm your email") ||
		strings.Contains(lowerDesc, "email not confirmed") ||
		strings.Contains(lowerDesc, "confirm your email")
}
