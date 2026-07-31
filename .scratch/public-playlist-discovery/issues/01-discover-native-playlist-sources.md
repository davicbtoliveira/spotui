# Discover native playlist sources

Type: research

## Question

Which authenticated native Spotify protocols, callable from the pinned
go-librespot fork, can provide paginated Public Playlist summaries for each of
these surfaces: a text search, a personalized recommendation feed with an
editorial fallback, and playlists relevant to an artist? Establish whether the
same source can return User Playlists when Spotify ranks them, and identify
the response fields needed by `PlaylistSummary`.

## Constraints

- Do not add a separate official Spotify Web API client or client secret.
- Use primary sources: the pinned fork, its upstream protocol implementation,
  and Spotify-owned protocol responses only where safely reproducible.
