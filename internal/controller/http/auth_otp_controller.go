package http

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/controller/base"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type AuthOTPController struct {
	*base.BaseController
	authOTPUsecase usecase.AuthOTPUsecase
	cookieHelper   *helper.CookieHelper
}

func NewAuthOTPController(authOTPUsecase usecase.AuthOTPUsecase, cfg *config.Config) *AuthOTPController {
	return &AuthOTPController{
		BaseController: base.NewBaseController(nil, nil),
		authOTPUsecase: authOTPUsecase,
		cookieHelper:   helper.NewCookieHelper(cfg),
	}
}

// RegisterWithOTP initiates registration by sending OTP to user's email.
//
// @Summary Registrasi dengan OTP
// @Description Memulai proses registrasi dengan mengirim kode OTP ke email pengguna.
// @Tags Auth OTP
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Data registrasi"
// @Success 201 {object} domain.SuccessResponse{data=domain.OTPResponse} "Kode OTP telah dikirim ke email"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 409 {object} domain.ErrorResponse "Email sudah terdaftar"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/otp/register [post]
func (ctrl *AuthOTPController) RegisterWithOTP(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	result, err := ctrl.authOTPUsecase.RegisterWithOTP(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendCreated(c, result, "Kode OTP telah dikirim ke email Anda")
}

// VerifyRegisterOTP verifies the OTP code sent during registration and creates the account.
//
// @Summary Verifikasi OTP registrasi
// @Description Memverifikasi kode OTP yang dikirim saat registrasi dan membuat akun pengguna.
// @Tags Auth OTP
// @Accept json
// @Produce json
// @Param request body domain.VerifyOTPRequest true "Email dan kode OTP"
// @Success 201 {object} domain.SuccessResponse{data=domain.AuthResponse} "Verifikasi OTP berhasil, akun dibuat"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 401 {object} domain.ErrorResponse "Kode OTP tidak valid atau expired"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/otp/register/verify [post]
func (ctrl *AuthOTPController) VerifyRegisterOTP(c *fiber.Ctx) error {
	var req domain.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	refreshToken, result, err := ctrl.authOTPUsecase.VerifyRegisterOTP(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return ctrl.SendCreated(c, result, "Verifikasi OTP berhasil, akun dibuat")
}

// ResendRegisterOTP resends the OTP code for registration.
//
// @Summary Kirim ulang OTP registrasi
// @Description Mengirim ulang kode OTP untuk proses registrasi.
// @Tags Auth OTP
// @Accept json
// @Produce json
// @Param request body domain.ResendOTPRequest true "Email untuk kirim ulang OTP"
// @Success 200 {object} domain.SuccessResponse{data=domain.OTPResponse} "Kode OTP baru telah dikirim"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 429 {object} domain.ErrorResponse "Terlalu banyak permintaan, coba lagi nanti"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/otp/register/resend [post]
func (ctrl *AuthOTPController) ResendRegisterOTP(c *fiber.Ctx) error {
	var req domain.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	result, err := ctrl.authOTPUsecase.ResendRegisterOTP(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Kode OTP baru telah dikirim")
}

// InitiateResetPassword initiates password reset by sending OTP to user's email.
//
// @Summary Minta reset password dengan OTP
// @Description Mengirim kode OTP ke email untuk reset password.
// @Tags Auth OTP
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "Email untuk reset password"
// @Success 200 {object} domain.SuccessResponse{data=domain.OTPResponse} "Kode OTP telah dikirim ke email"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 429 {object} domain.ErrorResponse "Terlalu banyak permintaan, coba lagi nanti"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/otp/reset-password [post]
func (ctrl *AuthOTPController) InitiateResetPassword(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	result, err := ctrl.authOTPUsecase.InitiateResetPassword(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Kode OTP reset password telah dikirim ke email Anda")
}

// VerifyResetPasswordOTP verifies the OTP code for password reset.
//
// @Summary Verifikasi OTP reset password
// @Description Memverifikasi kode OTP untuk reset password.
// @Tags Auth OTP
// @Accept json
// @Produce json
// @Param request body domain.VerifyOTPRequest true "Email dan kode OTP"
// @Success 200 {object} domain.SuccessResponse{data=domain.OTPResponse} "Verifikasi OTP berhasil"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 401 {object} domain.ErrorResponse "Kode OTP tidak valid atau expired"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/otp/reset-password/verify [post]
func (ctrl *AuthOTPController) VerifyResetPasswordOTP(c *fiber.Ctx) error {
	var req domain.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	result, err := ctrl.authOTPUsecase.VerifyResetPasswordOTP(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Verifikasi OTP reset password berhasil")
}

// ConfirmResetPasswordWithOTP completes password reset using verified OTP and new password.
//
// @Summary Konfirmasi reset password dengan OTP
// @Description Mengatur password baru menggunakan OTP yang sudah diverifikasi.
// @Tags Auth OTP
// @Accept json
// @Produce json
// @Param request body domain.ConfirmResetPasswordOTPRequest true "Email, kode OTP, dan password baru"
// @Success 200 {object} domain.SuccessResponse{data=domain.AuthResponse} "Password berhasil direset"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 401 {object} domain.ErrorResponse "Kode OTP tidak valid atau expired"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/otp/reset-password/confirm [post]
func (ctrl *AuthOTPController) ConfirmResetPasswordWithOTP(c *fiber.Ctx) error {
	var req domain.ConfirmResetPasswordOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	refreshToken, result, err := ctrl.authOTPUsecase.ConfirmResetPasswordWithOTP(req.Email, req.NewPassword)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return ctrl.SendSuccess(c, result, "Password berhasil direset")
}

// ResendResetPasswordOTP resends the OTP code for password reset.
//
// @Summary Kirim ulang OTP reset password
// @Description Mengirim ulang kode OTP untuk reset password.
// @Tags Auth OTP
// @Accept json
// @Produce json
// @Param request body domain.ResendOTPRequest true "Email untuk kirim ulang OTP"
// @Success 200 {object} domain.SuccessResponse{data=domain.OTPResponse} "Kode OTP baru telah dikirim"
// @Failure 400 {object} domain.ErrorResponse "Format request tidak valid"
// @Failure 404 {object} domain.ErrorResponse "Email tidak ditemukan"
// @Failure 429 {object} domain.ErrorResponse "Terlalu banyak permintaan, coba lagi nanti"
// @Failure 500 {object} domain.ErrorResponse "Terjadi kesalahan pada server"
// @Router /api/v1/auth/otp/reset-password/resend [post]
func (ctrl *AuthOTPController) ResendResetPasswordOTP(c *fiber.Ctx) error {
	var req domain.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil // validation error response already sent
	}

	result, err := ctrl.authOTPUsecase.ResendResetPasswordOTP(req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Kode OTP reset password baru telah dikirim")
}
