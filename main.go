package main

import (
	"fmt"
	"internship-telkom/api"
	"internship-telkom/infrastructure"
	"internship-telkom/service"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// @title           API Sistem Informasi Pendaftaran PKL - Telkom Witel Lampung
// @version         1.0
// @description     Dokumentasi API untuk pengelolaan pendaftaran Praktik Kerja Lapangan (PKL).
// @contact.name    Refrizar Dwiardito

// @host      104.43.88.99
// @BasePath  /api/v1

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description                 Use format: Bearer <your_token_jwt>

func main() {
		if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

		// 1. Inisialisasi Koneksi Database
		db, err := infrastructure.ConnectDB()
		if err != nil {
				log.Fatalf("Could not connect to the database: %v", err)
		}
		log.Println("Successfully connected to the database")

		// 2. Inisialisasi Koneksi Redis
		redisClient, err := infrastructure.NewRedisClient()
		if err != nil {
				log.Fatalf("Could not connect to Redis: %v", err)
		} 
		log.Println("Successfully connected to Redis")

		// Inisialisasi Mailer (Infrastruktur)
		mailer := infrastructure.NewGmailMailer(
		os.Getenv("SMTP_SENDER_EMAIL"),
		os.Getenv("SMTP_SENDER_PASSWORD"),
		)

		// Inisialisasi JWT Service
		jwtSrv := service.NewJwtService()

		// 3. Inisialisasi Aplikasi Fiber
		app := fiber.New(fiber.Config{
				AppName: "Sistem Informasi Pendaftaran PKL - Telkom Witel",
		})

		// 4. Setup Routing
		api.SetupRoutes(app, db, redisClient, mailer, jwtSrv)

		// 5. Menjalankan Server
		port := getEnv("APP_PORT", "8080")
		log.Println("Starting the server on port 8080...")
		err = app.Listen(":8080")
		
		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}