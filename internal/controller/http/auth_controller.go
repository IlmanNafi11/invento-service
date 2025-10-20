package http

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	authUsecase  usecase.AuthUsecase
	cookieHelper *helper.CookieHelper
	logger       *helper.Logger
}

func NewAuthController(authUsecase usecase.AuthUsecase, cfg *config.Config) *AuthController {
	return &AuthController{
		authUsecase:  authUsecase,
		cookieHelper: helper.NewCookieHelper(cfg),
		logger:       helper.NewLogger(),
	}
}

func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	refreshToken, result, err := ctrl.authUsecase.Register(req)
	if err != nil {
		switch err.Error() {
		case "email sudah terdaftar":
			return helper.SendConflictResponse(c, err.Error())
		case "hanya email dengan domain polije.ac.id yang dapat digunakan",
			"subdomain email tidak valid, gunakan student atau teacher",
			"role tidak tersedia, silakan hubungi administrator":
			return helper.SendBadRequestResponse(c, err.Error())
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return helper.SendSuccessResponse(c, helper.StatusCreated, "Registrasi berhasil", result)
}

func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req domain.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	refreshToken, result, err := ctrl.authUsecase.Login(req)
	if err != nil {
		switch err.Error() {
		case "email atau password salah":
			return helper.SendUnauthorizedResponse(c)
		case "akun belum diaktifkan":
			return helper.SendErrorResponse(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, refreshToken)

	return helper.SendSuccessResponse(c, helper.StatusOK, "Login berhasil", result)
}

func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	oldRefreshToken := helper.GetRefreshTokenFromCookie(c)
	if oldRefreshToken == "" {
		return helper.SendBadRequestResponse(c, "Refresh token diperlukan")
	}

	newRefreshToken, result, err := ctrl.authUsecase.RefreshToken(oldRefreshToken)
	if err != nil {
		switch err.Error() {
		case "refresh token tidak valid atau sudah expired", "user tidak ditemukan":
			return helper.SendUnauthorizedResponse(c)
		default:
			return helper.SendInternalServerErrorResponse(c)
		}
	}

	ctrl.cookieHelper.SetRefreshTokenCookie(c, newRefreshToken)

	return helper.SendSuccessResponse(c, helper.StatusOK, "Token berhasil diperbarui", result)
}

func (ctrl *AuthController) ResetPassword(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	err := ctrl.authUsecase.ResetPassword(req)
	if err != nil {
		if err.Error() == "email tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Link reset password telah dikirim ke email Anda", nil)
}

func (ctrl *AuthController) ConfirmResetPassword(c *fiber.Ctx) error {
	var req domain.NewPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendBadRequestResponse(c, "Format request tidak valid")
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	err := ctrl.authUsecase.ConfirmResetPassword(req)
	if err != nil {
		if err.Error() == "token reset password tidak valid atau sudah expired" {
			return helper.SendNotFoundResponse(c, "Token reset password tidak valid")
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, helper.StatusOK, "Password berhasil direset", nil)
}

func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	token := helper.GetRefreshTokenFromCookie(c)
	if token == "" {
		return helper.SendBadRequestResponse(c, "Refresh token diperlukan")
	}

	err := ctrl.authUsecase.Logout(token)
	if err != nil {
		if err.Error() == "refresh token tidak valid" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	ctrl.cookieHelper.ClearRefreshTokenCookie(c)

	return helper.SendSuccessResponse(c, helper.StatusOK, "Logout berhasil", nil)
}
