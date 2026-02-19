package usecase

import (
	"context"
	"errors"
	"fmt"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/usecase/repo"
	"strings"
	"time"

	apperrors "invento-service/internal/errors"
	supabaseAuth "invento-service/internal/supabase"

	"github.com/rs/zerolog"
	"github.com/supabase-community/supabase-go"
	"gorm.io/gorm"
)

type AuthUsecase interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*domain.RegisterResult, error)
	Login(ctx context.Context, req dto.AuthRequest) (string, *dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, *dto.RefreshTokenResponse, error)
	RequestPasswordReset(ctx context.Context, req dto.ResetPasswordRequest) error
	Logout(ctx context.Context, token string) error
}

type authUsecase struct {
	userRepo           repo.UserRepository
	roleRepo           repo.RoleRepository
	authService        domain.AuthService
	supabaseClient     *supabase.Client
	supabaseServiceKey string
	config             *config.Config
	logger             zerolog.Logger
}

func NewAuthUsecase(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	supabaseClient *supabase.Client,
	supabaseServiceKey string,
	config *config.Config,
	logger zerolog.Logger,
) (AuthUsecase, error) {
	authURL := config.Supabase.URL + "/auth/v1"
	authService, err := supabaseAuth.NewAuthService(authURL, supabaseServiceKey)
	if err != nil {
		return nil, fmt.Errorf("auth service init: %w", err)
	}

	return &authUsecase{
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		authService:        authService,
		supabaseClient:     supabaseClient,
		supabaseServiceKey: supabaseServiceKey,
		config:             config,
		logger:             logger.With().Str("component", "AuthUsecase").Logger(),
	}, nil
}

func NewAuthUsecaseWithDeps(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	authService domain.AuthService,
	config *config.Config,
	logger zerolog.Logger,
) AuthUsecase {
	return &authUsecase{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		authService: authService,
		config:      config,
		logger:      logger.With().Str("component", "AuthUsecase").Logger(),
	}
}

func (uc *authUsecase) Register(ctx context.Context, req dto.RegisterRequest) (*domain.RegisterResult, error) {
	emailInfo, err := validatePolijeEmail(req.Email)
	if err != nil {
		return nil, apperrors.NewValidationError(err.Error(), err)
	}

	// Self-registration restricted to student email domain only (AUTH-06)
	if emailInfo.RoleName != "mahasiswa" {
		return nil, apperrors.NewValidationError(
			"Pendaftaran mandiri hanya tersedia untuk email mahasiswa (@student.polije.ac.id)",
			nil,
		)
	}

	// Check if user already exists locally
	existingUser, _ := uc.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, apperrors.NewConflictError("Email sudah terdaftar")
	}

	// Get role for the user
	role, err := uc.roleRepo.GetByName(ctx, emailInfo.RoleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role " + emailInfo.RoleName)
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("AuthUsecase.Register: get role: %w", err))
	}

	// Register with Supabase (AutoConfirm=false: requires email confirmation)
	supabaseReq := domain.AuthServiceRegisterRequest{
		Email:       req.Email,
		Password:    req.Password,
		Name:        req.Name,
		AutoConfirm: false,
	}

	authResp, err := uc.authService.Register(ctx, supabaseReq)
	if err != nil {
		// Check if user already exists in Supabase
		if strings.Contains(err.Error(), "already registered") || strings.Contains(err.Error(), "already exists") {
			return nil, apperrors.NewConflictError("Email sudah terdaftar di sistem autentikasi")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("AuthUsecase.Register: supabase register: %w", err))
	}

	// Create user in local database with Supabase user ID
	roleID := int(role.ID)
	user := &domain.User{
		ID:       authResp.User.ID, // Use Supabase user ID
		Name:     req.Name,
		Email:    req.Email,
		RoleID:   &roleID,
		IsActive: true,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		// Rollback: Delete user from Supabase if local DB creation fails
		uc.logger.Error().Err(err).Str("supabase_user_id", authResp.User.ID).Msg("failed to create local user, rolling back Supabase user")
		if deleteErr := uc.authService.DeleteUser(ctx, authResp.User.ID); deleteErr != nil {
			uc.logger.Error().Err(deleteErr).Str("supabase_user_id", authResp.User.ID).Msg("failed to rollback Supabase user deletion")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("AuthUsecase.Register: create local user: %w", err))
	}

	return &domain.RegisterResult{
		NeedsConfirmation: true,
		Message:           "Registrasi berhasil! Silakan cek email Anda untuk konfirmasi akun sebelum login.",
	}, nil
}

func (uc *authUsecase) Login(ctx context.Context, req dto.AuthRequest) (string, *dto.AuthResponse, error) {
	authResp, err := uc.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.ErrEmailNotConfirmed {
			resendErr := uc.authService.ResendConfirmation(ctx, req.Email)
			if resendErr != nil {
				uc.logger.Warn().Err(resendErr).Str("email", req.Email).Msg("Gagal mengirim ulang email konfirmasi")
			}
			return "", nil, apperrors.NewEmailNotConfirmedError(
				"Email belum dikonfirmasi. Email konfirmasi telah dikirim ulang, silakan cek inbox Anda.",
			)
		}
		return "", nil, apperrors.NewUnauthorizedError("Email atau password salah")
	}

	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			emailInfo, validateErr := validatePolijeEmail(req.Email)
			if validateErr != nil {
				return "", nil, apperrors.NewValidationError(validateErr.Error(), validateErr)
			}

			role, roleErr := uc.roleRepo.GetByName(ctx, emailInfo.RoleName)
			if roleErr != nil {
				return "", nil, apperrors.NewInternalError(fmt.Errorf("AuthUsecase.Login: get role: %w", roleErr))
			}

			roleID := int(role.ID)
			name := authResp.User.Name
			if name == "" {
				name = req.Email
			}

			user = &domain.User{
				ID:       authResp.User.ID,
				Name:     name,
				Email:    req.Email,
				RoleID:   &roleID,
				IsActive: true,
			}

			if createErr := uc.userRepo.Create(ctx, user); createErr != nil {
				uc.logger.Error().Err(createErr).Str("email", req.Email).Msg("failed to sync Supabase user to local database")
				return "", nil, apperrors.NewInternalError(fmt.Errorf("AuthUsecase.Login: sync user: %w", createErr))
			}

			user.Role = role
		} else {
			return "", nil, apperrors.NewInternalError(fmt.Errorf("AuthUsecase.Login: get user: %w", err))
		}
	}

	if !user.IsActive {
		return "", nil, apperrors.NewForbiddenError("Akun belum diaktifkan")
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	domainAuthResp := &dto.AuthResponse{
		User: &dto.AuthUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      roleName,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
		AccessToken: authResp.AccessToken,
		TokenType:   authResp.TokenType,
		ExpiresIn:   authResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second).Unix(),
	}

	return authResp.RefreshToken, domainAuthResp, nil
}

func (uc *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, *dto.RefreshTokenResponse, error) {
	authResp, err := uc.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		return "", nil, apperrors.NewUnauthorizedError("Refresh token tidak valid atau sudah expired")
	}

	domainResp := &dto.RefreshTokenResponse{
		AccessToken: authResp.AccessToken,
		TokenType:   authResp.TokenType,
		ExpiresIn:   authResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second).Unix(),
	}

	return authResp.RefreshToken, domainResp, nil
}

func (uc *authUsecase) RequestPasswordReset(ctx context.Context, req dto.ResetPasswordRequest) error {
	redirectURL := uc.config.App.CorsOriginDev + "/reset-password"
	if uc.config.App.Env == config.EnvProduction {
		redirectURL = uc.config.App.CorsOriginProd + "/reset-password"
	}

	if err := uc.authService.RequestPasswordReset(ctx, req.Email, redirectURL); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("AuthUsecase.RequestPasswordReset: %w", err))
	}

	return nil
}

func (uc *authUsecase) Logout(ctx context.Context, token string) error {
	if err := uc.authService.Logout(ctx, token); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("AuthUsecase.Logout: %w", err))
	}

	return nil
}

// emailDomainInfo holds parsed email domain information for Polije email validation.
type emailDomainInfo struct {
	IsValid   bool
	Subdomain string
	RoleName  string
}

// validatePolijeEmail validates that the email belongs to the polije.ac.id domain
// and returns the parsed domain info including the mapped role name.
func validatePolijeEmail(email string) (*emailDomainInfo, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil, errors.New("format email tidak valid")
	}

	domain := parts[1]

	if !strings.HasSuffix(domain, "polije.ac.id") {
		return nil, errors.New("hanya email dengan domain polije.ac.id yang dapat digunakan")
	}

	subdomain := ""
	if strings.Contains(domain, ".") {
		domainParts := strings.Split(domain, ".")
		if len(domainParts) >= 3 {
			subdomain = domainParts[0]
		}
	}

	info := &emailDomainInfo{
		IsValid:   true,
		Subdomain: subdomain,
	}

	switch subdomain {
	case "student":
		info.RoleName = "mahasiswa"
	case "teacher":
		info.RoleName = "dosen"
	default:
		return nil, errors.New("subdomain email tidak valid, gunakan student atau teacher")
	}

	return info, nil
}
