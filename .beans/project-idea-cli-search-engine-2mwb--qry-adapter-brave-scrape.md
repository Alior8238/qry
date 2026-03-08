---
# project-idea-cli-search-engine-2mwb
title: qry-adapter-brave-scrape
status: completed
type: task
priority: normal
created_at: 2026-03-08T02:58:33Z
updated_at: 2026-03-08T03:21:57Z
parent: project-idea-cli-search-engine-usbw
---

Adapter for Brave Search via scraping. No API key required. An alternative to brave-api for users who don't want to pay. Output: adapters/brave-scrape/main.go

## Summary of Changes\n\nadapters/brave-scrape/main.go:\n- GETs https://search.brave.com/search?q=...&source=web\n- Forces HTTP/1.1 (ForceAttemptHTTP2: false) — Brave fingerprints HTTP/2\n- Splits page into per-result blocks using 'snippet svelte-jmfu5f' class + data-type=web\n- Extracts URL from first <a href> in each block\n- Extracts title from title= attribute on .title.search-snippet-title div\n- Extracts snippet from .content.desktop-default-regular div, strips tags + unescapes HTML\n- Maps 429 → rate_limited, captcha page → rate_limited, no blocks → unavailable\n- Supports optional config: safe (moderate/strict/off)
