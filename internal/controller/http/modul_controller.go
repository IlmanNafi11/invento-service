package http

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ModulController struct {
	modulUsecase usecase.ModulUsecase
}

func NewModulController(modulUsecase usecase.ModulUsecase) *ModulController {
	return &ModulController{
		modulUsecase: modulUsecase,
	}
}

func (ctrl *ModulController) Create(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	userEmailVal := c.Locals("user_email")
	if userEmailVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userEmail, ok := userEmailVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	userRoleVal := c.Locals("user_role")
	if userRoleVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userRole, ok := userRoleVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	files := form.File["files"]
	if len(files) == 0 {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "File wajib diupload", nil)
	}

	namaFilesStr := form.Value["nama_file"]

	if len(namaFilesStr) != len(files) {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Jumlah file dan nama file harus sama", nil)
	}

	for _, nama := range namaFilesStr {
		if nama == "" {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Nama file wajib diisi", nil)
		}
	}

	result, err := ctrl.modulUsecase.Create(userID, userEmail, userRole, files, namaFilesStr)
	if err != nil {
		if err.Error() == "file harus berupa pdf, docx, xlsx, pptx, jpg, jpeg, png, atau gif" {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusCreated, "Modul berhasil dibuat", result)
}

func (ctrl *ModulController) GetList(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var params domain.ModulListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Parameter query tidak valid", nil)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	result, err := ctrl.modulUsecase.GetList(userID, params.Search, params.FilterType, params.Page, params.Limit)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar modul: "+err.Error(), nil)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar modul berhasil diambil", result)
}

func (ctrl *ModulController) Update(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	userEmailVal := c.Locals("user_email")
	if userEmailVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userEmail, ok := userEmailVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	userRoleVal := c.Locals("user_role")
	if userRoleVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userRole, ok := userRoleVal.(string)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	modulID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID modul tidak valid", nil)
	}

	namaFile := c.FormValue("nama_file")

	file, _ := c.FormFile("file")

	result, err := ctrl.modulUsecase.Update(uint(modulID), userID, userEmail, userRole, namaFile, file)
	if err != nil {
		if err.Error() == "modul tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke modul ini" {
			return helper.SendForbiddenResponse(c)
		}
		if err.Error() == "file harus berupa pdf, docx, xlsx, pptx, jpg, jpeg, png, atau gif" {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Modul berhasil diperbarui", result)
}

func (ctrl *ModulController) Delete(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	modulID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID modul tidak valid", nil)
	}

	err = ctrl.modulUsecase.Delete(uint(modulID), userID)
	if err != nil {
		if err.Error() == "modul tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke modul ini" {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Modul berhasil dihapus", nil)
}

func (ctrl *ModulController) Download(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var req domain.ModulDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if len(req.IDs) == 0 {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID modul tidak boleh kosong", nil)
	}

	filePath, err := ctrl.modulUsecase.Download(userID, req.IDs)
	if err != nil {
		if err.Error() == "modul tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return c.Download(filePath)
}
