# qry — Adapter Contract

An adapter is any executable that speaks the qry adapter protocol. This document defines
what an adapter is, how `qry` invokes it, what it must do, and how to build one.

For the full JSON schemas referenced here, see [schema.md](./schema.md).

---

## What is an adapter?

An adapter is a **standalone executable** (binary, script, or any runnable file) that:

1. Reads a search request from **stdin** as JSON
2. Executes a search against some backend (an API, a scraper, a local index — anything)
3. Writes results to **stdout** as JSON, or an error to **stderr** as JSON

That's the entire contract. `qry` knows nothing about how the adapter works internally.
The adapter owns its backend, its credentials, and its logic. `qry` owns the orchestration.

This means:
- Adapters can be written in **any language**
- An API-based adapter and a scraping adapter for the same engine are both valid
- Adapters can be swapped, mixed, and extended without changing `qry`

---

## Lifecycle

```
qry invokes adapter binary
        │
        ▼
adapter reads JSON from stdin
        │
        ▼
adapter executes search
        │
   ┌────┴────┐
success    failure
   │            │
   ▼            ▼
write JSON   write JSON    exit non-zero
array to     error to
stdout       stderr
   │
   ▼
exit 0
```

`qry` enforces a **timeout** per adapter (configured in `config.toml`). If the adapter does not
exit within the timeout, `qry` kills the process and treats it as a `timeout` failure.

---

## Protocol

### Invocation

`qry` invokes the adapter as a subprocess:

```bash
/path/to/adapter-binary
```

No command-line arguments are passed. All input is via stdin.

### Input (stdin)

`qry` writes a single JSON object to the adapter's stdin, then closes stdin:

```json
{
  "query":  "what is the latest version of numpy",
  "num":    10,
  "config": {
    "api_key": "YOUR_KEY"
  }
}
```

The adapter must read stdin to EOF before processing.

### Output — success (stdout)

Write a JSON array to stdout and exit 0:

```json
[
  {
    "title":   "NumPy 2.0 Release Notes",
    "url":     "https://numpy.org/doc/stable/release/2.0.0-notes.html",
    "snippet": "NumPy 2.0.0 is the first major release since 2006..."
  }
]
```

- An empty array `[]` is valid — it means the search returned no results, not that something failed
- Write **only** the JSON array to stdout — no logging, no extra text
- Results should be ordered by relevance (most relevant first)
- Return at most `num` results

### Output — failure (stderr + exit code)

Write a JSON error object to stderr and exit non-zero:

```json
{
  "error":   "rate_limited",
  "message": "429 Too Many Requests from Brave API"
}
```

- Write **only** the JSON error to stderr — no extra text
- Write **nothing** to stdout on failure
- Use a standard error code from [schema.md](./schema.md#standard-error-codes)
- `qry` will handle fallback routing based on the error code

---

## Rules

| Rule | Details |
|---|---|
| No args | `qry` passes no CLI arguments — all input is stdin |
| Stdin to EOF | Read all of stdin before doing anything |
| Stdout is sacred | On success: only the JSON array. On failure: nothing. |
| Stderr is sacred | On failure: only the JSON error. On success: nothing. |
| Exit codes | 0 = success, non-zero = failure (any non-zero value is treated as failure) |
| Timeout | The adapter process will be killed if it exceeds the configured timeout |
| No state | Adapters are stateless — each invocation is independent |
| No side effects | Adapters must not write files, open browsers, or produce side effects |

---

## Reference Implementation

A minimal adapter in shell. This returns hardcoded results and is useful for testing the
`qry` pipeline end-to-end without a real backend.

**`qry-adapter-mock`**

```bash
#!/usr/bin/env bash
set -euo pipefail

# Read stdin (required — must consume input before responding)
input=$(cat)

# Optionally parse the query for logging/debugging (not required)
# query=$(echo "$input" | jq -r '.query')

# Return hardcoded results
cat <<'EOF'
[
  {
    "title":   "Mock Result 1",
    "url":     "https://example.com/result-1",
    "snippet": "This is a mock search result for testing qry."
  },
  {
    "title":   "Mock Result 2",
    "url":     "https://example.com/result-2",
    "snippet": "Another mock result to validate the adapter contract."
  }
]
EOF

exit 0
```

Save as `qry-adapter-mock`, make executable, register in config:

```toml
[adapters.mock]
  bin = "/usr/local/bin/qry-adapter-mock"
```

---

## Building a Real Adapter

### Step 1 — Read and parse stdin

```go
// Go example
var req struct {
    Query  string            `json:"query"`
    Num    int               `json:"num"`
    Config map[string]string `json:"config"`
}
if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
    writeError("unknown", "failed to parse request: "+err.Error())
    os.Exit(1)
}
```

### Step 2 — Extract config

Your adapter's config block from `config.toml` is passed verbatim as the `config` object.
Pull what you need:

```go
apiKey := req.Config["api_key"]
if apiKey == "" {
    writeError("auth_failed", "api_key is required but not set in config")
    os.Exit(1)
}
```

### Step 3 — Execute the search

Call your backend. Handle errors by mapping them to standard error codes:

```go
results, err := callBraveAPI(apiKey, req.Query, req.Num)
if err != nil {
    if isRateLimit(err) {
        writeError("rate_limited", err.Error())
    } else {
        writeError("unknown", err.Error())
    }
    os.Exit(1)
}
```

### Step 4 — Write results and exit 0

```go
if err := json.NewEncoder(os.Stdout).Encode(results); err != nil {
    writeError("unknown", "failed to encode results: "+err.Error())
    os.Exit(1)
}
os.Exit(0)
```

### Error helper

```go
func writeError(code, message string) {
    json.NewEncoder(os.Stderr).Encode(map[string]string{
        "error":   code,
        "message": message,
    })
}
```

---

## Naming Convention

Adapter binaries should follow the naming pattern:

```
qry-adapter-<name>
```

Examples:
- `qry-adapter-brave-api`
- `qry-adapter-google-api`
- `qry-adapter-ddg-scrape`
- `qry-adapter-brave-scrape`
- `qry-adapter-mock`

The `<name>` part is what you use in `config.toml` as the adapter key and in `--adapter` flag.

---

## Registration

Install the binary somewhere on your system, then add a block to `~/.config/qry/config.toml`:

```toml
[adapters.brave-api]
  bin = "/usr/local/bin/qry-adapter-brave-api"
  timeout = "5s"

  [adapters.brave-api.config]
    api_key = "YOUR_KEY"
```

Then add it to your pool or fallback:

```toml
[routing]
  mode = "first"
  pool = ["brave-api"]
  fallback = ["ddg-scrape"]
```

No restart required — `qry` reads config on every invocation.

---

## Checklist for adapter authors

- [ ] Reads all of stdin before doing anything
- [ ] Parses the JSON request envelope
- [ ] Validates required config fields and exits with `auth_failed` if missing
- [ ] Returns a JSON array on stdout on success, nothing else
- [ ] Returns a JSON error on stderr on failure, nothing on stdout
- [ ] Uses standard error codes from schema.md
- [ ] Exits 0 on success, non-zero on failure
- [ ] Binary named `qry-adapter-<name>`
- [ ] Tested with the mock adapter pattern before wiring to a real backend
