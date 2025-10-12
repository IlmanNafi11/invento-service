package http

import (
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ProjectController struct {
	projectUsecase usecase.ProjectUsecase
}

func NewProjectController(projectUsecase usecase.ProjectUsecase) *ProjectController {
	return &ProjectController{
		projectUsecase: projectUsecase,
	}
}

func (ctrl *ProjectController) Create(c *fiber.Ctx) error {
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

	namaProjectsStr := form.Value["nama_project"]
	kategoriStr := form.Value["kategori"]
	semestersStr := form.Value["semester"]

	if len(namaProjectsStr) != len(files) || len(semestersStr) != len(files) {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Jumlah file, nama project, dan semester harus sama", nil)
	}

	var semesters []int
	for _, semStr := range semestersStr {
		sem, err := strconv.Atoi(semStr)
		if err != nil || sem < 1 || sem > 8 {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Semester harus antara 1 sampai 8", nil)
		}
		semesters = append(semesters, sem)
	}

	for _, nama := range namaProjectsStr {
		if nama == "" {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Nama project wajib diisi", nil)
		}
	}

	result, err := ctrl.projectUsecase.Create(userID, userEmail, userRole, files, namaProjectsStr, kategoriStr, semesters)
	if err != nil {
		if err.Error() == "file harus berupa zip" {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusCreated, "Project berhasil dibuat", result)
}

func (ctrl *ProjectController) GetList(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var params domain.ProjectListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Parameter query tidak valid", nil)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	result, err := ctrl.projectUsecase.GetList(userID, params.Search, params.FilterSemester, params.FilterKategori, params.Page, params.Limit)
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar project: "+err.Error(), nil)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Daftar project berhasil diambil", result)
}

func (ctrl *ProjectController) Update(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID project tidak valid", nil)
	}

	namaProject := c.FormValue("nama_project")
	kategori := c.FormValue("kategori")
	semesterStr := c.FormValue("semester")

	var semester int
	if semesterStr != "" {
		semester, err = strconv.Atoi(semesterStr)
		if err != nil || semester < 1 || semester > 8 {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Semester harus antara 1 sampai 8", nil)
		}
	}

	file, _ := c.FormFile("file")

	result, err := ctrl.projectUsecase.Update(uint(projectID), userID, namaProject, kategori, semester, file)
	if err != nil {
		if err.Error() == "project tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke project ini" {
			return helper.SendForbiddenResponse(c)
		}
		if err.Error() == "file harus berupa zip" {
			return helper.SendErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Project berhasil diperbarui", result)
}

func (ctrl *ProjectController) Delete(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID project tidak valid", nil)
	}

	err = ctrl.projectUsecase.Delete(uint(projectID), userID)
	if err != nil {
		if err.Error() == "project tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		if err.Error() == "tidak memiliki akses ke project ini" {
			return helper.SendForbiddenResponse(c)
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return helper.SendSuccessResponse(c, fiber.StatusOK, "Project berhasil dihapus", nil)
}

func (ctrl *ProjectController) Download(c *fiber.Ctx) error {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return helper.SendUnauthorizedResponse(c)
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		return helper.SendUnauthorizedResponse(c)
	}

	var req domain.ProjectDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "Format request tidak valid", nil)
	}

	if len(req.IDs) == 0 {
		return helper.SendErrorResponse(c, fiber.StatusBadRequest, "ID project tidak boleh kosong", nil)
	}

	filePath, err := ctrl.projectUsecase.Download(userID, req.IDs)
	if err != nil {
		if err.Error() == "project tidak ditemukan" {
			return helper.SendNotFoundResponse(c, err.Error())
		}
		return helper.SendInternalServerErrorResponse(c)
	}

	return c.Download(filePath)
}
