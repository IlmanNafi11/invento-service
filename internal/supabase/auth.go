package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"invento-service/internal/domain"
)

var _ domain.AuthService = (*AuthService)(nil)

type AuthService struct {
	authURL     string
	serviceKey  string
	httpClient  *http.Client
	jwtVerifier *JWTVerifier
}

type supabaseUser struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
}

type supabaseAuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	User         supabaseUser `json:"user"`
}

func NewAuthService(authURL, serviceKey string) (*AuthService, error) {
	jwksURL := authURL + "/.well-known/jwks.json"
	verifier, err := NewJWTVerifier(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi JWT verifier: %w", err)
	}

	return &AuthService{
		authURL:     authURL,
		serviceKey:  serviceKey,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		jwtVerifier: verifier,
	}, nil
}

func (s *AuthService) VerifyJWT(token string) (domain.AuthClaims, error) {
	return s.jwtVerifier.Verify(token)
}

func (s *AuthService) Register(ctx context.Context, req domain.AuthServiceRegisterRequest) (*domain.AuthServiceResponse, error) {
	body := map[string]interface{}{
		"email":         req.Email,
		"password":      req.Password,
		"email_confirm": true,
		"data": map[string]interface{}{
			"name": req.Name,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.authURL+"/signup", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("supabase signup failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ParseAuthError(resp)
	}

	var authResp supabaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &domain.AuthServiceResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenType:    authResp.TokenType,
		ExpiresIn:    authResp.ExpiresIn,
		User: &domain.AuthServiceUserInfo{
			ID:    authResp.User.ID,
			Email: authResp.User.Email,
			Name:  req.Name,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.AuthServiceResponse, error) {
	body := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.authURL+"/token?grant_type=password", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("supabase login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ParseAuthError(resp)
	}

	var authResp supabaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	name := ""
	if authResp.User.UserMetadata != nil {
		if n, ok := authResp.User.UserMetadata["name"].(string); ok {
			name = n
		}
	}

	return &domain.AuthServiceResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenType:    authResp.TokenType,
		ExpiresIn:    authResp.ExpiresIn,
		User: &domain.AuthServiceUserInfo{
			ID:    authResp.User.ID,
			Email: authResp.User.Email,
			Name:  name,
		},
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthServiceResponse, error) {
	body := map[string]interface{}{
		"refresh_token": refreshToken,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.authURL+"/token?grant_type=refresh_token", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ParseAuthError(resp)
	}

	var authResp supabaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &domain.AuthServiceResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenType:    authResp.TokenType,
		ExpiresIn:    authResp.ExpiresIn,
		User:         &domain.AuthServiceUserInfo{ID: authResp.User.ID},
	}, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string, redirectTo string) error {
	body := map[string]interface{}{
		"email":       email,
		"redirect_to": redirectTo,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.authURL+"/recover", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("password reset failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("password reset failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.authURL+"/logout", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ParseAuthError(resp)
	}

	return nil
}

func (s *AuthService) DeleteUser(ctx context.Context, uid string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", s.authURL+"/admin/users/"+uid, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.serviceKey)
	httpReq.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ParseAuthError(resp)
	}

	return nil
}
