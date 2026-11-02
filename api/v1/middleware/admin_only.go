package middleware

import (
	"internship-telkom/model"

	"github.com/gofiber/fiber/v2"
)

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != model.RoleAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Forbidden: admin access only",
			})
		}
		return c.Next()
	}
}

func StaffOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || (role != model.RoleAdmin && role != model.RolePegawai) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Forbidden: staff access only",
			})
		}
		return c.Next()
	}
}