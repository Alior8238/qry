---
# project-idea-cli-search-engine-8iqp
title: Adapter contract doc
status: completed
type: task
priority: normal
created_at: 2026-03-07T17:46:19Z
updated_at: 2026-03-07T19:09:08Z
parent: project-idea-cli-search-engine-rgag
---

Document the full adapter contract: what an adapter is, how to implement one, the stdin/stdout/stderr protocol, exit codes, lifecycle, and a minimal reference adapter example (shell script). Output: docs/adapters.md

## Summary of Changes\n\nWrote docs/adapters.md covering:\n- What an adapter is and the core philosophy\n- Full lifecycle diagram\n- Protocol: invocation, stdin, stdout (success), stderr (failure)\n- Rules table\n- Reference mock adapter in shell\n- Step-by-step guide for building a real adapter in Go\n- Naming convention (qry-adapter-<name>)\n- Registration in config.toml\n- Checklist for adapter authors
