package api

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func GetMe(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	return c.JSON(fiber.Map{
		"email":      u.Email,
		"apiKey":     u.APIKey,
		"plan":       u.Plan,
		"usageToday": u.Usage,
		"dailyLimit": PlanLimit(u.Plan),
		"isAdmin":    u.IsAdmin,
		"verified":   u.Verified,
	})
}

func RegenerateKey(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	newKey := generateAPIKey()
	DB.Exec(context.Background(), "UPDATE users SET api_key = $1 WHERE id = $2", newKey, u.ID)
	return c.JSON(fiber.Map{"apiKey": newKey})
}

func LogoutAllDevices(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	DB.Exec(context.Background(), "UPDATE users SET token_version = token_version + 1 WHERE id = $1", u.ID)
	c.ClearCookie("tq_token")
	return c.JSON(fiber.Map{"success": true, "message": "Logged out of all devices"})
}
