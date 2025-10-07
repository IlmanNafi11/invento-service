package http

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthController(authUsecase usecase.AuthUsecase) *AuthController {
	return &AuthController{
		authUsecase: authUsecase,
	}
}

func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authUsecase.Register(req)
	if err != nil {
		if err.Error() == "email sudah terdaftar" {
			return helper.SendErrorResponse(c, fiber.StatusConflict, err.Error(), nil)
		}
		if err.Error() == "hanya email dengan domain polije.ac.id yang dapat digunakan" ||
			err.Error() == "subdomain email tidak valid, gunakan student atau teacher" {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusCreated, "Registrasi berhasil", result)
}

func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req domain.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authUsecase.Login(req)
	if err != nil {
		if err.Error() == "email atau password salah" {
			return helper.SendErrorResponse(c, fiber.StatusUnauthorized, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Login berhasil", result)
}

func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	var req domain.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if validationErrors := helper.ValidateStruct(req); len(validationErrors) > 0 {
		return helper.SendValidationErrorResponse(c, validationErrors)
	}

	result, err := ctrl.authUsecase.RefreshToken(req)
	if err != nil {
		if err.Error() == "refresh token tidak valid atau sudah expired" {
			return helper.SendErrorResponse(c, fiber.StatusUnauthorized, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Token berhasil diperbarui", result)
}

func (ctrl *AuthController) ResetPassword(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
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

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Link reset password telah dikirim ke email Anda", nil)
}

func (ctrl *AuthController) ConfirmResetPassword(c *fiber.Ctx) error {
	var req domain.NewPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
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

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Password berhasil direset", nil)
}

func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	token := c.Get("X-Refresh-Token")
	if token == "" {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Refresh token diperlukan", nil)
	}

	err := ctrl.authUsecase.Logout(token)
	if err != nil {
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Logout berhasil", nil)
}
