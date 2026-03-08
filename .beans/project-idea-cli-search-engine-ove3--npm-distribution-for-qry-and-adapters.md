---
# project-idea-cli-search-engine-ove3
title: npm distribution for qry and adapters
status: completed
type: feature
priority: normal
created_at: 2026-03-08T16:44:46Z
updated_at: 2026-03-08T16:47:59Z
---

Publish @justestif/qry and all adapters to npm using the platform optional-deps pattern (same as esbuild/biome). Each binary gets a main package with a JS wrapper bin and platform-specific optional packages containing the actual Go binary.

## Packages to create
- @justestif/qry + platform packages (linux-x64, linux-arm64, darwin-x64, darwin-arm64, win32-x64)
- @justestif/qry-adapter-brave-api + platform packages
- @justestif/qry-adapter-brave-scrape + platform packages
- @justestif/qry-adapter-ddg-scrape + platform packages
- @justestif/qry-adapter-exa + platform packages

## Todo
- [x] Create npm/ directory structure with all package.json files
- [x] Create bin.js wrapper scripts for each main package
- [x] Update release.yml to publish all npm packages after GitHub release
- [x] Document npm install in README

## Summary of Changes

- Created `npm/` directory with 30 packages total (5 binaries × 5 platforms + 5 main wrappers)
- Each main package (`@justestif/qry`, `@justestif/qry-adapter-*`) has a `bin.js` JS wrapper that resolves the right platform package at runtime
- Platform packages (`@justestif/qry-linux-x64` etc.) have `os`/`cpu` fields so npm auto-selects the right one
- Updated `release.yml` with a `publish-npm` job that: extracts binaries from the existing build artifacts, stamps versions from the git tag, publishes platform packages first, then main packages
- Updated README with `npm install -g @justestif/qry` as the recommended install method
- Requires `NPM_TOKEN` secret to be set in the GitHub repo settings

## Revision: Simplified to postinstall pattern

Scrapped the platform-optional-packages approach (too many files). Each package now has:
- `package.json` with a `postinstall` script
- `install.js` downloads the right binary from GitHub Releases at install time
- `bin.js` wrapper that execs the downloaded binary

5 packages, 15 files total. CI just stamps versions and publishes — no binary copying needed.
