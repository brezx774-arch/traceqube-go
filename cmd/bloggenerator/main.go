package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
        "github.com/yuin/goldmark"
        "github.com/yuin/goldmark/renderer/html"
)

const NVIDIA_API = "https://integrate.api.nvidia.com/v1/chat/completions"
const MODEL = "meta/llama-3.2-11b-vision-instruct"

type BlogPost struct {
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Date     string `json:"date"`
	ReadTime string `json:"read_time"`
	Excerpt  string `json:"excerpt"`
}

type Topic struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Slug     string `json:"slug"`
	Excerpt  string `json:"excerpt"`
}

func callNvidia(prompt string, maxTokens int) (string, error) {
	apiKey := os.Getenv("NVIDIA_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("NVIDIA_API_KEY not set")
	}
	body := map[string]interface{}{
		"model":       MODEL,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens":  maxTokens,
		"temperature": 0.8,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", NVIDIA_API, bytes.NewBuffer(jsonBody))
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

// repairJSON fixes the common AI mistake of missing opening quote on string values
// e.g. "excerpt":Some text  ->  "excerpt":"Some text"
func repairJSON(s string) string {
	fields := []string{"title", "category", "slug", "excerpt"}
	for _, f := range fields {
		// Pattern: "field": followed by non-quote non-special char
		bad := `"` + f + `":`
		idx := 0
		for {
			pos := strings.Index(s[idx:], bad)
			if pos == -1 {
				break
			}
			abs := idx + pos + len(bad)
			// Skip whitespace
			for abs < len(s) && (s[abs] == ' ' || s[abs] == '\t') {
				abs++
			}
			if abs < len(s) && s[abs] != '"' && s[abs] != '{' && s[abs] != '[' {
				// Find end of value
				end := abs
				for end < len(s) && s[end] != ',' && s[end] != '\n' && s[end] != '}' {
					end++
				}
				val := strings.TrimSpace(s[abs:end])
				// Remove any stray quotes
				val = strings.Trim(val, `"`)
				// Rebuild with proper quotes
				s = s[:abs] + `"` + val + `"` + s[end:]
			}
			idx = abs + 1
			if idx >= len(s) {
				break
			}
		}
	}
	return s
}

func generateTopics(count int) ([]Topic, error) {
	prompt := fmt.Sprintf(`Generate %d unique blog post topics for a network diagnostics website called TraceQube. Topics must be about networking, DNS, IP addresses, SSL/TLS, web security, server administration, internet protocols, or web performance. Return ONLY a JSON array, no other text: [{"title":"...","category":"...","slug":"...","excerpt":"..."}] Rules: slug must be lowercase letters and hyphens only. category must be one of: Networking, Security, Developer, Operations, Domains, Performance. excerpt must be 1 sentence.`, count)
	content, err := callNvidia(prompt, 1500)
	if err != nil {
		return nil, err
	}
	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON array: %s", content)
	}
	raw := repairJSON(content[start : end+1])
	var topics []Topic
	if err := json.Unmarshal([]byte(raw), &topics); err != nil {
		return nil, fmt.Errorf("parse error: %s\nrepaired: %s", content, raw)
	}
	return topics, nil
}

func generatePost(title, category string) (string, error) {
	prompt := fmt.Sprintf(`Write a detailed blog post about "%s" for a network diagnostics website called TraceQube. Category: %s. Requirements: 600-900 words. Start directly with content no title. Use h2 tags for headings, p tags for paragraphs, strong for key terms, ul/li for lists. Include practical examples. Professional tone. HTML only no markdown.`, title, category)
	return callNvidia(prompt, 2500)
}

func slugExists(slug string) bool {
	_, err := os.Stat(fmt.Sprintf("web/templates/blog-%s.html", slug))
	return err == nil
}

func createTemplate(slug, title, category, date, content string) error {
	tmpl := "{{define \"content\"}}" + `
<div class="container">
  <div class="blog-article">
    <a href="/blog" class="back-link">
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 2L4 6l4 4"/></svg>
      All posts
    </a>
    <div class="blog-tag" style="margin-top:20px">` + category + `</div>
    <h1>` + title + `</h1>
    <div class="meta"><span>` + date + `</span><span>5 min read</span></div>
    ` + content + `
    <div style="margin-top:48px;padding-top:32px;border-top:1px solid var(--border)">
      <a href="/tools"><button class="btn-lg primary">Explore All Tools</button></a>
    </div>
  </div>
</div>` + "\n{{end}}"
	return os.WriteFile(fmt.Sprintf("web/templates/blog-%s.html", slug), []byte(tmpl), 0644)
}

func readBlogJSON() ([]BlogPost, error) {
	data, err := os.ReadFile("data/blog.json")
	if err != nil {
		return nil, err
	}
	var posts []BlogPost
	return posts, json.Unmarshal(data, &posts)
}

func rebuildBlogTemplate(posts []BlogPost) error {
	// Newest first
	sorted := make([]BlogPost, len(posts))
	copy(sorted, posts)
	for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}
	const perPage = 10
	totalPages := (len(sorted) + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}
	for page := 1; page <= totalPages; page++ {
		start := (page - 1) * perPage
		end := start + perPage
		if end > len(sorted) {
			end = len(sorted)
		}
		pagePosts := sorted[start:end]
		cards := ""
		for _, p := range pagePosts {
			cards += fmt.Sprintf(`
    <a href="/blog/%s" class="blog-card">
      <div class="blog-tag">%s</div>
      <h3>%s</h3>
      <p>%s</p>
      <div class="blog-meta"><span>%s</span><span>%s</span></div>
    </a>`, p.Slug, p.Category, p.Title, p.Excerpt, p.Date, p.ReadTime)
		}
		pagination := `<div class="blog-pagination" style="display:flex;justify-content:space-between;align-items:center;margin-top:32px">`
		if page > 1 {
			prevHref := "/blog"
			if page-1 > 1 {
				prevHref = fmt.Sprintf("/blog/page/%d", page-1)
			}
			pagination += fmt.Sprintf(`<a href="%s" class="page-link">&larr; Newer</a>`, prevHref)
		} else {
			pagination += `<span></span>`
		}
		pagination += fmt.Sprintf(`<span class="page-status" style="color:var(--text2);font-size:0.85rem">Page %d of %d</span>`, page, totalPages)
		if page < totalPages {
			pagination += fmt.Sprintf(`<a href="/blog/page/%d" class="page-link">Older &rarr;</a>`, page+1)
		} else {
			pagination += `<span></span>`
		}
		pagination += `</div>`
		tmpl := `{{define "content"}}
<div class="container">
  <div style="padding:52px 0 36px">
    <div class="eyebrow">Blog</div>
    <h1 class="display" style="font-size:2rem;letter-spacing:-1px">Networking &amp; Web Guides</h1>
    <p style="font-size:0.9rem;color:var(--text2);margin-top:8px;line-height:1.7;max-width:520px">Deep-dives on DNS, SSL, networking, and web diagnostics.</p>
  </div>
  <div class="blog-grid">` + cards + `
  </div>
  ` + pagination + `
</div>
{{end}}`
		filename := "web/templates/blog.html"
		if page > 1 {
			filename = fmt.Sprintf("web/templates/blog-page-%d.html", page)
		}
		if err := os.WriteFile(filename, []byte(tmpl), 0644); err != nil {
			return err
		}
	}
	// Remove stale higher-numbered page files from previous runs
	for p := totalPages + 1; p <= totalPages+30; p++ {
		os.Remove(fmt.Sprintf("web/templates/blog-page-%d.html", p))
	}
	return nil
}

func generateSitemap(posts []BlogPost) error {
	staticURLs := []string{
		`<url><loc>https://traceqube.com/</loc><priority>1.0</priority></url>`,
		`<url><loc>https://traceqube.com/tools</loc><priority>0.9</priority></url>`,
		`<url><loc>https://traceqube.com/tools/what-is-my-ip</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/dns-lookup</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/ping-test</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/http-header-checker</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/port-checker</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/whois-lookup</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/ssl-checker</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/ip-geolocation</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/reverse-dns</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/subnet-calculator</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/down-checker</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/tools/speed-test</loc><priority>0.8</priority></url>`,
		`<url><loc>https://traceqube.com/blog</loc><priority>0.8</priority></url>`,
	}
	var blogURLs []string
	for _, p := range posts {
		blogURLs = append(blogURLs, fmt.Sprintf(`<url><loc>https://traceqube.com/blog/%s</loc><priority>0.7</priority></url>`, p.Slug))
	}
	tailURLs := []string{
		`<url><loc>https://traceqube.com/pricing</loc><priority>0.7</priority></url>`,
		`<url><loc>https://traceqube.com/api-docs</loc><priority>0.7</priority></url>`,
		`<url><loc>https://traceqube.com/about</loc><priority>0.6</priority></url>`,
		`<url><loc>https://traceqube.com/contact</loc><priority>0.6</priority></url>`,
		`<url><loc>https://traceqube.com/privacy-policy</loc><priority>0.4</priority></url>`,
		`<url><loc>https://traceqube.com/terms</loc><priority>0.4</priority></url>`,
	}
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, u := range staticURLs {
		sb.WriteString("  " + u + "\n")
	}
	for _, u := range blogURLs {
		sb.WriteString("  " + u + "\n")
	}
	for _, u := range tailURLs {
		sb.WriteString("  " + u + "\n")
	}
	sb.WriteString(`</urlset>` + "\n")
	return os.WriteFile("web/static/sitemap.xml", []byte(sb.String()), 0644)
}

func addRouteToMain(slug, title, excerpt string) error {
	data, err := os.ReadFile("cmd/server/main.go")
	if err != nil {
		return err
	}
	code := string(data)
	if strings.Contains(code, "/blog/"+slug) {
		return nil
	}
	newRoute := "\n\tapp.Get(\"/blog/" + slug + "\", func(c *fiber.Ctx) error {\n\t\treturn handlers.RenderPage(c, \"blog-" + slug + ".html\", handlers.PageData{Title: \"" + title + " | TraceQube\", Description: \"" + excerpt + "\", Canonical: \"/blog/" + slug + "\"})\n\t})"
	code = strings.Replace(code, "\t// API endpoints", newRoute+"\n\t// API endpoints", 1)
	return os.WriteFile("cmd/server/main.go", []byte(code), 0644)
}

func removeRouteFromMain(slug string) error {
	data, err := os.ReadFile("cmd/server/main.go")
	if err != nil {
		return err
	}
	code := string(data)
	pattern := `\n\tapp\.Get\("/blog/` + regexp.QuoteMeta(slug) + `", func\(c \*fiber\.Ctx\) error \{.*?\n\t\}\)`
	re := regexp.MustCompile("(?s)" + pattern)
	newCode := re.ReplaceAllString(code, "")
	if newCode == code {
		log.Printf("No route found to remove for slug: %s", slug)
		return nil
	}
	return os.WriteFile("cmd/server/main.go", []byte(newCode), 0644)
}

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = "/root/traceqube-go"
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func deleteOldTemplates(posts []BlogPost) {
	for _, p := range posts {
		path := fmt.Sprintf("web/templates/blog-%s.html", p.Slug)
		if err := os.Remove(path); err == nil {
			log.Printf("Deleted old template: %s", path)
		}
		if err := removeRouteFromMain(p.Slug); err != nil {
			log.Printf("Error removing route for %s: %v", p.Slug, err)
		} else {
			log.Printf("Removed route for old slug: %s", p.Slug)
		}
	}
}

func main() {
	godotenv.Load(".env.api")
	log.Println("Blog generator starting...")
	// Backup main.go before making any route edits, in case the build fails
	if data, err := os.ReadFile("cmd/server/main.go"); err == nil {
		os.WriteFile("cmd/server/main.go.bak", data, 0644)
		log.Println("Backed up main.go")
	}

	// Load existing posts — these will be deleted and replaced
	oldPosts, err := readBlogJSON()
	if err != nil {
		log.Fatal("Could not read blog.json:", err)
	}

	log.Println("Generating 5 new topics with AI...")
	newTopics, err := generateTopics(5)
	if err != nil {
		log.Printf("Topic generation error: %v", err)
		return
	}
	log.Printf("Got %d topics from AI", len(newTopics))

	var newPosts []BlogPost
	for _, topic := range newTopics {
		log.Printf("Generating post: %s", topic.Title)
		content, err := generatePost(topic.Title, topic.Category)
		if err != nil {
			log.Printf("Error generating post: %v", err)
			continue
		}
                content = normalizeContent(content)
		date := time.Now().Format("January 2, 2006")
		if err := createTemplate(topic.Slug, topic.Title, topic.Category, date, content); err != nil {
			log.Printf("Template error: %v", err)
			continue
		}
		newPost := BlogPost{
			Slug:     topic.Slug,
			Title:    topic.Title,
			Category: topic.Category,
			Date:     date,
			ReadTime: "5 min read",
			Excerpt:  topic.Excerpt,
		}
		newPosts = append(newPosts, newPost)
		if err := addRouteToMain(topic.Slug, topic.Title, topic.Excerpt); err != nil {
			log.Printf("Route error: %v", err)
		}
		log.Printf("Done: %s", topic.Title)
	}

	if len(newPosts) == 0 {
		log.Println("No new posts generated, aborting.")
		return
	}

	// Delete old templates and save new blog.json with only 5 new posts
	log.Println("Skipping deletion — keeping old posts")
	// deleteOldTemplates(oldPosts) -- disabled, old posts are now kept permanently

	allPosts := append(oldPosts, newPosts...)
	rebuildBlogTemplate(allPosts)
	if err := generateSitemap(allPosts); err != nil {
		log.Printf("Sitemap generation error: %v", err)
	} else {
		log.Println("Sitemap regenerated")
	}
	data, _ := json.MarshalIndent(allPosts, "", "  ")
	os.WriteFile("data/blog.json", data, 0644)
	log.Printf("Saved %d total posts to blog.json", len(allPosts))

	log.Println("Rebuilding site...")
	os.Remove("traceqube")
	out, err := runCmd("go", "build", "-o", "traceqube", "cmd/server/main.go")
	if err != nil {
		log.Printf("Build error: %s", out)
		log.Println("Restoring main.go from backup due to build failure...")
		if backup, berr := os.ReadFile("cmd/server/main.go.bak"); berr == nil {
			os.WriteFile("cmd/server/main.go", backup, 0644)
			log.Println("Restored main.go.bak -> main.go")
		} else {
			log.Printf("Could not restore backup: %v", berr)
		}
	} else {
		runCmd("systemctl", "restart", "traceqube-go")
		log.Println("Site rebuilt and restarted!")
	}
}

func stripLeakedTokens(s string) string {
	s = regexp.MustCompile(`<\|[a-z_]+\|>`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i)</h[1-6]assistant`).ReplaceAllString(s, "</h2")
	return s
}

var mdConverter = goldmark.New(goldmark.WithRendererOptions(html.WithUnsafe()))

// normalizeContent guards against the model ignoring the "HTML only" instruction
// and returning raw markdown instead. WithUnsafe lets already-valid HTML pass
// through unchanged, while markdown syntax (**, ###, ```, - lists) gets properly
// converted to real HTML tags.
func normalizeContent(raw string) string {
	raw = stripLeakedTokens(raw)
	var buf bytes.Buffer
	if err := mdConverter.Convert([]byte(raw), &buf); err != nil {
		log.Printf("markdown normalization failed, using raw content: %v", err)
		return raw
	}
	out := buf.String()
	// The page template already renders its own <h1>{title}</h1>.
	// If the AI's markdown started with its own top-level heading,
	// goldmark will have converted it to a leading <h1>...</h1> here too.
	// Strip exactly one leading <h1> to avoid a duplicate title.
	out = strings.TrimSpace(out)
	if strings.HasPrefix(out, "<h1>") {
		if end := strings.Index(out, "</h1>"); end != -1 {
			out = strings.TrimSpace(out[end+len("</h1>"):])
		}
	}
	return out
}
