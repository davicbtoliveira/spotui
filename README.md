# SpotUI

SpotUI is a zero configuration Spotify client for the terminal. It combines
collection browsing, grouped search, Spotify Connect, and local audio playback
in one process: no playback daemon, sidecar, developer dashboard, or official
Spotify Web API application is required.

> **Spotify Premium is required.** Free accounts are stopped after Login and
> can only log out or quit.

SpotUI uses an unofficial protocol through a pinned `go-librespot` fork.
Spotify service changes can therefore break the client. SpotUI is not affiliated
with or endorsed by Spotify.

## Features

- Browse Liked Tracks, User Playlists, and Saved Albums.
- Open playlist, album, and artist details and play tracks in their context.
- Explore recommendations built from top artists and tracks.
- Search tracks, albums, artists, and playlists with pagination.
- Play locally through the operating system's Default Audio Output.
- Control play/pause, next/previous, volume, Autoplay, Shuffle, and seeking.
- Appear as a Spotify Connect device and keep playback local after a transfer.
- Reconnect after transient engine failures without asking for Login again.

## Support and requirements

| Target | Release artifact | Notes |
| --- | --- | --- |
| Linux x86_64 | `spotui-linux-amd64.tar.gz` | PulseAudio or ALSA |
| macOS Intel | `spotui-darwin-amd64.tar.gz` | AudioToolbox |
| macOS Apple Silicon | `spotui-darwin-arm64.tar.gz` | AudioToolbox |

Windows is not supported and no Windows artifact is published. The first
Login must run in a local terminal on Linux or macOS; remote SSH and headless
first-login flows are not supported.

## Install a release

Download the archive for your platform from the
[releases page](https://github.com/dcbto/spotui/releases). Releases also
provide `SHA256SUMS` for verification. Each archive contains the `spotui`
binary, `LICENSE`, and `THIRD_PARTY_NOTICES.md`.

```bash
tar -xzf spotui-linux-amd64.tar.gz
cd spotui-linux-amd64
./spotui
```

Replace the archive and directory names with the macOS artifact when needed.
On first launch, press `Enter` and finish authorization in the browser. If the
browser does not open, SpotUI displays the authorization URL so it can be
copied manually.

## Build from source

The required Go version is declared in `go.mod`. A source build also needs the
native audio dependencies for the target platform.

### Linux (Debian or Ubuntu)

```bash
sudo apt-get update
sudo apt-get install -y libasound2-dev libflac-dev pkg-config
```

Use the equivalent ALSA, FLAC, and `pkg-config` development packages on other
Linux distributions.

### macOS

Install the Xcode command-line tools and, if needed, Homebrew:

```bash
xcode-select --install
brew install flac pkg-config
```

Then build and run:

```bash
go mod verify
go test -race ./...
make build
./spotui
```

`make run` runs directly with `go run`; `make clean` removes the local binary.
Automated tests use fakes and require no Spotify credentials, browser, network,
or audio device.

## Login and local state

SpotUI starts on the Logged-Out screen. Press `Enter` to begin Login. After a
successful authorization, a private reusable Local Session is stored on disk;
later launches reuse it without opening the browser. SpotUI does not resume
the previously playing track after startup or reconnect.

The directory is resolved with Go's `os.UserConfigDir`:

| System | SpotUI directory |
| --- | --- |
| Linux | `$XDG_CONFIG_HOME/spotui`, or `~/.config/spotui` when the variable is unset |
| macOS | `~/Library/Application Support/spotui` |

Files in that directory are:

| File | Purpose |
| --- | --- |
| `session.json` | Reusable local session state. Stored with owner-only permissions and removed by `L`. |
| `settings.json` | Local preferences, currently the Autoplay setting. |

The session file contains reusable authentication state. Do not share it or
commit it to a repository. SpotUI does not require a Spotify developer app or
client credentials.

## Browse and play

After Login, the browse shell opens on **Liked Tracks**. On a wide terminal the
navigation panel is on the left; on a smaller terminal, press `Tab` to switch
between navigation and content focus.

- **Library**: Liked Tracks, Playlists, and Saved Albums.
- **Recommended**: top artists and tracks, plus albums and playlists from top
  artists.
- **Search**: grouped results for tracks, albums, artists, and playlists.

Press `Enter` on a track to play it. Press `Enter` on an album, artist, or
playlist to open its details; `Esc` or `Backspace` returns to the previous
view. `[` and `]` move between pages. Artwork is loaded lazily for the selected
item when artwork metadata is available.

## Keyboard controls

| Key | Action |
| --- | --- |
| `Enter` | Start Login, open the selected catalog item, or play the selected track |
| `/` | Open grouped Search |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Tab` | Switch navigation/content focus |
| `[` / `]` | Previous / next page |
| `r` | Retry a failed catalog request |
| `o` | Open the selected Spotify link in the browser |
| `Space` | Play / pause |
| `n` | Next track |
| `p` | Previous track |
| `-` / `+` | Lower / raise volume by 5% |
| `a` | Enable or disable Autoplay |
| `s` | Enable or disable Shuffle for the current playback context |
| `h` / `l` | Seek backward / forward 10 seconds |
| `Esc` / `Backspace` | Cancel search input or go back from details |
| `L` | Log out; press `y` to confirm |
| `?` | Show the keyboard help overlay |
| `q` / `Ctrl+C` | Quit |

`Autoplay` is saved in `settings.json`. `Shuffle` is available when the
selected track belongs to a playback context such as an album or playlist.

## Audio and Spotify Connect

SpotUI always targets the system Default Audio Output. It does not provide a
separate output-device selector.

- **Linux**: PulseAudio is preferred when `PULSE_SERVER`, the WSLg socket, or
  the user's PulseAudio socket is available; otherwise SpotUI uses ALSA's
  `default` device.
- **macOS**: SpotUI uses AudioToolbox's default output.

SpotUI advertises itself as a Spotify Connect computer named `SpotUI`. If
playback is transferred to another device, it remains there and SpotUI does
not take it back automatically. Local playback controls are unavailable while
the transfer is active.

Transient connection failures show a reconnect state. Search and playback
controls are temporarily disabled, and SpotUI retries with bounded backoff.

## Troubleshooting

### The browser did not open

Keep SpotUI running and copy the authorization URL shown on screen. Press `r`
to try opening it again. The first Login must be performed from a local
terminal on the same machine as SpotUI.

### There is no audio

Check the operating system's Default Audio Output and confirm that the output
device is connected before starting playback. On Linux, check the PulseAudio
connection (`PULSE_SERVER`, WSLg, or `$XDG_RUNTIME_DIR/pulse/native`) and then
the ALSA `default` device. SpotUI does not select a named output device.

### SpotUI is reconnecting

Wait for the reconnect attempt to finish. The current track is not resumed
automatically after reconnect. If the problem persists, check the network and
run SpotUI again; a new Login is only needed if the Local Session has expired.

### The session is invalid or corrupted

Use `L` to log out when possible. If SpotUI cannot load at all, remove only
`session.json` from the platform-specific SpotUI directory above and start it
again. This preserves `settings.json` and causes the next run to show Login.

### The terminal is too small

Resize the terminal to at least 50 columns by 16 rows. A terminal of at least
80 columns by 24 rows enables the full two-panel browse layout.

## Development documentation

- [Architecture and runtime boundaries](docs/architecture.md)
- [Premium release validation](docs/manual-playback-test.md)
- [Contributing guide](CONTRIBUTING.md)
- [Architecture decision records](docs/adr/)
- [Third-party notices](THIRD_PARTY_NOTICES.md)

## License

SpotUI is licensed under GPL-3.0-only. See [LICENSE](LICENSE) and
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
