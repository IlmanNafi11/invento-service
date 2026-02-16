package base

import (
	"errors"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"invento-service/internal/rbac"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// BaseController provides common functionality for all HTTP controllers.
// It wraps RBAC authorization, validation, and response helpers.
type BaseController struct {
	SupabaseURL string
	Casbin      *rbac.CasbinEnforcer
	Validator   *validator.Validate
}

// NewBaseController creates a new base controller instance.
// All parameters are optional; pass nil for components not needed by the controller.
func NewBaseController(supabaseURL string, casbin *rbac.CasbinEnforcer) *BaseController {
	return &BaseController{
		SupabaseURL: supabaseURL,
		Casbin:      casbin,
		Validator:   validator.New(),
	}
}

// GetAuthenticatedUserID extracts the authenticated user ID from context locals.
// Returns empty string and sends unauthorized response if user ID is not found or invalid.
//
// Usage:
//
//	userID := ctrl.GetAuthenticatedUserID(c)
//	if userID == "" {
//	    return nil // response already sent
//	}
func (bc *BaseController) GetAuthenticatedUserID(c *fiber.Ctx) string {
	userIDVal := c.Locals(middleware.LocalsKeyUserID)
	if userIDVal == nil {
		httputil.SendUnauthorizedResponse(c)
		return ""
	}
	userID, ok := userIDVal.(string)
	if !ok {
		httputil.SendUnauthorizedResponse(c)
		return ""
	}
	return userID
}

// GetAuthenticatedUserEmail extracts the authenticated user email from context locals.
// Returns empty string and sends unauthorized response if email is not found or invalid.
func (bc *BaseController) GetAuthenticatedUserEmail(c *fiber.Ctx) string {
	emailVal := c.Locals(middleware.LocalsKeyUserEmail)
	if emailVal == nil {
		httputil.SendUnauthorizedResponse(c)
		return ""
	}
	email, ok := emailVal.(string)
	if !ok || email == "" {
		httputil.SendUnauthorizedResponse(c)
		return ""
	}
	return email
}

// GetAuthenticatedUserRole extracts the authenticated user role from context locals.
// Returns empty string and sends unauthorized response if role is not found or invalid.
func (bc *BaseController) GetAuthenticatedUserRole(c *fiber.Ctx) string {
	roleVal := c.Locals(middleware.LocalsKeyUserRole)
	if roleVal == nil {
		httputil.SendUnauthorizedResponse(c)
		return ""
	}
	role, ok := roleVal.(string)
	if !ok || role == "" {
		httputil.SendUnauthorizedResponse(c)
		return ""
	}
	return role
}

// CheckPermission performs RBAC authorization check using Casbin.
// Returns nil if permission is granted, otherwise sends appropriate error response.
//
// Parameters:
//   - c: Fiber context
//   - resource: Resource identifier (e.g., "projects", "modules")
//   - action: Action identifier (e.g., "read", "write", "delete")
//
// Returns error if check fails (authorization or internal error), nil if authorized.
func (bc *BaseController) CheckPermission(c *fiber.Ctx, resource, action string) error {
	if bc.Casbin == nil {
		httputil.SendInternalServerErrorResponse(c)
		return errors.New("casbin enforcer not configured")
	}

	role := bc.GetAuthenticatedUserRole(c)
	if role == "" {
		return errors.New("failed to get user role")
	}

	allowed, err := bc.Casbin.CheckPermission(role, resource, action)
	if err != nil {
		httputil.SendInternalServerErrorResponse(c)
		return err
	}

	if !allowed {
		httputil.SendForbiddenResponse(c)
		return errors.New("permission denied")
	}

	return nil
}

// ParsePathID parses an ID parameter from the URL path.
// Returns the parsed ID or an error if parsing fails.
//
// Usage:
//
//	id, err := ctrl.ParsePathID(c)
//	if err != nil {
//	    return err // error response already sent
//	}
func (bc *BaseController) ParsePathID(c *fiber.Ctx) (uint, error) {
	idParam := c.Params("id")
	if idParam == "" {
		httputil.SendBadRequestResponse(c, "ID tidak valid")
		return 0, errors.New("id parameter is empty")
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		httputil.SendBadRequestResponse(c, "ID tidak valid")
		return 0, errors.New("id is not a valid number")
	}

	return uint(id), nil
}

// ParsePathUUID parses a UUID string from the URL path parameter.
// Returns an error and sends a BadRequest response if the parameter is missing, empty, or not a valid UUID.
//
// Usage:
//
//	id, err := ctrl.ParsePathUUID(c)
//	if err != nil {
//	    return err // error response already sent
//	}
func (bc *BaseController) ParsePathUUID(c *fiber.Ctx) (string, error) {
	idParam := c.Params("id")
	if idParam == "" {
		httputil.SendBadRequestResponse(c, "ID tidak valid")
		return "", errors.New("id parameter is empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(idParam); err != nil {
		httputil.SendBadRequestResponse(c, "ID tidak valid")
		return "", errors.New("id is not a valid UUID")
	}

	return idParam, nil
}

// ParsePagination parses page and limit parameters from query string.
// Returns defaults (page=1, limit=10) if parameters are invalid or missing.
//
// Usage:
//
//	page, limit, err := ctrl.ParsePagination(c)
//	if err != nil {
//	    return err // error response already sent
//	}
func (bc *BaseController) ParsePagination(c *fiber.Ctx) (page, limit int, err error) {
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	page, err = strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	return page, limit, nil
}

// SendSuccess sends a success response with consistent structure.
// Uses httputil.StatusOK (200) as default status code.
//
// Parameters:
//   - c: Fiber context
//   - data: Response data (can be nil)
//   - message: Success message in Indonesian
func (bc *BaseController) SendSuccess(c *fiber.Ctx, data interface{}, message string) error {
	return httputil.SendSuccessResponse(c, httputil.StatusOK, message, data)
}

// SendCreated sends a created response (201) with consistent structure.
//
// Parameters:
//   - c: Fiber context
//   - data: Response data (can be nil)
//   - message: Success message in Indonesian
func (bc *BaseController) SendCreated(c *fiber.Ctx, data interface{}, message string) error {
	return httputil.SendSuccessResponse(c, httputil.StatusCreated, message, data)
}

// SendError sends an error response with consistent structure.
// If err is nil, sends internal server error response.
//
// Parameters:
//   - c: Fiber context
//   - err: Error object (used for message, not exposed to client)
//   - defaultMessage: Fallback message in Indonesian if err.Error() is not user-friendly
func (bc *BaseController) SendError(c *fiber.Ctx, err error, defaultMessage string) error {
	if err == nil {
		httputil.SendInternalServerErrorResponse(c)
		return nil
	}

	// Use provided default message, or fallback to generic message
	message := defaultMessage
	if message == "" {
		message = httputil.GetDefaultMessage(httputil.StatusInternalServerError)
	}

	return httputil.SendErrorResponse(c, httputil.StatusInternalServerError, message, nil)
}

// SendBadRequest sends a bad request response (400).
//
// Parameters:
//   - c: Fiber context
//   - message: Error message in Indonesian
func (bc *BaseController) SendBadRequest(c *fiber.Ctx, message string) error {
	if message == "" {
		message = httputil.GetDefaultMessage(httputil.StatusBadRequest)
	}
	return httputil.SendBadRequestResponse(c, message)
}

// SendUnauthorized sends an unauthorized response (401).
func (bc *BaseController) SendUnauthorized(c *fiber.Ctx) error {
	return httputil.SendUnauthorizedResponse(c)
}

// SendForbidden sends a forbidden response (403).
func (bc *BaseController) SendForbidden(c *fiber.Ctx) error {
	return httputil.SendForbiddenResponse(c)
}

// SendNotFound sends a not found response (404).
//
// Parameters:
//   - c: Fiber context
//   - message: Error message in Indonesian (uses default if empty)
func (bc *BaseController) SendNotFound(c *fiber.Ctx, message string) error {
	return httputil.SendNotFoundResponse(c, message)
}

// SendConflict sends a conflict response (409).
//
// Parameters:
//   - c: Fiber context
//   - message: Error message in Indonesian
func (bc *BaseController) SendConflict(c *fiber.Ctx, message string) error {
	return httputil.SendConflictResponse(c, message)
}

// SendInternalError sends an internal server error response (500).
func (bc *BaseController) SendInternalError(c *fiber.Ctx) error {
	return httputil.SendInternalServerErrorResponse(c)
}

// ValidateStruct validates a struct using the validator.
// Returns true if valid, false if validation fails (response already sent).
//
// Usage:
//
//	var req domain.CreateUserRequest
//	if err := c.BodyParser(&req); err != nil {
//	    return ctrl.SendBadRequest(c, "Format request tidak valid")
//	}
//	if !ctrl.Validate(c, req) {
//	    return nil // validation error response already sent
//	}
func (bc *BaseController) ValidateStruct(c *fiber.Ctx, data interface{}) bool {
	validationErrors := httputil.ValidateStruct(data)
	if len(validationErrors) > 0 {
		httputil.SendValidationErrorResponse(c, validationErrors)
		return false
	}
	return true
}
