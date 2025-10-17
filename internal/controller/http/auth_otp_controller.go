package http

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type AuthOTPController struct {
	authOTPUsecase usecase.AuthOTPUsecase
	cookieHelper   *helper.CookieHelper
	logger         *helper.Logger
}

func NewAuthOTPController(authOTPUsecase usecase.AuthOTPUsecase, cfg *config.Config) *AuthOTPController {
	return &AuthOTPController{
		authOTPUsecase: authOTPUsecase,
		cookieHelper:   helper.NewCookieHelper(cfg),
		logger:         helper.NewLogger(),
	}
}

func (ctrl *AuthOTPController) RegisterWithOTP(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authOTPUsecase.RegisterWithOTP(req)
	if err != nil {
		switch err.Error() {
		case "email sudah terdaftar":
			return helper.SendConflictResponse(c, err.Error())
		case "hanya email dengan domain polije.ac.id yang dapat digunakan",
			"subdomain email tidak valid, gunakan student atau teacher",
			"role tidak tersedia, silakan hubungi administrator":
			return helper.SendBadRequestResponse(c, err.Error())
		case "gagal generate kode otp",
			"gagal menyimpan kode otp",
			"gagal mengirim kode otp ke email":
			return helper.SendInternalServerErrorResponse(c)
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	return helper.SendSuccessResponse(c, helper.StatusCreated, "Kode OTP telah dikirim ke email Anda", result)
}

func (ctrl *AuthOTPController) VerifyRegisterOTP(c *fiber.Ctx) error {
	var req domain.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	refreshToken, result, err := ctrl.authOTPUsecase.VerifyRegisterOTP(req)
	if err != nil {
		switch err.Error() {
		case "email tidak ditemukan", "kode otp tidak valid atau sudah expired":
			return helper.SendNotFoundResponse(c, err.Error())
		case "kode otp salah, coba lagi", "silakan daftar ulang, sesi OTP telah berakhir":
			return helper.SendUnauthorizedResponse(c)
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return helper.SendSuccessResponse(c, helper.StatusCreated, "Verifikasi OTP berhasil, akun dibuat", result)
}

func (ctrl *AuthOTPController) ResendRegisterOTP(c *fiber.Ctx) error {
	var req domain.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authOTPUsecase.ResendRegisterOTP(req)
	if err != nil {
		switch err.Error() {
		case "silakan daftar ulang, sesi OTP telah berakhir":
			return helper.SendNotFoundResponse(c, err.Error())
		case "silakan tunggu beberapa saat sebelum mengirim ulang kode otp":
			return helper.SendTooManyRequestsResponse(c, err.Error())
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Kode OTP baru telah dikirim", result)
}

func (ctrl *AuthOTPController) InitiateResetPassword(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authOTPUsecase.InitiateResetPassword(req)
	if err != nil {
		errMsg := err.Error()

		if errMsg == "email tidak ditemukan" {
			return helper.SendNotFoundResponse(c, errMsg)
		}

		errorDetail := map[string]interface{}{
			"error": errMsg,
			"trace": "InitiateResetPassword failed",
		}
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", errorDetail)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Kode OTP reset password telah dikirim ke email Anda", result)
}

func (ctrl *AuthOTPController) VerifyResetPasswordOTP(c *fiber.Ctx) error {
	var req domain.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authOTPUsecase.VerifyResetPasswordOTP(req)
	if err != nil {
		errMsg := err.Error()

		if errMsg == "kode otp tidak valid atau sudah expired" {
			return helper.SendNotFoundResponse(c, errMsg)
		}

		if errMsg == "kode otp salah, coba lagi" {
			return helper.SendUnauthorizedResponse(c)
		}

		errorDetail := map[string]interface{}{
			"error": errMsg,
			"trace": "VerifyResetPasswordOTP failed",
		}
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server", errorDetail)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Verifikasi OTP reset password berhasil", result)
}

func (ctrl *AuthOTPController) ConfirmResetPasswordWithOTP(c *fiber.Ctx) error {
	var req domain.ConfirmResetPasswordOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	refreshToken, result, err := ctrl.authOTPUsecase.ConfirmResetPasswordWithOTP(req.Email, req.NewPassword)
	if err != nil {
		switch err.Error() {
		case "email tidak ditemukan", "verifikasi OTP gagal":
			return helper.SendNotFoundResponse(c, err.Error())
		case "password tidak memenuhi kriteria":
			return helper.SendBadRequestResponse(c, err.Error())
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return helper.SendSuccessResponse(c, helper.StatusOK, "Password berhasil direset", result)
}

func (ctrl *AuthOTPController) ResendResetPasswordOTP(c *fiber.Ctx) error {
	var req domain.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authOTPUsecase.ResendResetPasswordOTP(req)
	if err != nil {
		switch err.Error() {
		case "silakan mulai ulang proses reset password":
			return helper.SendNotFoundResponse(c, err.Error())
		case "silakan tunggu beberapa saat sebelum mengirim ulang kode otp":
			return helper.SendTooManyRequestsResponse(c, err.Error())
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Kode OTP reset password baru telah dikirim", result)
}
