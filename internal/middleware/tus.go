package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func TusProtocolMiddleware(tusVersion string, maxSize int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()

		if method == "OPTIONS" {
			c.Set("Tus-Resumable", tusVersion)
			c.Set("Tus-Version", tusVersion)
			c.Set("Tus-Extension", "creation,termination")
			c.Set("Tus-Max-Size", strconv.FormatInt(maxSize, 10))
			return c.SendStatus(fiber.StatusNoContent)
		}

		tusResumable := c.Get("Tus-Resumable")
		if method != "GET" && method != "POST" && tusResumable == "" {
			c.Set("Tus-Resumable", tusVersion)
			return c.SendStatus(fiber.StatusPreconditionFailed)
		}

		if tusResumable != "" && tusResumable != tusVersion {
			c.Set("Tus-Resumable", tusVersion)
			return c.SendStatus(fiber.StatusPreconditionFailed)
		}

		return c.Next()
	}
}
