package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ToolIP(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	ip := c.Query("ip")
	if ip == "" { ip = c.Get("CF-Connecting-IP") }
	if ip == "" { ip = c.IP() }
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil { return c.JSON(fiber.Map{"ip": ip}) }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	result["ip"] = ip
	IncrementUsage(u.ID)
	return c.JSON(result)
}

func ToolDNS(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	domain := c.Query("domain")
	recordType := strings.ToUpper(c.Query("type", "A"))
	if domain == "" { return c.Status(400).JSON(fiber.Map{"error": "domain is required"}) }
	var results []string
	var err error
	switch recordType {
	case "A":
		ips, e := net.LookupHost(domain); err = e; results = ips
	case "MX":
		mxs, e := net.LookupMX(domain); err = e
		for _, mx := range mxs { results = append(results, fmt.Sprintf("%s (priority: %d)", mx.Host, mx.Pref)) }
	case "NS":
		nss, e := net.LookupNS(domain); err = e
		for _, ns := range nss { results = append(results, ns.Host) }
	case "TXT":
		txts, e := net.LookupTXT(domain); err = e; results = txts
	case "CNAME":
		cname, e := net.LookupCNAME(domain); err = e; results = []string{cname}
	default:
		ips, e := net.LookupHost(domain); err = e; results = ips
	}
	if err != nil { return c.JSON(fiber.Map{"error": err.Error(), "domain": domain, "type": recordType}) }
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"domain": domain, "type": recordType, "results": results})
}

func ToolPing(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	host := c.Query("host")
	if host == "" { return c.Status(400).JSON(fiber.Map{"error": "host is required"}) }
	var results []map[string]interface{}
	for i := 0; i < 4; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", host), 3*time.Second)
		elapsed := time.Since(start).Milliseconds()
		if err != nil {
			conn2, err2 := net.DialTimeout("tcp", fmt.Sprintf("%s:443", host), 3*time.Second)
			elapsed = time.Since(start).Milliseconds()
			if err2 != nil { results = append(results, map[string]interface{}{"seq": i+1, "status": "timeout"}); continue }
			conn2.Close()
		} else { conn.Close() }
		results = append(results, map[string]interface{}{"seq": i+1, "status": "ok", "ms": elapsed})
	}
	var total int64; success := 0
	for _, r := range results { if r["status"] == "ok" { total += r["ms"].(int64); success++ } }
	var avg interface{}
	if success > 0 { avg = total / int64(success) }
	loss := ((4 - success) * 100) / 4
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"host": host, "results": results, "avg_ms": avg, "packet_loss": fmt.Sprintf("%d%%", loss)})
}

func ToolPort(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	host := c.Query("host"); port := c.Query("port")
	if host == "" || port == "" { return c.Status(400).JSON(fiber.Map{"error": "host and port are required"}) }
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 5*time.Second)
	if err != nil { return c.JSON(fiber.Map{"host": host, "port": port, "status": "closed", "message": "Port is closed or filtered"}) }
	conn.Close()
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"host": host, "port": port, "status": "open", "message": "Port is open"})
}

func ToolSSL(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	domain := c.Query("domain")
	if domain == "" { return c.Status(400).JSON(fiber.Map{"error": "domain is required"}) }
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10*time.Second}, "tcp", fmt.Sprintf("%s:443", domain), &tls.Config{ServerName: domain})
	if err != nil { return c.JSON(fiber.Map{"domain": domain, "valid": false, "error": err.Error()}) }
	defer conn.Close()
	cert := conn.ConnectionState().PeerCertificates[0]
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"domain": domain, "valid": true, "issuer": cert.Issuer.CommonName, "subject": cert.Subject.CommonName, "expires": cert.NotAfter.Format("2006-01-02"), "issued": cert.NotBefore.Format("2006-01-02"), "days_left": daysLeft, "dns_names": cert.DNSNames})
}

func ToolWhois(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	domain := strings.ToLower(strings.TrimSpace(c.Query("domain")))
	if domain == "" { return c.Status(400).JSON(fiber.Map{"error": "domain is required"}) }
	conn, err := net.DialTimeout("tcp", "whois.iana.org:43", 10*time.Second)
	if err != nil { return c.JSON(fiber.Map{"error": "Could not connect to WHOIS server"}) }
	defer conn.Close()
	fmt.Fprintf(conn, "%s\r\n", domain)
	body, _ := io.ReadAll(conn)
	result := string(body)
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "refer:") {
			server := strings.TrimSpace(strings.TrimPrefix(trimmed, "refer:"))
			if conn2, err2 := net.DialTimeout("tcp", server+":43", 10*time.Second); err2 == nil {
				defer conn2.Close()
				fmt.Fprintf(conn2, "%s\r\n", domain)
				body2, _ := io.ReadAll(conn2)
				result = string(body2)
			}
			break
		}
	}
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"domain": domain, "result": result})
}

func ToolGeo(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	ip := c.Query("ip")
	if ip == "" { return c.Status(400).JSON(fiber.Map{"error": "ip is required"}) }
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil { return c.JSON(fiber.Map{"error": "Could not fetch location"}) }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	IncrementUsage(u.ID)
	return c.JSON(result)
}

func ToolRDNS(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	ip := c.Query("ip")
	if ip == "" { return c.Status(400).JSON(fiber.Map{"error": "ip is required"}) }
	hostnames, err := net.LookupAddr(ip)
	if err != nil { return c.JSON(fiber.Map{"ip": ip, "hostnames": []string{}, "error": "No PTR record found"}) }
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"ip": ip, "hostnames": hostnames})
}

func ToolSubnet(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	cidr := c.Query("cidr")
	if cidr == "" { return c.Status(400).JSON(fiber.Map{"error": "cidr is required"}) }
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil { return c.JSON(fiber.Map{"error": "Invalid CIDR notation"}) }
	mask := ipNet.Mask
	ones, bits := mask.Size()
	hosts := (1 << uint(bits-ones)) - 2
	if hosts < 0 { hosts = 0 }
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"cidr": cidr, "network": ipNet.IP.String(), "prefix_length": ones, "total_hosts": hosts, "subnet_mask": fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])})
}

func ToolHeaders(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	url := c.Query("url")
	if url == "" { return c.Status(400).JSON(fiber.Map{"error": "url is required"}) }
	if !strings.HasPrefix(url, "http") { url = "https://" + url }
	client := &http.Client{Timeout: 10*time.Second}
	resp, err := client.Head(url)
	if err != nil { return c.JSON(fiber.Map{"error": fmt.Sprintf("Could not reach %s", url)}) }
	defer resp.Body.Close()
	headers := make(map[string]string)
	for k, v := range resp.Header { headers[k] = v[0] }
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"url": url, "status": resp.Status, "status_code": resp.StatusCode, "headers": headers})
}

func ToolDown(c *fiber.Ctx) error {
	u := c.Locals("user").(User)
	if !CheckLimit(c) { return nil }
	url := c.Query("url")
	if url == "" { return c.Status(400).JSON(fiber.Map{"error": "url is required"}) }
	if !strings.HasPrefix(url, "http") { url = "https://" + url }
	start := time.Now()
	client := &http.Client{Timeout: 10*time.Second}
	resp, err := client.Get(url)
	elapsed := time.Since(start).Milliseconds()
	if err != nil { return c.JSON(fiber.Map{"url": url, "status": "down", "message": "Site is unreachable", "response_ms": elapsed}) }
	defer resp.Body.Close()
	status := "up"
	if resp.StatusCode >= 500 { status = "down" }
	IncrementUsage(u.ID)
	return c.JSON(fiber.Map{"url": url, "status": status, "status_code": resp.StatusCode, "response_ms": elapsed})
}

// Keep context import used
var _ = context.Background
