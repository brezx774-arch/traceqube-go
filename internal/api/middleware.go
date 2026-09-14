package api

import (
	"context"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID           string
	Email        string
	APIKey       string
	Plan         string
	Usage        int
	Verified     bool
	IsAdmin      bool
	TokenVersion int
}

func AuthMiddleware(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	apiKey := c.Get("X-API-Key")

	var userID string
	var tokenVersion int
	checkVersion := false

	if apiKey != "" {
		row := DB.QueryRow(context.Background(),
			"SELECT id FROM users WHERE api_key = $1", apiKey)
		if err := row.Scan(&userID); err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid API key"})
		}
	} else if header != "" {
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
		}
		claims := token.Claims.(jwt.MapClaims)
		userID = claims["id"].(string)
		if tv, ok := claims["tv"].(float64); ok {
			tokenVersion = int(tv)
			checkVersion = true
		}
	} else if cookieToken := c.Cookies("tq_token"); cookieToken != "" {
		token, err := jwt.Parse(cookieToken, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
		}
		claims := token.Claims.(jwt.MapClaims)
		userID = claims["id"].(string)
		if tv, ok := claims["tv"].(float64); ok {
			tokenVersion = int(tv)
			checkVersion = true
		}
	} else {
		return c.Status(401).JSON(fiber.Map{"error": "Authentication required"})
	}

	var u User
	err := DB.QueryRow(context.Background(),
		"SELECT id, email, api_key, plan, usage, verified, is_admin, token_version FROM users WHERE id = $1", userID).
		Scan(&u.ID, &u.Email, &u.APIKey, &u.Plan, &u.Usage, &u.Verified, &u.IsAdmin, &u.TokenVersion)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
	}

	if checkVersion && tokenVersion != u.TokenVersion {
		return c.Status(401).JSON(fiber.Map{"error": "Session expired, please log in again"})
	}

	c.Locals("user", u)
	return c.Next()
}

func AdminMiddleware(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !u.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Admin only"})
	}
	return c.Next()
}

func PlanLimit(plan string) int {
	switch plan {
	case "pro":
		return 5000
	case "enterprise":
		return 50000
	default:
		return 50
	}
}

func CheckLimit(c *fiber.Ctx) bool {
	u := c.Locals("user").(User)
	limit := PlanLimit(u.Plan)
	if u.Usage >= limit {
		c.Status(429).JSON(fiber.Map{"error": "Plan limit reached. Upgrade at app.traceqube.com"})
		return false
	}
	return true
}

func IncrementUsage(userID string) {
	DB.Exec(context.Background(),
		"UPDATE users SET usage = usage + 1 WHERE id = $1", userID)
}
