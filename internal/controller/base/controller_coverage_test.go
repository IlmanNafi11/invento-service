package base

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewBaseController_Success tests successful base controller creation
func TestNewBaseController_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)

	assert.NotNil(t, controller)
	assert.NotNil(t, controller.Validator)
}

// TestBaseController_ParsePathID_Success tests successful ID parsing
func TestBaseController_ParsePathID_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test/:id", func(c *fiber.Ctx) error {
		id, err := controller.ParsePathID(c)
		assert.NoError(t, err)
		assert.Equal(t, uint(123), id)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test/123", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_ParsePathID_Empty_Success tests ID parsing with empty ID
func TestBaseController_ParsePathID_Empty_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		_, err := controller.ParsePathID(c)
		assert.Error(t, err)
		return nil
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestBaseController_ParsePathID_Invalid_Success tests ID parsing with invalid ID
func TestBaseController_ParsePathID_Invalid_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test/:id", func(c *fiber.Ctx) error {
		_, err := controller.ParsePathID(c)
		assert.Error(t, err)
		return nil
	})

	req, _ := http.NewRequest("GET", "/test/invalid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestBaseController_ParsePagination_Default_Success tests pagination parsing with defaults
func TestBaseController_ParsePagination_Default_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		page, limit, err := controller.ParsePagination(c)
		assert.NoError(t, err)
		assert.Equal(t, 1, page)
		assert.Equal(t, 10, limit)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_ParsePagination_WithParams_Success tests pagination with parameters
func TestBaseController_ParsePagination_WithParams_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		page, limit, err := controller.ParsePagination(c)
		assert.NoError(t, err)
		assert.Equal(t, 2, page)
		assert.Equal(t, 20, limit)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test?page=2&limit=20", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_ParsePagination_Invalid_Success tests pagination with invalid parameters
func TestBaseController_ParsePagination_Invalid_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		page, limit, err := controller.ParsePagination(c)
		assert.NoError(t, err)
		assert.Equal(t, 1, page) // defaults to 1
		assert.Equal(t, 10, limit) // defaults to 10
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test?page=abc&limit=xyz", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_SendSuccess_Success tests successful success response
func TestBaseController_SendSuccess_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		return controller.SendSuccess(c, map[string]string{"key": "value"}, "Berhasil")
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_SendCreated_Success tests successful created response
func TestBaseController_SendCreated_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		return controller.SendCreated(c, map[string]string{"id": "123"}, "Data berhasil dibuat")
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

// TestBaseController_SendBadRequest_Success tests bad request response
func TestBaseController_SendBadRequest_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return controller.SendBadRequest(c, "Request tidak valid")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestBaseController_SendUnauthorized_Success tests unauthorized response
func TestBaseController_SendUnauthorized_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return controller.SendUnauthorized(c)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBaseController_SendForbidden_Success tests forbidden response
func TestBaseController_SendForbidden_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return controller.SendForbidden(c)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

// TestBaseController_SendNotFound_Success tests not found response
func TestBaseController_SendNotFound_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return controller.SendNotFound(c, "Data tidak ditemukan")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

// TestBaseController_SendConflict_Success tests conflict response
func TestBaseController_SendConflict_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		return controller.SendConflict(c, "Data sudah ada")
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
}

// TestBaseController_SendInternalError_Success tests internal server error response
func TestBaseController_SendInternalError_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return controller.SendInternalError(c)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

// TestBaseController_SendError_Success tests error response
func TestBaseController_SendError_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return controller.SendError(c, assert.AnError, "Terjadi kesalahan")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

// TestBaseController_SendError_Nil tests error response with nil error
func TestBaseController_SendError_Nil(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		return controller.SendError(c, nil, "")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

// TestBaseController_ValidateStruct_Success tests struct validation
func TestBaseController_ValidateStruct_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	type TestStruct struct {
		Email string `validate:"required,email"`
	}

	app.Post("/test", func(c *fiber.Ctx) error {
		data := TestStruct{Email: "test@example.com"}
		valid := controller.ValidateStruct(c, data)
		assert.True(t, valid)
		return c.SendStatus(200)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_ValidateStruct_Failure tests struct validation failure
func TestBaseController_ValidateStruct_Failure(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	type TestStruct struct {
		Email string `validate:"required,email"`
	}

	app.Post("/test", func(c *fiber.Ctx) error {
		data := TestStruct{Email: "invalid-email"}
		valid := controller.ValidateStruct(c, data)
		assert.False(t, valid)
		return nil
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestBaseController_GetAuthenticatedUserID_Success tests getting user ID from context
func TestBaseController_GetAuthenticatedUserID_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(123))
		userID := controller.GetAuthenticatedUserID(c)
		assert.Equal(t, uint(123), userID)
		return c.SendStatus(200)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_GetAuthenticatedUserID_Missing tests missing user ID
func TestBaseController_GetAuthenticatedUserID_Missing(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		userID := controller.GetAuthenticatedUserID(c)
		assert.Equal(t, uint(0), userID)
		return nil
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBaseController_GetAuthenticatedUserID_InvalidType tests invalid user ID type
func TestBaseController_GetAuthenticatedUserID_InvalidType(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", "not-a-uint")
		userID := controller.GetAuthenticatedUserID(c)
		assert.Equal(t, uint(0), userID)
		return nil
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBaseController_GetAuthenticatedUserEmail_Success tests getting user email from context
func TestBaseController_GetAuthenticatedUserEmail_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_email", "test@example.com")
		email := controller.GetAuthenticatedUserEmail(c)
		assert.Equal(t, "test@example.com", email)
		return c.SendStatus(200)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_GetAuthenticatedUserEmail_Missing tests missing user email
func TestBaseController_GetAuthenticatedUserEmail_Missing(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		email := controller.GetAuthenticatedUserEmail(c)
		assert.Equal(t, "", email)
		return nil
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBaseController_GetAuthenticatedUserRole_Success tests getting user role from context
func TestBaseController_GetAuthenticatedUserRole_Success(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_role", "admin")
		role := controller.GetAuthenticatedUserRole(c)
		assert.Equal(t, "admin", role)
		return c.SendStatus(200)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestBaseController_GetAuthenticatedUserRole_Missing tests missing user role
func TestBaseController_GetAuthenticatedUserRole_Missing(t *testing.T) {
	controller := NewBaseController(nil, nil)
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		role := controller.GetAuthenticatedUserRole(c)
		assert.Equal(t, "", role)
		return nil
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

// TestBaseController_CheckPermission_NoCasbin tests check permission without casbin
func TestBaseController_CheckPermission_NoCasbin(t *testing.T) {
	controller := NewBaseController(nil, nil) // No casbin
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		err := controller.CheckPermission(c, "projects", "read")
		assert.Error(t, err)
		return nil
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

