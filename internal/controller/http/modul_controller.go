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

type ModulController struct {
	*base.BaseController
	modulUsecase    usecase.ModulUsecase
	tusModulUsecase usecase.TusModulUsecase
	config          *config.Config
}

func NewModulController(
	modulUsecase usecase.ModulUsecase,
	tusModulUsecase usecase.TusModulUsecase,
	config *config.Config,
	baseCtrl *base.BaseController,
) *ModulController {
	return &ModulController{
		BaseController:  baseCtrl,
		modulUsecase:    modulUsecase,
		tusModulUsecase: tusModulUsecase,
		config:          config,
	}
}

func (ctrl *ModulController) GetList(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	var params domain.ModulListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	result, err := ctrl.modulUsecase.GetList(userID, params.Search, params.FilterType, params.FilterSemester, params.Page, params.Limit)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar modul berhasil diambil")
}

func (ctrl *ModulController) Delete(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	modulID, err := ctrl.ParsePathID(c)
	if err != nil {
		return nil
	}

	err = ctrl.modulUsecase.Delete(modulID, userID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Modul berhasil dihapus")
}

func (ctrl *ModulController) UpdateMetadata(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	modulID, err := ctrl.ParsePathID(c)
	if err != nil {
		return nil
	}

	var req domain.ModulUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	err = ctrl.modulUsecase.UpdateMetadata(modulID, userID, req)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, nil, "Metadata modul berhasil diperbarui")
}

func (ctrl *ModulController) Download(c *fiber.Ctx) error {
	userID := ctrl.GetAuthenticatedUserID(c)
	if userID == "" {
		return nil
	}

	var req domain.ModulDownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if len(req.IDs) == 0 {
		return ctrl.SendBadRequest(c, "ID modul tidak boleh kosong")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	filePath, err := ctrl.modulUsecase.Download(userID, req.IDs)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return helper.SendAppError(c, appErr)
		}
		return ctrl.SendInternalError(c)
	}

	return c.Download(filePath)
}
