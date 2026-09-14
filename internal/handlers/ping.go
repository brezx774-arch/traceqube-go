package handlers

import (
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
)

func PingTestPage(c *fiber.Ctx) error {
	return RenderPage(c, "ping-test.html", PageData{
		Title:       "Ping Test – Check Host Latency & Reachability | TraceQube",
		Description: "Test the latency and reachability of any host or IP address with a free online ping tool.",
		Canonical:   "/tools/ping-test",
	})
}

func PingTest(c *fiber.Ctx) error {
	host := c.Query("host")
	if host == "" {
		return c.JSON(fiber.Map{"error": "host is required"})
	}
	var results []map[string]interface{}
	for i := 0; i < 4; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", host), 3*time.Second)
		elapsed := time.Since(start).Milliseconds()
		if err != nil {
			conn2, err2 := net.DialTimeout("tcp", fmt.Sprintf("%s:443", host), 3*time.Second)
			elapsed = time.Since(start).Milliseconds()
			if err2 != nil {
				results = append(results, map[string]interface{}{"seq": i + 1, "status": "timeout", "ms": nil})
				continue
			}
			conn2.Close()
		} else {
			conn.Close()
		}
		results = append(results, map[string]interface{}{"seq": i + 1, "status": "ok", "ms": elapsed})
	}
	var total int64
	success := 0
	for _, r := range results {
		if r["status"] == "ok" {
			total += r["ms"].(int64)
			success++
		}
	}
	var avg interface{}
	if success > 0 {
		avg = total / int64(success)
	}
	loss := ((4 - success) * 100) / 4
	return c.JSON(fiber.Map{"host": host, "results": results, "avg_ms": avg, "packet_loss": fmt.Sprintf("%d%%", loss)})
}
