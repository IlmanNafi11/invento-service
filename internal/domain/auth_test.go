package domain

import (
	"testing"
	"time"

	"invento-service/internal/dto"

	"github.com/stretchr/testify/assert"
)

type testAuthClaims struct {
	userID string
}

func (c testAuthClaims) GetUserID() string {
	return c.userID
}

func TestUser_Creation(t *testing.T) {
	t.Parallel()
	user := User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
	assert.True(t, user.IsActive)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestAuthRequest_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		request dto.AuthRequest
		valid   bool
	}{
		{
			name: "valid auth request",
			request: dto.AuthRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			valid: true,
		},
		{
			name: "empty email",
			request: dto.AuthRequest{
				Email:    "",
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "empty password",
			request: dto.AuthRequest{
				Email:    "test@example.com",
				Password: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.valid {
				assert.NotEmpty(t, tt.request.Email)
				assert.NotEmpty(t, tt.request.Password)
			} else {
				if tt.request.Email == "" {
					assert.Empty(t, tt.request.Email)
				}
				if tt.request.Password == "" {
					assert.Empty(t, tt.request.Password)
				}
			}
		})
	}
}

func TestRegisterRequest_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		request dto.RegisterRequest
		valid   bool
	}{
		{
			name: "valid register request",
			request: dto.RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			valid: true,
		},
		{
			name: "empty name",
			request: dto.RegisterRequest{
				Name:     "",
				Email:    "test@example.com",
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "empty email",
			request: dto.RegisterRequest{
				Name:     "Test User",
				Email:    "",
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "empty password",
			request: dto.RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.valid {
				assert.NotEmpty(t, tt.request.Name)
				assert.NotEmpty(t, tt.request.Email)
				assert.NotEmpty(t, tt.request.Password)
			} else {
				isEmpty := tt.request.Name == "" || tt.request.Email == "" || tt.request.Password == ""
				assert.True(t, isEmpty)
			}
		})
	}
}

func TestAuthResponse_Structure(t *testing.T) {
	t.Parallel()
	user := dto.AuthUserResponse{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      "mahasiswa",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	authResponse := dto.AuthResponse{
		User:        &user,
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(time.Hour).Unix(),
	}

	assert.Equal(t, &user, authResponse.User)
	assert.Equal(t, "access_token_123", authResponse.AccessToken)
	assert.Equal(t, "Bearer", authResponse.TokenType)
	assert.Equal(t, 3600, authResponse.ExpiresIn)
	assert.NotZero(t, authResponse.ExpiresAt)
	assert.Equal(t, "mahasiswa", authResponse.User.Role)
}

func TestRefreshTokenResponse_Structure(t *testing.T) {
	t.Parallel()
	response := dto.RefreshTokenResponse{
		AccessToken: "new_access_token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(time.Hour).Unix(),
	}

	assert.Equal(t, "new_access_token", response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, 3600, response.ExpiresIn)
	assert.NotZero(t, response.ExpiresAt)
}

func TestAuthClaims_Interface(t *testing.T) {
	t.Parallel()
	claims := testAuthClaims{userID: "user-claims-123"}

	var authClaims AuthClaims = claims
	assert.Equal(t, "user-claims-123", authClaims.GetUserID())
}
