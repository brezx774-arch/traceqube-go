package api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

func AdminGetStats(c *fiber.Ctx) error {
	var total, free, pro, enterprise, verified int
	var totalUsage int64
	DB.QueryRow(context.Background(), `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE plan = 'free'),
			COUNT(*) FILTER (WHERE plan = 'pro'),
			COUNT(*) FILTER (WHERE plan = 'enterprise'),
			COUNT(*) FILTER (WHERE verified = true),
			COALESCE(SUM(usage), 0)
		FROM users`).Scan(&total, &free, &pro, &enterprise, &verified, &totalUsage)
	return c.JSON(fiber.Map{
		"totalUsers": total, "proUsers": pro, "enterpriseUsers": enterprise,
		"requestsToday": totalUsage, "totalRequests": totalUsage,
	})
}

func AdminGetUsers(c *fiber.Ctx) error {
	rows, err := DB.Query(context.Background(),
		"SELECT id, email, plan, usage, verified, is_admin, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()
	var users []fiber.Map
	for rows.Next() {
		var id, email, plan string
		var usage int
		var verified, isAdmin bool
		var createdAt time.Time
		if err := rows.Scan(&id, &email, &plan, &usage, &verified, &isAdmin, &createdAt); err != nil {
			continue
		}
		users = append(users, fiber.Map{
			"id": id, "email": email, "plan": plan,
			"usage": usage, "verified": verified, "is_admin": isAdmin,
			"created_at": createdAt.Format("2006-01-02"),
		})
	}
	if users == nil {
		users = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"users": users, "total": len(users)})
}

func AdminUpdatePlan(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct{ Plan string `json:"plan"` }
	c.BodyParser(&body)
	valid := map[string]bool{"free": true, "pro": true, "enterprise": true}
	if !valid[body.Plan] {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid plan"})
	}
	DB.Exec(context.Background(), "UPDATE users SET plan = $1 WHERE id = $2", body.Plan, id)
	return c.JSON(fiber.Map{"success": true})
}

func AdminResetUsage(c *fiber.Ctx) error {
	id := c.Params("id")
	DB.Exec(context.Background(), "UPDATE users SET usage = 0 WHERE id = $1", id)
	return c.JSON(fiber.Map{"success": true})
}

func AdminDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var email string
	DB.QueryRow(context.Background(), "DELETE FROM users WHERE id = $1 RETURNING email", id).Scan(&email)
	return c.JSON(fiber.Map{"success": true, "deleted": email})
}

func AdminResetAllUsage(c *fiber.Ctx) error {
	DB.Exec(context.Background(), "UPDATE users SET usage = 0")
	return c.JSON(fiber.Map{"success": true, "message": "All usage reset"})
}
