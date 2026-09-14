package handlers

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/gofiber/fiber/v2"
)

func SubnetPage(c *fiber.Ctx) error {
	return RenderPage(c, "subnet-calculator.html", PageData{
		Title:       "Subnet Calculator – CIDR Network Calculator | TraceQube",
		Description: "Calculate network address, broadcast address, usable IP range, and host count from any CIDR notation.",
		Canonical:   "/tools/subnet-calculator",
	})
}

func SubnetCalc(c *fiber.Ctx) error {
	cidr := c.Query("cidr")
	if cidr == "" {
		return c.JSON(fiber.Map{"error": "cidr is required"})
	}
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return c.JSON(fiber.Map{"error": "Invalid CIDR notation"})
	}
	mask := ipNet.Mask
	ones, bits := mask.Size()
	networkIP := ipNet.IP
	broadcastIP := make(net.IP, 4)
	n := binary.BigEndian.Uint32(networkIP.To4())
	m := binary.BigEndian.Uint32(mask)
	b := n | ^m
	binary.BigEndian.PutUint32(broadcastIP, b)
	hosts := (1 << uint(bits-ones)) - 2
	if hosts < 0 {
		hosts = 0
	}
	firstHost := make(net.IP, 4)
	binary.BigEndian.PutUint32(firstHost, n+1)
	lastHost := make(net.IP, 4)
	binary.BigEndian.PutUint32(lastHost, b-1)
	return c.JSON(fiber.Map{
		"ip":            ip.String(),
		"cidr":          cidr,
		"network":       networkIP.String(),
		"broadcast":     broadcastIP.String(),
		"subnet_mask":   fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
		"prefix_length": ones,
		"total_hosts":   hosts,
		"first_host":    firstHost.String(),
		"last_host":     lastHost.String(),
	})
}
