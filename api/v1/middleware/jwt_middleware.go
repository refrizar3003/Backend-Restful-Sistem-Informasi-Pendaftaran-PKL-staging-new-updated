package middleware

import (
	"internship-telkom/service"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTMiddleware(jwtSrv service.JwtService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return unauthorizedResponse(c, "Authorization header is missing")
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return unauthorizedResponse(c, "Invalid authorization format, use: Bearer <token>")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return unauthorizedResponse(c, "Token is missing")
		}

		claims, err := jwtSrv.ValidateToken(tokenString)
		if err != nil {
			return unauthorizedResponse(c, "Invalid or expired token")
		}

		scope, ok := claims["scope"].(string)
		if !ok || scope != "login" {
			return unauthorizedResponse(c, "Invalid token scope")
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return unauthorizedResponse(c, "Invalid token claims")
		}

		role, ok := claims["role"].(string)
		if !ok || role == "" {
			role = "applicant"
		}
		
		c.Locals("user_id", uint(userIDFloat))
		c.Locals("role", role)
		return c.Next()
	}
}

func unauthorizedResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}