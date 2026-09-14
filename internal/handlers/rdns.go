package handlers

import (
	"net"

	"github.com/gofiber/fiber/v2"
)

func ReverseDNSPage(c *fiber.Ctx) error {
	return RenderPage(c, "reverse-dns.html", PageData{
		Title:       "Reverse DNS Lookup – Resolve IP to Hostname | TraceQube",
		Description: "Perform a reverse DNS lookup to resolve an IP address back to its hostname using PTR records.",
		Canonical:   "/tools/reverse-dns",
	})
}

func ReverseDNS(c *fiber.Ctx) error {
	ip := c.Query("ip")
	if ip == "" {
		return c.JSON(fiber.Map{"error": "ip is required"})
	}
	hostnames, err := net.LookupAddr(ip)
	if err != nil {
		return c.JSON(fiber.Map{"ip": ip, "error": "No PTR record found", "hostnames": []string{}})
	}
	return c.JSON(fiber.Map{"ip": ip, "hostnames": hostnames})
}
