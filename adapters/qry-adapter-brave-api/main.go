// qry-adapter-brave-api searches via the Brave Search API.
//
// Required config:
//   api_key  — Brave Search API key (X-Subscription-Token)
//
// Optional config:
//   country      — 2-char country code e.g. "US" (default: unset)
//   search_lang  — content language e.g. "en" (default: unset)
//   freshness    — pd | pw | pm | py | date range (default: unset)
package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
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

// --- Brave API response types ---

type braveResponse struct {
	Web struct {
		Results []braveResult `json:"results"`
	} `json:"web"`
}

type braveResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
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

	// 2. Validate required config
	apiKey := req.Config["api_key"]
	if apiKey == "" {
		writeError("auth_failed", "api_key is required but not set in adapter config")
		os.Exit(1)
	}

	if req.Query == "" {
		writeError("invalid_query", "query must not be empty")
		os.Exit(1)
	}

	// 3. Build request URL
	count := req.Num
	if count <= 0 || count > 20 {
		count = 20
	}

	params := url.Values{}
	params.Set("q", req.Query)
	params.Set("count", strconv.Itoa(count))

	if v := req.Config["country"]; v != "" {
		params.Set("country", v)
	}
	if v := req.Config["search_lang"]; v != "" {
		params.Set("search_lang", v)
	}
	if v := req.Config["freshness"]; v != "" {
		params.Set("freshness", v)
	}

	endpoint := "https://api.search.brave.com/res/v1/web/search?" + params.Encode()

	// 4. Execute the HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		writeError("unknown", "failed to build HTTP request: "+err.Error())
		os.Exit(1)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Accept-Encoding", "gzip")
	httpReq.Header.Set("X-Subscription-Token", apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		writeError("unavailable", "HTTP request failed: "+err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 5. Handle HTTP error codes
	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusUnauthorized, http.StatusForbidden:
		writeError("auth_failed", fmt.Sprintf("Brave API returned %d — check your api_key", resp.StatusCode))
		os.Exit(1)
	case http.StatusTooManyRequests:
		writeError("rate_limited", "Brave API returned 429 Too Many Requests")
		os.Exit(1)
	case http.StatusBadRequest:
		writeError("invalid_query", fmt.Sprintf("Brave API returned 400 Bad Request for query: %q", req.Query))
		os.Exit(1)
	default:
		writeError("unavailable", fmt.Sprintf("Brave API returned unexpected status %d", resp.StatusCode))
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

	// 7. Parse the Brave response
	var braveResp braveResponse
	if err := json.NewDecoder(reader).Decode(&braveResp); err != nil {
		writeError("unknown", "failed to parse Brave API response: "+err.Error())
		os.Exit(1)
	}

	// 8. Map to qry result format
	results := make([]Result, 0, len(braveResp.Web.Results))
	for _, r := range braveResp.Web.Results {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Description,
		})
	}

	// 9. Write results to stdout
	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		writeError("unknown", "failed to encode results: "+err.Error())
		os.Exit(1)
	}
}
