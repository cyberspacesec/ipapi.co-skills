# ipapi.co-skills

[![CLI Release](https://img.shields.io/github/v/release/cyberspacesec/ipapi.co-skills)](https://github.com/cyberspacesec/ipapi.co-skills/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/cyberspacesec/ipapi.co-skills.svg)](https://pkg.go.dev/github.com/cyberspacesec/ipapi.co-skills)
[![CI](https://github.com/cyberspacesec/ipapi.co-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/cyberspacesec/ipapi.co-skills/actions/workflows/ci.yml)
[![GitHub License](https://img.shields.io/github/license/cyberspacesec/ipapi.co-skills)](https://github.com/cyberspacesec/ipapi.co-skills/blob/main/LICENSE)
[![Docs](https://img.shields.io/badge/docs-vitepress-3c8c4a)](https://cyberspacesec.github.io/ipapi.co-skills/)

**A command-line IP geolocation tool** that wraps the [ipapi.co](https://ipapi.co) API. Default JSON output, stable exit codes, AI-Agent friendly — backed by a zero-dependency Go SDK.

📖 **Full documentation**: [https://cyberspacesec.github.io/ipapi.co-skills/](https://cyberspacesec.github.io/ipapi.co-skills/) · 📄 **中文版**: [README_CN.md](./README_CN.md)

## ✨ Why use it

- 🤖 **Agent friendly**: default JSON envelope `{ok, command, data, meta}`, errors identified by stable `code` + exit code — Agents can call it without reading docs
- 🚪 **Stable exit codes**: `0` success, `3-12` map to 10 business error categories, `6/8/9` mark retryable errors
- 🎯 **Field-level queries**: `ipapi field 8.8.8.8 country` fetches a single field in one line; `--human` outputs a plain value for easy piping
- 📡 **Multiple formats**: JSON / JSONP / XML / CSV / YAML; `raw` command returns original bytes
- 🧭 **Self-describing**: `ipapi fields` lists all 28 queryable fields, grouped
- 🧩 **Also a Go SDK**: the CLI is just a thin shell; `pkg/ipapi` is a zero-runtime-dependency Go library you can embed directly

## 📥 Installation

```bash
go install github.com/cyberspacesec/ipapi.co-skills/cmd/ipapi@latest
```

Or download a precompiled binary from [Releases](https://github.com/cyberspacesec/ipapi.co-skills/releases).

## ⚡ 30-second quickstart

```bash
ipapi 8.8.8.8                    # Query a specific IP (default JSON envelope)
ipapi 8.8.8.8 -H                 # Human-readable table
ipapi field 8.8.8.8 country      # Fetch a single field
ipapi field 8.8.8.8 country -H   # Plain value output, easy to pipe
ipapi me                          # Look up this machine's public IP
ipapi fields                      # List all queryable fields
ipapi raw 8.8.8.8 -f yaml         # Raw YAML response
```

JSON output example:

```json
{
  "ok": true,
  "command": "info",
  "args": {"ip": "8.8.8.8", "format": "json"},
  "data": {
    "ip": "8.8.8.8",
    "city": "Mountain View",
    "country_name": "United States",
    "country_code": "US",
    "latitude": 37.42301,
    "longitude": -122.083352,
    "timezone": "America/Los_Angeles",
    "asn": "AS15169",
    "org": "Google LLC"
  },
  "meta": {"format": "json", "durationMs": 312, "retrievedAt": "2026-07-04T10:01:22Z"}
}
```

On errors the exit code distinguishes categories, making Agent branching easy:

```bash
ipapi info 999.1.1.1; echo "exit=$?"   # exit=3 (INVALID_IP)
```

## 🤖 AI Agent integration

The Agent calling pattern: first `ipapi fields` to discover queryable fields, then `ipapi info/field` to query, and rely on the exit code plus `error.code` to interpret results. See [SKILLS.md](./SKILLS.md) and the [Agent integration guide](https://cyberspacesec.github.io/ipapi.co-skills/cli/agent).

```bash
# Extract the country name
ipapi info 8.8.8.8 | jq -r '.data.country_name'

# Error handling
if ! ipapi info "$IP" > /tmp/info.json 2>/tmp/err.json; then
  code=$(jq -r '.error.code' /tmp/err.json)
  case "$code" in
    RATE_LIMITED) sleep 60 && retry ;;
    INVALID_IP)   echo "bad ip" ;;
  esac
fi
```

## 📋 Command overview

| Command | Description |
|---|---|
| `ipapi info <ip>` | Full info for a given IP |
| `ipapi me` | Full info for this machine's public IP |
| `ipapi field <ip> <field>` | A single field for a given IP |
| `ipapi me-field <field>` | A single field for this machine's IP |
| `ipapi raw <ip> -f <fmt>` | Raw format (xml/csv/yaml/jsonp) for a given IP |
| `ipapi me-raw -f <fmt>` | Raw format for this machine's IP |
| `ipapi fields` | List all queryable fields (local, no network) |
| `ipapi version` | Version info |
| `ipapi completion <shell>` | Generate shell completions |

Global flags: `--api-key`, `--api-key-mode`, `-f/--format`, `--base-url`, `--retries`, `--timeout`, `-H/--human`, `--config`, `--callback`. Config priority: flag > env var > `~/.ipapi.json` > defaults.

## 🧩 Using it as a Go SDK

The CLI is a thin shell over `pkg/ipapi`. To embed it in a Go program:

```bash
go get github.com/cyberspacesec/ipapi.co-skills
```

```go
package main

import (
	"context"
	"fmt"
	"github.com/cyberspacesec/ipapi.co-skills/pkg/ipapi"
	"time"
)

func main() {
	client := ipapi.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, _ := client.GetIPInfo(ctx, "8.8.8.8", string(ipapi.FormatJSON))
	fmt.Printf("Location: %s, %s\n", info.City, info.CountryName)
}
```

SDK features: 10 functional options (`WithAPIKey` / `WithCustomHTTPClient` / `WithErrorHandler` / `WithCallback` / `WithBaseURL` / `WithUserAgent` / `WithRetries` / `WithTimeout` / `WithRateLimiter`, etc.), 5 response formats, 10 sentinel errors (`errors.Is` matching), automatic retry and rate limiting, full IPv4/IPv6 support, 100% test coverage.

- 📖 [Library guide](https://cyberspacesec.github.io/ipapi.co-skills/guide/intro)
- 📚 [API reference](https://cyberspacesec.github.io/ipapi.co-skills/api/methods)
- 🐙 [Go docs](https://pkg.go.dev/github.com/cyberspacesec/ipapi.co-skills)

## 🏗 Build

```bash
# CLI
go build ./cmd/ipapi/

# Tests
go test ./...
```

Cross-platform binary releases are automated by GitHub Actions + GoReleaser (triggered by pushing a `v*` tag).

## 🤝 Contributing

Issues and PRs are welcome. Please ensure:
- Code follows Go conventions and passes `go vet`
- New features include tests
- Documentation stays in sync

## 📄 License

MIT, see [LICENSE](LICENSE).

---

> 📌 Use of the API is subject to ipapi.co's terms of service; for production, applying for an API Key is recommended.
