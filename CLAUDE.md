# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Espace-Tech Cloud CLI (`espacetech`) — a Go CLI for deploying and managing apps on Espace-Tech Cloud. Built with Cobra for commands and promptui for interactive prompts. The Go module name is `paas-cli`.

## Build & Run

```bash
make build              # builds ./espacetech binary with version/commit/date ldflags
make release-all        # cross-compile for linux/darwin/windows into dist/
make clean              # remove dist/
go build -o espacetech . # quick build without ldflags
```

There are no tests in this project currently.

## Architecture

**Entry point:** `main.go` calls `cmd.Execute()`.

**`cmd/`** — each file registers one top-level Cobra command (or subcommand group) onto `rootCmd`:
- `root.go` — root command, version command, ldflags vars (`version`, `commit`, `date`)
- `deploy.go` — tarball upload + deployment polling loop; has monorepo detection via `turbo.json`
- `init.go` — interactive project init, writes `.espacetech.json`
- `site.go` — `site add|list|use` subcommands
- `db.go` — `db create|list|info|credentials|link|unlink|expose|unexpose|stop|start|rotate|delete`
- `storage.go` — `storage create|list|info|credentials|link|unlink|expose|unexpose|rotate|delete`
- `auth.go` — `auth create|list|info|config|users|stats|rotate-keys|delete`
- `domain.go`, `env.go`, `logs.go`, `rollback.go`, `status.go`, `delete.go`, `login.go`, `logout.go`, `register.go`

**`internal/api/client.go`** — single API client struct wrapping `net/http`. All backend calls go through `authRequest()` which adds the Bearer token. The client handles projects, deployments, domains, env vars, databases, storage buckets, and auth apps.

**`internal/api/tar.go`** — creates gzip tarballs for deploy uploads, skipping `node_modules`, `.next`, `.git`, `.turbo`, `dist`.

**`internal/config/config.go`** — reads/writes `~/.paas-cli.json` (token, api_host, user_id, email). Default API host: `https://api.espace-tech.com`.

## Key Patterns

- Version info is injected at build time via `-ldflags` (see `Makefile` `LDFLAGS`)
- Project config lives in `.espacetech.json` in the project directory (not the CLI config)
- Deploy detects monorepos by walking up to find `turbo.json`, then scans for `.espacetech.json` files
- Env var operations auto-detect single-site projects and use site-scoped endpoints; multi-site projects require `site_id` in `.espacetech.json`
- Releases are triggered by pushing a `v*` tag (see `.github/workflows/release.yml`)
