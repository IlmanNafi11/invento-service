package http

import (
	"errors"
	"invento-service/config"
	"invento-service/internal/controller/base"
	"invento-service/internal/dto"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"invento-service/internal/usecase"

	apperrors "invento-service/internal/errors"

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
// @Param request body dto.AuthRequest true "Credential login (email, password)"
// @Success 200 {object} dto.SuccessResponse{data=dto.AuthResponse} "Login berhasil"
// @Failure 400 {object} dto.ErrorResponse "Format request tidak valid"
// @Failure 401 {object} dto.ErrorResponse "Email atau password salah"
// @Failure 403 {object} dto.ErrorResponse "Akun belum diaktifkan"
// @Failure 500 {object} dto.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/login [post]
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req dto.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	ctx := c.UserContext()
	refreshToken, result, err := ctrl.authUsecase.Login(ctx, req)
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
// @Param request body dto.RegisterRequest true "Data registrasi (name, email, password)"
// @Success 201 {object} dto.SuccessResponse{data=dto.AuthResponse} "Registrasi berhasil"
// @Failure 400 {object} dto.ErrorResponse "Data validasi tidak valid"
// @Failure 409 {object} dto.ErrorResponse "Email sudah terdaftar"
// @Failure 500 {object} dto.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/register [post]
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	ctx := c.UserContext()
	refreshToken, result, err := ctrl.authUsecase.Register(ctx, req)
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
// @Success 200 {object} dto.SuccessResponse{data=dto.RefreshTokenResponse} "Token berhasil diperbarui"
// @Failure 400 {object} dto.ErrorResponse "Refresh token diperlukan"
// @Failure 401 {object} dto.ErrorResponse "Token tidak valid atau expired"
// @Failure 500 {object} dto.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/refresh [post]
func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	refreshToken := ctrl.cookieHelper.GetRefreshTokenFromCookie(c)
	if refreshToken == "" {
		return httputil.SendErrorResponse(c, fiber.StatusUnauthorized, "Refresh token tidak ditemukan", nil)
	}

	ctx := c.UserContext()
	newRefreshToken, result, err := ctrl.authUsecase.RefreshToken(ctx, refreshToken)
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
// @Success 200 {object} dto.SuccessResponse "Logout berhasil"
// @Failure 400 {object} dto.ErrorResponse "Refresh token diperlukan"
// @Failure 404 {object} dto.ErrorResponse "Token tidak valid"
// @Failure 500 {object} dto.ErrorResponse "Terjadi kesalahan pada server"
// @Security BearerAuth
// @Router /auth/logout [post]
func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	accessToken, _ := c.Locals(middleware.LocalsKeyAccessToken).(string)
	if accessToken != "" {
		ctx := c.UserContext()
		if err := ctrl.authUsecase.Logout(ctx, accessToken); err != nil {
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
// @Param request body dto.ResetPasswordRequest true "Email untuk reset password"
// @Success 200 {object} dto.SuccessResponse "Link reset password telah dikirim"
// @Failure 400 {object} dto.ErrorResponse "Format request tidak valid"
// @Failure 404 {object} dto.ErrorResponse "Email tidak ditemukan"
// @Failure 500 {object} dto.ErrorResponse "Terjadi kesalahan pada server"
// @Router /auth/reset-password [post]
func (ctrl *AuthController) RequestPasswordReset(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	ctx := c.UserContext()
	if err := ctrl.authUsecase.RequestPasswordReset(ctx, req); err != nil {
		ctrl.logger.Warn().Err(err).Str("email", req.Email).Msg("request password reset failed")
	}

	return ctrl.SendSuccess(c, nil, "Link reset password telah dikirim ke email Anda")
}
