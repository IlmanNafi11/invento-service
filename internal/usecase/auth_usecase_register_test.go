package usecase

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"testing"
)

func TestLogin_UserInactive(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := dto.AuthRequest{
		Email:    "inactive@student.polije.ac.id",
		Password: "password123",
	}

	inactiveUser := &domain.User{
		ID:       "user-uuid-123",
		Email:    req.Email,
		Name:     "Inactive User",
		IsActive: false,
	}

	mockAuth.On("Login", mock.Anything, req.Email, req.Password).Return(&domain.AuthServiceResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "bearer",
		ExpiresIn:    3600,
		User:         &domain.AuthServiceUserInfo{ID: "user-uuid-123", Email: req.Email},
	}, nil)
	mockUser.On("GetByEmail", req.Email).Return(inactiveUser, nil)

	refreshToken, authResp, err := uc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, authResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "belum diaktifkan")
}

func TestRequestPasswordReset_Success(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	req := dto.ResetPasswordRequest{
		Email: "test@student.polije.ac.id",
	}

	mockAuth.On("RequestPasswordReset", mock.Anything, req.Email, mock.Anything).Return(nil)

	err := uc.RequestPasswordReset(context.Background(), req)

	assert.NoError(t, err)
	mockAuth.AssertCalled(t, "RequestPasswordReset", mock.Anything, req.Email, mock.Anything)
}

func TestRefreshToken_Success(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	mockAuth.On("RefreshToken", mock.Anything, "refresh_token_old").Return(&domain.AuthServiceResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "bearer",
		ExpiresIn:    3600,
	}, nil)

	newRefreshToken, resp, err := uc.RefreshToken(context.Background(), "refresh_token_old")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "new_refresh_token", newRefreshToken)
	assert.Equal(t, "new_access_token", resp.AccessToken)
	assert.NotZero(t, resp.ExpiresAt)
}

func TestRefreshToken_ServiceFails(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())

	mockAuth.On("RefreshToken", mock.Anything, "refresh_token_old").Return(nil, errors.New("service error"))
	newRefreshToken, resp, err := uc.RefreshToken(context.Background(), "refresh_token_old")

	assert.Error(t, err)
	assert.Empty(t, newRefreshToken)
	assert.Nil(t, resp)
}

func TestLogout_Success(t *testing.T) {
	t.Parallel()
	mockAuth := new(MockAuthService)
	mockUser := new(authTestUserRepo)
	mockRole := new(authTestRoleRepo)
	cfg := newTestConfig()

	uc := NewAuthUsecaseWithDeps(mockUser, mockRole, mockAuth, cfg, zerolog.Nop())
	mockAuth.On("Logout", mock.Anything, "access_token").Return(nil)

	err := uc.Logout(context.Background(), "access_token")
	assert.NoError(t, err)
	mockAuth.AssertCalled(t, "Logout", mock.Anything, "access_token")
}
