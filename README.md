# SpotUI

SpotUI is a zero configuration Spotify terminal client with local audio
playback. It runs as one process: no playback daemon, sidecar, developer
dashboard setup, or official Spotify Web API client is required.

> **Spotify Premium is required.** Free accounts are stopped after Login and
> can only log out or quit.

SpotUI uses an unofficial protocol through a pinned `go-librespot` fork. It can
break when Spotify changes its service and is not affiliated with or endorsed
by Spotify.

## Install

Download the archive for your platform from the
[releases page](https://github.com/dcbto/spotui/releases):

| Platform | Archive |
|----------|---------|
| Linux x86_64 | `spotui-linux-amd64.tar.gz` |
| macOS Intel | `spotui-darwin-amd64.tar.gz` |
| macOS Apple Silicon | `spotui-darwin-arm64.tar.gz` |

Windows is not supported and no Windows artifact is published. Each archive
contains the `spotui` binary, `LICENSE`, and `THIRD_PARTY_NOTICES.md`.

```bash
tar -xzf spotui-linux-amd64.tar.gz
cd spotui-linux-amd64
./spotui
```

Audio is sent to the operating system Default Audio Output. On Linux, SpotUI
uses PulseAudio when `PULSE_SERVER` or the WSLg PulseAudio socket is available,
and otherwise uses ALSA. On macOS, it uses AudioToolbox. All backends are
opened by go-librespot.

## Login

The first launch stays on the Logged-Out screen. Press `Enter` to start Login
and complete authorization in the browser.

First Login must run in a local terminal on Linux or macOS. Remote SSH and
headless first-login flows are not supported. If the browser cannot open,
SpotUI shows the authorization URL so it can be copied without losing the
Login attempt.

The private Local Session is stored at `~/.config/spotui/session.json` with
owner-only permissions and reused on later runs. `L` logs out and deletes it.
SpotUI does not resume the previously playing track after startup or reconnect.

## Controls

| Key | Action |
|-----|--------|
| `/` | Search tracks |
| `[` / `]` | Previous / next search page |
| `esc` | Cancel search input |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `enter` | Play selected track |
| `space` | Play / pause |
| `n` | Next track |
| `p` | Previous track |
| `-` / `+` | Lower / raise volume |
| `a` | Toggle Autoplay |
| `L` | Log out |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

SpotUI also appears as a Spotify Connect device. Playback transferred to
another device remains there; SpotUI does not steal it back automatically.
Transient engine failures enter a visible reconnect state with bounded backoff.

## Build from source

Requirements:

- Go version declared in `go.mod`
- Spotify Premium account for manual playback validation
- Linux: ALSA and FLAC development packages plus `pkg-config`
- macOS: Xcode command-line tools, FLAC, and `pkg-config`

```bash
go mod verify
make build
./spotui
```

Dependencies and the modified `go-librespot` fork revision are pinned by
`go.mod` and `go.sum`. Automated tests use fakes and require no Spotify
credentials, browser, network, or audio device:

```bash
go test -race ./...
```

## License

SpotUI is licensed under GPL-3.0-only. See [LICENSE](LICENSE) and
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
