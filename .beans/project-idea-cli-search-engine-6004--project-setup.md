---
# project-idea-cli-search-engine-6004
title: Project setup
status: completed
type: task
priority: normal
created_at: 2026-03-08T02:48:16Z
updated_at: 2026-03-08T02:50:20Z
---

Initialize the Go module, directory structure, and basic scaffolding for qry. Includes go.mod, cobra/viper wiring, and mise tasks.

## Summary of Changes\n\n- Initialized Go module: github.com/justestif/qry\n- Added cobra + viper dependencies\n- Scaffolded directory structure: cmd/, internal/config/, internal/router/, internal/result/\n- main.go entry point\n- cmd/root.go: CLI flags, config init, router wiring\n- internal/config/config.go: typed Config struct, Validate(), ResolvedAdapter()\n- internal/result/result.go: Result, AdapterError, FirstOutput, MergeOutput, FailureOutput, Deduplicate()\n- internal/router/invoke.go: subprocess invocation with timeout, stdin/stdout/stderr handling\n- internal/router/router.go: first + merge routing modes, AllAdaptersFailedError\n- mise.toml: build, run, test, tidy, lint tasks\n- Binary builds and --help works
