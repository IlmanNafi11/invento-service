package http

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// AuthController handles authentication-related HTTP requests.
// It embeds BaseController for common functionality like user ID extraction,
// validation helpers, and standardized response methods.
type AuthController struct {
	*base.BaseController
	authUsecase  usecase.AuthUsecase
	cookieHelper *helper.CookieHelper
	logger       *helper.Logger
}

// NewAuthController creates a new AuthController instance.
// Initializes base controller without JWT/Casbin since auth endpoints
// handle credentials directly (no authentication required).
func NewAuthController(authUsecase usecase.AuthUsecase, cfg *config.Config) *AuthController {
	return &AuthController{
		BaseController: base.NewBaseController(nil, nil),
		authUsecase:    authUsecase,
		cookieHelper:   helper.NewCookieHelper(cfg),
		logger:         helper.NewLogger(),
	}
}

// handleAppError standardizes error handling for AppError types.
// Returns true if error was handled (response sent), false otherwise.
func (ctrl *AuthController) handleAppError(c *fiber.Ctx, err error) bool {
	if appErr, ok := err.(*apperrors.AppError); ok {
		helper.SendAppError(c, appErr)
		return true
	}
	return false
}

// Login authenticates a user with email and password.
//
// @Summary Login pengguna
// @Description Autentikasi pengguna dengan email dan password. Refresh token dikirim via httpOnly cookie.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.AuthRequest true "Credential login"
// @Success 200 {object} domain.SuccessResponse{data=domain.AuthResponse} "Login berhasil"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 401 {object} domain.ErrorResponse "Email atau password salah"
// @Failure 403 {object} domain.ErrorResponse "Akun belum diaktifkan"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/login [post]
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
		if ctrl.handleAppError(c, err) {
			return nil
		}

		// Handle string-based errors from usecase (migration in progress)
		if err.Error() == "email atau password salah" {
			return ctrl.SendUnauthorized(c)
		}
		if err.Error() == "akun belum diaktifkan" {
			return helper.SendErrorResponse(c, fiber.StatusForbidden, "akun belum diaktifkan", nil)
		}
		return ctrl.SendInternalError(c)
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)
	return ctrl.SendSuccess(c, result, "Login berhasil")
}

// Register creates a new user account.
//
// @Summary Registrasi pengguna baru
// @Description Membuat akun pengguna baru dengan email polije.ac.id.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Data registrasi"
// @Success 201 {object} domain.SuccessResponse{data=domain.AuthResponse} "Registrasi berhasil"
// @Failure 400 {object} domain.ErrorResponse "Data validasi tidak valid"
// @Failure 409 {object} domain.ErrorResponse "Email sudah terdaftar"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/register [post]
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
		if ctrl.handleAppError(c, err) {
			return nil
		}

		// Handle string-based errors from usecase (migration in progress)
		if err.Error() == "email sudah terdaftar" {
			return ctrl.SendConflict(c, "Email sudah terdaftar")
		}
		if err.Error() == "hanya email dengan domain polije.ac.id yang dapat digunakan" ||
			err.Error() == "subdomain email tidak valid, gunakan student atau teacher" ||
			err.Error() == "role tidak tersedia, silakan hubungi administrator" {
			return ctrl.SendBadRequest(c, err.Error())
		}
		return ctrl.SendInternalError(c)
	}

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
// @Router /api/v1/auth/refresh [post]
func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	oldRefreshToken := helper.GetRefreshTokenFromCookie(c)
	if oldRefreshToken == "" {
		return ctrl.SendBadRequest(c, "Refresh token diperlukan")
	}

	newRefreshToken, result, err := ctrl.authUsecase.RefreshToken(oldRefreshToken)
	if err != nil {
		if ctrl.handleAppError(c, err) {
			return nil
		}

		// Handle string-based errors from usecase (migration in progress)
		if err.Error() == "refresh token tidak valid atau sudah expired" ||
			err.Error() == "user tidak ditemukan" {
			return ctrl.SendUnauthorized(c)
		}
		return ctrl.SendInternalError(c)
	}

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
// @Router /api/v1/auth/logout [post]
func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	token := helper.GetRefreshTokenFromCookie(c)
	if token == "" {
		return ctrl.SendBadRequest(c, "Refresh token diperlukan")
	}

	err := ctrl.authUsecase.Logout(token)
	if err != nil {
		if ctrl.handleAppError(c, err) {
			return nil
		}

		// Handle string-based errors from usecase (migration in progress)
		if err.Error() == "refresh token tidak valid" {
			return ctrl.SendNotFound(c, "Token tidak valid")
		}
		return ctrl.SendInternalError(c)
	}

	ctrl.cookieHelper.ClearRefreshTokenCookie(c)
	return ctrl.SendSuccess(c, nil, "Logout berhasil")
}

// ResetPassword initiates password reset by sending OTP to email.
//
// @Summary Minta reset password
// @Description Mengirim OTP ke email untuk reset password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "Email untuk reset password"
// @Success 200 {object} domain.SuccessResponse "Link reset password telah dikirim"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/reset-password/otp [post]
func (ctrl *AuthController) ResetPassword(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	err := ctrl.authUsecase.ResetPassword(req)
	if err != nil {
		if ctrl.handleAppError(c, err) {
			return nil
		}

		// Handle string-based errors from usecase (migration in progress)
		if err.Error() == "email tidak ditemukan" {
			return ctrl.SendNotFound(c, "Email tidak ditemukan")
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Link reset password telah dikirim ke email Anda")
}

// ConfirmResetPassword completes password reset using OTP token.
//
// @Summary Konfirmasi reset password
// @Description Mengatur password baru menggunakan token reset yang valid.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.NewPasswordRequest true "Token dan password baru"
// @Success 200 {object} domain.SuccessResponse "Password berhasil direset"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 404 {object} domain.ErrorResponse "Token tidak valid"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/reset-password/confirm-otp [post]
func (ctrl *AuthController) ConfirmResetPassword(c *fiber.Ctx) error {
	var req domain.NewPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	err := ctrl.authUsecase.ConfirmResetPassword(req)
	if err != nil {
		if ctrl.handleAppError(c, err) {
			return nil
		}

		// Handle string-based errors from usecase (migration in progress)
		if err.Error() == "token reset password tidak valid atau sudah expired" {
			return ctrl.SendNotFound(c, "Token reset password tidak valid")
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Password berhasil direset")
}
