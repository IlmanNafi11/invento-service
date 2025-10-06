package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthUsecase interface {
	Register(req domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(req domain.AuthRequest) (*domain.AuthResponse, error)
	RefreshToken(req domain.RefreshTokenRequest) (*domain.RefreshTokenResponse, error)
	ResetPassword(req domain.ResetPasswordRequest) error
	ConfirmResetPassword(req domain.NewPasswordRequest) error
	Logout(token string) error
}

type authUsecase struct {
	userRepo         repo.UserRepository
	refreshTokenRepo repo.RefreshTokenRepository
	resetTokenRepo   repo.PasswordResetTokenRepository
	config           *config.Config
}

func NewAuthUsecase(
	userRepo repo.UserRepository,
	refreshTokenRepo repo.RefreshTokenRepository,
	resetTokenRepo repo.PasswordResetTokenRepository,
	config *config.Config,
) AuthUsecase {
	return &authUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		resetTokenRepo:   resetTokenRepo,
		config:           config,
	}
}

func (uc *authUsecase) Register(req domain.RegisterRequest) (*domain.AuthResponse, error) {
	existingUser, _ := uc.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsActive: true,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, errors.New("gagal membuat user")
	}

	return uc.generateAuthResponse(user)
}

func (uc *authUsecase) Login(req domain.AuthRequest) (*domain.AuthResponse, error) {
	user, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("email atau password salah")
		}
		return nil, errors.New("gagal mengambil data user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	return uc.generateAuthResponse(user)
}

func (uc *authUsecase) RefreshToken(req domain.RefreshTokenRequest) (*domain.RefreshTokenResponse, error) {
	refreshToken, err := uc.refreshTokenRepo.GetByToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("refresh token tidak valid atau sudah expired")
	}

	user, err := uc.userRepo.GetByID(refreshToken.UserID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	if err := uc.refreshTokenRepo.RevokeToken(req.RefreshToken); err != nil {
		return nil, errors.New("gagal revoke refresh token lama")
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	accessToken, err := helper.GenerateAccessToken(user.ID, user.Email, user.RoleID, roleName, uc.config.JWT.Secret, uc.config.JWT.ExpireHours)
	if err != nil {
		return nil, errors.New("gagal generate access token")
	}

	newRefreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("gagal generate refresh token")
	}

	refreshTokenExpiry := time.Now().Add(time.Hour * time.Duration(uc.config.JWT.RefreshTokenExpireHours))
	if _, err := uc.refreshTokenRepo.Create(user.ID, newRefreshToken, refreshTokenExpiry); err != nil {
		return nil, errors.New("gagal simpan refresh token")
	}

	return &domain.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    uc.config.JWT.ExpireHours * 3600,
	}, nil
}

func (uc *authUsecase) ResetPassword(req domain.ResetPasswordRequest) error {
	user, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("email tidak ditemukan")
		}
		return errors.New("gagal mengambil data user")
	}

	resetToken, err := helper.GenerateResetToken()
	if err != nil {
		return errors.New("gagal generate reset token")
	}

	expiresAt := time.Now().Add(time.Hour * 1)
	if _, err := uc.resetTokenRepo.Create(user.Email, resetToken, expiresAt); err != nil {
		return errors.New("gagal simpan reset token")
	}

	return nil
}

func (uc *authUsecase) ConfirmResetPassword(req domain.NewPasswordRequest) error {
	resetToken, err := uc.resetTokenRepo.GetByToken(req.Token)
	if err != nil {
		return errors.New("token reset password tidak valid atau sudah expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal mengenkripsi password")
	}

	if err := uc.userRepo.UpdatePassword(resetToken.Email, string(hashedPassword)); err != nil {
		return errors.New("gagal update password")
	}

	if err := uc.resetTokenRepo.MarkAsUsed(req.Token); err != nil {
		return errors.New("gagal mark reset token sebagai used")
	}

	return nil
}

func (uc *authUsecase) Logout(token string) error {
	return uc.refreshTokenRepo.RevokeToken(token)
}

func (uc *authUsecase) generateAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	accessToken, err := helper.GenerateAccessToken(user.ID, user.Email, user.RoleID, roleName, uc.config.JWT.Secret, uc.config.JWT.ExpireHours)
	if err != nil {
		return nil, errors.New("gagal generate access token")
	}

	refreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		return nil, errors.New("gagal generate refresh token")
	}

	refreshTokenExpiry := time.Now().Add(time.Hour * time.Duration(uc.config.JWT.RefreshTokenExpireHours))
	if _, err := uc.refreshTokenRepo.Create(user.ID, refreshToken, refreshTokenExpiry); err != nil {
		return nil, errors.New("gagal simpan refresh token")
	}

	return &domain.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    uc.config.JWT.ExpireHours * 3600,
	}, nil
}
