package http

import (
	"errors"
	"invento-service/config"
	"invento-service/internal/controller/base"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// AuthController handles authentication-related HTTP requests.
// It embeds BaseController for common functionality like user ID extraction,
// validation helpers, and standardized response methods.
type AuthController struct {
	*base.BaseController
	authUsecase  usecase.AuthUsecase
	cookieHelper *httputil.CookieHelper
	logger       zerolog.Logger
}

// NewAuthController creates a new AuthController instance.
// Initializes base controller without JWT/Casbin since auth endpoints
// handle credentials directly (no authentication required).
func NewAuthController(authUsecase usecase.AuthUsecase, cookieHelper *httputil.CookieHelper, cfg *config.Config, logger zerolog.Logger) *AuthController {
	return &AuthController{
		BaseController: base.NewBaseController(cfg.Supabase.URL, nil),
		authUsecase:    authUsecase,
		cookieHelper:   cookieHelper,
		logger:         logger.With().Str("component", "AuthController").Logger(),
	}
}

// Login authenticates a user with email and password.
//
// @Summary Login pengguna
// @Description Autentikasi pengguna melalui Supabase Auth dengan email dan password. Data profil lokal disinkronkan jika diperlukan.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.AuthRequest true "Credential login (email, password)"
// @Success 200 {object} domain.SuccessResponse{data=domain.AuthResponse} "Login berhasil"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 401 {object} domain.ErrorResponse "Email atau password salah"
// @Failure 403 {object} domain.ErrorResponse "Akun belum diaktifkan"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req domain.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	refreshToken, result, err := ctrl.authUsecase.Login(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	ctrl.cookieHelper.SetAccessTokenCookie(c, result.AccessToken, result.ExpiresIn)
	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return ctrl.SendSuccess(c, result, "Login berhasil")
}

// Register creates a new user account.
//
// @Summary Registrasi pengguna baru
// @Description Membuat akun pengguna baru melalui Supabase Auth dan menyimpan data profil ke database lokal.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Data registrasi (name, email, password)"
// @Success 201 {object} domain.SuccessResponse{data=domain.AuthResponse} "Registrasi berhasil"
// @Failure 400 {object} domain.ErrorResponse "Data validasi tidak valid"
// @Failure 409 {object} domain.ErrorResponse "Email sudah terdaftar"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/register [post]
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	refreshToken, result, err := ctrl.authUsecase.Register(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	ctrl.cookieHelper.SetAccessTokenCookie(c, result.AccessToken, result.ExpiresIn)
	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return ctrl.SendCreated(c, result, "Registrasi berhasil")
}

// RefreshToken refreshes an access token using a valid refresh token.
//
// @Summary Refresh access token
// @Description Memperbarui access token menggunakan refresh token dari cookie.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=domain.RefreshTokenResponse} "Token berhasil diperbarui"
// @Failure 400 {object} domain.ErrorResponse "Refresh token diperlukan"
// @Failure 401 {object} domain.ErrorResponse "Token tidak valid atau expired"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/refresh [post]
func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	refreshToken := ctrl.cookieHelper.GetRefreshTokenFromCookie(c)
	if refreshToken == "" {
		return httputil.SendErrorResponse(c, fiber.StatusUnauthorized, "Refresh token tidak ditemukan", nil)
	}

	newRefreshToken, result, err := ctrl.authUsecase.RefreshToken(refreshToken)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return httputil.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	ctrl.cookieHelper.SetAccessTokenCookie(c, result.AccessToken, result.ExpiresIn)
	ctrl.cookieHelper.SetRefreshTokenCookie(c, newRefreshToken)

	return ctrl.SendSuccess(c, result, "Token berhasil diperbarui")
}

// Logout invalidates the user's refresh token.
//
// @Summary Logout pengguna
// @Description Menghapus refresh token dari cookie dan database.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse "Logout berhasil"
// @Failure 400 {object} domain.ErrorResponse "Refresh token diperlukan"
// @Failure 404 {object} domain.ErrorResponse "Token tidak valid"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Security BearerAuth
// @Router /auth/logout [post]
func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	accessToken, _ := c.Locals("access_token").(string)
	if accessToken != "" {
		if err := ctrl.authUsecase.Logout(accessToken); err != nil {
			ctrl.logger.Warn().Err(err).Msg("Supabase logout failed, clearing cookies anyway")
		}
	}

	ctrl.cookieHelper.ClearAllAuthCookies(c)
	return ctrl.SendSuccess(c, nil, "Logout berhasil")
}

// RequestPasswordReset initiates password reset by sending magic link to email.
//
// @Summary Minta reset password
// @Description Mengirim link reset password ke email yang terdaftar melalui Supabase Auth.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "Email untuk reset password"
// @Success 200 {object} domain.SuccessResponse "Link reset password telah dikirim"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/reset-password [post]
func (ctrl *AuthController) RequestPasswordReset(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	if err := ctrl.authUsecase.RequestPasswordReset(req); err != nil {
		ctrl.logger.Warn().Err(err).Str("email", req.Email).Msg("request password reset failed")
	}

	return ctrl.SendSuccess(c, nil, "Link reset password telah dikirim ke email Anda")
}
