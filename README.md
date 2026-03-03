
![README](./images/img.png)

# Red Team TL;DR

[English](./README.md) | [中文简体](./README-zh.md)

[![GitHub release](https://img.shields.io/github/release/Rvn0xsy/red-tldr.svg)](https://github.com/Rvn0xsy/red-tldr/releases)

## What is Red Team TL;DR?

red-tldr is a lightweight red team command lookup tool. It helps experienced red team operators quickly find the commands, techniques, and key points they need — like a `man` command built for offensive security.

Starting from v0.5.0, red-tldr also ships with a built-in [MCP](https://modelcontextprotocol.io/) server, so AI assistants like Claude Desktop and Cursor can query your red team knowledge base directly.

## Why Red Team TL;DR?

In daily red team work, you need to memorize hundreds of commands — and usually you only remember the first few characters. Searching through documentation or search engines is slow and noisy. red-tldr gives you instant, offline, deterministic answers from a curated database you control.

## Features

- **Full-text search** — Bleve-powered BM25 search with weighted scoring (name×5, tags×3, data×1) and fuzzy matching
- **Multi-format output** — `text` (colored terminal), `json`, `markdown`
- **MCP Server** — Built-in Model Context Protocol server for AI assistant integration (stdio + Streamable HTTP)
- **MITRE ATT&CK mapping** — Entries can be tagged with tactics and techniques
- **Enhanced data model** — category, platforms, mitre_attack, metadata (author/source/confidence), related entries
- **Backward compatible** — Works with existing YAML databases, new fields are optional
- **Offline-first** — No cloud dependencies, no API keys, fully deterministic
- **Secure** — Zip-slip vulnerability fixed, path validation on all archive operations

## Quick Start

### Install

#### macOS

```bash
$ brew install red-tldr
```

#### Arch Linux

```bash
# AUR: https://aur.archlinux.org/packages/red-tldr
$ sudo pacman -S red-tldr
```

#### From Source

```bash
$ git clone https://github.com/Rvn0xsy/red-tldr
$ cd red-tldr
$ go build
```

#### Binary

Download from [Releases](https://github.com/Rvn0xsy/red-tldr/releases/).

```bash
$ tar -zxvf red-tldr_latest_linux_amd64.tar.gz
$ ./red-tldr
```

> It is recommended to add red-tldr to your PATH.

## Usage

### Search

```bash
# Keyword search
$ red-tldr mimikatz
```

![search-mimikatz](./images/img_1.png)

```bash
# Fuzzy matching — type a few characters
$ red-tldr mi
```

![Fuzzy-match](./images/img_2.png)

When multiple results are found, select by number:

![Select-Number](./images/img_3.png)

### Output Formats

```bash
# JSON output
$ red-tldr mimikatz -f json

# Markdown output
$ red-tldr mimikatz -f markdown
```

### Update & Upgrade

```bash
# Rebuild local index (JSON + Bleve)
$ red-tldr update

# Download latest database from GitHub
$ red-tldr upgrade
```

## MCP Server

red-tldr includes a built-in MCP server that exposes your red team knowledge base to AI assistants.

### Tools

| Tool | Description |
|------|-------------|
| `search_redteam_commands` | Search commands by keyword, platform, category, tactic, or technique |
| `get_command_details` | Get full details of a specific command entry |
| `list_techniques` | List available MITRE ATT&CK techniques in the database |

### stdio Mode (Claude Desktop / Cursor)

```bash
$ red-tldr serve
```

Add to your Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "red-tldr": {
      "command": "red-tldr",
      "args": ["serve"]
    }
  }
}
```

### Streamable HTTP Mode

```bash
$ red-tldr serve --http --addr localhost:8080 --endpoint /mcp
```

## Configuration

Default config path: `~/.red-tldr/config.toml`

```toml
[red-tldr]
  index-update = false
  github-update = false
  path = ""
  color = true
```

| Item | Description | Type |
|------|-------------|------|
| index-update | Auto-rebuild index on search | Bool |
| github-update | Auto-download database from GitHub | Bool |
| path | Database storage path | String |
| color | Colored terminal output | Bool |

## Data Model

Each entry is a YAML file in the [red-tldr-db](https://github.com/Rvn0xsy/red-tldr-db) repository. The enhanced format (v0.5.0+) supports additional fields while remaining backward compatible:

```yaml
name: mimikatz-sekurlsa
tags:
  - mimikatz
  - credentials
  - lsass
data: |
  # Mimikatz Sekurlsa
  ```
  privilege::debug
  sekurlsa::logonpasswords
  ```
category: credential-access
platforms:
  - windows
mitre_attack:
  tactics:
    - credential-access
  techniques:
    - T1003.001
metadata:
  author: Rvn0xsy
  confidence: high
related:
  - mimikatz-dpapi
```

All new fields (`category`, `platforms`, `mitre_attack`, `metadata`, `related`) are optional — existing YAML files work without modification.

## Contributing

We welcome contributions!

* Report bugs via [Issues](https://github.com/Rvn0xsy/red-tldr/issues).
* Submit bug fixes as Pull Requests to the **dev** branch.
* For new features, create an issue first to discuss the approach before submitting a PR.
* Documentation improvements are always welcome.
* Contact: rvn0xsy@gmail.com

## Stargazers over time

[![Stargazers over time](https://starchart.cc/Rvn0xsy/red-tldr.svg)](https://starchart.cc/Rvn0xsy/red-tldr)
