# Premium Release Validation

Run this matrix before publishing a release. It requires a dedicated Spotify
Premium test account and speakers or headphones connected to the system
Default Audio Output. Never record credentials, authorization URLs, or
session-file contents.

## Supported platform matrix

| Target | Fresh host | Output backend | Result | Evidence |
|--------|------------|----------------|--------|----------|
| Linux amd64 | Supported distribution, no SpotUI services installed | PulseAudio or ALSA Default Audio Output | Pending | Release notes |
| macOS amd64 | Intel Mac, no SpotUI services installed | AudioToolbox Default Audio Output | Pending | Release notes |
| macOS arm64 | Apple Silicon Mac, no SpotUI services installed | AudioToolbox Default Audio Output | Pending | Release notes |

`Pending` is intentional until a human runs the release candidate on real
hardware. Automated tests do not replace audible playback validation.

For Linux releases, validate ALSA on a native host and PulseAudio on a host
with `PULSE_SERVER` configured, a user PulseAudio socket, or under WSLg.

## Per-platform procedure

Use the release archive, not `go run`, so this also validates packaging.

1. Confirm the archive contains only `spotui`, `LICENSE`, and
   `THIRD_PARTY_NOTICES.md`.
2. Confirm no SpotUI daemon, sidecar, official Spotify player, or environment
   credential is installed or running.
3. Start `spotui` from a local terminal. Confirm it remains logged out and does
   not open a browser until `Enter` is pressed.
4. Complete Login in the browser with the Premium test account. Confirm the
   private session file is created and the browse shell opens on Liked Tracks.
5. Restart SpotUI. Confirm the session is reused, no browser opens, and the
   previous track does not resume.
6. Exercise the browse shell. Confirm Liked Tracks, Playlists, Saved Albums, and
   Recommended can be selected from navigation, and that `Tab` switches between
   navigation and content focus.
7. Open a playlist, album, and artist detail. Verify detail metadata, track
   pagination with `[` and `]`, selected artwork when available, and playback
   of a track in its album or playlist context.
8. Press `/` and search for a known query. Confirm grouped Tracks, Albums,
   Artists, and Playlists results, pagination, cursor movement, selection, and
   the `o` external-link action. If a non-track group is unavailable, confirm
   the UI labels it as unavailable instead of failing the whole search.
9. Press `Enter` on a known track. Confirm audio starts through the system
   Default Audio Output and metadata, buffering, playing state, and progress
   update in the TUI.
10. Verify play/pause, next, previous, volume down/up, Autoplay, Shuffle for a
    contextual track, and ten-second backward/forward seeking.
11. Select SpotUI from Spotify Connect, then transfer playback away. Confirm
    SpotUI reports the transfer, disables local playback controls, and does not
    steal playback back.
12. Interrupt connectivity temporarily. Confirm a visible reconnect state,
    disabled search/playback controls, and recovery without browser Login or
    automatic track resume.
13. Disconnect the output device and play a track. Confirm a visible error and
    a responsive terminal.
14. Press `L`. Confirm playback stops, the session is deleted, the Autoplay
    preference remains local, and the Logged-Out screen returns.
15. Quit. Confirm audio stops and no SpotUI process remains.

## Free-account guard

Separately log in with a Free test account. Confirm SpotUI explains that
Premium is required and exposes only logout and quit. Confirm no playback
engine, audio stream, or Connect device starts.

## Evidence template

Record one block per target in the release notes:

```text
Target:
SpotUI commit/tag:
Archive SHA-256:
OS version:
CPU:
Output device:
Catalog sections checked:
Track URI:
Steps 1-15:
Free-account guard:
Observed errors:
Tester/date:
```
