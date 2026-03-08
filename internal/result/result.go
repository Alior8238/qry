package result

// Result is a single search result returned by an adapter.
type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// AdapterError is the error structure an adapter writes to stderr on failure.
type AdapterError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Attempt records the outcome of a single adapter invocation (used in failure reporting).
type Attempt struct {
	Adapter string `json:"adapter"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// FirstOutput is the stdout shape for "first" mode — a plain array of results.
type FirstOutput = []Result

// MergeOutput is the stdout shape for "merge" mode — results plus any warnings.
type MergeOutput struct {
	Results  []Result `json:"results"`
	Warnings []string `json:"warnings,omitempty"`
}

// FailureOutput is written to stderr when all adapters are exhausted.
type FailureOutput struct {
	Error    string    `json:"error"`
	Message  string    `json:"message"`
	Attempts []Attempt `json:"attempts"`
}

// Deduplicate removes results with duplicate URLs, keeping the first occurrence.
// Pool order determines priority — results earlier in the slice win.
func Deduplicate(results []Result) []Result {
	seen := make(map[string]struct{}, len(results))
	out := make([]Result, 0, len(results))
	for _, r := range results {
		if _, ok := seen[r.URL]; ok {
			continue
		}
		seen[r.URL] = struct{}{}
		out = append(out, r)
	}
	return out
}
