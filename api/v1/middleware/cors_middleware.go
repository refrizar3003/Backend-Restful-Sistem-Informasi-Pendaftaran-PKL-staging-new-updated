package middleware

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORSMiddleware() fiber.Handler {
	allowedOrigins := os.Getenv("CORS_ORIGIN")
	if allowedOrigins == "" {
		log.Println("Warning: CORS_ORIGIN is not set, defaulting to http://localhost:3000")
		allowedOrigins = "http://localhost:3000"
	}

	origins := strings.Split(allowedOrigins, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}
	cleanedOrigins := strings.Join(origins, ",")

	return cors.New(cors.Config{
		AllowOrigins:     cleanedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	})
}