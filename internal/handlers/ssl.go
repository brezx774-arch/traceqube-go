package handlers

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SSLCheckerPage(c *fiber.Ctx) error {
	return RenderPage(c, "ssl-checker.html", PageData{
		Title:       "SSL Certificate Checker – Verify SSL/TLS Certificates | TraceQube",
		Description: "Check SSL certificate validity, expiry date, issuer, and domain coverage for any website.",
		Canonical:   "/tools/ssl-checker",
	})
}

func SSLCheck(c *fiber.Ctx) error {
	domain := c.Query("domain")
	if domain == "" {
		return c.JSON(fiber.Map{"error": "domain is required"})
	}
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		fmt.Sprintf("%s:443", domain),
		&tls.Config{ServerName: domain},
	)
	if err != nil {
		return c.JSON(fiber.Map{"domain": domain, "valid": false, "error": err.Error()})
	}
	defer conn.Close()
	cert := conn.ConnectionState().PeerCertificates[0]
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	return c.JSON(fiber.Map{
		"domain":     domain,
		"valid":      true,
		"issuer":     cert.Issuer.CommonName,
		"subject":    cert.Subject.CommonName,
		"expires":    cert.NotAfter.Format("2006-01-02"),
		"issued":     cert.NotBefore.Format("2006-01-02"),
		"days_left":  daysLeft,
		"dns_names":  cert.DNSNames,
	})
}
