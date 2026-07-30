# Disable Audio Cache in the First Increment

## Status

Accepted

## Context

go-librespot can cache encrypted audio files to reduce repeated downloads.
Caching adds disk usage, eviction policy, configuration, cleanup, and privacy
expectations that are not needed to prove the Playback Tracer Bullet.

## Decision

SpotUI will disable go-librespot's on-disk audio cache in the first increment.

Only the private Local Session and minimal non-audio application state may be
persisted.

## Consequences

- Replaying a track downloads its audio again.
- Logout does not need to locate or clear an audio cache.
- Disk use remains small and predictable.
- An encrypted, size-bounded cache can be introduced later as an explicit
  opt-in setting.
