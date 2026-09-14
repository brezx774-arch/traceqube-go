package handlers

import (
	"bytes"
	"html/template"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type PageData struct {
	Title       string
	Description string
	Canonical   string
	LoggedIn    bool
	Data        interface{}
}

func isLoggedIn(c *fiber.Ctx) bool {
	tokenStr := c.Cookies("tq_token")
	if tokenStr == "" {
		return false
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	return err == nil && token.Valid
}

func RenderPage(c *fiber.Ctx, contentTmpl string, data PageData) error {
	data.LoggedIn = isLoggedIn(c)
	t, err := template.ParseFiles("web/templates/base.html", "web/templates/"+contentTmpl)
	if err != nil {
		return c.Status(500).SendString("Template error: " + err.Error())
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return c.Status(500).SendString("Render error: " + err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(buf.Bytes())
}
