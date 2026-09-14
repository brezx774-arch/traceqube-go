package handlers

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func DownCheckerPage(c *fiber.Ctx) error {
	return RenderPage(c, "down-checker.html", PageData{
		Title:       "Website Down Checker – Is It Down For Everyone? | TraceQube",
		Description: "Check if a website is down for everyone or just you. Get instant HTTP status and response time.",
		Canonical:   "/tools/down-checker",
	})
}

func DownCheck(c *fiber.Ctx) error {
	url := c.Query("url")
	if url == "" {
		return c.JSON(fiber.Map{"error": "url is required"})
	}
	if len(url) < 8 || (url[:7] != "http://" && url[:8] != "https://") {
		url = "https://" + url
	}
	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return c.JSON(fiber.Map{"url": url, "status": "down", "message": "Site is unreachable", "response_ms": elapsed})
	}
	defer resp.Body.Close()
	status := "up"
	if resp.StatusCode >= 500 {
		status = "down"
	}
	return c.JSON(fiber.Map{"url": url, "status": status, "status_code": resp.StatusCode, "response_ms": elapsed})
}

func SpeedTestPage(c *fiber.Ctx) error {
	return RenderPage(c, "speed-test.html", PageData{
		Title:       "Internet Speed Test – Test Your Download Speed | TraceQube",
		Description: "Test your internet download speed for free. Get accurate Mbps results instantly.",
		Canonical:   "/tools/speed-test",
	})
}
