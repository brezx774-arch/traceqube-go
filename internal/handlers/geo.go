package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func IPGeoPage(c *fiber.Ctx) error {
	return RenderPage(c, "ip-geolocation.html", PageData{
		Title:       "IP Geolocation – Look Up Any IP Address Location | TraceQube",
		Description: "Look up the country, city, ISP, ASN, and timezone of any IP address with our free geolocation tool.",
		Canonical:   "/tools/ip-geolocation",
	})
}

func IPGeoLookup(c *fiber.Ctx) error {
	ip := c.Query("ip")
	if ip == "" {
		return c.JSON(fiber.Map{"error": "ip is required"})
	}
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return c.JSON(fiber.Map{"error": "Could not fetch location data"})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return c.JSON(result)
}
