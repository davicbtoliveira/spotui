# Use Authenticated Pathfinder for Public Playlist Discovery

## Status

Accepted

## Context

SpotUI's native catalog can list User Playlists and open a playlist with a
known URI, but it cannot discover Public Playlists in search, recommendations,
or artist details. The public Spotify Web API is deliberately excluded because
its shared application quota previously made catalog browsing unreliable.

## Decision

The pinned go-librespot fork will issue Spotify Pathfinder persisted queries
through the existing authenticated `Spclient` token flow. SpotUI uses those
queries to obtain Public Playlist summaries for Search, Recommended, and
artist pages, then continues to use the established native context resolver
for playlist detail and playback.

Each discovery surface is best-effort. A failed Pathfinder query renders that
playlist section as unavailable without failing tracks, albums, artists, or
the rest of the page.

## Consequences

- Public Playlist discovery remains inside the one-process authenticated
  runtime and does not add a client secret or Web API dependency.
- Persisted-query names and hashes are an unofficial integration contract and
  must be revalidated when Spotify changes its web client or responses.
- Playlist opening and context playback remain unchanged, which keeps the
  playback contract independent of the discovery source.
