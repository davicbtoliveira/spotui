# Deliver a Playback Tracer Bullet First

## Status

Accepted

## Context

Replacing the official Web API client with go-librespot affects authentication,
session persistence, catalog access, playback, player events, audio output, and
every library surface.

Migrating all current library views before validating local playback would delay
feedback on the highest-risk integration.

## Decision

The first functional increment will deliver one end-to-end path containing:

- explicit browser login;
- persisted session reuse;
- confirmed logout;
- Track Search;
- selection of a Search Result;
- local playback through go-librespot;
- play, pause, next, and previous controls;
- five-percent volume controls;
- current-track and progress display.

Playlists, saved tracks, top artists, shuffle, and other library surfaces will
be migrated in later increments.

## Consequences

- The highest-risk authentication-to-audio path is validated first.
- Existing library features may be temporarily unavailable on the migration
  branch but must not be silently presented as working.
- Internal interfaces should allow later library surfaces without coupling the
  TUI directly to go-librespot packages.
- The first increment needs integration tests around the boundary and manual
  end-to-end validation with a Spotify Premium account.
