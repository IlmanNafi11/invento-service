package http

import (
	"invento-service/internal/controller/base"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/helper"
	"invento-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

// RoleController handles role-related HTTP requests.
// It embeds BaseController for common functionality like authentication,
// authorization, validation, and response helpers.
type RoleController struct {
	*base.BaseController
	roleUsecase usecase.RoleUsecase
}

// NewRoleController creates a new RoleController instance.
// Parameters:
//   - roleUsecase: Role business logic layer
//   - baseController: Base controller for common functionality (optional)
func NewRoleController(roleUsecase usecase.RoleUsecase, baseController *base.BaseController) *RoleController {
	if baseController == nil {
		baseController = base.NewBaseController("", nil)
	}

	return &RoleController{
		BaseController: baseController,
		roleUsecase:    roleUsecase,
	}
}

// GetAvailablePermissions retrieves all available permissions grouped by resource.
//
// @Summary  Get available permissions
// @Description Retrieves all available permissions grouped by resource
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Success  200 {object} domain.SuccessResponse "Permissions retrieved successfully"
// @Failure  500 {object} domain.ErrorResponse "Internal server error"
// @Router   /role/permissions [get]
// @Security  BearerAuth
func (ctrl *RoleController) GetAvailablePermissions(c *fiber.Ctx) error {
	permissions, err := ctrl.roleUsecase.GetAvailablePermissions()
	if err != nil {
		return ctrl.SendInternalError(c)
	}

	response := map[string]interface{}{
		"items": permissions,
	}

	return ctrl.SendSuccess(c, response, "Daftar resource dan permission berhasil diambil")
}

// GetRoleList retrieves a paginated list of roles.
//
// @Summary  Get role list
// @Description Retrieves a paginated list of roles with optional search
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   search query  string false "Search keyword"
// @Param   page  query  int  false "Page number"  default(1)
// @Param   limit  query  int  false "Items per page" default(10)
// @Success  200  {object} domain.SuccessResponse "Roles retrieved successfully"
// @Failure  400  {object} domain.ErrorResponse "Invalid query parameters"
// @Failure  500  {object} domain.ErrorResponse "Internal server error"
// @Router   /role [get]
// @Security  BearerAuth
func (ctrl *RoleController) GetRoleList(c *fiber.Ctx) error {
	var params domain.RoleListQueryParams
	if err := c.QueryParser(&params); err != nil {
		return ctrl.SendBadRequest(c, "Parameter query tidak valid")
	}

	result, err := ctrl.roleUsecase.GetRoleList(params)
	if err != nil {
		return ctrl.SendInternalError(c)
	}

	return ctrl.SendSuccess(c, result, "Daftar role berhasil diambil")
}

// CreateRole creates a new role with specified permissions.
//
// @Summary  Create role
// @Description Creates a new role with the specified name and permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   request body  domain.RoleCreateRequest true "Role creation request"
// @Success  201  {object} domain.SuccessResponse "Role created successfully"
// @Failure  400  {object} domain.ErrorResponse "Invalid request format"
// @Failure  409  {object} domain.ErrorResponse "Role name already exists"
// @Failure  500  {object} domain.ErrorResponse "Internal server error"
// @Router   /role [post]
// @Security  BearerAuth
func (ctrl *RoleController) CreateRole(c *fiber.Ctx) error {
	var req domain.RoleCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	result, err := ctrl.roleUsecase.CreateRole(req)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendCreated(c, result, "Role berhasil dibuat")
}

// GetRoleDetail retrieves details of a specific role by ID.
//
// @Summary  Get role detail
// @Description Retrieves detailed information about a specific role including its permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   id path  int true "Role ID"
// @Success  200 {object} domain.SuccessResponse "Role details retrieved successfully"
// @Failure  400 {object} domain.ErrorResponse "Invalid role ID"
// @Failure  404 {object} domain.ErrorResponse "Role not found"
// @Failure  500 {object} domain.ErrorResponse "Internal server error"
// @Router   /role/{id} [get]
// @Security  BearerAuth
func (ctrl *RoleController) GetRoleDetail(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	result, err := ctrl.roleUsecase.GetRoleDetail(id)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendSuccess(c, result, "Detail role berhasil diambil")
}

// UpdateRole updates an existing role by ID.
//
// @Summary  Update role
// @Description Updates an existing role's name and permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   id  path  int      true "Role ID"
// @Param   request body  domain.RoleUpdateRequest true "Role update request"
// @Success  200  {object} domain.SuccessResponse "Role updated successfully"
// @Failure  400  {object} domain.ErrorResponse "Invalid request format"
// @Failure  404  {object} domain.ErrorResponse "Role not found"
// @Failure  409  {object} domain.ErrorResponse "Role name already exists"
// @Failure  500  {object} domain.ErrorResponse "Internal server error"
// @Router   /role/{id} [put]
// @Security  BearerAuth
func (ctrl *RoleController) UpdateRole(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	var req domain.RoleUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ctrl.SendBadRequest(c, "Format request tidak valid")
	}

	if !ctrl.ValidateStruct(c, req) {
		return nil
	}

	result, err := ctrl.roleUsecase.UpdateRole(id, req)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendSuccess(c, result, "Role berhasil diperbarui")
}

// DeleteRole deletes a role by ID.
//
// @Summary  Delete role
// @Description Deletes a role and removes all associated permissions
// @Tags   Roles
// @Accept   json
// @Produce  json
// @Param   id path int true "Role ID"
// @Success  200 {object} domain.SuccessResponse "Role deleted successfully"
// @Failure  400 {object} domain.ErrorResponse "Invalid role ID"
// @Failure  404 {object} domain.ErrorResponse "Role not found"
// @Failure  500 {object} domain.ErrorResponse "Internal server error"
// @Router   /role/{id} [delete]
// @Security  BearerAuth
func (ctrl *RoleController) DeleteRole(c *fiber.Ctx) error {
	id, err := ctrl.ParsePathID(c)
	if err != nil {
		return err
	}

	err = ctrl.roleUsecase.DeleteRole(id)
	if err != nil {
		return ctrl.handleRoleError(c, err)
	}

	return ctrl.SendSuccess(c, nil, "Role berhasil dihapus")
}

// handleRoleError handles role usecase errors and maps them to appropriate HTTP responses.
// Uses type-safe error handling with AppError types.
func (ctrl *RoleController) handleRoleError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return helper.SendAppError(c, appErr)
	}
	return ctrl.SendInternalError(c)
}
