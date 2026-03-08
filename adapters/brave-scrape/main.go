// qry-adapter-brave-scrape searches via Brave Search HTML scraping.
// No API key required.
//
// Optional config:
//   safe  — "moderate" (default) | "strict" | "off"
package main

import (
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
	reBlock   = regexp.MustCompile(`class="snippet[^"]*svelte-jmfu5f[^"]*" data-pos="\d+" data-type="web"`)
	reURL     = regexp.MustCompile(`<a href="(https?://[^"]+)"`)
	reTitle   = regexp.MustCompile(`class="title search-snippet-title[^"]*" title="([^"]+)"`)
	reSnippet = regexp.MustCompile(`class="content desktop-default-regular[^"]*">(.*?)</div>`)
	reTag     = regexp.MustCompile(`<[^>]+>`)
)

func stripTags(s string) string {
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

// splitBlocks splits the page into per-result chunks using the snippet class as a delimiter.
func splitBlocks(page string) []string {
	indices := reBlock.FindAllStringIndex(page, -1)
	if len(indices) == 0 {
		return nil
	}
	blocks := make([]string, 0, len(indices))
	for i, loc := range indices {
		end := len(page)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		blocks = append(blocks, page[loc[0]:end])
	}
	return blocks
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
	safe := req.Config["safe"]
	if safe == "" {
		safe = "moderate"
	}

	// 3. Build request URL
	params := url.Values{}
	params.Set("q", req.Query)
	params.Set("source", "web")
	params.Set("safe", safe)
	endpoint := "https://search.brave.com/search?" + params.Encode()

	// 4. Execute request — HTTP/1.1 only, Brave fingerprints HTTP/2
	transport := &http.Transport{
		ForceAttemptHTTP2: false,
	}
	client := &http.Client{Transport: transport}

	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		writeError("unknown", "failed to build HTTP request: "+err.Error())
		os.Exit(1)
	}
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	httpReq.Header.Set("Accept", "text/html,application/xhtml+xml")
	httpReq.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(httpReq)
	if err != nil {
		writeError("unavailable", "HTTP request failed: "+err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusTooManyRequests:
		writeError("rate_limited", "Brave Search returned 429 Too Many Requests")
		os.Exit(1)
	default:
		writeError("unavailable", fmt.Sprintf("Brave Search returned unexpected status %d", resp.StatusCode))
		os.Exit(1)
	}

	// 5. Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeError("unknown", "failed to read response body: "+err.Error())
		os.Exit(1)
	}
	page := string(body)

	// 6. Parse result blocks
	blocks := splitBlocks(page)
	if len(blocks) == 0 {
		// Either blocked or page structure changed
		if strings.Contains(page, "captcha") || strings.Contains(page, "rate-limit") {
			writeError("rate_limited", "Brave Search served a challenge page")
		} else {
			writeError("unavailable", "no results found or page structure changed")
		}
		os.Exit(1)
	}

	// 7. Extract fields from each block
	num := req.Num
	if num <= 0 {
		num = len(blocks)
	}

	results := make([]Result, 0, num)
	for _, b := range blocks {
		if len(results) >= num {
			break
		}

		urlM := reURL.FindStringSubmatch(b)
		if urlM == nil {
			continue
		}
		u := urlM[1]

		title := ""
		if m := reTitle.FindStringSubmatch(b); m != nil {
			title = html.UnescapeString(m[1])
		}

		snippet := ""
		if m := reSnippet.FindStringSubmatch(b); m != nil {
			snippet = stripTags(m[1])
		}

		results = append(results, Result{
			Title:   title,
			URL:     u,
			Snippet: snippet,
		})
	}

	// 8. Write results to stdout
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		writeError("unknown", "failed to encode results: "+err.Error())
		os.Exit(1)
	}
}
