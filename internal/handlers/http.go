package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

func HTTPHeaderPage(c *fiber.Ctx) error {
	return RenderPage(c, "http-header-checker.html", PageData{
		Title:       "HTTP Header Checker – Inspect Response Headers | TraceQube",
		Description: "Inspect HTTP response headers for any URL. Check security headers, cache control, content type, and more.",
		Canonical:   "/tools/http-header-checker",
	})
}

func HTTPHeaderCheck(c *fiber.Ctx) error {
	url := c.Query("url")
	if url == "" {
		return c.JSON(fiber.Map{"error": "url is required"})
	}
	if len(url) < 8 || (url[:7] != "http://" && url[:8] != "https://") {
		url = "https://" + url
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		resp2, err2 := client.Get(url)
		if err2 != nil {
			return c.JSON(fiber.Map{"error": fmt.Sprintf("Could not reach %s", url)})
		}
		defer resp2.Body.Close()
		resp = resp2
	}
	defer resp.Body.Close()
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = values[0]
	}
	return c.JSON(fiber.Map{"url": url, "status": resp.Status, "status_code": resp.StatusCode, "headers": headers})
}
