<div align="center">

# lazyactions

**A lazygit-style TUI for GitHub Actions**



 [![CI](https://github.com/nnnkkk7/lazyactions/actions/workflows/ci.yaml/badge.svg)](https://github.com/nnnkkk7/lazyactions/actions/workflows/ci.yaml)
 [![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
 [![Go Report Card](https://goreportcard.com/badge/github.com/nnnkkk7/lazyactions)](https://goreportcard.com/report/github.com/nnnkkk7/lazyactions)
 [![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Monitor workflows, view logs, trigger runs, and manage GitHub Actions — all from your terminal.

![lazyactions demo](assets/demo.gif)

[Features](#features) • [Installation](#installation) • [Usage](#usage) • [Keybindings](#keybindings) • [Contributing](#contributing)

</div>

---

![lazyactions demo](demo.gif)

## Why lazyactions?

Inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [lazydocker](https://github.com/jesseduffield/lazydocker).

Tired of switching between your terminal and browser to check CI status? **lazyactions** brings GitHub Actions to your terminal with a familiar lazygit-style interface.

- **Stay in flow** — No context switching to the browser
- **Keyboard-first** — Vim-style navigation you already know
- **Mouse support** — Click and scroll for quick navigation
- **Real-time** — Adaptive polling keeps you updated
- **Full control** — Trigger, cancel, and rerun workflows without leaving the terminal

## Features

| Feature | Description |
|---------|-------------|
| **Browse Workflows** | View all workflows in your repository |
| **Monitor Runs** | See run status with real-time updates |
| **View Logs** | Stream job logs directly in the terminal |
| **Trigger Workflows** | Start `workflow_dispatch` workflows |
| **Cancel Runs** | Stop running workflows instantly |
| **Rerun Workflows** | Rerun entire workflows or just failed jobs |
| **Filter** | Quickly find workflows and runs with fuzzy search |
| **Copy URLs** | Yank workflow/run URLs to clipboard |
| **Mouse Support** | Click and scroll for quick navigation |

## Installation

### Homebrew

```bash
brew install nnnkkk7/tap/lazyactions
```

### Using Go

```bash
go install github.com/nnnkkk7/lazyactions/cmd/lazyactions@latest
```

### From Source

```bash
git clone https://github.com/nnnkkk7/lazyactions.git
cd lazyactions
make build
# Binary: ./bin/lazyactions
```

### Prerequisites

- Go 1.21+
- GitHub CLI (`gh`) authenticated, **OR** `GITHUB_TOKEN` environment variable

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

---

<div align="center">

**[Star this repo](https://github.com/nnnkkk7/lazyactions)** if you find it useful!

</div>
