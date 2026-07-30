# Premium Release Validation

Run this matrix before publishing a release. It requires a dedicated Spotify
Premium test account and speakers or headphones connected to the system
Default Audio Output. Never record credentials or session-file contents.

## Supported platform matrix

| Target | Fresh host | Output backend | Result | Evidence |
|--------|------------|----------------|--------|----------|
| Linux amd64 | Supported distribution, no SpotUI services installed | ALSA Default Audio Output | Pending | Release notes |
| macOS amd64 | Intel Mac, no SpotUI services installed | AudioToolbox Default Audio Output | Pending | Release notes |
| macOS arm64 | Apple Silicon Mac, no SpotUI services installed | AudioToolbox Default Audio Output | Pending | Release notes |

`Pending` is intentional until a human runs the release candidate on real
hardware. Automated tests do not replace audible playback validation.

## Per-platform procedure

Use the release archive, not `go run`, so this also validates packaging.

1. Confirm the archive contains only `spotui`, `LICENSE`, and
   `THIRD_PARTY_NOTICES.md`.
2. Confirm no SpotUI daemon, sidecar, official Spotify player, or environment
   credential is installed or running.
3. Start `spotui` from a local terminal. Confirm it remains logged out and does
   not open a browser until `Enter` is pressed.
4. Complete Login in the browser with the Premium test account. Confirm the
   private session file is created and search becomes available.
5. Restart SpotUI. Confirm the session is reused, no browser opens, and the
   previous track does not resume.
6. Search for a known track. Verify title, artist, duration, pagination, cursor
   stability, and selection.
7. Press `Enter`. Confirm audio starts through the system Default Audio Output
   and metadata, buffering, playing state, and progress update in the TUI.
8. Verify play/pause, next, previous, volume down/up, and Autoplay.
9. Select SpotUI from Spotify Connect, then transfer playback away. Confirm
   SpotUI reports the transfer and does not steal playback back.
10. Interrupt connectivity temporarily. Confirm a visible reconnect state,
    disabled search/playback controls, and recovery without browser Login or
    automatic track resume.
11. Disconnect the output device and play a track. Confirm a visible error and
    a responsive terminal.
12. Press `L`. Confirm playback stops, the session is deleted, and the logged
    out screen returns.
13. Quit. Confirm audio stops and no SpotUI process remains.

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
Track URI:
Steps 1-13:
Free-account guard:
Observed errors:
Tester/date:
```
