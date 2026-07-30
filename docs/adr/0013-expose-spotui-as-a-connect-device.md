# Expose SpotUI as a Spotify Connect Device

## Status

Accepted

## Context

go-librespot implements a Spotify Connect-compatible device. SpotUI must work
without another Spotify player, but users may still benefit from controlling
the active terminal player from a phone or desktop.

External control can change track, position, pause state, and queue while the
TUI is running.

## Decision

An authenticated SpotUI instance will register as a Spotify Connect device
named `SpotUI`.

External Spotify clients may control it, but no external client is required to
start or manage playback from the TUI.

go-librespot player events will be the source of truth. Changes initiated
externally will update the TUI's current track, progress, and playback state.

## Consequences

- SpotUI participates in the user's Spotify Connect device list.
- The TUI must consume event-driven state instead of assuming commands are the
  only source of player changes.
- Conflicting local and external commands resolve according to the latest state
  reported by go-librespot.
- Device registration must stop on Quit or Logout.
- The device name can become configurable later, but defaults to `SpotUI`.
