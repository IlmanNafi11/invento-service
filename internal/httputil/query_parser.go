package httputil

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ParsePaginationQuery(c *fiber.Ctx) (page int, limit int) {
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	return page, limit
}

func ParseSearchQuery(c *fiber.Ctx) string {
	return c.Query("search", "")
}

func ParseFilterQuery(c *fiber.Ctx, filterName string) string {
	return c.Query(filterName, "")
}

func ParseIntQuery(c *fiber.Ctx, key string, defaultValue int) int {
	valueStr := c.Query(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func ParseBoolQuery(c *fiber.Ctx, key string, defaultValue bool) bool {
	valueStr := c.Query(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
