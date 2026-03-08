---
# project-idea-cli-search-engine-ykdu
title: Architecture doc
status: completed
type: task
priority: normal
created_at: 2026-03-07T17:46:13Z
updated_at: 2026-03-08T02:31:24Z
parent: project-idea-cli-search-engine-rgag
---

Document the overall architecture of qry: the hub+adapter model, how the core binary works, config loading, adapter invocation, fallback routing, and how it all fits together. Output: docs/architecture.md

## Summary of Changes\n\nWrote docs/architecture.md covering:\n- Overview diagram of the full system\n- Four components: CLI layer, config loader, router, adapters\n- Adapter invocation step-by-step\n- Both routing modes with flow diagrams (first + merge)\n- Data flow diagrams for success and partial failure cases\n- Project directory structure\n- Key design decisions with rationale
