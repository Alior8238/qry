// qry-adapter-mock is a reference adapter that returns hardcoded results.
// Use it to validate the qry pipeline end-to-end without a real backend.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

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

func writeError(code, message string) {
	json.NewEncoder(os.Stderr).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

func main() {
	var req Request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		writeError("unknown", "failed to parse request: "+err.Error())
		os.Exit(1)
	}

	results := []Result{
		{
			Title:   fmt.Sprintf("Mock Result 1 for: %s", req.Query),
			URL:     "https://example.com/result-1",
			Snippet: "This is a mock search result for testing qry.",
		},
		{
			Title:   fmt.Sprintf("Mock Result 2 for: %s", req.Query),
			URL:     "https://example.com/result-2",
			Snippet: "Another mock result to validate the adapter contract.",
		},
	}

	// Respect the num limit
	if req.Num > 0 && req.Num < len(results) {
		results = results[:req.Num]
	}

	if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
		writeError("unknown", "failed to encode results: "+err.Error())
		os.Exit(1)
	}
}
