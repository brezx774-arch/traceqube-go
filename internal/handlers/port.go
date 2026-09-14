package handlers

import (
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
)

func PortCheckerPage(c *fiber.Ctx) error {
	return RenderPage(c, "port-checker.html", PageData{
		Title:       "Port Checker – Check if a TCP Port is Open | TraceQube",
		Description: "Check if a specific TCP port is open or closed on any host or IP address. Free online port checker.",
		Canonical:   "/tools/port-checker",
	})
}

func PortCheck(c *fiber.Ctx) error {
	host := c.Query("host")
	port := c.Query("port")
	if host == "" || port == "" {
		return c.JSON(fiber.Map{"error": "host and port are required"})
	}
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return c.JSON(fiber.Map{"host": host, "port": port, "status": "closed", "message": "Port is closed or filtered"})
	}
	conn.Close()
	return c.JSON(fiber.Map{"host": host, "port": port, "status": "open", "message": "Port is open"})
}
