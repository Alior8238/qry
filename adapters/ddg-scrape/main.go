// qry-adapter-ddg-scrape searches via DuckDuckGo's HTML endpoint.
// No API key required.
//
// Optional config:
//   region  — DDG region code e.g. "us-en" (default: "wt-wt" = no region)
//   safe    — safe search: "1" (on, default) or "-1" (off)
package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// --- qry adapter protocol types ---

type Request struct {
	Query  string            `json:"query"`
	Num    int               `json:"num"`
	Config map[string]string `json:"config"`
}

type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// --- HTML parsing ---

var (
	reURL     = regexp.MustCompile(`class="result__a" href="([^"]+)"`)
	reTitle   = regexp.MustCompile(`class="result__a"[^>]*>([\s\S]*?)</a>`)
	reSnippet = regexp.MustCompile(`class="result__snippet"[^>]*>([\s\S]*?)</a>`)
	reTag     = regexp.MustCompile(`<[^>]+>`)
)

func stripTags(s string) string {
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

// --- error helpers ---

func writeError(code, message string) {
	json.NewEncoder(os.Stderr).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

func main() {
	// 1. Read and parse the qry request from stdin
	var req Request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		writeError("unknown", "failed to parse request: "+err.Error())
		os.Exit(1)
	}

	if req.Query == "" {
		writeError("invalid_query", "query must not be empty")
		os.Exit(1)
	}

	// 2. Read optional config
	region := req.Config["region"]
	if region == "" {
		region = "wt-wt"
	}
	safe := req.Config["safe"]
	if safe == "" {
		safe = "1"
	}

	// 3. Build POST body
	// DDG HTML endpoint requires specific params to avoid bot detection.
	// kf=-1: no favicons, kh=1: HTTPS always, k1=-1: no ads
	form := url.Values{}
	form.Set("q", req.Query)
	form.Set("b", "")
	form.Set("kf", "-1")
	form.Set("kh", "1")
	form.Set("kl", region)
	form.Set("kp", safe)
	form.Set("k1", "-1")

	// 4. Build HTTP client — must use HTTP/1.1, DDG blocks HTTP/2 clients
	transport := &http.Transport{
		ForceAttemptHTTP2: false,
	}
	client := &http.Client{Transport: transport}

	httpReq, err := http.NewRequest("POST", "https://html.duckduckgo.com/html", strings.NewReader(form.Encode()))
	if err != nil {
		writeError("unknown", "failed to build HTTP request: "+err.Error())
		os.Exit(1)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept-Encoding", "gzip")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	httpReq.Header.Set("DNT", "1")

	// 5. Execute request
	resp, err := client.Do(httpReq)
	if err != nil {
		writeError("unavailable", "HTTP request failed: "+err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 202 {
		writeError("rate_limited", "DuckDuckGo returned 202 — rate limited, try again shortly")
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		writeError("unavailable", fmt.Sprintf("DuckDuckGo returned unexpected status %d", resp.StatusCode))
		os.Exit(1)
	}

	// 6. Decompress if gzip
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			writeError("unknown", "failed to decompress gzip response: "+err.Error())
			os.Exit(1)
		}
		defer gz.Close()
		reader = gz
	}

	// 7. Parse HTML
	body, err := io.ReadAll(reader)
	if err != nil {
		writeError("unknown", "failed to read response body: "+err.Error())
		os.Exit(1)
	}
	page := string(body)

	if strings.Contains(page, "anomaly-modal") {
		writeError("rate_limited", "DuckDuckGo served a bot challenge — try again shortly")
		os.Exit(1)
	}

	urls := reURL.FindAllStringSubmatch(page, -1)
	titles := reTitle.FindAllStringSubmatch(page, -1)
	snippets := reSnippet.FindAllStringSubmatch(page, -1)

	// 8. Build results
	num := req.Num
	if num <= 0 {
		num = len(urls)
	}

	results := make([]Result, 0, num)
	for i := 0; i < len(urls) && len(results) < num; i++ {
		u := urls[i][1]
		title := ""
		if i < len(titles) {
			title = stripTags(titles[i][1])
		}
		snippet := ""
		if i < len(snippets) {
			snippet = stripTags(snippets[i][1])
		}
		results = append(results, Result{
			Title:   title,
			URL:     u,
			Snippet: snippet,
		})
	}

	// 9. Write results to stdout
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		writeError("unknown", "failed to encode results: "+err.Error())
		os.Exit(1)
	}
}
