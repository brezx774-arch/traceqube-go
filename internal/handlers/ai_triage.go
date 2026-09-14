package handlers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------
// Incident Triage — classify + diagnose raw/messy diagnostic dumps
// (curl -v, nginx errors, DNS errors, mixed traceroute+ping pastes,
// or a screenshot of any of the above — text is OCR'd out of images
// and run through the same proven text-diagnosis path).
// ---------------------------------------------------------------------

const maxTriageImageBytes = 5 * 1024 * 1024 // 5MB

var allowedTriageImageTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpeg",
	"image/webp": "webp",
}

func buildTriagePrompt(raw string) string {
	return fmt.Sprintf(`You are a senior network engineer helping triage a problem from raw diagnostic output pasted by a user. The input could be a traceroute, ping, curl -v output, DNS error, nginx/server log snippet, or something else — you must figure out which.
Do not repeat back any long tokens that look like API keys, auth headers, passwords, or session cookies from the input — refer to them generically (e.g. "an auth header was present") instead of quoting them.
Raw input:
%s
Respond in exactly this format, nothing else:
Type: <what kind of output this is>
Failure point: <the specific line/hop/detail indicating a problem, or "none apparent">
Likely cause: <one or two sentence plain-English diagnosis>
Suggested next step: <one concrete action, and if relevant, name it as one of: DNS Lookup, Ping Test, Port Checker, SSL Checker, WHOIS Lookup, Reverse DNS, HTTP Header Checker, Speed Test>`, raw)
}

type aiTriageRequest struct {
	Output string `json:"output"`
}

// AITriage handles POST /api/ai/triage
// Text mode:  JSON body { "output": "<raw pasted diagnostic dump>" }
// Image mode: multipart/form-data with a "screenshot" file field —
//             text is OCR'd out of the image, then handled identically.
func AITriage(c *fiber.Ctx) error {
	if strings.HasPrefix(c.Get("Content-Type"), "multipart/form-data") {
		return handleTriageImage(c)
	}

	var req aiTriageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Output = strings.TrimSpace(req.Output)
	if req.Output == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "output is required"})
	}
	analyzeAndRespond(c, req.Output)
	return nil
}

func handleTriageImage(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("screenshot")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "screenshot file is required"})
	}
	if fileHeader.Size > maxTriageImageBytes {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "image too large (max 5MB)"})
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	ext, ok := allowedTriageImageTypes[mimeType]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported image type — use PNG, JPEG, or WebP"})
	}

	tmpFile, err := os.CreateTemp("", "triage-*."+ext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not process uploaded file"})
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := c.SaveFile(fileHeader, tmpPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not save uploaded file"})
	}

	extractedText, err := extractTextFromImage(tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read text from image"})
	}
	extractedText = strings.TrimSpace(extractedText)
	if len(extractedText) < 15 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "couldn't find readable diagnostic text in that screenshot — try a clearer image or paste the text directly"})
	}

	analyzeAndRespond(c, extractedText)
	return nil
}

// extractTextFromImage inverts/contrasts the image (terminal screenshots are
// typically light-on-dark, which OCR reads far worse than dark-on-light) then
// runs Tesseract against the processed version.
func extractTextFromImage(path string) (string, error) {
	processedPath := path + "-processed.png"
	defer os.Remove(processedPath)

	convertCmd := exec.Command("convert", path, "-colorspace", "Gray", "-negate", "-contrast-stretch", "0", processedPath)
	var convertErr bytes.Buffer
	convertCmd.Stderr = &convertErr
	if err := convertCmd.Run(); err != nil {
		return "", fmt.Errorf("image preprocessing failed: %v: %s", err, convertErr.String())
	}

	cmd := exec.Command("tesseract", processedPath, "stdout", "-l", "eng")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract failed: %v: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// analyzeAndRespond runs the shared rate-limit + prompt + AI-call + JSON
// response path used by both the text and OCR'd-image inputs.
func analyzeAndRespond(c *fiber.Ctx, raw string) {
	const maxRawChars = 6000
	if len(raw) > maxRawChars {
		raw = raw[:maxRawChars] + "\n... (truncated)"
	}
	ip := aiRealIP(c)
	isPaid := isPaidUser(c)
	allowed, remaining := checkAndIncrementUsage(ip, isPaid)
	if !allowed {
		c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":   "daily AI analysis limit reached",
			"limit":   freeDailyLimit,
			"upgrade": "/pricing",
		})
		return
	}
	prompt := buildTriagePrompt(raw)
	analysis, err := callNvidiaAI(prompt, 300)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI analysis failed, please try again"})
		return
	}
	c.JSON(fiber.Map{
		"analysis":  analysis,
		"remaining": remaining,
	})
}

func TriagePage(c *fiber.Ctx) error {
	return RenderPage(c, "triage.html", PageData{
		Title:       "Incident Triage – AI Network Diagnostics | TraceQube",
		Description: "Paste any raw diagnostic output — traceroute, ping, curl, DNS errors — or upload a screenshot, and get an instant AI-powered diagnosis.",
		Canonical:   "/tools/triage",
	})
}
