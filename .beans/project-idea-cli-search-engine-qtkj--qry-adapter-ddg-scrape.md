---
# project-idea-cli-search-engine-qtkj
title: qry-adapter-ddg-scrape
status: completed
type: task
priority: normal
created_at: 2026-03-08T02:58:33Z
updated_at: 2026-03-08T03:08:21Z
parent: project-idea-cli-search-engine-usbw
---

Adapter for DuckDuckGo via scraping. No API key required. Output: adapters/ddg-scrape/main.go

## Summary of Changes\n\nadapters/ddg-scrape/main.go:\n- POSTs to https://html.duckduckgo.com/html with kf=-1, kh=1, k1=-1 params\n- Forces HTTP/1.1 (ForceAttemptHTTP2: false) — DDG blocks HTTP/2 clients\n- Handles gzip decompression\n- Maps 202 → rate_limited, anomaly-modal → rate_limited\n- Parses result__a (URL+title) and result__snippet with regexp\n- Strips HTML tags and unescapes HTML entities from titles/snippets\n- Supports optional config: region (default: wt-wt), safe (default: 1)
