# SpotUI

Spotify TUI client written in Go. Controls Spotify playback from the terminal.

> **Requires a Spotify Premium account.** Free accounts can browse the library but cannot start or control playback.

## Install

Download the latest binary for your platform from the [releases page](https://github.com/dcbto/spotui/releases):

| Platform | File |
|----------|------|
| Linux (x86_64) | `spotui-linux-amd64` |
| macOS (Intel) | `spotui-darwin-amd64` |
| macOS (Apple Silicon) | `spotui-darwin-arm64` |
| Windows (x86_64) | `spotui-windows-amd64.exe` |

```bash
# Linux / macOS
chmod +x spotui-linux-amd64
./spotui-linux-amd64
```

Windows: run `spotui-windows-amd64.exe` directly.

## First Run

On first launch a browser window opens for Spotify login. After authorizing, the token is saved to `~/.config/spotui/token.json` and reused on subsequent runs.

## Keybindings

| Key | Action |
|-----|--------|
| `1` | Playlists tab |
| `2` | Tracks tab |
| `3` | Artists tab |
| `4` | Search tab |
| `/` | Search tracks |
| `esc` | Cancel search input |
| `j` / `↓` | Cursor down |
| `k` / `↑` | Cursor up |
| `enter` | Play selected |
| `space` | Play / Pause |
| `n` | Next track |
| `p` | Previous track |
| `s` | Toggle shuffle |
| `c` | Open Spotify account settings |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

## Build from Source

```bash
# Self-contained binary (ID baked in — no env var needed at runtime)
SPOTIFY_CLIENT_ID=<your_id> make build
./spotui

# Binary without baked ID (requires SPOTIFY_CLIENT_ID env var at runtime)
make build
SPOTIFY_CLIENT_ID=<your_id> ./spotui
```

## Contributing / Development

1. Go to https://developer.spotify.com/dashboard and create a new app
2. Add `http://localhost:8080/callback` as a Redirect URI
3. Copy the **Client ID** (no client secret needed)

```bash
export SPOTIFY_CLIENT_ID=your_client_id_here
make run
```

## Requirements

- Go 1.22+
- Spotify Premium account (required for playback control)
- An active Spotify session on any device before using playback controls
