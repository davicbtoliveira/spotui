# Enable Configurable Autoplay

## Status

Accepted

## Context

When the current playback context has no next track, go-librespot can resolve a
Spotify Autoplay station and continue with recommendations.

Continuous playback is the preferred default, but users need a way to stop at
the end of the selected context.

go-librespot currently exposes Autoplay only as daemon configuration. A
race-free runtime toggle needs an explicit player request rather than mutating
shared configuration from the TUI.

## Decision

Spotify Autoplay will be enabled by default.

SpotUI will expose a persistent user option to disable or re-enable it. Changing
the option will affect the running player without requiring Logout or restart.

The `a` key will toggle Autoplay while the player is active. The player and help
views will show its current state, and a status message will confirm each
change.

The temporary go-librespot fork and upstream proposal will include a
runtime-safe Autoplay control request if upstream does not expose one before
implementation.

When Autoplay is disabled and no queued or contextual next track exists,
playback will stop at the end of the current track.

## Consequences

- `next` can advance into Spotify recommendations while Autoplay is enabled.
- The preference belongs to SpotUI settings, not the account credential file.
- The preference is reused on later SpotUI starts.
- External Spotify Connect commands may supply their own context or queue; those
  tracks take precedence over Autoplay.
- The TUI must display current Autoplay state and confirm changes.
