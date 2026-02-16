package httputil

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestParsePaginationQuery_Success tests pagination query parsing
func TestParsePaginationQuery_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	// Test default values
	app.Get("/test1", func(c *fiber.Ctx) error {
		page, limit := ParsePaginationQuery(c)
		assert.Equal(t, 1, page)
		assert.Equal(t, 10, limit)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with values
	app.Get("/test2", func(c *fiber.Ctx) error {
		page, limit := ParsePaginationQuery(c)
		assert.Equal(t, 2, page)
		assert.Equal(t, 20, limit)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test2?page=2&limit=20", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with invalid values (should return defaults)
	app.Get("/test3", func(c *fiber.Ctx) error {
		page, limit := ParsePaginationQuery(c)
		assert.Equal(t, 1, page)
		assert.Equal(t, 10, limit)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test3?page=abc&limit=xyz", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test max limit cap
	app.Get("/test4", func(c *fiber.Ctx) error {
		page, limit := ParsePaginationQuery(c)
		assert.Equal(t, 1, page)
		assert.Equal(t, 100, limit) // capped at 100
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test4?page=1&limit=200", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestParseSearchQuery_Success tests search query parsing
func TestParseSearchQuery_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	// Test empty search
	app.Get("/test1", func(c *fiber.Ctx) error {
		search := ParseSearchQuery(c)
		assert.Equal(t, "", search)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with search term
	app.Get("/test2", func(c *fiber.Ctx) error {
		search := ParseSearchQuery(c)
		assert.Equal(t, "test", search)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test2?search=test", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestParseFilterQuery_Success tests filter query parsing
func TestParseFilterQuery_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	// Test empty filter
	app.Get("/test1", func(c *fiber.Ctx) error {
		filter := ParseFilterQuery(c, "category")
		assert.Equal(t, "", filter)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with filter value
	app.Get("/test2", func(c *fiber.Ctx) error {
		filter := ParseFilterQuery(c, "category")
		assert.Equal(t, "active", filter)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test2?category=active", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestParseIntQuery_Success tests integer query parsing
func TestParseIntQuery_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	// Test default value
	app.Get("/test1", func(c *fiber.Ctx) error {
		result := ParseIntQuery(c, "semester", 0)
		assert.Equal(t, 0, result)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with valid value
	app.Get("/test2", func(c *fiber.Ctx) error {
		result := ParseIntQuery(c, "semester", 0)
		assert.Equal(t, 2, result)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test2?semester=2", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with invalid value (should return default)
	app.Get("/test3", func(c *fiber.Ctx) error {
		result := ParseIntQuery(c, "semester", 0)
		assert.Equal(t, 0, result)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test3?semester=abc", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestParseBoolQuery_Success tests boolean query parsing
func TestParseBoolQuery_Success(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	// Test default value (false)
	app.Get("/test1", func(c *fiber.Ctx) error {
		result := ParseBoolQuery(c, "active", false)
		assert.False(t, result)
		return c.SendString("ok")
	})

	req, _ := http.NewRequest("GET", "/test1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with true
	app.Get("/test2", func(c *fiber.Ctx) error {
		result := ParseBoolQuery(c, "active", false)
		assert.True(t, result)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test2?active=true", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with false
	app.Get("/test3", func(c *fiber.Ctx) error {
		result := ParseBoolQuery(c, "active", true)
		assert.False(t, result)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test3?active=false", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with invalid value (should return default)
	app.Get("/test4", func(c *fiber.Ctx) error {
		result := ParseBoolQuery(c, "active", false)
		assert.False(t, result)
		return c.SendString("ok")
	})

	req, _ = http.NewRequest("GET", "/test4?active=invalid", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
