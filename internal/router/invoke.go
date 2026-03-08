package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/justestif/qry/internal/config"
	"github.com/justestif/qry/internal/result"
)

// Request is the JSON payload written to an adapter's stdin.
type Request struct {
	Query  string            `json:"query"`
	Num    int               `json:"num"`
	Config map[string]string `json:"config"`
}

// invokeAdapter runs a single adapter binary, writes the request to stdin,
// and returns parsed results or a structured error.
func invokeAdapter(ctx context.Context, name string, adapter config.Adapter, query string) ([]result.Result, *result.Attempt) {
	req := Request{
		Query:  query,
		Num:    adapter.Num,
		Config: adapter.Config,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, &result.Attempt{
			Adapter: name,
			Error:   "unknown",
			Message: fmt.Sprintf("failed to marshal request: %s", err),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, adapter.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, adapter.Bin)
	cmd.Stdin = bytes.NewReader(reqJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Parse stderr for structured error
		adapterErr := &result.AdapterError{}
		if jsonErr := json.Unmarshal(stderr.Bytes(), adapterErr); jsonErr != nil {
			// stderr was not valid JSON — surface raw message
			adapterErr.Error = "unknown"
			adapterErr.Message = stderr.String()
		}
		if adapterErr.Error == "" {
			adapterErr.Error = "unknown"
		}
		if adapterErr.Message == "" {
			adapterErr.Message = err.Error()
		}
		return nil, &result.Attempt{
			Adapter: name,
			Error:   adapterErr.Error,
			Message: adapterErr.Message,
		}
	}

	// Parse stdout for results
	var results []result.Result
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		return nil, &result.Attempt{
			Adapter: name,
			Error:   "unknown",
			Message: fmt.Sprintf("failed to parse adapter output: %s", err),
		}
	}

	return results, nil
}
