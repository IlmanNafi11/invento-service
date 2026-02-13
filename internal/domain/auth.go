package domain

import (
	"context"
	"time"
)

// AuthService defines the authentication operations interface.
// This interface allows for dependency injection and mocking in tests.
type AuthService interface {
	VerifyJWT(token string) (AuthClaims, error)
	Login(ctx context.Context, email, password string) (*AuthServiceResponse, error)
	Register(ctx context.Context, req AuthServiceRegisterRequest) (*AuthServiceResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthServiceResponse, error)
	Logout(ctx context.Context, accessToken string) error
	RequestPasswordReset(ctx context.Context, email string, redirectTo string) error
	DeleteUser(ctx context.Context, uid string) error
}

type AuthClaims interface {
	GetUserID() string
}

// AuthServiceRegisterRequest represents the registration request for auth service.
type AuthServiceRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

// AuthServiceLoginRequest represents the login request for auth service.
type AuthServiceLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthServiceResponse represents the response from auth service operations.
type AuthServiceResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token,omitempty"`
	TokenType    string               `json:"token_type"`
	ExpiresIn    int                  `json:"expires_in"`
	User         *AuthServiceUserInfo `json:"user"`
}

// AuthServiceUserInfo represents user information from auth service.
type AuthServiceUserInfo struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	Name         string                 `json:"name"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
}

type User struct {
	ID           string    `json:"id" gorm:"column:id;type:uuid;primary_key"`
	Email        string    `json:"email" gorm:"column:email;type:text;not null"`
	Name         string    `json:"name" gorm:"column:name;type:text;not null"`
	JenisKelamin *string   `json:"jenis_kelamin,omitempty" gorm:"column:jenis_kelamin;type:text"`
	FotoProfil   *string   `json:"foto_profil,omitempty" gorm:"column:foto_profil;type:varchar(500)"`
	RoleID       *int      `json:"role_id,omitempty" gorm:"column:role_id"`
	IsActive     bool      `json:"is_active" gorm:"column:is_active;type:boolean;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
	Role         *Role     `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

// TableName specifies the table name for User model to match Supabase schema
func (User) TableName() string {
	return "user_profiles"
}

type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// AuthUserResponse represents safe user data returned in auth responses.
// Excludes sensitive fields like RoleID, IsActive from client exposure.
type AuthUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type AuthResponse struct {
	User        *AuthUserResponse `json:"user"`
	AccessToken string            `json:"access_token"`
	TokenType   string            `json:"token_type"`
	ExpiresIn   int               `json:"expires_in"`
	ExpiresAt   int64             `json:"expires_at"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   int64  `json:"expires_at"`
}
