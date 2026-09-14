package handlers

import (
	"fmt"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func DNSLookupPage(c *fiber.Ctx) error {
	return RenderPage(c, "dns-lookup.html", PageData{
		Title:       "DNS Lookup – Query DNS Records for Any Domain | TraceQube",
		Description: "Look up A, AAAA, MX, CNAME, TXT, and NS DNS records for any domain name. Free DNS lookup tool.",
		Canonical:   "/tools/dns-lookup",
	})
}

func DNSLookup(c *fiber.Ctx) error {
	domain := c.Query("domain")
	recordType := strings.ToUpper(c.Query("type", "A"))
	if domain == "" {
		return c.JSON(fiber.Map{"error": "domain is required"})
	}
	var results []string
	var err error
	switch recordType {
	case "A":
		ips, e := net.LookupHost(domain)
		err = e
		results = ips
	case "MX":
		mxs, e := net.LookupMX(domain)
		err = e
		for _, mx := range mxs {
			results = append(results, fmt.Sprintf("%s (priority: %d)", mx.Host, mx.Pref))
		}
	case "NS":
		nss, e := net.LookupNS(domain)
		err = e
		for _, ns := range nss {
			results = append(results, ns.Host)
		}
	case "TXT":
		txts, e := net.LookupTXT(domain)
		err = e
		results = txts
	case "CNAME":
		cname, e := net.LookupCNAME(domain)
		err = e
		results = []string{cname}
	default:
		ips, e := net.LookupHost(domain)
		err = e
		results = ips
	}
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error(), "domain": domain, "type": recordType})
	}
	return c.JSON(fiber.Map{"domain": domain, "type": recordType, "results": results})
}
