package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func sendFile(path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendFile(path)
	}
}

func main() {
	godotenv.Load()
	port := os.Getenv("DASH_PORT")
	if port == "" {
		port = "5501"
	}

	app := fiber.New(fiber.Config{AppName: "TraceQube Dashboard"})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Static("/static", "./dashboard/static")

	app.Get("/", sendFile("./dashboard/templates/auth.html"))
	app.Get("/login", sendFile("./dashboard/templates/auth.html"))
	app.Get("/oauth-callback", sendFile("./dashboard/templates/oauth-callback.html"))
	app.Get("/register", sendFile("./dashboard/templates/auth.html"))
	app.Get("/dashboard", sendFile("./dashboard/templates/dashboard.html"))
	app.Get("/forgot-password", sendFile("./dashboard/templates/forgot.html"))
	app.Get("/reset-password", sendFile("./dashboard/templates/reset.html"))
	app.Get("/admin", sendFile("./dashboard/templates/admin.html"))

	log.Printf("Dashboard starting on port %s", port)
	log.Fatal(app.Listen(":"+port))
}