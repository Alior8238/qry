# qry — MVP Definition

## What

`qry` is a terminal-native, agent-first web search CLI written in Go.

It acts as a **hub**: you run one command, it routes your query through a configured search adapter,
and returns structured JSON results. Adapters are separate binaries installed independently and
registered in a config file. The same search engine can have multiple adapters (e.g. a Brave API
adapter and a Brave scraping adapter) — you pick which one to use.

### MVP CLI shape

```bash
# Basic search — uses default adapter from config
qry "what is the latest version of numpy"

# Override adapter for this invocation
qry --adapter brave-api "what is the latest version of numpy"

# Limit results
qry --num 5 "what is the latest version of numpy"

# Output is always JSON
[
  { "title": "...", "url": "...", "snippet": "..." },
  ...
]
```

---

## Why

### The problem

Terminal agents (like pi) need web search to do research tasks. The current standard solution
(`ddgr` + DuckDuckGo) has two key failure modes:

1. **Rate limiting** — DuckDuckGo blocks requests when an agent fires many queries in a short
   window, breaking multi-step research tasks.
2. **Single source** — being tied to one engine means no fallback when results are poor or
   unavailable.

Existing tools don't solve this:

- `ddgr` / `surfraw` — single engine, human-oriented (opens browser), no structured output
- `agent-browser` — browser _automation_, not search; a different problem entirely

### The solution

`qry` decouples the **search interface** from the **search backend**. You get:

- A stable CLI contract that agents and humans can rely on
- Swappable adapters so you are never locked into one engine or one strategy
- Config-driven control over adapter selection, fallback behavior, and timeouts
- Always-JSON output — clean for agents, pipeable for humans

---

## Alternatives Considered

Understanding what already exists makes `qry`'s position clear.

### curl / raw HTTP

The baseline. You can query any search API directly with curl. No structure, no abstraction, no
config — you write the whole integration yourself every time. Works, but zero reuse, no fallback
logic, and no agent ergonomics. `qry` is what you'd build on top of curl, standardized.

### CLI scrapers — ddgr, googler, surfraw

| Tool      | Engine          | Output        | Agent-friendly | Status                                              |
| --------- | --------------- | ------------- | -------------- | --------------------------------------------------- |
| `ddgr`    | DuckDuckGo      | JSON          | Yes            | Rate-limits hard under agent load                   |
| `googler` | Google (scrape) | JSON          | Yes            | **Broken** — Google changed SERP structure Jan 2025 |
| `surfraw` | 100+ engines    | Opens browser | No             | Human-only, no structured output                    |

The core problem: scraping-based tools are getting **less reliable over time**. Google actively
broke scrapers in January 2025. DuckDuckGo blocks agents making concurrent requests. These are not
bugs that get fixed — they are the nature of scraping against a hostile target.

### Official provider APIs

| Provider           | What it is                               | Structured results?                        | Cost                                                  |
| ------------------ | ---------------------------------------- | ------------------------------------------ | ----------------------------------------------------- |
| Brave Search API   | Independent search index, clean REST API | Yes — `{title, url, description}`          | ~$5/month for ~1,000 queries (free tier removed 2025) |
| Perplexity (Sonar) | LLM that cites sources                   | No — synthesized answer, not a result list | Token-based pricing, pay-as-you-go                    |
| Firecrawl          | Web crawling + content extraction        | No — page content, not search results      | Separate use case — give it a URL, get clean content  |

Key observations:

- **Brave** is the cleanest option for structured web search results, but now requires payment and
  locks you into one provider
- **Perplexity** is a different tool — great for synthesis and Q&A, not for returning a raw list of
  results an agent can reason over
- **Firecrawl** is complementary — pair it with `qry` (qry finds URLs, Firecrawl fetches content),
  not a replacement

### The gap `qry` fills

No existing tool gives you:

1. A **stable CLI contract** that doesn't break when an upstream scraping target changes
2. **Swappable backends** so you can use an API today, a scraper tomorrow, or both with fallback
3. **Agent-first output** (always JSON, no pager, no browser, no interaction)
4. The ability to have **multiple strategies for the same engine** (e.g. Brave API adapter vs. a
   Brave scraping adapter) and choose based on cost, reliability, or context

The scraping tools are becoming less viable. The API tools are becoming more expensive and more
siloed. `qry` sits in between: a stable interface that lets you swap the backend as the landscape
shifts — without changing your agents or your workflows.

---

## Core Concepts

### Hub + Adapter model

`qry` is the hub. Adapters are the backends. An adapter is a standalone binary that:

- Receives a query request (via stdin as JSON)
- Returns results (via stdout as JSON)
- Speaks a defined interface — nothing more

This mirrors how `kubectl` plugins and `git` extensions work. Any language can implement an adapter.

### Config file

Location: `~/.config/qry/config.toml`

Responsibilities:

- Declare which adapters are installed and where their binaries live
- Set the default adapter
- Configure per-adapter settings (API keys, timeouts, result limits)
- Control routing behavior (e.g. fallback order, timeout thresholds)

```toml
default_adapter = "brave-api"

[adapters.brave-api]
  bin = "/usr/local/bin/qry-adapter-brave-api"
  api_key = "YOUR_KEY"
  timeout = "5s"
  results = 10

[adapters.google-api]
  bin = "/usr/local/bin/qry-adapter-google-api"
  api_key = "YOUR_KEY"
  cx = "YOUR_CX"
  timeout = "5s"
  results = 10

[routing]
  fallback = ["brave-api", "google-api"]  # try in order on failure
```

---

## MVP Scope

### In scope

| Feature                     | Notes                                                       |
| --------------------------- | ----------------------------------------------------------- |
| Core `qry` binary in Go     | Parses config, invokes adapter, returns JSON                |
| Adapter interface spec      | Stdin/stdout JSON contract all adapters must follow         |
| `qry-adapter-brave-api`     | Uses Brave Search API — requires API key                    |
| `qry-adapter-google-api`    | Uses Google Custom Search API — requires API key + cx       |
| Config file (`config.toml`) | Adapter registration, defaults, per-adapter config          |
| CLI flags                   | `--adapter`, `--num`, override config values per invocation |
| Always-JSON output          | `[{title, url, snippet}]` — no other output modes           |
| Fallback routing            | Try next adapter in config order on timeout or error        |

### Out of scope for MVP

| Feature                | Reason                                                   |
| ---------------------- | -------------------------------------------------------- |
| Scraping adapters      | Valid future plugins — not needed to prove the model     |
| TUI / interactive mode | Human UX is a later concern                              |
| Page content fetching  | Out of search scope; pair with `agent-browser` or `curl` |
| Image / news search    | Web results only for now                                 |
| Adapter auto-install   | Manual binary install + config registration for MVP      |

---

## Adapter Interface Contract (v1)

Every adapter binary must implement this protocol:

**Input** (stdin):

```json
{
  "query": "what is the latest version of numpy",
  "num": 10
}
```

**Output** (stdout):

```json
[
  {
    "title": "NumPy 2.0 Release Notes",
    "url": "https://numpy.org/doc/stable/release/2.0.0-notes.html",
    "snippet": "NumPy 2.0.0 is the first major release since 2006..."
  }
]
```

**On error** (stderr + non-zero exit):

```json
{ "error": "rate_limited", "message": "429 from Brave API" }
```

`qry` treats a non-zero exit as a failure and moves to the next fallback adapter if configured.

---

## Success Criteria for MVP

- [ ] `qry "some query"` returns valid JSON results using the configured default adapter
- [ ] `--adapter` flag overrides the default for a single invocation
- [ ] Fallback routing works: if adapter A fails, adapter B is tried automatically
- [ ] Both `qry-adapter-brave-api` and `qry-adapter-google-api` are functional with valid API keys
- [ ] A new adapter can be added by: installing a binary + adding a `[adapters.name]` block to config
- [ ] Usable as a pi skill with no changes to `qry` itself
