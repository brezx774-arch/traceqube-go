package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func generateToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateAPIKey() string {
	return "tq_" + generateToken(24)
}

func generateJWT(id string, email string, tokenVersion int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    id,
		"tv":    tokenVersion,
		"email": email,
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func sendMail(to, subject, body string) {
	log.Printf("sendMail called: to=%s", to)
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if user == "" {
		log.Println("SMTP: no user configured")
		return
	}
	auth := smtp.PlainAuth("", user, pass, host)
	msg := fmt.Sprintf("From: TraceQube <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", user, to, subject, body)
	err := smtp.SendMail(host+":"+port, auth, user, []string{to}, []byte(msg))
	if err != nil {
		log.Println("SMTP error:", err)
	} else {
		log.Println("SMTP: email sent to", to)
	}
}

func Register(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	if body.Email == "" || body.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password required"})
	}
	if len(body.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters"})
	}
	var existing int
	DB.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", body.Email).Scan(&existing)
	if existing > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Email already exists"})
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	apiKey := generateAPIKey()
	verifyToken := generateToken(32)
	var id string
	err := DB.QueryRow(context.Background(),
		`INSERT INTO users (email, password, api_key, plan, usage, verified, verify_token)
		 VALUES ($1, $2, $3, 'free', 0, false, $4) RETURNING id`,
		body.Email, string(hashed), apiKey, verifyToken).Scan(&id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	go sendMail(body.Email, "Verify your TraceQube account",
		fmt.Sprintf(`<div style="font-family:Arial,sans-serif;max-width:500px;margin:0 auto">
		<h2 style="color:#c47d2e">Welcome to TraceQube!</h2>
		<p>Click the link below to verify your email address:</p>
		<a href="https://api.traceqube.com/api/verify-email?token=%s"
		  style="display:inline-block;padding:12px 24px;background:#c47d2e;color:white;border-radius:8px;font-weight:bold;text-decoration:none">
		  Verify Email
		</a></div>`, verifyToken))
	token, _ := generateJWT(id, body.Email, 0)
	return c.JSON(fiber.Map{"token": token, "email": body.Email, "apiKey": apiKey, "plan": "free", "usage": 0, "verified": false})
}

func VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	var id string
	err := DB.QueryRow(context.Background(), "SELECT id FROM users WHERE verify_token = $1", token).Scan(&id)
	if err != nil {
		return c.Status(400).SendString("<h2>Invalid or expired link.</h2>")
	}
	DB.Exec(context.Background(), "UPDATE users SET verified = true, verify_token = NULL WHERE verify_token = $1", token)
	c.Set("Content-Type", "text/html")
	return c.SendString(`<html><body style="font-family:Arial;text-align:center;padding:60px;background:#faf9f7">
		<h2 style="color:#c47d2e">Email Verified!</h2>
		<p>Your TraceQube account is now active.</p>
		<a href="https://app.traceqube.com" style="color:#c47d2e">Go to Dashboard</a>
		</body></html>`)
}

func Login(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	var u User
	var hashedPw string
	err := DB.QueryRow(context.Background(),
		"SELECT id, email, password, api_key, plan, usage, verified, is_admin, token_version FROM users WHERE email = $1",
		body.Email).Scan(&u.ID, &u.Email, &hashedPw, &u.APIKey, &u.Plan, &u.Usage, &u.Verified, &u.IsAdmin, &u.TokenVersion)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPw), []byte(body.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	token, _ := generateJWT(u.ID, u.Email, u.TokenVersion)
	setAuthCookie(c, token)
	return c.JSON(fiber.Map{"token": token, "email": u.Email, "apiKey": u.APIKey, "plan": u.Plan, "usage": u.Usage, "verified": u.Verified, "isAdmin": u.IsAdmin})
}

func ForgotPassword(c *fiber.Ctx) error {
	var body struct{ Email string `json:"email"` }
	c.BodyParser(&body)
	var id string
	if err := DB.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", body.Email).Scan(&id); err != nil {
		return c.JSON(fiber.Map{"message": "If that email exists, a reset link has been sent."})
	}
	resetToken := generateToken(32)
	expiry := time.Now().Add(time.Hour)
	DB.Exec(context.Background(), "UPDATE users SET reset_token = $1, reset_expiry = $2 WHERE email = $3", resetToken, expiry, body.Email)
	go sendMail(body.Email, "Reset your TraceQube password",
		fmt.Sprintf(`<div style="font-family:Arial,sans-serif;max-width:500px;margin:0 auto">
		<h2 style="color:#c47d2e">Password Reset</h2>
		<p>Click the link below to reset your password. Expires in 1 hour.</p>
		<a href="https://app.traceqube.com/reset-password?token=%s"
		  style="display:inline-block;padding:12px 24px;background:#c47d2e;color:white;border-radius:8px;font-weight:bold;text-decoration:none">
		  Reset Password
		</a></div>`, resetToken))
	return c.JSON(fiber.Map{"message": "If that email exists, a reset link has been sent."})
}

func ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	c.BodyParser(&body)
	if len(body.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters"})
	}
	var id int
	var expiry time.Time
	err := DB.QueryRow(context.Background(),
		"SELECT id, reset_expiry FROM users WHERE reset_token = $1", body.Token).Scan(&id, &expiry)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired token"})
	}
	if time.Now().After(expiry) {
		return c.Status(400).JSON(fiber.Map{"error": "Token has expired"})
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	DB.Exec(context.Background(), "UPDATE users SET password = $1, reset_token = NULL, reset_expiry = NULL WHERE id = $2", string(hashed), id)
	return c.JSON(fiber.Map{"message": "Password reset successfully"})
}
