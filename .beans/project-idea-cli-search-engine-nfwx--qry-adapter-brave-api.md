---
# project-idea-cli-search-engine-nfwx
title: qry-adapter-brave-api
status: completed
type: task
priority: high
created_at: 2026-03-08T02:58:33Z
updated_at: 2026-03-08T03:02:41Z
parent: project-idea-cli-search-engine-usbw
---

Adapter for the Brave Search API. Requires an API key configured in config.toml. Output: adapters/brave-api/main.go

## Summary of Changes\n\nadapters/brave-api/main.go:\n- Reads qry request from stdin\n- Validates api_key (auth_failed if missing)\n- Builds GET request to https://api.search.brave.com/res/v1/web/search\n- Sets X-Subscription-Token, Accept, Accept-Encoding: gzip headers\n- Supports optional config: country, search_lang, freshness\n- Maps HTTP errors to standard qry error codes: auth_failed, rate_limited, invalid_query, unavailable\n- Handles gzip response decompression\n- Maps braveResult{title,url,description} → qry Result{title,url,snippet}\n- Exits 0 with JSON array on success, 1 with JSON error on stderr on failure
