# Stop Playback on Quit and Preserve Session

## Status

Accepted

## Context

With local playback, quitting SpotUI terminates the process producing audio.
This differs from the previous remote-control behavior, where another Spotify
player could continue after the TUI exited.

Quitting is also distinct from Logout: quitting should not force browser
authentication on the next start.

## Decision

Quitting SpotUI will:

1. stop local audio;
2. close the go-librespot runtime cleanly;
3. preserve the Local Session;
4. exit the TUI.

Starting SpotUI again will reuse the Local Session but will not automatically
resume the previous track.

## Consequences

- `q` always silences SpotUI-owned playback.
- `q` never removes credentials.
- Shutdown must wait for bounded player cleanup without hanging the terminal.
- Crash recovery must not interpret an unclean exit as a Logout.
