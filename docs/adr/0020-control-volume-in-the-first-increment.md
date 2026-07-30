# Control Volume in the First Increment

## Status

Accepted

## Context

A standalone local player should control its own playback volume. go-librespot
already exposes runtime volume requests and Spotify Connect volume events.

## Decision

The Playback Tracer Bullet will include player volume:

- `-` lowers volume by five percentage points;
- `+` raises volume by five percentage points;
- values clamp between 0% and 100%;
- the player view displays the current percentage;
- local and external Spotify Connect volume changes update the same displayed
  state.

Mute is deferred.

## Consequences

- Volume becomes part of event-driven player state.
- SpotUI must avoid feedback loops when receiving externally initiated volume
  changes.
- The last go-librespot volume is reused unless the user changes it.
- OS master volume remains independent.
