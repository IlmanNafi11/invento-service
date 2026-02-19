package domain

import (
	"context"
	"time"
)

type AuthService interface {
	VerifyJWT(token string) (AuthClaims, error)
	Login(ctx context.Context, email, password string) (*AuthServiceResponse, error)
	Register(ctx context.Context, req AuthServiceRegisterRequest) (*AuthServiceResponse, error)
	ResendConfirmation(ctx context.Context, email string) error
	RefreshToken(ctx context.Context, refreshToken string) (*AuthServiceResponse, error)
	Logout(ctx context.Context, accessToken string) error
	RequestPasswordReset(ctx context.Context, email, redirectTo string) error
	DeleteUser(ctx context.Context, uid string) error
}

type AuthClaims interface {
	GetUserID() string
}

type AuthServiceRegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	AutoConfirm bool   `json:"-"`
}

type AuthServiceLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthServiceResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token,omitempty"`
	TokenType    string               `json:"token_type"`
	ExpiresIn    int                  `json:"expires_in"`
	User         *AuthServiceUserInfo `json:"user"`
}

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

func (User) TableName() string {
	return "user_profiles"
}
