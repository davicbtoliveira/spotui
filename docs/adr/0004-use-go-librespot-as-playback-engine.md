# Use go-librespot as the Playback Engine

## Status

Accepted

## Context

SpotUI must authenticate users without a Spotify developer application or
allowlist and must play audio locally without another Spotify player.

The official Spotify Web API cannot provide that experience to an unapproved
public application. The Rust librespot project would require a sidecar process
or a Rust-to-Go integration layer.

go-librespot is a Spotify Connect-compatible client implemented in Go. It
supports interactive browser OAuth, local audio playback, player state,
credentials persistence, and programmatic playback control.

## Decision

SpotUI will use `github.com/devgianlu/go-librespot` as its authentication,
session, metadata, and local playback engine.

SpotUI will integrate go-librespot in-process behind internal interfaces. The
TUI will not require a separately installed daemon or another Spotify client.

## Consequences

- SpotUI can remain primarily a Go project.
- Users need Spotify Premium for music playback.
- The existing Spotify Web API OAuth client and remote-control playback path
  will be replaced.
- Authentication credentials and playback state must be stored with restricted
  filesystem permissions.
- go-librespot uses unofficial Spotify protocols and can break when Spotify
  changes private services.
- go-librespot is GPL-3.0 licensed. Distribution of an in-process integration
  requires SpotUI to use a compatible license and satisfy GPL obligations.
- Audio backends and native codec dependencies must be handled by build and
  release packaging.

## Reference

- https://github.com/devgianlu/go-librespot
