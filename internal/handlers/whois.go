package handlers

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func WhoisPage(c *fiber.Ctx) error {
	return RenderPage(c, "whois-lookup.html", PageData{
		Title:       "WHOIS Lookup - Domain Info | TraceQube",
		Description: "Look up domain registration and expiry info.",
		Canonical:   "/tools/whois-lookup",
	})
}

func WhoisLookup(c *fiber.Ctx) error {
	domain := c.Query("domain")
	if domain == "" {
		return c.JSON(fiber.Map{"error": "domain is required"})
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	conn, err := net.DialTimeout("tcp", "whois.iana.org:43", 10*time.Second)
	if err != nil {
		return c.JSON(fiber.Map{"error": "Could not connect to WHOIS server"})
	}
	defer conn.Close()
	fmt.Fprintf(conn, "%s\r\n", domain)
	body, _ := io.ReadAll(conn)
	result := string(body)
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "refer:") {
			server := strings.TrimSpace(strings.TrimPrefix(trimmed, "refer:"))
			conn2, err2 := net.DialTimeout("tcp", server+":43", 10*time.Second)
			if err2 == nil {
				defer conn2.Close()
				fmt.Fprintf(conn2, "%s\r\n", domain)
				body2, _ := io.ReadAll(conn2)
				result = string(body2)
			}
			break
		}
	}
	return c.JSON(fiber.Map{"domain": domain, "result": result})
}
