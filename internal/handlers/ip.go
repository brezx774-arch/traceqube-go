package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type IPInfo struct {
	IP       string `json:"ip"`
	Country  string `json:"country"`
	Region   string `json:"regionName"`
	City     string `json:"city"`
	Zip      string `json:"zip"`
	Timezone string `json:"timezone"`
	ISP      string `json:"isp"`
	ASN      string `json:"as"`
	Query    string `json:"query"`
}

func GetMyIP(c *fiber.Ctx) error {
	ip := c.Get("X-Real-IP")
	if ip == "" {
		ip = c.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.IP()
	}

	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return c.JSON(fiber.Map{"ip": ip, "error": "Could not fetch location"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info IPInfo
	json.Unmarshal(body, &info)
	info.IP = ip

	return c.JSON(info)
}

func WhatIsMyIPPage(c *fiber.Ctx) error {
	return RenderPage(c, "what-is-my-ip.html", PageData{
		Title:       "What Is My IP Address? – Find Your Public IP | TraceQube",
		Description: "Instantly find your public IP address, location, ISP, and timezone. Free IP lookup tool by TraceQube.",
		Canonical:   "/tools/what-is-my-ip",
	})
}
