package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fiber-boiler-plate/internal/domain"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/supabase-community/supabase-go"
)

var _ domain.AuthService = (*AuthService)(nil)

type AuthService struct {
	client     *supabase.Client
	authURL    string
	ServiceKey string
	httpClient *http.Client
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

func NewAuthService(client *supabase.Client, authURL string) *AuthService {
	return &AuthService{
		client:     client,
		authURL:    authURL,
		ServiceKey: "",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
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
	httpReq.Header.Set("apikey", s.ServiceKey)

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

func (s *AuthService) Login(ctx context.Context, req domain.AuthServiceLoginRequest) (*domain.AuthServiceResponse, error) {
	body := map[string]interface{}{
		"email":    req.Email,
		"password": req.Password,
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
	httpReq.Header.Set("apikey", s.ServiceKey)

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
	httpReq.Header.Set("apikey", s.ServiceKey)

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
	httpReq.Header.Set("apikey", s.ServiceKey)

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

func (s *AuthService) GetUser(ctx context.Context, accessToken string) (*domain.AuthServiceUserInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", s.authURL+"/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("apikey", s.ServiceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user failed (status %d): %s", resp.StatusCode, string(body))
	}

	var user supabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	name := ""
	if user.UserMetadata != nil {
		if n, ok := user.UserMetadata["name"].(string); ok {
			name = n
		}
	}

	return &domain.AuthServiceUserInfo{
		ID:    user.ID,
		Email: user.Email,
		Name:  name,
	}, nil
}

// DeleteUser deletes a user from Supabase Auth by their UID using the admin API.
func (s *AuthService) DeleteUser(ctx context.Context, uid string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", s.authURL+"/admin/users/"+uid, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	httpReq.Header.Set("apikey", s.ServiceKey)

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
