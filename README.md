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

First launch shows the Logged-Out Screen without opening a browser. Press `Enter`
to authorize SpotUI with Spotify. The private Local Session is saved to
`~/.config/spotui/session.json` and reused on subsequent runs.

## Keybindings

| Key | Action |
|-----|--------|
| `1` | Playlists tab |
| `2` | Tracks tab |
| `3` | Artists tab |
| `4` | Search tab |
| `/` | Search tracks |
| `[` / `]` | Previous / next search page |
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
make build
./spotui
```

## Contributing / Development

```bash
make run
```

## Requirements

- Go 1.22+
- Spotify Premium account (required for playback control)
- A local Linux or macOS terminal for first-time browser Login

## License

SpotUI is licensed under GPL-3.0-only. See [LICENSE](LICENSE) and
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
