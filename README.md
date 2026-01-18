# lazyactions

A lazygit-style TUI for GitHub Actions. Monitor workflows, view logs, and manage runs from your terminal.

## Prerequisites

- Go 1.21 or later
- GitHub CLI (`gh`) installed and authenticated, OR a GitHub token

## Installation

### From Source

```bash
git clone https://github.com/nnnkkk7/lazyactions.git
cd lazyactions
make build
```

The binary will be created at `./bin/lazyactions`.

### Go Install

```bash
go install github.com/nnnkkk7/lazyactions/cmd/lazyactions@latest
```

## Authentication

lazyactions uses the following authentication methods (in order of priority):

1. **GitHub CLI (Recommended)**: If you have `gh` installed and authenticated, no configuration needed.
   ```bash
   gh auth login
   ```

2. **Environment Variable**: Set `GITHUB_TOKEN` with a personal access token.
   ```bash
   export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
   ```

   Required token scopes: `repo`, `workflow`

## Usage

Navigate to any directory within a GitHub repository and run:

```bash
lazyactions
```

Or specify a path:

```bash
lazyactions /path/to/repo
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Move to left pane |
| `l` / `→` | Move to right pane |
| `Tab` | Next pane |
| `Shift+Tab` | Previous pane |
| `Enter` | Select / Expand |
| `t` | Trigger workflow |
| `c` | Cancel run |
| `r` | Rerun workflow |
| `R` | Rerun failed jobs |
| `y` | Copy URL to clipboard |
| `/` | Filter |
| `Ctrl+r` | Refresh |
| `L` | View full log |
| `?` | Help |
| `q` | Quit |
| `Esc` | Back / Clear filter |

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make cover
```

### Lint

```bash
make lint
```

### Full CI Che

```bash
make ci
```

## Project Structure

```
lazyactions/
├── cmd/lazyactions/    # Entry point
├── app/                # TUI application (BubbleTea)
├── github/             # GitHub API client
├── auth/               # Authentication (SecureToken)
├── repo/               # Repository detection
└── docs/               # Design documents
```

## License

MIT
