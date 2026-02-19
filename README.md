# JobGo

A Go CLI that crawls jobs from your target companies, matches them against your profile, and notifies you in real-time — with built-in H1B visa sponsorship tracking for international students and new grads.

## Features

- **Multi-platform scraping** — Fetches jobs from Lever and Greenhouse career pages
- **Concurrent worker pool** — Scrapes multiple companies in parallel with configurable concurrency
- **Smart matching** — Keyword-based scoring with optional LLM-powered semantic matching via Claude API
- **H1B sponsorship tracking** — Import USCIS employer data, link companies, and boost/penalize jobs based on visa stance
- **Job classification** — Auto-detects experience level (intern/entry/mid/senior/staff), new grad roles, and visa sentiment
- **H1B-aware scoring** — Adjusts match scores based on sponsorship history, visa sentiment, and new grad fit
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
  h1b/                  H1B importer, classifier, and scorer
migrations/             Versioned SQL migrations
data/                   CSV data files (companies.csv, h1b_employers.csv)
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
  --experience 1 \
  --visa   # set if you need H1B sponsorship

jobgo profile show
```

### Add companies to track

Add one at a time:

```bash
jobgo company add --name "Stripe" --platform lever --slug stripe
jobgo company add --name "Airbnb" --platform greenhouse --slug airbnb
```

Or bulk import from CSV:

```bash
jobgo company import data/companies.csv
jobgo company list
```

The CSV format is `name,platform,slug` — edit [data/companies.csv](data/companies.csv) to add your targets.

### Import H1B sponsorship data

Download the USCIS H1B Employer Data Hub CSV from [uscis.gov](https://www.uscis.gov/tools/reports-and-studies/h-1b-employer-data-hub) and run:

```bash
jobgo h1b import data/h1b_employers.csv
jobgo h1b status
```

This imports ~24,000+ employer records and auto-links your tracked companies to their sponsorship history.

### Search for jobs

```bash
jobgo search
```

This scrapes all tracked companies, stores new jobs, scores them against your profile, classifies them (experience level, visa stance), and applies H1B adjustments if `--visa` is set.

### Browse results

```bash
# All jobs sorted by match score
jobgo jobs list

# Filter by score, remote, or new only
jobgo jobs list --min-match 50 --remote --new

# Only visa-friendly jobs (H1B sponsors, no negative visa sentiment)
jobgo jobs list --visa-friendly

# Only new grad roles
jobgo jobs list --new-grad

# Combine filters
jobgo jobs list --visa-friendly --new-grad --remote

# View full job details (includes H1B and classification info)
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

## H1B Workflow

For international students and new grads needing sponsorship:

```bash
# 1. Mark visa required in profile
jobgo profile set --visa

# 2. Import USCIS data
jobgo h1b import data/h1b_employers.csv

# 3. Scrape + score + classify + adjust
jobgo search

# 4. View visa-friendly new grad jobs
jobgo jobs list --visa-friendly --new-grad

# 5. Check H1B status of tracked companies
jobgo h1b status
```

**Score adjustments applied automatically when `visa_required = true`:**

| Signal | Score Delta |
|--------|------------|
| Company H1B sponsor, ≥90% approval | +15 |
| Company H1B sponsor, ≥70% approval | +10 |
| Company H1B sponsor, <70% approval | +5 |
| Job mentions visa sponsorship positively | +10 |
| Job says no visa sponsorship | −20 |
| New grad role + ≤2 years experience | +5 |

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
| GET | /api/jobs | List jobs (query: min_score, company_id, new, remote, visa_friendly, new_grad) |
| GET | /api/jobs/:id | Job details (includes H1B + classification fields) |
| GET | /api/companies | List companies |
| POST | /api/companies | Add a company |
| DELETE | /api/companies/:id | Remove a company |
| GET | /api/profile | Current profile |
| GET | /api/stats | Application pipeline summary |
| GET | /api/h1b/sponsors | H1B sponsorship status for tracked companies |
| GET | /api/h1b/status | Total H1B records in database |

### MCP Tools

| Tool | Description |
|------|-------------|
| search_jobs | Search jobs with filters (visa_friendly, new_grad, remote, min_score) |
| get_job_details | Full job description + match info + H1B/classification data |
| list_companies | Tracked companies with H1B sponsorship info |
| get_profile | User profile including visa requirement |
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
- **USCIS H1B Employer Data Hub** — Visa sponsorship history data
