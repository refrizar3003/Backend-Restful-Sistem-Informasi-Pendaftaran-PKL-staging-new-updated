package api

import (
	"internship-telkom/api/v1/controller"
	"internship-telkom/api/v1/middleware"
	"internship-telkom/infrastructure"
	"internship-telkom/repository"
	"internship-telkom/service"
	"time"

	_ "internship-telkom/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/swagger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SetupRoutes mengatur semua rute
func SetupRoutes(app *fiber.App, db *gorm.DB, rdb *redis.Client, mailer infrastructure.Mailer, jwtSrv service.JwtService,) {
	app.Use(middleware.CORSMiddleware())
	// =========================================================
	// Inisialisasi Infrastructure
	// =========================================================
	cld, err := infrastructure.NewCloudinary()
	if err != nil {
		panic("failed to initialize Cloudinary: " + err.Error())
	}

	// =========================================================
	// Inisialisasi User Layer
	// =========================================================
	userRepo := repository.NewUserRepository(db)
	notifRepo    := repository.NewNotificationRepository(db)
	deviceRepo := repository.NewTrustedDeviceRepository(db)
	userService := service.NewUserService(userRepo, rdb, cld, jwtSrv, mailer, deviceRepo)
	userController := controller.NewUserController(userService)

	// =========================================================
	// Inisialisasi Application Layer (sisi applicant)
	// =========================================================
	appRepo       := repository.NewApplicationRepository(db)
	appService    := service.NewApplicationService(appRepo, cld, userRepo, notifRepo)
	appController := controller.NewApplicationController(appService)

	// =========================================================
	// Inisialisasi Admin Layer
	// =========================================================
	unitRepo := repository.NewUnitRepository(db)
	mentorRepo := repository.NewMentorRepository(db)
	letterRepo := repository.NewLetterRepository(db)

	letterService := service.NewLetterService(appRepo, letterRepo, cld)
	letterController := controller.NewLetterController(letterService)

	adminAppService := service.NewAdminApplicationService(appRepo, unitRepo, mentorRepo, notifRepo, userRepo, mailer, letterService)
	adminAppController := controller.NewAdminApplicationController(adminAppService)

	notifService := service.NewNotificationService(notifRepo)
	notifController := controller.NewNotificationController(notifService)

	// =========================================================
	// Inisialisasi Employee (Kelola Pegawai) Layer
	// =========================================================
	employeeService := service.NewEmployeeService(userRepo, cld)
	employeeController := controller.NewEmployeeController(employeeService)

	// =========================================================
	// Setup Routing
	// =========================================================
	v1 := app.Group("/api/v1")

	loginLimiter := limiter.New(limiter.Config{
		Max:          10,
		Expiration:   1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"status":  "error",
			"message": "Too many requests, please try again later",
			})
		},
	})

	// Public routes
	v1.Post("/register", userController.Register)
	v1.Post("/login", loginLimiter, userController.Login)
	v1.Post("/login/verify-otp", loginLimiter, userController.VerifyLoginOTP)
	v1.Post("/login/resend-otp", loginLimiter, userController.ResendLoginOTP)
	v1.Get("/verify", userController.VerifyAccount)
	v1.Post("/resend-verification", userController.ResendVerification)
	v1.Post("/forgot-password", userController.ForgotPassword)
	v1.Post("/resend-otp", userController.ResendOTP)
	v1.Post("/verify-code", userController.VerifyCode)
	v1.Post("/reset-password", userController.ResetPassword)

	protected := v1.Group("", middleware.JWTMiddleware(jwtSrv))

	profile := protected.Group("/profile")
	profile.Get("/", userController.GetProfile)
	profile.Put("/", userController.UpdateProfile)

	applications := protected.Group("/applications")
	applications.Post("/",   appController.Apply)
	applications.Get("/",    appController.GetMyApplications)
	applications.Get("/:id", appController.GetApplicationDetail)

	// Admin routes
	staff := protected.Group("/admin", middleware.StaffOnly())

	staff.Get("/applications", adminAppController.GetApplications)
	staff.Get("/applications/:id", adminAppController.GetApplicationDetail)
	staff.Patch("/applications/:id/status", adminAppController.UpdateApplicationStatus)
	staff.Post("/applications/:id/letter", letterController.GenerateLetter)

	staff.Get("/units", adminAppController.GetUnits) 
	staff.Get("/units/:unit_id/mentors", adminAppController.GetMentorsByUnit)
	staff.Get("/dashboard/summary", adminAppController.GetDashboardSummary)

	// Kelola Pegawai — hanya Administrator
	adminOnly := protected.Group("/admin", middleware.AdminOnly())

	adminOnly.Post("/employees", employeeController.CreateEmployee)
	adminOnly.Get("/employees", employeeController.GetEmployees)
	adminOnly.Get("/employees/:id", employeeController.GetEmployeeDetail)
	adminOnly.Put("/employees/:id", employeeController.UpdateEmployee)
	adminOnly.Delete("/employees/:id", employeeController.DeleteEmployee)

	// Notifikasi
	notifications := protected.Group("/notifications")
	
	notifications.Get("/",              notifController.GetNotifications)
	notifications.Patch("/read-all",    notifController.MarkAllAsRead)
	notifications.Patch("/:id/read",    notifController.MarkAsRead)

	app.Get("/swagger/*", swagger.HandlerDefault)
}