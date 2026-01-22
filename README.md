<div align="center">

<img src="assets/logo.svg" alt="lazyactions logo" width="120" height="120">

**A lazygit-style TUI for GitHub Actions**



 [![CI](https://github.com/nnnkkk7/lazyactions/actions/workflows/ci.yaml/badge.svg)](https://github.com/nnnkkk7/lazyactions/actions/workflows/ci.yaml)
 [![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
 [![Go Report Card](https://goreportcard.com/badge/github.com/nnnkkk7/lazyactions)](https://goreportcard.com/report/github.com/nnnkkk7/lazyactions)
 [![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Monitor workflows, view logs, trigger runs, and manage GitHub Actions — all from your terminal.

![lazyactions demo](assets/demo.gif)

[Features](#features) • [Installation](#installation) • [Usage](#usage) • [Keybindings](#keybindings) • [Contributing](#contributing)

</div>

---


## Why lazyactions?

Inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [lazydocker](https://github.com/jesseduffield/lazydocker).

Tired of switching between your terminal and browser to check CI status? **lazyactions** brings GitHub Actions to your terminal with a familiar lazygit-style interface.

## Features

- **Browse & Monitor** — View workflows and runs with real-time status updates
- **View Logs** — Stream job logs directly in the terminal
- **Trigger Workflows** — Start `workflow_dispatch` workflows
- **Cancel & Rerun** — Stop running workflows or rerun failed jobs
- **Filter** — Quickly find workflows and runs with fuzzy search
- **Copy URLs** — Yank workflow/run URLs to clipboard
- **Keyboard & Mouse** — Vim-style keys and mouse support for navigation

## Installation

### Homebrew (Recommended)

```bash
brew install nnnkkk7/tap/lazyactions
```

### Using Go

Requires Go 1.24+

```bash
go install github.com/nnnkkk7/lazyactions/cmd/lazyactions@latest
```

### From Source

Requires Go 1.24+

```bash
git clone https://github.com/nnnkkk7/lazyactions.git
cd lazyactions
make build
# Binary: ./bin/lazyactions
```

## Authentication

**Option 1: GitHub CLI (Recommended)**
```bash
gh auth login
```

**Option 2: Personal Access Token**
```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
# Required scopes: repo, workflow
```

> **Note:** GitHub Enterprise is not currently supported. Only `github.com` repositories are supported at this time.

## Usage

```bash
# Run in any git repository
lazyactions

# Or specify a path
lazyactions /path/to/repo
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move between panels |
| `↑` / `↓` | Move up/down in list |
| `h` / `←` | Previous pane |
| `l` / `→` | Next pane |
| `Tab` / `Shift+Tab` | Cycle panes |
| `1` | Info tab |
| `2` | Logs tab |

### Actions

| Key | Action |
|-----|--------|
| `t` | Trigger workflow |
| `c` | Cancel run |
| `r` | Rerun workflow |
| `R` | Rerun failed jobs only |
| `y` | Copy URL to clipboard |

### General

| Key | Action |
|-----|--------|
| `/` | Filter mode |
| `Ctrl+r` | Refresh all data |
| `L` | Toggle fullscreen log |
| `?` | Show help |
| `Esc` | Back / Clear error |
| `q` | Quit |

### Mouse

| Action | Description |
|--------|-------------|
| **Click** | Select item / Switch pane |
| **Scroll** | Navigate lists and logs |

## Development

```bash
make build      # Build binary
make test       # Run all tests
make lint       # Run linter
make ci         # Full CI check
```


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.


## License

MIT License - see [LICENSE](LICENSE) for details.


## Support

If you find lazyactions useful, please consider giving it a star on GitHub!
It helps others discover the project and motivates continued development.

[![Star on GitHub](https://img.shields.io/github/stars/nnnkkk7/lazyactions?style=social)](https://github.com/nnnkkk7/lazyactions)
