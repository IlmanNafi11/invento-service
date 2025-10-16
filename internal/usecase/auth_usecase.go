package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"time"

	"gorm.io/gorm"
)

type AuthUsecase interface {
	Register(req domain.RegisterRequest) (string, *domain.AuthResponse, error)
	Login(req domain.AuthRequest) (string, *domain.AuthResponse, error)
	RefreshToken(refreshToken string) (string, *domain.RefreshTokenResponse, error)
	ResetPassword(req domain.ResetPasswordRequest) error
	ConfirmResetPassword(req domain.NewPasswordRequest) error
	Logout(token string) error
}

type authUsecase struct {
	userRepo         repo.UserRepository
	refreshTokenRepo repo.RefreshTokenRepository
	resetTokenRepo   repo.PasswordResetTokenRepository
	roleRepo         repo.RoleRepository
	authHelper       *helper.AuthHelper
	jwtManager       *helper.JWTManager
	config           *config.Config
}

func NewAuthUsecase(
	userRepo repo.UserRepository,
	refreshTokenRepo repo.RefreshTokenRepository,
	resetTokenRepo repo.PasswordResetTokenRepository,
	roleRepo repo.RoleRepository,
	config *config.Config,
) AuthUsecase {
	jwtManager, err := helper.NewJWTManager(config)
	if err != nil {
		panic("Gagal inisialisasi JWT Manager: " + err.Error())
	}

	authHelper := helper.NewAuthHelper(refreshTokenRepo, jwtManager, config)

	return &authUsecase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		resetTokenRepo:   resetTokenRepo,
		roleRepo:         roleRepo,
		authHelper:       authHelper,
		jwtManager:       jwtManager,
		config:           config,
	}
}

func (uc *authUsecase) Register(req domain.RegisterRequest) (string, *domain.AuthResponse, error) {
	emailInfo, err := helper.ValidatePolijeEmail(req.Email)
	if err != nil {
		return "", nil, err
	}

	existingUser, _ := uc.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return "", nil, errors.New("email sudah terdaftar")
	}

	role, err := uc.roleRepo.GetByName(emailInfo.RoleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("role tidak tersedia, silakan hubungi administrator")
		}
		return "", nil, errors.New("gagal mengambil data role")
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return "", nil, err
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		RoleID:   &role.ID,
		IsActive: true,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return "", nil, errors.New("gagal membuat user")
	}

	user.Role = role

	return uc.authHelper.GenerateAuthResponse(user)
}

func (uc *authUsecase) Login(req domain.AuthRequest) (string, *domain.AuthResponse, error) {
	user, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("email atau password salah")
		}
		return "", nil, errors.New("gagal mengambil data user")
	}

	if err := helper.ComparePassword(user.Password, req.Password); err != nil {
		return "", nil, err
	}

	return uc.authHelper.GenerateAuthResponse(user)
}

func (uc *authUsecase) RefreshToken(refreshToken string) (string, *domain.RefreshTokenResponse, error) {
	hashedToken := helper.HashRefreshToken(refreshToken)

	tokenRecord, err := uc.refreshTokenRepo.GetByToken(hashedToken)
	if err != nil {
		return "", nil, errors.New("refresh token tidak valid atau sudah expired")
	}

	if tokenRecord.IsRevoked {
		if err := uc.refreshTokenRepo.RevokeAllUserTokens(tokenRecord.UserID); err == nil {
		}
		return "", nil, errors.New("refresh token tidak valid atau sudah expired")
	}

	user, err := uc.userRepo.GetByID(tokenRecord.UserID)
	if err != nil {
		return "", nil, errors.New("user tidak ditemukan")
	}

	return uc.authHelper.RevokeAndGenerateNewTokens(refreshToken, user)
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

	hashedPassword, err := helper.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePassword(resetToken.Email, hashedPassword); err != nil {
		return errors.New("gagal update password")
	}

	if err := uc.resetTokenRepo.MarkAsUsed(req.Token); err != nil {
		return errors.New("gagal mark reset token sebagai used")
	}

	return nil
}

func (uc *authUsecase) Logout(token string) error {
	hashedToken := helper.HashRefreshToken(token)
	if err := uc.refreshTokenRepo.RevokeToken(hashedToken); err != nil {
		return errors.New("refresh token tidak valid")
	}

	return nil
}
