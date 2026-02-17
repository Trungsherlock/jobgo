# JobGo

A Go CLI that crawls jobs from your target companies, matches them against your profile, and notifies you in real-time.

## Features

- **Multi-platform scraping** — Fetches jobs from Lever and Greenhouse career pages
- **Concurrent worker pool** — Scrapes multiple companies in parallel with configurable concurrency
- **Smart matching** — Keyword-based scoring with optional LLM-powered semantic matching via Claude API
- **Watch mode** — Background polling with desktop, terminal, and webhook notifications
- **Application tracking** — Track your pipeline from applied to offer
- **REST API** — Serve job data over HTTP for integrations
- **MCP server** — Expose tools for AI agents via Model Context Protocol
- **JSON output** — Pipe any list command to `jq` with `-o json`

## Architecture

```
cmd/jobgo/              CLI entrypoint
internal/
  cli/                  Cobra command definitions
  database/             SQLite + migration runner + repositories
  scraper/              Scraper interface + Lever/Greenhouse adapters
  matcher/              Matcher interface + keyword/LLM/hybrid implementations
  worker/               Goroutine worker pool with channel-based job queue
  notifier/             Notifier interface + terminal/desktop/webhook
  server/               REST API (chi) + MCP server (stdio/SSE)
migrations/             Versioned SQL migrations
```

## Quick Start

### Install

```bash
go install github.com/Trungsherlock/jobgocli/cmd/jobgo@latest
```

Or build from source:

```bash
git clone https://github.com/Trungsherlock/jobgocli.git
cd jobgocli
make build
```

### Set up your profile

```bash
jobgo profile set \
  --name "Your Name" \
  --skills "Go,PostgreSQL,Docker,Kubernetes" \
  --roles "backend engineer,SRE" \
  --locations "remote,San Francisco" \
  --experience 3

jobgo profile show
```

### Add companies to track

```bash
jobgo company add --name "Spotify" --platform lever --slug spotify
jobgo company add --name "Airbnb" --platform greenhouse --slug airbnb
jobgo company list
```

### Search for jobs

```bash
jobgo search
```

This scrapes all tracked companies, stores new jobs, and scores them against your profile.

### Browse results

```bash
# List all jobs sorted by match score
jobgo jobs list

# Filter by score, remote, or new only
jobgo jobs list --min-match 50 --remote --new

# View full job details
jobgo jobs show <job-id>

# Open in browser
jobgo jobs open <job-id>

# JSON output
jobgo jobs list -o json
```

### Track applications

```bash
jobgo apply <job-id> --notes "Applied via website"
jobgo jobs update <job-id> --status interview --notes "Scheduled Friday"
jobgo status
```

### Watch mode

```bash
jobgo watch --interval 30m --min-score 50
```

Runs in the foreground, scraping on a loop and notifying you of new high-match jobs.

## Configuration

Config file at `~/.jobgo/config.yaml`:

```yaml
# Matcher: keyword, llm, or hybrid
matcher: keyword

# Anthropic API key for LLM/hybrid matching
anthropic_api_key: sk-ant-...

# Keyword score threshold for hybrid mode
hybrid_threshold: 30.0

# Notification channels
notify:
  - desktop
  # - webhook

# webhook_url: https://hooks.slack.com/services/...
```

Or use environment variables:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

## API Server

```bash
# REST API
jobgo serve --port 8080

# MCP server (stdio, for Claude Code)
jobgo serve --mcp

# MCP server (SSE, for remote clients)
jobgo serve --mcp-sse --port 9090
```

### REST Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/jobs | List jobs (query: min_score, company_id, new, remote) |
| GET | /api/jobs/:id | Job details |
| GET | /api/companies | List companies |
| POST | /api/companies | Add a company |
| DELETE | /api/companies/:id | Remove a company |
| GET | /api/profile | Current profile |
| GET | /api/stats | Application pipeline summary |

### MCP Tools

| Tool | Description |
|------|-------------|
| search_jobs | Search jobs with filters |
| get_job_details | Full job description + match info |
| list_companies | Tracked companies |
| get_profile | User profile |
| get_stats | Application pipeline stats |

## Supported Platforms

| Platform | API | Auth |
|----------|-----|------|
| Lever | `api.lever.co/v0/postings/{slug}` | None (public) |
| Greenhouse | `boards.greenhouse.io/v1/boards/{slug}/jobs` | None (public) |

## Development

```bash
# Run tests
make test

# Run tests (skip integration)
go test ./... -short

# Lint
make lint

# Build
make build
```

## Tech Stack

- **Go** — CLI, concurrency, HTTP server
- **SQLite** (modernc.org/sqlite) — Zero-config embedded database
- **Cobra/Viper** — CLI framework + config management
- **Chi** — Lightweight HTTP router
- **mcp-go** — Model Context Protocol server SDK
- **Claude API** — LLM-powered job matching
