package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------
// NVIDIA client (same pattern as cmd/bloggenerator/main.go)
// ---------------------------------------------------------------------

const aiNvidiaAPI = "https://integrate.api.nvidia.com/v1/chat/completions"
const aiModel = "meta/llama-3.2-11b-vision-instruct"

func callNvidiaAI(prompt string, maxTokens int) (string, error) {
	apiKey := os.Getenv("NVIDIA_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("NVIDIA_API_KEY not set")
	}
	body := map[string]interface{}{
		"model":       aiModel,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens":  maxTokens,
		"temperature": 0.3, // lower than blog gen — we want precise, factual analysis
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", aiNvidiaAPI, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse error: %s", string(respBody))
	}
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices: %s", string(respBody))
	}
	msg := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	return strings.TrimSpace(msg["content"].(string)), nil
}

// ---------------------------------------------------------------------
// Rate limiting — simple in-memory daily counter keyed by IP.
// Swap for a DB-backed version when accounts/plans are wired up.
// ---------------------------------------------------------------------

const freeDailyLimit = 3

type aiUsage struct {
	count int
	day   string // "2006-01-02"
}

var (
	// FLAG: in-memory + per-process, keyed by aiRealIP (CF-Connecting-IP).
	// Resets on restart/redeploy and is NOT shared across multiple instances.
	// Shared/CGNAT IPs share one quota; VPN or IP rotation effectively resets the limit.
	// Fine for a single instance / soft limit -- revisit before scaling out or if abused.
	aiUsageMu  sync.Mutex
	aiUsageMap = map[string]*aiUsage{}
)

func aiRealIP(c *fiber.Ctx) string {
	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return c.IP()
}

// checkAndIncrementUsage returns (allowed, remaining)
// isPaid bypasses the limit entirely.
func checkAndIncrementUsage(ip string, isPaid bool) (bool, int) {
	if isPaid {
		return true, -1 // unlimited
	}
	today := time.Now().UTC().Format("2006-01-02")

	aiUsageMu.Lock()
	defer aiUsageMu.Unlock()

	u, ok := aiUsageMap[ip]
	if !ok || u.day != today {
		u = &aiUsage{count: 0, day: today}
		aiUsageMap[ip] = u
	}
	if u.count >= freeDailyLimit {
		return false, 0
	}
	u.count++
	return true, freeDailyLimit - u.count
}

// ---------------------------------------------------------------------
// Prompt templates per tool
// ---------------------------------------------------------------------

// toolType -> prompt builder. Each builder gets the raw output (as a
// string) plus an optional "target" (domain/IP/host the scan was run
// against) for context.
var aiPromptBuilders = map[string]func(target, raw string) string{

	"mtr": func(target, raw string) string {
		return fmt.Sprintf(`You are a senior network engineer. Analyze this MTR/traceroute output for target "%s".

%s

In plain English, explain:
1. Where (which hop/ISP) packet loss or high latency occurs, if any
2. Whether the issue is likely on the user's local network, their ISP, an upstream/transit provider, or the destination server
3. One or two concrete next steps the user could take

Keep it under 120 words. Be direct and avoid restating the raw data line by line.`, target, raw)
	},

	"traceroute": func(target, raw string) string {
		return fmt.Sprintf(`You are a senior network engineer. Analyze this traceroute output for target "%s".

%s

In plain English, explain:
1. Any hops with notably high latency jumps or timeouts (*)
2. Whether the routing path looks normal or has a likely bottleneck, and roughly where (local network, ISP, transit, destination)
3. One or two concrete next steps

Keep it under 120 words.`, target, raw)
	},

	"dns": func(target, raw string) string {
		return fmt.Sprintf(`You are a DNS specialist. Analyze this DNS lookup result for domain "%s".

%s

In plain English, explain:
1. Whether the records look correctly configured for normal operation
2. Anything missing or unusual (e.g. no MX for a domain that sends mail, no SPF/DKIM in TXT, multiple conflicting A records, CNAME chains)
3. One or two concrete fixes if something looks off — otherwise confirm it looks healthy

Keep it under 100 words.`, target, raw)
	},

	"reverse-dns": func(target, raw string) string {
		return fmt.Sprintf(`You are a network engineer. Analyze this reverse DNS (PTR) lookup result for IP "%s".

%s

In plain English, explain:
1. Whether a PTR record exists and if it matches expected naming conventions
2. Why missing/mismatched PTR records matter (e.g. email deliverability, server identification)
3. One or two concrete next steps if something looks off

Keep it under 90 words.`, target, raw)
	},

	"ping": func(target, raw string) string {
		return fmt.Sprintf(`You are a network engineer. Analyze this ping test result for "%s".

%s

In plain English, explain:
1. Whether latency and packet loss look normal, elevated, or concerning
2. What jitter (variance between ping times) suggests about the connection quality
3. One or two concrete next steps if results look poor — otherwise confirm it looks healthy

Keep it under 90 words.`, target, raw)
	},

	"port-checker": func(target, raw string) string {
		return fmt.Sprintf(`You are a security-conscious network engineer. Analyze this port scan result for "%s".

%s

In plain English, explain:
1. Which ports are open/closed/filtered and what services typically run on them
2. Any security concerns from open ports (e.g. exposed databases, admin panels, unnecessary services)
3. One or two concrete recommendations

Keep it under 110 words.`, target, raw)
	},

	"http-headers": func(target, raw string) string {
		return fmt.Sprintf(`You are a web security and performance specialist. Analyze these HTTP response headers for "%s".

%s

In plain English, explain:
1. Server/technology info these headers reveal
2. Missing or weak security headers (e.g. HSTS, CSP, X-Frame-Options, X-Content-Type-Options) and why they matter
3. Any caching/performance-related headers worth noting
4. One or two concrete recommendations

Keep it under 120 words.`, target, raw)
	},

	"whois": func(target, raw string) string {
		return fmt.Sprintf(`You are a domain and network specialist. Analyze this WHOIS result for "%s".

%s

In plain English, summarize:
1. Registrar, registration/expiry dates, and whether expiry is coming up soon (flag if within ~60 days)
2. Nameservers and what they suggest about hosting/DNS setup
3. Anything notable (privacy protection, recent transfer, lock status)

Keep it under 100 words.`, target, raw)
	},

	"ssl": func(target, raw string) string {
		return fmt.Sprintf(`You are a security engineer. Analyze this SSL/TLS certificate check for "%s".

%s

In plain English, explain:
1. Certificate validity, issuer, and expiry status (flag clearly if expiring within ~30 days or expired)
2. Any chain, trust, or protocol issues (e.g. "chain issues", weak protocol versions, self-signed)
3. If there's a problem, give concrete steps to fix it on a typical Nginx/Apache + Let's Encrypt setup

Keep it under 120 words.`, target, raw)
	},

	"geo": func(target, raw string) string {
		return fmt.Sprintf(`You are a network engineer. Analyze this IP geolocation result for "%s".

%s

In plain English, briefly summarize:
1. Location, ISP/organization, and connection type
2. Whether anything looks unusual (e.g. hosting provider IP vs residential, mismatched location vs expected)

Keep it under 70 words.`, target, raw)
	},

	"subnet": func(target, raw string) string {
		return fmt.Sprintf(`You are a network engineer. Explain this subnet calculation result for "%s".

%s

In plain English, explain:
1. Usable host range and broadcast/network addresses
2. How many usable hosts this subnet provides and what size network that's suitable for (e.g. small office, single rack, etc.)

Keep it under 80 words.`, target, raw)
	},

	"down-checker": func(target, raw string) string {
		return fmt.Sprintf(`You are a network engineer. Analyze this uptime/down-checker result for "%s".

%s

In plain English, explain:
1. Whether the site/host appears up or down, and from how many vantage points if given
2. Likely causes if it's down (server issue, DNS issue, firewall, regional outage)
3. One or two next steps

Keep it under 90 words.`, target, raw)
	},

	"speed-test": func(target, raw string) string {
		return fmt.Sprintf(`You are a network engineer. Analyze this speed test result for "%s".

%s

In plain English, explain:
1. Whether download/upload speeds and latency look good, average, or poor for typical use cases (browsing, streaming, video calls, gaming)
2. What might cause poor results if applicable (ISP throttling, local network congestion, server distance)

Keep it under 80 words.`, target, raw)
	},

	"ip": func(target, raw string) string {
		return fmt.Sprintf(`You are a network engineer. Briefly explain this "what is my IP" result.

%s

In 2-3 sentences, explain what this IP reveals (ISP, rough location, IPv4 vs IPv6) and whether it looks like a residential, mobile, or hosting/datacenter connection.`, raw)
	},
}

// ---------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------

type aiAnalyzeRequest struct {
	Tool   string `json:"tool"`
	Target string `json:"target"`
	Output string `json:"output"`
}

// AIAnalyze handles POST /api/ai/analyze
// Body: { "tool": "mtr", "target": "google.com", "output": "<raw tool output>" }
func AIAnalyze(c *fiber.Ctx) error {
	var req aiAnalyzeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Tool = strings.ToLower(strings.TrimSpace(req.Tool))
	req.Output = strings.TrimSpace(req.Output)

	if req.Tool == "" || req.Output == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tool and output are required"})
	}

	builder, ok := aiPromptBuilders[req.Tool]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("unknown tool type: %s", req.Tool)})
	}

	// Cap raw output size sent to the model — keeps cost/latency down
	// and avoids blowing context on pathological inputs.
	const maxRawChars = 6000
	raw := req.Output
	if len(raw) > maxRawChars {
		raw = raw[:maxRawChars] + "\n... (truncated)"
	}

	// --- rate limiting ---
	ip := aiRealIP(c)
	isPaid := isPaidUser(c) // FLAG: always false today, so ALL users (incl. future paid accounts) hit freeDailyLimit. TODO: wire to real auth/plan lookup once cmd/api auth is integrated here.
	allowed, remaining := checkAndIncrementUsage(ip, isPaid)
	if !allowed {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":   "daily AI analysis limit reached",
			"limit":   freeDailyLimit,
			"upgrade": "/pricing",
		})
	}

	prompt := builder(req.Target, raw)

	analysis, err := callNvidiaAI(prompt, 350)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI analysis failed, please try again"})
	}

	return c.JSON(fiber.Map{
		"analysis":  analysis,
		"remaining": remaining, // -1 means unlimited (paid)
	})
}

// isPaidUser is a placeholder. Replace with real session/plan lookup
// once auth + subscriptions are wired into this server (currently
// auth lives in cmd/api + cmd/dashboard).
func isPaidUser(c *fiber.Ctx) bool {
	return false
}

// callNvidiaAIVision sends a text prompt plus an image (as a base64 data URI)
// to the vision-capable model. Used for screenshot-based triage.
func callNvidiaAIVision(prompt string, imageDataURI string, maxTokens int) (string, error) {
	apiKey := os.Getenv("NVIDIA_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("NVIDIA_API_KEY not set")
	}
	content := []map[string]interface{}{
		{"type": "text", "text": prompt},
		{"type": "image_url", "image_url": map[string]string{"url": imageDataURI}},
	}
	body := map[string]interface{}{
		"model":       aiModel,
		"messages":    []map[string]interface{}{{"role": "user", "content": content}},
		"max_tokens":  maxTokens,
		"temperature": 0.3,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", aiNvidiaAPI, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse error: %s", string(respBody))
	}
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices: %s", string(respBody))
	}
	msg := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	return strings.TrimSpace(msg["content"].(string)), nil
}
