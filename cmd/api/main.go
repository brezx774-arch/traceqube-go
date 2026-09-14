package main

import (
	"log"
	"os"

	"traceqube/internal/api"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"context"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.api")

	if err := api.InitDB(); err != nil {
		log.Fatal("DB connection failed:", err)
	}
	log.Println("Database connected")

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "4000"
	}

	app := fiber.New(fiber.Config{AppName: "TraceQube API v2"})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{AllowOrigins: "https://traceqube.com, https://app.traceqube.com, https://api.traceqube.com", AllowHeaders: "Origin, Content-Type, Authorization, X-API-Key", AllowCredentials: true}))


	app.Get("/debug/smtp", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"smtp_user": os.Getenv("SMTP_USER"),
			"smtp_host": os.Getenv("SMTP_HOST"),
			"smtp_port": os.Getenv("SMTP_PORT"),
			"has_pass":  os.Getenv("SMTP_PASS") != "",
		})
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "TraceQube API running", "version": "2.0"})
	})

	// Auth routes
	app.Post("/api/register", api.Register)
	app.Post("/api/login", api.Login)
	app.Get("/api/verify-email", api.VerifyEmail)
	app.Post("/api/forgot-password", api.ForgotPassword)
	app.Post("/api/reset-password", api.ResetPassword)

	// OAuth routes
	app.Get("/auth/google", api.GoogleLogin)
	app.Get("/auth/google/callback", api.GoogleCallback)
	app.Get("/auth/github", api.GithubLogin)
	app.Get("/auth/github/callback", api.GithubCallback)

	// Protected routes
	auth := app.Group("/api", api.AuthMiddleware)
	auth.Get("/me", api.GetMe)
	auth.Post("/me/regenerate-key", api.RegenerateKey)
	auth.Post("/me/logout-all", api.LogoutAllDevices)

	// API tool routes (v1)
	v1 := app.Group("/v1", api.AuthMiddleware)
	v1.Get("/ip", api.ToolIP)
	v1.Get("/dns", api.ToolDNS)
	v1.Get("/ping", api.ToolPing)
	v1.Get("/port", api.ToolPort)
	v1.Get("/ssl", api.ToolSSL)
	v1.Get("/whois", api.ToolWhois)
	v1.Get("/geo", api.ToolGeo)
	v1.Get("/rdns", api.ToolRDNS)
	v1.Get("/subnet", api.ToolSubnet)
	v1.Get("/headers", api.ToolHeaders)
	v1.Get("/down", api.ToolDown)

	// Admin routes
	admin := app.Group("/api/admin", api.AuthMiddleware, api.AdminMiddleware)
	admin.Get("/stats", api.AdminGetStats)
	admin.Get("/users", api.AdminGetUsers)
	admin.Patch("/users/:id/plan", api.AdminUpdatePlan)
	admin.Patch("/users/:id/reset-usage", api.AdminResetUsage)
	admin.Delete("/users/:id", api.AdminDeleteUser)
	admin.Post("/reset-usage-all", api.AdminResetAllUsage)


	app.Post("/internal/reset-usage", func(c *fiber.Ctx) error {
		if c.IP() != "127.0.0.1" {
			return c.Status(403).JSON(fiber.Map{"error": "localhost only"})
		}
		api.DB.Exec(context.Background(), "UPDATE users SET usage = 0")
		return c.JSON(fiber.Map{"success": true})
	})
	log.Printf("TraceQube API starting on port %s", port)
	log.Fatal(app.Listen(":"+port))
}
