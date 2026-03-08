---
# project-idea-cli-search-engine-rpfv
title: Schema doc
status: completed
type: task
priority: normal
created_at: 2026-03-07T17:46:16Z
updated_at: 2026-03-07T17:47:46Z
parent: project-idea-cli-search-engine-rgag
---

Document all JSON schemas used by qry: the config.toml structure, the adapter input schema (stdin), the adapter output schema (stdout), and the error schema (stderr). Output: docs/schema.md

## Summary of Changes\n\nWrote docs/schema.md covering all data structures:\n- Config file (config.toml) — full field reference with duration format note\n- Adapter request (stdin) — query, num, config passthrough\n- Adapter response (stdout) — title, url, snippet array\n- Adapter error (stderr + exit code) — standard error codes and rules\n- qry final output — passthrough + all-adapters-failed envelope
